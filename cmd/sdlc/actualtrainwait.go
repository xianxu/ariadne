package main

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

// selectCarWaits selects descriptive overlap, never an activity subtraction.
// The full retained invocation is covered after verified car membership, even
// when a later split omits this car or it is released back into the queue.
func selectCarWaits(waits []landingline.WaitInterval, heads map[string]queueHeadMembership, engagements []tracker.EngagementWindow, repository string, issueID int, since, until time.Time, sessions []transcripts.SessionEvidence, windows map[string][]activetime.ActivityWindow, slots map[string]string) ([]gatestate.MeasurementCarWait, []string) {
	var selected []gatestate.MeasurementCarWait
	var notes []string
	// Compute ambiguous ownership windows once, rather than rescanning the
	// engagement history for every selected window of every waiter.
	ambiguousWindows := carWaitAmbiguousEngagements(engagements)
	for _, w := range waits {
		if w.Issue != issueID || w.Identity.Repo != repository || !w.Start.Before(until) || !w.End.After(since) {
			continue
		}
		head, ok := heads[w.Identity.Head]
		if !ok || !head.Known {
			notes = append(notes, fmt.Sprintf("ticket %d head %s membership unknown", w.Seq, w.Identity.Head))
			continue
		}
		if !slices.Contains(head.Issues, issueID) {
			continue
		}
		before := len(selected)
		ownedOverlap := false
		for index, e := range engagements {
			ownerCount := 0
			for _, owner := range e.Owners {
				if owner.Workspace == w.Identity.Slot && filepath.Clean(owner.Worktree) == w.Identity.Root {
					ownerCount++
				}
			}
			if ownerCount == 0 {
				continue
			}
			start, end := carWaitIntersection(w.Start, w.End, since, until)
			start, end = carWaitIntersection(start, end, e.Start, e.End)
			if !end.After(start) {
				continue
			}
			ownedOverlap = true
			ambiguous := ownerCount != 1 || ambiguousWindows[index]
			if ambiguous {
				notes = append(notes, fmt.Sprintf("ticket %d ownership overlap is ambiguous", w.Seq))
				continue
			}
			for _, session := range sessions {
				if slots[session.File] != w.Identity.Slot || filepath.Clean(session.CWD) != w.Identity.Root {
					continue
				}
				for _, window := range windows[session.File] {
					s, t := carWaitIntersection(start, end, window.Start, window.End)
					if !t.After(s) {
						continue
					}
					q := gatestate.MeasurementCarWait{Ticket: w.Seq, Trains: append([]int64(nil), w.Trains...), Slot: w.Identity.Slot, Head: w.Identity.Head, RecordDigest: w.RecordDigest, Start: s.UTC().Format(time.RFC3339Nano), End: t.UTC().Format(time.RFC3339Nano), Sources: []string{session.Harness + "/" + session.ID}}
					found := false
					for i := range selected {
						old := &selected[i]
						if old.Ticket == q.Ticket && old.Start == q.Start && old.End == q.End {
							if !slices.Contains(old.Sources, q.Sources[0]) {
								old.Sources = append(old.Sources, q.Sources[0])
								slices.Sort(old.Sources)
							}
							found = true
							break
						}
					}
					if !found {
						selected = append(selected, q)
					}
					if len(selected) > gatestate.SnapshotLimitCarWaits {
						return nil, []string{"selected retained train waits exceed snapshot capacity"}
					}
				}
			}
		}
		if ownedOverlap && len(selected) == before {
			notes = append(notes, fmt.Sprintf("ticket %d has no unambiguous pinned source overlap", w.Seq))
		}
	}
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Start < selected[j].Start || selected[i].Start == selected[j].Start && selected[i].Ticket < selected[j].Ticket
	})
	return selected, boundedCarWaitNotes(notes)
}

func carWaitAmbiguousEngagements(windows []tracker.EngagementWindow) map[int]bool {
	order := make([]int, len(windows))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool { return windows[order[i]].Start.Before(windows[order[j]].Start) })
	out := map[int]bool{}
	maxIndex := -1
	var maxEnd time.Time
	for _, index := range order {
		w := windows[index]
		if maxIndex >= 0 && w.Start.Before(maxEnd) {
			out[index] = true
			out[maxIndex] = true
		}
		if maxIndex < 0 || w.End.After(maxEnd) {
			maxEnd = w.End
			maxIndex = index
		}
	}
	return out
}

func carWaitIntersection(start, end, lower, upper time.Time) (time.Time, time.Time) {
	if start.Before(lower) {
		start = lower
	}
	if end.After(upper) {
		end = upper
	}
	return start, end
}

// attachCarWaitCoverage validates a copy, so diagnostic capacity or malformed
// optional facts cannot invalidate otherwise valid accounting evidence.
func attachCarWaitCoverage(m gatestate.MeasurementInputs, waits []gatestate.MeasurementCarWait, notes []string, sources []gatestate.MeasurementSource) (gatestate.MeasurementInputs, []string) {
	if len(waits) == 0 && len(notes) == 0 {
		return m, nil
	}
	candidate := m
	candidate.CarWaits = waits
	candidate.CarWaitNotes = boundedCarWaitNotes(notes)
	candidate.Sources = append([]gatestate.MeasurementSource(nil), m.Sources...)
	needed := map[string]bool{}
	for _, w := range waits {
		for _, source := range w.Sources {
			needed[source] = true
		}
	}
	for _, source := range candidate.Sources {
		delete(needed, source.Harness+"/"+source.ID)
	}
	for _, source := range sources {
		if needed[source.Harness+"/"+source.ID] {
			candidate.Sources = append(candidate.Sources, source)
			delete(needed, source.Harness+"/"+source.ID)
		}
	}
	err := candidate.ValidateCarWaitCoverage()
	if err == nil {
		err = candidate.Validate()
	}
	if err == nil {
		return candidate, nil
	}
	note := "retained train wait evidence incomplete: optional snapshot capacity or validation limit"
	// Keep a compact replay limitation only when it fits; otherwise preserve
	// the original bytes and report the limitation in this invocation alone.
	fallback := m
	fallback.CarWaitNotes = []string{note}
	if fallback.Validate() == nil {
		return fallback, nil
	}
	return m, []string{note}
}

func boundedCarWaitNotes(notes []string) []string {
	var out []string
	for _, note := range notes {
		if len(note) > 256 {
			note = note[:256]
		}
		if !slices.Contains(out, note) {
			out = append(out, note)
		}
		if len(out) == 16 {
			break
		}
	}
	return out
}

func carWaitCoverageWarnings(waits []gatestate.MeasurementCarWait, notes []string) []string {
	type interval struct{ start, end time.Time }
	var intervals []interval
	for _, w := range waits {
		s, e1 := time.Parse(time.RFC3339Nano, w.Start)
		e, e2 := time.Parse(time.RFC3339Nano, w.End)
		if e1 == nil && e2 == nil && e.After(s) {
			intervals = append(intervals, interval{s, e})
		}
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i].start.Before(intervals[j].start) })
	var total time.Duration
	if len(intervals) > 0 {
		cur := intervals[0]
		for _, next := range intervals[1:] {
			if next.start.After(cur.end) {
				total += cur.end.Sub(cur.start)
				cur = next
			} else if next.end.After(cur.end) {
				cur.end = next.end
			}
		}
		total += cur.end.Sub(cur.start)
	}
	var out []string
	if total > 0 {
		out = append(out, fmt.Sprintf("unverified coverage: %.2f min in a retained train wait (whole bound invocation; no subtraction or idle proof)", total.Minutes()))
	}
	for _, note := range boundedCarWaitNotes(notes) {
		out = append(out, "unverified coverage: "+strings.TrimPrefix(note, "unverified coverage: "))
	}
	return out
}
