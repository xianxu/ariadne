package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderRetroEntry(t *testing.T) {
	b := board{Done: 3, Total: 7, RemainingHours: 22, Deadline: "2026-09-01", Frontier: []string{"ariadne#182", "metis#9"}}
	got := renderRetroEntry(b, "2026-07-20")
	for _, want := range []string{"### 2026-07-20 — retro", "**board:** 3/7 done · Σ remaining ≈ 22h · deadline 2026-09-01 (43 days) · frontier: ariadne#182, metis#9", "<where we are + what changed + new forecast"} {
		if !strings.Contains(got, want) {
			t.Errorf("retro missing %q:\n%s", want, got)
		}
	}
}

func TestRunProjectRetroAppendsAndDryRuns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.md")
	doc := "---\ntype: project\nname: demo\nstatus: executing\n---\n## Breakdown\n- [ ] plain\n## Log\n"
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
	orig := projectTodayFn
	projectTodayFn = func() string { return "2026-07-20" }
	t.Cleanup(func() { projectTodayFn = orig })
	var out, errOut bytes.Buffer
	f := &projectRetroFlags{Slug: "demo", ProjectsDir: dir}
	if err := runProjectRetro(&out, &errOut, f); err != nil {
		t.Fatal(err)
	}
	if err := runProjectRetro(&out, &errOut, f); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if got := strings.Count(string(b), "### 2026-07-20 — retro"); got != 2 {
		t.Fatalf("retro count = %d:\n%s", got, b)
	}
	before := string(b)
	f.DryRun = true
	out.Reset()
	if err := runProjectRetro(&out, &errOut, f); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(path)
	if string(after) != before || !strings.Contains(out.String(), "### 2026-07-20 — retro") {
		t.Fatal("dry-run wrote or failed to print")
	}
}

func TestProjectRetroCommandRegistered(t *testing.T) {
	project, _, _ := buildRoot().Find([]string{"project"})
	found, _, err := project.Find([]string{"retro"})
	if err != nil || found == project {
		t.Fatalf("project retro not registered: %v", err)
	}
}
