package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

// writeFleetProject writes a project record with a status, a task ref, and an
// optional **phase-a:** line, under <parent>/<repo>/workshop/projects/<name>.md.
func writeFleetProject(t *testing.T, parent, repo, name, status, taskRef, phaseA string) string {
	t.Helper()
	dir := filepath.Join(parent, repo, "workshop", "projects")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	est := ""
	if phaseA != "" {
		est = "\n## Estimate\n\n**phase-a:** " + phaseA + "\n"
	}
	body := "---\ntype: project\nname: " + name + "\ngoal: g\ndone_when: w\nstatus: " + status +
		"\n---\n\n# " + name + est + "\n\n## Breakdown\n\n- [ ] a task [" + taskRef + "]\n"
	p := filepath.Join(dir, name+".md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// stubIssueLookup makes projectIssueLookupFn return a fixed estimate for refs
// present in the map, and an error otherwise (an unresolvable breakdown row).
func stubIssueLookup(t *testing.T, estimates map[string]float64) {
	t.Helper()
	orig := projectIssueLookupFn
	t.Cleanup(func() { projectIssueLookupFn = orig })
	projectIssueLookupFn = func(ref, _ string) (issueMeta, error) {
		if h, ok := estimates[ref]; ok {
			return issueMeta{Identity: ref, Status: "working", EstimateHours: h}, nil
		}
		return issueMeta{}, os.ErrNotExist
	}
}

func TestListFleetProjects(t *testing.T) {
	parent := t.TempDir()
	subject := writeFleetProject(t, parent, "ariadne", "subject", "executing", "ariadne#182", "")
	writeFleetProject(t, parent, "metis", "m", "executing", "metis#2", "")
	writeFleetProject(t, parent, "nous", "paused-proj", "paused", "nous#9", "")
	// A sibling whose breakdown row won't resolve but carries a Phase-A total.
	writeFleetProject(t, parent, "kbench", "phasea-proj", "committed", "kbench#99", "12h")
	// A sibling with neither resolvable rows nor Phase-A → unknown.
	writeFleetProject(t, parent, "pair", "mystery", "executing", "pair#404", "")

	stubIssueLookup(t, map[string]float64{
		"metis#2": 20,
		"nous#9":  14,
	})

	loads := ListFleetProjects(parent, subject)
	byName := map[string]projectdoc.ProjectLoad{}
	for _, l := range loads {
		byName[l.Name] = l
	}
	if _, ok := byName["subject"]; ok {
		t.Error("subject excluded")
	}
	if m := byName["m"]; m.Status != "executing" || m.RemainingHours != 20 || m.RemainingSource != "board" {
		t.Errorf("metis load wrong: %+v", m)
	}
	if p := byName["paused-proj"]; p.Status != "paused" {
		t.Errorf("nous paused load wrong: %+v", p)
	}
	if pa := byName["phasea-proj"]; pa.RemainingHours != 12 || pa.RemainingSource != "phase-a" {
		t.Errorf("phase-a fallback wrong: %+v", pa)
	}
	if my := byName["mystery"]; my.RemainingSource != "unknown" || my.Warning == "" {
		t.Errorf("unknown load should carry a warning: %+v", my)
	}
}

// TestListFleetProjects_AllTerminalReadsBoardZero pins the M2-review Important:
// an open project whose breakdown fully resolved to terminal issues reads board
// source with 0 remaining (NOT the stale Phase-A total), and does not contend.
func TestListFleetProjects_AllTerminalReadsBoardZero(t *testing.T) {
	parent := t.TempDir()
	subject := writeFleetProject(t, parent, "ariadne", "subject", "executing", "ariadne#182", "")
	// A committed sibling whose one breakdown row resolved to a DONE issue, but
	// which carries a Phase-A of 12h. Must read board/0, not phase-a/12.
	writeFleetProject(t, parent, "kbench", "burned-down", "committed", "kbench#1", "12h")

	orig := projectIssueLookupFn
	t.Cleanup(func() { projectIssueLookupFn = orig })
	projectIssueLookupFn = func(ref, _ string) (issueMeta, error) {
		// kbench#1 resolves to a terminal (done) issue → board remaining 0.
		return issueMeta{Identity: ref, Status: "done", EstimateHours: 5}, nil
	}

	loads := ListFleetProjects(parent, subject)
	var bd projectdoc.ProjectLoad
	for _, l := range loads {
		if l.Name == "burned-down" {
			bd = l
		}
	}
	if bd.RemainingSource != "board" {
		t.Errorf("all-terminal project should read board source (not phase-a), got %q", bd.RemainingSource)
	}
	if bd.RemainingHours != 0 {
		t.Errorf("all-terminal project remaining = %v, want 0", bd.RemainingHours)
	}
	// And it must not contend: a solo subject's forecast stays N=1.
	f, err := projectdoc.ComputeForecast(
		estimate.ThroughputBaseline{HoursPerWeek: 40, Ceiling: 2},
		projectdoc.ProjectLoad{Name: "subject", Status: "executing", RemainingHours: 40, RemainingSource: "board"},
		loads, "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	if f.N != 1 {
		t.Errorf("burned-down project must not contend: N = %d, want 1", f.N)
	}
}

func TestLoadThroughputBaseline_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.tsv")
	content := estimate.BaselineHeader() + "\n2026-07-19\t2026-06-22\t2026-07-19\t110.00\t280\t2\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_THROUGHPUT_BASELINE", path)
	b, err := loadThroughputBaseline("/nonexistent-brain")
	if err != nil {
		t.Fatal(err)
	}
	if b.HoursPerWeek != 110.0 {
		t.Errorf("HoursPerWeek = %.2f, want 110", b.HoursPerWeek)
	}
}

func TestLoadThroughputBaseline_AbsentIsErrNoBaseline(t *testing.T) {
	t.Setenv("WF_THROUGHPUT_BASELINE", filepath.Join(t.TempDir(), "missing.tsv"))
	if _, err := loadThroughputBaseline("/nonexistent"); err != errNoBaseline {
		t.Errorf("absent baseline should map to errNoBaseline, got %v", err)
	}
}

func TestLoadThroughputBaseline_UnparsableIsErrNoBaseline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.tsv")
	// header + a row with a bad float column
	if err := os.WriteFile(path, []byte(estimate.BaselineHeader()+"\n2026-07-19\ts\te\tNOPE\t1\t2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_THROUGHPUT_BASELINE", path)
	if _, err := loadThroughputBaseline("/nonexistent"); err != errNoBaseline {
		t.Errorf("unparsable baseline should map to errNoBaseline, got %v", err)
	}
}

// A no-baseline forecastForProject bubbles errNoBaseline so consumers pick
// their own fallback.
func TestForecastForProject_NoBaseline(t *testing.T) {
	parent := t.TempDir()
	subject := writeFleetProject(t, parent, "ariadne", "subj", "committed", "ariadne#182", "40h")
	stubIssueLookup(t, map[string]float64{})
	t.Setenv("WF_THROUGHPUT_BASELINE", filepath.Join(t.TempDir(), "missing.tsv"))
	d, err := readProject(subject)
	if err != nil {
		t.Fatal(err)
	}
	_, _, ferr := forecastForProject(d, subject, parent, "/nonexistent-brain", "2026-09-01")
	if ferr != errNoBaseline {
		t.Errorf("want errNoBaseline, got %v", ferr)
	}
}

func TestForecastForProject_WithBaseline(t *testing.T) {
	parent := t.TempDir()
	// subject committed with a Phase-A of 55h, deadline set; no other projects.
	subject := writeFleetProjectDeadline(t, parent, "ariadne", "subj", "committed", "ariadne#182", "55h", "2026-09-01")
	stubIssueLookup(t, map[string]float64{})
	blPath := filepath.Join(t.TempDir(), "baseline.tsv")
	if err := os.WriteFile(blPath, []byte(estimate.BaselineHeader()+"\n2026-07-19\ts\te\t55.00\t10\t2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WF_THROUGHPUT_BASELINE", blPath)
	d, err := readProject(subject)
	if err != nil {
		t.Fatal(err)
	}
	f, deadline, ferr := forecastForProject(d, subject, parent, "/nonexistent-brain", "2026-09-01")
	if ferr != nil {
		t.Fatalf("forecast: %v", ferr)
	}
	if deadline != "2026-09-01" {
		t.Errorf("deadline = %q, want 2026-09-01", deadline)
	}
	// 55h ÷ 55h/wk ÷ 1 = 1wk = 7 days → 2026-09-08.
	if f.ProjectedFinish != "2026-09-08" {
		t.Errorf("ProjectedFinish = %q, want 2026-09-08", f.ProjectedFinish)
	}
	if f.RemainingSource != "phase-a" {
		t.Errorf("RemainingSource = %q, want phase-a", f.RemainingSource)
	}
}

func writeFleetProjectDeadline(t *testing.T, parent, repo, name, status, taskRef, phaseA, deadline string) string {
	t.Helper()
	dir := filepath.Join(parent, repo, "workshop", "projects")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	est := "\n## Estimate\n\n**phase-a:** " + phaseA + "\n"
	body := "---\ntype: project\nname: " + name + "\ngoal: g\ndone_when: w\nstatus: " + status +
		"\ndeadline: " + deadline + "\n---\n\n# " + name + est + "\n\n## Breakdown\n\n- [ ] a task [" + taskRef + "]\n"
	p := filepath.Join(dir, name+".md")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}
