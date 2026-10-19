package main

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// creditedIntervals protects only Build/Converge time included in the adopted
// number. A fractionally allocated run protects its whole intersection with a
// counted phase: the engine has no finer evidence for that fraction.
func creditedIntervals(issue string, segments []activetime.Segment, counted []activetime.Phase, slots map[string]string) ([]activetime.CreditedInterval, error) {
	var out []activetime.CreditedInterval
	for _, segment := range segments {
		if segment.Alloc[issue] <= 0 {
			continue
		}
		for _, phase := range counted {
			if phase.Kind == activetime.Tail {
				continue
			}
			start, end := segment.Start, segment.End
			if start.Before(phase.Start) {
				start = phase.Start
			}
			if end.After(phase.End) {
				end = phase.End
			}
			if !start.Before(end) {
				continue
			}
			slot := slots[segment.Source]
			if slot == "" {
				return nil, fmt.Errorf("credited source %q for issue #%s during %s → %s has no pinned slot", segment.Source, issue, start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))
			}
			out = append(out, activetime.CreditedInterval{Issue: issue, Slot: slot, Start: start, End: end})
		}
	}
	return out, nil
}

type actualSlotWindow struct {
	slot       string
	start, end time.Time
}

func candidateSlotWindows(inputs *ownedActualInputs) ([]actualSlotWindow, error) {
	var out []actualSlotWindow
	for source, windows := range inputs.Windows {
		slot := inputs.SourceSlots[source]
		for _, window := range windows {
			if slot == "" {
				return nil, fmt.Errorf("candidate source %q during %s → %s has no pinned slot", source, window.Start.Format(time.RFC3339Nano), window.End.Format(time.RFC3339Nano))
			}
			out = append(out, actualSlotWindow{slot, window.Start, window.End})
		}
	}
	return out, nil
}

func overlappingSlotInterval(slot string, start, end time.Time, windows []actualSlotWindow) (time.Time, time.Time, bool) {
	for _, window := range windows {
		if window.slot == slot && start.Before(window.end) && window.start.Before(end) {
			if start.Before(window.start) {
				start = window.start
			}
			if end.After(window.end) {
				end = window.end
			}
			return start, end, true
		}
	}
	return time.Time{}, time.Time{}, false
}

type adoptedWindowState uint8

const (
	windowDisjoint adoptedWindowState = iota
	windowOverlap
	windowMissing
)

type adoptedSlotState uint8

const (
	slotDifferent adoptedSlotState = iota
	slotSame
	slotUnknown
)

type adoptedDecision uint8

const (
	adoptionSkip adoptedDecision = iota
	adoptionReplay
	adoptionRefuse
)

// classifyAdopted is the complete window × slot decision table. Invalid
// states fail closed. In particular, unknown is never treated as not-same.
func classifyAdopted(window adoptedWindowState, slot adoptedSlotState) adoptedDecision {
	switch window {
	case windowDisjoint:
		return adoptionSkip
	case windowOverlap:
		switch slot {
		case slotDifferent:
			return adoptionSkip
		case slotSame:
			return adoptionReplay
		case slotUnknown:
			return adoptionRefuse
		}
	case windowMissing:
		switch slot {
		case slotDifferent:
			return adoptionSkip
		case slotSame, slotUnknown:
			return adoptionRefuse
		}
	}
	return adoptionRefuse
}

// adoptedSlotRelationship requires complete source and ownership identity.
// A different slot is proven only when every owner has a known workspace and
// none matches a candidate source slot. Mixed evidence stays unknown.
func adoptedSlotRelationship(engagements []tracker.EngagementWindow, windows []actualSlotWindow) (adoptedSlotState, time.Time, time.Time) {
	if len(engagements) == 0 || len(windows) == 0 {
		return slotUnknown, time.Time{}, time.Time{}
	}
	for _, window := range windows {
		if window.slot == "" {
			return slotUnknown, time.Time{}, time.Time{}
		}
	}
	var same bool
	var sameStart, sameEnd time.Time
	for _, engagement := range engagements {
		if len(engagement.Owners) == 0 {
			return slotUnknown, time.Time{}, time.Time{}
		}
		for _, owner := range engagement.Owners {
			if owner.Workspace == "" {
				return slotUnknown, time.Time{}, time.Time{}
			}
			for _, window := range windows {
				if owner.Workspace != window.slot {
					continue
				}
				same = true
				if engagement.Start.Before(window.end) && window.start.Before(engagement.End) {
					sameStart, sameEnd = engagement.Start, engagement.End
					if sameStart.Before(window.start) {
						sameStart = window.start
					}
					if sameEnd.After(window.end) {
						sameEnd = window.end
					}
				}
			}
		}
	}
	if same {
		return slotSame, sameStart, sameEnd
	}
	return slotDifferent, time.Time{}, time.Time{}
}

// cardDateMayOverlap is only a cheap rejection filter. Card dates are local
// calendar dates with no zone, whereas evidence uses UTC; widen both ends by
// one day so a timezone edge cannot silently hide an adopted issue. Missing or
// malformed dates stay relevant. Neither a date nor an open claim is exclusion
// authority.
func cardDateMayOverlap(created, updated string, since, until time.Time) bool {
	first, firstErr := time.Parse("2006-01-02", created)
	last, lastErr := time.Parse("2006-01-02", updated)
	if firstErr == nil && !first.Add(-24*time.Hour).Before(until) {
		return false
	}
	if lastErr == nil && !last.Add(48*time.Hour).After(since) {
		return false
	}
	return true
}

// savedMeasurementMayOverlap compares credited-time bounds, not the lifetime of
// a claim that may remain on a done card after its actual was adopted.
func savedMeasurementMayOverlap(savedSince, savedUntil string, since, until time.Time) bool {
	start, _ := time.Parse(time.RFC3339Nano, savedSince) // validated snapshot
	end, _ := time.Parse(time.RFC3339Nano, savedUntil)
	return start.Before(until) && since.Before(end)
}

// loadAdoptedExclusions reads only already adopted same-slot actuals. A
// potential overlap whose pinned replay cannot be established is unknown, not
// zero. The measured issue's own completion never excludes itself.
func loadAdoptedExclusions(ctx context.Context, root, issueNum, plans string, current *ownedActualInputs) ([]activetime.CreditedInterval, error) {
	since, err := time.Parse(time.RFC3339Nano, current.Snapshot.Since)
	if err != nil {
		return nil, err
	}
	until, err := time.Parse(time.RFC3339Nano, current.Snapshot.Until)
	if err != nil {
		return nil, err
	}
	repo, err := recordsRepository(ctx, root)
	if err != nil {
		return nil, err
	}
	rs, err := actualInputRecords(ctx, repo, root, filepath.Join(root, envOr("WF_ISSUES_DIR", "workshop/issues")))
	if err != nil {
		return nil, err
	}
	if rs.Stale {
		return nil, fmt.Errorf("adopted-actual tracker evidence is stale: %v", rs.FetchErr)
	}
	var candidates []tracker.IssueRecord
	for _, rec := range rs.All() {
		id := strings.TrimLeft(rec.ID, "0")
		if id == strings.TrimLeft(issueNum, "0") {
			continue
		}
		if !cardDateMayOverlap(rec.Text("created"), rec.Text("updated"), since, until) {
			continue
		}
		if rec.CardErr != nil {
			return nil, fmt.Errorf("potentially overlapping issue #%s card unreadable: %w", id, rec.CardErr)
		}
		if rec.Card == nil {
			continue
		}
		if _, err := strconv.ParseFloat(rec.Text("actual_hours"), 64); err != nil {
			continue // N/A and unadopted issues reserve no time
		}
		candidates = append(candidates, rec)
	}
	if len(candidates) == 0 {
		return nil, nil // no adopted actual can overlap; source slots are irrelevant
	}
	var windows []actualSlotWindow
	var protected []activetime.CreditedInterval
	for _, rec := range candidates {
		id := strings.TrimLeft(rec.ID, "0")
		adoptedHours, _ := strconv.ParseFloat(rec.Text("actual_hours"), 64)
		context := fmt.Sprintf("issue #%s adopted actual", id)
		completion, completed, err := issue.CardCompletion(rec.Card.Raw)
		if err != nil {
			return nil, fmt.Errorf("%s has no trustworthy completion: %w", context, err)
		}
		if !completed {
			return nil, fmt.Errorf("%s has no trustworthy completion", context)
		}
		saved, ledger, _, savedErr := completedMeasurementSnapshot(ctx, root, rec, completion, current.Snapshot.Repository, plans, true)
		window := windowMissing
		if savedErr == nil && saved != nil {
			window = windowOverlap
			if !savedMeasurementMayOverlap(saved.Since, saved.Until, since, until) {
				window = windowDisjoint
			}
		}
		// The saved numeric's own window, not its retained claim, is the
		// first and strongest disjointness proof.
		if classifyAdopted(window, slotUnknown) == adoptionSkip {
			continue
		}
		slot := slotUnknown
		var slotErr error
		if windows == nil {
			var candidateWindows []actualSlotWindow
			candidateWindows, slotErr = candidateSlotWindows(current)
			if slotErr == nil && len(candidateWindows) == 0 {
				slotErr = fmt.Errorf("candidate has no pinned slot windows")
			}
			if slotErr == nil {
				windows = candidateWindows
			}
		}
		var overlapStart, overlapEnd time.Time
		if slotErr == nil {
			var events []tracker.OwnershipEvent
			events, slotErr = repo.EngagementHistory(rs.Ref, rec.Card.Path)
			if slotErr == nil {
				cutoff := until
				if window == windowOverlap {
					cutoff, _ = time.Parse(time.RFC3339Nano, saved.Until)
				}
				var engagements []tracker.EngagementWindow
				engagements, slotErr = tracker.DeriveEngagements(events, cutoff)
				if slotErr == nil {
					slot, overlapStart, overlapEnd = adoptedSlotRelationship(engagements, windows)
				}
			}
		}
		switch classifyAdopted(window, slot) {
		case adoptionSkip:
			continue
		case adoptionRefuse:
			if savedErr != nil {
				return nil, fmt.Errorf("%s has no trustworthy saved window: %w", context, savedErr)
			}
			if saved == nil {
				return nil, fmt.Errorf("%s has no trustworthy saved window", context)
			}
			if slotErr != nil {
				return nil, fmt.Errorf("%s has unknown slot relationship: %w", context, slotErr)
			}
			return nil, fmt.Errorf("%s has unknown slot relationship during overlapping saved window %s → %s", context, saved.Since, saved.Until)
		case adoptionReplay:
			// Only a proven same-slot overlap may use the retained replay.
		}
		if overlapStart.IsZero() {
			context = fmt.Sprintf("issue #%s adopted actual has overlapping saved window %s → %s and the same pinned slot", id, saved.Since, saved.Until)
		} else {
			context = fmt.Sprintf("issue #%s adopted actual overlaps this slot during %s → %s", id, overlapStart.Format(time.RFC3339Nano), overlapEnd.Format(time.RFC3339Nano))
		}
		prior, err := resolveOwnedReplay(ctx, root, id, *saved, ledger)
		if err != nil {
			return nil, fmt.Errorf("%s but cannot be replayed: %w", context, err)
		}
		if prior == nil || !prior.Replay {
			return nil, fmt.Errorf("%s but lacks pinned replay inputs", context)
		}
		measurement := computeOwnedActualFor(ctx, root, id, "", prior)
		if measurement.Status != actualMeasured {
			return nil, fmt.Errorf("%s but replay is unavailable: %s", context, measurement.Detail)
		}
		if math.Abs(measurement.Hours-adoptedHours) > 0.00501 {
			return nil, fmt.Errorf("%s but pinned replay %.4fh disagrees with adopted actual %.2fh", context, measurement.Hours, adoptedHours)
		}
		credited, err := creditedIntervals(id, measurement.Segments, measurement.CountedPhases, prior.SourceSlots)
		if err != nil {
			return nil, fmt.Errorf("issue #%s adopted actual: %w", id, err)
		}
		for _, interval := range credited {
			if _, _, ok := overlappingSlotInterval(interval.Slot, interval.Start, interval.End, windows); ok {
				if interval.Start.Before(since) {
					interval.Start = since
				}
				if interval.End.After(until) {
					interval.End = until
				}
				protected = append(protected, interval)
			}
		}
	}
	return protected, nil
}
