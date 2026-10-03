package workspace

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// OperationMarkers are the files and directories Git keeps in a worktree's own
// git directory while an operation is in progress: the index and HEAD are then
// not the user's settled intent. The one list (#289) for every reader —
// landing, move, move-detail, fleet readiness and weave refresh.
var OperationMarkers = []string{
	"MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD", "REBASE_HEAD",
	"rebase-merge", "rebase-apply", "sequencer", "BISECT_LOG", "BISECT_START",
}

// ActiveOperation is the first marker present in gitDir, or "" when none is.
// Any lstat failure other than absence is an error: an unreadable directory
// never reads as "no operation".
func ActiveOperation(gitDir string, lstat func(string) error) (string, error) {
	for _, m := range OperationMarkers {
		err := lstat(filepath.Join(gitDir, m))
		switch {
		case err == nil:
			return m, nil
		case !errors.Is(err, fs.ErrNotExist):
			return "", fmt.Errorf("check %s for an operation in progress: %w", gitDir, err)
		}
	}
	return "", nil
}

// Lstat is ActiveOperation's file probe.
func Lstat(p string) error {
	_, err := os.Lstat(p)
	return err
}

// WorktreeGitDir is a worktree's own git directory, from its .git entry: the
// directory itself, or the target of a `gitdir:` file (a linked worktree or a
// separate git dir), relative targets resolved against the worktree. read
// returns the entry's content, or isDir for a directory.
func WorktreeGitDir(root string, read func(string) (content []byte, isDir bool, err error)) (string, error) {
	entry := filepath.Join(root, ".git")
	content, isDir, err := read(entry)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", entry, err)
	}
	if isDir {
		return entry, nil
	}
	line := bytes.TrimSpace(content)
	target, ok := bytes.CutPrefix(line, []byte("gitdir: "))
	if !ok || len(bytes.TrimSpace(target)) == 0 || bytes.ContainsAny(line, "\n") {
		return "", fmt.Errorf("%s is not a gitdir pointer", entry)
	}
	dir := string(bytes.TrimSpace(target))
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(root, dir)
	}
	return filepath.Clean(dir), nil
}

// ReadGitPointer is WorktreeGitDir's reader over the filesystem.
func ReadGitPointer(p string) ([]byte, bool, error) {
	info, err := os.Lstat(p)
	if err != nil {
		return nil, false, err
	}
	if info.IsDir() {
		return nil, true, nil
	}
	raw, err := os.ReadFile(p)
	return raw, false, err
}
