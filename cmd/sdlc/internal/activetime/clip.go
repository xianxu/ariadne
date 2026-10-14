package activetime

import (
	"fmt"
	"sort"
	"time"
)

// ActivityWindow is a half-open interval of confirmed source eligibility.
type ActivityWindow struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	// Began identifies the continuous engagement across relocation slices.
	// Zero preserves callers without engagement metadata, using Start.
	Began time.Time `json:"began,omitempty"`
}

// ClipActivity admits only the source's confirmed ownership intervals. Nil
// windows preserves the standalone/legacy calculation; an explicit empty map
// admits nothing. Overlapping windows union before intersection, so transfers
// and repeated evidence cannot count the same source time twice.
func ClipActivity(runs []ActivityRun, windows map[string][]ActivityWindow) ([]ActivityRun, error) {
	if windows == nil {
		return append([]ActivityRun(nil), runs...), nil
	}
	merged := make(map[string][]ActivityWindow, len(windows))
	for source, intervals := range windows {
		ordered := append([]ActivityWindow(nil), intervals...)
		for _, interval := range ordered {
			if interval.Start.IsZero() || !interval.End.After(interval.Start) {
				return nil, fmt.Errorf("source %s has invalid activity window", source)
			}
		}
		sort.Slice(ordered, func(i, j int) bool { return ordered[i].Start.Before(ordered[j].Start) })
		for _, interval := range ordered {
			last := len(merged[source]) - 1
			if last < 0 || interval.Start.After(merged[source][last].End) {
				merged[source] = append(merged[source], interval)
			} else if interval.End.After(merged[source][last].End) {
				merged[source][last].End = interval.End
			}
		}
	}
	var out []ActivityRun
	for _, run := range runs {
		for _, window := range merged[sourceKey(run.Source)] {
			start, end := run.Start, run.End
			if start.Before(window.Start) {
				start = window.Start
			}
			if end.After(window.End) {
				end = window.End
			}
			if !end.After(start) {
				continue
			}
			part := run
			part.Start, part.End, part.Active = start, end, end.Sub(start).Minutes()
			out = append(out, part)
		}
	}
	return out, nil
}

// AdmitTaskSpans binds a subagent execution to the ownership engagement that
// admitted its original dispatch. It clips to that engagement before span union,
// so neither a release nor a fresh reclaim can revive earlier dispatched work.
func AdmitTaskSpans(spans []TaskSpan, windows map[string][]ActivityWindow) ([]TaskSpan, error) {
	if windows == nil {
		return append([]TaskSpan(nil), spans...), nil
	}
	origin := func(w ActivityWindow) time.Time {
		if w.Began.IsZero() {
			return w.Start
		}
		return w.Began
	}
	for source, ws := range windows {
		for _, w := range ws {
			if w.Start.IsZero() || !w.End.After(w.Start) || origin(w).After(w.Start) {
				return nil, fmt.Errorf("source %s has invalid activity window", source)
			}
		}
	}
	var out []TaskSpan
	for _, sp := range spans {
		dispatch := sp.Dispatched
		if dispatch.IsZero() {
			dispatch = sp.Start
		}
		var began time.Time
		ws := windows[sourceKey(sp.Source)]
		for _, w := range ws {
			if !dispatch.Before(w.Start) && dispatch.Before(w.End) {
				if !began.IsZero() && !began.Equal(origin(w)) {
					return nil, fmt.Errorf("source %s has ambiguous dispatch engagement", sp.Source)
				}
				began = origin(w)
			}
		}
		if began.IsZero() {
			continue
		}
		for _, w := range ws {
			if !origin(w).Equal(began) {
				continue
			}
			part := sp
			if part.Start.Before(w.Start) {
				part.Start = w.Start
			}
			if part.End.After(w.End) {
				part.End = w.End
			}
			if part.End.After(part.Start) {
				out = append(out, part)
			}
		}
	}
	return out, nil
}
