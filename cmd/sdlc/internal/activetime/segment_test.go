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
	segs := buildSegments(events, commits, nil, 1.0 /*commitWeight*/, 0.5 /*prefixWeight*/, 15 /*thresholdMin*/)
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
	// Last segment is after the final commit; the global-boundary model uses the
	// previous commit fallback rather than leaving suffix work unanchored.
	last := segs[len(segs)-1]
	if last.Commit == nil || last.Commit.SHA != "bbb2222" {
		t.Fatalf("suffix segment should fall back to previous commit bbb2222, got %+v", last.Commit)
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
	segs := buildSegments(events, commits, nil, 1.0, 0.5, 15)
	for _, s := range segs {
		if s.IsPrefix {
			t.Fatalf("did not expect a prefix segment: %+v", s)
		}
	}
}

func TestBuildSegmentsFillsSpan(t *testing.T) {
	// One commit; a 40-min Agent span sits in the suffix between two events that
	// are 40 min apart (bare gap would cap at 15). The span fills it.
	events := []Event{
		{Time: tm("2026-01-01T00:50:00Z"), Mentions: map[string]int{"8": 1}},
		{Time: tm("2026-01-01T01:00:00Z")},                                   // dispatch event
		{Time: tm("2026-01-01T01:40:00Z"), Mentions: map[string]int{"8": 1}}, // next turn after return
	}
	commits := []Commit{{Time: tm("2026-01-01T00:50:00Z"), SHA: "aaa", Subject: "#8", Issues: []string{"8"}}}
	spans := []TaskSpan{{Start: tm("2026-01-01T01:00:00Z"), End: tm("2026-01-01T01:40:00Z")}}
	withSpan := buildSegments(events, commits, spans, 1.0, 0.5, 15)
	noSpan := buildSegments(events, commits, nil, 1.0, 0.5, 15)
	var aw, an float64
	for _, s := range withSpan {
		aw += s.Active
	}
	for _, s := range noSpan {
		an += s.Active
	}
	// noSpan: 10 (kept) + 15 (capped) = 25. withSpan: 10 + 40 (filled) = 50.
	if !approx(an, 25) || !approx(aw, 50) {
		t.Fatalf("want noSpan=25 withSpan=50, got %v / %v", an, aw)
	}
}

func TestBuildSegmentsCommitInsideSpan(t *testing.T) {
	// The blocker case: a commit lands STRICTLY INSIDE a span (a subagent that
	// commits mid-run). The post-commit tail segment has NO events (the return is
	// a dropped tool_result), so it must not be skipped — the full span counts and
	// attributes to the commits anchoring each piece.
	events := []Event{
		{Time: tm("2026-01-01T00:00:00Z"), Mentions: map[string]int{"8": 1}}, // dispatch event
		{Time: tm("2026-01-01T00:50:00Z"), Mentions: map[string]int{"8": 1}}, // next turn (after return)
	}
	commits := []Commit{
		{Time: tm("2026-01-01T00:10:00Z"), SHA: "aaa", Subject: "#8 mid", Issues: []string{"8"}}, // INSIDE span
		{Time: tm("2026-01-01T00:50:00Z"), SHA: "bbb", Subject: "#8 end", Issues: []string{"8"}},
	}
	spans := []TaskSpan{{Start: tm("2026-01-01T00:00:00Z"), End: tm("2026-01-01T00:40:00Z")}}
	segs := buildSegments(events, commits, spans, 1.0 /*commitWeight*/, 1.0 /*prefixWeight*/, 15)
	var tot, toIssue8 float64
	for _, s := range segs {
		tot += s.Active
		toIssue8 += s.Alloc["8"]
	}
	// Span is 40 min; the [00:00,00:10) piece (10) + [00:10,00:40) tail (30) must
	// both count → 40 total, all attributed to #8 (weight 1.0, both anchors #8).
	if !approx(tot, 40) {
		t.Fatalf("commit-inside-span: want 40 total active, got %v (tail dropped?)", tot)
	}
	if !approx(toIssue8, 40) {
		t.Fatalf("commit-inside-span: want 40 min → #8, got %v", toIssue8)
	}
}

func TestBuildSegmentsSpanTailPastLastEvent(t *testing.T) {
	// Dispatch is the LAST event; the return lands 30 min later (no further
	// event). The final boundary must extend so the tail is not cut.
	events := []Event{
		{Time: tm("2026-01-01T00:00:00Z"), Mentions: map[string]int{"8": 1}},
		{Time: tm("2026-01-01T00:05:00Z")}, // dispatch, last event
	}
	commits := []Commit{{Time: tm("2026-01-01T00:00:00Z"), SHA: "aaa", Subject: "#8", Issues: []string{"8"}}}
	spans := []TaskSpan{{Start: tm("2026-01-01T00:05:00Z"), End: tm("2026-01-01T00:35:00Z")}}
	segs := buildSegments(events, commits, spans, 1.0, 0.5, 15)
	var tot float64
	for _, s := range segs {
		tot += s.Active
	}
	// 5-min gap kept + 30-min span filled = 35.
	if !approx(tot, 35) {
		t.Fatalf("want 35 (tail not cut), got %v", tot)
	}
}

func TestActivityRunsUnionWithinSourceOnly(t *testing.T) {
	events := []Event{
		{Time: tm("2026-01-01T10:00:00Z"), Source: "a"},
		{Time: tm("2026-01-01T10:15:00Z"), Source: "a"},
		{Time: tm("2026-01-01T10:00:00Z"), Source: "b"},
		{Time: tm("2026-01-01T10:15:00Z"), Source: "b"},
	}
	runs := activityRuns(events, nil, 15)
	if len(runs) != 2 {
		t.Fatalf("same-time intervals from two sources should remain two runs, got %+v", runs)
	}
	for _, r := range runs {
		if !approx(r.Active, 15) {
			t.Fatalf("each run should carry 15 active minutes, got %+v", runs)
		}
	}

	sameSource := []TaskSpan{
		{Start: tm("2026-01-01T11:00:00Z"), End: tm("2026-01-01T11:20:00Z"), Source: "a"},
		{Start: tm("2026-01-01T11:10:00Z"), End: tm("2026-01-01T11:30:00Z"), Source: "a"},
	}
	runs = activityRuns(nil, sameSource, 15)
	if len(runs) != 1 || !approx(runs[0].Active, 30) {
		t.Fatalf("overlaps within one source should union to one 30-min run, got %+v", runs)
	}
}

func TestBuildSegments_GlobalBoundariesPreventLongIssueAbsorption(t *testing.T) {
	commits := []Commit{
		{Time: tm("2026-01-01T00:00:00Z"), SHA: "c11", Subject: "#1 c11", Issues: []string{"1"}},
		{Time: tm("2026-01-01T00:20:00Z"), SHA: "c21", Subject: "#2 c21", Issues: []string{"2"}},
		{Time: tm("2026-01-01T00:40:00Z"), SHA: "c22", Subject: "#2 c22", Issues: []string{"2"}},
		{Time: tm("2026-01-01T01:00:00Z"), SHA: "c31", Subject: "#3 c31", Issues: []string{"3"}},
		{Time: tm("2026-01-01T01:20:00Z"), SHA: "c23", Subject: "#2 c23", Issues: []string{"2"}},
		{Time: tm("2026-01-01T01:40:00Z"), SHA: "c32", Subject: "#3 c32", Issues: []string{"3"}},
		{Time: tm("2026-01-01T03:00:00Z"), SHA: "c12", Subject: "#1 c12", Issues: []string{"1"}},
	}
	events := []Event{
		{Time: tm("2026-01-01T00:05:00Z"), Source: "issue1a"},
		{Time: tm("2026-01-01T00:10:00Z"), Source: "issue1a"},
		{Time: tm("2026-01-01T00:25:00Z"), Source: "issue2a"},
		{Time: tm("2026-01-01T00:35:00Z"), Source: "issue2a"},
		{Time: tm("2026-01-01T01:10:00Z"), Source: "issue2b"},
		{Time: tm("2026-01-01T01:15:00Z"), Source: "issue2b"},
		{Time: tm("2026-01-01T01:30:00Z"), Source: "issue3"},
		{Time: tm("2026-01-01T01:35:00Z"), Source: "issue3"},
		{Time: tm("2026-01-01T02:45:00Z"), Source: "issue1b"},
		{Time: tm("2026-01-01T03:00:00Z"), Source: "issue1b"},
	}
	segs := buildSegments(events, commits, nil, 1.0, 1.0, 15)
	if got := sumAlloc(segs, "1"); !approx(got, 20) {
		t.Fatalf("#1 should receive only adjacent issue1 runs, got %v segs=%+v", got, segs)
	}
	if got := sumAlloc(segs, "2"); !approx(got, 15) {
		t.Fatalf("#2 should receive c22/c23-nearest runs, got %v segs=%+v", got, segs)
	}
	if got := sumAlloc(segs, "3"); !approx(got, 5) {
		t.Fatalf("#3 should receive c32-nearest run, got %v segs=%+v", got, segs)
	}
	var totalActive, totalAllocated float64
	for _, s := range segs {
		totalActive += s.Active
		for _, mins := range s.Alloc {
			totalAllocated += mins
		}
	}
	if !approx(totalActive, 40) || !approx(totalAllocated, totalActive) {
		t.Fatalf("allocation should conserve active minutes: active=%v allocated=%v segs=%+v", totalActive, totalAllocated, segs)
	}
}

func TestClaimActivityRuns_PrefersNextCommitOnTie(t *testing.T) {
	runs := []ActivityRun{{
		Start:  tm("2026-01-01T10:10:00Z"),
		End:    tm("2026-01-01T10:20:00Z"),
		Active: 10,
	}}
	commits := []Commit{
		{Time: tm("2026-01-01T10:00:00Z"), SHA: "prev", Subject: "#1 prev", Issues: []string{"1"}},
		{Time: tm("2026-01-01T10:30:00Z"), SHA: "next", Subject: "#2 next", Issues: []string{"2"}},
	}
	segs := claimActivityRuns(runs, commits, 1.0, 1.0)
	if got := sumAlloc(segs, "2"); !approx(got, 10) {
		t.Fatalf("tie should prefer next commit #2, got #2=%v segs=%+v", got, segs)
	}
	if got := sumAlloc(segs, "1"); got != 0 {
		t.Fatalf("tie should not allocate to previous commit #1, got %v", got)
	}
}

func TestClaimActivityRuns_PreviousCommitFallback(t *testing.T) {
	runs := []ActivityRun{{
		Start:  tm("2026-01-01T10:10:00Z"),
		End:    tm("2026-01-01T10:20:00Z"),
		Active: 10,
	}}
	commits := []Commit{{Time: tm("2026-01-01T10:00:00Z"), SHA: "prev", Subject: "#1 prev", Issues: []string{"1"}}}
	segs := claimActivityRuns(runs, commits, 1.0, 1.0)
	if got := sumAlloc(segs, "1"); !approx(got, 10) {
		t.Fatalf("previous issue commit should claim when no next issue commit exists, got %v segs=%+v", got, segs)
	}
}

func TestClaimActivityRuns_NeutralCommitCutsButDoesNotClaim(t *testing.T) {
	runs := []ActivityRun{{
		Start:    tm("2026-01-01T10:10:00Z"),
		End:      tm("2026-01-01T10:20:00Z"),
		Active:   10,
		Mentions: map[string]int{"9": 1},
	}}
	commits := []Commit{
		{Time: tm("2026-01-01T10:00:00Z"), SHA: "prev", Subject: "#1 prev", Issues: []string{"1"}},
		{Time: tm("2026-01-01T10:05:00Z"), SHA: "neutral", Subject: "chore: no refs"},
	}
	segs := claimActivityRuns(runs, commits, 1.0, 1.0)
	if got := sumAlloc(segs, "1"); got != 0 {
		t.Fatalf("neutral commit should cut off previous #1 claimant, got #1=%v segs=%+v", got, segs)
	}
	if got := sumAlloc(segs, "9"); !approx(got, 10) {
		t.Fatalf("neutral commit should not claim; run should fall back to mentions, got #9=%v segs=%+v", got, segs)
	}
}

func TestClaimActivityRuns_BoundarySuppressesMentionAllocation(t *testing.T) {
	runs := []ActivityRun{{
		Start:    tm("2026-01-01T10:00:00Z"),
		End:      tm("2026-01-01T10:10:00Z"),
		Active:   10,
		Mentions: map[string]int{"9": 1},
	}}
	commits := []Commit{{Time: tm("2026-01-01T10:12:00Z"), SHA: "next", Subject: "#8 work", Issues: []string{"8"}}}
	segs := claimActivityRuns(runs, commits, 0.5, 0.5)
	if got := sumAlloc(segs, "8"); !approx(got, 5) {
		t.Fatalf("commit-weighted share should go to #8, got %v segs=%+v", got, segs)
	}
	if got := sumAlloc(segs, "9"); got != 0 {
		t.Fatalf("mentions must not claim when a boundary exists, got #9=%v segs=%+v", got, segs)
	}
	if got := sumAlloc(segs, UnattributedKey); !approx(got, 5) {
		t.Fatalf("non-commit share should be unattributed, got %v segs=%+v", got, segs)
	}
}

func TestClaimActivityRuns_MentionFallbackOnlyWithoutIssueBoundary(t *testing.T) {
	runs := []ActivityRun{{
		Start:    tm("2026-01-01T10:00:00Z"),
		End:      tm("2026-01-01T10:10:00Z"),
		Active:   10,
		Mentions: map[string]int{"9": 1},
	}}
	commits := []Commit{{Time: tm("2026-01-01T10:05:00Z"), SHA: "neutral", Subject: "chore"}}
	segs := claimActivityRuns(runs, commits, 1.0, 1.0)
	if got := sumAlloc(segs, "9"); !approx(got, 10) {
		t.Fatalf("mentions should claim only without issue commit boundaries, got %v segs=%+v", got, segs)
	}
}

func TestBuildSegments_ParallelRunsCanAttributeToDifferentIssues(t *testing.T) {
	events := []Event{
		{Time: tm("2026-01-01T10:00:00Z"), Source: "a"},
		{Time: tm("2026-01-01T10:15:00Z"), Source: "a"},
		{Time: tm("2026-01-01T10:10:00Z"), Source: "b"},
		{Time: tm("2026-01-01T10:25:00Z"), Source: "b"},
	}
	commits := []Commit{
		{Time: tm("2026-01-01T10:16:00Z"), SHA: "aaa", Subject: "#8 work", Issues: []string{"8"}},
		{Time: tm("2026-01-01T10:26:00Z"), SHA: "bbb", Subject: "#9 work", Issues: []string{"9"}},
	}
	segs := buildSegments(events, commits, nil, 1.0, 1.0, 15)
	if got := sumAlloc(segs, "8"); !approx(got, 15) {
		t.Fatalf("#8 should receive its own overlapping run, got %v segs=%+v", got, segs)
	}
	if got := sumAlloc(segs, "9"); !approx(got, 15) {
		t.Fatalf("#9 should receive its own overlapping run, got %v segs=%+v", got, segs)
	}
	var total float64
	for _, s := range segs {
		total += s.Active
	}
	if !approx(total, 30) {
		t.Fatalf("parallel runs should sum per-issue work to 30 min, got %v segs=%+v", total, segs)
	}
}

func sumAlloc(segs []Segment, issue string) float64 {
	var total float64
	for _, s := range segs {
		total += s.Alloc[issue]
	}
	return total
}
