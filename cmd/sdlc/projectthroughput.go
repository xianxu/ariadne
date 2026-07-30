// projectthroughput.go — `sdlc project throughput` (#182 M1). Blesses a
// representative span into a measured throughput baseline (operator picks the
// span, the machinery measures the rate from the calibration ledger), and — in
// its bare form — prints the current baseline plus a trailing-4-week comparison
// so a stale baseline surfaces on demand. The baseline never auto-substitutes
// the trailing number; it's read by the #182 calendar forecast.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

type projectThroughputFlags struct {
	Bless    string // "FROM..TO" ISO span to bless; empty → bare (show) mode
	Ceiling  int    // attention ceiling recorded with a bless (default 2, #117)
	BrainDir string
}

func newProjectThroughputCmd() *cobra.Command {
	f := projectThroughputFlags{}
	cmd := &cobra.Command{
		Use:           "throughput",
		Short:         "Bless a measured throughput baseline, or show the current one",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// A bless writes the baseline file; a bare show is read-only. Only
			// the mutating form takes the repo lock, so the bare form stays
			// cheap. cobra can't vary markMutatingCommand per-invocation, so
			// runProjectThroughput does its own writes without the checkpoint lock
			// (the baseline file is in brain, outside the repo transaction).
			return runProjectThroughput(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	cmd.Flags().StringVar(&f.Bless, "bless", "", "bless a representative span FROM..TO (ISO dates) as the throughput baseline")
	cmd.Flags().IntVar(&f.Ceiling, "ceiling", 2, "attention ceiling to record with a bless (concurrent projects, #117)")
	cmd.Flags().StringVar(&f.BrainDir, "brain-dir", "../brain", "brain root holding the velocity ledgers")
	return cmd
}

// calibrationLedgerPath resolves the calibration ledger, honoring WF_CALIB_LEDGER
// first (mirroring appendCalibrationRow), else the brain velocity path.
func calibrationLedgerPath(brainDir string) string {
	if p := os.Getenv("WF_CALIB_LEDGER"); p != "" {
		return p
	}
	return estimate.VelocityPath(brainDir, "calibration-ledger.tsv")
}

// throughputBaselinePath resolves the baseline file, honoring
// WF_THROUGHPUT_BASELINE first (test override), else the brain velocity path.
func throughputBaselinePath(brainDir string) string {
	if p := os.Getenv("WF_THROUGHPUT_BASELINE"); p != "" {
		return p
	}
	return estimate.VelocityPath(brainDir, "throughput-baseline.tsv")
}

func runProjectThroughput(stdout, stderr io.Writer, f *projectThroughputFlags) error {
	if f.Bless != "" {
		return blessThroughput(stdout, stderr, f)
	}
	return showThroughput(stdout, stderr, f)
}

func blessThroughput(stdout, stderr io.Writer, f *projectThroughputFlags) error {
	from, to, ok := strings.Cut(f.Bless, "..")
	from, to = strings.TrimSpace(from), strings.TrimSpace(to)
	if !ok || from == "" || to == "" {
		return fmt.Errorf("--bless wants a span FROM..TO (ISO dates), got %q", f.Bless)
	}
	ledgerText, err := os.ReadFile(calibrationLedgerPath(f.BrainDir))
	if err != nil {
		return fmt.Errorf("read calibration ledger: %w", err)
	}
	measure, err := estimate.SpanThroughput(estimate.ParseRows(string(ledgerText)), from, to)
	if err != nil {
		return err
	}
	row := estimate.ThroughputBaseline{
		BlessedDate:  projectTodayFn(),
		SpanStart:    from,
		SpanEnd:      to,
		HoursPerWeek: measure.HoursPerWeek,
		Rows:         measure.Issues, // distinct issues since #192; see ThroughputBaseline.Rows
		Ceiling:      f.Ceiling,
	}
	if err := appendBaselineRow(throughputBaselinePath(f.BrainDir), row); err != nil {
		return err
	}
	// Print BOTH counts: the gap between issues and rows scanned is the re-close duplication,
	// and the old single "%d rows" is exactly what made a 1.41x inflation look like data (#192).
	cok(stdout, fmt.Sprintf("blessed baseline: %.2f h/wk over %s..%s (%d issues from %d ledger rows, %d days, ceiling %d)",
		row.HoursPerWeek, from, to, measure.Issues, measure.RowsScanned, measure.Days, f.Ceiling))
	if measure.UntrustedIssues > 0 {
		cwarn(stderr, fmt.Sprintf("%d of %d issues have window_trusted=no newest measurements (truncated actuals) — the rate may run low", measure.UntrustedIssues, measure.Issues))
	}
	if measure.Skipped > 0 {
		cwarn(stderr, fmt.Sprintf("%d ledger row(s) had an unparsable date and were excluded", measure.Skipped))
	}
	return nil
}

// appendBaselineRow appends row to the baseline file, creating it with a header
// comment + column header if absent. Append-only = re-blessing history.
func appendBaselineRow(path string, row estimate.ThroughputBaseline) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	var b strings.Builder
	if len(existing) == 0 {
		b.WriteString("# Throughput baseline (ariadne#182) — measured focused h/wk over an\n")
		b.WriteString("# operator-blessed representative span. Append-only; last row = current.\n")
		b.WriteString(estimate.BaselineHeader() + "\n")
	} else {
		b.Write(existing)
		if !strings.HasSuffix(string(existing), "\n") {
			b.WriteString("\n")
		}
	}
	b.WriteString(estimate.RenderBaselineRow(row) + "\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func showThroughput(stdout, stderr io.Writer, f *projectThroughputFlags) error {
	baselineText, err := os.ReadFile(throughputBaselinePath(f.BrainDir))
	if err != nil {
		if !os.IsNotExist(err) {
			// A real IO error (permission, etc.) is worth surfacing honestly
			// rather than reporting it as "not blessed yet".
			return fmt.Errorf("read throughput baseline: %w", err)
		}
		cinfo(stdout, "no blessed throughput baseline yet")
		cinfo(stdout, "  bless one: sdlc project throughput --bless <FROM>..<TO>  (a representative ~4-week span)")
		return nil
	}
	rows, err := estimate.ParseBaselineTSV(string(baselineText))
	if err != nil {
		return fmt.Errorf("parse throughput baseline: %w", err)
	}
	if len(rows) == 0 {
		cinfo(stdout, "throughput baseline file has no rows — bless one with --bless <FROM>..<TO>")
		return nil
	}
	cur := rows[len(rows)-1]
	cok(stdout, fmt.Sprintf("current baseline: %.2f h/wk over %s..%s (blessed %s, ceiling %d)",
		cur.HoursPerWeek, cur.SpanStart, cur.SpanEnd, cur.BlessedDate, cur.Ceiling))

	// Trailing-4-week comparison (informational only; never substituted).
	if ledgerText, lerr := os.ReadFile(calibrationLedgerPath(f.BrainDir)); lerr == nil {
		today := projectTodayFn()
		if t, perr := time.Parse(isoLayout, today); perr == nil {
			from := t.AddDate(0, 0, -28).Format(isoLayout)
			if m, merr := estimate.SpanThroughput(estimate.ParseRows(string(ledgerText)), from, today); merr == nil {
				delta := m.HoursPerWeek - cur.HoursPerWeek
				cinfo(stdout, fmt.Sprintf("trailing 4wk: %.2f h/wk (%+.2f vs baseline) — re-bless if this is the new normal", m.HoursPerWeek, delta))
			}
		}
	}
	return nil
}

const isoLayout = "2006-01-02"
