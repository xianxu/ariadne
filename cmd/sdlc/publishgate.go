// publishgate.go — the deterministic pre-publish gate for `sdlc merge` and
// `sdlc push` (#160). It REPLACES the pre-merge plan/specs/lessons LLM judges:
// all LLM review is now close-time (the boundary review), so the publish gate
// carries no LLM. It enforces the reviewed-HEAD-unchanged invariant
// (codecomplete ⟹ the close boundary review covered HEAD) and flips the merged
// codecomplete issues to done. For a quick-flow issue it also re-measures the
// final diff (#231): fixes made after the small-diff verdict are unmeasured by
// close, and past flow.MaxAddedLinesAfterReview the issue goes back to close.
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
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
	anchor, _ := codecompleteAnchorCommitAt("HEAD", issuePath, gitx.RunGit)
	return anchor
}

// mergedCodecompleteIssues returns the repo-relative paths of issue files changed in
// baseRef..HEAD whose CURRENT (working-tree) status is codecomplete — the set a
// publish is about to flip to done. Mirrors touchedIssuesNotDone's window scan
// (ARCH-DRY).
func mergedCodecompleteIssues(baseRef, issuesDir string) ([]string, error) {
	refs, err := scanIssueFiles(baseRef, issuesDir, gitx.RunGit)
	if err != nil {
		if scanErr, ok := err.(*issueFileScanError); ok {
			return nil, fmt.Errorf("git diff %s..HEAD: %w", baseRef, scanErr.Err)
		}
		return nil, fmt.Errorf("git diff %s..HEAD: %w", baseRef, err)
	}
	codecomplete := codecompleteIssueFiles(refs)
	paths := make([]string, 0, len(codecomplete))
	for _, ref := range codecomplete {
		paths = append(paths, ref.Path)
	}
	return paths, nil
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
	return validatePublishIssues(issues, stderr)
}

// validatePublishIssues shares the reviewed-head, docs-only and quick-flow
// checks across ordinary diff selection and immutable landing ownership.
func validatePublishIssues(issues []string, stderr io.Writer) error {
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
		ahead, ok := revCount(a + "..HEAD")
		if !ok {
			// Fail-closed: if we can't verify HEAD vs the anchor, refuse rather than
			// silently pass (unreachable in practice — the anchor is from HEAD's log).
			return fmt.Errorf("publish gate: could not compute rev-list %s..HEAD (git error) — refusing to publish unverified", shortSHA(a))
		}
		if minAhead < 0 || ahead < minAhead {
			minAhead, newestAnchor = ahead, a
		}
	}
	if minAhead > 0 {
		// #174: a post-close delta with no code surface (lessons.md, plan
		// ticks, atlas — the bookkeeping the close itself invites) does not
		// weaken the review's claims, which are about code behavior. Same
		// "no code surface" definition as the atlas gate's auto-satisfy
		// (#177, hasCodePath). Git errors keep the fail-closed posture.
		paths, derr := gitx.DiffNames(newestAnchor, "HEAD")
		if derr != nil {
			return fmt.Errorf("publish gate: could not diff %s..HEAD (%v) — refusing to publish unverified", shortSHA(newestAnchor), derr)
		}
		if !publishGateHasCodeSurface(paths) {
			cinfo(stderr, formatPublishGateDocsOnly(minAhead, shortSHA(newestAnchor)))
			return quickGrewPastReview(issues)
		}
		return fmt.Errorf(
			"publish gate: %d commit(s) landed after `sdlc close` (anchor %s) — the boundary review no longer covers HEAD.\n"+
				"  Re-run `sdlc close --issue <N> --verified '<evidence>'` to re-review the code delta, then retry the publish.\n"+
				"  (Next time: bundle post-close bookkeeping into the close commit — doc-only deltas pass on their own, #174.)",
			minAhead, shortSHA(newestAnchor))
	}
	if err := quickGrewPastReview(issues); err != nil {
		return err
	}
	cok(stderr, fmt.Sprintf("publish gate: HEAD unchanged since close (anchor %s) — reviewed-HEAD-unchanged ✓", shortSHA(newestAnchor)))
	return nil
}

// quickGrewPastReview refuses the publish of a quick-flow issue whose final diff
// grew past flow.MaxAddedLinesAfterReview (#231). Close measured the head its
// small-diff review saw; the fixes made after the verdict ride into the close
// commit, which is the anchor above, so nothing else measures them. The window
// is close's own (boundaryWindowBase), extended to HEAD, so the publish check
// and the close it sends the issue back to measure the same diff — and that
// close finds the shell crossed, upgrades the issue and runs the full review.
// Deterministic, like the rest of the gate: the review runs in close. A full
// issue (or one without a readable quick record) is not the quick flow's to
// re-measure.
func quickGrewPastReview(issues []string) error {
	for _, p := range issues {
		content, err := os.ReadFile(p)
		if err != nil {
			return fmt.Errorf("publish gate: read %s to check its flow: %v", p, err)
		}
		fm, _, perr := issue.Parse(string(content))
		if perr != nil {
			continue
		}
		rec, rerr := flow.Recorded(fm)
		if rerr != nil || rec == nil || rec.Kind() != flow.Quick {
			continue
		}
		n := issueIDFromPath(p)
		base := boundaryWindowBase(strconv.Itoa(n), "", "")
		if base == "" {
			continue // no #N commit anchors a window: nothing was measured at close either
		}
		stats, err := windowFileStats(base, "HEAD")
		if err != nil {
			return fmt.Errorf("publish gate: could not measure #%d's diff %s..HEAD (%v) — refusing to publish unverified", n, shortSHA(base), err)
		}
		if why := flow.Measure(stats, 0, nil).GrewPastReview(); why != "" {
			return fmt.Errorf(
				"publish gate: #%d is on the quick flow, and its diff grew to %s.\n"+
					"  Re-run `sdlc close --issue %d --verified '<evidence>'`: close measures the crossing, upgrades the issue\n"+
					"  to the full flow and runs the full review. Then retry the publish.", n, why, n)
		}
	}
	return nil
}

// publishCodecompleteIssues flips every codecomplete issue in issuesDir to done —
// the deterministic merge/push publish flip (#160). Run AFTER the invariant check +
// the merge/push, BEFORE archiving (which keys on IsTerminal). actual_hours was set
// at close, so the compiled done-guard is already satisfied. Returns the flipped
// issue paths (for logging); the caller's archive step stages + commits the moves.
//
// Scope is DIR-WIDE (glob), not window-scoped, matching archiveDoneIssues' existing
// behavior — on a healthy main no codecomplete issue persists outside a publish (each
// merge/push flips them), so the only codecomplete issues present are this publish's.
// (The invariant that gates un-reviewed drift is runPublishGate; this flip is the
// mechanical state change once that gate passed.)
func publishCodecompleteIssues(issuesDir string) ([]string, error) {
	refs, err := scanIssueFiles("", issuesDir, nil)
	if err != nil {
		return nil, err
	}
	today := time.Now().Format("2006-01-02")
	var flipped []string
	for _, ref := range codecompleteIssueFiles(refs) {
		content, err := publishedIssueContent(ref.Frontmatter, ref.Body, today)
		if err != nil {
			return flipped, err
		}
		if werr := os.WriteFile(ref.Path, content, 0o644); werr != nil {
			return flipped, fmt.Errorf("flip %s → done: %w", ref.Path, werr)
		}
		flipped = append(flipped, ref.Path)
	}
	return flipped, nil
}

// publishGateHasCodeSurface is the publish gate's code-surface predicate
// (#174 close review I1): hasCodePath's docs definition (#177), TIGHTENED so
// anything under cmd/ counts as code even when it's *.md — helptext is
// //go:embed'ed into the binary and shapes shipped agent-facing behavior, so
// a post-close helptext edit must not ride the doc-only pass. The atlas
// gate keeps plain hasCodePath (there, embedded docs SHOULD satisfy a docs
// demand); only the publish decision needs the stricter read. Pure.
func publishGateHasCodeSurface(paths []string) bool {
	for _, p := range paths {
		if !churn.IsDoc(p) || churn.IsEmbedded(p) {
			return true
		}
	}
	return false
}

// formatPublishGateDocsOnly renders the docs-only pass line (#174). Phrased
// to share NO vocabulary with the drift refusal ("landed after") — gatesig
// classifies transcripts by substring, so a pass line echoing refusal words
// would corrupt friction attribution (#172). Pure.
func formatPublishGateDocsOnly(n int, anchorShort string) string {
	return fmt.Sprintf("publish gate: %d doc-only commit(s) since close (anchor %s) — no code surface, "+
		"reviewed-HEAD-unchanged holds for code (#174)", n, anchorShort)
}

// revCount returns the commit count of a `git rev-list --count` range. ok is false
// when git errored (Capture returns "" — a valid count is always a number like "0"),
// so the caller can fail-closed rather than treat a git error as "no drift".
func revCount(rangeSpec string) (count int, ok bool) {
	out := strings.TrimSpace(gitx.Capture("rev-list", "--count", rangeSpec))
	if out == "" {
		return 0, false
	}
	n, err := strconv.Atoi(out)
	return n, err == nil
}

// codecompleteAnchorCommitAt is the bounded immutable counterpart of the legacy
// HEAD query. Query failure cannot masquerade as no close evidence.
func codecompleteAnchorCommitAt(ref, issuePath string, runGit func(...string) ([]byte, error)) (string, error) {
	out, err := runGit("--literal-pathspecs", "log", "--format=%H", "--max-count=10001", ref, "--", issuePath)
	if err != nil {
		return "", fmt.Errorf("read close history: %w", err)
	}
	commits := strings.Fields(string(out))
	if len(commits) > 10000 {
		return "", fmt.Errorf("close history exceeds 10000 commits")
	}
	for _, sha := range commits {
		entries, err := runGit("--literal-pathspecs", "ls-tree", "-z", sha, "--", issuePath)
		if err != nil {
			return "", fmt.Errorf("read close entry: %w", err)
		}
		if len(entries) == 0 {
			continue
		}
		content, err := runGit("cat-file", "blob", sha+":"+issuePath)
		if err != nil {
			return "", fmt.Errorf("read close body: %w", err)
		}
		fm, _, err := issue.Parse(string(content))
		if err != nil {
			return "", fmt.Errorf("parse close body: %w", err)
		}
		if status, _ := issue.GetField(fm, "status"); status == "codecomplete" {
			return sha, nil
		}
	}
	return "", nil
}
