// startplan.go — `sdlc start-plan`: the workflow's planning-entry transition
// (#75 M2). The flow had `claim` (start work) and `change-code` (the plan-quality
// *review* gate, which is too late — the design is already made) but no marker
// for "I'm now designing". start-plan fills it: it delivers architecture.md's
// `at-plan` lens to the agent's main thread so the design accounts for the
// architectural principles from the start (the highest-leverage injection —
// architecture is decided here). It's the *forward* counterpart to change-code's
// plan-quality judge (the *backward* check), both consuming the one registry.
//
// Re-run it each time a new design begins: agents don't reread a static doc, so
// re-delivering keeps the principles live in attention.
package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// NewStartPlanCmd returns the cobra command for `sdlc start-plan`.
func NewStartPlanCmd() *cobra.Command {
	var issue int
	cmd := &cobra.Command{
		Use:           "start-plan",
		Short:         "Enter planning: deliver the architecture principles to design against (#75)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"start-plan\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			runStartPlan(cmd.OutOrStdout(), issue)
			return nil
		},
	}
	cmd.Flags().IntVar(&issue, "issue", 0, "issue being planned (optional, for the label)")
	return cmd
}

// runStartPlan emits the planning framing + the at-plan architecture lens.
func runStartPlan(stdout io.Writer, issue int) {
	label := "this issue"
	if issue > 0 {
		label = fmt.Sprintf("#%d", issue)
	}
	cinfo(stdout, fmt.Sprintf("Entering planning for %s. Design with these architectural", label))
	fmt.Fprintln(stdout, "    principles in mind — the plan-quality gate (`sdlc change-code`) checks the")
	fmt.Fprintln(stdout, "    plan against them, and the boundary review checks the code. Cite ARCH-* in")
	fmt.Fprintln(stdout, "    your plan where a principle shaped a decision.")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, judge.ArchitectureBlock("at-plan"))
}
