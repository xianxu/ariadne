// archprinciples.go — `sdlc arch-principles`: print the ARCH-* architecture
// principles from the one registry (#128). The registry (architecture.md) is the
// single source; it's PUSHED at the gates (start-plan to the main thread, the
// plan-quality + boundary-review judges inline). This command is the standalone
// PULL: for non-gate work (§7 autonomous fixes, quick edits, Q&A) that never runs
// start-plan, and as a tested CLI consumer that DERIVES from the registry rather
// than a hand-maintained restatement (ARCH-PURPOSE, #126). It shares the one
// render primitive judge.ArchitectureBlock with start-plan (ARCH-DRY) — this is
// just a second entry point to it.
package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// NewArchPrinciplesCmd returns the cobra command for `sdlc arch-principles`.
func NewArchPrinciplesCmd() *cobra.Command {
	var lens string
	cmd := &cobra.Command{
		Use:           "arch-principles",
		Short:         "Print the ARCH-* architecture principles (single source for non-gate work)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"arch-principles\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runArchPrinciples(cmd.OutOrStdout(), lens)
		},
	}
	cmd.Flags().StringVar(&lens, "lens", "at-plan",
		"which lens to foreground: at-plan (design time) | at-review (on a diff)")
	return cmd
}

// runArchPrinciples prints the registry under the requested lens. Rejects an
// unknown lens so a typo can't silently render the wrong framing. The render
// itself is the shared pure primitive judge.ArchitectureBlock (ARCH-DRY); this is
// the thin IO seam (ARCH-PURE).
func runArchPrinciples(stdout io.Writer, lens string) error {
	if lens != "at-plan" && lens != "at-review" {
		return fmt.Errorf("unknown --lens %q: want at-plan or at-review", lens)
	}
	fmt.Fprintln(stdout, judge.ArchitectureBlock(lens))
	return nil
}
