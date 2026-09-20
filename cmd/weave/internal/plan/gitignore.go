package plan

import (
	"fmt"
	"os"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
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
// mergeManagedBlock transform live here as data + a string function; the actual
// .gitignore read/write is the IO seam (applyEnsureGitignore, called from
// plan.Apply). The compile lowering (main.planActions) appends exactly ONE
// EnsureGitignore action per weave run.
//
// WHAT IS NOT IGNORED is decided by OWNERSHIP, not by a list. The paragraph that
// used to sit here justified tracking the whole symlink class as "pre-weave
// BOOTSTRAP scaffolding … a fresh clone must commit those BEFORE weave can run
// (the bootstrap chicken-and-egg)". #225 built the owner-resolution fallbacks
// that dissolved that chicken-and-egg, and the list was never shrunk — a stale
// justification nothing could fail on, which is how it survived months of
// manifest edits (ariadne#239 defect 1).
//
// The rule that replaces it: weave IGNORES WHAT IT RE-DERIVES and TRACKS WHAT IT
// MERELY PROVISIONS. M3 makes the entry list derive from the manifest walk under
// that rule; until then the fixed list below stands.

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
	"/" + walk.GeneratedRel + "/", // per-repo dynamic-skill materialization (#115 M3, single-sourced) — regenerated every compile
}

// EnsureGitignore makes the repo's .gitignore carry exactly Entries inside
// weave's delimited region, REPLACING that region wholesale (#239 M2) — so an
// entry weave no longer produces loses its line. Lines outside the markers are
// the repo's own and are preserved. It is weave's owned mechanism for keeping
// its generated artifacts out of `git status`. The planner emits one per
// compile; Apply runs the pure mergeManagedBlock and writes back only on change
// (the IO seam, applyEnsureGitignore), failing closed on an unreadable file or
// an unparseable region. A pure Action — it carries only the entry list.
type EnsureGitignore struct {
	Entries []string
}

func (EnsureGitignore) isAction() {}

// The managed region's delimiters, matched as EXACT WHOLE LINES — a line that
// merely CONTAINS a marker (say, prose about this mechanism) is not a marker.
//
// This is NOT protection against a line that *is* the marker. A .gitignore that
// quotes the open marker verbatim on its own line — e.g. a comment block
// documenting the convention — reads as a second opening marker and makes
// `make weave` fail closed. That is the workshop/lessons.md lesson in its exact
// form ("a search that keys on content cannot see the content that describes
// it"), and whole-line matching does not dissolve it: the marker is the only
// thing distinguishing weave's region, so a verbatim copy IS ambiguous. Pinned
// by TestManagedBlockTreatsAQuotedMarkerAsAMarker; the failure is loud and the
// message says to delete the block, which is the right repair.
const (
	managedBlockOpen  = "# >>> weave-generated — managed by `make weave`, do not edit >>>"
	managedBlockClose = "# <<< weave-generated <<<"
)

// legacyBlanketEntries are pre-#239 BLANKET directory ignores that the per-path
// derivation supersedes. They must be absorbed even though they match no derived
// entry, because a derivative never runs the intermediate M2 binary: it goes
// straight from the hardcoded list to the derived one, and an exact-line absorb
// would strand these outside the block FOREVER as permanent directory ignores —
// the pair#64 hazard (a blanket `bin/` ignore made tracked shell scripts look
// disposable and a propagate-base sweep git-rm'd them).
//
// ARCH-FUNERAL: this list is a one-time migration aid and names its own end.
// Delete it once `git grep` finds no repo carrying these lines outside a managed
// block (checked at #239's close).
var legacyBlanketEntries = []string{
	"/.claude/skills/",
	"/.agents/skills/",
	"/.colima/",
}

// mergeManagedBlock is the pure transform behind applyEnsureGitignore: given a
// .gitignore's current content and the entries weave owns, it returns the next
// content, whether anything changed, and an error if the existing block cannot
// be parsed.
//
//   - INSIDE the markers: replaced WHOLESALE, so an entry weave no longer
//     produces (a retired manifest row) loses its ignore line. The append-only
//     predecessor could not do this.
//   - OUTSIDE the markers: preserved, in their original relative order. (A block
//     that sat mid-file is moved to the end once, and trailing blank lines are
//     normalized — both one-time and stable thereafter.)
//   - MIGRATION: a loose line outside the block is absorbed when it exactly
//     matches an entry the block now owns, or when it is a legacyBlanketEntry
//     the per-path derivation supersedes.
//   - No block yet → append one at the end.
//
// Fails closed (ARCH-SECURE): an unterminated or duplicated marker pair is an
// error naming the remedy, never a guessed splice point. The duplicated case is
// what a git MERGE CONFLICT between two branches that both regenerated the block
// produces, and it surfaces as a failing `make weave` — so the message has to
// say what to do about it.
func mergeManagedBlock(current string, entries []string) (string, bool, error) {
	const remedy = " — delete the managed block and re-run `make weave`"
	lines := strings.Split(current, "\n")
	openIdx, closeIdx := -1, -1
	for i, line := range lines {
		switch line {
		case managedBlockOpen:
			if openIdx != -1 {
				return "", false, fmt.Errorf("duplicate weave-generated opening marker (line %d) — a merge conflict?%s", i+1, remedy)
			}
			openIdx = i
		case managedBlockClose:
			if closeIdx != -1 {
				return "", false, fmt.Errorf("duplicate weave-generated closing marker (line %d) — a merge conflict?%s", i+1, remedy)
			}
			closeIdx = i
		}
	}
	switch {
	case openIdx == -1 && closeIdx != -1:
		return "", false, fmt.Errorf("weave-generated closing marker with no opening marker%s", remedy)
	case openIdx != -1 && closeIdx == -1:
		return "", false, fmt.Errorf("weave-generated opening marker with no closing marker%s", remedy)
	case openIdx != -1 && closeIdx < openIdx:
		return "", false, fmt.Errorf("weave-generated markers are inverted%s", remedy)
	}

	absorb := map[string]bool{}
	for _, e := range entries {
		absorb[e] = true
	}
	for _, e := range legacyBlanketEntries {
		absorb[e] = true
	}
	var outside []string
	for i, line := range lines {
		if openIdx != -1 && i >= openIdx && i <= closeIdx {
			continue
		}
		if absorb[line] {
			continue
		}
		outside = append(outside, line)
	}
	// Split/Join round-trips a trailing newline as a final empty element; drop
	// trailing blanks so the block appends cleanly.
	for len(outside) > 0 && outside[len(outside)-1] == "" {
		outside = outside[:len(outside)-1]
	}

	// DEDUPE the entry list. The retired ensureGitignoreText guarded this
	// explicitly ("guard against a duplicate entry in the input list") and
	// dropping the guard regressed it silently — the test named for the property
	// had been rewritten into a tautology and could not see it (#239 M2 BR-19).
	// Order-stable: first occurrence wins, so the block stays deterministic.
	block := []string{managedBlockOpen}
	emitted := map[string]bool{}
	for _, e := range entries {
		if emitted[e] {
			continue
		}
		emitted[e] = true
		block = append(block, e)
	}
	block = append(block, managedBlockClose)

	next := strings.Join(append(outside, block...), "\n") + "\n"
	return next, next != current, nil
}

// applyEnsureGitignore is the IO seam for EnsureGitignore: read the repo's
// .gitignore, rewrite weave's region via the pure mergeManagedBlock, and write
// back ONLY when something changed (no churn on a re-weave — running weave twice
// is byte-identical). gitignorePath is the absolute path to the repo's
// .gitignore. All IO lives here (ARCH-PURE); the transform is the pure function
// above.
//
// Two FAIL-CLOSED paths, both because the block is written wholesale:
//   - an unreadable .gitignore is an ERROR, not an empty file (treating it as
//     empty would replace every repo-owned entry with weave's block alone);
//   - an unparseable region is an error naming the remedy.
func applyEnsureGitignore(fs weavefs.FS, gitignorePath string, entries []string) error {
	var current string
	if data, err := fs.ReadFile(gitignorePath); err == nil {
		current = string(data)
	} else if !os.IsNotExist(err) {
		// Absent is fine (⇒ empty). Any OTHER read failure must NOT be treated
		// as an empty file: the block is written WHOLESALE (#239 M2), so doing
		// so would replace the repo's own entries with weave's block alone.
		// Harmless while this appended; fatal now.
		return fmt.Errorf("apply ensure-gitignore: read %s: %w", gitignorePath, err)
	}
	next, changed, err := mergeManagedBlock(current, entries)
	if err != nil {
		// The pure transform's messages carry no path prefix — the seam supplies
		// it, so prefixing in both places would double it.
		return fmt.Errorf("apply ensure-gitignore: %s: %w", gitignorePath, err)
	}
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
