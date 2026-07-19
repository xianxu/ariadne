package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// seedBrainLedger writes a calibration-ledger fixture under a temp brain dir's
// velocity path and returns (brainDir, baselinePath).
func seedBrainLedger(t *testing.T) (string, string) {
	t.Helper()
	brain := t.TempDir()
	velDir := filepath.Dir(estimate.VelocityPath(brain, "calibration-ledger.tsv"))
	if err := os.MkdirAll(velDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ledger := estimate.Header() + "\n" +
		"a#1\t1.00\t0.00\t0.00\t4.00\t0.25\tm\t-\tyes\t2026-06-02\n" +
		"a#2\t1.00\t0.00\t0.00\t6.00\t0.17\tm\t-\tno\t2026-06-09\n" +
		"a#3\t1.00\t0.00\t0.00\t10.00\t0.10\tm\t-\tyes\t2026-06-23\n"
	if err := os.WriteFile(estimate.VelocityPath(brain, "calibration-ledger.tsv"), []byte(ledger), 0o644); err != nil {
		t.Fatal(err)
	}
	return brain, estimate.VelocityPath(brain, "throughput-baseline.tsv")
}

func TestProjectThroughput_Registered(t *testing.T) {
	root := buildRoot()
	project, _, _ := root.Find([]string{"project"})
	if found, _, err := project.Find([]string{"throughput"}); err != nil || found == project {
		t.Fatalf("project throughput not registered: %v", err)
	}
}

func TestProjectThroughput_Bless(t *testing.T) {
	brain, baselinePath := seedBrainLedger(t)
	// span 2026-06-01..2026-06-28 = 28 days = 4wk; in-span 4+6+10 = 20h → 5.0/wk.
	var out strings.Builder
	f := &projectThroughputFlags{Bless: "2026-06-01..2026-06-28", Ceiling: 2, BrainDir: brain}
	if err := runProjectThroughput(&out, &out, f); err != nil {
		t.Fatalf("bless: %v\n%s", err, out.String())
	}
	data, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("baseline not written: %v", err)
	}
	rows, err := estimate.ParseBaselineTSV(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].HoursPerWeek < 4.99 || rows[0].HoursPerWeek > 5.01 {
		t.Fatalf("baseline row wrong: %+v", rows)
	}
	if rows[0].Ceiling != 2 {
		t.Errorf("ceiling = %d, want 2", rows[0].Ceiling)
	}
	if !strings.Contains(out.String(), "5.00") {
		t.Errorf("output should report the measured rate: %q", out.String())
	}
	// The untrusted row (a#2) should be surfaced as a warning.
	if !strings.Contains(out.String(), "window_trusted=no") {
		t.Errorf("output should warn about the 1 window_trusted=no row: %q", out.String())
	}
}

func TestProjectThroughput_BlessAppends(t *testing.T) {
	brain, baselinePath := seedBrainLedger(t)
	for _, span := range []string{"2026-06-01..2026-06-28", "2026-06-08..2026-06-28"} {
		f := &projectThroughputFlags{Bless: span, Ceiling: 3, BrainDir: brain}
		var out strings.Builder
		if err := runProjectThroughput(&out, &out, f); err != nil {
			t.Fatalf("bless %s: %v\n%s", span, err, out.String())
		}
	}
	data, _ := os.ReadFile(baselinePath)
	rows, err := estimate.ParseBaselineTSV(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 appended rows, got %d", len(rows))
	}
	if rows[1].SpanStart != "2026-06-08" {
		t.Errorf("last (current) baseline should be the second bless: %+v", rows[1])
	}
}

func TestProjectThroughput_EmptySpanRefuses(t *testing.T) {
	brain, _ := seedBrainLedger(t)
	f := &projectThroughputFlags{Bless: "2020-01-01..2020-01-31", BrainDir: brain}
	var out strings.Builder
	if err := runProjectThroughput(&out, &out, f); err == nil {
		t.Fatal("bless over an empty span must refuse")
	}
}

func TestProjectThroughput_BareShowsCurrentAndTrailing(t *testing.T) {
	brain, _ := seedBrainLedger(t)
	// bless first so there's a current baseline.
	blessF := &projectThroughputFlags{Bless: "2026-06-01..2026-06-28", Ceiling: 2, BrainDir: brain}
	var blessOut strings.Builder
	if err := runProjectThroughput(&blessOut, &blessOut, blessF); err != nil {
		t.Fatal(err)
	}
	// bare form prints the current baseline + a trailing comparison.
	var out strings.Builder
	if err := runProjectThroughput(&out, &out, &projectThroughputFlags{BrainDir: brain}); err != nil {
		t.Fatalf("bare: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "5.00") {
		t.Errorf("bare form should show the current baseline rate: %q", out.String())
	}
}

func TestProjectThroughput_NoBaselineBareHints(t *testing.T) {
	brain, _ := seedBrainLedger(t) // ledger exists, no baseline blessed yet
	var out strings.Builder
	if err := runProjectThroughput(&out, &out, &projectThroughputFlags{BrainDir: brain}); err != nil {
		t.Fatalf("bare with no baseline should not error: %v", err)
	}
	if !strings.Contains(out.String(), "--bless") {
		t.Errorf("no-baseline bare form should hint the bless command: %q", out.String())
	}
}
