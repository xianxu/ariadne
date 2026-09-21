package plan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// GeneratedRuntimeGitignoreEntries is the exact historical fixed list, retained
// only for migration. New output coverage comes from GeneratedGitignoreEntries.
var GeneratedRuntimeGitignoreEntries = []string{
	"/AGENTS.md", // codex entry file (composed prose)
	"/CLAUDE.md", // claude entry file (composed prose) — Option B #107
	"/GEMINI.md", // gemini entry file (composed prose) — Option B #107
	"/.claude/skills/",
	"/.agents/skills/", // codex + gemini skill dir — Option B #107
	"/.claude/settings.json",
	"/.colima/",
	"/construct/scripts/vm-log.sh",
	"/" + walk.GeneratedRel + "/", // per-repo dynamic-skill materialization (#115 M3, single-sourced) — regenerated every compile
}

// EnsureGitignore carries exact generated entries for plain Apply. ApplyManaged
// derives these from its complete scoped inventory instead.
type EnsureGitignore struct{ Entries []string }

func (EnsureGitignore) isAction() {}

func ensureGitignoreText(current string, entries []string) (string, bool, error) {
	next, err := managedIgnoreText(current, entries)
	return next, next != current, err
}

func applyEnsureGitignore(fs weavefs.FS, gitignorePath string, entries []string) error {
	if fi, e := fs.Lstat(gitignorePath); e == nil && !fi.Mode().IsRegular() {
		return fmt.Errorf("gitignore is not a regular file: %s", gitignorePath)
	} else if e != nil && !os.IsNotExist(e) {
		return e
	}
	current, err := fs.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	next, changed, err := ensureGitignoreText(string(current), entries)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	if err := ensureParent(fs, gitignorePath); err != nil {
		return err
	}
	return writeManagedIgnore(fs, filepath.Dir(gitignorePath), next)
}
