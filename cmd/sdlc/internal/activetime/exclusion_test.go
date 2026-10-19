package activetime

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func exTime(minute int) time.Time {
	return tm("2026-01-01T00:00:00Z").Add(time.Duration(minute) * time.Minute)
}

func TestClaimActivityRunsRetainsSource(t *testing.T) {
	runs := []ActivityRun{{Start: exTime(0), End: exTime(10), Active: 10, Source: "session-a", Mentions: map[string]int{"8": 1}}}
	got := claimActivityRuns(runs, nil, 1, 1)
	if len(got) != 1 || got[0].Source != "session-a" {
		t.Fatalf("segment lost run source: %+v", got)
	}
}

func TestSubtractProtectedIntervalsSameSlotUnionAndOtherSlot(t *testing.T) {
	runs := []ActivityRun{
		{Start: exTime(0), End: exTime(60), Active: 60, Source: "same", Mentions: map[string]int{"8": 1}},
		{Start: exTime(0), End: exTime(60), Active: 60, Source: "other", Mentions: map[string]int{"9": 1}},
	}
	protected := []CreditedInterval{
		{Issue: "636", Slot: "ariadne:1", Start: exTime(10), End: exTime(30)},
		{Issue: "636", Slot: "ariadne:1", Start: exTime(20), End: exTime(40)},
	}
	got, excluded, err := subtractProtectedIntervals(runs, protected, map[string]string{"same": "ariadne:1", "other": "ariadne:2"})
	if err != nil {
		t.Fatal(err)
	}
	var same, other float64
	for _, run := range got {
		switch run.Source {
		case "same":
			same += run.Active
		case "other":
			other += run.Active
		}
	}
	if !approx(same, 30) || !approx(other, 60) {
		t.Fatalf("same slot should lose 30 union minutes, other slot none: %+v", got)
	}
	var minutes float64
	for _, diagnostic := range excluded {
		if diagnostic.Issue != "636" || diagnostic.Slot != "ariadne:1" || !diagnostic.End.After(diagnostic.Start) {
			t.Fatalf("bad exclusion diagnostic: %+v", diagnostic)
		}
		minutes += diagnostic.Minutes
	}
	if !approx(minutes, 30) {
		t.Fatalf("diagnostics double counted overlapping protection: %+v", excluded)
	}
}

func TestSubtractProtectedIntervalsUnknownOverlappingSlotFailsClosed(t *testing.T) {
	runs := []ActivityRun{{Start: exTime(0), End: exTime(60), Active: 60, Source: "unknown"}}
	protected := []CreditedInterval{{Issue: "636", Slot: "ariadne:1", Start: exTime(10), End: exTime(20)}}
	if _, _, err := subtractProtectedIntervals(runs, protected, nil); err == nil || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("unknown overlapping source must fail closed: %v", err)
	}
	protected[0].Start, protected[0].End = exTime(60), exTime(70)
	if _, _, err := subtractProtectedIntervals(runs, protected, nil); err != nil {
		t.Fatalf("disjoint unknown source should not block: %v", err)
	}
}

func TestComputeProtectedIntervalBeforeFractionalAllocation(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-01-01T01:00:00Z", "neutral work")
	file := filepath.Join(t.TempDir(), "session.jsonl")
	writeJSONL(t, file, []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8 start"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"user","message":{"content":"#9 done"}}`,
	})
	got, err := Compute(Options{
		Files: []string{file}, GitRepo: repo, SinceISO: "2026-01-01T00:00:00Z", UntilISO: "2026-01-01T01:00:00Z",
		Issues: []string{"8", "9"}, ThresholdMin: 120, CommitWeight: 1, IncludeAssistant: true,
		ProtectedIntervals: []CreditedInterval{{Issue: "636", Slot: "ariadne:1", Start: exTime(15), End: exTime(45)}},
		SourceSlots:        map[string]string{file: "ariadne:1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !approx(got.TotalActive, 30) || !approx(got.PerIssue["8"], 15) || !approx(got.PerIssue["9"], 15) {
		t.Fatalf("excluded minutes must be removed before fractional attribution: %+v", got)
	}
	if len(got.Exclusions) == 0 || !approx(got.Exclusions[0].Minutes, 30) {
		t.Fatalf("missing exclusion diagnostic: %+v", got.Exclusions)
	}
}

func TestComputeProtectedIntervalExcludesOnlyPostQueueActivity(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-01-01T01:00:00Z", "#8 work")
	file := filepath.Join(t.TempDir(), "session.jsonl")
	writeJSONL(t, file, []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8 start"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"user","message":{"content":"#8 done"}}`,
	})
	got, err := Compute(Options{
		Files: []string{file}, GitRepo: repo, SinceISO: "2026-01-01T00:00:00Z", UntilISO: "2026-01-01T01:00:00Z",
		Issues: []string{"8"}, ThresholdMin: 120, CommitWeight: 1, IncludeAssistant: true,
		QueueIntervals: []QueueInterval{{Start: exTime(10), End: exTime(20), Slot: "ariadne:1", AttemptID: "suite"}},
		ProtectedIntervals: []CreditedInterval{{Issue: "636", Slot: "ariadne:1", Start: exTime(15), End: exTime(45)}},
		SourceSlots: map[string]string{file: "ariadne:1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !approx(got.QueueAttribution.ExcludedMinutes, 10) || !approx(got.TotalActive, 25) || !approx(got.PerIssue["8"], 25) {
		t.Fatalf("queue then protection should leave 25 minutes: %+v", got)
	}
	if len(got.Exclusions) != 1 || !approx(got.Exclusions[0].Minutes, 25) || !got.Exclusions[0].Start.Equal(exTime(20)) || !got.Exclusions[0].End.Equal(exTime(45)) {
		t.Fatalf("diagnostic must describe only post-queue removed activity: %+v", got.Exclusions)
	}
}
