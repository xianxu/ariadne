package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// fakeRunner records each Run call and returns a canned error — so the generate
// stage is tested with NO real binary spawned (the production ExecRunner is
// integration-tested in weavefs/runner_test.go).
type fakeRunner struct {
	dirs []string
	err  error
}

func (f *fakeRunner) Run(dir string, _ []string) error {
	f.dirs = append(f.dirs, dir)
	return f.err
}

var _ weavefs.Runner = (*fakeRunner)(nil)

// writeMarker lays an executable .dynamic-skill in a leaf skill-package dir.
func writeMarker(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".dynamic-skill"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestGenerateDynamicSkills_InvokesWithCwd: the leaf's executable .dynamic-skill
// is run with cwd = its package dir.
func TestGenerateDynamicSkills_InvokesWithCwd(t *testing.T) {
	parent := t.TempDir()
	leaf := filepath.Join(parent, "leaf")
	pkg := filepath.Join(leaf, "construct", "local", "datatype")
	writeMarker(t, pkg)

	layers := []layer.Layer{
		{Name: "leaf", Path: leaf, Intents: skillIntents("construct/local", "construct/adapted")},
	}
	fr := &fakeRunner{}
	if err := generateDynamicSkills(layers, weavefs.OSFS{}, fr); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(fr.dirs) != 1 || fr.dirs[0] != pkg {
		t.Fatalf("ran %v, want [%s] (cwd = package dir)", fr.dirs, pkg)
	}
}

// TestGenerateDynamicSkills_RunnerErrorAborts: a non-zero exit (a Runner error)
// aborts the compile — generate returns the error.
func TestGenerateDynamicSkills_RunnerErrorAborts(t *testing.T) {
	parent := t.TempDir()
	leaf := filepath.Join(parent, "leaf")
	writeMarker(t, filepath.Join(leaf, "construct", "local", "datatype"))

	layers := []layer.Layer{
		{Name: "leaf", Path: leaf, Intents: skillIntents("construct/local")},
	}
	sentinel := errors.New("exit 3")
	fr := &fakeRunner{err: sentinel}
	if err := generateDynamicSkills(layers, weavefs.OSFS{}, fr); !errors.Is(err, sentinel) {
		t.Fatalf("generate err = %v, want it to wrap the Runner error (compile aborts)", err)
	}
}

// TestCompileDryRunSkipsDynamicSkills drives run(..., dryRun=true) over a leaf
// carrying an EXECUTABLE .dynamic-skill that, if exec'd, would create a sentinel
// file in its dir. The dry-run gate must skip the generate stage entirely — so no
// sentinel appears (and crucially no real process is spawned). This is the
// read-only-path exclusion (--dry-run / golden / verify-complete never mutate).
func TestCompileDryRunSkipsDynamicSkills(t *testing.T) {
	derived := buildSkillRepoFixture(t)
	pkg := filepath.Join(derived, "construct", "local", "datatype")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	// A real script that would touch sentinel IF the generate stage ran it.
	script := "#!/bin/sh\ntouch sentinel\n"
	if err := os.WriteFile(filepath.Join(pkg, ".dynamic-skill"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetAll, true, &out); err != nil {
		t.Fatalf("run --dry-run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pkg, "sentinel")); !os.IsNotExist(err) {
		t.Fatalf("dry-run exec'd the .dynamic-skill (sentinel exists, err=%v); read-only paths must not run it", err)
	}
}

// TestCompileRunsDynamicSkills drives run(..., dryRun=false) over a leaf carrying
// an EXECUTABLE .dynamic-skill that touches a sentinel in its dir — the positive
// counterpart to the dry-run test. It exercises the PRODUCTION path end-to-end:
// the real ExecRunner running the RELATIVE "./.dynamic-skill" argv with cwd = the
// package dir (the one path the fake-runner + absolute-/bin/sh unit tests don't
// cover, so a runner/argv refactor can't silently break it with units still green).
func TestCompileRunsDynamicSkills(t *testing.T) {
	derived := buildSkillRepoFixture(t)
	pkg := filepath.Join(derived, "construct", "local", "datatype")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\ntouch sentinel\n"
	if err := os.WriteFile(filepath.Join(pkg, ".dynamic-skill"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetAll, false, &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(pkg, "sentinel")); err != nil {
		t.Fatalf("non-dry-run did NOT exec the .dynamic-skill (sentinel missing, err=%v); the production relative-marker exec path is broken", err)
	}
}

// skillIntents builds `skill <dir>` intents for the leaf's manifest in these
// tests (mirrors walk.skillRows, which is test-only in another package).
func skillIntents(sources ...string) []intent.Intent {
	rows := make([]intent.Intent, len(sources))
	for i, s := range sources {
		rows[i] = intent.Intent{Kind: intent.Skill, Source: s}
	}
	return rows
}
