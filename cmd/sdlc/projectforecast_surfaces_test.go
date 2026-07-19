package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// surfaceFixture writes a committed project (Phase-A 55h, deadline) under a
// real repo layout and blesses a baseline. Returns the ProjectsDir.
func surfaceFixture(t *testing.T, hpw float64) string {
	t.Helper()
	parent := t.TempDir()
	projectsDir := filepath.Join(parent, "ariadne", "workshop", "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\ntype: project\nname: demo\ngoal: g\ndone_when: d\nstatus: executing\n" +
		"deadline: 2026-09-01\nplanned_finish: 2026-08-20\nupdated: 2026-07-01\n---\n## PRD\np\n" +
		"## Estimate\n\n**phase-a:** 55h\n## Breakdown\n- [ ]\n## Log\n"
	if err := os.WriteFile(filepath.Join(projectsDir, "demo.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	blPath := filepath.Join(t.TempDir(), "baseline.tsv")
	row := estimate.RenderBaselineRow(estimate.ThroughputBaseline{
		BlessedDate: "2026-07-19", SpanStart: "2026-06-22", SpanEnd: "2026-07-19",
		HoursPerWeek: hpw, Rows: 10, Ceiling: 2,
	})
	if err := os.WriteFile(blPath, []byte(estimate.BaselineHeader()+"\n"+row+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_THROUGHPUT_BASELINE", blPath)
	orig := projectTodayFn
	projectTodayFn = func() string { return "2026-09-01" }
	t.Cleanup(func() { projectTodayFn = orig })
	return projectsDir
}

func TestProjectShow_IncludesForecast(t *testing.T) {
	dir := surfaceFixture(t, 55)
	var out strings.Builder
	if err := runProjectShow(&out, &out, &projectShowFlags{Slug: "demo", ProjectsDir: dir, BrainDir: "/nonexistent"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "forecast:") || !strings.Contains(out.String(), "lands ~2026-09-08") {
		t.Errorf("show missing forecast drift line:\n%s", out.String())
	}
}

func TestProjectStatus_IncludesForecast(t *testing.T) {
	dir := surfaceFixture(t, 55)
	stubIssueLookup(t, map[string]float64{})
	var out strings.Builder
	if err := runProjectStatus(&out, &out, &projectStatusFlags{Slug: "demo", ProjectsDir: dir, BrainDir: "/nonexistent"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "forecast:") {
		t.Errorf("status missing forecast line:\n%s", out.String())
	}
}

func TestProjectShow_NoBaselineQuietLine(t *testing.T) {
	dir := surfaceFixture(t, 55)
	t.Setenv("WF_THROUGHPUT_BASELINE", filepath.Join(t.TempDir(), "missing.tsv"))
	var out strings.Builder
	if err := runProjectShow(&out, &out, &projectShowFlags{Slug: "demo", ProjectsDir: dir, BrainDir: "/nonexistent"}); err != nil {
		t.Fatalf("show must not fail without a baseline: %v", err)
	}
	if !strings.Contains(out.String(), "no blessed baseline") {
		t.Errorf("show should print a quiet no-baseline line:\n%s", out.String())
	}
}

// --- Calendar ledger row at close ---

func TestPrepareCalendarLedgerRow(t *testing.T) {
	text := "# L\n\n## Calendar ledger\n\n| project | planned | actual | slip_days | closed |\n|---|---|---|---:|---|\n"
	// planned 2026-08-20, actual 2026-09-01 → 12 days late.
	got, err := prepareCalendarLedgerRow(text, "demo", "2026-08-20", "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "| demo | 2026-08-20 | 2026-09-01 | 12 | 2026-09-01 |") {
		t.Errorf("calendar row wrong:\n%s", got)
	}
}

func TestPrepareCalendarLedgerRow_NoPlannedFinish(t *testing.T) {
	text := "# L\n\n## Calendar ledger\n\n| project | planned | actual | slip_days | closed |\n|---|---|---|---:|---|\n"
	got, err := prepareCalendarLedgerRow(text, "demo", "", "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "| demo | n/a | 2026-09-01 | n/a | 2026-09-01 |") {
		t.Errorf("absent planned_finish should record n/a slip:\n%s", got)
	}
}

func TestPrepareCalendarLedgerRow_MissingHeadingErrors(t *testing.T) {
	text := "# L\n\n## Fog ledger\n\n| p |\n|---|\n"
	_, err := prepareCalendarLedgerRow(text, "demo", "2026-08-20", "2026-09-01")
	if err == nil || !strings.Contains(err.Error(), "Calendar ledger") {
		t.Errorf("missing Calendar ledger heading should error naming that heading: %v", err)
	}
}

func TestAppendProjectLedgerRow_FogHeadingUnchanged(t *testing.T) {
	// Regression: the fog path still works with the heading-parameterized fn.
	text := "# L\n\n## Fog ledger\n\n| project | phase-a | actuals | fog | closed |\n|---|---:|---:|---:|---|\n"
	got, err := appendProjectLedgerRow(text, "Fog ledger", "| demo | 5h | 4h | 0.80 | 2026-09-01 |")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "| demo | 5h | 4h | 0.80 | 2026-09-01 |") {
		t.Errorf("fog row not appended:\n%s", got)
	}
}
