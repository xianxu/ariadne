package refresh

import (
	"path/filepath"
	"testing"
)

// From a linked worktree, refresh checks a sourced substrate's origin where
// discovery resolved it, beside the primary checkout (#230 close review BR-8).
func TestRefreshLinkedWorktreeSubstrate(t *testing.T) {
	base := t.TempDir()
	dep := repo(t, base, "dep")
	host := repo(t, base, "host")
	commit(t, host.root, "construct/deps", "substrate ../dep "+dep.remote+"\n")
	git(t, host.root, "push", "origin", "main")
	tree := filepath.Join(filepath.Dir(host.root), "worktree", "host", "br")
	git(t, host.root, "worktree", "add", "-b", "br", tree)
	git(t, tree, "branch", "--set-upstream-to=origin/main")
	_, e, n := run(t, fixture{root: tree}, client(), false)
	if e != nil || n != 1 {
		t.Fatalf("refresh from a linked worktree: %v (compiles %d)", e, n)
	}
}
