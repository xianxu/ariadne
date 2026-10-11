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
