package plan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// Apply executes []Action against an FS, idempotently, ported from setup.sh's
// create_symlink / create_scaffold / WriteFile behaviors. It is the only
// mutating code; we test it against a real t.TempDir()-rooted OSFS (the seam
// exercised end-to-end, no mocks — ARCH: faithful over mocked).

func TestApplyWriteFile(t *testing.T) {
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		WriteFile{Path: "AGENTS.md", Content: "BASE\n\nLOCAL\n"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(got) != "BASE\n\nLOCAL\n" {
		t.Fatalf("AGENTS.md = %q, want %q", got, "BASE\n\nLOCAL\n")
	}
}

func TestApplyWriteFileCreatesParents(t *testing.T) {
	// WriteFile to a nested path creates parents (ensure_parent / mkdir -p).
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		WriteFile{Path: "workshop/lessons.md", Content: ""},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "workshop", "lessons.md")); err != nil {
		t.Fatalf("nested file not created: %v", err)
	}
}

func TestApplyTouchCreatesWhenMissing(t *testing.T) {
	// Touch creates an empty file (with parents) when none exists.
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		Touch{Path: "workshop/lessons.md"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	fi, err := os.Stat(filepath.Join(root, "workshop", "lessons.md"))
	if err != nil {
		t.Fatalf("touched file not created: %v", err)
	}
	if fi.Size() != 0 {
		t.Fatalf("touched file size = %d, want 0", fi.Size())
	}
}

func TestApplyTouchDoesNotClobberExisting(t *testing.T) {
	// The golden-diff finding: Touch must NOT overwrite an existing,
	// content-bearing file (workshop/lessons.md accumulates lessons over time).
	root := t.TempDir()
	path := filepath.Join(root, "workshop", "lessons.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("ACCUMULATED LESSONS"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{
		Touch{Path: "workshop/lessons.md"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "ACCUMULATED LESSONS" {
		t.Fatalf("Touch clobbered existing content: got %q, want ACCUMULATED LESSONS", got)
	}
}

func TestApplyMkdir(t *testing.T) {
	root := t.TempDir()
	if err := Apply(weavefs.OSFS{}, root, []Action{
		Mkdir{Path: ".claude/skills"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	info, err := os.Stat(filepath.Join(root, ".claude", "skills"))
	if err != nil || !info.IsDir() {
		t.Fatalf("scaffold dir not created: err=%v info=%v", err, info)
	}
}

func TestApplySymlinkRelativeToTarget(t *testing.T) {
	// Symlink{Src absolute upstream, Dst repo-relative} → a RELATIVE symlink
	// from the target dir to the upstream file (ports create_symlink's
	// rel_path(src, dirname(dst))). We point Src at a real upstream file.
	root := t.TempDir()
	upstream := t.TempDir()
	srcAbs := filepath.Join(upstream, "AGENTS.md")
	if err := os.WriteFile(srcAbs, []byte("UPSTREAM"), 0o644); err != nil {
		t.Fatalf("seed upstream: %v", err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{
		Symlink{Src: srcAbs, Dst: "AGENTS.md"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	dst := filepath.Join(root, "AGENTS.md")
	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	// link target must be RELATIVE (not absolute) — survives a repo move.
	if filepath.IsAbs(target) {
		t.Fatalf("symlink target %q is absolute, want relative", target)
	}
	// and must resolve to the upstream content.
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read through symlink: %v", err)
	}
	if string(got) != "UPSTREAM" {
		t.Fatalf("symlink resolves to %q, want UPSTREAM", got)
	}
}

func TestApplySymlinkReplacesExistingSymlink(t *testing.T) {
	// Idempotency: a re-run with a different upstream replaces a stale symlink
	// (ports create_symlink's [[ -L ]] → rm + relink). Also: re-applying the
	// SAME link is a no-op that still leaves a correct link.
	root := t.TempDir()
	upstream := t.TempDir()
	srcA := filepath.Join(upstream, "a.md")
	srcB := filepath.Join(upstream, "b.md")
	if err := os.WriteFile(srcA, []byte("A"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcB, []byte("B"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "doc.md")

	if err := Apply(weavefs.OSFS{}, root, []Action{Symlink{Src: srcA, Dst: "doc.md"}}); err != nil {
		t.Fatalf("Apply A: %v", err)
	}
	// Re-apply pointing elsewhere — must replace, not error.
	if err := Apply(weavefs.OSFS{}, root, []Action{Symlink{Src: srcB, Dst: "doc.md"}}); err != nil {
		t.Fatalf("Apply B: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "B" {
		t.Fatalf("symlink resolves to %q after replace, want B", got)
	}
}

func TestApplySymlinkReplacesExistingRegularFile(t *testing.T) {
	// A regular file/dir occupying the slot is removed and relinked
	// (create_symlink's [[ -e ]] → rm -rf).
	root := t.TempDir()
	upstream := t.TempDir()
	src := filepath.Join(upstream, "x.md")
	if err := os.WriteFile(src, []byte("UP"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(root, "x.md")
	if err := os.WriteFile(dst, []byte("STALE REGULAR FILE"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Apply(weavefs.OSFS{}, root, []Action{Symlink{Src: src, Dst: "x.md"}}); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if fi, err := os.Lstat(dst); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("dst is not a symlink after relink: err=%v mode=%v", err, fi.Mode())
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "UP" {
		t.Fatalf("relinked content = %q, want UP", got)
	}
}

// fakeGoMod records AddTool calls instead of shelling out to the toolchain — the
// exec seam's fake (ARCH-PURE: the planner stays pure; Apply's go.mod mutation
// is the only exec, isolated behind GoModEditor and verifiable without `go`).
type fakeGoMod struct {
	calls []string // each "gomodPath -tool toolImportPath"
	err   error
}

func (f *fakeGoMod) AddTool(gomodPath, toolImportPath string) error {
	f.calls = append(f.calls, gomodPath+" -tool "+toolImportPath)
	return f.err
}

func TestApplyToolDerivativeAppendsSubstrate(t *testing.T) {
	// Derivative mode (Owner != repoRoot): ApplyWith appends
	// `substrate <relpath-to-owner>` to the repo's construct/deps, creating the
	// file when absent. The relpath is repo-root-relative (../ariadne), matching
	// ensure_go_tool_dependency's cross-target branch.
	root := t.TempDir()
	owner := filepath.Join(filepath.Dir(root), "ariadne")
	ed := &fakeGoMod{}
	if err := ApplyWith(weavefs.OSFS{}, root, ed, []Action{
		ToolDep{Owner: owner, Path: "cmd/sdlc"},
	}); err != nil {
		t.Fatalf("ApplyWith: %v", err)
	}
	if len(ed.calls) != 0 {
		t.Fatalf("derivative must not touch go.mod, got calls %v", ed.calls)
	}
	deps, err := os.ReadFile(filepath.Join(root, "construct", "deps"))
	if err != nil {
		t.Fatalf("read construct/deps: %v", err)
	}
	want := "substrate ../ariadne\n"
	if string(deps) != want {
		t.Fatalf("construct/deps = %q, want %q", deps, want)
	}
}

func TestApplyToolDerivativeIdempotent(t *testing.T) {
	// A second ApplyWith must NOT re-append the substrate row (ParseDeps detects
	// it is already present), and must preserve existing content. Matches
	// ensure_go_tool_dependency's awk present-check.
	root := t.TempDir()
	owner := filepath.Join(filepath.Dir(root), "ariadne")
	depsPath := filepath.Join(root, "construct", "deps")
	if err := os.MkdirAll(filepath.Dir(depsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	// Pre-existing content: a data row + the already-present substrate row.
	if err := os.WriteFile(depsPath, []byte("data ../some-data git@x\nsubstrate ../ariadne\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ed := &fakeGoMod{}
	if err := ApplyWith(weavefs.OSFS{}, root, ed, []Action{
		ToolDep{Owner: owner, Path: "cmd/sdlc"},
	}); err != nil {
		t.Fatalf("ApplyWith: %v", err)
	}
	deps, err := os.ReadFile(depsPath)
	if err != nil {
		t.Fatalf("read construct/deps: %v", err)
	}
	want := "data ../some-data git@x\nsubstrate ../ariadne\n"
	if string(deps) != want {
		t.Fatalf("construct/deps = %q, want unchanged %q", deps, want)
	}
}

func TestApplyToolOwnerEditsGoMod(t *testing.T) {
	// Owner self-walk (Owner == repoRoot): ApplyWith reads the repo's go.mod
	// module line and asks the GoModEditor to add a `tool <module>/<path>`
	// directive. construct/deps is NOT touched. Matches the self-walk branch.
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"),
		[]byte("module github.com/xianxu/ariadne\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ed := &fakeGoMod{}
	if err := ApplyWith(weavefs.OSFS{}, root, ed, []Action{
		ToolDep{Owner: root, Path: "cmd/sdlc"},
	}); err != nil {
		t.Fatalf("ApplyWith: %v", err)
	}
	wantCall := filepath.Join(root, "go.mod") + " -tool github.com/xianxu/ariadne/cmd/sdlc"
	if len(ed.calls) != 1 || ed.calls[0] != wantCall {
		t.Fatalf("editor calls = %v, want exactly [%q]", ed.calls, wantCall)
	}
	if _, err := os.Stat(filepath.Join(root, "construct", "deps")); !os.IsNotExist(err) {
		t.Fatalf("owner self-walk must not create construct/deps (err=%v)", err)
	}
}
