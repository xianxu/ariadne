package gitx

import (
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"path/filepath"
	"testing"
)

func TestParseWorktreesConformanceNewlinePath(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit())
	path := filepath.Join(t.TempDir(), "linked\nworktree")
	testfix.Git(t, repo, "worktree", "add", "-b", "linked", path, "HEAD")
	t.Cleanup(func() { testfix.Git(t, repo, "worktree", "remove", "--force", path) })
	wantPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}

	porcelain := testfix.Capture(t, repo, "worktree", "list", "--porcelain", "-z")
	worktrees, err := ParseWorktrees([]byte(porcelain))
	if err != nil {
		t.Fatal(err)
	}
	for _, worktree := range worktrees {
		if worktree.Path == wantPath {
			if worktree.Branch != "linked" {
				t.Fatalf("linked worktree branch = %q, want linked", worktree.Branch)
			}
			return
		}
	}
	t.Fatalf("newline path %q absent from %+v", wantPath, worktrees)
}
