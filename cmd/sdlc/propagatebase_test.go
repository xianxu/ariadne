package main

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
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
