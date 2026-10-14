package activetime

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestComputeClipsOwnershipActivity(t *testing.T) {
	dir := eventsDir(t,
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8"}}`,
		`{"timestamp":"2026-01-01T00:10:00Z","type":"user","message":{"content":"#8"}}`,
		`{"timestamp":"2026-01-01T00:20:00Z","type":"user","message":{"content":"#8"}}`,
		`{"timestamp":"2026-01-01T00:30:00Z","type":"user","message":{"content":"#8"}}`,
		`{"timestamp":"2026-01-01T00:40:00Z","type":"user","message":{"content":"#8"}}`,
	)
	withGitRun(t, func(repo string, args ...string) ([]byte, error) { return nil, nil })
	at := func(min int) time.Time { return time.Date(2026, 1, 1, 0, min, 0, 0, time.UTC) }
	res, err := Compute(Options{Dirs: []string{dir}, GitRepo: "/repo", SinceISO: wideSince, UntilISO: wideUntil,
		Issues: []string{"8"}, CommitWeight: 1, ThresholdMin: 15, IncludeAssistant: true,
		ActivityWindows: map[string][]ActivityWindow{filepath.Join(dir, "s.jsonl"): {{Start: at(0), End: at(3)}, {Start: at(30), End: at(35)}}},
	})
	if err != nil || res.PerIssue["8"] != 8 || res.TotalActive != 8 {
		t.Fatalf("engine admitted released activity: %+v %v", res, err)
	}
}

func TestClipActivityOwnershipIntervals(t *testing.T) {
	at := func(min int) time.Time { return time.Date(2026, 10, 13, 0, min, 0, 0, time.UTC) }
	runs := []ActivityRun{
		{Start: at(0), End: at(40), Active: 40, Source: "owner", Mentions: map[string]int{"335": 1}},
		{Start: at(10), End: at(30), Active: 20, Source: "other-slot"},
	}
	before := append([]ActivityRun(nil), runs...)
	got, err := ClipActivity(runs, map[string][]ActivityWindow{
		"owner": {{Start: at(0), End: at(5)}, {Start: at(20), End: at(25)}, {Start: at(23), End: at(30)}},
	})
	if err != nil || len(got) != 2 || got[0].Active != 5 || got[1].Active != 10 || got[1].Start != at(20) || got[1].End != at(30) {
		t.Fatalf("released/unowned time included or overlapping windows double counted: %+v, %v", got, err)
	}
	if !reflect.DeepEqual(runs, before) || got[0].Mentions["335"] != 1 {
		t.Fatal("clipping changed its input or lost attribution context")
	}
}

func TestClipActivityLegacyAndInvalidWindows(t *testing.T) {
	at := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	runs := []ActivityRun{{Start: at, End: at.Add(time.Minute), Active: 1, Source: "owner"}}
	got, err := ClipActivity(runs, nil)
	if err != nil || !reflect.DeepEqual(got, runs) {
		t.Fatalf("legacy unrestricted input changed: %+v %v", got, err)
	}
	got, err = ClipActivity(runs, map[string][]ActivityWindow{})
	if err != nil || len(got) != 0 {
		t.Fatalf("explicitly empty eligibility admitted activity: %+v %v", got, err)
	}
	for _, window := range []ActivityWindow{{Start: at, End: at}, {Start: at, End: at.Add(-time.Minute)}} {
		if _, err := ClipActivity(runs, map[string][]ActivityWindow{"owner": {window}}); err == nil {
			t.Fatal("invalid ownership interval was silently accepted")
		}
	}
}

// A long task span must be clipped after source activity is formed, so a
// release cannot disappear merely because the task's two events straddle it.
func TestClipActivityTaskSpanAcrossRelease(t *testing.T) {
	at := func(min int) time.Time { return time.Date(2026, 10, 13, 0, min, 0, 0, time.UTC) }
	runs := activityRuns(nil, []TaskSpan{{Start: at(0), End: at(40), Source: "owner"}}, 15)
	got, err := ClipActivity(runs, map[string][]ActivityWindow{"owner": {{Start: at(0), End: at(3)}, {Start: at(30), End: at(35)}}})
	var total float64
	for _, run := range got {
		total += run.Active
	}
	if err != nil || total != 8 {
		t.Fatalf("span outside ownership counted: %+v %v", got, err)
	}
}
