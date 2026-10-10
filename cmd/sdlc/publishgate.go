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
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/flow"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
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
func mergedCodecompleteIssues(ctx context.Context, baseRef, issuesDir string) ([]string, error) {
	refs, err := scanIssueFiles(ctx, baseRef, issuesDir, gitx.RunGit)
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
func runPublishGate(ctx context.Context, baseRef, issuesDir string, stderr io.Writer) error {
	if err := guardTransferredDetailsFn(ctx); err != nil {
		return err
	}
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.Fresh)
	if err != nil {
		return err
	}
	if rs.Tracker {
		// #252: this publish owns the closes whose evidence it carries.
		env, err := openTracker(ctx)
		if err != nil {
			return err
		}
		owned, err := ownedCompletions(env, rs, "HEAD", baseRef, false)
		if err != nil {
			return err
		}
		if err := refuseUnownedCompletions(ctx, env, rs, owned); err != nil {
			return err
		}
		return validatePublishAnchors(ctx, ownedPublishIssues(owned), stderr)
	}
	issues, err := mergedCodecompleteIssues(ctx, baseRef, issuesDir)
	if err != nil {
		return err
	}
	return validatePublishIssues(ctx, issues, stderr)
}

// refuseUnownedCompletions makes an orphaned close LOUD (#304 D8). A codecomplete card
// bound to this repository whose details this branch's patch changes, but which no rule
// owns, is a close the branch no longer carries (a squash or an amend dropped its
// Close-Token). Skipping it would publish code with no gate and leave the card
// codecomplete forever, which is how a rebase used to fail open.
func refuseUnownedCompletions(ctx context.Context, env *trackerEnv, rs tracker.Records, owned []ownedCompletion) error {
	mainBase, _, err := reviewMainBase(ctx, gitx.Capture("rev-parse", "HEAD"))
	if err != nil || mainBase == "" {
		return err // on main there is no branch patch; a criss-cross is refused below anyway
	}
	changed, err := gitx.DiffNames(mainBase, "HEAD")
	if err != nil {
		return fmt.Errorf("publish gate: could not list the branch patch (%v) — refusing to publish unverified", err)
	}
	inPatch := map[string]bool{}
	for _, p := range changed {
		inPatch[p] = true
	}
	isOwned := map[string]bool{}
	for _, oc := range owned {
		isOwned[oc.ID] = true
	}
	var orphans []string
	for _, rec := range rs.All() {
		if rec.Card == nil || rec.Duplicate || rec.Status() != "codecomplete" || isOwned[rec.ID] || rec.DetailPath == "" {
			continue
		}
		b, ok, err := issue.CardCompletion(rec.Card.Raw)
		if err != nil || !ok || b.Repository != env.target.Repository {
			continue
		}
		rel, err := filepath.Rel(canonRoot(env.root), canonRoot(rec.DetailPath))
		if err != nil || !inPatch[filepath.ToSlash(rel)] {
			continue
		}
		orphans = append(orphans, issue.CLIRef(rec.ID))
	}
	if len(orphans) == 0 {
		return nil
	}
	return fmt.Errorf("publish gate: #%s is codecomplete but this branch carries no close for it "+
		"(a squash or an amend dropped its Close-Token, #304).\n"+
		"  Re-run `sdlc close --issue %s --verified '<evidence>'`, then retry the publish.",
		strings.Join(orphans, ", #"), orphans[0])
}

// ownedPublishIssues anchors each owned completion on its evidence commit.
func ownedPublishIssues(owned []ownedCompletion) []publishIssue {
	entries := make([]publishIssue, 0, len(owned))
	for _, oc := range owned {
		entries = append(entries, publishIssue{Path: oc.DetailPath, Anchor: oc.Binding.EvidenceCommit})
	}
	return entries
}

// validatePublishIssues anchors pre-tracker issues on the newest commit that
// left them codecomplete, then applies the shared checks.
func validatePublishIssues(ctx context.Context, issues []string, stderr io.Writer) error {
	entries := make([]publishIssue, 0, len(issues))
	for _, p := range issues {
		a := codecompleteAnchorCommit(p)
		if a == "" {
			return fmt.Errorf(
				"publish gate: %s is codecomplete but has no close commit reachable from HEAD.\n"+
					"  Commit the `sdlc close` (its status flip must be committed), then retry the publish.", p)
		}
		entries = append(entries, publishIssue{Path: p, Anchor: a})
	}
	return validatePublishAnchors(ctx, entries, stderr)
}

// validatePublishAnchors shares the reviewed-head, docs-only and quick-flow
// checks across ordinary diff selection, immutable landing ownership and card
// completion bindings (#252).
func validatePublishAnchors(ctx context.Context, entries []publishIssue, stderr io.Writer) error {
	issues := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.Path != "" {
			issues = append(issues, e.Path)
		}
	}
	if len(entries) == 0 {
		// No codecomplete issue in this window (e.g. an intermediate push of
		// not-yet-closed work) — no invariant to enforce. Deterministic no-op.
		cinfo(stderr, "publish gate: no codecomplete issues in this window — nothing to verify")
		return nil
	}
	// #304 D9: each issue is checked against ITS OWN reviewed patch, replayed onto
	// today's main — never a commit count after an anchor, which reads every main
	// commit merged in after the close as unreviewed (A4). The anchor is the close's
	// evidence commit: under the FIX-THEN-SHIP protocol the fixes ride into it (#174).
	mainBase, stale, merr := reviewMainBase(ctx, gitx.Capture("rev-parse", "HEAD"))
	if stale != "" {
		cwarn(stderr, "publish gate: measured against "+stale)
	}
	deltas := make([]publishDelta, 0, len(entries))
	for _, e := range entries {
		d := publishDelta{Anchor: e.Anchor}
		switch {
		case merr != nil:
			d.Unresolvable = merr.Error()
		case !isResolvedSHA(e.Anchor) || gitx.Capture("rev-parse", "--verify", "-q", e.Anchor+"^{commit}") == "":
			d.Unresolvable = fmt.Sprintf("the reviewed commit %s is not in this repository (rebased in another clone, or its pin was removed)", shortOID(e.Anchor))
		default:
			base := e.Anchor // on main: nothing to replay onto
			if mainBase != "" {
				s, conflicted, err := gitx.RebasedReviewedBase(mainBase, e.Anchor)
				if err != nil {
					d.Unresolvable = err.Error()
					break
				}
				base, d.Conflicted = s, conflicted
			}
			paths, err := gitx.DiffNames(base, "HEAD")
			if err != nil {
				return fmt.Errorf("publish gate: could not diff the reviewed patch %s against HEAD (%v) — refusing to publish unverified", shortOID(e.Anchor), err)
			}
			d.Paths = paths
		}
		deltas = append(deltas, d)
	}
	// Every whole-issue close reviewed the ENTIRE branch patch, so on a branch carrying
	// several closes the newest one covers the rest: HEAD passes when any reviewed patch
	// covers it (the "newest anchor" rule, generalized off commit counts). A refusal
	// reports the closest — the patch with the fewest code paths left uncovered.
	best := deltas[0]
	for _, d := range deltas[1:] {
		if publishDeltaRank(d) < publishDeltaRank(best) {
			best = d
		}
	}
	pass, msg := classifyPublishDelta(best)
	if !pass {
		return errors.New(msg)
	}
	cinfo(stderr, msg)
	return quickGrewPastReview(ctx, issues)
}

// publishDeltaRank orders deltas from best covered to worst: an exact match, then
// doc-only, then by the number of uncovered code paths; unreplayable or conflicting
// patches rank last.
func publishDeltaRank(d publishDelta) int {
	if d.Unresolvable != "" || len(d.Conflicted) > 0 {
		return 1 << 30
	}
	code := 0
	for _, p := range d.Paths {
		if publishGateHasCodeSurface([]string{p}) {
			code++
		}
	}
	if code > 0 {
		return 2 + code
	}
	if len(d.Paths) > 0 {
		return 1
	}
	return 0
}

// publishDelta is what HEAD carries beyond one issue's reviewed patch.
type publishDelta struct {
	Anchor       string
	Conflicted   []string // main rewrote lines the review read; the resolution is unreviewed
	Paths        []string // diff(reviewed patch on today's main, HEAD)
	Unresolvable string   // why the reviewed patch could not be replayed
}

// publishGateRefusal prefixes every branch-patch refusal, the token gatesig keys on.
const publishGateRefusal = "publish gate: the reviewed patch no longer covers HEAD"

// classifyPublishDelta decides one issue's publish (#304 D9). Pure. Each refusal names
// its own cause and the one next action, a re-close of the code delta.
func classifyPublishDelta(d publishDelta) (pass bool, msg string) {
	reclose := "\n  Re-run `sdlc close --issue <N> --verified '<evidence>'` to review it, then retry the publish."
	switch {
	case d.Unresolvable != "":
		return false, fmt.Sprintf("%s: %s.%s", publishGateRefusal, d.Unresolvable, reclose)
	case len(d.Conflicted) > 0:
		return false, fmt.Sprintf("%s: main's changes conflict with the reviewed patch in %s — the resolution is unreviewed.%s",
			publishGateRefusal, strings.Join(d.Conflicted, ", "), reclose)
	}
	var code []string
	for _, p := range d.Paths {
		if publishGateHasCodeSurface([]string{p}) {
			code = append(code, p)
		}
	}
	switch {
	case len(code) > 0:
		return false, fmt.Sprintf("%s: code changed after `sdlc close` (reviewed %s): %s.%s\n"+
			"  (Merging or rebasing main never counts — only the branch's own changes do, #304. Doc-only deltas pass on their own, #174.)",
			publishGateRefusal, shortOID(d.Anchor), strings.Join(code, ", "), reclose)
	case len(d.Paths) > 0:
		return true, formatPublishGateDocsOnly(len(d.Paths), shortOID(d.Anchor))
	default:
		return true, fmt.Sprintf("publish gate: HEAD carries exactly the reviewed patch (%s) — reviewed-HEAD-unchanged ✓", shortOID(d.Anchor))
	}
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
func quickGrewPastReview(ctx context.Context, issues []string) error {
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
		base := boundaryWindowBase(ctx, strconv.Itoa(n), "", "", "")
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
func publishCodecompleteIssues(ctx context.Context, issuesDir string) ([]string, error) {
	if done, tracked, err := publishTrackerCompletions(ctx, issuesDir); tracked || err != nil {
		return done, err
	}
	refs, err := scanIssueFiles(ctx, "", issuesDir, nil)
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
	return fmt.Sprintf("publish gate: %d doc-only file(s) beyond the reviewed patch (%s) — no code surface, "+
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

// publishTrackerCompletions is the publish flip for a tracked repository (#252):
// every codecomplete card bound here whose evidence commit fresh main now
// carries goes done for its close generation. The archive that follows mirrors
// that done card into the details (#275); the card stays the authority.
func publishTrackerCompletions(ctx context.Context, issuesDir string) (done []string, tracked bool, err error) {
	rs, err := loadIssueRecords(ctx, issuesDir, tracker.Fresh)
	if err != nil || !rs.Tracker {
		return nil, rs.Tracker, err
	}
	abs, err := filepath.Abs(issuesDir)
	if err != nil {
		return nil, true, err
	}
	env, err := openTrackerAt(ctx, abs)
	if err != nil {
		return nil, true, err
	}
	if done, err = settleLandedCompletions(ctx, env, abs); err != nil {
		return done, true, err
	}
	return done, true, nil
}
