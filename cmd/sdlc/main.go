// Command sdlc is the unified SDLC checkpoint binary for ariadne.
//
// Subcommands are checkpoint guards — they defend known commit moments
// (close, merge, push, milestone-close) against drift. Subcommands are
// added incrementally when the same drift recurs at a stage; the binary
// does not model the SDLC as a state machine.
//
// Help disclosure is progressive — `sdlc --help` is the single workflow
// contract (the xx-sdlc skill is a static pointer to it):
//
//	sdlc --help              top-level workflow contract + verb list
//	sdlc <verb> --help       per-checkpoint contract + flags + examples
//
// Design rationale: workshop/issues/000031-sdlc-checkpoint-binary.md +
// docs/vision/2026-05-25-01-pensive-sdlc-checkpoint-binary.md.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/helptext"
)

func main() {
	if err := buildRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// buildRoot assembles the full cobra command tree. Extracted from main so
// the tree itself is testable (e.g. the flat→group alias wiring for
// set-status / fetch — #56 M2).
func buildRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "sdlc",
		Short:         "SDLC checkpoint binary — guards known commit moments against drift",
		Long:          helptext.MustGet("root"),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.RunE = func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	}

	closeCmd := NewCloseCmd()
	closeCmd.Long = helptext.MustGet("close")
	root.AddCommand(closeCmd)

	stateCmd := NewStateCmd()
	stateCmd.Long = helptext.MustGet("state")
	root.AddCommand(stateCmd)

	judgeCmd := NewJudgeCmd()
	judgeCmd.Long = helptext.MustGet("judge")
	root.AddCommand(judgeCmd)

	// fetch folded into `sdlc issue new --from-github` (#56 M2). Hidden,
	// deprecated alias kept for one cycle (retains its --github-issue flag).
	fetchCmd := NewFetchCmd()
	fetchCmd.Long = helptext.MustGet("fetch")
	fetchCmd.Hidden = true
	fetchCmd.Deprecated = "use `sdlc issue new --from-github N`"
	root.AddCommand(fetchCmd)

	// `sdlc start` is a hidden stub that errors with a migration
	// message (#39). No Long help wired — Short + RunE carry the
	// message. Helptext file removed.
	startCmd := NewStartCmd()
	root.AddCommand(startCmd)

	claimCmd := NewClaimCmd()
	claimCmd.Long = helptext.MustGet("claim")
	root.AddCommand(claimCmd)

	changeCodeCmd := NewChangeCodeCmd()
	changeCodeCmd.Long = helptext.MustGet("change-code")
	root.AddCommand(changeCodeCmd)

	// set-status moved under `sdlc issue` (#56 M2). Keep a hidden,
	// deprecated flat alias for one cycle so existing `sdlc set-status`
	// callers + references keep working while they migrate.
	flatSetStatus := NewSetStatusCmd()
	flatSetStatus.Long = helptext.MustGet("set-status")
	flatSetStatus.Hidden = true
	flatSetStatus.Deprecated = "use `sdlc issue set-status`"
	root.AddCommand(flatSetStatus)

	pushCmd := NewPushCmd()
	pushCmd.Long = helptext.MustGet("push")
	root.AddCommand(pushCmd)

	prCmd := NewPRCmd()
	prCmd.Long = helptext.MustGet("pr")
	root.AddCommand(prCmd)

	mergeCmd := NewMergeCmd()
	mergeCmd.Long = helptext.MustGet("merge")
	root.AddCommand(mergeCmd)

	milestoneCloseCmd := NewMilestoneCloseCmd()
	milestoneCloseCmd.Long = helptext.MustGet("milestone-close")
	root.AddCommand(milestoneCloseCmd)

	issueCmd := NewIssueCmd()
	issueCmd.Long = helptext.MustGet("issue")
	root.AddCommand(issueCmd)

	return root
}
