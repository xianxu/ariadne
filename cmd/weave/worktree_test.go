package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// worktreeFixture is buildFixture's base + derived with derived committed and
// a linked worktree at <parent>/worktree/derived/br: ../base exists beside the
// primary checkout, not beside the worktree (parley.nvim#254, #230).
func worktreeFixture(t *testing.T) (parent, base, derived, tree string) {
	t.Helper()
	parent, base, derived = buildFixture(t)
	setupGit(t, derived, "init", "-q", "-b", "main")
	setupGit(t, derived, "add", "-f", ".") // -f: a global excludes file must not drop fixture prose
	setupGit(t, derived, "commit", "-qm", "derived")
	tree = filepath.Join(parent, "worktree", "derived", "br")
	setupGit(t, derived, "worktree", "add", "-qb", "br", tree)
	for _, p := range []*string{&parent, &base, &derived, &tree} {
		c, err := workspace.CanonicalPath(*p)
		if err != nil {
			t.Fatal(err)
		}
		*p = c
	}
	return
}

func TestCompileLinkedWorktreeResolvesViaPrimary(t *testing.T) {
	parent, base, derived, tree := worktreeFixture(t)
	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, tree, plan.TargetAll, false, &out); err != nil {
		t.Fatalf("run: %v\n%s", err, out.String())
	}
	agents, err := os.ReadFile(filepath.Join(tree, "AGENTS.md"))
	if err != nil || string(agents) != "BASE PROSE\n\nDERIVED PROSE\n" {
		t.Fatalf("AGENTS.md = %q, %v", agents, err)
	}
	want := "resolved via primary checkout " + derived + ": " + base
	if !strings.Contains(out.String(), want) {
		t.Fatalf("output lacks %q:\n%s", want, out.String())
	}
	if _, err := os.Lstat(filepath.Join(parent, "worktree", "derived", "base")); !os.IsNotExist(err) {
		t.Fatalf("compile created a sibling beside the worktree: %v", err)
	}
}

// A dependency absent beside both the worktree and the primary refuses before
// compile writes anything: previously generated output is untouched.
func TestCompileLinkedWorktreeMissingDependencyPreservesOutput(t *testing.T) {
	_, base, _, tree := worktreeFixture(t)
	var out bytes.Buffer
	if err := run(weavefs.OSFS{}, tree, plan.TargetAll, false, &out); err != nil {
		t.Fatalf("first run: %v\n%s", err, out.String())
	}
	before := snapshotTree(t, tree)
	if err := os.RemoveAll(base); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	err := run(weavefs.OSFS{}, tree, plan.TargetAll, false, &out)
	if err == nil || !strings.Contains(err.Error(), "absent beside this worktree") {
		t.Fatalf("err = %v\n%s", err, out.String())
	}
	if after := snapshotTree(t, tree); !reflect.DeepEqual(before, after) {
		t.Fatalf("refused compile changed the worktree:\nbefore %v\nafter  %v", before, after)
	}
}

// snapshotTree maps every path under root (outside .git) to its content or
// symlink target.
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		switch {
		case rel == ".git":
			return nil
		case d.Type()&fs.ModeSymlink != 0:
			target, err := os.Readlink(p)
			snap[rel] = "-> " + target
			return err
		case d.IsDir():
			snap[rel] = "dir"
			return nil
		}
		b, err := os.ReadFile(p)
		snap[rel] = string(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return snap
}
