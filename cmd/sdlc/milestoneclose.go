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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

type milestoneCloseFlags struct {
	Issue     int
	Milestone string
	Actual    string
	Verified  string
	Force     bool
	DryRun    bool
	NoJudge   bool   // skip the auto-dispatched milestone-review
	Agent     string // forwarded to the judge dispatch
	BrainDir  string
	IssuesDir string

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
	Verdict  judge.Verdict
	Reason   string // populated for not-run / unknown
	Base     string // short SHA
	Head     string // short SHA ("HEAD" fine in dry-run)
	BaseLong string // long SHA, used by trailer-verifier lookups in close
}

func NewMilestoneCloseCmd() *cobra.Command {
	f := milestoneCloseFlags{}
	cmd := &cobra.Command{
		Use:           "milestone-close",
		Short:         "Close one milestone of an issue + auto-dispatch post-milestone review (AGENTS.md §3)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"milestone-close\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMilestoneClose(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	cmd.Flags().IntVar(&f.Issue, "issue", 0, "ariadne workshop issue ID (required, positive)")
	cmd.Flags().StringVar(&f.Milestone, "milestone", "", "milestone tag e.g. M4 (required)")
	cmd.Flags().StringVar(&f.Actual, "actual", "", "focused dev-hours for this milestone")
	cmd.Flags().StringVar(&f.Verified, "verified", "", "one-line evidence the milestone meets done-when")
	cmd.Flags().BoolVar(&f.Force, "force", false, "bypass ALL close gates (≡ every --no-* flag); reason in --verified")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "plan only; do not write or dispatch judge")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip the auto-dispatched milestone-review")
	// Per-gate close bypasses (#67) — forwarded to runClose; --force waives all.
	cmd.Flags().BoolVar(&f.NoActual, "no-actual", false, "bypass the ACTUAL-hours requirement")
	cmd.Flags().BoolVar(&f.NoVerified, "no-verified", false, "bypass the VERIFIED-evidence requirement")
	cmd.Flags().BoolVar(&f.NoReclose, "no-reclose-guard", false, "bypass the already-done refusal")
	cmd.Flags().BoolVar(&f.NoAtlas, "no-atlas", false, "bypass the atlas/ change check (no new surface)")
	cmd.Flags().BoolVar(&f.NoVerdict, "no-verdict", false, "bypass the milestone Review-Verdict check")
	cmd.Flags().BoolVar(&f.NoPlanCheck, "no-plan-check", false, "bypass the unchecked-## Plan-items refusal")
	cmd.Flags().BoolVar(&f.NoProject, "no-project", false, "bypass the project detail-block update requirement")
	cmd.Flags().StringVar(&f.Agent, "agent", envOr("AGENT_CMD", ""), "agent CLI for judge dispatch (claude | codex | gemini)")
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
		Issue:       f.Issue,
		Milestone:   f.Milestone,
		Actual:      f.Actual,
		Verified:    f.Verified,
		Force:       f.Force,
		DryRun:      f.DryRun,
		BrainDir:    f.BrainDir,
		IssuesDir:   f.IssuesDir,
		NoActual:    f.NoActual,
		NoVerified:  f.NoVerified,
		NoReclose:   f.NoReclose,
		NoAtlas:     f.NoAtlas,
		NoVerdict:   f.NoVerdict,
		NoPlanCheck: f.NoPlanCheck,
		NoProject:   f.NoProject,
	}
	if err := runClose(stderr, closeF); err != nil {
		return err
	}

	// Step 2: figure out the review window (used regardless of whether
	// the judge actually runs — the trailer always carries it).
	base, baseLong, head := resolveReviewWindow(fmt.Sprintf("#%d %s", f.Issue, f.Milestone))

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
			IssueRef:  fmt.Sprintf("ariadne#%d %s", f.Issue, f.Milestone),
			Label:     fmt.Sprintf("#%d %s", f.Issue, f.Milestone),
			Base:      base,
			BaseLong:  baseLong,
			Head:      head,
			IssuesDir: f.IssuesDir,
			Agent:     f.Agent,
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
// boundary-review window, given the commit-subject substring that the
// boundary's commits carry (`#69 M1` for a milestone, `#69` for a whole
// issue). base is short, baseLong is the full 40-char SHA (used by the
// verifier in close.go to locate the same window in `git log`), head is
// "HEAD" — the close hasn't been committed yet, so HEAD is the operator's
// pre-close tip and the diff is what got reviewed. The base is the parent
// of the FIRST commit referencing the subject — the same window source as
// close.go's atlas gate (ARCH-DRY), so the review and the atlas check agree.
//
// Returns ("?", "", "HEAD") when no commit references the subject (e.g., a
// docs-only milestone with no code commits) so the trailer still has
// something to write.
func resolveReviewWindow(refSubject string) (base, baseLong, head string) {
	head = "HEAD"
	entries, err := gitx.LogReverse()
	if err != nil {
		return "?", "", head
	}
	var firstSHA string
	for _, e := range entries {
		if strings.Contains(e.Subject, refSubject) {
			firstSHA = e.SHA
			break
		}
	}
	if firstSHA == "" {
		return "?", "", head
	}
	baseLong = firstSHA + "^"
	// Verify the parent exists (initial-commit edge case).
	if gitx.Capture("rev-parse", "--verify", baseLong) == "" {
		baseLong = firstSHA
	}
	base = shortSHA(baseLong)
	return base, baseLong, head
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
	issueID := fmt.Sprintf("%06d", issueNum)
	pattern := filepath.Join(issuesDir, issueID+"-*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob %s: %w", pattern, err)
	}
	sort.Strings(matches)
	if len(matches) == 0 {
		return fmt.Errorf("no issue file matches %s", pattern)
	}
	if len(matches) > 1 {
		return fmt.Errorf("multiple issue files match: %v", matches)
	}
	issuePath := matches[0]
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
	IssueRef             string // prompt label, e.g. "ariadne#69" or "ariadne#69 M1"
	Label                string // commit-subject substring for messages, e.g. "#69 M1"
	Base, BaseLong, Head string // review window (short base, long base, "HEAD")
	IssuesDir            string
	Agent                string
}

// dispatchBoundaryReview invokes the one fresh-context review on p's window.
// Returns a reviewResult capturing the verdict + reason. Never returns an error:
// the close has already happened; the review is a follow-on, so any failure here
// is recorded as VerdictNotRun with a Reason and the caller still emits a trailer.
func dispatchBoundaryReview(stdout, stderr io.Writer, p boundaryReviewParams) reviewResult {
	res := func(v judge.Verdict, reason string) reviewResult {
		return reviewResult{Verdict: v, Reason: reason, Base: p.Base, Head: p.Head, BaseLong: p.BaseLong}
	}
	if p.BaseLong == "" {
		reason := fmt.Sprintf("no commits reference '%s' — cannot determine review window", p.Label)
		cwarn(stderr, "boundary review skipped: "+reason)
		cwarn(stderr, "close succeeded; re-run judge manually if needed")
		return res(judge.VerdictNotRun, reason)
	}

	diff, _, err := collectDiff(judge.MilestoneReview, p.BaseLong, "HEAD", p.IssuesDir, "workshop/history")
	if err != nil {
		reason := fmt.Sprintf("collect diff: %v", err)
		cwarn(stderr, "boundary review failed: "+reason)
		cwarn(stderr, "close succeeded; re-run judge manually if needed")
		return res(judge.VerdictNotRun, reason)
	}

	in := judge.PromptInput{Diff: diff, Base: p.BaseLong, Head: "HEAD", IssueRef: p.IssueRef}
	prompt := judge.BuildPrompt(judge.MilestoneReview, in)

	agent := judge.AgentCLI(orStr(p.Agent, "claude"))
	opts := judge.DispatchOptions{
		Agent:        agent,
		Prompt:       prompt,
		AllowedTools: judge.MilestoneReview.AllowedTools(),
		IsSandbox:    isSandbox(),
		Stdout:       stdout,
		Stderr:       stderr,
	}

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
	return reviewResult{Verdict: verdict, Base: p.Base, Head: p.Head, BaseLong: p.BaseLong}
}
