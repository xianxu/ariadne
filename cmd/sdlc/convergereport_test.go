package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// #234: the report reads the ledger at $WF_CALIB_LEDGER and prints k per
// group; with no split rows it says why the report is empty.
func TestConvergeReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ledger.tsv")
	t.Setenv("WF_CALIB_LEDGER", path)
	run := func() string {
		cmd := NewConvergeReportCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetArgs([]string{})
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		return out.String()
	}
	legacy := estimate.FormatRow(estimate.LedgerRow{Issue: "a#1", Estimate: 1, Actual: 2, WindowTrusted: true})
	if err := os.WriteFile(path, []byte(estimate.Header()+"\n"+legacy+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := run(); !strings.Contains(got, "no rows with a build/converge split yet") {
		t.Fatalf("legacy-only ledger: %q", got)
	}
	split := estimate.FormatRow(estimate.LedgerRow{Issue: "a#2", Estimate: 1, Actual: 1.5, WindowTrusted: true, Model: "estimate-logic-v3.1",
		FlowKind: "full", BuildHours: 1, ConvergeHours: 0.5, BoundaryRounds: 2, SplitSource: "dispatch", ReviewRecipe: "milestone-review"})
	if err := os.WriteFile(path, []byte(estimate.Header()+"\n"+legacy+"\n"+split+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := run()
	if !strings.Contains(got, "milestone-review") || !strings.Contains(got, "0.50") || strings.Contains(got, "no rows") {
		t.Fatalf("report: %q", got)
	}
}

// Calibration correction semantics must survive the command IO boundary.
func TestConvergeReportCorrectionResolutionAndErrors(t *testing.T) {
	for _, mode := range []string{"exclude", "malformed", "missing-target"} {
		t.Run(mode, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "ledger.tsv")
			t.Setenv("WF_CALIB_LEDGER", path)
			text := calibrationConsumerCorrectionFixture(t, mode)
			if err := os.WriteFile(path, []byte(text), 0644); err != nil {
				t.Fatal(err)
			}
			cmd := NewConvergeReportCmd()
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs([]string{})
			err := cmd.Execute()
			if mode == "exclude" {
				if err != nil || !strings.Contains(out.String(), "no rows") {
					t.Fatalf("excluded target returned: %v %q", err, out.String())
				}
			} else if err == nil || !strings.Contains(err.Error(), "calibration ledger") ||
				strings.Contains(out.String(), "converge h") || strings.Contains(out.String(), "no rows with a build/converge split") {
				t.Fatalf("corrupt ledger not reported: %v %q", err, out.String())
			}
		})
	}
}

func calibrationConsumerCorrectionFixture(t *testing.T, mode string) string {
	t.Helper()
	raw := estimate.FormatRow(estimate.LedgerRow{Issue: "repo#1", Estimate: 2, Actual: 1, WindowTrusted: true, Model: "estimate-logic-v2", BuildHours: 1, ConvergeHours: 0.25, SplitSource: "dispatch"})
	rows, err := estimate.ParseRowsChecked(raw)
	if err != nil {
		t.Fatal(err)
	}
	correction := estimate.LedgerRow{Issue: "repo#1", Correction: &estimate.CorrectionRecord{Version: 1, Operation: "consumer-op", Generation: "g", Disposition: "exclude", Targets: []string{rows[0].RawID}, InputDigest: strings.Repeat("a", 64)}}
	if mode == "missing-target" {
		correction.Correction.Targets = []string{strings.Repeat("b", 64)}
	}
	repair := estimate.FormatRow(correction)
	if mode == "malformed" {
		repair = estimate.FormatRow(estimate.LedgerRow{Issue: "repo#1"}) + "\t{}"
	}
	return estimate.Header() + "\n" + raw + "\n" + repair + "\n"
}
