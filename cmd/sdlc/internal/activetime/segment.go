package activetime

import (
	"sort"
	"time"
)

// Segment is the unit of attribution: an [Start,End) event span, its gap-
// truncated active minutes, and the resulting per-issue allocation. There is one
// optional prefix segment (events before the first commit), one segment per
// commit, and one optional suffix (events after the last commit).
type Segment struct {
	Start, End time.Time
	Active     float64
	Commit     *Commit
	Mentions   map[string]int
	Alloc      map[string]float64
	IsPrefix   bool
}

// UnattributedKey buckets transcript share that has no mention signal. Stored
// with a leading '#' as a sentinel so it never collides with a numeric issue
// key; renderers display it as "unattributed" (the Python's f"#{iss}" doubled it
// to "##unattributed" — a cosmetic bug we don't reproduce). Exported so the CLI
// renderer can recognize the bucket.
const UnattributedKey = "#unattributed"

// sortTimes sorts a slice of times ascending, in place.
func sortTimes(ts []time.Time) {
	sort.Slice(ts, func(i, j int) bool { return ts[i].Before(ts[j]) })
}

// interval is a half-open [s,e) time span. Active minutes are computed as a
// UNION of intervals (capped inter-event gaps ∪ full-length task spans) so that
// overlapping/parallel subagent spans collapse to wall-clock, not summed effort.
type interval struct{ s, e time.Time }

// unionMinutes merges overlapping intervals and returns the total covered
// duration in minutes. Pure. Touching intervals (iv.s == cur.e) merge, which is
// load-bearing for parity: consecutive uncapped gaps are contiguous, so their
// merged length equals the sum.
func unionMinutes(ivals []interval) float64 {
	if len(ivals) == 0 {
		return 0
	}
	sort.Slice(ivals, func(i, j int) bool { return ivals[i].s.Before(ivals[j].s) })
	var total time.Duration
	cur := ivals[0]
	for _, iv := range ivals[1:] {
		if iv.s.After(cur.e) {
			total += cur.e.Sub(cur.s)
			cur = iv
		} else if iv.e.After(cur.e) {
			cur.e = iv.e
		}
	}
	total += cur.e.Sub(cur.s)
	return total.Minutes()
}

// activeMinutesUnion is the single gap-math core: each inter-event gap counts up
// to thresholdMin (idle truncation), each task span counts in full, and the
// result is the UNION of those intervals (#118). With spans == nil the gap
// intervals are adjacent and non-overlapping, so the union equals the old
// sum-of-capped-gaps — parity is exact (activeMinutes delegates here).
func activeMinutesUnion(times []time.Time, spans []TaskSpan, thresholdMin int) float64 {
	sorted := make([]time.Time, len(times))
	copy(sorted, times)
	sortTimes(sorted)
	capGap := time.Duration(thresholdMin) * time.Minute
	var ivals []interval
	for i := 1; i < len(sorted); i++ {
		end := sorted[i]
		if sorted[i].Sub(sorted[i-1]) > capGap {
			end = sorted[i-1].Add(capGap)
		}
		ivals = append(ivals, interval{sorted[i-1], end})
	}
	for _, sp := range spans {
		if sp.End.After(sp.Start) {
			ivals = append(ivals, interval{sp.Start, sp.End})
		}
	}
	return unionMinutes(ivals)
}

// clampSpans intersects each span with [start,end), dropping empties, so a span
// that straddles a commit boundary is split across segments (each counts only
// its portion — no double-count when the per-segment actives are summed). The
// per-segment-sum-equals-whole-window-union invariant requires that a segment
// carrying a clamped span is never skipped — see buildSegments.
func clampSpans(spans []TaskSpan, start, end time.Time) []TaskSpan {
	var out []TaskSpan
	for _, sp := range spans {
		s, e := sp.Start, sp.End
		if s.Before(start) {
			s = start
		}
		if e.After(end) {
			e = end
		}
		if e.After(s) {
			out = append(out, TaskSpan{Start: s, End: e})
		}
	}
	return out
}

// activeMinutes sums inter-event gaps, each capped at thresholdMin, and returns
// minutes. Mirrors active-time-v3.py active_minutes — the v2.1 active-time
// procedure. Now a thin wrapper over activeMinutesUnion (no task spans), so the
// gap-math has a single implementation (ARCH-DRY); behavior is unchanged.
func activeMinutes(times []time.Time, thresholdMin int) float64 {
	return activeMinutesUnion(times, nil, thresholdMin)
}

// attributeSegment allocates active minutes of one segment per the v3 rule,
// returning {issue: minutes}. Mirrors active-time-v3.py attribute_segment:
//
//   - commit-named issues split weight·active equally;
//   - mentioned issues split (1-weight)·active proportionally to mention count;
//   - with no commit signal, the full segment goes by mention;
//   - transcript share with no mentions falls into the unattributed bucket.
func attributeSegment(active float64, commitIssues []string, mentions map[string]int, weight float64) map[string]float64 {
	out := map[string]float64{}
	if active <= 0 {
		return out
	}
	var transcriptShare float64
	if len(commitIssues) > 0 {
		perCommit := weight * active / float64(len(commitIssues))
		for _, iss := range commitIssues {
			out[iss] += perCommit
		}
		transcriptShare = (1 - weight) * active
	} else {
		// No commit signal → the full segment goes by mention.
		transcriptShare = active
	}
	if transcriptShare <= 0 {
		return out
	}
	total := 0
	for _, n := range mentions {
		total += n
	}
	if total > 0 {
		for iss, n := range mentions {
			out[iss] += transcriptShare * float64(n) / float64(total)
		}
	} else {
		// No mentions either; leave the transcript share unattributed.
		out[UnattributedKey] += transcriptShare
	}
	return out
}

// buildSegments constructs the v3 segments from time-sorted events + commits and
// returns them with each segment's allocation filled in. Pure — operates on
// already-loaded slices; prefix-weight defaulting is resolved by the caller
// (Compute) so this takes concrete scalar weights.
//
// Mirrors the Python original's main() segment loop: boundaries are the first
// event, every commit time, and one second past the last event, deduped by
// instant; each [start,end) span with ≥1 event (or a clamped task span) becomes
// a segment anchored by the commit whose time equals its end (suffix has none);
// the very first segment uses prefixWeight iff there is a real pre-first-commit
// prefix. Task spans (#118) are NOT added as boundaries — segments keep ending
// at commits so commit-anchored attribution is preserved; each span is clamped
// into the segments it overlaps and counted in full there.
func buildSegments(events []Event, commits []Commit, spans []TaskSpan, commitWeight, prefixWeight float64, thresholdMin int) []Segment {
	// Boundaries, deduped by instant (Python's sorted(set(aware datetimes))
	// dedupes by UTC instant; we key on UTC UnixNano).
	bset := map[int64]time.Time{}
	add := func(t time.Time) { bset[t.UTC().UnixNano()] = t }
	add(events[0].Time)
	for _, c := range commits {
		add(c.Time)
	}
	// Final boundary extends past the last event to cover any span whose return
	// lands after it (a trailing subagent run), so its tail is not cut.
	last := events[len(events)-1].Time.Add(time.Second)
	for _, sp := range spans {
		if sp.End.After(last) {
			last = sp.End
		}
	}
	add(last)
	boundaries := make([]time.Time, 0, len(bset))
	for _, t := range bset {
		boundaries = append(boundaries, t)
	}
	sortTimes(boundaries)

	// A real pre-first-commit prefix exists only if the first event is strictly
	// before the first commit (open-session-then-immediate-commit has none).
	hasPrefix := events[0].Time.Before(commits[0].Time)

	var segs []Segment
	eIdx := 0
	for i := 0; i < len(boundaries)-1; i++ {
		segStart, segEnd := boundaries[i], boundaries[i+1]
		var segEvents []Event
		for eIdx < len(events) && events[eIdx].Time.Before(segEnd) {
			if !events[eIdx].Time.Before(segStart) {
				segEvents = append(segEvents, events[eIdx])
			}
			eIdx++
		}
		// Clamp task spans into this segment. A segment with a span but no events
		// is still emitted (else the post-commit tail of a span — whose return is
		// a dropped tool_result, hence no event — would be silently lost).
		clampedSpans := clampSpans(spans, segStart, segEnd)
		if len(segEvents) == 0 && len(clampedSpans) == 0 {
			continue
		}
		times, mentions := eventTimesAndMentions(segEvents)
		active := activeMinutesUnion(times, clampedSpans, thresholdMin)

		// Anchor: the commit at seg_end, if any (works because every commit
		// time is in the boundary set, so equality holds by instant).
		var anchor *Commit
		for ci := range commits {
			if commits[ci].Time.Equal(segEnd) {
				anchor = &commits[ci]
				break
			}
		}
		var commitIssues []string
		if anchor != nil {
			commitIssues = anchor.Issues
		}
		isPrefix := hasPrefix && i == 0
		weight := commitWeight
		if isPrefix {
			weight = prefixWeight
		}
		segs = append(segs, Segment{
			Start: segStart, End: segEnd, Active: active,
			Commit: anchor, Mentions: mentions,
			Alloc:    attributeSegment(active, commitIssues, mentions, weight),
			IsPrefix: isPrefix,
		})
	}
	return segs
}
