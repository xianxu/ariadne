package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// confidenceLevel says how much of a measured actual was attributed (#603).
type confidenceLevel string

const (
	confidenceExact      confidenceLevel = "exact"      // every seen minute attributed
	confidencePartial    confidenceLevel = "partial"    // some seen minutes not attributed
	confidenceUnmeasured confidenceLevel = "unmeasured" // no number at all
)

// actualConfidence is derived from a measurement, never typed. Partial hours
// exclude every minute whose attribution is unproven but whose extent is known,
// so they are a lower bound. An unknown with no extent (an unbounded queue
// attempt) makes Bounded false: the hours are then an upper bound, the share 0,
// and calibration gives the row no weight.
type actualConfidence struct {
	Level               confidenceLevel
	AttributedMinutes   float64
	UnattributedMinutes float64
	// OverstatedMinutes is how much of the counted time unknown queue
	// attempts overlap: the most an upper bound may overstate.
	OverstatedMinutes float64
	Bounded           bool
	Reasons           []string
}

// Share is the attributed fraction of the seen minutes: 1 when exact, 0 when
// unmeasured or unbounded.
func (c actualConfidence) Share() float64 {
	switch {
	case c.Level == confidenceExact:
		return 1
	case c.Level == confidenceUnmeasured || !c.Bounded:
		return 0
	case c.AttributedMinutes+c.UnattributedMinutes <= 0:
		return 0
	}
	return c.AttributedMinutes / (c.AttributedMinutes + c.UnattributedMinutes)
}

// String renders the one line close, milestone-close and sdlc actual print and
// the evidence trailer records.
func (c actualConfidence) String() string {
	reasons := strings.Join(c.Reasons, "; ")
	switch {
	case c.Level == confidenceExact:
		return "exact"
	case c.Level == confidenceUnmeasured:
		return "unmeasured (" + reasons + ")"
	case !c.Bounded && c.OverstatedMinutes > 0:
		return fmt.Sprintf("partial, upper bound, may overstate by up to %.0fm (%s)", c.OverstatedMinutes, reasons)
	case !c.Bounded:
		return "partial, upper bound (" + reasons + ")"
	}
	return fmt.Sprintf("partial %.2f (%.0fm unattributed: %s)", c.Share(), c.UnattributedMinutes, reasons)
}

// Token is the confidence as a bare machine value: exact, partial <share>,
// partial upper-bound or unmeasured. It is what frontmatter persists, so it
// holds no reason text (no colon or hash a YAML scalar would misread).
func (c actualConfidence) Token() string {
	switch {
	case c.Level == confidencePartial && !c.Bounded:
		return "partial upper-bound"
	case c.Level == confidencePartial:
		return "partial " + c.ShareField()
	}
	return string(c.Level)
}

// ShareField is the ledger's attributed_share column.
func (c actualConfidence) ShareField() string {
	return strconv.FormatFloat(c.Share(), 'f', 2, 64)
}

// maxConfidenceReasons bounds how many distinct reasons a rendering names.
const maxConfidenceReasons = 3

// classifyActual is the one mapping from a measurement outcome to its
// confidence. Resting and stage exclusions never lower it: they are not this
// issue's work and are shown as excluded minutes (#591, #594).
func classifyActual(res actualResult) actualConfidence {
	if res.Status != actualMeasured {
		reason := res.Detail
		if reason == "" {
			reason = res.Status.String()
		}
		if res.Recovery != "" {
			reason += "; " + res.Recovery
		}
		return actualConfidence{Level: confidenceUnmeasured, Reasons: []string{confidenceReason(reason)}}
	}
	c := actualConfidence{Level: confidenceExact, AttributedMinutes: res.Hours * 60, Bounded: true}
	if res.Inputs == nil {
		return c
	}
	var reasons []string
	seen := map[string]bool{}
	for _, u := range res.Inputs.Unattributed {
		// An unknown queue attempt's wait is never subtracted, so even a
		// bounded one leaves the hours an upper bound (#603, D4).
		if !u.Bounded || u.Kind == gatestate.UnattributedQueueUnknown {
			c.Bounded = false
		}
		if r := confidenceReason(u.Reason); !seen[r] && len(reasons) < maxConfidenceReasons {
			seen[r] = true
			reasons = append(reasons, r)
		}
	}
	c.UnattributedMinutes = res.Inputs.UnattributedMinutes
	c.OverstatedMinutes = res.OverstatedMinutes
	if c.UnattributedMinutes > 0 || !c.Bounded {
		c.Level, c.Reasons = confidencePartial, reasons
	}
	return c
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// maxReasonBytes bounds one reason in a rendering, which reaches the evidence
// commit's trailer and so the pushed history.
const maxReasonBytes = 240

// confidenceReason is one bounded, single-line reason. A path under the home
// directory is shown from ~, so the trailer names no account layout.
func confidenceReason(s string) string {
	s = oneLine(s)
	if home, err := os.UserHomeDir(); err == nil && len(home) > 1 {
		s = strings.ReplaceAll(s, home+string(filepath.Separator), "~"+string(filepath.Separator))
	}
	if len(s) > maxReasonBytes {
		s = strings.ToValidUTF8(s[:maxReasonBytes-3], "") + "..."
	}
	return s
}

type span struct{ start, end time.Time }

func clipSpans(in []span, lo, hi time.Time) []span {
	var out []span
	for _, x := range in {
		if x.start.Before(lo) {
			x.start = lo
		}
		if x.end.After(hi) {
			x.end = hi
		}
		if x.end.After(x.start) {
			out = append(out, x)
		}
	}
	return out
}

func unionSpans(in []span) []span {
	in = append([]span(nil), in...)
	sort.Slice(in, func(i, j int) bool { return in[i].start.Before(in[j].start) })
	var out []span
	for _, x := range in {
		if n := len(out) - 1; n >= 0 && !x.start.After(out[n].end) {
			if x.end.After(out[n].end) {
				out[n].end = x.end
			}
			continue
		}
		out = append(out, x)
	}
	return out
}

// spanMinutes and overlapMinutes take unions.
func spanMinutes(in []span) float64 {
	total := 0.0
	for _, x := range in {
		total += x.end.Sub(x.start).Minutes()
	}
	return total
}

func overlapMinutes(a, b []span) float64 {
	total := 0.0
	for _, x := range a {
		total += spanMinutes(clipSpans(b, x.start, x.end))
	}
	return total
}

// recordSpans is the union of the bounded records of the kinds wanted, inside
// the sample and, for a milestone, inside the phases it counts.
func recordSpans(records []gatestate.MeasurementUnattributed, queue bool, phases []activetime.Phase, since, until time.Time) []span {
	var seen []span
	for _, u := range records {
		if !u.Bounded || (u.Kind == gatestate.UnattributedQueueUnknown) != queue {
			continue
		}
		start, e1 := time.Parse(time.RFC3339Nano, u.Start)
		end, e2 := time.Parse(time.RFC3339Nano, u.End)
		if e1 == nil && e2 == nil {
			seen = append(seen, span{start, end})
		}
	}
	seen = unionSpans(clipSpans(seen, since, until))
	if len(phases) == 0 {
		return seen
	}
	var limits, in []span
	for _, p := range phases {
		limits = append(limits, span{p.Start, p.End})
	}
	for _, l := range unionSpans(limits) {
		in = append(in, clipSpans(seen, l.start, l.end)...)
	}
	return unionSpans(in)
}

// issueCover is the wall time this issue's counted segments cover.
func issueCover(issue string, counted []activetime.Segment) []span {
	var cover []span
	for _, s := range counted {
		if s.Alloc[issue] > 0 {
			cover = append(cover, span{s.Start, s.End})
		}
	}
	return unionSpans(cover)
}

// unattributedMinutes is the wall cover of the bounded non-queue records, less
// the time this issue's counted segments already cover: a minute another
// source counted for this issue is not lost. phases limits it to what a
// milestone counts; none means the whole sample.
func unattributedMinutes(issue string, records []gatestate.MeasurementUnattributed, counted []activetime.Segment, phases []activetime.Phase, since, until time.Time) float64 {
	seen := recordSpans(records, false, phases, since, until)
	return spanMinutes(seen) - overlapMinutes(seen, issueCover(issue, counted))
}

// overstatedMinutes is the counted time unknown queue attempts overlap: their
// waits were not subtracted, so this is the most the hours may overstate.
func overstatedMinutes(issue string, records []gatestate.MeasurementUnattributed, counted []activetime.Segment, phases []activetime.Phase, since, until time.Time) float64 {
	return overlapMinutes(recordSpans(records, true, phases, since, until), issueCover(issue, counted))
}

// retainUnattributed keeps a bounded prefix of the records for the snapshot;
// the minutes field carries the whole total. Unbounded records come first, so
// a retained snapshot still says when its hours are an upper bound.
func retainUnattributed(records []gatestate.MeasurementUnattributed) []gatestate.MeasurementUnattributed {
	out := append([]gatestate.MeasurementUnattributed(nil), records...)
	sort.SliceStable(out, func(i, j int) bool { return !out[i].Bounded && out[j].Bounded })
	if len(out) > gatestate.SnapshotLimitUnattributed {
		out = out[:gatestate.SnapshotLimitUnattributed]
	}
	return out
}

// String names a status for an unmeasured reason that carries no detail.
func (s actualStatus) String() string {
	switch s {
	case actualMeasured:
		return "measured"
	case actualTelemetryGap:
		return "telemetry gap: commits exist but no transcript events"
	case actualEmptyWindow:
		return "empty window: no attributed activity"
	case actualNoWindow:
		return "no window: no commits reference the issue"
	}
	return "measurement error"
}
