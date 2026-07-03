// publishgate.go — the deterministic pre-publish gate for `sdlc merge` and
// `sdlc push` (#160). It REPLACES the pre-merge plan/specs/lessons LLM judges:
// all LLM review is now close-time (the boundary review), so the publish gate
// carries no LLM. It enforces the reviewed-HEAD-unchanged invariant
// (codecomplete ⟹ the close boundary review covered HEAD) and flips the merged
// codecomplete issues to done.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// codecompleteAnchorCommit returns the SHA of the NEWEST commit touching issuePath
// that leaves the file at `status: codecomplete` — the anchor for the
// reviewed-HEAD-unchanged invariant (#160). Because `sdlc close` is the SOLE writer
// of codecomplete (set-status refuses it), this is a close commit; a re-close after
// drift produces a newer such commit, so the anchor ADVANCES (the drift-recovery
// flow stays clean). "" when the file has no codecomplete-leaving commit reachable
// from HEAD.
//
// Derivation is a CONTENT READ (not the commit-message trailer grep of
// previousReviewBoundary — a genuinely different signal, so a distinct helper, not
// ARCH-DRY reuse): walk the file's commits newest-first and return the first whose
// content parses to codecomplete.
//
// Residual (by design, does not arise in practice): a SINGLE commit that both edits
// the issue file AND changes code WITHOUT going through close would be mis-picked as
// the anchor. But post-close code changes must re-close, set-status can't write
// codecomplete, and hand-editing frontmatter is off-convention — so it doesn't occur.
func codecompleteAnchorCommit(issuePath string) string {
	out, err := gitx.RunGit("log", "--format=%H", "--", issuePath)
	if err != nil {
		return ""
	}
	for _, sha := range strings.Fields(string(out)) {
		content, err := gitx.RunGit("show", sha+":"+issuePath)
		if err != nil {
			continue
		}
		fm, _, perr := issue.Parse(string(content))
		if perr != nil {
			continue
		}
		if st, _ := issue.GetField(fm, "status"); st == "codecomplete" {
			return sha
		}
	}
	return ""
}

// mergedCodecompleteIssues returns the repo-relative paths of issue files changed in
// baseRef..HEAD whose CURRENT (working-tree) status is codecomplete — the set a
// publish is about to flip to done. Mirrors touchedIssuesNotDone's window scan
// (ARCH-DRY).
func mergedCodecompleteIssues(baseRef, issuesDir string) ([]string, error) {
	out, err := gitx.RunGit("diff", "--name-only", baseRef+"..HEAD", "--", issuesDir+"/*.md")
	if err != nil {
		return nil, fmt.Errorf("git diff %s..HEAD: %w", baseRef, err)
	}
	var cc []string
	for _, p := range splitNonEmptyLines(string(out)) {
		data, derr := os.ReadFile(p)
		if derr != nil {
			continue
		}
		fm, _, perr := issue.Parse(string(data))
		if perr != nil {
			continue
		}
		if st, _ := issue.GetField(fm, "status"); st == "codecomplete" {
			cc = append(cc, p)
		}
	}
	return cc, nil
}

// runPublishGate is the deterministic pre-publish check (#160) — no LLM. It
// enumerates the codecomplete issues this publish will flip, finds the NEWEST close
// anchor among them (the last `sdlc close`, whose whole-issue boundary review
// covered branch-point..anchor — hence a branch-level check suffices, no false
// per-issue "drift" refusal on multi-issue branches), and refuses unless HEAD is
// unchanged since that anchor. On refusal the message points at re-running close.
func runPublishGate(baseRef, issuesDir string, stderr io.Writer) error {
	issues, err := mergedCodecompleteIssues(baseRef, issuesDir)
	if err != nil {
		return err
	}
	if len(issues) == 0 {
		// No codecomplete issue in this window (e.g. an intermediate push of
		// not-yet-closed work) — no invariant to enforce. Deterministic no-op.
		cinfo(stderr, "publish gate: no codecomplete issues in this window — nothing to verify")
		return nil
	}
	newestAnchor, minAhead := "", -1
	for _, p := range issues {
		a := codecompleteAnchorCommit(p)
		if a == "" {
			return fmt.Errorf(
				"publish gate: %s is codecomplete but has no close commit reachable from HEAD.\n"+
					"  Commit the `sdlc close` (its status flip must be committed), then retry the publish.", p)
		}
		if ahead := revCount(a + "..HEAD"); minAhead < 0 || ahead < minAhead {
			minAhead, newestAnchor = ahead, a
		}
	}
	if minAhead > 0 {
		return fmt.Errorf(
			"publish gate: %d commit(s) landed after `sdlc close` (anchor %s) — the boundary review no longer covers HEAD.\n"+
				"  Re-run `sdlc close --issue <N> --verified '<evidence>'` to re-review the delta, then retry the publish.",
			minAhead, shortSHA(newestAnchor))
	}
	cok(stderr, fmt.Sprintf("publish gate: HEAD unchanged since close (anchor %s) — reviewed-HEAD-unchanged ✓", shortSHA(newestAnchor)))
	return nil
}

// publishCodecompleteIssues flips every codecomplete issue in issuesDir to done —
// the deterministic merge/push publish flip (#160). Run AFTER the invariant check +
// the merge/push, BEFORE archiving (which keys on IsTerminal). actual_hours was set
// at close, so the compiled done-guard is already satisfied. Returns the flipped
// issue paths (for logging); the caller's archive step stages + commits the moves.
func publishCodecompleteIssues(issuesDir string) ([]string, error) {
	matches, _ := filepath.Glob(filepath.Join(issuesDir, "[0-9][0-9][0-9][0-9][0-9][0-9]-*.md"))
	sort.Strings(matches)
	today := time.Now().Format("2006-01-02")
	var flipped []string
	for _, p := range matches {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		fm, body, perr := issue.Parse(string(data))
		if perr != nil {
			continue
		}
		if st, _ := issue.GetField(fm, "status"); st != "codecomplete" {
			continue
		}
		fm = issue.SetField(fm, "status", "done")
		fm = issue.SetField(fm, "updated", today)
		if werr := os.WriteFile(p, []byte(issue.Compose(fm, body)), 0o644); werr != nil {
			return flipped, fmt.Errorf("flip %s → done: %w", p, werr)
		}
		flipped = append(flipped, p)
	}
	return flipped, nil
}

// revCount returns the number of commits in a `git rev-list --count` range (0 on error).
func revCount(rangeSpec string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(gitx.Capture("rev-list", "--count", rangeSpec)))
	return n
}
