package fleet

import (
	"os"
	"path/filepath"
	"strings"
)

// stagedRecordName is the per-worktree file `sdlc stage` (#344) writes after
// it detaches a checkout at another slot's commit for a smoke test.
const stagedRecordName = "sdlc-staged"

// StagedRecordPath is the staged record's path in the worktree's own git dir
// (never the shared one, so each slot has its own).
func StagedRecordPath(git GitReader, tree string) (string, error) {
	out, err := git.GitInDir(tree, "rev-parse", "--git-path", stagedRecordName)
	if err != nil {
		return "", err
	}
	p := gitLine(out)
	if !filepath.IsAbs(p) {
		p = filepath.Join(tree, p)
	}
	return p, nil
}

// StagedCommit is the commit sdlc last staged in tree ("" when none). A
// record is trusted only where it equals the detached HEAD exactly; a stale
// one (`:0` switched by hand) never vouches for another commit.
func StagedCommit(git GitReader, tree string) string {
	p, err := StagedRecordPath(git, tree)
	if err != nil {
		return ""
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}
