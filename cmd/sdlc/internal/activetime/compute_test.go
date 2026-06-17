package activetime

import (
	"path/filepath"
	"testing"
)

const wideSince = "2026-01-01T00:00:00Z"
const wideUntil = "2026-01-02T00:00:00Z"

// eventsDir writes a one-file transcript dir and returns its path.
func eventsDir(t *testing.T, lines ...string) string {
	t.Helper()
	dir := t.TempDir()
	writeJSONL(t, filepath.Join(dir, "s.jsonl"), lines)
	return dir
}

func TestComputeTelemetryGap(t *testing.T) {
	// Commits in window, but zero transcript events → TelemetryGap (#68).
	withGitRun(t, func(repo string, args ...string) ([]byte, error) {
		return []byte("aaaaaaa\t2026-01-01T00:50:00Z\t#8 work"), nil
	})
	res, err := Compute(Options{
		Dirs: []string{t.TempDir()}, GitRepo: "/repo",
		SinceISO: wideSince, UntilISO: wideUntil,
		Issues: []string{"8"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != TelemetryGap {
		t.Fatalf("want TelemetryGap, got %v", res.Status)
	}
}

func TestComputeEmptyWindow(t *testing.T) {
	// No commits, no events → EmptyWindow.
	withGitRun(t, func(repo string, args ...string) ([]byte, error) { return []byte(""), nil })
	res, err := Compute(Options{
		Dirs: []string{t.TempDir()}, GitRepo: "/repo",
		SinceISO: wideSince, UntilISO: wideUntil,
		Issues: []string{"8"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != EmptyWindow {
		t.Fatalf("want EmptyWindow, got %v", res.Status)
	}
}

func TestComputeMeasuredSegments(t *testing.T) {
	// Events (no mentions) + an interior commit → commit-weighted attribution.
	// Segment [00:00,00:45) holds 00:00 + 00:30 → active = cap(30min,15) = 15,
	// all to #8 at commit-weight 1.0. The 01:00 event is a single-event suffix
	// (active 0).
	dir := eventsDir(t,
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"a"}}`,
		`{"timestamp":"2026-01-01T00:30:00Z","type":"user","message":{"content":"b"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"user","message":{"content":"c"}}`,
	)
	withGitRun(t, func(repo string, args ...string) ([]byte, error) {
		return []byte("aaaaaaa\t2026-01-01T00:45:00Z\t#8 work"), nil
	})
	res, err := Compute(Options{
		Dirs: []string{dir}, GitRepo: "/repo",
		SinceISO: wideSince, UntilISO: wideUntil,
		Issues: []string{"8"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != Measured {
		t.Fatalf("want Measured, got %v", res.Status)
	}
	if !approx(res.PerIssue["8"], 15) {
		t.Fatalf("want #8=15min, got %v (all: %v)", res.PerIssue["8"], res.PerIssue)
	}
	if !approx(res.TotalActive, 15) {
		t.Fatalf("want total active 15, got %v", res.TotalActive)
	}
	if res.NumEvents != 3 || res.NumCommits != 1 {
		t.Fatalf("counts wrong: events=%d commits=%d", res.NumEvents, res.NumCommits)
	}
}

func TestComputeNoCommitsMentionFallback(t *testing.T) {
	// Events present, no commits → whole-window mention attribution.
	// 00:00(#8) + 00:10(#10) → active = 10min split by mentions 1:1 → 5 each.
	dir := eventsDir(t,
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"on #8"}}`,
		`{"timestamp":"2026-01-01T00:10:00Z","type":"user","message":{"content":"on #10"}}`,
	)
	withGitRun(t, func(repo string, args ...string) ([]byte, error) { return []byte(""), nil })
	res, err := Compute(Options{
		Dirs: []string{dir}, GitRepo: "/repo",
		SinceISO: wideSince, UntilISO: wideUntil,
		Issues: []string{"8", "10"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != Measured {
		t.Fatalf("want Measured, got %v", res.Status)
	}
	if !approx(res.PerIssue["8"], 5) || !approx(res.PerIssue["10"], 5) {
		t.Fatalf("want #8=5,#10=5, got %v", res.PerIssue)
	}
}

// PrefixWeight is a *float64: an explicit 0 must be honored, not treated as
// unset. With a real prefix segment, prefixWeight 0 means the commit gets no
// commit-weighted share — all of the prefix's active time goes by mention.
func TestComputePrefixWeightZeroHonored(t *testing.T) {
	// Prefix events 00:00(#8) + 00:10(#10) before commit #8 at 00:20.
	dir := eventsDir(t,
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8"}}`,
		`{"timestamp":"2026-01-01T00:10:00Z","type":"user","message":{"content":"#10"}}`,
		`{"timestamp":"2026-01-01T00:25:00Z","type":"user","message":{"content":"after"}}`,
	)
	withGitRun(t, func(repo string, args ...string) ([]byte, error) {
		return []byte("aaaaaaa\t2026-01-01T00:20:00Z\t#8 work"), nil
	})
	zero := 0.0
	res, err := Compute(Options{
		Dirs: []string{dir}, GitRepo: "/repo",
		SinceISO: wideSince, UntilISO: wideUntil,
		Issues: []string{"8", "10"}, CommitWeight: 1.0, PrefixWeight: &zero,
		ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Prefix segment [00:00,00:20): events 00:00,00:10 → active 10min, anchor #8.
	// With prefixWeight 0, the commit share is 0; full 10min splits by mentions
	// #8:#10 = 1:1 → 5 each. (A buggy !=0 sentinel would route through
	// commit-weight 1.0 and give #8 all 10.)
	if !approx(res.PerIssue["8"], 5) || !approx(res.PerIssue["10"], 5) {
		t.Fatalf("prefixWeight 0 not honored: want #8=5,#10=5, got %v", res.PerIssue)
	}
}
