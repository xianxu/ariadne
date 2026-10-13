package layergraph_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/layergraph"
	"github.com/xianxu/ariadne/pkg/workspace"
)

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=Test", "-c", "user.email=test@example.com", "-c", "core.hooksPath=/dev/null"}, args...)...)
	c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_TERMINAL_PROMPT=0")
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// layer makes dir a layer (a base.manifest) declaring deps, if any.
func layer(t *testing.T, dir, deps string) {
	t.Helper()
	write(t, filepath.Join(dir, "construct", "base.manifest"), "")
	if deps != "" {
		write(t, filepath.Join(dir, "construct", "deps"), deps)
	}
}

type worktreeFleet struct{ fleet, base, leaf, tree string }

// newWorktreeFleet is fleet/{base,leaf}, leaf declaring `substrate ../base`
// and committed, with a linked worktree at fleet/worktree/leaf/br — the
// parley.nvim#254 shape: ../base exists beside the primary, not the worktree.
func newWorktreeFleet(t *testing.T, fleetName string) worktreeFleet {
	t.Helper()
	root, err := workspace.CanonicalPath(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := worktreeFleet{fleet: filepath.Join(root, fleetName)}
	f.base, f.leaf = filepath.Join(f.fleet, "base"), filepath.Join(f.fleet, "leaf")
	f.tree = filepath.Join(f.fleet, "worktree", "leaf", "br")
	layer(t, f.base, "")
	layer(t, f.leaf, "substrate ../base\n")
	gitIn(t, f.leaf, "init", "-b", "main")
	gitIn(t, f.leaf, "add", "-A")
	gitIn(t, f.leaf, "commit", "-m", "init")
	gitIn(t, f.leaf, "worktree", "add", "-b", "br", f.tree)
	return f
}

func walk(t *testing.T, root string) []string {
	t.Helper()
	got, err := layergraph.Walk(layergraph.OSFS{}, root)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestWalkLinkedWorktreeFallsBackToPrimarySibling(t *testing.T) {
	for _, name := range []string{"fleet", "my fleet"} {
		t.Run(name, func(t *testing.T) {
			f := newWorktreeFleet(t, name)
			deps, _ := os.ReadFile(filepath.Join(f.tree, "construct", "deps"))
			if got := walk(t, f.tree); !reflect.DeepEqual(got, []string{f.base, f.tree}) {
				t.Fatalf("Walk = %v", got)
			}
			declared, err := layergraph.DeclaredSubstrates(layergraph.OSFS{}, f.tree)
			if err != nil {
				t.Fatal(err)
			}
			want := []layergraph.DeclaredSubstrate{{Path: f.base, Owner: f.tree, Declared: "../base", Present: true, Via: f.leaf}}
			if !reflect.DeepEqual(declared, want) {
				t.Fatalf("declared %+v, want %+v", declared, want)
			}
			after, _ := os.ReadFile(filepath.Join(f.tree, "construct", "deps"))
			if string(after) != string(deps) {
				t.Fatal("construct/deps rewritten")
			}
			if _, err := os.Lstat(filepath.Join(f.fleet, "worktree", "leaf", "base")); !os.IsNotExist(err) {
				t.Fatalf("resolution created a sibling: %v", err)
			}
		})
	}
}

func TestWalkLinkedWorktreeLocalSiblingWins(t *testing.T) {
	f := newWorktreeFleet(t, "fleet")
	local := filepath.Join(f.fleet, "worktree", "leaf", "base")
	layer(t, local, "")
	if got := walk(t, f.tree); !reflect.DeepEqual(got, []string{local, f.tree}) {
		t.Fatalf("Walk = %v", got)
	}
}

func TestWalkLinkedWorktreeInvalidLocalNotBypassed(t *testing.T) {
	f := newWorktreeFleet(t, "fleet")
	local := filepath.Join(f.fleet, "worktree", "leaf", "base")
	if err := os.MkdirAll(local, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := layergraph.Walk(layergraph.OSFS{}, f.tree)
	var nl *layergraph.NotLayerError
	if !errors.As(err, &nl) || nl.Path != local {
		t.Fatalf("err = %v, want NotLayerError for %s", err, local)
	}
}

func TestWalkLinkedWorktreeInvalidFallbackReported(t *testing.T) {
	f := newWorktreeFleet(t, "fleet")
	if err := os.Remove(filepath.Join(f.base, "construct", "base.manifest")); err != nil {
		t.Fatal(err)
	}
	_, err := layergraph.Walk(layergraph.OSFS{}, f.tree)
	var nl *layergraph.NotLayerError
	if !errors.As(err, &nl) || nl.Path != f.base {
		t.Fatalf("err = %v, want NotLayerError for %s", err, f.base)
	}
}

func TestWalkLinkedWorktreeTransitive(t *testing.T) {
	f := newWorktreeFleet(t, "fleet")
	core := filepath.Join(f.fleet, "core")
	layer(t, core, "")
	write(t, filepath.Join(f.base, "construct", "deps"), "substrate ../core\n")
	if got := walk(t, f.tree); !reflect.DeepEqual(got, []string{core, f.base, f.tree}) {
		t.Fatalf("Walk = %v", got)
	}
}

func TestWalkLinkedWorktreeAbsoluteNeverFallsBack(t *testing.T) {
	f := newWorktreeFleet(t, "fleet")
	// An absolute row naming an absent dir whose relative spelling from the
	// primary would exist: it stays absent.
	write(t, filepath.Join(f.tree, "construct", "deps"), "substrate "+filepath.Join(f.fleet, "worktree", "leaf", "base")+"\n")
	if got := walk(t, f.tree); !reflect.DeepEqual(got, []string{f.tree}) {
		t.Fatalf("Walk = %v", got)
	}
}

func TestWalkLinkedWorktreeThroughSymlink(t *testing.T) {
	f := newWorktreeFleet(t, "fleet")
	link := filepath.Join(filepath.Dir(f.fleet), "link")
	if err := os.Symlink(f.tree, link); err != nil {
		t.Fatal(err)
	}
	if got := walk(t, link); !reflect.DeepEqual(got, []string{f.base, f.tree}) {
		t.Fatalf("Walk = %v", got)
	}
}

func TestWalkNoAnchorKeepsPresentSkip(t *testing.T) {
	f := newWorktreeFleet(t, "fleet")
	if err := os.RemoveAll(f.base); err != nil {
		t.Fatal(err)
	}
	for _, root := range []string{f.leaf, f.tree} {
		if got := walk(t, root); !reflect.DeepEqual(got, []string{root}) {
			t.Fatalf("Walk(%s) = %v", root, got)
		}
	}
	declared, err := layergraph.DeclaredSubstrates(layergraph.OSFS{}, f.leaf)
	if err != nil || len(declared) != 1 || !strings.Contains(declared[0].Unanchored, "primary checkout") {
		t.Fatalf("declared %+v, %v", declared, err)
	}
	declared, err = layergraph.DeclaredSubstrates(layergraph.OSFS{}, f.tree)
	if err != nil || len(declared) != 1 || !strings.Contains(declared[0].Unanchored, "too") {
		t.Fatalf("declared %+v, %v", declared, err)
	}
}

// The parley.nvim#254 regression: from a linked worktree, the vocabulary
// merge cmd/vocabulary performs (and sdlc's merge validation reads) finds the
// substrate's issue schema.
func TestWalkLinkedWorktreeVocabulary(t *testing.T) {
	f := newWorktreeFleet(t, "fleet")
	write(t, filepath.Join(f.base, "construct", "vocabulary", "issue.cue"), "package vocabulary\n")
	var dirs []string
	for _, r := range walk(t, f.tree) {
		dirs = append(dirs, filepath.Join(r, "construct", "vocabulary"))
	}
	got, err := layergraph.MergeByName(layergraph.OSFS{}, dirs, ".cue")
	if err != nil {
		t.Fatal(err)
	}
	if got["issue"] != filepath.Join(f.base, "construct", "vocabulary", "issue.cue") {
		t.Fatalf("merged %v", got)
	}
}
