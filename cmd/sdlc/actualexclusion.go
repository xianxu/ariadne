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

// creditedIntervals is the conservative occupied-time bound of a pinned
// actual. A fractionally allocated run protects its whole wall interval: the
// engine has no evidence for which sub-interval earned that fraction.
func creditedIntervals(issue string, segments []activetime.Segment, slots map[string]string) ([]activetime.CreditedInterval, error) {
	var out []activetime.CreditedInterval
	for _, segment := range segments {
		if segment.Alloc[issue] <= 0 {
			continue
		}
		slot := slots[segment.Source]
		if slot == "" {
			return nil, fmt.Errorf("credited source %q for issue #%s during %s → %s has no pinned slot", segment.Source, issue, segment.Start.Format(time.RFC3339Nano), segment.End.Format(time.RFC3339Nano))
		}
		out = append(out, activetime.CreditedInterval{Issue: issue, Slot: slot, Start: segment.Start, End: segment.End})
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
		saved, ledger, _, err := completedMeasurementSnapshot(ctx, root, rec, completion, current.Snapshot.Repository, plans, true)
		if err != nil {
			return nil, fmt.Errorf("%s has no trustworthy saved window: %w", context, err)
		}
		if saved == nil {
			return nil, fmt.Errorf("%s has no trustworthy saved window", context)
		}
		priorEnd, _ := time.Parse(time.RFC3339Nano, saved.Until)
		if !savedMeasurementMayOverlap(saved.Since, saved.Until, since, until) {
			continue
		}
		if windows == nil {
			windows, err = candidateSlotWindows(current)
			if err != nil || len(windows) == 0 {
				return nil, err
			}
		}
		events, err := repo.EngagementHistory(rs.Ref, rec.Card.Path)
		if err != nil {
			return nil, fmt.Errorf("issue #%s adopted actual engagement: %w", id, err)
		}
		engagements, err := tracker.DeriveEngagements(events, priorEnd)
		if err != nil {
			return nil, fmt.Errorf("issue #%s adopted actual engagement: %w", id, err)
		}
		relevant := false
		var overlapStart, overlapEnd time.Time
		for _, engagement := range engagements {
			for _, owner := range engagement.Owners {
				if start, end, ok := overlappingSlotInterval(owner.Workspace, engagement.Start, engagement.End, windows); ok {
					relevant = true
					overlapStart, overlapEnd = start, end
				}
			}
		}
		if !relevant {
			continue
		}
		context = fmt.Sprintf("issue #%s adopted actual overlaps this slot during %s → %s", id, overlapStart.Format(time.RFC3339Nano), overlapEnd.Format(time.RFC3339Nano))
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
		credited, err := creditedIntervals(id, measurement.Segments, prior.SourceSlots)
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
