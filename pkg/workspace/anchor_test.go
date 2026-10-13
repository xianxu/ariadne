package workspace_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/workspace"
)

func TestDependencyAnchor(t *testing.T) {
	f := setup(t)
	feature := filepath.Join(f.root, "worktree", "repo", "br")
	git(t, f.primary, "worktree", "add", "-b", "br", feature)
	if err := os.Mkdir(filepath.Join(f.primary, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	plain := filepath.Join(f.root, "plain")
	if err := os.Mkdir(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	bare := filepath.Join(f.root, "b.git")
	git(t, f.root, "init", "--bare", "-b", "main", bare)
	git(t, f.primary, "push", bare, "main")
	bareTree := filepath.Join(f.root, "barewt")
	git(t, bare, "worktree", "add", bareTree, "main")

	for _, c := range []struct{ name, dir, want, err string }{
		{"linked worktree", feature, f.primary, ""},
		{"ordinary linked worktree", f.ordinary, f.primary, ""},
		{"primary", f.primary, "", "primary checkout"},
		{"subdirectory", filepath.Join(f.primary, "sub"), "", "not a Git checkout root"},
		{"non-Git", plain, "", "not a Git checkout root"},
		{"bare primary", bareTree, "", "bare"},
		{"numbered environment host", f.slot, "", "numbered environment"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got, err := workspace.DependencyAnchor(f.real, c.dir)
			if c.err == "" {
				if err != nil || got != c.want {
					t.Fatalf("got %q, %v; want %q", got, err, c.want)
				}
				return
			}
			var none *workspace.NoAnchorError
			if !errors.As(err, &none) || !strings.Contains(err.Error(), c.err) || got != "" {
				t.Fatalf("got %q, %v; want NoAnchorError containing %q", got, err, c.err)
			}
		})
	}
}

// A layer that is not a checkout root never runs Git (an enclosing
// repository must not lend its primary), nor does one with a .git directory
// (its own primary checkout).
func TestDependencyAnchorSkipsGitOutsideCheckouts(t *testing.T) {
	f := setup(t)
	for _, dir := range []string{filepath.Join(f.root, "absent"), f.primary} {
		f.real.reads = 0
		if _, err := workspace.DependencyAnchor(f.real, dir); err == nil {
			t.Fatalf("%s anchored", dir)
		}
		if f.real.reads != 0 {
			t.Fatalf("%s: ran Git %d times", dir, f.real.reads)
		}
	}
}

func TestDependencyAnchorSpacesAndSymlink(t *testing.T) {
	root, err := workspace.CanonicalPath(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fleet := filepath.Join(root, "my fleet")
	primary := filepath.Join(fleet, "leaf")
	if err := os.MkdirAll(primary, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, primary, "init", "-b", "main")
	git(t, primary, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "initial")
	tree := filepath.Join(fleet, "worktree", "leaf", "my br")
	git(t, primary, "worktree", "add", "-b", "br", tree)
	link := filepath.Join(root, "link")
	if err := os.Symlink(tree, link); err != nil {
		t.Fatal(err)
	}
	got, err := workspace.DependencyAnchor(&realGit{}, link)
	if err != nil || got != primary {
		t.Fatalf("got %q, %v; want %q", got, err, primary)
	}
}

// A feature worktree of a numbered environment's own dependency clone falls
// back inside the environment (#230 Revisions).
func TestDependencyAnchorInsideEnvironment(t *testing.T) {
	f := setup(t)
	env := filepath.Dir(f.slot)
	dep := filepath.Join(env, "dep")
	if err := os.Mkdir(dep, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, dep, "init", "-b", "main")
	git(t, dep, "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--allow-empty", "-m", "initial")
	feature := filepath.Join(env, ".worktrees", "dep", "br")
	git(t, dep, "worktree", "add", "-b", "br", feature)
	got, err := workspace.DependencyAnchor(f.real, feature)
	if err != nil || got != dep {
		t.Fatalf("got %q, %v; want %q", got, err, dep)
	}
}

// A Git failure is not an expected absence: the caller must not mistake it
// for "no anchor" and act as if the layer were not a worktree.
func TestDependencyAnchorGitFailureIsNotNoAnchor(t *testing.T) {
	f := setup(t)
	feature := filepath.Join(f.root, "worktree", "repo", "br")
	git(t, f.primary, "worktree", "add", "-b", "br", feature)
	_, err := workspace.DependencyAnchor(failingGit{}, feature)
	var none *workspace.NoAnchorError
	if err == nil || errors.As(err, &none) {
		t.Fatalf("err = %v, want a Git failure", err)
	}
}

type failingGit struct{}

func (failingGit) GitInDir(string, ...string) ([]byte, error) {
	return nil, errors.New("git unavailable")
}
