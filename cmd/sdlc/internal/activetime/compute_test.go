package activetime

import (
	"path/filepath"
	"strings"
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
	// Events (no mentions) + an interior commit → commit-boundary attribution.
	// The source-local event gaps are each capped at 15 minutes and both are
	// nearest to the #8 commit, so both count toward #8.
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
	if !approx(res.PerIssue["8"], 30) {
		t.Fatalf("want #8=30min, got %v (all: %v)", res.PerIssue["8"], res.PerIssue)
	}
	if !approx(res.TotalActive, 30) {
		t.Fatalf("want total active 30, got %v", res.TotalActive)
	}
	if res.NumEvents != 3 || res.NumCommits != 1 {
		t.Fatalf("counts wrong: events=%d commits=%d", res.NumEvents, res.NumCommits)
	}
}

func TestComputeNoCommitsMentionFallback(t *testing.T) {
	// Events present, no commits → source-run mention attribution.
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
	if len(res.Segments) != 1 {
		t.Fatalf("no-commit fallback should still render attribution segments, got %+v", res.Segments)
	}
}

func TestComputeNoCommitsFallbackPreservesOverlappingSources(t *testing.T) {
	dir := t.TempDir()
	writeJSONL(t, filepath.Join(dir, "issue8.jsonl"),
		[]string{
			`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8 start"}}`,
			`{"timestamp":"2026-01-01T00:15:00Z","type":"user","message":{"content":"#8 done"}}`,
		})
	writeJSONL(t, filepath.Join(dir, "issue9.jsonl"),
		[]string{
			`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#9 start"}}`,
			`{"timestamp":"2026-01-01T00:15:00Z","type":"user","message":{"content":"#9 done"}}`,
		})
	withGitRun(t, func(repo string, args ...string) ([]byte, error) { return []byte(""), nil })
	res, err := Compute(Options{
		Dirs: []string{dir}, GitRepo: "/repo",
		SinceISO: wideSince, UntilISO: wideUntil,
		Issues: []string{"8", "9"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !approx(res.TotalActive, 30) {
		t.Fatalf("overlapping no-commit sources should remain separate, got total=%v", res.TotalActive)
	}
	if !approx(res.PerIssue["8"], 15) || !approx(res.PerIssue["9"], 15) {
		t.Fatalf("want #8=15,#9=15, got %v", res.PerIssue)
	}
	if len(res.Segments) != 2 {
		t.Fatalf("want one fallback segment per source, got %+v", res.Segments)
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
	// The source-local run spans the commit; with no later claimant, the inside
	// commit claims it. Explicit prefixWeight 0 must not erase non-prefix work.
	if !approx(res.PerIssue["8"], 25) || res.PerIssue["10"] != 0 {
		t.Fatalf("inside commit should claim the run: want #8=25,#10=0, got %v", res.PerIssue)
	}
}

func TestComputeDiscoversInterveningIssueClaimants(t *testing.T) {
	dir := t.TempDir()
	writeJSONL(t, filepath.Join(dir, "issue2.jsonl"),
		[]string{
			`{"timestamp":"2026-01-01T00:25:00Z","type":"user","message":{"content":"issue 2 work"}}`,
			`{"timestamp":"2026-01-01T00:35:00Z","type":"user","message":{"content":"issue 2 done"}}`,
		})
	writeJSONL(t, filepath.Join(dir, "issue3.jsonl"),
		[]string{
			`{"timestamp":"2026-01-01T00:50:00Z","type":"user","message":{"content":"issue 3 work"}}`,
			`{"timestamp":"2026-01-01T00:58:00Z","type":"user","message":{"content":"issue 3 done"}}`,
		})
	writeJSONL(t, filepath.Join(dir, "issue1.jsonl"),
		[]string{
			`{"timestamp":"2026-01-01T02:40:00Z","type":"user","message":{"content":"issue 1 resumes"}}`,
			`{"timestamp":"2026-01-01T02:55:00Z","type":"user","message":{"content":"issue 1 done"}}`,
		})
	withGitRun(t, func(repo string, args ...string) ([]byte, error) {
		return []byte(strings.Join([]string{
			"aaaaaaaaaaaaaaaa\t2026-01-01T00:00:00Z\t#1 c11",
			"bbbbbbbbbbbbbbbb\t2026-01-01T00:20:00Z\t#2 c21",
			"cccccccccccccccc\t2026-01-01T00:40:00Z\t#2 c22",
			"dddddddddddddddd\t2026-01-01T01:00:00Z\t#3 c31",
			"eeeeeeeeeeeeeeee\t2026-01-01T03:00:00Z\t#1 c12",
		}, "\n")), nil
	})

	res, err := Compute(Options{
		Dirs: []string{dir}, GitRepo: "/repo",
		SinceISO: wideSince, UntilISO: wideUntil,
		Issues: []string{"1"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := res.PerIssue["1"]; !approx(got, 15) {
		t.Fatalf("#1 should receive only its own resumed run, got %v all=%v", got, res.PerIssue)
	}
	if got := res.PerIssue["2"]; !approx(got, 10) {
		t.Fatalf("intervening #2 commit should claim its nearby run, got %v all=%v", got, res.PerIssue)
	}
	if got := res.PerIssue["3"]; !approx(got, 8) {
		t.Fatalf("intervening #3 commit should claim its nearby run, got %v all=%v", got, res.PerIssue)
	}
	if !approx(res.TotalActive, 33) {
		t.Fatalf("total active should preserve all source runs, got %v", res.TotalActive)
	}
}

func TestComputeDominantBoundaryWarning(t *testing.T) {
	segs := []Segment{{
		Start:  tm("2026-01-01T10:00:00Z"),
		End:    tm("2026-01-01T13:00:00Z"),
		Active: 180,
		Commit: &Commit{SHA: "aaa", Subject: "#8 work", Issues: []string{"8"}},
		Alloc:  map[string]float64{"8": 180},
	}}
	warnings := attributionWarnings(segs, map[string]float64{"8": 180})
	if len(warnings) != 1 {
		t.Fatalf("want one warning, got %+v", warnings)
	}
	if warnings[0].Issue != "8" || warnings[0].Reason == "" || warnings[0].Share < 0.99 {
		t.Fatalf("warning should name issue/share/reason, got %+v", warnings[0])
	}
}

func TestComputeMentionFallbackWarning(t *testing.T) {
	segs := []Segment{{
		Start:    tm("2026-01-01T10:00:00Z"),
		End:      tm("2026-01-01T10:10:00Z"),
		Active:   10,
		Mentions: map[string]int{"9": 1},
		Alloc:    map[string]float64{"9": 10},
	}}
	warnings := attributionWarnings(segs, map[string]float64{"9": 10})
	if len(warnings) != 1 {
		t.Fatalf("want mention-fallback warning, got %+v", warnings)
	}
	if warnings[0].Issue != "9" || !strings.Contains(warnings[0].Reason, "fallback") {
		t.Fatalf("warning should describe mention fallback for #9, got %+v", warnings[0])
	}
}

func TestComputeUnattributedFallbackWarning(t *testing.T) {
	segs := []Segment{{
		Start:  tm("2026-01-01T10:00:00Z"),
		End:    tm("2026-01-01T10:10:00Z"),
		Active: 10,
		Alloc:  map[string]float64{UnattributedKey: 10},
	}}
	warnings := attributionWarnings(segs, map[string]float64{UnattributedKey: 10})
	if len(warnings) != 1 {
		t.Fatalf("want unattributed fallback warning, got %+v", warnings)
	}
	if warnings[0].Issue != UnattributedKey || !strings.Contains(warnings[0].Reason, "unattributed") {
		t.Fatalf("warning should describe unattributed fallback, got %+v", warnings[0])
	}
}

func TestComputeNoCommitsMentionFallbackWarning(t *testing.T) {
	dir := eventsDir(t,
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8 start"}}`,
		`{"timestamp":"2026-01-01T00:10:00Z","type":"user","message":{"content":"#8 done"}}`,
	)
	withGitRun(t, func(repo string, args ...string) ([]byte, error) { return []byte(""), nil })
	res, err := Compute(Options{
		Dirs: []string{dir}, GitRepo: "/repo",
		SinceISO: wideSince, UntilISO: wideUntil,
		Issues: []string{"8"}, CommitWeight: 1.0, ThresholdMin: 15, IncludeAssistant: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Warnings) != 1 || !strings.Contains(res.Warnings[0].Reason, "fallback") {
		t.Fatalf("no-commit mention fallback should warn, got %+v", res.Warnings)
	}
}
