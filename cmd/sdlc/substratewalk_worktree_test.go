package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xianxu/ariadne/pkg/layergraph"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// From a linked worktree, sdlc's substrate walk resolves ../A beside the
// worktree's primary checkout, as weave and vocabulary do (#230).
func TestSubstrateWalkLinkedWorktree(t *testing.T) {
	parent, err := workspace.CanonicalPath(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a, b := filepath.Join(parent, "A"), filepath.Join(parent, "B")
	for dir, deps := range map[string]string{a: "", b: "substrate ../A\n"} {
		if err := os.MkdirAll(filepath.Join(dir, "construct"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "construct", "deps"), []byte(deps), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tree := filepath.Join(parent, "worktree", "B", "br")
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"add", "-f", "."}, {"commit", "-qm", "b"}, {"worktree", "add", "-qb", "br", tree}} {
		c := exec.Command("git", append([]string{"-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "core.hooksPath=/dev/null"}, args...)...)
		c.Dir = b
		c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	order, absent := substrateWalk(tree)
	if !reflect.DeepEqual(order, []string{a}) || len(absent) != 0 {
		t.Fatalf("substrateWalk = %v, absent %v", order, absent)
	}
}

// failingAnchorsFS is an FS whose Anchors cannot answer (Git failed).
type failingAnchorsFS struct{ layergraph.OSFS }

func (failingAnchorsFS) PrimaryCheckout(string) (string, error) {
	return "", errors.New("git unavailable")
}

// When Git cannot say whether a layer is a worktree, the row is reported
// absent, once, under its local spelling — never resolved by a guess.
func TestSubstrateWalkAnchorFailureReportsAbsent(t *testing.T) {
	parent, err := workspace.CanonicalPath(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	leaf := filepath.Join(parent, "leaf")
	if err := os.MkdirAll(filepath.Join(leaf, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(leaf, "construct", "deps"), []byte("substrate ../base\nsubstrate ../base\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	order, absent := substrateWalkWith(leaf, layergraph.NewResolver(failingAnchorsFS{}))
	if len(order) != 0 || !reflect.DeepEqual(absent, []string{filepath.Join(parent, "base")}) {
		t.Fatalf("order %v, absent %v", order, absent)
	}
}
