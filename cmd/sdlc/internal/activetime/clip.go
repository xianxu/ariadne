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
