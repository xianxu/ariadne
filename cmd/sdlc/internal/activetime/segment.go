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

// unattributedKey buckets transcript share that has no mention signal. Stored
// with a leading '#' as a sentinel so it never collides with a numeric issue
// key; rendered as "unattributed" (the Python's f"#{iss}" doubled it to
// "##unattributed" — a cosmetic bug we don't reproduce).
const unattributedKey = "#unattributed"

// sortTimes sorts a slice of times ascending, in place.
func sortTimes(ts []time.Time) {
	sort.Slice(ts, func(i, j int) bool { return ts[i].Before(ts[j]) })
}

// activeMinutes sums inter-event gaps, each capped at thresholdMin, and returns
// minutes. Mirrors active-time-v3.py active_minutes — the v2.1 active-time
// procedure. The caller may pass an unsorted slice; we sort a copy defensively
// (Python sorts internally and does not mutate its input).
func activeMinutes(times []time.Time, thresholdMin int) float64 {
	if len(times) < 2 {
		return 0
	}
	sorted := make([]time.Time, len(times))
	copy(sorted, times)
	sortTimes(sorted)
	capGap := time.Duration(thresholdMin) * time.Minute
	var total time.Duration
	for i := 1; i < len(sorted); i++ {
		gap := sorted[i].Sub(sorted[i-1])
		if gap <= capGap {
			total += gap
		} else {
			total += capGap
		}
	}
	return total.Minutes()
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
		out[unattributedKey] += transcriptShare
	}
	return out
}

// buildSegments constructs the v3 segments from time-sorted events + commits and
// returns them with each segment's allocation filled in. Pure — operates on
// already-loaded slices; prefix-weight defaulting is resolved by the caller
// (Compute) so this takes concrete scalar weights.
//
// Mirrors active-time-v3.py main() lines 324–372: boundaries are the first
// event, every commit time, and one second past the last event, deduped by
// instant; each [start,end) span with ≥1 event becomes a segment anchored by the
// commit whose time equals its end (suffix has none); the very first segment
// uses prefixWeight iff there is a real pre-first-commit prefix.
func buildSegments(events []Event, commits []Commit, commitWeight, prefixWeight float64, thresholdMin int) []Segment {
	// Boundaries, deduped by instant (Python's sorted(set(aware datetimes))
	// dedupes by UTC instant; we key on UTC UnixNano).
	bset := map[int64]time.Time{}
	add := func(t time.Time) { bset[t.UTC().UnixNano()] = t }
	add(events[0].Time)
	for _, c := range commits {
		add(c.Time)
	}
	add(events[len(events)-1].Time.Add(time.Second))
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
		if len(segEvents) == 0 {
			continue
		}
		times := make([]time.Time, len(segEvents))
		mentions := map[string]int{}
		for k, e := range segEvents {
			times[k] = e.Time
			for iss, n := range e.Mentions {
				mentions[iss] += n
			}
		}
		active := activeMinutes(times, thresholdMin)

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
