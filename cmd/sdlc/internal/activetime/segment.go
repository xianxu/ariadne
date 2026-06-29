package activetime

import (
	"sort"
	"time"
)

// Segment is the rendered unit of attribution: one source-scoped activity run,
// the commit boundary that claimed it (if any), and the resulting per-issue
// allocation.
type Segment struct {
	Start, End time.Time
	Active     float64
	Commit     *Commit
	Mentions   map[string]int
	Alloc      map[string]float64
	IsPrefix   bool
}

// ActivityRun is one claimable unit of transcript work. Source is retained so
// overlapping sessions can each count toward their claimed issue while overlaps
// inside one source still collapse to wall-clock.
type ActivityRun struct {
	Start, End time.Time
	Active     float64
	Mentions   map[string]int
	Source     string
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

type runInterval struct {
	interval
	mentions map[string]int
}

func sourceKey(s string) string {
	if s == "" {
		return "(default)"
	}
	return s
}

func addMentions(dst, src map[string]int) {
	for iss, n := range src {
		dst[iss] += n
	}
}

// activityRuns turns event gaps and task spans into source-scoped active runs.
// It unions overlaps within a source only; overlaps across sources remain
// separate claimable work.
func activityRuns(events []Event, spans []TaskSpan, thresholdMin int) []ActivityRun {
	eventsBySource := map[string][]Event{}
	for _, e := range events {
		eventsBySource[sourceKey(e.Source)] = append(eventsBySource[sourceKey(e.Source)], e)
	}
	ivalsBySource := map[string][]runInterval{}
	capGap := time.Duration(thresholdMin) * time.Minute
	for source, evs := range eventsBySource {
		sort.Slice(evs, func(i, j int) bool { return evs[i].Time.Before(evs[j].Time) })
		for i := 1; i < len(evs); i++ {
			start := evs[i-1].Time
			end := evs[i].Time
			if end.Sub(start) > capGap {
				end = start.Add(capGap)
			}
			if !end.After(start) {
				continue
			}
			mentions := map[string]int{}
			addMentions(mentions, evs[i-1].Mentions)
			addMentions(mentions, evs[i].Mentions)
			ivalsBySource[source] = append(ivalsBySource[source], runInterval{
				interval: interval{s: start, e: end},
				mentions: mentions,
			})
		}
	}
	for _, sp := range spans {
		if sp.End.After(sp.Start) {
			source := sourceKey(sp.Source)
			ivalsBySource[source] = append(ivalsBySource[source], runInterval{
				interval: interval{s: sp.Start, e: sp.End},
				mentions: map[string]int{},
			})
		}
	}

	var runs []ActivityRun
	for source, ivals := range ivalsBySource {
		if len(ivals) == 0 {
			continue
		}
		sort.Slice(ivals, func(i, j int) bool { return ivals[i].s.Before(ivals[j].s) })
		cur := ivals[0]
		for _, iv := range ivals[1:] {
			if iv.s.After(cur.e) {
				runs = append(runs, ActivityRun{
					Start: cur.s, End: cur.e, Active: cur.e.Sub(cur.s).Minutes(),
					Mentions: cur.mentions, Source: source,
				})
				cur = iv
				continue
			}
			if iv.e.After(cur.e) {
				cur.e = iv.e
			}
			addMentions(cur.mentions, iv.mentions)
		}
		runs = append(runs, ActivityRun{
			Start: cur.s, End: cur.e, Active: cur.e.Sub(cur.s).Minutes(),
			Mentions: cur.mentions, Source: source,
		})
	}
	sort.Slice(runs, func(i, j int) bool {
		if runs[i].Start.Equal(runs[j].Start) {
			return runs[i].Source < runs[j].Source
		}
		return runs[i].Start.Before(runs[j].Start)
	})
	return runs
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

func claimActivityRuns(runs []ActivityRun, commits []Commit, commitWeight, prefixWeight float64) []Segment {
	var firstIssueCommit *Commit
	for i := range commits {
		if len(commits[i].Issues) > 0 {
			firstIssueCommit = &commits[i]
			break
		}
	}
	segs := make([]Segment, 0, len(runs))
	for _, r := range runs {
		anchor := selectClaimant(r, commits)
		isPrefix := firstIssueCommit != nil && (r.End.Before(firstIssueCommit.Time) || r.End.Equal(firstIssueCommit.Time))
		weight := commitWeight
		if isPrefix {
			weight = prefixWeight
		}
		segs = append(segs, Segment{
			Start: r.Start, End: r.End, Active: r.Active,
			Commit: anchor, Mentions: r.Mentions,
			Alloc:    attributeRun(r.Active, issuesOf(anchor), r.Mentions, weight),
			IsPrefix: isPrefix,
		})
	}
	return segs
}

func issuesOf(c *Commit) []string {
	if c == nil {
		return nil
	}
	return c.Issues
}

func selectClaimant(run ActivityRun, commits []Commit) *Commit {
	var prevNeutral, nextNeutral *Commit
	for i := range commits {
		c := &commits[i]
		if len(c.Issues) > 0 {
			continue
		}
		switch {
		case c.Time.After(run.Start) && c.Time.Before(run.End):
			return nil
		case c.Time.Before(run.Start) || c.Time.Equal(run.Start):
			if prevNeutral == nil || c.Time.After(prevNeutral.Time) {
				prevNeutral = c
			}
		case c.Time.After(run.End) || c.Time.Equal(run.End):
			if nextNeutral == nil || c.Time.Before(nextNeutral.Time) {
				nextNeutral = c
			}
		}
	}

	var prev, next, inside *Commit
	for i := range commits {
		c := &commits[i]
		if len(c.Issues) == 0 {
			continue
		}
		switch {
		case c.Time.After(run.Start) && c.Time.Before(run.End):
			if inside == nil || c.Time.After(inside.Time) {
				inside = c
			}
		case c.Time.Before(run.Start) || c.Time.Equal(run.Start):
			if prevNeutral != nil && !c.Time.After(prevNeutral.Time) {
				continue
			}
			if prev == nil || c.Time.After(prev.Time) {
				prev = c
			}
		case c.Time.After(run.End) || c.Time.Equal(run.End):
			if nextNeutral != nil && !c.Time.Before(nextNeutral.Time) {
				continue
			}
			if next == nil || c.Time.Before(next.Time) {
				next = c
			}
		}
	}
	if inside != nil && next == nil {
		return inside
	}
	switch {
	case prev == nil:
		return next
	case next == nil:
		return prev
	default:
		prevDist := run.Start.Sub(prev.Time)
		nextDist := next.Time.Sub(run.End)
		if nextDist <= prevDist {
			return next
		}
		return prev
	}
}

func attributeRun(active float64, commitIssues []string, mentions map[string]int, weight float64) map[string]float64 {
	if len(commitIssues) == 0 {
		return attributeSegment(active, nil, mentions, weight)
	}
	out := map[string]float64{}
	if active <= 0 {
		return out
	}
	commitShare := weight * active
	perCommit := commitShare / float64(len(commitIssues))
	for _, iss := range commitIssues {
		out[iss] += perCommit
	}
	if remainder := active - commitShare; remainder > 0 {
		out[UnattributedKey] += remainder
	}
	return out
}

// buildSegments constructs rendered attribution rows from source-scoped activity
// runs claimed by nearby issue commit boundaries. Pure — operates on already-
// loaded slices; prefix-weight defaulting is resolved by the caller (Compute) so
// this takes concrete scalar weights.
func buildSegments(events []Event, commits []Commit, spans []TaskSpan, commitWeight, prefixWeight float64, thresholdMin int) []Segment {
	if len(events) == 0 && len(spans) == 0 {
		return nil
	}
	return claimActivityRuns(activityRuns(events, spans, thresholdMin), commits, commitWeight, prefixWeight)
}
