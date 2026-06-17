package plan

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// gitignore.go is weave's generated-runtime ignore mechanism: weave GENERATES a
// fixed set of runtime artifacts (the composed AGENTS.md, the .claude/skills
// symlinks, the merged .claude/settings.json, the .colima/ VM tree, the
// vm-log.sh helper), so weave OWNS ensuring the repo's .gitignore covers them
// (ARCH-DRY — one owner for "this artifact is weave-produced"). Without this a
// fresh `weave compile` leaves a dirty `git status` in every derivative; the
// hand-added /AGENTS.md ignore (parley) was the symptom. The ensure runs on
// EVERY compile so derivatives get it automatically, with no per-repo hand-edit.
//
// The pure core stays pure (ARCH-PURE): the entry LIST + the pure
// ensure-text transform live here as data + a string function; the actual
// .gitignore read/write is the IO seam (applyEnsureGitignore, called from
// plan.Apply). The compile lowering (main.planActions) appends exactly ONE
// EnsureGitignore action per weave run.
//
// What is NOT ignored — the pre-weave BOOTSTRAP scaffolding (bootstrap.sh,
// Makefile, Makefile.workflow, construct/scripts/{lib-deps,bootstrap-peers,
// clone-data-deps,list-peers,apply-gitignore-entries}.sh, the construct/* dir
// symlinks, .claude/settings.ariadne.json). A fresh clone must commit those
// BEFORE weave can run (the bootstrap chicken-and-egg), and they are stable —
// so they are tracked, not ignored. Only weave's OWN generated outputs go here.

// GeneratedRuntimeGitignoreEntries is the FIXED set of repo-relative paths weave
// generates and therefore ensures the repo's .gitignore covers. Order is the
// order appended to a .gitignore missing them. Leading-slash anchored to the
// repo root (the artifacts live at fixed top-level locations), trailing-slash on
// directories — matching git's own .gitignore grammar and the existing
// hand-added `/AGENTS.md` entry.
var GeneratedRuntimeGitignoreEntries = []string{
	"/AGENTS.md", // codex entry file (composed prose)
	"/CLAUDE.md", // claude entry file (composed prose) — Option B #107
	"/GEMINI.md", // gemini entry file (composed prose) — Option B #107
	"/.claude/skills/",
	"/.agents/skills/", // codex + gemini skill dir — Option B #107
	"/.claude/settings.json",
	"/.colima/",
	"/construct/scripts/vm-log.sh",
}

// EnsureGitignore ensures the repo's .gitignore contains every entry in Entries,
// appending the absent ones (idempotent: a present entry is never duplicated,
// existing entries/comments are preserved). It is weave's owned mechanism for
// keeping its generated-runtime artifacts out of `git status`. The planner emits
// one per compile carrying GeneratedRuntimeGitignoreEntries; Apply reads the
// live .gitignore and appends what is missing (the IO seam,
// applyEnsureGitignore). A pure Action — it carries only the entry list; the
// read/write is the seam's.
type EnsureGitignore struct {
	Entries []string
}

func (EnsureGitignore) isAction() {}

// ensureGitignoreText is the pure transform behind applyEnsureGitignore: given a
// .gitignore's current content and the entries to ensure, it returns the next
// content with every ABSENT entry appended (exact whole-line match, the
// `grep -qxF` semantics of the retired apply-gitignore-entries.sh) and whether
// anything changed. Present entries are never duplicated; existing lines/comments
// are preserved verbatim. A non-empty file not ending in a newline gets one
// before the first appended entry, so entries never glue onto a trailing line.
// Pure (string in/out) ⇒ unit-tested directly; the IO seam only does the
// read/write around it.
func ensureGitignoreText(current string, entries []string) (string, bool) {
	present := map[string]bool{}
	for _, line := range strings.Split(current, "\n") {
		present[line] = true
	}
	next := current
	changed := false
	for _, entry := range entries {
		if present[entry] {
			continue
		}
		if next != "" && !strings.HasSuffix(next, "\n") {
			next += "\n"
		}
		next += entry + "\n"
		present[entry] = true // guard against a duplicate entry in the input list
		changed = true
	}
	return next, changed
}

// applyEnsureGitignore is the IO seam for EnsureGitignore: read the repo's
// .gitignore (absent ⇒ empty), append the missing entries via the pure
// ensureGitignoreText, and write it back ONLY when something changed (no churn
// on a re-weave once the entries are present — running weave twice never
// duplicates a line). gitignorePath is the absolute path to the repo's
// .gitignore. All IO lives here (ARCH-PURE); the transform is the pure function
// above.
func applyEnsureGitignore(fs weavefs.FS, gitignorePath string, entries []string) error {
	var current string
	if data, err := fs.ReadFile(gitignorePath); err == nil {
		current = string(data)
	}
	next, changed := ensureGitignoreText(current, entries)
	if !changed {
		return nil
	}
	if err := ensureParent(fs, gitignorePath); err != nil {
		return err
	}
	if err := fs.WriteFile(gitignorePath, []byte(next)); err != nil {
		return fmt.Errorf("apply ensure-gitignore: write %s: %w", gitignorePath, err)
	}
	return nil
}
