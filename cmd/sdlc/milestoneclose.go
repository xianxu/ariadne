// milestoneclose.go — `sdlc milestone-close` subcommand.
//
// Thin wrapper over `sdlc close --milestone Mx` that adds the
// AGENTS.md §3 mandatory post-milestone code review as an auto-dispatched
// follow-on: after the milestone close completes, fires the one binary-owned
// boundary review (dispatchBoundaryReview, shared with `sdlc close` since #69)
// against the commit window for the milestone.
//
// Promotes milestone close from "a flag on close" to its own verb so the
// auto-dispatch is implicit. `sdlc close --milestone Mx` still works
// (operators may want it without the auto-judge), but the canonical
// closing flow is `sdlc milestone-close`.
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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
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

	// Per-gate close bypasses (#67), threaded into the delegated runClose.
	NoActual    bool
	NoVerified  bool
	NoReclose   bool
	NoAtlas     bool
	NoVerdict   bool
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
	Head        string // short SHA ("HEAD" fine in dry-run)
	BaseLong    string // long SHA, used by trailer-verifier lookups in close
	SidecarPath string // #136: durable review transcript path ("" when no review ran)
}

func NewMilestoneCloseCmd() *cobra.Command {
	f := milestoneCloseFlags{}
	cmd := markMutatingCommand(&cobra.Command{
		Use:           "milestone-close",
		Short:         "Close one milestone of an issue + auto-dispatch post-milestone review (AGENTS.md §3)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"milestone-close\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			f.AgentExplicit = cmd.Flags().Changed("agent")
			return runMilestoneClose(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "ariadne workshop issue ID (required, positive)")
	cmd.Flags().StringVar(&f.Milestone, "milestone", "", "milestone tag e.g. M4 (required)")
	cmd.Flags().StringVar(&f.Actual, "actual", "", "focused dev-hours for this milestone")
	cmd.Flags().StringVar(&f.Verified, "verified", "", "one-line evidence the milestone meets done-when")
	cmd.Flags().BoolVar(&f.Force, "force", false, "bypass ALL close gates (≡ every --no-* flag); reason in --verified")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "plan only; do not write or dispatch judge")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip the auto-dispatched milestone-review")
	// Per-gate close bypasses (#67) — forwarded to runClose; --force waives all.
	cmd.Flags().BoolVar(&f.NoActual, "no-actual", false, "record actual_hours: N/A on issue close / skip actual on milestone close")
	cmd.Flags().BoolVar(&f.NoVerified, "no-verified", false, "bypass the VERIFIED-evidence requirement")
	cmd.Flags().BoolVar(&f.NoReclose, "no-reclose-guard", false, "bypass the already-done refusal")
	cmd.Flags().BoolVar(&f.NoAtlas, "no-atlas", false, "bypass the atlas/ change check (no new surface)")
	cmd.Flags().BoolVar(&f.NoVerdict, "no-verdict", false, "bypass the milestone Review-Verdict check")
	cmd.Flags().BoolVar(&f.NoPlanCheck, "no-plan-check", false, "bypass the unchecked-## Plan-items refusal")
	cmd.Flags().BoolVar(&f.NoProject, "no-project", false, "bypass the project detail-block update requirement")
	cmd.Flags().StringVar(&f.Agent, "agent", "", "agent CLI for judge dispatch (claude | codex | gemini)")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "path to the brain repo (for project-file lookup)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	return cmd
}

func runMilestoneClose(stdout, stderr io.Writer, f *milestoneCloseFlags) error {
	if f.Milestone == "" {
		die(stderr, "--milestone is required for milestone-close (use `sdlc close` without it for full-issue close)")
	}
	if f.Issue <= 0 {
		die(stderr, fmt.Sprintf("--issue is required and must be positive (got %d)", f.Issue))
	}

	// Step 1: delegate the mechanical close to runClose.
	closeF := &closeFlags{
		Issue:         f.Issue,
		Milestone:     f.Milestone,
		Actual:        f.Actual,
		Verified:      f.Verified,
		Force:         f.Force,
		DryRun:        f.DryRun,
		BrainDir:      f.BrainDir,
		IssuesDir:     f.IssuesDir,
		Agent:         f.Agent,
		AgentExplicit: f.AgentExplicit,
		NoActual:      f.NoActual,
		NoVerified:    f.NoVerified,
		NoReclose:     f.NoReclose,
		NoAtlas:       f.NoAtlas,
		NoVerdict:     f.NoVerdict,
		NoPlanCheck:   f.NoPlanCheck,
		NoProject:     f.NoProject,
	}
	if err := runClose(stderr, closeF); err != nil {
		return err
	}

	// Step 2: figure out the review window (used regardless of whether
	// the judge actually runs — the trailer always carries it). The base is
	// the prior review boundary so inter-milestone #N-but-not-Mx commits are
	// covered (#58); resolving needs the issue file to find that boundary.
	issuePath, perr := issueFilePath(f.IssuesDir, f.Issue)
	if perr != nil {
		cwarn(stderr, fmt.Sprintf("resolve issue file for review window: %v", perr))
	}
	base, baseLong, head := resolveReviewWindow(strconv.Itoa(f.Issue), f.Milestone, issuePath)

	// Step 3: dispatch the judge (or short-circuit if skipped).
	var result reviewResult
	switch {
	case f.NoJudge:
		cinfo(stderr, "skipping milestone-review per --no-judge")
		result = reviewResult{Verdict: judge.VerdictNotRun, Reason: "--no-judge", Base: base, Head: head, BaseLong: baseLong}
	case f.DryRun:
		cinfo(stderr, "dry-run — would dispatch judge milestone-review")
		result = reviewResult{Verdict: judge.VerdictNotRun, Reason: "--dry-run", Base: base, Head: head, BaseLong: baseLong}
	default:
		result = dispatchBoundaryReview(stdout, stderr, boundaryReviewParams{
			Label:         fmt.Sprintf("#%d %s", f.Issue, f.Milestone),
			Base:          base,
			BaseLong:      baseLong,
			Head:          head,
			IssuesDir:     f.IssuesDir,
			Agent:         f.Agent,
			AgentExplicit: f.AgentExplicit,
			IssueNum:      f.Issue,
			Milestone:     f.Milestone,
			PlansDir:      envOr("WF_PLANS_DIR", "workshop/plans"),
		})
	}

	// Step 4: emit the trailer block to stdout (the agent pastes this
	// into the close commit message; close.go's verifier later greps
	// for Review-Verdict: to confirm review evidence per milestone).
	emitTrailerBlock(stdout, result, "milestone-close")

	// Step 5: mirror the verdict into the issue file's just-written log
	// line so a human grep finds it. Skip in --dry-run (file wasn't
	// written) and on hard failures (the log line may not exist).
	if !f.DryRun {
		if err := annotateLogLineWithVerdict(f.IssuesDir, f.Issue, f.Milestone, result.Verdict); err != nil {
			cwarn(stderr, fmt.Sprintf("log-line verdict annotation skipped: %v", err))
		}
	}

	return nil
}

// resolveReviewWindow computes the (base, baseLong, head) tuple for a
// boundary-review window. base is short; baseLong is the full base ref (used by
// the verifier in close.go to locate the same window in `git log`); head is
// "HEAD" — the close hasn't been committed yet, so HEAD is the operator's
// pre-close tip and the diff is what got reviewed. The base itself comes from
// boundaryWindowBase: the prior review boundary for a milestone close, or the
// branch start for a whole-issue close / first milestone (see that helper). It
// is the same window source as close.go's atlas gate (ARCH-DRY), so the review
// and the atlas check provably cover the same commits.
//
// Returns ("?", "", "HEAD") when no commit anchors the window (e.g., a docs-only
// milestone with no #N commits) so the trailer still has something to write.
func resolveReviewWindow(issueStr, milestone, issuePath string) (base, baseLong, head string) {
	head = "HEAD"
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
// branch start. The first-milestone fallback (no prior boundary) also uses the
// branch start. Returns "" when no anchor exists (no #N commit yet).
func boundaryWindowBase(issueStr, milestone, issuePath string) string {
	if milestone != "" {
		// Milestone close: the prior review boundary, else the branch start (#58).
		if prev := previousReviewBoundary(issuePath); prev != "" {
			return prev
		}
		return branchStartByIssue(issueStr)
	}
	// Whole-issue close: the branch point, else (on main) the issue's branch start.
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
	if gitx.Capture("rev-parse", "--verify", parent) == "" {
		return firstSHA
	}
	return parent
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
	out, err := gitx.RunGit("log", "--grep=Review-Verdict:", "--max-count=1", "--pretty=format:%H", "--", issuePath)
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
//	Review-Window: abc1234..HEAD
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
	fmt.Fprintf(stdout, "Review-Window: %s..%s\n", r.Base, r.Head)
	if r.Reason != "" {
		fmt.Fprintf(stdout, "Review-Reason: %s\n", r.Reason)
	}
}

// annotateLogLineWithVerdict re-reads the issue file and appends
// "; review verdict: <verdict>" to the just-written close log line for
// this milestone. Idempotent: if the line already carries a verdict
// suffix (re-run case), it's left alone.
//
// Why post-mutation rather than threading the verdict through runClose:
// runClose runs before the judge has a verdict to record. The cleanest
// seam is to let runClose own its log-line shape and let milestone-close
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
	Base, BaseLong, Head string // review window (short base, long base, "HEAD")
	IssuesDir            string
	Agent                string
	AgentExplicit        bool
	// Sidecar persistence (#136): the issue id + milestone + plans dir needed to
	// name and write the durable review transcript. Milestone "" ⇒ whole-issue close.
	IssueNum  int
	Milestone string
	PlansDir  string
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
		return reviewResult{Verdict: v, Reason: reason, Base: p.Base, Head: p.Head, BaseLong: p.BaseLong}
	}
	opts, ok, reason := boundaryReviewDispatchOptions(stdout, stderr, p)
	if !ok {
		cwarn(stderr, "boundary review skipped: "+reason)
		cwarn(stderr, "close succeeded; re-run judge manually if needed")
		return res(judge.VerdictNotRun, reason)
	}

	agent := opts.Agent
	cinfo(stderr, fmt.Sprintf("dispatching boundary review (%s..HEAD) via %s …", p.BaseLong, agent))
	output, derr := judge.Dispatch(context.Background(), opts)
	if derr != nil {
		cwarn(stderr, fmt.Sprintf("boundary review failed: %v", derr))
		cwarn(stderr, "close succeeded; re-run judge manually if needed")
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
		cwarn(stderr, "boundary review: findings reported — address before crossing the boundary")
	}
	verdict := judge.ParseVerdict(output)
	if verdict == judge.VerdictUnknown {
		cwarn(stderr, "boundary review: no leading 'SHIP | FIX-THEN-SHIP | REWORK' verdict found — recording verdict as 'unknown'")
	}
	rr := reviewResult{Verdict: verdict, Base: p.Base, Head: p.Head, BaseLong: p.BaseLong}
	// Persist the full transcript to a durable sidecar (#136) so an agent can
	// reopen it after scrollback loss / compaction. Non-fatal: the review already
	// ran, so a write failure is warned, not propagated (matches the philosophy above).
	// Record the RESOLVED reviewer (opts.Agent), not the raw --agent flag — the
	// latter defaults to "" so the sidecar's reviewer cell would otherwise be empty.
	p.Agent = string(agent)
	if path, werr := writeReviewSidecar(p, string(verdict), output, nowRFC3339()); werr != nil {
		cwarn(stderr, fmt.Sprintf("review sidecar not written: %v", werr))
	} else {
		rr.SidecarPath = path
		cok(stderr, "review sidecar: "+path)
	}
	return rr
}

func boundaryReviewDispatchOptions(stdout, stderr io.Writer, p boundaryReviewParams) (judge.DispatchOptions, bool, string) {
	if p.BaseLong == "" {
		return judge.DispatchOptions{}, false, fmt.Sprintf("no commits reference '%s' — cannot determine review window", p.Label)
	}

	diff, _, err := collectDiff(judge.MilestoneReview, p.BaseLong, "HEAD", p.IssuesDir, "workshop/history")
	if err != nil {
		return judge.DispatchOptions{}, false, fmt.Sprintf("collect diff: %v", err)
	}

	// Orient the fresh reviewer to the ACTUAL repo (#137) — derived from the live
	// git context, not a hardcoded "ariadne". Computed once here, the single site
	// both close and milestone-close funnel through (ARCH-DRY).
	o := boundaryOrientation(p.IssuesDir, p.IssueNum, p.Milestone)
	in := judge.PromptInput{
		Diff: diff, Base: p.BaseLong, Head: "HEAD",
		IssueRef: o.IssueRef, Repo: o.Repo, RepoRoot: o.RepoRoot,
		IssueFile: o.IssueFile, Boundary: o.Boundary, RepoNote: o.RepoNote,
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
