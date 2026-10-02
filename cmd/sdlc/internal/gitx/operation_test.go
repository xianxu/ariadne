package gitx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// #289: end to end, a conflicted rebase in a linked worktree is seen through
// that worktree's gitdir file (Git writes REBASE_HEAD and rebase-merge; the
// first in marker order is reported).
func TestActiveOperationSeesARebaseInALinkedWorktree(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())
	write := func(dir, body string) {
		if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(repo, "base\n")
	testfix.Git(t, repo, "add", "f.txt")
	testfix.Git(t, repo, "commit", "-qm", "base")
	linked := filepath.Join(t.TempDir(), "linked")
	testfix.Git(t, repo, "worktree", "add", "-q", "-b", "topic", linked)
	write(linked, "topic\n")
	testfix.Git(t, linked, "commit", "-qam", "topic")
	write(repo, "main\n")
	testfix.Git(t, repo, "commit", "-qam", "main")
	_ = runIgnoringExit(linked, "rebase", "main") // conflicts by construction
	dir, err := workspace.WorktreeGitDir(linked, workspace.ReadGitPointer)
	if err != nil {
		t.Fatal(err)
	}
	if op, err := workspace.ActiveOperation(dir, workspace.Lstat); !strings.Contains(strings.ToLower(op), "rebase") || err != nil {
		t.Fatalf("operation %q, %v", op, err)
	}
	if op, _ := workspace.ActiveOperation(filepath.Join(repo, ".git"), workspace.Lstat); op != "" {
		t.Fatalf("the primary is not rebasing, got %q", op)
	}
}

func runIgnoringExit(dir string, args ...string) error {
	_, err := RunGit(append([]string{"-C", dir}, args...)...)
	if err != nil && !strings.Contains(err.Error(), "exit") {
		return err
	}
	return nil
}
