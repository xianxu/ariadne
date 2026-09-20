package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
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

// The synchronous stateful fake has no background producer lifetime.
func (f *fakeRunner) RunOwned(dir string, argv []string, stage string) error { return f.Run(dir, argv) }

// writeMarker lays an executable .dynamic-skill in a skill-package dir.
func writeMarker(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".dynamic-skill"), []byte("#!/bin/sh\n# weave-output: argv1\n"), 0o755); err != nil {
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
	dyns, derr := walk.DynamicSkills(weavefs.OSFS{}, layers)
	if derr != nil {
		t.Fatalf("select: %v", derr)
	}
	if err := generateDynamicSkills(weavefs.OSFS{}, dyns, layers[len(layers)-1].Path, filepath.Join(parent, "stage"), fr); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(fr.calls) != 1 {
		t.Fatalf("ran %d markers, want 1: %v", len(fr.calls), fr.calls)
	}
	c := fr.calls[0]
	if c.dir != leaf {
		t.Errorf("cwd = %q, want the DERIVATIVE root %q (leaf-rooted output)", c.dir, leaf)
	}
	wantArgv := []string{"sh", ancMarker, filepath.Join(parent, "stage", "datatype")}
	if len(c.argv) != 3 || c.argv[0] != wantArgv[0] || c.argv[1] != wantArgv[1] || c.argv[2] != wantArgv[2] {
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
	dyns, derr := walk.DynamicSkills(weavefs.OSFS{}, layers)
	if derr != nil {
		t.Fatalf("select: %v", derr)
	}
	if err := generateDynamicSkills(weavefs.OSFS{}, dyns, layers[len(layers)-1].Path, filepath.Join(parent, "stage"), fr); !errors.Is(err, sentinel) {
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
	script := "#!/bin/sh\n# weave-output: argv1\ntouch sentinel\nmkdir -p \"$1\"\nprintf body > \"$1/SKILL.md\"\n"
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
	script := "#!/bin/sh\n# weave-output: argv1\nout=\"${1:-" + outDir + "}\"\nmkdir -p \"$out\"\n" +
		"printf '%s\\n' '---' 'name: xx-datatype' 'description: generated' '---' '' 'BODY' > \"$out/SKILL.md\"\n"

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

// Unknown generated-looking paths have no ownership evidence and must survive.
func TestCompilePreservesUnownedGeneratedDir(t *testing.T) {
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
	if got, err := os.ReadFile(filepath.Join(goneDir, "SKILL.md")); err != nil || string(got) != "stale" {
		t.Errorf("unowned generated file changed: %q, %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(derived, "construct", "generated", "datatype", "SKILL.md")); err != nil {
		t.Errorf("in-use construct/generated/datatype was destroyed: %v", err)
	}
}

func TestCompilePreservesEditedGeneratorDestination(t *testing.T) {
	for _, replacement := range []string{"file", "link"} {
		t.Run(replacement, func(t *testing.T) {
			leaf := buildSkillRepoFixture(t)
			base := filepath.Join(filepath.Dir(leaf), "base")
			realDatatypeMarker(t, filepath.Join(base, "construct/local/datatype"), "construct/generated/datatype")
			var out bytes.Buffer
			if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err != nil {
				t.Fatal(err)
			}
			body := filepath.Join(leaf, "construct/generated/datatype/SKILL.md")
			if replacement == "link" {
				authored := filepath.Join(leaf, "authored.md")
				mkfile(t, authored, "authored edit\n")
				if err := os.Remove(body); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(authored, body); err != nil {
					t.Fatal(err)
				}
			} else {
				mkfile(t, body, "authored edit\n")
			}
			if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err == nil {
				t.Fatal("edited generated destination accepted")
			}
			if got, err := os.ReadFile(body); err != nil || string(got) != "authored edit\n" {
				t.Fatalf("edit overwritten: %q %v", got, err)
			}
		})
	}
}

func TestCompileFailedGeneratorDoesNotPublishAndRetryRetires(t *testing.T) {
	leaf := buildSkillRepoFixture(t)
	base := filepath.Join(filepath.Dir(leaf), "base")
	pkg := filepath.Join(base, "construct/local/datatype")
	realDatatypeMarker(t, pkg, "construct/generated/datatype")
	marker := filepath.Join(pkg, ".dynamic-skill")
	good, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, append(append([]byte{}, good...), []byte("exit 1\n")...), 0755); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err == nil {
		t.Fatal("failed generator accepted")
	}
	body := filepath.Join(leaf, "construct/generated/datatype/SKILL.md")
	if _, err := os.Lstat(body); !os.IsNotExist(err) {
		t.Fatalf("failed generator published output: %v", err)
	}
	if err := os.WriteFile(marker, good, 0755); err != nil {
		t.Fatal(err)
	}
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(body); err != nil {
		t.Fatal("retry missing output:", err)
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(body); !os.IsNotExist(err) {
		t.Fatalf("retry output not retired: %v", err)
	}
}

func TestCompileRejectsLegacyMarkerBeforeExecution(t *testing.T) {
	leaf := buildSkillRepoFixture(t)
	marker := filepath.Join(leaf, "construct/local/legacy/.dynamic-skill")
	mkfile(t, marker, "#!/bin/sh\ntouch unsafe-sentinel\n")
	if err := os.Chmod(marker, 0755); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err == nil || !strings.Contains(err.Error(), "weave-output: argv1") {
		t.Fatalf("want marker contract error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(leaf, "unsafe-sentinel")); !os.IsNotExist(err) {
		t.Fatalf("legacy marker executed: %v", err)
	}
}

func TestCompilePreservesColdAuthoredGeneratorOutput(t *testing.T) {
	leaf := buildSkillRepoFixture(t)
	base := filepath.Join(filepath.Dir(leaf), "base")
	realDatatypeMarker(t, filepath.Join(base, "construct/local/datatype"), "construct/generated/datatype")
	body := filepath.Join(leaf, "construct/generated/datatype/SKILL.md")
	mkfile(t, body, "authored before first compile\n")
	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err == nil {
		t.Fatal("cold authored output accepted")
	}
	if got, err := os.ReadFile(body); err != nil || string(got) != "authored before first compile\n" {
		t.Fatalf("authored output overwritten: %q %v", got, err)
	}
}

func TestCompileKilledGeneratorHelper(t *testing.T) {
	leaf := os.Getenv("WEAVE_KILLED_GENERATOR_FIXTURE")
	if leaf == "" {
		t.Skip("subprocess helper")
	}
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, os.Stdout); err != nil {
		t.Fatal(err)
	}
}

func TestCompileRecoversAfterProcessDeathDuringGeneration(t *testing.T) {
	leaf := buildSkillRepoFixture(t)
	base := filepath.Join(filepath.Dir(leaf), "base")
	pkg := filepath.Join(base, "construct/local/datatype")
	realDatatypeMarker(t, pkg, "construct/generated/datatype")
	marker := filepath.Join(pkg, ".dynamic-skill")
	good, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	killed := append(append([]byte{}, good...), []byte("kill -KILL \"$PPID\"\nexit 0\n")...)
	if err := os.WriteFile(marker, killed, 0755); err != nil {
		t.Fatal(err)
	}
	child := exec.Command(os.Args[0], "-test.run=^TestCompileKilledGeneratorHelper$")
	child.Env = append(os.Environ(), "WEAVE_KILLED_GENERATOR_FIXTURE="+leaf)
	if output, err := child.CombinedOutput(); err == nil {
		t.Fatalf("helper not killed: %s", output)
	}
	body := filepath.Join(leaf, "construct/generated/datatype/SKILL.md")
	if _, err := os.Lstat(body); !os.IsNotExist(err) {
		t.Fatalf("dead generator published: %v", err)
	}
	if err := os.WriteFile(marker, good, 0755); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(leaf, "construct/generated/weave"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("owned stage survived retry: %s", entry.Name())
		}
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(body); !os.IsNotExist(err) {
		t.Fatalf("recovered output did not retire: %v", err)
	}
}

func TestCompilePreservesGeneratedExecutablePermissions(t *testing.T) {
	leaf := buildSkillRepoFixture(t)
	base := filepath.Join(filepath.Dir(leaf), "base")
	pkg := filepath.Join(base, "construct/local/datatype")
	realDatatypeMarker(t, pkg, "construct/generated/datatype")
	marker := filepath.Join(pkg, ".dynamic-skill")
	original, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	for _, mode := range []string{"755", "644"} {
		script := string(original) + "printf '#!/bin/sh\\necho generated\\n' > \"$out/run.sh\"\nchmod " + mode + " \"$out/run.sh\"\n"
		if err := os.WriteFile(marker, []byte(script), 0755); err != nil {
			t.Fatal(err)
		}
		if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, &out); err != nil {
			t.Fatal(err)
		}
		info, err := os.Stat(filepath.Join(leaf, "construct/generated/datatype/run.sh"))
		if err != nil {
			t.Fatal(err)
		}
		want := os.FileMode(0755)
		if mode == "644" {
			want = 0644
		}
		if info.Mode().Perm() != want {
			t.Fatalf("generated file mode %o, want %o", info.Mode().Perm(), want)
		}
	}
}
