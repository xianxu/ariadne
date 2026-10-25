package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// adoptionReadSet is invocation-local measurement authority, never persisted.
// Copy the card bytes at their original read so an intervening adoption cannot
// be blessed by a later capture. The measurement window pins the shared filter.
type adoptionReadSet struct {
	issue, details string
	since, until   time.Time
	cards          map[string]adoptionCardAuthority
}

type adoptionCardAuthority struct{ path, blob, raw string }

// adoptedExclusionCandidates is the one cheap admission filter, shared by the
// measurement and its finalization validator. It deliberately does not decide
// saved-window/slot overlap: that remains loadAdoptedExclusions' authority.
func adoptedExclusionCandidates(records []tracker.IssueRecord, own string, since, until time.Time) ([]tracker.IssueRecord, error) {
	var candidates []tracker.IssueRecord
	for _, rec := range records {
		id := strings.TrimLeft(rec.ID, "0")
		if id == strings.TrimLeft(own, "0") {
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
			continue
		}
		candidates = append(candidates, rec)
	}
	return candidates, nil
}

func captureAdoptionReadSet(own, details string, since, until time.Time, candidates []tracker.IssueRecord) *adoptionReadSet {
	set := &adoptionReadSet{issue: own, details: details, since: since, until: until, cards: make(map[string]adoptionCardAuthority, len(candidates))}
	for _, rec := range candidates {
		set.cards[rec.ID] = adoptionCardAuthority{path: rec.Card.Path, blob: rec.Card.BlobOID, raw: string(rec.Card.Raw)}
	}
	return set
}

func compareAdoptionReadSet(saved *adoptionReadSet, records []tracker.IssueRecord) error {
	if saved == nil {
		return nil
	}
	if saved.cards == nil {
		return fmt.Errorf("adopted-actual preparation read-set is unavailable")
	}
	candidates, err := adoptedExclusionCandidates(records, saved.issue, saved.since, saved.until)
	if err != nil {
		return err
	}
	current := captureAdoptionReadSet(saved.issue, saved.details, saved.since, saved.until, candidates)
	for id, card := range current.cards {
		if prior, ok := saved.cards[id]; !ok || prior != card {
			return fmt.Errorf("adopted-actual preparation stale: issue #%s adoption added or changed", strings.TrimLeft(id, "0"))
		}
	}
	for id := range saved.cards {
		if _, ok := current.cards[id]; !ok {
			return fmt.Errorf("adopted-actual preparation stale: issue #%s adoption removed", strings.TrimLeft(id, "0"))
		}
	}
	return nil
}

// observeClosePreparationCards starts a fresh observation, shared by own-card
// and adoption comparisons. It must not reuse preparation's cached generation.
func observeClosePreparationCards(ctx context.Context, root, details string) (tracker.Records, error) {
	ctx = withIssueRecordsScope(commandContext(ctx))
	return loadIssueRecordsAt(ctx, root, details, tracker.Fresh)
}
