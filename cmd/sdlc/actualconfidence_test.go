package main

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// #603: the closed (status × unattributed kind × bounded) domain maps to one
// confidence, and every rendering is one line naming its level.
func TestClassifyActual(t *testing.T) {
	unknown := gatestate.MeasurementUnattributed{Kind: gatestate.UnattributedUnknownBranch, Bounded: true, Start: "2026-10-24T10:00:00Z", End: "2026-10-24T10:30:00Z", Reason: "concurrent claims have unknown, detached or ambiguous active branch"}
	source := gatestate.MeasurementUnattributed{Kind: gatestate.UnattributedSourceAttribution, Bounded: true, Start: "2026-10-24T11:00:00Z", End: "2026-10-24T11:10:00Z", Reason: "ambiguous historical owner for /w"}
	queueBounded := gatestate.MeasurementUnattributed{Kind: gatestate.UnattributedQueueUnknown, Bounded: true, Start: "2026-10-24T12:00:00Z", End: "2026-10-24T12:05:00Z", Reason: "suite attempt ccdd has unknown slot provenance"}
	queueOpen := gatestate.MeasurementUnattributed{Kind: gatestate.UnattributedQueueUnknown, Reason: "legacy suite attempt aabb lacks exact queue endpoints"}
	measured := func(minutes float64, records ...gatestate.MeasurementUnattributed) actualResult {
		return actualResult{Status: actualMeasured, Hours: 1.5, Inputs: &gatestate.MeasurementInputs{Unattributed: records, UnattributedMinutes: minutes}}
	}
	for _, tc := range []struct {
		name    string
		res     actualResult
		level   confidenceLevel
		share   float64
		bounded bool
		render  string
	}{
		{"measured, nothing unattributed", measured(0), confidenceExact, 1, true, "exact"},
		{"measured, no retained inputs", actualResult{Status: actualMeasured, Hours: 1}, confidenceExact, 1, true, "exact"},
		{"unknown branch", measured(30, unknown), confidencePartial, 90.0 / 120, true, "partial 0.75 (30m unattributed: concurrent claims"},
		{"source attribution", measured(10, source), confidencePartial, 90.0 / 100, true, "partial 0.90 (10m unattributed: ambiguous historical owner"},
		{"bounded queue unknown is an upper bound", measured(0, queueBounded), confidencePartial, 0, false, "partial, upper bound"},
		{"unbounded queue unknown", measured(0, queueOpen), confidencePartial, 0, false, "partial, upper bound (legacy suite attempt aabb"},
		{"mixed, upper bound wins", measured(30, unknown, queueOpen), confidencePartial, 0, false, "partial, upper bound"},
		{"telemetry gap", actualResult{Status: actualTelemetryGap}, confidenceUnmeasured, 0, false, "unmeasured (telemetry gap"},
		{"empty window", actualResult{Status: actualEmptyWindow, Detail: "no activity from ownership-proven transcript sources"}, confidenceUnmeasured, 0, false, "unmeasured (no activity from ownership-proven"},
		{"no window", actualResult{Status: actualNoWindow}, confidenceUnmeasured, 0, false, "unmeasured (no window"},
		{"error with recovery", actualResult{Status: actualError, Detail: "branch source lineage cycle", Recovery: "claim the issue"}, confidenceUnmeasured, 0, false, "unmeasured (branch source lineage cycle; claim the issue)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := classifyActual(tc.res)
			if c.Level != tc.level || math.Abs(c.Share()-tc.share) > 1e-9 || (c.Level != confidenceUnmeasured && c.Bounded != tc.bounded) {
				t.Fatalf("classify = %+v share %v; want %s share %v bounded %v", c, c.Share(), tc.level, tc.share, tc.bounded)
			}
			if got := c.String(); !strings.HasPrefix(got, tc.render) || strings.ContainsAny(got, "\r\n") {
				t.Fatalf("render %q, want prefix %q on one line", got, tc.render)
			}
		})
	}
	multi := actualResult{Status: actualError, Detail: "first\nsecond   third"}
	if got := classifyActual(multi).String(); got != "unmeasured (first second third)" {
		t.Fatalf("a multi-line reason renders %q", got)
	}
}

// Unattributed minutes count a seen minute once, inside the sample and the
// counted phases, and never one this issue's segments already counted.
func TestUnattributedAndOverstatedMinutes(t *testing.T) {
	at := func(m int) time.Time { return time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC).Add(time.Duration(m) * time.Minute) }
	rec := func(kind string, a, b int) gatestate.MeasurementUnattributed {
		return gatestate.MeasurementUnattributed{Kind: kind, Bounded: true, Start: at(a).Format(time.RFC3339), End: at(b).Format(time.RFC3339), Reason: "r"}
	}
	records := []gatestate.MeasurementUnattributed{
		rec(gatestate.UnattributedUnknownBranch, 0, 30),
		rec(gatestate.UnattributedSourceAttribution, 20, 40), // overlaps the first: counted once
		rec(gatestate.UnattributedUnknownBranch, 100, 130),   // outside the sample: clipped away
		rec(gatestate.UnattributedQueueUnknown, 50, 60),      // a queue record is overstatement, not unattributed
		{Kind: gatestate.UnattributedQueueUnknown, Reason: "no endpoints"},
	}
	counted := []activetime.Segment{
		{Start: at(35), End: at(55), Alloc: map[string]float64{"603": 1}}, // this issue counted 35–40 and 50–55
		{Start: at(0), End: at(10), Alloc: map[string]float64{"601": 1}},  // another issue's minute is still ours lost
	}
	if got := unattributedMinutes("603", records, counted, nil, at(0), at(90)); got != 35 {
		t.Fatalf("unattributed = %v, want 0–40 less 35–40 = 35", got)
	}
	if got := overstatedMinutes("603", records, counted, nil, at(0), at(90)); got != 5 {
		t.Fatalf("overstated = %v, want the 50–55 overlap = 5", got)
	}
	phases := []activetime.Phase{{Start: at(25), End: at(45), Kind: activetime.Build}}
	if got := unattributedMinutes("603", records, counted, phases, at(0), at(90)); got != 10 {
		t.Fatalf("milestone unattributed = %v, want 25–40 less 35–40 = 10", got)
	}
	var many []gatestate.MeasurementUnattributed
	for i := 0; i < gatestate.SnapshotLimitUnattributed; i++ {
		many = append(many, rec(gatestate.UnattributedUnknownBranch, i, i+1))
	}
	if kept := retainUnattributed(append(many, records[4])); len(kept) != gatestate.SnapshotLimitUnattributed || kept[0].Bounded {
		t.Fatalf("retained %d records, first bounded=%v: an unbounded record must survive the cap", len(kept), kept[0].Bounded)
	}
}
