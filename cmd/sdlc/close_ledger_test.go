package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

func ledgerTestBody() string {
	return "# T\n\n## Estimate\n\n```estimate\n" +
		"model: estimate-logic-v2\n" +
		"item: greenfield-go-module design=0.3 impl=0.6\n" +
		"item: smaller-go-module design=0.2 impl=0.6\n" +
		"item: smaller-go-module design=0.2 impl=0.5\n" +
		"item: atlas-docs design=0.0 impl=0.2\n" +
		"item: milestone-review design=0.0 impl=0.6\n" +
		"design-buffer: 0.30\ntotal: 3.4\n```\n"
}

// TestAppendCalibrationRow_Happy pins #117 mechanism 3: a full-issue close appends
// one estimate↔actual row (header on first write) with the right values, the
// design/impl subtotals parsed from the ## Estimate block, and window-trusted=no
// (no `started:` until #116).
func TestAppendCalibrationRow_Happy(t *testing.T) {
	path := t.TempDir() + "/ledger.tsv"
	t.Setenv("WF_CALIB_LEDGER", path)
	var errb bytes.Buffer
	f := &closeFlags{Actual: "1.7", Mode: "supervised"}
	fm := "id: 1\nstatus: working\nestimate_hours: 3.4\n"

	appendCalibrationRow(&errb, f, fm, ledgerTestBody(), "ariadne", "117", "2026-06-17")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ledger not written: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want header + 1 row, got %d lines: %q", len(lines), string(data))
	}
	row := lines[1]
	for _, want := range []string{"ariadne#117", "3.40", "1.70", "0.70", "2.50", "estimate-logic-v2", "supervised", "2026-06-17"} {
		if !strings.Contains(row, want) {
			t.Errorf("row missing %q: %q", want, row)
		}
	}
	if !strings.Contains(row, "\tno\t") {
		t.Errorf("expected window-trusted=no (no started:), got %q", row)
	}
}

func TestAppendCalibrationRow_TrustedWithStarted(t *testing.T) {
	path := t.TempDir() + "/ledger.tsv"
	t.Setenv("WF_CALIB_LEDGER", path)
	var errb bytes.Buffer
	f := &closeFlags{Actual: "2.0"}
	fm := "estimate_hours: 3.4\nstarted: 2026-06-17T10:00:00Z\n"

	appendCalibrationRow(&errb, f, fm, ledgerTestBody(), "ariadne", "117", "2026-06-17")

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), "\tyes\t") {
		t.Errorf("started: present should mark window-trusted=yes: %q", string(data))
	}
}

// TestAppendCalibrationRow_BrainAbsentSkips pins the M2 plan-quality finding #1:
// with no override and no brain ledger dir (base-layer downstream), close must
// skip-with-warning, never break.
func TestAppendCalibrationRow_BrainAbsentSkips(t *testing.T) {
	t.Setenv("WF_CALIB_LEDGER", "") // force the brain-path branch
	var errb bytes.Buffer
	f := &closeFlags{Actual: "1.0", BrainDir: "/no/such/brain/dir"}

	appendCalibrationRow(&errb, f, "estimate_hours: 1\n", "# T\n", "ariadne", "117", "2026-06-17")

	if !strings.Contains(errb.String(), "skipped") {
		t.Errorf("expected a skip warning, got %q", errb.String())
	}
}

// TestShouldLogCalibration pins the ledger's core integrity contract (#117 M3
// review): only a full-issue close with a measured actual logs a row; a milestone
// close (partial actual) must not.
func TestShouldLogCalibration(t *testing.T) {
	cases := []struct {
		milestone, actual string
		want              bool
	}{
		{"", "0.9", true},   // full-issue close with actual → log
		{"M1", "0.9", false}, // milestone close → never log (partial actual)
		{"", "", false},      // full-issue close, no actual → nothing to log
		{"M2", "", false},
	}
	for _, c := range cases {
		got := shouldLogCalibration(&closeFlags{Milestone: c.milestone, Actual: c.actual})
		if got != c.want {
			t.Errorf("shouldLogCalibration{milestone:%q actual:%q} = %v, want %v", c.milestone, c.actual, got, c.want)
		}
	}
}

// TestRunClose_IssueClose_WritesLedgerRow is the integration proof (#117 M3
// review): a full-issue close routed through runClose appends exactly one
// calibration row; a milestone close on a fresh repo appends none.
func TestRunClose_LedgerGuardWiring(t *testing.T) {
	// Milestone close → no ledger row.
	t.Run("milestone writes nothing", func(t *testing.T) {
		issuesDir := closeRepo(t, 71)
		ledger := filepath.Join(t.TempDir(), "l.tsv")
		t.Setenv("WF_CALIB_LEDGER", ledger)
		f := &closeFlags{Issue: 71, Milestone: "M1", Actual: "1", Verified: "slice", NoAtlas: true, IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}
		if err := runClose(io.Discard, f); err != nil {
			t.Fatalf("runClose (milestone): %v", err)
		}
		if _, err := os.Stat(ledger); !os.IsNotExist(err) {
			t.Errorf("milestone close must NOT write the calibration ledger (stat err: %v)", err)
		}
	})
	// Full-issue close → exactly one ledger row.
	t.Run("issue close writes one row", func(t *testing.T) {
		issuesDir := closeRepo(t, 72)
		ledger := filepath.Join(t.TempDir(), "l.tsv")
		t.Setenv("WF_CALIB_LEDGER", ledger)
		f := &closeFlags{Issue: 72, Actual: "1", Verified: "done", NoAtlas: true, IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}
		if err := runClose(io.Discard, f); err != nil {
			t.Fatalf("runClose (issue): %v", err)
		}
		data, err := os.ReadFile(ledger)
		if err != nil {
			t.Fatalf("ledger not written on issue close: %v", err)
		}
		if rows := estimate.ParseRows(string(data)); len(rows) != 1 {
			t.Fatalf("want exactly 1 data row, got %d: %q", len(rows), string(data))
		}
	})
}

func TestAppendCalibrationRow_DriftWarns(t *testing.T) {
	path := t.TempDir() + "/ledger.tsv"
	t.Setenv("WF_CALIB_LEDGER", path)

	// Pre-populate 4 trusted, >2× over-estimate rows.
	var sb strings.Builder
	sb.WriteString(estimate.Header() + "\n")
	for i := 0; i < 4; i++ {
		sb.WriteString(estimate.FormatRow(estimate.LedgerRow{
			Issue: "ariadne#" + strconv.Itoa(i), Estimate: 5, Actual: 0.5,
			Model: "estimate-logic-v2", WindowTrusted: true, Date: "2026-06-17",
		}) + "\n")
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	// Append a 5th trusted, >2× over row → last 5 trusted all over → drift.
	var errb bytes.Buffer
	f := &closeFlags{Actual: "0.3"}
	fm := "estimate_hours: 3.4\nstarted: 2026-06-17T10:00:00Z\n"
	appendCalibrationRow(&errb, f, fm, ledgerTestBody(), "ariadne", "117", "2026-06-17")

	if !strings.Contains(errb.String(), "drift") {
		t.Errorf("expected a drift warning, got %q", errb.String())
	}
}
