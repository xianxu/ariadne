package gitx

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// #289: the active operation is the first marker present in the worktree's
// own git directory; any lstat failure other than absence is an error, never
// "no operation".
func TestActiveOperation(t *testing.T) {
	denied := errors.New("permission denied")
	for _, tc := range []struct {
		name    string
		present map[string]error // marker -> lstat result (nil = present)
		want    string
		wantErr bool
	}{
		{"none", nil, "", false},
		{"merge", map[string]error{"MERGE_HEAD": nil}, "MERGE_HEAD", false},
		{"rebase dir", map[string]error{"rebase-merge": nil}, "rebase-merge", false},
		{"bisect", map[string]error{"BISECT_START": nil}, "BISECT_START", false},
		{"sequencer", map[string]error{"sequencer": nil}, "sequencer", false},
		{"probe fails", map[string]error{"REVERT_HEAD": denied}, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lstat := func(p string) error {
				if err, ok := tc.present[filepath.Base(p)]; ok {
					return err
				}
				return fs.ErrNotExist
			}
			got, err := ActiveOperation("/g", lstat)
			if got != tc.want || (err != nil) != tc.wantErr {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

// #289: a worktree's git directory, from its .git directory or gitdir file.
func TestWorktreeGitDir(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{}
	read := func(p string) ([]byte, bool, error) { // content, isDir, err
		if p == filepath.Join(root, "plain", ".git") {
			return nil, true, nil
		}
		if c, ok := files[p]; ok {
			return []byte(c), false, nil
		}
		return nil, false, fs.ErrNotExist
	}
	files[filepath.Join(root, "abs", ".git")] = "gitdir: /primary/.git/worktrees/abs\n"
	files[filepath.Join(root, "rel", ".git")] = "gitdir: ../primary/.git/worktrees/rel\n"
	files[filepath.Join(root, "bad", ".git")] = "not a pointer\n"
	for name, want := range map[string]string{
		"plain": filepath.Join(root, "plain", ".git"),
		"abs":   "/primary/.git/worktrees/abs",
		"rel":   filepath.Join(root, "primary", ".git", "worktrees", "rel"),
	} {
		if got, err := WorktreeGitDir(filepath.Join(root, name), read); err != nil || got != want {
			t.Errorf("%s: %q %v, want %q", name, got, err, want)
		}
	}
	for _, name := range []string{"bad", "missing"} {
		if _, err := WorktreeGitDir(filepath.Join(root, name), read); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

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
	dir, err := WorktreeGitDir(linked, ReadGitPointer)
	if err != nil {
		t.Fatal(err)
	}
	if op, err := ActiveOperation(dir, Lstat); !strings.Contains(strings.ToLower(op), "rebase") || err != nil {
		t.Fatalf("operation %q, %v", op, err)
	}
	if op, _ := ActiveOperation(filepath.Join(repo, ".git"), Lstat); op != "" {
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
