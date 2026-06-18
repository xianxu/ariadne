package activetime

import (
	"math"
	"testing"
	"time"
)

func tm(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestActiveMinutes(t *testing.T) {
	if activeMinutes(nil, 15) != 0 {
		t.Fatal("empty should be 0")
	}
	if activeMinutes([]time.Time{tm("2026-01-01T00:00:00Z")}, 15) != 0 {
		t.Fatal("single should be 0")
	}
	// 5-min gap (kept) + 30-min gap (capped to 15) = 20.
	got := activeMinutes([]time.Time{
		tm("2026-01-01T00:00:00Z"),
		tm("2026-01-01T00:05:00Z"),
		tm("2026-01-01T00:35:00Z"),
	}, 15)
	if !approx(got, 20) {
		t.Fatalf("want 20, got %v", got)
	}
	// Unsorted input must yield the same answer (defensive sort).
	got = activeMinutes([]time.Time{
		tm("2026-01-01T00:35:00Z"),
		tm("2026-01-01T00:00:00Z"),
		tm("2026-01-01T00:05:00Z"),
	}, 15)
	if !approx(got, 20) {
		t.Fatalf("unsorted want 20, got %v", got)
	}
}

func TestActiveMinutesUnionFillsSpan(t *testing.T) {
	// Two events 30 min apart: bare gap caps at 15. With a Task span covering the
	// full 30 min, the span fills it → 30 (subagent work, not idle).
	times := []time.Time{tm("2026-01-01T00:00:00Z"), tm("2026-01-01T00:30:00Z")}
	if got := activeMinutesUnion(times, nil, 15); !approx(got, 15) {
		t.Fatalf("bare 30-min gap should cap at 15, got %v", got)
	}
	spans := []TaskSpan{{Start: tm("2026-01-01T00:00:00Z"), End: tm("2026-01-01T00:30:00Z")}}
	if got := activeMinutesUnion(times, spans, 15); !approx(got, 30) {
		t.Fatalf("span-covered 30-min gap should fill to 30, got %v", got)
	}
}

func TestActiveMinutesUnionParityNoSpans(t *testing.T) {
	// Identical to TestActiveMinutes: 5-min kept + 30-min capped = 20.
	times := []time.Time{
		tm("2026-01-01T00:00:00Z"), tm("2026-01-01T00:05:00Z"), tm("2026-01-01T00:35:00Z"),
	}
	if got := activeMinutesUnion(times, nil, 15); !approx(got, 20) {
		t.Fatalf("want 20, got %v", got)
	}
}

func TestActiveMinutesUnionOverlapIsWallClock(t *testing.T) {
	// Two overlapping (parallel) spans union to wall-clock, not summed effort.
	spans := []TaskSpan{
		{Start: tm("2026-01-01T00:00:00Z"), End: tm("2026-01-01T00:20:00Z")},
		{Start: tm("2026-01-01T00:10:00Z"), End: tm("2026-01-01T00:30:00Z")},
	}
	if got := activeMinutesUnion(nil, spans, 15); !approx(got, 30) {
		t.Fatalf("overlapping spans should union to 30, got %v (not 40)", got)
	}
}

func TestClampSpans(t *testing.T) {
	spans := []TaskSpan{{Start: tm("2026-01-01T00:00:00Z"), End: tm("2026-01-01T01:00:00Z")}}
	got := clampSpans(spans, tm("2026-01-01T00:20:00Z"), tm("2026-01-01T00:40:00Z"))
	if len(got) != 1 || !got[0].Start.Equal(tm("2026-01-01T00:20:00Z")) || !got[0].End.Equal(tm("2026-01-01T00:40:00Z")) {
		t.Fatalf("clamp to [00:20,00:40), got %+v", got)
	}
	// Non-overlapping span → dropped.
	if out := clampSpans(spans, tm("2026-01-01T02:00:00Z"), tm("2026-01-01T03:00:00Z")); len(out) != 0 {
		t.Fatalf("non-overlapping span should drop, got %+v", out)
	}
}

func TestAttributeSegmentCommitOnly(t *testing.T) {
	// weight 1.0, two commit issues, no transcript share → 30 each, no unattributed.
	out := attributeSegment(60, []string{"8", "10"}, map[string]int{"8": 3}, 1.0)
	if !approx(out["8"], 30) || !approx(out["10"], 30) {
		t.Fatalf("want 30/30, got %v", out)
	}
	if _, ok := out[UnattributedKey]; ok {
		t.Fatalf("weight 1.0 must not produce unattributed: %v", out)
	}
}

func TestAttributeSegmentMixedWeight(t *testing.T) {
	// weight 0.5, commit #8 → 30; transcript 30 split 1:3 → #8 +7.5, #10 +22.5.
	out := attributeSegment(60, []string{"8"}, map[string]int{"8": 1, "10": 3}, 0.5)
	if !approx(out["8"], 37.5) || !approx(out["10"], 22.5) {
		t.Fatalf("want 37.5/22.5, got %v", out)
	}
}

func TestAttributeSegmentNoCommitMentionOnly(t *testing.T) {
	out := attributeSegment(40, nil, map[string]int{"8": 1, "10": 1}, 1.0)
	if !approx(out["8"], 20) || !approx(out["10"], 20) {
		t.Fatalf("want 20/20, got %v", out)
	}
}

func TestAttributeSegmentNoMentionUnattributed(t *testing.T) {
	out := attributeSegment(40, nil, nil, 1.0)
	if !approx(out[UnattributedKey], 40) {
		t.Fatalf("want 40 unattributed, got %v", out)
	}
}

func TestAttributeSegmentZeroActive(t *testing.T) {
	if out := attributeSegment(0, []string{"8"}, map[string]int{"8": 1}, 1.0); len(out) != 0 {
		t.Fatalf("zero active should allocate nothing, got %v", out)
	}
}

func TestBuildSegmentsPrefixAndAnchor(t *testing.T) {
	// Two commits, both interior to the event span, so the prefix segment and a
	// non-prefix commit-anchored segment are distinct. (With a single commit the
	// first segment is necessarily both the prefix AND its anchor.)
	events := []Event{
		{Time: tm("2026-01-01T00:00:00Z"), Mentions: map[string]int{"8": 1}}, // prefix
		{Time: tm("2026-01-01T00:10:00Z")},                                   // prefix
		{Time: tm("2026-01-01T00:40:00Z"), Mentions: map[string]int{"8": 1}}, // commit1 seg
		{Time: tm("2026-01-01T01:10:00Z"), Mentions: map[string]int{"8": 1}}, // commit2 seg
		{Time: tm("2026-01-01T01:30:00Z")},                                   // suffix
	}
	commits := []Commit{
		{Time: tm("2026-01-01T00:50:00Z"), SHA: "aaa1111", Subject: "#8 m1", Issues: []string{"8"}},
		{Time: tm("2026-01-01T01:20:00Z"), SHA: "bbb2222", Subject: "#8 m2", Issues: []string{"8"}},
	}
	segs := buildSegments(events, commits, 1.0 /*commitWeight*/, 0.5 /*prefixWeight*/, 15 /*thresholdMin*/)
	if len(segs) == 0 {
		t.Fatal("expected segments")
	}
	if !segs[0].IsPrefix {
		t.Fatalf("first segment should be prefix: %+v", segs[0])
	}
	// There must be a non-prefix segment anchored by the second commit.
	var anchored *Segment
	for i := range segs {
		if segs[i].Commit != nil && segs[i].Commit.SHA == "bbb2222" {
			anchored = &segs[i]
		}
	}
	if anchored == nil {
		t.Fatalf("expected a segment anchored by bbb2222, got %+v", segs)
	}
	if anchored.IsPrefix {
		t.Fatalf("the bbb2222-anchored segment should not be the prefix: %+v", anchored)
	}
	// Last segment is the suffix: events after the final commit, no anchor.
	last := segs[len(segs)-1]
	if last.Commit != nil {
		t.Fatalf("suffix segment should have no anchor, got %+v", last.Commit)
	}
}

// When the first event coincides with the first commit there is no prefix.
func TestBuildSegmentsNoPrefix(t *testing.T) {
	at := tm("2026-01-01T00:50:00Z")
	events := []Event{
		{Time: at, Mentions: map[string]int{"8": 1}},
		{Time: tm("2026-01-01T00:55:00Z")},
	}
	commits := []Commit{{Time: at, SHA: "abc1234", Subject: "#8 work", Issues: []string{"8"}}}
	segs := buildSegments(events, commits, 1.0, 0.5, 15)
	for _, s := range segs {
		if s.IsPrefix {
			t.Fatalf("did not expect a prefix segment: %+v", s)
		}
	}
}
