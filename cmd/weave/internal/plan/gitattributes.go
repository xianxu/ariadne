package plan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// UnionMergeAttributes are the append-only workshop lists every woven repo merges
// with git's built-in union driver (#320). Parallel landings append to them on both
// sides, and keeping both is the right resolution, so no conflict is ever raised.
// Only files whose lines are appended, never edited in place, belong here: union
// keeps both versions of a line edited on both sides.
var UnionMergeAttributes = []string{
	"workshop/lessons.md merge=union",
}

// EnsureGitattributes writes weave's delimited block into .gitattributes, keeping
// the authored lines after it (later lines win, so authored ones can override).
type EnsureGitattributes struct{ Entries []string }

func (EnsureGitattributes) isAction() {}

func applyEnsureGitattributes(fs weavefs.FS, root string, entries []string) error {
	p := filepath.Join(root, ".gitattributes")
	if fi, e := fs.Lstat(p); e == nil && !fi.Mode().IsRegular() {
		return fmt.Errorf("gitattributes is not a regular file: %s", p)
	} else if e != nil && !os.IsNotExist(e) {
		return e
	}
	current, err := fs.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	next, err := managedIgnoreText(string(current), entries, false)
	if err != nil {
		return fmt.Errorf("%s: %w", p, err)
	}
	return writeManagedFile(fs, root, ".gitattributes", next)
}
