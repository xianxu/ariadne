package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

func writeStatusProject(t *testing.T, dir, status, prd, estimate, extraFM string) string {
	t.Helper()
	path := filepath.Join(dir, "demo.md")
	body := "---\ntype: project\nname: demo\ngoal: g\ndone_when: d\nstatus: " + status + "\n" + extraFM + "updated: 2026-07-01\n---\n## PRD\n" + prd + "\n## Estimate\n" + estimate + "\n## Breakdown\n- [ ]\n## Log\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestApplyProjectStatusUsesModelAndGuards(t *testing.T) {
	dir := t.TempDir()
	path := writeStatusProject(t, dir, "ideation", "A real PRD.", "", "")
	prev, changed, err := applyProjectStatus(path, "defined", false, projectdoc.GuardCtx{Today: "2026-07-16"})
	if err != nil || prev != "ideation" || !changed {
		t.Fatalf("apply = %q,%v,%v", prev, changed, err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "status: defined") || !strings.Contains(string(b), "updated: 2026-07-16") {
		t.Fatalf("file not updated:\n%s", b)
	}

	if _, _, err := applyProjectStatus(path, "executing", true, projectdoc.GuardCtx{Today: "2026-07-16"}); err == nil || !strings.Contains(err.Error(), "legal from") {
		t.Fatalf("force bypassed lifecycle legality: %v", err)
	}
	if _, _, err := applyProjectStatus(path, "bogus", true, projectdoc.GuardCtx{}); err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("unknown status accepted: %v", err)
	}
}

func TestApplyProjectStatusForceWaivesNamedGuardsAndDoneRoutesToClose(t *testing.T) {
	dir := t.TempDir()
	path := writeStatusProject(t, dir, "ideation", "", "", "")
	if _, _, err := applyProjectStatus(path, "defined", false, projectdoc.GuardCtx{}); err == nil || !strings.Contains(err.Error(), "prd-present") {
		t.Fatalf("missing PRD guard = %v", err)
	}
	if _, changed, err := applyProjectStatus(path, "defined", true, projectdoc.GuardCtx{Today: "2026-07-16"}); err != nil || !changed {
		t.Fatalf("force did not waive guard: %v", err)
	}

	path = writeStatusProject(t, dir, "executing", "ok", "", "deadline: 2026-09-01\nplanned_finish: 2026-08-20\n")
	if _, _, err := applyProjectStatus(path, "done", true, projectdoc.GuardCtx{}); err == nil || !strings.Contains(err.Error(), "sdlc project close") {
		t.Fatalf("done did not route to close: %v", err)
	}
}

func TestApplyProjectStatusRefusesUnknownModelGuard(t *testing.T) {
	dir := t.TempDir()
	path := writeStatusProject(t, dir, "ideation", "ok", "", "")
	orig := projectGuardsFn
	projectGuardsFn = func() map[string]projectdoc.GuardFunc { return map[string]projectdoc.GuardFunc{} }
	t.Cleanup(func() { projectGuardsFn = orig })
	if _, _, err := applyProjectStatus(path, "defined", false, projectdoc.GuardCtx{}); err == nil || !strings.Contains(err.Error(), "unknown project guard") {
		t.Fatalf("unknown guard = %v", err)
	}
}

func TestApplyProjectStatusAppendsEvidence(t *testing.T) {
	dir := t.TempDir()
	path := writeStatusProject(t, dir, "defined", "ok", "\n**phase-a:** 3h\n", "deadline: 2026-09-01\nplanned_finish: 2026-08-20\n")
	ctx := projectdoc.GuardCtx{Today: "2026-07-16", Evidence: map[string]string{"reality-check": "capacity checked"}}
	if _, _, err := applyProjectStatus(path, "committed", false, ctx); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "reality-check: capacity checked") {
		t.Fatalf("evidence not logged:\n%s", b)
	}
}
