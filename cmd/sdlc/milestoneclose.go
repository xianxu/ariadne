// milestoneclose.go — `sdlc milestone-close` subcommand.
//
// THE milestone close: runs the mechanical close (via computeClose with the
// milestone set, then applyClose — #139) and adds the AGENTS.md §3 mandatory post-milestone
// code review as an auto-dispatched follow-on — after the close completes, fires
// the one binary-owned boundary review (dispatchBoundaryReview, shared with
// `sdlc close` since #69) against the commit window for the milestone.
//
// #146: `sdlc close --milestone Mx` used to be the "flag on close" spelling that
// ran the mechanical close WITHOUT the review — a redundant, unlabeled bypass. It
// was removed (close refuses --milestone and redirects here). To close a milestone
// without the review, use the self-labeling `sdlc milestone-close --no-judge`.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/pkg/vocab"
)

type milestoneCloseFlags struct {
	Issue         int
	Milestone     string
	Actual        string
	Verified      string
	Force         bool
	DryRun        bool
	NoJudge       bool   // skip the auto-dispatched milestone-review
	Agent         string // forwarded to the judge dispatch
	AgentExplicit bool
	BrainDir      string
	IssuesDir     string
	PlansDir      string

	// Per-gate close bypasses (#67), threaded into the delegated computeClose.
	NoActual    bool
	NoVerified  bool
	NoReclose   bool
	NoAtlas     bool
	NoVerdict   bool
	NoLedger    bool
	NoPlanCheck bool
	NoProject   bool
}

// reviewResult bundles the outputs of the post-milestone judge call that
// downstream artifacts (commit trailer, log-line mirror) need to embed.
// "not-run" verdict + a Reason populated when the judge was skipped or
// errored — the operator should still be able to reconstruct what
// happened from the trailer alone.
type reviewResult struct {
	Verdict     judge.Verdict
	Reason      string // populated for not-run / unknown
	Base        string // short SHA
	Head        string // reviewed head: long SHA (falls back to "HEAD" only when rev-parse fails)
	BaseLong    string // long SHA, used by trailer-verifier lookups in close
	SidecarPath string // #136: durable final-review-response path ("" when no review ran)
	Output      string // semantic review body, retained when sidecar writing is deferred
	Agent       string // resolved reviewer CLI, retained for deferred sidecar metadata
	// Round is the findings block this review emitted (#194 M2), parsed but NOT yet
	// applied: dispatch runs with the repo transaction lock released, and the ledger is
	// a repo write, so it is persisted at finalize time — the same deferral the sidecar
	// uses. nil when the reviewer emitted no valid fence.
	Round *gatestate.RoundReport
	// ProtocolError records a reviewer that produced no parseable findings block. The
	// round is still persisted: dropping it would leave len(Rounds) at 0 forever for a
	// CLI that never emits the fence, so the round cap could never bound the loop it
	// exists to bound (the reasoning changecode.go:490 spells out for the plan gate).
	ProtocolError string
}

func NewMilestoneCloseCmd() *cobra.Command {
	f := milestoneCloseFlags{}
	cmd := markManualLockCommand(&cobra.Command{
		Use:           "milestone-close",
		Short:         "Close one milestone of an issue + auto-dispatch post-milestone review (AGENTS.md §3)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"milestone-close\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard
			f.AgentExplicit = cmd.Flags().Changed("agent")
			return runMilestoneCloseLocked(cmd, cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "ariadne workshop issue ID (required, positive)")
	cmd.Flags().StringVar(&f.Milestone, "milestone", "", "milestone tag e.g. M4 (required)")
	cmd.Flags().StringVar(&f.Actual, "actual", "", "focused dev-hours for this milestone")
	cmd.Flags().StringVar(&f.Verified, "verified", "", "one-line evidence the milestone meets done-when")
	cmd.Flags().BoolVar(&f.Force, "force", false, "bypass ALL close gates (≡ every --no-* flag); reason in --verified")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "plan only; do not write or dispatch judge")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip the auto-dispatched milestone-review")
	// Per-gate close bypasses (#67) — forwarded to computeClose; --force waives all.
	cmd.Flags().BoolVar(&f.NoActual, "no-actual", false, "record actual_hours: N/A on issue close / skip actual on milestone close")
	cmd.Flags().BoolVar(&f.NoVerified, "no-verified", false, "bypass the VERIFIED-evidence requirement")
	cmd.Flags().BoolVar(&f.NoReclose, "no-reclose-guard", false, "bypass the already-done refusal")
	cmd.Flags().BoolVar(&f.NoAtlas, "no-atlas", false, "bypass the atlas/ change check (no new surface)")
	cmd.Flags().BoolVar(&f.NoLedger, "no-ledger", false, "bypass the boundary gate ledger's open-findings refusal (#194)")
	cmd.Flags().BoolVar(&f.NoVerdict, "no-verdict", false, "bypass the milestone Review-Verdict check")
	cmd.Flags().BoolVar(&f.NoPlanCheck, "no-plan-check", false, "bypass the unchecked-## Plan-items refusal")
	cmd.Flags().BoolVar(&f.NoProject, "no-project", false, "bypass the project detail-block update requirement")
	cmd.Flags().StringVar(&f.Agent, "agent", "", "agent CLI for judge dispatch (claude | codex | gemini)")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "path to the brain repo (for project-file lookup)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.PlansDir, "plans-dir", envOr("WF_PLANS_DIR", "workshop/plans"), "directory holding durable plans + gate sidecars (#187)")
	return cmd
}

func (f *milestoneCloseFlags) closeFlags() *closeFlags {
	return &closeFlags{
		Issue:         f.Issue,
		Milestone:     f.Milestone,
		Actual:        f.Actual,
		Verified:      f.Verified,
		Force:         f.Force,
		DryRun:        f.DryRun,
		BrainDir:      f.BrainDir,
		IssuesDir:     f.IssuesDir,
		PlansDir:      f.PlansDir,
		Agent:         f.Agent,
		AgentExplicit: f.AgentExplicit,
		NoActual:      f.NoActual,
		NoVerified:    f.NoVerified,
		NoReclose:     f.NoReclose,
		NoAtlas:       f.NoAtlas,
		NoVerdict:     f.NoVerdict,
		NoLedger:      f.NoLedger,
		NoPlanCheck:   f.NoPlanCheck,
		NoProject:     f.NoProject,
	}
}

func runMilestoneClose(stdout, stderr io.Writer, f *milestoneCloseFlags) error {
	if f.Milestone == "" {
		die(stderr, "--milestone is required for milestone-close (use `sdlc close` without it for full-issue close)")
	}
	if f.Issue <= 0 {
		die(stderr, fmt.Sprintf("--issue is required and must be positive (got %d)", f.Issue))
	}

	// Step 1: build the closeFlags for the mechanical close (computed below via
	// computeClose — #139's compute→review→finalize; NOT runClose, which is test-only).
	closeF := f.closeFlags()
	// Step 1: COMPUTE the mechanical close — write NOTHING yet (#139). The review
	// runs against the un-mutated tree; applyClose fires only after a finalizing
	// verdict, so a REWORK/unexpected milestone review leaves nothing written.
	r := computeClose(stderr, closeF)

	// Step 2: figure out the review window (used regardless of whether the judge
	// actually runs — the trailer always carries it). The base is the prior review
	// boundary so inter-milestone #N-but-not-Mx commits are covered (#58); resolving
	// needs the issue file to find that boundary.
	issuePath, perr := issueFilePath(f.IssuesDir, f.Issue)
	if perr != nil {
		cwarn(stderr, fmt.Sprintf("resolve issue file for review window: %v", perr))
	}
	base, baseLong, head := resolveReviewWindow(strconv.Itoa(f.Issue), f.Milestone, issuePath)

	// Step 3: dispatch → finalize-on-verdict, or short-circuit the explicit skips.
	switch {
	case f.NoJudge || f.Force:
		// Explicit operator skip → finalize (annotate + trailer as before). --force
		// implies --no-judge per its "bypass ALL gates" contract (#139 I2), matching
		// full-issue close's f.skip("judge"); otherwise a --force milestone-close would
		// still dispatch and could halt/rework, defeating the emergency bypass.
		cinfo(stderr, "skipping milestone-review per --no-judge (or --force)")
		applyClose(stdout, stderr, closeRunner, closeF, r)
		emitTrailerBlock(stdout, reviewResult{Verdict: judge.VerdictNotRun, Reason: "--no-judge", Base: base, Head: head, BaseLong: baseLong}, "milestone-close")
		if err := annotateLogLineWithVerdict(f.IssuesDir, f.Issue, f.Milestone, judge.VerdictNotRun); err != nil {
			cwarn(stderr, fmt.Sprintf("log-line verdict annotation skipped: %v", err))
		}
		return nil
	case f.DryRun:
		cinfo(stderr, "dry-run — would dispatch judge milestone-review")
		printCloseDryRun(stderr, r)
		emitTrailerBlock(stdout, reviewResult{Verdict: judge.VerdictNotRun, Reason: "--dry-run", Base: base, Head: head, BaseLong: baseLong}, "milestone-close")
		return nil
	}

	return reviewThenFinalize(stdout, stderr, closeF, r, boundaryReviewParams{
		Label:         fmt.Sprintf("#%d %s", f.Issue, f.Milestone),
		Base:          base,
		BaseLong:      baseLong,
		Head:          head,
		IssuesDir:     f.IssuesDir,
		Agent:         f.Agent,
		AgentExplicit: f.AgentExplicit,
		IssueNum:      f.Issue,
		Milestone:     f.Milestone,
		PlansDir:      resolvePlansDir(f.PlansDir),
	})
}

func runMilestoneCloseLocked(cmd *cobra.Command, stdout, stderr io.Writer, f *milestoneCloseFlags) error {
	if f.Milestone == "" || f.Issue <= 0 || f.NoJudge || f.Force || f.DryRun {
		return withRequiredRepoTransactionLock(cmd, func() error {
			return runMilestoneClose(stdout, stderr, f)
		})
	}

	closeF := f.closeFlags()
	var r closeResult
	var base, baseLong, head, prior string
	var snapshot closeReviewSnapshot
	if err := withRequiredRepoTransactionLock(cmd, func() error {
		r = computeClose(stderr, closeF)
		issuePath, perr := issueFilePath(f.IssuesDir, f.Issue)
		if perr != nil {
			cwarn(stderr, fmt.Sprintf("resolve issue file for review window: %v", perr))
		}
		base, baseLong, head = resolveReviewWindow(strconv.Itoa(f.Issue), f.Milestone, issuePath)
		snapshot = captureCloseReviewSnapshot(r, head, f.Milestone)
		prior = boundaryPriorFindings(stderr, boundaryReviewParams{
			IssuesDir: f.IssuesDir, IssueNum: f.Issue, Milestone: f.Milestone, PlansDir: resolvePlansDir(f.PlansDir),
		})
		return nil
	}); err != nil {
		return err
	}

	return reviewThenFinalizeLocked(cmd, stdout, stderr, closeF, r, boundaryReviewParams{
		Label:         fmt.Sprintf("#%d %s", f.Issue, f.Milestone),
		Base:          base,
		BaseLong:      baseLong,
		Head:          head,
		IssuesDir:     f.IssuesDir,
		Agent:         f.Agent,
		AgentExplicit: f.AgentExplicit,
		IssueNum:      f.Issue,
		Milestone:     f.Milestone,
		PlansDir:      resolvePlansDir(f.PlansDir),
		PriorFindings: prior,
	}, snapshot)
}

// resolveReviewWindow computes the (base, baseLong, head) tuple for a
// boundary-review window. base is short; baseLong is the full base ref (used by
// the verifier in close.go to locate the same window in `git log`); head is the
// CONCRETE SHA at the operator's pre-close tip — the close hasn't been committed
// yet, so that commit is what gets reviewed.
//
// #194: head used to be the literal string "HEAD", which floats. Three things
// spend it — the dispatched diff, the durable record (Review-Window trailer +
// sidecar), and the finalize check — and none of them can be right against a
// floating ref: the diff is collected AFTER the repo lock is released, so a
// literal "HEAD" there could resolve to a different commit than the snapshot
// recorded, and a finalize check cannot classify a mid-review delta with no
// anchor to measure from. Callers resolve this under the lock and pass the SAME
// value everywhere. The base itself comes from
// boundaryWindowBase: the prior review boundary for a milestone close, or the
// branch start for a whole-issue close / first milestone (see that helper). It
// is the same window source as close.go's atlas gate (ARCH-DRY), so the review
// and the atlas check provably cover the same commits.
//
// Returns ("?", "", "HEAD") when no commit anchors the window (e.g., a docs-only
// milestone with no #N commits) so the trailer still has something to write.
func resolveReviewWindow(issueStr, milestone, issuePath string) (base, baseLong, head string) {
	// Degrade to the literal only when rev-parse cannot answer (empty repo, or a
	// dry-run outside a work tree) — the window still needs SOMETHING to print.
	head = "HEAD"
	if sha := gitx.Capture("rev-parse", "HEAD"); sha != "" {
		head = sha
	}
	baseLong = boundaryWindowBase(issueStr, milestone, issuePath)
	if baseLong == "" {
		return "?", "", head
	}
	base = shortSHA(baseLong)
	return base, baseLong, head
}

// boundaryWindowBase returns the base ref of a close/review window — the commit
// the diff runs *from* (exclusive in `base..HEAD`) — for both the atlas-coverage
// gate (close.go) and the boundary review (milestoneclose.go). Keeping the two
// on one window source means they provably cover the same commits (ARCH-DRY).
//
// Milestone close: the PREVIOUS review boundary — the most recent prior commit
// touching the issue file that carries a Review-Verdict: trailer (the prior
// milestone close). Everything since that boundary, including inter-milestone
// #N-but-not-Mx commits (side-quests, fixes), then lands in exactly one window
// (#58) — previously such commits could slip between the M(x-1) and Mx windows
// and escape review entirely.
//
// Whole-issue close (milestone == ""): the branch point — `merge-base(main,
// HEAD)` — so the end-of-issue integration review covers exactly this branch's
// commits, not unrelated history merged before the issue's first commit (#77).
// On main (no divergence) merge-base == HEAD, so it falls back to the issue's
// branch start. A first milestone (no prior boundary) uses that same feature-
// branch point; only direct-on-main/no-divergence work uses the issue-specific
// fallback. Returns "" when no anchor exists (no #N commit yet).
func boundaryWindowBase(issueStr, milestone, issuePath string) string {
	if milestone != "" {
		// A prior milestone boundary wins. Without one, continue to the shared
		// feature-branch point below (#58/#162).
		if prev := previousReviewBoundary(issuePath); prev != "" {
			return prev
		}
	}
	// Whole-issue close or first milestone: the branch point, else (on main)
	// the issue's own branch start.
	if mb := gitx.MergeBaseWithMain(); mb != "" {
		return mb
	}
	return branchStartByIssue(issueStr)
}

// branchStartByIssue returns the parent of the first commit referencing #N (the
// issue's branch start), or that commit itself when it has no parent
// (initial-commit edge). "" when no #N commit exists. Shared by the milestone
// no-prior-boundary fallback and the whole-issue on-main fallback (ARCH-DRY).
func branchStartByIssue(issueStr string) string {
	firstSHA := firstCommitReferencing("#" + issueStr)
	if firstSHA == "" {
		return ""
	}
	parent := firstSHA + "^"
	resolvedParent := gitx.Capture("rev-parse", "--verify", parent)
	if resolvedParent == "" {
		return firstSHA
	}
	return resolvedParent
}

// previousReviewBoundary returns the SHA of the most recent commit touching
// issuePath whose message carries a Review-Verdict: trailer — the prior
// milestone-close boundary on this branch — or "" if none. Scoping to the issue
// file means only this issue's close commits qualify (the same scoping
// milestoneHasVerdictCommit uses, ARCH-DRY). It is called before the current
// close is committed, so the most-recent match is genuinely the *previous*
// boundary, not this one.
//
// If a prior close's trailer was never pasted into its commit, this finds
// nothing and the caller falls back to the branch start — over-covering
// (re-reviewing prior work) rather than under-covering, the safe direction.
func previousReviewBoundary(issuePath string) string {
	if issuePath == "" {
		return ""
	}
	// Anchored (#194 M2 review): an UNANCHORED --grep matches the commit BODY, so prose
	// merely discussing the trailer becomes a false window base. Commit 23d5b8a in this
	// very issue's history — whose body reads "the Review-Verdict trailer" — was one
	// character away from silently re-basing a review window. Same class as the lesson
	// this issue added to lessons.md: a substring anchor in a self-referential document.
	out, err := gitx.RunGit("log", "--grep=^Review-Verdict:", "--max-count=1", "--pretty=format:%H", "--", issuePath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// issueFilePath resolves the single workshop issue file for issueNum under
// issuesDir (glob NNNNNN-*.md). Errors when zero or multiple files match — the
// same resolution close.go's locate-issue step and annotateLogLineWithVerdict
// rely on, kept in one place (ARCH-DRY).
func issueFilePath(issuesDir string, issueNum int) (string, error) {
	pattern := filepath.Join(issuesDir, fmt.Sprintf("%06d", issueNum)+"-*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob %s: %w", pattern, err)
	}
	sort.Strings(matches)
	if len(matches) == 0 {
		return "", fmt.Errorf("no issue file matches %s", pattern)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("multiple issue files match: %v", matches)
	}
	return matches[0], nil
}

// firstCommitReferencing returns the SHA of the FIRST (oldest) commit whose
// subject contains refSubject (e.g. "#69" — the issue ref). "" when `git log`
// fails or nothing matches. It locates the branch start (that commit's parent)
// for boundaryWindowBase — the single window source (ARCH-DRY) feeding both the
// atlas gate and the boundary review.
//
// Match is a bare substring on the subject: refSubject always starts with '#',
// which never appears in a SHA or ISO date, so subject-only matching is precise.
// (Caveat: "#69" substring-matches "#690" — a theoretical collision; commit
// subjects on a feature branch are the issue's own, so it doesn't bite in
// practice.)
func firstCommitReferencing(refSubject string) string {
	entries, err := gitx.LogReverse()
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if strings.Contains(e.Subject, refSubject) {
			return e.SHA
		}
	}
	return ""
}

// shortSHA returns the abbreviated SHA via `git rev-parse --short`. Falls
// back to manual truncation if rev-parse fails (e.g., the ref doesn't
// resolve — shouldn't happen on the path that calls this but safer than
// returning empty).
func shortSHA(ref string) string {
	if ref == "" {
		return "?"
	}
	if s := gitx.Capture("rev-parse", "--short", ref); s != "" {
		return s
	}
	if len(ref) >= 8 {
		return ref[:8]
	}
	return ref
}

// emitTrailerBlock writes the conventional git-trailer block to stdout
// so the operator/agent can paste it into the milestone-close commit
// message. The block is prefixed with a marker comment so it's easy to
// locate in the captured output.
//
// Shape (per AGENTS.md trailer conventions):
//
//	── milestone-close trailers (paste into commit message) ──
//
//	Review-Verdict: SHIP
//	Review-Window: abc1234..def5678
//	[Review-Reason: --no-judge]   (only when verdict is not-run)
//
// The blank line before the trailers matches git's `interpret-trailers`
// expectation: trailers form a contiguous block at the message bottom,
// separated from the body by one blank line.
func emitTrailerBlock(stdout io.Writer, r reviewResult, kind string) {
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "── %s trailers (paste into commit message) ──\n", kind)
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Review-Verdict: %s\n", r.Verdict)
	// Both ends through the SAME abbreviator so the window reads symmetrically —
	// r.Base is git's minimal-unique short form (often 7), abbrevSHA is a fixed 8.
	base := r.Base
	if r.BaseLong != "" {
		base = abbrevSHA(r.BaseLong)
	}
	fmt.Fprintf(stdout, "Review-Window: %s..%s\n", base, abbrevSHA(r.Head))
	if r.Reason != "" {
		fmt.Fprintf(stdout, "Review-Reason: %s\n", r.Reason)
	}
}

// annotateLogLineWithVerdict re-reads the issue file and appends
// "; review verdict: <verdict>" to the just-written close log line for
// this milestone. Idempotent: if the line already carries a verdict
// suffix (re-run case), it's left alone.
//
// Why post-mutation rather than threading the verdict through the mechanical close:
// computeClose/applyClose run before the judge has a verdict to record. The cleanest
// seam is to let the mechanical close own its log-line shape and let milestone-close
// extend it afterwards. The cost is one extra file read+write; the
// benefit is that close.go doesn't grow a verdict-aware code path that
// only ever fires from this wrapper.
func annotateLogLineWithVerdict(issuesDir string, issueNum int, milestone string, verdict judge.Verdict) error {
	issuePath, err := issueFilePath(issuesDir, issueNum)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(issuePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", issuePath, err)
	}
	updated, ok := appendVerdictSuffix(string(data), milestone, verdict)
	if !ok {
		return fmt.Errorf("no matching '- YYYY-MM-DD: closed %s — ...' line", milestone)
	}
	if updated == string(data) {
		return nil // already annotated, idempotent no-op
	}
	if err := os.WriteFile(issuePath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", issuePath, err)
	}
	return nil
}

// appendVerdictSuffix finds the first log line matching
//
//   - YYYY-MM-DD: closed <milestone> — <verified>
//
// and appends "; review verdict: <verdict>" if it isn't already
// present. Returns (updated, true) when a target line was located,
// (text, false) otherwise. Idempotent on re-runs.
//
// Pure: no IO, deterministic. Lives next to the writer in this package
// rather than in internal/issue/ because it's a milestone-close-specific
// shape (the closed-with-milestone log line, not arbitrary log lines).
func appendVerdictSuffix(text, milestone string, verdict judge.Verdict) (string, bool) {
	lines := strings.Split(text, "\n")
	verdictSuffix := "; review verdict: " + string(verdict)
	// Match "- <date>: closed <milestone> — ..." (milestone close) or
	// "- <date>: closed — ..." (whole-issue close, milestone==""). The close
	// writer emits "closed[ Mx] — <verified>"; mirror that exactly.
	prefix := "closed — "
	if milestone != "" {
		prefix = "closed " + milestone + " — "
	}
	for i, line := range lines {
		if !strings.HasPrefix(line, "- ") {
			continue
		}
		// Format the close writer emits: "- <date>: closed <Mx> — ..."
		// Find ":" after "- " and check the post-colon prefix.
		colon := strings.Index(line, ": ")
		if colon < 0 {
			continue
		}
		rest := line[colon+2:]
		if !strings.HasPrefix(rest, prefix) {
			continue
		}
		// Idempotency guard: line already carries a verdict suffix —
		// don't append a second one.
		if strings.Contains(line, "; review verdict: ") {
			return text, true
		}
		lines[i] = line + verdictSuffix
		return strings.Join(lines, "\n"), true
	}
	return text, false
}

// boundaryReviewParams are the inputs to the one binary-owned boundary review
// (#69). milestone-close passes a per-milestone window (`#69 M1`); close passes
// a whole-issue window (`#69`). Both invoke the same MilestoneReview prompt (the
// embedded code-review.md procedure) on the resolved window.
type boundaryReviewParams struct {
	// The prompt's issue ref is no longer carried here — it's derived from the
	// live git context in boundaryReviewDispatchOptions (#137, via IssueNum +
	// Milestone + the repo root), so it names the actual repo (e.g. pair#69).
	Label                string // commit-subject substring for messages, e.g. "#69 M1"
	Base, BaseLong, Head string // review window (short base, long base, long reviewed head — #194)
	IssuesDir            string
	Agent                string
	AgentExplicit        bool
	// Sidecar persistence (#136): the issue id + milestone + plans dir needed to
	// name and write the durable final review response. Milestone "" ⇒ whole-issue close.
	IssueNum  int
	Milestone string
	PlansDir  string
	// ReviewPlansDir carries the read-only plan discovery root across the unlocked
	// dispatch while PlansDir is blanked to defer sidecar writes until relock.
	ReviewPlansDir string
	// ForcedRationale records a bypass of the ledger's refusal (--no-ledger / --force),
	// stamped onto the round so the durable record does not read a waived refusal as a
	// clean pass (#194 close review BR-17/BR-39). NOT the gate_forced metric — that reads
	// the plan-gate ledger only (churnreport.go); the boundary gate has no metric consumer
	// yet, and claiming one was part of what made the first version of this inert.
	ForcedRationale string
	// PriorFindings is the rendered prior-round block for THIS boundary (#194 M2),
	// read by the caller UNDER THE REPO LOCK and carried here rather than fetched at
	// dispatch time.
	//
	// M2's boundary review found the reason this is a field and not a lookup: dispatch
	// runs with the lock released, and reviewThenFinalizeLocked blanks PlansDir there to
	// defer the sidecar write — so a dispatch-time read keyed on PlansDir returned ""
	// in 100% of live reviews. The ledger then blocked on findings the reviewer was
	// never shown and therefore could not dispose. Reading it beside
	// captureCloseReviewSnapshot also makes the ledger and the snapshot provably see the
	// same repo state, which is M1's anchor argument applied one artifact over.
	PriorFindings string
}

func printBoundaryReviewDryRun(stdout, stderr io.Writer, p boundaryReviewParams) error {
	opts, ok, reason := boundaryReviewDispatchOptions(stdout, stderr, p)
	if !ok {
		return errors.New(reason)
	}
	cmdLine, err := judge.FormatCommandLine(opts)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, "── command (would invoke) ──")
	fmt.Fprintln(stdout, cmdLine)
	return nil
}

// dispatchBoundaryReview invokes the one fresh-context review on p's window.
// Returns a reviewResult capturing the verdict + reason. Never returns an error:
// the close has already happened; the review is a follow-on, so any failure here
// is recorded as VerdictNotRun with a Reason and the caller still emits a trailer.
func dispatchBoundaryReview(stdout, stderr io.Writer, p boundaryReviewParams) reviewResult {
	res := func(v judge.Verdict, reason string) reviewResult {
		// ProtocolError distinguishes "the reviewer ran and emitted no fence" from "the
		// review never started" — both reach persistBoundaryRound with Round == nil, and
		// a ledger that cannot tell them apart mis-reports why a round contributed
		// nothing (#194 M2 review).
		return reviewResult{Verdict: v, Reason: reason, Base: p.Base, Head: p.Head, BaseLong: p.BaseLong,
			ProtocolError: "review did not run: " + reason}
	}
	opts, ok, reason := boundaryReviewDispatchOptions(stdout, stderr, p)
	if !ok {
		// The caller (reviewThenFinalize) maps this VerdictNotRun → closeHalt and
		// prints the outcome ("close NOT finalized") — so don't claim success here (#139 I1).
		cwarn(stderr, "boundary review could not run: "+reason)
		return res(judge.VerdictNotRun, reason)
	}

	agent := opts.Agent
	cinfo(stderr, fmt.Sprintf("dispatching boundary review (%s..%s) via %s …", shortSHA(p.BaseLong), abbrevSHA(p.Head), agent))
	output, derr := judge.Dispatch(context.Background(), opts)
	if derr != nil {
		// Dispatch error → VerdictNotRun → the caller halts (does NOT finalize); the
		// outcome message is the caller's, not a false "close succeeded" here (#139 I1).
		cwarn(stderr, fmt.Sprintf("boundary review failed: %v", derr))
		return res(judge.VerdictNotRun, derr.Error())
	}
	fmt.Fprint(stdout, output)
	if !strings.HasSuffix(output, "\n") {
		fmt.Fprintln(stdout)
	}
	switch judge.Classify(output) {
	case judge.Clean:
		cok(stderr, "boundary review: clean")
	case judge.Info:
		cinfo(stderr, "boundary review: info")
	case judge.Failure:
		cwarn(stderr, "boundary review: findings reported — address before crossing the boundary"+fixTheClassNote())
	}
	verdict := judge.ParseVerdict(output)
	if verdict == judge.VerdictUnknown {
		cwarn(stderr, fmt.Sprintf("boundary review: no '%s' verdict found (block or line) — recording verdict as 'unknown'",
			strings.Join(vocab.Verdict().Emitted(), " | ")))
	}
	rr := reviewResult{Verdict: verdict, Base: p.Base, Head: p.Head, BaseLong: p.BaseLong, Output: output, Agent: string(agent)}
	// #194 M2: capture the structured findings handoff. Persisting it is the caller's
	// job (it holds the lock); parsing here keeps the raw output in one place.
	if round, ok := gatestate.ParseFindingsBlock(output); ok {
		rr.Round = &round
	} else {
		rr.ProtocolError = "no valid findings block"
	}
	// Persist the semantic final review response to a durable sidecar (#136) so an agent can
	// reopen it after scrollback loss / compaction. Non-fatal: the review already
	// ran, so a write failure is warned, not propagated (matches the philosophy above).
	// Record the RESOLVED reviewer (opts.Agent), not the raw --agent flag — the
	// latter defaults to "" so the sidecar's reviewer cell would otherwise be empty.
	p.Agent = string(agent)
	if p.PlansDir != "" {
		if path, werr := writeReviewSidecar(p, string(verdict), output, nowRFC3339()); werr != nil {
			cwarn(stderr, fmt.Sprintf("review sidecar not written: %v", werr))
		} else {
			rr.SidecarPath = path
			cok(stderr, "review sidecar: "+path)
		}
	}
	return rr
}

func boundaryReviewDispatchOptions(stdout, stderr io.Writer, p boundaryReviewParams) (judge.DispatchOptions, bool, string) {
	if p.BaseLong == "" {
		return judge.DispatchOptions{}, false, fmt.Sprintf("no commits reference '%s' — cannot determine review window", p.Label)
	}

	// p.Head, not "HEAD" (#194): this runs after reviewThenFinalizeLocked released
	// the lock, so re-resolving here could collect a DIFFERENT commit than the
	// snapshot pinned — the review would then be attributed to a commit it never read.
	plansDir := p.ReviewPlansDir
	if plansDir == "" {
		plansDir = p.PlansDir
	}
	manifest, err := resolveBoundaryReviewManifest(liveBoundaryGit{}, boundaryReviewManifestRequest{
		BaseRef: p.BaseLong, HeadRef: p.Head,
		IssuesDir: p.IssuesDir, HistoryDir: "workshop/history",
		IssueNum: p.IssueNum, PlansDir: plansDir,
	})
	if err != nil {
		return judge.DispatchOptions{}, false, fmt.Sprintf("resolve review manifest: %v", err)
	}
	// Automatic close captures concrete anchors under the repo lock. Resolving a
	// symbolic ref here would reintroduce the moving-window race #194 removed.
	if manifest.BaseSHA != p.BaseLong || manifest.HeadSHA != p.Head {
		return judge.DispatchOptions{}, false, "automatic boundary review requires concrete base/head commit object ids"
	}
	reviewWindow, err := judge.RenderReviewWindow(manifest)
	if err != nil {
		return judge.DispatchOptions{}, false, fmt.Sprintf("render review manifest: %v", err)
	}

	// Orient the fresh reviewer to the ACTUAL repo (#137) — derived from the live
	// git context, not a hardcoded "ariadne". Computed once here, the single site
	// both close and milestone-close funnel through (ARCH-DRY).
	o := boundaryOrientation(p.IssuesDir, p.IssueNum, p.Milestone)
	o.RepoRoot = manifest.RepoRoot
	o.Repo = filepath.Base(manifest.RepoRoot)
	o.IssueFile = manifest.IssueFile
	in := judge.PromptInput{
		ReviewWindow: reviewWindow, Base: manifest.BaseSHA, Head: manifest.HeadSHA,
		IssueRef: o.IssueRef, Repo: o.Repo, RepoRoot: o.RepoRoot,
		IssueFile: o.IssueFile, Boundary: o.Boundary, RepoNote: o.RepoNote,
		PriorFindings: p.PriorFindings,
	}
	prompt := judge.BuildPrompt(judge.MilestoneReview, in)

	return judge.DispatchOptions{
		Agent:        judge.ResolveAgentCLI(p.Agent, p.AgentExplicit, judge.CurrentAgentDefaultEnv()),
		Prompt:       prompt,
		AllowedTools: judge.MilestoneReview.AllowedTools(),
		IsSandbox:    isSandbox(),
		Stdout:       stdout,
		Stderr:       stderr,
	}, true, ""
}
