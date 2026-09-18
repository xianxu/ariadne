package gitx

import (
	"fmt"
	"strings"
)

// EntryAt reports the tree entry of path at ref: its mode, and whether it is
// there at all. It separates "absent" (present=false, err=nil) from "could not
// tell" (err) — the distinction a failed `git show` alone cannot make — and is
// the one implementation of that question (#231: TrunkFile and close's
// shared-surface read both ask it). --end-of-options keeps a ref that begins
// with "-" from being read as a flag; --full-tree keeps the pathspec rooted at
// the top of the tree whatever dir is. dir "" runs in the current directory.
func EntryAt(dir, ref, path string) (mode string, present bool, err error) {
	out, errOut, err := runGitIn(dir, nil, "ls-tree", "--full-tree", "--end-of-options", ref, "--", path)
	if err != nil {
		return "", false, fmt.Errorf("ls-tree %s -- %s: %v\n%s", ref, path, err, errOut)
	}
	f := strings.Fields(strings.TrimSpace(string(out)))
	if len(f) == 0 {
		return "", false, nil
	}
	return f[0], true, nil
}

// BlobAt reads the blob at ref:path. Callers establish presence with EntryAt
// first; here a failure is an error, never "empty".
func BlobAt(dir, ref, path string) ([]byte, error) {
	out, errOut, err := runGitIn(dir, nil, "cat-file", "blob", ref+":"+path)
	if err != nil {
		return nil, fmt.Errorf("cat-file blob %s:%s: %v\n%s", ref, path, err, errOut)
	}
	return out, nil
}

// DiffPathsBothSides lists every path a window touches with rename detection
// OFF, so a rename contributes its source AND its destination (#231 BR-24).
// DiffNames reports a rename's destination only, which is right for counting
// what the window changed and wrong for asking what it touched: moving a file
// out of a protected directory must count as touching that directory. -z keeps
// non-ASCII paths unquoted (lessons: git's porcelain answers a human's question).
func DiffPathsBothSides(since, until string) ([]string, error) {
	out, errOut, err := runGitIn("", nil, "diff", "--name-only", "--no-renames", "-z", since+".."+until)
	if err != nil {
		return nil, fmt.Errorf("diff --name-only --no-renames %s..%s: %v\n%s", since, until, err, errOut)
	}
	var paths []string
	for _, p := range strings.Split(string(out), "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	return paths, nil
}
