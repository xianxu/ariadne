package acquire

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/workspace"
)

// worktreeLeaf is fleet/{base,leaf}, leaf declaring deps and committed, with
// a linked worktree at fleet/worktree/leaf/br (#230; parley.nvim#254 shape).
func worktreeLeaf(t *testing.T, deps string) (fleet, base, leaf, tree string) {
	t.Helper()
	root, err := workspace.CanonicalPath(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fleet = filepath.Join(root, "fleet")
	base, leaf = filepath.Join(fleet, "base"), filepath.Join(fleet, "leaf")
	tree = filepath.Join(fleet, "worktree", "leaf", "br")
	put(t, filepath.Join(base, "construct/base.manifest"), "")
	put(t, filepath.Join(leaf, "construct/base.manifest"), "")
	put(t, filepath.Join(leaf, "construct/deps"), deps)
	gitFixture(t, leaf, "init", "-b", "main")
	gitFixture(t, leaf, "add", ".")
	gitFixture(t, leaf, "commit", "-m", "init")
	gitFixture(t, leaf, "worktree", "add", "-b", "br", tree)
	return
}

func TestRestoreLinkedWorktreeUsesPrimarySibling(t *testing.T) {
	src := origin(t, t.TempDir(), "base", "", true)
	for _, source := range []string{"", src} {
		t.Run("source="+source, func(t *testing.T) {
			fleet, base, leaf, tree := worktreeLeaf(t, strings.TrimSpace("substrate ../base "+source)+"\n")
			if source != "" {
				// The primary's sibling is a checkout of the declared source,
				// which Restore verifies rather than re-cloning.
				if err := os.RemoveAll(base); err != nil {
					t.Fatal(err)
				}
				gitFixture(t, fleet, "clone", "-q", src, base)
			}
			got, err := Restore(context.Background(), tree, false)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Layers, []string{base, tree}) {
				t.Fatalf("layers %v", got.Layers)
			}
			want := []Fallback{{Owner: tree, Path: "../base", Resolved: base, Primary: leaf}}
			if !reflect.DeepEqual(got.Fallbacks, want) {
				t.Fatalf("fallbacks %+v, want %+v", got.Fallbacks, want)
			}
			if _, err := os.Lstat(filepath.Join(fleet, "worktree", "leaf", "base")); !os.IsNotExist(err) {
				t.Fatalf("restore created a duplicate beside the worktree: %v", err)
			}
		})
	}
}

func TestRestoreLinkedWorktreeLocalSiblingWins(t *testing.T) {
	fleet, _, _, tree := worktreeLeaf(t, "substrate ../base\n")
	local := filepath.Join(fleet, "worktree", "leaf", "base")
	put(t, filepath.Join(local, "construct/base.manifest"), "")
	got, err := Restore(context.Background(), tree, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Layers, []string{local, tree}) || len(got.Fallbacks) != 0 {
		t.Fatalf("%+v", got)
	}
}

// Absent beside both the worktree and the primary: refuse with the reason,
// and never clone a duplicate beside the worktree, even with a source.
func TestRestoreLinkedWorktreeMissingEverywhere(t *testing.T) {
	orig := t.TempDir()
	src := origin(t, orig, "base", "", true)
	for _, dryRun := range []bool{false, true} {
		fleet, base, _, tree := worktreeLeaf(t, "substrate ../base "+src+"\n")
		if err := os.RemoveAll(base); err != nil {
			t.Fatal(err)
		}
		_, err := Restore(context.Background(), tree, dryRun)
		if err == nil || !strings.Contains(err.Error(), "absent beside this worktree") || !strings.Contains(err.Error(), "primary checkout") {
			t.Fatalf("dryRun=%v: err = %v", dryRun, err)
		}
		for _, p := range []string{filepath.Join(fleet, "worktree", "leaf", "base"), base} {
			if _, err := os.Lstat(p); !os.IsNotExist(err) {
				t.Fatalf("dryRun=%v: restore created %s", dryRun, p)
			}
		}
	}
}

type fakeAnchors struct {
	primary string
	err     error
}

func (f fakeAnchors) PrimaryCheckout(string) (string, error) { return f.primary, f.err }

// The reason no fallback applied reaches the missing-substrate error
// (close review BR-2).
func TestRestoreMissingNamesNoFallbackReason(t *testing.T) {
	root := t.TempDir()
	leaf := filepath.Join(root, "leaf")
	put(t, filepath.Join(leaf, "construct/deps"), "substrate ../base\n")
	reason := &workspace.NoAnchorError{Dir: leaf, Reason: "belongs to a repository whose primary worktree is bare"}
	_, err := Client{Anchors: fakeAnchors{err: reason}}.Restore(context.Background(), leaf, false)
	if err == nil || !strings.Contains(err.Error(), "record its source") || !strings.Contains(err.Error(), "no worktree fallback: "+reason.Error()) {
		t.Fatalf("err = %v", err)
	}
}

// A failure to find the anchor is not "not a worktree": Restore fails
// instead of cloning beside what may be a worktree.
func TestRestoreAnchorFailureFailsClosed(t *testing.T) {
	root := t.TempDir()
	src := origin(t, filepath.Join(root, "origins"), "base", "", true)
	leaf := filepath.Join(root, "leaf")
	put(t, filepath.Join(leaf, "construct/deps"), "substrate ../base "+src+"\n")
	_, err := Client{Anchors: fakeAnchors{err: errors.New("git unavailable")}}.Restore(context.Background(), leaf, false)
	if err == nil || !strings.Contains(err.Error(), "git unavailable") {
		t.Fatalf("err = %v", err)
	}
	if _, err := os.Lstat(filepath.Join(root, "base")); !os.IsNotExist(err) {
		t.Fatalf("cloned despite the anchor failure: %v", err)
	}
}
