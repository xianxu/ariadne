package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// runCall records one Runner.Run invocation: its cwd and argv.
type runCall struct {
	dir  string
	argv []string
}

// fakeRunner records each Run call and returns a canned error — so the generate
// stage is tested with NO real binary spawned (the production ExecRunner is
// integration-tested in weavefs/runner_test.go).
type fakeRunner struct {
	calls []runCall
	err   error
}

func (f *fakeRunner) Run(dir string, argv []string) error {
	f.calls = append(f.calls, runCall{dir: dir, argv: argv})
	return f.err
}

var _ weavefs.Runner = (*fakeRunner)(nil)

// writeMarker lays an executable .dynamic-skill in a skill-package dir.
func writeMarker(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".dynamic-skill"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// writeConfig writes a layer's construct/config.json localPrefix (so DynamicSkills'
// prefix resolution is deterministic in these cmd/weave tests).
func writeConfig(t *testing.T, root, prefix string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "construct", "config.json"),
		[]byte(`{"localPrefix": "`+prefix+`"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestGenerateDynamicSkills_AncestorMarkerInvokedWithLeafCwd: compiling a
// DERIVATIVE whose dynamic skill (datatype) is owned by an ANCESTOR invokes the
// ancestor's marker with cwd = the DERIVATIVE's root (leaf-rooted output), and the
// ancestor's tree is untouched (no marker run with the ancestor as cwd).
func TestGenerateDynamicSkills_AncestorMarkerInvokedWithLeafCwd(t *testing.T) {
	parent := t.TempDir()
	ancestor := filepath.Join(parent, "ancestor")
	leaf := filepath.Join(parent, "leaf")
	writeConfig(t, ancestor, "xx-")
	writeConfig(t, leaf, "xx-")
	ancMarker := filepath.Join(ancestor, "construct", "local", "datatype", ".dynamic-skill")
	writeMarker(t, filepath.Dir(ancMarker))

	layers := []layer.Layer{
		{Name: "ancestor", Path: ancestor, Intents: skillIntents("construct/local")},
		{Name: "leaf", Path: leaf, Intents: skillIntents("construct/local", "construct/adapted")},
	}
	fr := &fakeRunner{}
	if err := generateDynamicSkills(layers, weavefs.OSFS{}, fr); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(fr.calls) != 1 {
		t.Fatalf("ran %d markers, want 1: %v", len(fr.calls), fr.calls)
	}
	c := fr.calls[0]
	if c.dir != leaf {
		t.Errorf("cwd = %q, want the DERIVATIVE root %q (leaf-rooted output)", c.dir, leaf)
	}
	wantArgv := []string{"sh", ancMarker}
	if len(c.argv) != 2 || c.argv[0] != wantArgv[0] || c.argv[1] != wantArgv[1] {
		t.Errorf("argv = %v, want %v (the ANCESTOR's marker, run via sh)", c.argv, wantArgv)
	}
}

// TestGenerateDynamicSkills_RunnerErrorAborts: a non-zero exit (a Runner error)
// aborts the compile — generate returns the error.
func TestGenerateDynamicSkills_RunnerErrorAborts(t *testing.T) {
	parent := t.TempDir()
	leaf := filepath.Join(parent, "leaf")
	writeConfig(t, leaf, "xx-")
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
// at the repo ROOT (its cwd). The dry-run gate must skip the generate stage
// entirely — so no sentinel appears (and crucially no real process is spawned).
// This is the read-only-path exclusion (--dry-run / golden / verify-complete never
// mutate).
func TestCompileDryRunSkipsDynamicSkills(t *testing.T) {
	derived := buildSkillRepoFixture(t)
	pkg := filepath.Join(derived, "construct", "local", "datatype")
	if err := os.MkdirAll(pkg, 0o755); err != nil {
		t.Fatal(err)
	}
	// A real script that would touch sentinel (at cwd = the repo root) IF the
	// generate stage ran it.
	script := "#!/bin/sh\ntouch sentinel\n"
	if err := os.WriteFile(filepath.Join(pkg, ".dynamic-skill"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetAll, true, &out); err != nil {
		t.Fatalf("run --dry-run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(derived, "sentinel")); !os.IsNotExist(err) {
		t.Fatalf("dry-run exec'd the .dynamic-skill (sentinel exists, err=%v); read-only paths must not run it", err)
	}
}

// TestCompileRunsDynamicSkills drives run(..., dryRun=false) over a leaf carrying
// an EXECUTABLE .dynamic-skill that touches a sentinel at the repo ROOT — the
// positive counterpart to the dry-run test. It exercises the PRODUCTION path
// end-to-end: the real ExecRunner running `sh <markerPath>` with cwd = R's ROOT
// (#115 M3 leaf-rooted output — the one path the fake-runner unit tests don't
// cover, so a runner/argv refactor can't silently break it with units still green).
// The sentinel lands at the repo root precisely BECAUSE cwd = root now.
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
	if _, err := os.Stat(filepath.Join(derived, "sentinel")); err != nil {
		t.Fatalf("non-dry-run did NOT exec the .dynamic-skill at the repo root (sentinel missing, err=%v); the production leaf-rooted marker exec path is broken", err)
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

// realDatatypeMarker writes an executable .dynamic-skill that, like the production
// marker, writes a SKILL.md into construct/generated/<dir> RELATIVE to cwd (=the
// compiling repo's root) — a self-contained stand-in for the datatype binary so the
// e2e tests don't depend on it being on PATH.
func realDatatypeMarker(t *testing.T, pkgDir, outDir string) {
	t.Helper()
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nmkdir -p " + outDir + "\n" +
		"printf '%s\\n' '---' 'name: xx-datatype' 'description: generated' '---' '' 'BODY' > " + outDir + "/SKILL.md\n"
	if err := os.WriteFile(filepath.Join(pkgDir, ".dynamic-skill"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestCompileDerivativeMaterializesAndVerifyCompleteGreen (#115 M3, plan-review I6):
// compiling a DERIVATIVE whose datatype marker is owned by the BASE layer
// materializes the body under the DERIVATIVE's construct/generated (leaf-rooted
// output), lowers exactly ONE xx-datatype skill (NO derived-datatype duplicate, the
// C1 gate), and the read-only `verify-complete` stays GREEN post-compile.
func TestCompileDerivativeMaterializesAndVerifyCompleteGreen(t *testing.T) {
	derived := buildSkillRepoFixture(t)
	// buildSkillRepoFixture's base lives at ../base relative to derived. Add the
	// datatype marker to the BASE's construct/local (the owner), writing to
	// construct/generated/datatype relative to cwd (the derivative root at compile).
	base := filepath.Join(filepath.Dir(derived), "base")
	realDatatypeMarker(t, filepath.Join(base, "construct", "local", "datatype"), "construct/generated/datatype")

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetAll, false, &out); err != nil {
		t.Fatalf("compile derivative: %v\n%s", err, out.String())
	}

	// The body materialized under the DERIVATIVE (not the base).
	if _, err := os.Stat(filepath.Join(derived, "construct", "generated", "datatype", "SKILL.md")); err != nil {
		t.Fatalf("derivative did not materialize its own construct/generated/datatype/SKILL.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "construct", "generated", "datatype", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("the BASE (ancestor) tree was mutated by the derivative compile (err=%v); leaf-rooted output violated", err)
	}

	// Exactly one xx-datatype lowered, no derived-datatype duplicate (C1 gate).
	skillsDir := filepath.Join(derived, ".claude", "skills")
	var datatypeLinks []string
	entries, _ := os.ReadDir(skillsDir)
	for _, e := range entries {
		if strings.Contains(e.Name(), "datatype") {
			datatypeLinks = append(datatypeLinks, e.Name())
		}
	}
	if len(datatypeLinks) != 1 || datatypeLinks[0] != "xx-datatype" {
		t.Fatalf(".claude/skills datatype links = %v, want exactly [xx-datatype] (no derived-datatype — C1)", datatypeLinks)
	}

	// verify-complete (read-only) stays green post-compile.
	var vout bytes.Buffer
	if err := runVerifyComplete(weavefs.OSFS{}, derived, []string{derived}, plan.TargetAll, &vout); err != nil {
		t.Fatalf("verify-complete reported under-production post-compile: %v\n%s", err, vout.String())
	}
}

// TestCompilePrunesOrphanedGeneratedDir (#115 M3): a stale construct/generated/gone
// (no longer produced by any marker) is GC'd by the compile, while the in-use
// construct/generated/datatype survives.
func TestCompilePrunesOrphanedGeneratedDir(t *testing.T) {
	derived := buildSkillRepoFixture(t)
	base := filepath.Join(filepath.Dir(derived), "base")
	realDatatypeMarker(t, filepath.Join(base, "construct", "local", "datatype"), "construct/generated/datatype")
	// Pre-seed an ORPHAN generated dir (no marker produces "gone").
	goneDir := filepath.Join(derived, "construct", "generated", "gone")
	if err := os.MkdirAll(goneDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(goneDir, "SKILL.md"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, derived, plan.TargetAll, false, &out); err != nil {
		t.Fatalf("compile: %v\n%s", err, out.String())
	}
	if _, err := os.Stat(goneDir); !os.IsNotExist(err) {
		t.Errorf("orphan construct/generated/gone survived the compile (err=%v); generated-class GC failed", err)
	}
	if _, err := os.Stat(filepath.Join(derived, "construct", "generated", "datatype", "SKILL.md")); err != nil {
		t.Errorf("in-use construct/generated/datatype was destroyed: %v", err)
	}
}
