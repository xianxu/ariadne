// convergereport.go — `sdlc converge-report`: the converge factor k per model,
// flow and review recipe (#234). Read-only.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// NewConvergeReportCmd returns the cobra command for `sdlc converge-report`.
func NewConvergeReportCmd() *cobra.Command {
	var brainDir string
	cmd := &cobra.Command{
		Use:           "converge-report",
		Short:         "Converge factor k = converge/build per model, flow and review recipe (#234)",
		Long:          "Placeholder — replaced by helptext.MustGet(\"converge-report\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			brainAbs, _ := filepath.Abs(brainDir)
			path := calibrationLedgerPath(brainAbs)
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("calibration ledger %s: %w", path, err)
			}
			rows, err := estimate.ParseRowsChecked(string(data))
			if err == nil {
				rows, err = estimate.ResolveCorrections(rows)
			}
			if err != nil {
				return fmt.Errorf("calibration ledger %s: %w", path, err)
			}
			renderConvergeReport(cmd.OutOrStdout(), estimate.ConvergeFactors(rows))
			return nil
		},
	}
	cmd.Flags().StringVar(&brainDir, "brain-dir", "../brain", "brain root holding the calibration ledger")
	return cmd
}

// renderConvergeReport prints one line per group. Pure over its input.
func renderConvergeReport(w io.Writer, groups []estimate.ConvergeFactor) {
	if len(groups) == 0 {
		fmt.Fprintln(w, "no rows with a build/converge split yet (ariadne#234; historical backfill is ariadne#323)")
		return
	}
	// #603: n counts weighted rows; partial says how many of them a partial
	// attributed share weights, so the hours are weighted hours.
	fmt.Fprintf(w, "%-6s %-22s %-36s %4s %7s %8s %10s %6s\n", "flow", "model", "recipe", "n", "partial", "build h", "converge h", "k")
	for _, g := range groups {
		fmt.Fprintf(w, "%-6s %-22s %-36s %4d %7d %8.2f %10.2f %6.2f\n", estimate.Dash(g.Flow), estimate.Dash(g.Model), estimate.Dash(g.Recipe), g.N, g.Partial, g.Build, g.Converge, g.K())
	}
}
