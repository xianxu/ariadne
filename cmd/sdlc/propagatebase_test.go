package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// orderDependentsFoundationFirst: a dependent that is itself in ANOTHER dependent's
// chain comes first (nous before the brains). Pure.
func TestOrderDependentsFoundationFirst(t *testing.T) {
	deps := []propDep{
		{root: "/ws/brain", chain: []string{"/ws/nous", "/ws/ariadne"}}, // → nous → ariadne
		{root: "/ws/pair", chain: []string{"/ws/ariadne"}},
		{root: "/ws/nous", chain: []string{"/ws/ariadne"}},
	}
	got := orderDependentsFoundationFirst(deps)
	var names []string
	for _, d := range got {
		names = append(names, filepath.Base(d.root))
	}
	// rank 0 (no other dep in chain): nous, pair (tie → path order); rank 1: brain (LAST).
	want := []string{"nous", "pair", "brain"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("order = %v, want %v (foundation-first: nous before brain)", names, want)
	}
}

// workingTreeDirty gates the per-dependent precheck (#109): a clean repo proceeds;
// pre-existing uncommitted work (untracked non-ignored files OR modified tracked
// files) is dirty → skipped. The hinge: gitignored woven output (CLAUDE.md, …) must
// NOT read as dirty, else a previously-propagated clean dependent is falsely skipped.
func TestWorkingTreeDirty(t *testing.T) {
	dir := t.TempDir()
	git(t, "", "init", "-b", "main", dir)
	git(t, dir, "config", "user.email", "t@example.com")
	git(t, dir, "config", "user.name", "t")
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustDirty := func(stage string, want bool) {
		t.Helper()
		got, err := workingTreeDirty(dir)
		if err != nil {
			t.Fatalf("%s: workingTreeDirty err = %v", stage, err)
		}
		if got != want {
			t.Fatalf("%s: dirty = %v, want %v", stage, got, want)
		}
	}

	write("f.txt", "hi\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "init")
	mustDirty("clean repo", false)

	// gitignored woven output present on disk → still CLEAN (the false-skip guard).
	write(".gitignore", "/CLAUDE.md\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "ignore")
	write("CLAUDE.md", "woven prose\n")
	mustDirty("gitignored woven output", false)

	// untracked non-ignored file → DIRTY (a concurrent session's WIP).
	write("wip.md", "scratch\n")
	mustDirty("untracked WIP", true)

	// remove the untracked file, modify a TRACKED file → DIRTY.
	if err := os.Remove(filepath.Join(dir, "wip.md")); err != nil {
		t.Fatal(err)
	}
	write("f.txt", "changed\n")
	mustDirty("modified tracked file", true)
}

// runPropagateBase SKIPS a dependent whose working tree is dirty (#109) — never
// `make weave`s or commits it — and exits non-zero. Hermetic: the skip path is
// reached BEFORE `make weave`, so no real weave binary is needed. Proves the
// end-to-end behavior the incident exposed (a concurrent session's dirty repo).
func TestPropagateBaseSkipsDirtyDependent(t *testing.T) {
	parent := t.TempDir()
	owner := filepath.Join(parent, "owner")
	if err := os.MkdirAll(owner, 0o755); err != nil { // owner just needs to exist (skipped as self)
		t.Fatal(err)
	}
	// A real git dependent: substrate ../owner, Makefile.workflow, an initial commit,
	// then an UNTRACKED file → dirty.
	dep := filepath.Join(parent, "dep")
	git(t, "", "init", "-b", "main", dep)
	git(t, dep, "config", "user.email", "t@example.com")
	git(t, dep, "config", "user.name", "t")
	if err := os.MkdirAll(filepath.Join(dep, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dep, "construct", "deps"), []byte("substrate ../owner\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dep, "Makefile.workflow"), []byte("# weave\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dep, "add", "-A")
	git(t, dep, "commit", "-q", "-m", "init")
	if err := os.WriteFile(filepath.Join(dep, "wip.md"), []byte("concurrent session WIP\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := runPropagateBase(canonRoot(owner), "test#1", false, &buf)
	if err == nil {
		t.Fatalf("expected non-zero (skipped) error, got nil\n%s", buf.String())
	}
	if !strings.Contains(err.Error(), "SKIPPED") {
		t.Fatalf("error = %q, want it to mention SKIPPED", err.Error())
	}
	if !strings.Contains(buf.String(), "SKIPPED: dirty working tree") {
		t.Fatalf("summary missing the SKIPPED line:\n%s", buf.String())
	}
	// Untouched: still exactly the one initial commit (no consume commit), WIP intact.
	if n := git(t, dep, "rev-list", "--count", "HEAD"); n != "1" {
		t.Fatalf("dependent commit count = %s, want 1 (must not commit a dirty dependent)", n)
	}
}

// recursiveDependents walks construct/deps across present siblings: a sibling is a
// dependent iff it's a git repo with a Makefile.workflow whose substrate chain
// transitively includes the owner.
func TestRecursiveDependents(t *testing.T) {
	parent := t.TempDir()
	mk := func(name, deps string, weaveRepo bool) string {
		root := filepath.Join(parent, name)
		if err := os.MkdirAll(filepath.Join(root, "construct"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		if deps != "" {
			if err := os.WriteFile(filepath.Join(root, "construct", "deps"), []byte(deps), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if weaveRepo {
			if err := os.WriteFile(filepath.Join(root, "Makefile.workflow"), []byte("# weave\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return root
	}
	owner := mk("ariadne", "", true)
	mk("nous", "substrate ../ariadne\n", true)            // direct dependent
	mk("brain", "substrate ../nous\n", true)              // transitive (→ nous → ariadne)
	mk("pair", "substrate ../ariadne\n", true)            // direct dependent
	mk("stranger", "substrate ../somewhere-else\n", true) // chain has no ariadne → excluded
	mk("noweave", "substrate ../ariadne\n", false)        // depends but no Makefile.workflow → excluded
	// A sibling that depends on ariadne + has Makefile.workflow but is NOT a git repo
	// (a setup.sh-era scratch dir) → EXCLUDED (can't commit a consumption there).
	scratch := filepath.Join(parent, "scratch")
	if err := os.MkdirAll(filepath.Join(scratch, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scratch, "construct", "deps"), []byte("substrate ../ariadne\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scratch, "Makefile.workflow"), []byte("# weave\n"), 0o644); err != nil {
		t.Fatal(err)
	} // deliberately no .git

	var names []string
	for _, d := range recursiveDependents(owner) {
		names = append(names, filepath.Base(d.root))
	}
	sort.Strings(names)
	want := []string{"brain", "nous", "pair"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("dependents = %v, want %v (transitive incl, non-weave/non-dep excl)", names, want)
	}
}
