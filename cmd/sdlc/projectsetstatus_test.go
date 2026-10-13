package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

const oneTrain = "| Train | Cars, in order | Slot |\n|---|---|---|\n| A | ariadne#1 → #2 | :1 |\n"

func writeStatusProject(t *testing.T, dir, status, trains, extraFM string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "demo.md")
	body := "---\ntype: project\nname: demo\ngoal: g\ndone_when: d\nstatus: " + status + "\n" + extraFM + "updated: 2026-07-01\n---\n## Charter\nWhy.\n## Trains\n" + trains + "\n## Breakdown\n- [ ]\n## Log\n## Decisions\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestApplyProjectStatusUsesModelAndGuards(t *testing.T) {
	dir := t.TempDir()
	path := writeStatusProject(t, dir, "defined", "", "")
	if _, _, err := applyProjectStatus(path, "executing", false, projectdoc.GuardCtx{}); err == nil || !strings.Contains(err.Error(), "trains-present") {
		t.Fatalf("missing trains guard = %v", err)
	}
	path = writeStatusProject(t, dir, "defined", oneTrain, "")
	prev, changed, err := applyProjectStatus(path, "executing", false, projectdoc.GuardCtx{Today: "2026-07-16"})
	if err != nil || prev != "defined" || !changed {
		t.Fatalf("apply = %q,%v,%v", prev, changed, err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "status: executing") || !strings.Contains(string(b), "updated: 2026-07-16") {
		t.Fatalf("file not updated:\n%s", b)
	}
	if _, _, err := applyProjectStatus(path, "defined", true, projectdoc.GuardCtx{}); err == nil || !strings.Contains(err.Error(), "legal from") {
		t.Fatalf("force bypassed lifecycle legality: %v", err)
	}
	if _, _, err := applyProjectStatus(path, "bogus", true, projectdoc.GuardCtx{}); err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Fatalf("unknown status accepted: %v", err)
	}
}

func TestApplyProjectStatusForceWaivesNamedGuardsAndClosedRoutesToClose(t *testing.T) {
	dir := t.TempDir()
	path := writeStatusProject(t, dir, "defined", "", "")
	if _, changed, err := applyProjectStatus(path, "executing", true, projectdoc.GuardCtx{Today: "2026-07-16"}); err != nil || !changed {
		t.Fatalf("force did not waive guard: %v", err)
	}
	path = writeStatusProject(t, dir, "closing", oneTrain, "")
	if _, _, err := applyProjectStatus(path, "closed", true, projectdoc.GuardCtx{}); err == nil || !strings.Contains(err.Error(), "sdlc project close") {
		t.Fatalf("closed did not route to close: %v", err)
	}
	// reopen: new MVP work sends a closing project back to executing, unguarded.
	if _, changed, err := applyProjectStatus(path, "executing", false, projectdoc.GuardCtx{Today: "2026-07-16"}); err != nil || !changed {
		t.Fatalf("reopen = %v", err)
	}
}

func TestApplyProjectStatusRefusesUnknownModelGuard(t *testing.T) {
	dir := t.TempDir()
	path := writeStatusProject(t, dir, "defined", oneTrain, "")
	orig := projectGuardsFn
	projectGuardsFn = func() map[string]projectdoc.GuardFunc { return map[string]projectdoc.GuardFunc{} }
	t.Cleanup(func() { projectGuardsFn = orig })
	if _, _, err := applyProjectStatus(path, "executing", false, projectdoc.GuardCtx{}); err == nil || !strings.Contains(err.Error(), "unknown project guard") {
		t.Fatalf("unknown guard = %v", err)
	}
}

// The wrap edge reads every mvp_scope card through the shared resolver, from
// the project's own checkout.
func TestProjectSetStatusWrapChecksMVPCards(t *testing.T) {
	_, _, slot := projectWorkspaceFixture(t)
	path := writeStatusProject(t, filepath.Join(slot, "workshop", "projects"), "executing", oneTrain, "mvp_scope: [ariadne#1, ariadne#2]\n")
	metas := map[string]issueMeta{"ariadne#1": {Status: "done"}, "ariadne#2": {Status: "working"}}
	var roots []string
	stubForecastLookup(t, metas, &roots)
	f := &projectSetStatusFlags{Slug: "demo", To: "closing", ProjectsDir: filepath.Dir(path)}
	err := runProjectSetStatus(context.Background(), &strings.Builder{}, nil, f)
	if err == nil || !strings.Contains(err.Error(), "mvp_scope issues not landed: ariadne#2") {
		t.Fatalf("wrap with an open MVP issue = %v", err)
	}
	metas["ariadne#2"] = issueMeta{Status: "done"}
	if err := runProjectSetStatus(context.Background(), &strings.Builder{}, nil, f); err != nil {
		t.Fatalf("wrap with every MVP issue landed = %v", err)
	}
	for _, r := range roots {
		if r != slot {
			t.Fatalf("lookup ran from %q, want %q", r, slot)
		}
	}
}
