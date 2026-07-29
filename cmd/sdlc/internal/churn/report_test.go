package churn

import (
	"math"
	"testing"
)

func TestSummarize(t *testing.T) {
	final := []FileStat{
		{Path: "cmd/sdlc/changecode.go", Insertions: 100},
		{Path: "cmd/sdlc/changecode_test.go", Insertions: 200},
		{Path: "atlas/index.md", Insertions: 5},
		{Path: "workshop/issues/000187-x.md", Insertions: 60},
	}
	// The window's commits inserted 1095 lines total to land those 365.
	r := Summarize(final, 1095)
	if r.Final.CodeProd != 100 || r.Final.CodeTest != 200 || r.Final.Atlas != 5 || r.Final.Workshop != 60 {
		t.Errorf("buckets = %+v", r.Final)
	}
	if r.FinalTotal != 365 {
		t.Errorf("FinalTotal = %d, want 365", r.FinalTotal)
	}
	if math.Abs(r.Rework-3.0) > 0.01 {
		t.Errorf("Rework = %.2f, want 3.00", r.Rework)
	}
}

// Rework is undefined, not +Inf, when a window lands no insertions (a pure-deletion or
// empty window) — a NaN/Inf in the TSV would poison every downstream reader.
func TestSummarizeZeroFinal(t *testing.T) {
	if r := Summarize(nil, 40); r.Rework != 0 {
		t.Errorf("Rework = %v, want 0 for an empty final diff", r.Rework)
	}
}

// Both halves of the ratio are zero on an empty window: 0/0 must not surface as NaN
// either. Checked with IsNaN rather than != 0 because NaN != 0 is true, so the test
// above would pass on a NaN and say nothing.
func TestSummarizeEmptyWindowIsNotNaN(t *testing.T) {
	r := Summarize(nil, 0)
	if math.IsNaN(r.Rework) || math.IsInf(r.Rework, 0) {
		t.Errorf("Rework = %v, want a finite 0", r.Rework)
	}
}

// Binary files show "-" for both counts in numstat; they must be skipped rather than
// abort the sum or land as a zero-insertion row that ClassifyPath still buckets.
func TestParseNumstatSkipsBinary(t *testing.T) {
	out := "12\t3\tcmd/sdlc/close.go\n" +
		"-\t-\tdocs/diagram.png\n" +
		"7\t0\tatlas/index.md\n"
	stats := ParseNumstat(out)
	if len(stats) != 2 {
		t.Fatalf("ParseNumstat returned %d stats, want 2: %+v", len(stats), stats)
	}
	if stats[0].Path != "cmd/sdlc/close.go" || stats[0].Insertions != 12 {
		t.Errorf("stats[0] = %+v", stats[0])
	}
	if stats[1].Path != "atlas/index.md" || stats[1].Insertions != 7 {
		t.Errorf("stats[1] = %+v", stats[1])
	}
}

// `git log --numstat --format=` separates each commit's stats with a blank line, and a
// rename shows as "ins\tdel\told => new" (or the NUL-ish three-field form). Neither may
// derail the parse — the commit-side total is a SUM, so one dropped line understates
// rework and one bogus line inflates it.
func TestParseNumstatHandlesLogOutput(t *testing.T) {
	out := "\n30\t0\tcmd/x.go\n\n30\t30\tcmd/x.go\n\n" +
		"4\t1\tcmd/{old.go => new.go}\n"
	stats := ParseNumstat(out)
	if len(stats) != 3 {
		t.Fatalf("ParseNumstat returned %d stats, want 3: %+v", len(stats), stats)
	}
	if got := TotalInsertions(stats); got != 64 {
		t.Errorf("TotalInsertions = %d, want 64", got)
	}
	// The rename's path keeps git's arrow form; it classifies by the leading segment,
	// which is all the bucketing needs.
	if stats[2].Insertions != 4 {
		t.Errorf("rename row = %+v", stats[2])
	}
}

// A file touched by three commits appears three times on the log side. TotalInsertions
// must count every appearance — collapsing by path would erase rework entirely, which is
// the one thing this metric exists to see.
func TestTotalInsertionsCountsRepeatedPaths(t *testing.T) {
	stats := []FileStat{
		{Path: "cmd/x.go", Insertions: 30},
		{Path: "cmd/x.go", Insertions: 30},
		{Path: "cmd/x.go", Insertions: 30},
	}
	if got := TotalInsertions(stats); got != 90 {
		t.Errorf("TotalInsertions = %d, want 90", got)
	}
}
