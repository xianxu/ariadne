package activetime

import (
	"fmt"
	"sort"
	"time"
)

// CreditedInterval is time already credited to Issue in Slot. It is half-open.
type CreditedInterval struct {
	Issue, Slot string
	Start, End  time.Time
}

// Exclusion records the portion of one credited interval actually removed
// from an activity source. Overlapping credited intervals never remove or
// report the same source minute twice.
type Exclusion struct {
	Issue, Slot, Source string
	Start, End          time.Time
	Minutes             float64
}

type ownedCut struct {
	interval
	issue string
}

// subtractProtectedIntervals removes the union of credited time from each run
// whose source belongs to the credited slot. Every overlapping source must have
// known slot ownership, including sources that would otherwise be unaffected.
func subtractProtectedIntervals(runs []ActivityRun, protected []CreditedInterval, slots map[string]string) ([]ActivityRun, []Exclusion, error) {
	ordered := append([]CreditedInterval(nil), protected...)
	for _, credit := range ordered {
		if credit.Issue == "" || credit.Slot == "" || credit.Start.IsZero() || !credit.End.After(credit.Start) {
			return nil, nil, fmt.Errorf("invalid credited interval or identity")
		}
	}
	sort.Slice(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.Slot != b.Slot {
			return a.Slot < b.Slot
		}
		if !a.Start.Equal(b.Start) {
			return a.Start.Before(b.Start)
		}
		if !a.End.Equal(b.End) {
			return a.End.Before(b.End)
		}
		return a.Issue < b.Issue
	})
	// Assign overlapping credit to the first interval in stable time order.
	// The union, not the number of records, determines excluded minutes.
	bySlot := map[string][]ownedCut{}
	covered := map[string][]interval{}
	for _, credit := range ordered {
		piece := []interval{{credit.Start, credit.End}}
		piece = subtractIntervals(piece, covered[credit.Slot])
		for _, p := range piece {
			bySlot[credit.Slot] = append(bySlot[credit.Slot], ownedCut{interval: p, issue: credit.Issue})
		}
		covered[credit.Slot] = mergeIntervals(append(covered[credit.Slot], interval{credit.Start, credit.End}))
	}
	var out []ActivityRun
	var excluded []Exclusion
	for _, run := range runs {
		if !run.End.After(run.Start) {
			continue
		}
		for _, credit := range ordered {
			if run.Start.Before(credit.End) && run.End.After(credit.Start) && slots[run.Source] == "" {
				return nil, nil, fmt.Errorf("credited issue %s overlaps source %s without unique slot ownership", credit.Issue, run.Source)
			}
		}
		pieces := []interval{{run.Start, run.End}}
		for _, cut := range bySlot[slots[run.Source]] {
			overlap := intersect(interval{run.Start, run.End}, cut.interval)
			if !overlap.e.After(overlap.s) {
				continue
			}
			excluded = append(excluded, Exclusion{
				Issue: cut.issue, Slot: slots[run.Source], Source: run.Source,
				Start: overlap.s, End: overlap.e, Minutes: overlap.e.Sub(overlap.s).Minutes(),
			})
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
	return out, excluded, nil
}
