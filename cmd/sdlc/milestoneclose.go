// milestoneclose.go — `sdlc milestone-close` subcommand.
//
// Thin wrapper over `sdlc close --milestone Mx` that adds the
// AGENTS.md §3 mandatory post-milestone code review as an auto-dispatched
// follow-on: after the milestone close completes, fires a fresh-context
// `judge milestone-review` against the commit window for the milestone.
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
	cmd.Flags().BoolVar(&f.Force, "force", false, "bypass guards (ACTUAL/VERIFIED/atlas/plan)")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "plan only; do not write or dispatch judge")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip the auto-dispatched milestone-review")
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
		Issue:     f.Issue,
		Milestone: f.Milestone,
		Actual:    f.Actual,
		Verified:  f.Verified,
		Force:     f.Force,
		DryRun:    f.DryRun,
		BrainDir:  f.BrainDir,
		IssuesDir: f.IssuesDir,
	}
	if err := runClose(stderr, closeF); err != nil {
		return err
	}

	// Step 2: auto-dispatch judge milestone-review on the commit window.
	if f.NoJudge {
		cinfo(stderr, "skipping milestone-review per --no-judge")
		return nil
	}
	if f.DryRun {
		cinfo(stderr, "dry-run — would dispatch judge milestone-review")
		return nil
	}

	if err := dispatchMilestoneReview(stdout, stderr, f); err != nil {
		// Don't fail the close — the milestone is already closed; the
		// review is a follow-on. Warn and let the operator re-run
		// `sdlc judge milestone-review --base SHA --head HEAD ...`
		// manually if they want.
		cwarn(stderr, fmt.Sprintf("milestone-review dispatch failed: %v", err))
		cwarn(stderr, "milestone close succeeded; re-run judge manually if needed")
		return nil
	}
	return nil
}

// dispatchMilestoneReview finds the first commit referencing
// `#<issue> <milestone>` and invokes the judge with that as the base.
// Returns an error if no commits matched the milestone (nothing to
// review) or the judge subprocess fails to launch.
func dispatchMilestoneReview(stdout, stderr io.Writer, f *milestoneCloseFlags) error {
	refSubject := fmt.Sprintf("#%d %s", f.Issue, f.Milestone)

	entries, err := gitx.LogReverse()
	if err != nil {
		return fmt.Errorf("git log: %w", err)
	}
	var firstSHA string
	for _, e := range entries {
		if strings.Contains(e.Subject, refSubject) {
			firstSHA = e.SHA
			break
		}
	}
	if firstSHA == "" {
		return fmt.Errorf("no commits reference %q — cannot determine review window", refSubject)
	}

	// Diff window: parent of first-matching..HEAD. Matches the way the
	// commit window is computed in close.go's atlas check.
	base := firstSHA + "^"
	// Verify the parent exists (initial-commit edge case).
	if gitx.Capture("rev-parse", "--verify", base) == "" {
		base = firstSHA
	}

	diff, _, err := collectDiff(judge.MilestoneReview, base, "HEAD", f.IssuesDir, "workshop/history")
	if err != nil {
		return fmt.Errorf("collect diff: %w", err)
	}

	in := judge.PromptInput{
		Diff:     diff,
		Base:     base,
		Head:     "HEAD",
		IssueRef: fmt.Sprintf("ariadne#%d %s", f.Issue, f.Milestone),
	}
	prompt := judge.BuildPrompt(judge.MilestoneReview, in)

	agent := judge.AgentCLI(orStr(f.Agent, "claude"))
	opts := judge.DispatchOptions{
		Agent:        agent,
		Prompt:       prompt,
		AllowedTools: judge.MilestoneReview.AllowedTools(),
		IsSandbox:    isSandbox(),
		Stdout:       stdout,
		Stderr:       stderr,
	}

	cinfo(stderr, fmt.Sprintf("dispatching milestone-review (%s..HEAD) via %s …", base, agent))
	output, derr := judge.Dispatch(context.Background(), opts)
	if derr != nil {
		return derr
	}
	fmt.Fprint(stdout, output)
	if !strings.HasSuffix(output, "\n") {
		fmt.Fprintln(stdout)
	}
	switch judge.Classify(output) {
	case judge.Clean:
		cok(stderr, "milestone-review: clean")
	case judge.Info:
		cinfo(stderr, "milestone-review: info")
	case judge.Failure:
		cwarn(stderr, "milestone-review: findings reported — address before next milestone")
	}
	return nil
}
