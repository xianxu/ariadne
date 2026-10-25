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
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// creditedIntervals protects only Build/Converge time included in the adopted
// number. A fractionally allocated run protects its whole intersection with a
// counted phase: the engine has no finer evidence for that fraction.
type adoptedCredit struct {
	Issue      string
	Slot       adoptedSlot
	Start, End time.Time
}

// adoptedSlot is the only identity admitted to exclusion comparisons. Zero
// means unknown; it can never prove that two slots differ.
type adoptedSlot struct {
	address workspace.Address
	known   bool
}

func parseAdoptedSlot(raw string) adoptedSlot {
	address, err := workspace.ParseAddress(raw)
	if err != nil || address.Repo == "" || address.String() != raw {
		return adoptedSlot{}
	}
	return adoptedSlot{address: address, known: true}
}

func (s adoptedSlot) String() string {
	if !s.known {
		return ""
	}
	return s.address.String()
}

func creditedIntervals(issue string, segments []activetime.Segment, counted []activetime.Phase, slots map[string]string) ([]adoptedCredit, error) {
	var out []adoptedCredit
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
			slot := parseAdoptedSlot(slots[segment.Source])
			if !slot.known {
				return nil, fmt.Errorf("credited source %q for issue #%s during %s → %s has no pinned slot (canonical identity required)", segment.Source, issue, start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))
			}
			out = append(out, adoptedCredit{Issue: issue, Slot: slot, Start: start, End: end})
		}
	}
	return out, nil
}

type actualSlotWindow struct {
	slot       adoptedSlot
	start, end time.Time
}

func candidateSlotWindows(inputs *ownedActualInputs) ([]actualSlotWindow, error) {
	var out []actualSlotWindow
	for source, windows := range inputs.Windows {
		raw := inputs.SourceSlots[source]
		slot := parseAdoptedSlot(raw)
		for _, window := range windows {
			if !slot.known {
				return nil, fmt.Errorf("candidate source %q during %s → %s has no pinned slot (canonical identity required)", source, window.Start.Format(time.RFC3339Nano), window.End.Format(time.RFC3339Nano))
			}
			out = append(out, actualSlotWindow{slot, window.Start, window.End})
		}
	}
	return out, nil
}

func overlappingSlotInterval(slot adoptedSlot, start, end time.Time, windows []actualSlotWindow) (time.Time, time.Time, bool) {
	for _, window := range windows {
		if slot.known && window.slot.known && window.slot.address == slot.address && start.Before(window.end) && window.start.Before(end) {
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
		if !window.slot.known {
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
			ownerSlot := parseAdoptedSlot(owner.Workspace)
			if !ownerSlot.known {
				return slotUnknown, time.Time{}, time.Time{}
			}
			for _, window := range windows {
				if ownerSlot.address != window.slot.address {
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

// legacyWindowDisjoint needs a proven immutable end and the earliest admitted
// engagement. Equality and missing bounds remain unknown.
func legacyWindowDisjoint(end, since time.Time) bool {
	return !end.IsZero() && !since.IsZero() && end.Before(since)
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
		completion, completed, completionErr := issue.CardCompletion(rec.Card.Raw)
		var saved *gatestate.MeasurementInputs
		var ledger gatestate.Ledger
		var savedErr error
		switch {
		case completionErr != nil:
			savedErr = fmt.Errorf("completion: %w", completionErr)
		case !completed:
			savedErr = fmt.Errorf("completion is missing")
		default:
			saved, ledger, _, savedErr = completedMeasurementSnapshot(ctx, root, rec, completion, current.Snapshot.Repository, plans, true)
		}
		window := windowMissing
		if savedErr == nil && saved != nil {
			window = windowOverlap
			if !savedMeasurementMayOverlap(saved.Since, saved.Until, since, until) {
				window = windowDisjoint
			}
		}
		// Legacy generations may replace an absent window only with a
		// revalidated immutable bound. The verifier also handles old imports
		// without provenance by proving tracker-root and source generation.
		// Saved input/correction failures never enter this recovery path.
		if completionErr == nil && ((!completed) || (savedErr == nil && saved == nil)) {
			legacy, explicit, legacyErr := issue.CardLegacyAdoption(rec.Card.Raw)
			if legacyErr != nil {
				savedErr = fmt.Errorf("legacy adoption: %w", legacyErr)
			} else if (!completed && (!explicit || legacy.WindowClass == issue.LegacyWindowlessTerminal)) ||
				(explicit && completed && legacy.WindowClass == issue.LegacyWindowlessClose) {
				end, proofErr := verifyLegacyAdoption(ctx, root, rs.Ref, rec, current.Snapshot.Repository, plans)
				if proofErr != nil {
					return nil, fmt.Errorf("%s has no trustworthy saved window: legacy adoption: %w", context, proofErr)
				} else if legacyWindowDisjoint(end, since) {
					window = windowDisjoint
				}
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
		if gatestate.IsSummaryMethod(saved.Method) {
			// A summarised close keeps its credited intervals, so exclusion
			// needs no replay (#603). Its hours still pin the adopted value.
			credits, lost, err := summaryProtection(id, *saved, adoptedHours, windows, since, until)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", context, err)
			}
			protected = append(protected, credits...)
			current.Unattributed = append(current.Unattributed, lost...)
			continue
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
		protected = append(protected, overlappingCredits(credited, windows, since, until)...)
	}
	return protected, nil
}

// overlappingCredits keeps the credits in this candidate's slot windows,
// clipped to its sample: the minutes it must not count again.
func overlappingCredits(credited []adoptedCredit, windows []actualSlotWindow, since, until time.Time) []activetime.CreditedInterval {
	var out []activetime.CreditedInterval
	for _, interval := range credited {
		if _, _, ok := overlappingSlotInterval(interval.Slot, interval.Start, interval.End, windows); !ok {
			continue
		}
		if interval.Start.Before(since) {
			interval.Start = since
		}
		if interval.End.After(until) {
			interval.End = until
		}
		if interval.End.After(interval.Start) {
			out = append(out, activetime.CreditedInterval{Issue: interval.Issue, Slot: interval.Slot.String(), Start: interval.Start, End: interval.End})
		}
	}
	return out
}

// summaryProtection is what a summarised earlier close protects in this
// candidate's slot windows (#603): its credited intervals, plus the coalesced
// gaps it over-protects, which the candidate records as unattributed.
func summaryProtection(id string, saved gatestate.MeasurementInputs, adoptedHours float64, windows []actualSlotWindow, since, until time.Time) ([]activetime.CreditedInterval, []gatestate.MeasurementUnattributed, error) {
	credits, widened, err := summaryCredits(id, saved, adoptedHours)
	if err != nil {
		return nil, nil, err
	}
	var lost []gatestate.MeasurementUnattributed
	for _, gap := range overlappingCredits(widened, windows, since, until) {
		lost = append(lost, gatestate.MeasurementUnattributed{Kind: gatestate.UnattributedSummaryWidened, Start: gap.Start.Format(time.RFC3339Nano), End: gap.End.Format(time.RFC3339Nano), Bounded: true, Reason: "issue #" + id + "'s summarised credits were coalesced over this gap"})
	}
	return overlappingCredits(credits, windows, since, until), lost, nil
}

// summaryCredits reads a summarised close's credited intervals (#603). Its
// recorded hours must match the adopted actual, as a replay's must.
func summaryCredits(id string, saved gatestate.MeasurementInputs, adoptedHours float64) (credited, widened []adoptedCredit, err error) {
	if saved.Summary == nil {
		return nil, nil, fmt.Errorf("summarised measurement lacks its summary")
	}
	if math.Abs(saved.Summary.Hours-adoptedHours) > 0.00501 {
		return nil, nil, fmt.Errorf("summarised measurement %.4fh disagrees with adopted actual %.2fh", saved.Summary.Hours, adoptedHours)
	}
	read := func(list []gatestate.MeasurementCredit) ([]adoptedCredit, error) {
		var out []adoptedCredit
		for _, c := range list {
			start, e1 := time.Parse(time.RFC3339Nano, c.Start)
			end, e2 := time.Parse(time.RFC3339Nano, c.End)
			slot := parseAdoptedSlot(c.Slot)
			if e1 != nil || e2 != nil || !slot.known || c.Issue != id {
				return nil, fmt.Errorf("summarised measurement credit %s %s → %s is not this issue's canonical credit", c.Slot, c.Start, c.End)
			}
			out = append(out, adoptedCredit{Issue: c.Issue, Slot: slot, Start: start, End: end})
		}
		return out, nil
	}
	if credited, err = read(saved.Summary.Credited); err != nil {
		return nil, nil, err
	}
	widened, err = read(saved.Summary.Widened)
	return credited, widened, err
}
