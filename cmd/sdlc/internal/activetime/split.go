package activetime

import (
	"sort"
	"time"
)

// PhaseKind labels a stretch of a measured window (#234).
type PhaseKind int

const (
	// Build is design and implementation up to "done as far as I can tell":
	// what an estimate predicts.
	Build PhaseKind = iota
	// Converge is a boundary's review-and-fix loop: from its first review
	// dispatch to the last #N commit at or before its final dispatch.
	Converge
	// Tail is a finalized boundary's verification after its last commit (the
	// final test run and review). It is uncounted at every boundary, the
	// operator's 2026-10-10 consistency rule.
	Tail
)

// Split sources: which evidence supplied a timeline's cuts.
const (
	SplitDispatch = "dispatch"  // every cut is a stamped dispatch
	SplitFinalize = "finalize"  // some cut fell back to a round's finalize time
	SplitNoReview = "no-review" // no boundary rounds: everything is build
)

// BoundaryRound is the view of one boundary-review round the split needs. The
// caller maps it from the gate ledger (activetime does not import gatestate):
// Dispatched is zero when the round has no usable stamp, and Finalizing marks
// a round that finalized its boundary (gatestate.Round.Reviewed != "").
type BoundaryRound struct {
	Boundary   string
	Dispatched time.Time
	Finalized  time.Time
	Finalizing bool
}

// Phase is one labelled stretch of the window. Boundary names the boundary it
// belongs to: a build phase belongs to the boundary it builds toward.
type Phase struct {
	Start, End time.Time
	Kind       PhaseKind
	Boundary   string
}

// Timeline is a window labelled into ordered, contiguous, non-overlapping
// phases, and the evidence its cuts came from.
type Timeline struct {
	Phases []Phase
	Source string
}

// For keeps the phases of one boundary: a milestone's own increment (E1).
func (tl Timeline) For(boundary string) Timeline {
	out := Timeline{Source: tl.Source}
	for _, p := range tl.Phases {
		if p.Boundary == boundary {
			out.Phases = append(out.Phases, p)
		}
	}
	return out
}

// BuildTimeline labels [since, until] from the boundary rounds and the times
// of the issue's own #N commits (#234 D1). For each boundary, in order of its
// first round:
//
//   - its cut is its first round's dispatch (the finalize time when the stamp
//     is missing or later than the finalize);
//   - once finalized, build runs to min(cut, tailStart), where tailStart is the
//     last commit at or before its final round's dispatch; converge is
//     [cut, tailStart] when tailStart is after the cut; the rest up to its
//     finalize is tail;
//   - still open (reviewed, no finalizing round yet), it is converge from its
//     cut to the window end.
//
// Time after the last finalized boundary is build toward current, the
// boundary being measured now, unless current has itself finalized (a past
// milestone measured later): then that time belongs to a boundary not yet
// reviewed, and is tagged "".
func BuildTimeline(rounds []BoundaryRound, issueCommits []time.Time, current string, since, until time.Time) Timeline {
	commits := append([]time.Time(nil), issueCommits...)
	sort.Slice(commits, func(i, j int) bool { return commits[i].Before(commits[j]) })
	tl := Timeline{Source: SplitNoReview}
	last := since
	emit := func(start, end time.Time, kind PhaseKind, boundary string) {
		if start.Before(last) {
			start = last
		}
		if end.After(until) {
			end = until
		}
		if !end.After(start) {
			return
		}
		tl.Phases = append(tl.Phases, Phase{Start: start, End: end, Kind: kind, Boundary: boundary})
		last = end
	}
	cursor := since
	finalized := map[string]bool{}
	for _, b := range groupBoundaries(rounds) {
		first := b[0]
		cut, stamped := dispatchOf(first)
		if tl.Source != SplitFinalize {
			tl.Source = SplitDispatch
			if !stamped {
				tl.Source = SplitFinalize
			}
		}
		final, ok := lastFinalizing(b)
		if !ok {
			emit(cursor, cut, Build, first.Boundary)
			emit(cut, until, Converge, first.Boundary)
			cursor = until
			continue
		}
		finalDispatch, finalStamped := dispatchOf(final)
		if !finalStamped {
			tl.Source = SplitFinalize
		}
		tailStart := lastAtOrBefore(commits, finalDispatch, cursor)
		if tailStart.After(cut) {
			emit(cursor, cut, Build, first.Boundary)
			emit(cut, tailStart, Converge, first.Boundary)
		} else {
			emit(cursor, tailStart, Build, first.Boundary)
		}
		emit(tailStart, final.Finalized, Tail, first.Boundary)
		finalized[first.Boundary] = true
		if final.Finalized.After(cursor) {
			cursor = final.Finalized
		}
	}
	if finalized[current] {
		current = ""
	}
	emit(cursor, until, Build, current)
	return tl
}

// groupBoundaries groups rounds by boundary in order of first appearance.
// Rounds that belong to every boundary carry no cut of their own.
func groupBoundaries(rounds []BoundaryRound) [][]BoundaryRound {
	idx := map[string]int{}
	var out [][]BoundaryRound
	for _, r := range rounds {
		if r.Boundary == "*" {
			continue
		}
		i, ok := idx[r.Boundary]
		if !ok {
			i = len(out)
			idx[r.Boundary] = i
			out = append(out, nil)
		}
		out[i] = append(out[i], r)
	}
	return out
}

// dispatchOf is a round's dispatch time, falling back to its finalize time
// when the stamp is missing or later than the finalize (a hand-edited or
// malformed value). stamped reports whether the stamp was used.
func dispatchOf(r BoundaryRound) (time.Time, bool) {
	if r.Dispatched.IsZero() || r.Dispatched.After(r.Finalized) {
		return r.Finalized, false
	}
	return r.Dispatched, true
}

// lastFinalizing is the boundary's last finalizing round: a narrow re-close
// (#183) finalizes again and extends converge.
func lastFinalizing(rounds []BoundaryRound) (BoundaryRound, bool) {
	for i := len(rounds) - 1; i >= 0; i-- {
		if rounds[i].Finalizing {
			return rounds[i], true
		}
	}
	return BoundaryRound{}, false
}

// lastAtOrBefore is the latest commit at or before t, or floor when there is
// none after floor.
func lastAtOrBefore(commits []time.Time, t, floor time.Time) time.Time {
	out := floor
	for _, c := range commits {
		if c.After(t) {
			break
		}
		if c.After(out) {
			out = c
		}
	}
	return out
}

// SplitResult is an issue's measured minutes by phase kind. Build + Converge
// is the new-rule actual (#234 D2); Tail is what the rule leaves uncounted.
type SplitResult struct {
	Build, Converge, Tail float64
}

// Split distributes each segment's allocation to issue over the timeline's
// phases in proportion to time. A run is one contiguous interval
// (Active == End−Start), so the cut is exact in minutes; the claimant stays
// the one Compute chose for the whole run (#234 D3). Minutes outside every
// phase (outside the window, or another boundary's for a For timeline) are
// not counted.
func Split(segs []Segment, issue string, tl Timeline) SplitResult {
	var out SplitResult
	add := func(k PhaseKind, m float64) {
		switch k {
		case Build:
			out.Build += m
		case Converge:
			out.Converge += m
		case Tail:
			out.Tail += m
		}
	}
	for _, s := range segs {
		a := s.Alloc[issue]
		if a <= 0 {
			continue
		}
		dur := s.End.Sub(s.Start)
		if dur <= 0 {
			for i, p := range tl.Phases {
				if !s.Start.Before(p.Start) && (s.Start.Before(p.End) || i == len(tl.Phases)-1 && s.Start.Equal(p.End)) {
					add(p.Kind, a)
					break
				}
			}
			continue
		}
		for _, p := range tl.Phases {
			lo, hi := p.Start, p.End
			if s.Start.After(lo) {
				lo = s.Start
			}
			if s.End.Before(hi) {
				hi = s.End
			}
			if hi.After(lo) {
				add(p.Kind, a*float64(hi.Sub(lo))/float64(dur))
			}
		}
	}
	return out
}
