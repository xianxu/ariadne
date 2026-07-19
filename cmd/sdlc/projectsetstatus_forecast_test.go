package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// commitFixture writes a `defined` project (PRD + Phase-A, deadline) under a
// real <parent>/<repo>/workshop/projects/ layout so the forecast's fleet parent
// resolves, and blesses a baseline via WF_THROUGHPUT_BASELINE. Returns the
// project's ProjectsDir. today is pinned to 2026-09-01.
func commitFixture(t *testing.T, hpw float64) (projectsDir string) {
	t.Helper()
	parent := t.TempDir()
	projectsDir = filepath.Join(parent, "ariadne", "workshop", "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\ntype: project\nname: demo\ngoal: g\ndone_when: d\nstatus: defined\n" +
		"deadline: 2026-09-01\nupdated: 2026-07-01\n---\n## PRD\nReal PRD.\n" +
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

func TestSetStatusCommit_ComputesAndDerivesPlannedFinish(t *testing.T) {
	dir := commitFixture(t, 55)
	var out strings.Builder
	f := &projectSetStatusFlags{Slug: "demo", To: "committed", ProjectsDir: dir, BrainDir: "/nonexistent-brain"}
	if err := runProjectSetStatus(&out, &out, f); err != nil {
		t.Fatalf("commit: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "forecast:") {
		t.Errorf("statement not printed: %q", out.String())
	}
	b, _ := os.ReadFile(filepath.Join(dir, "demo.md"))
	txt := string(b)
	// 55h ÷ 55h/wk ÷ 1 = 1wk = 7 days → 2026-09-08 derived.
	if !strings.Contains(txt, "planned_finish: 2026-09-08") {
		t.Errorf("planned_finish not derived:\n%s", txt)
	}
	if !strings.Contains(txt, "reality-check: computed:") {
		t.Errorf("computed statement not recorded as reality-check evidence:\n%s", txt)
	}
	if !strings.Contains(txt, "status: committed") {
		t.Errorf("status not flipped:\n%s", txt)
	}
}

func TestSetStatusCommit_PreExistingPlannedFinishKept(t *testing.T) {
	dir := commitFixture(t, 55)
	// Add a pre-existing planned_finish to the file.
	p := filepath.Join(dir, "demo.md")
	b, _ := os.ReadFile(p)
	txt := strings.Replace(string(b), "deadline: 2026-09-01\n", "deadline: 2026-09-01\nplanned_finish: 2026-08-15\n", 1)
	os.WriteFile(p, []byte(txt), 0o644)

	var out strings.Builder
	f := &projectSetStatusFlags{Slug: "demo", To: "committed", ProjectsDir: dir, BrainDir: "/nonexistent-brain"}
	if err := runProjectSetStatus(&out, &out, f); err != nil {
		t.Fatalf("commit: %v\n%s", err, out.String())
	}
	got, _ := os.ReadFile(p)
	if !strings.Contains(string(got), "planned_finish: 2026-08-15") {
		t.Errorf("pre-existing planned_finish should be kept:\n%s", got)
	}
	if !strings.Contains(string(got), "manual planned_finish kept") {
		t.Errorf("keep should be noted in the Log:\n%s", got)
	}
}

func TestSetStatusCommit_ExplicitPlannedFinishWins(t *testing.T) {
	dir := commitFixture(t, 55)
	var out strings.Builder
	f := &projectSetStatusFlags{Slug: "demo", To: "committed", PlannedFinish: "2026-10-01", ProjectsDir: dir, BrainDir: "/nonexistent-brain"}
	if err := runProjectSetStatus(&out, &out, f); err != nil {
		t.Fatalf("commit: %v\n%s", err, out.String())
	}
	got, _ := os.ReadFile(filepath.Join(dir, "demo.md"))
	if !strings.Contains(string(got), "planned_finish: 2026-10-01") {
		t.Errorf("explicit --planned-finish should win:\n%s", got)
	}
	if !strings.Contains(string(got), "planned_finish set manually") {
		t.Errorf("manual set should be noted:\n%s", got)
	}
}

func TestSetStatusCommit_NoBaselineRefusesWithHint(t *testing.T) {
	dir := commitFixture(t, 55)
	t.Setenv("WF_THROUGHPUT_BASELINE", filepath.Join(t.TempDir(), "missing.tsv"))
	var out strings.Builder
	f := &projectSetStatusFlags{Slug: "demo", To: "committed", ProjectsDir: dir, BrainDir: "/nonexistent-brain"}
	err := runProjectSetStatus(&out, &out, f)
	if err == nil || !strings.Contains(err.Error(), "reality-check") {
		t.Fatalf("no baseline + no --reality should refuse on the guard: %v", err)
	}
	if !strings.Contains(err.Error(), "throughput --bless") {
		t.Errorf("refusal should hint the bless command: %v", err)
	}
}

func TestSetStatusCommit_NoBaselineWithRealityPasses(t *testing.T) {
	dir := commitFixture(t, 55)
	t.Setenv("WF_THROUGHPUT_BASELINE", filepath.Join(t.TempDir(), "missing.tsv"))
	// Legacy flow: with no baseline the operator sets planned_finish by hand
	// (the pre-#182 baseline-set requirement) and attests via --reality.
	p := filepath.Join(dir, "demo.md")
	b, _ := os.ReadFile(p)
	os.WriteFile(p, []byte(strings.Replace(string(b), "deadline: 2026-09-01\n", "deadline: 2026-09-01\nplanned_finish: 2026-08-20\n", 1)), 0o644)

	var out strings.Builder
	f := &projectSetStatusFlags{Slug: "demo", To: "committed", Reality: "fits, checked manually", ProjectsDir: dir, BrainDir: "/nonexistent-brain"}
	if err := runProjectSetStatus(&out, &out, f); err != nil {
		t.Fatalf("no baseline + --reality should pass (legacy fallback): %v\n%s", err, out.String())
	}
	got, _ := os.ReadFile(filepath.Join(dir, "demo.md"))
	if !strings.Contains(string(got), "reality-check: fits, checked manually") {
		t.Errorf("legacy --reality evidence should be recorded:\n%s", got)
	}
}

func TestSetStatusNonCommit_Untouched(t *testing.T) {
	// A non-committed transition must not invoke the forecast at all.
	dir := t.TempDir()
	path := writeStatusProject(t, dir, "ideation", "A real PRD.", "", "")
	_ = path
	var out strings.Builder
	f := &projectSetStatusFlags{Slug: "demo", To: "defined", ProjectsDir: dir, BrainDir: "/nonexistent-brain"}
	if err := runProjectSetStatus(&out, &out, f); err != nil {
		t.Fatalf("defined transition: %v", err)
	}
	if strings.Contains(out.String(), "forecast:") {
		t.Errorf("non-committed transition must not compute a forecast: %q", out.String())
	}
}
