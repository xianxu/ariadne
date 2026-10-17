package activetime

import (
	"fmt"
	"sort"
	"time"
)

// QueueInterval is a recorded machine-suite wait. Source is assigned only to
// admitted activity from the attempt's pinned owner slot.
type QueueInterval struct {
	Start, End              time.Time
	Source, Slot, AttemptID string
}

// QueueAttribution describes minutes intersecting admitted activity runs.
type QueueAttribution struct {
	RawMinutes, ExcludedMinutes float64
}

// bindQueueSources applies a recorded wait to admitted activity sources in its
// pinned slot. No transcript poll or enclosing command supplies queue time.
func bindQueueSources(wait QueueInterval, runs []ActivityRun, slots map[string]string) ([]QueueInterval, error) {
	if wait.AttemptID == "" || wait.Slot == "" || wait.Start.IsZero() || !wait.End.After(wait.Start) {
		return nil, fmt.Errorf("invalid queue interval or identity")
	}
	seen := map[string]bool{}
	var bound []QueueInterval
	for _, run := range runs {
		if !run.Start.Before(wait.End) || !run.End.After(wait.Start) || seen[run.Source] {
			continue
		}
		seen[run.Source] = true
		slot := slots[run.Source]
		if slot == "" {
			return nil, fmt.Errorf("queue attempt %s overlaps source %s without unique slot ownership", wait.AttemptID, run.Source)
		}
		if slot == wait.Slot {
			copy := wait
			copy.Source = run.Source
			bound = append(bound, copy)
		}
	}
	return bound, nil
}

// subtractQueueIdle removes the union of source-bound waits from activity runs.
// Other source runs are unaffected. All intervals are half-open.
func subtractQueueIdle(runs []ActivityRun, waits []QueueInterval) ([]ActivityRun, QueueAttribution) {
	bySource := map[string][]interval{}
	for _, w := range waits {
		bySource[sourceKey(w.Source)] = append(bySource[sourceKey(w.Source)], interval{w.Start, w.End})
	}
	for source, ivs := range bySource {
		bySource[source] = mergeIntervals(ivs)
	}
	var out []ActivityRun
	stats := QueueAttribution{RawMinutes: queueRawMinutes(waits)}
	for _, run := range runs {
		if !run.End.After(run.Start) {
			continue
		}
		pieces := []interval{{run.Start, run.End}}
		for _, wait := range bySource[sourceKey(run.Source)] {
			overlap := intersect(interval{run.Start, run.End}, wait)
			if !overlap.e.After(overlap.s) {
				continue
			}
			stats.ExcludedMinutes += overlap.e.Sub(overlap.s).Minutes()
			pieces = subtractIntervals(pieces, []interval{overlap})
		}
		for _, part := range pieces {
			copy := run
			copy.Start, copy.End, copy.Active = part.s, part.e, part.e.Sub(part.s).Minutes()
			out = append(out, copy)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Start.Equal(out[j].Start) {
			return out[i].Source < out[j].Source
		}
		return out[i].Start.Before(out[j].Start)
	})
	return out, stats
}

// queueRawMinutes reports recorded wall-clock queue time once, irrespective of
// how many admitted sources had activity during the interval.
func queueRawMinutes(waits []QueueInterval) float64 {
	raw := map[string]interval{}
	for _, w := range waits {
		raw[w.AttemptID] = interval{w.Start, w.End}
	}
	var rawIntervals []interval
	for _, iv := range raw {
		rawIntervals = append(rawIntervals, iv)
	}
	var minutes float64
	for _, iv := range mergeIntervals(rawIntervals) {
		minutes += iv.e.Sub(iv.s).Minutes()
	}
	return minutes
}

func intersect(a, b interval) interval {
	if a.s.Before(b.s) {
		a.s = b.s
	}
	if a.e.After(b.e) {
		a.e = b.e
	}
	return a
}

func mergeIntervals(ivals []interval) []interval {
	sort.Slice(ivals, func(i, j int) bool { return ivals[i].s.Before(ivals[j].s) })
	var out []interval
	for _, iv := range ivals {
		if !iv.e.After(iv.s) {
			continue
		}
		if len(out) == 0 || iv.s.After(out[len(out)-1].e) {
			out = append(out, iv)
		} else if iv.e.After(out[len(out)-1].e) {
			out[len(out)-1].e = iv.e
		}
	}
	return out
}

func subtractIntervals(parts, cuts []interval) []interval {
	for _, cut := range cuts {
		var next []interval
		for _, p := range parts {
			x := intersect(p, cut)
			if !x.e.After(x.s) {
				next = append(next, p)
				continue
			}
			if x.s.After(p.s) {
				next = append(next, interval{p.s, x.s})
			}
			if p.e.After(x.e) {
				next = append(next, interval{x.e, p.e})
			}
		}
		parts = next
	}
	return parts
}
