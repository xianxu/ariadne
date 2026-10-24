package activetime

import (
	"fmt"
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

func TestComputeAdmitsClaudeTaskDispatchBeforeOwnershipUnion(t *testing.T) {
	at := func(min int) time.Time { return time.Date(2026, 1, 1, 0, min, 0, 0, time.UTC) }
	for _, tc := range []struct {
		name     string
		dispatch int
		since    string
		want     float64
		legacy   bool
	}{
		{name: "dispatch during release", dispatch: 10, want: 0},
		{name: "dispatch in earlier engagement", dispatch: 1, want: 2},
		{name: "dispatch in current engagement", dispatch: 31, want: 4},
		{name: "dispatch exactly at release", dispatch: 3, want: 0},
		{name: "original dispatch predates measurement", dispatch: 1, since: "2026-01-01T00:30:00Z", want: 0},
		{name: "legacy unrestricted", dispatch: 1, want: 39, legacy: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := eventsDir(t,
				fmt.Sprintf(`{"timestamp":"%s","type":"assistant","message":{"content":[{"type":"tool_use","id":"task","name":"Agent","input":{}}]}}`, at(tc.dispatch).Format(time.RFC3339)),
				`{"timestamp":"2026-01-01T00:40:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"task","content":"done"}]}}`,
				`{"timestamp":"2026-01-01T00:50:00Z","type":"user","message":{"content":"#8 done"}}`,
			)
			withGitRun(t, func(string, ...string) ([]byte, error) { return nil, nil })
			windows := map[string][]ActivityWindow{filepath.Join(dir, "s.jsonl"): {{Start: at(0), End: at(3)}, {Start: at(30), End: at(35)}}}
			if tc.legacy {
				windows = nil
			}
			since := wideSince
			if tc.since != "" {
				since = tc.since
			}
			res, err := Compute(Options{Dirs: []string{dir}, GitRepo: "/repo", SinceISO: since, UntilISO: wideUntil, Issues: []string{"8"}, CommitWeight: 1, ThresholdMin: 15, IncludeAssistant: true, ActivityWindows: windows})
			if err != nil || res.TotalActive != tc.want {
				t.Fatalf("dispatch %d: active=%v want=%v err=%v", tc.dispatch, res.TotalActive, tc.want, err)
			}
		})
	}
}

func TestAdmitTaskSpansRetainsRelocationButEndsAtReclaim(t *testing.T) {
	at := func(min int) time.Time { return time.Date(2026, 1, 1, 0, min, 0, 0, time.UTC) }
	spans := []TaskSpan{{Start: at(1), End: at(40), Source: "root"}}
	windows := map[string][]ActivityWindow{"root": {{Start: at(0), End: at(3), Began: at(0)}, {Start: at(3), End: at(10), Began: at(0)}, {Start: at(10), End: at(20), Began: at(10)}}}
	admitted, err := AdmitTaskSpans(spans, windows)
	if err != nil {
		t.Fatal(err)
	}
	runs := activityRuns(nil, admitted, 15)
	var total float64
	for _, run := range runs {
		total += run.Active
	}
	if total != 9 {
		t.Fatalf("dispatch crossed reclaim or lost relocation: %+v total=%v", admitted, total)
	}
	if len(spans) != 1 || spans[0].End != at(40) {
		t.Fatal("admission mutated input")
	}
}

func TestComputePartitionsTaskSpanAfterOriginalDispatchAdmission(t *testing.T) {
	dir := eventsDir(t,
		`{"timestamp":"2026-01-01T00:05:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"task","name":"Agent","input":{}}]}}`,
		`{"timestamp":"2026-01-01T00:30:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"task","content":"done"}]}}`,
	)
	withGitRun(t, func(string, ...string) ([]byte, error) { return nil, nil })
	at := func(m int) time.Time { return time.Date(2026, 1, 1, 0, m, 0, 0, time.UTC) }
	source := filepath.Join(dir, "s.jsonl")
	opts := Options{Dirs: []string{dir}, GitRepo: "/repo", SinceISO: wideSince, UntilISO: wideUntil, Issues: []string{"8"}, ThresholdMin: 15, IncludeAssistant: true,
		ActivityWindows: map[string][]ActivityWindow{source: {{Start: at(20), End: at(30), Began: at(2)}}},
		DispatchWindows: map[string][]ActivityWindow{source: {{Start: at(2), End: at(30), Began: at(2)}}},
	}
	got, err := Compute(opts)
	if err != nil || got.TotalActive != 10 {
		t.Fatalf("original dispatch lost or foreign prefix counted: %+v %v", got, err)
	}
	// A later claim never revives an earlier task, even if branch selection
	// happens to admit the same later slice.
	opts.DispatchWindows = map[string][]ActivityWindow{source: {{Start: at(20), End: at(30), Began: at(20)}}}
	got, err = Compute(opts)
	if err != nil || got.TotalActive != 0 {
		t.Fatalf("reclaim revived a pre-claim dispatch: %+v %v", got, err)
	}
}
