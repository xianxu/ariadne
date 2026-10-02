// startdecision.go — start-plan's pure core (#283): the owner starts the
// lifecycle. Claiming took the lock; this moves status, and only along the
// model's `start` edge.
package main

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// startDecision returns the card after start-plan, and whether it changed. An
// owned open card becomes working (stamped `started` if it never was); an owned
// started card is unchanged. Unowned, foreign and terminal cards refuse.
func startDecision(card []byte, id string, me issue.Claimant, today, started string) ([]byte, bool, error) {
	fm, _, err := issue.Parse(string(card))
	if err != nil {
		return nil, false, err
	}
	model := vocab.Issue()
	status, _ := issue.GetField(fm, "status")
	if !model.CanHoldOwner(status) {
		return nil, false, fmt.Errorf("#%s is %s; there is nothing to plan", id, status)
	}
	recorded, has, err := issue.CardClaimant(card)
	if err != nil {
		return nil, false, fmt.Errorf("card #%s: %w", id, err)
	}
	if !has {
		return nil, false, fmt.Errorf("#%s has no owner; `sdlc claim --issue %s` takes it before planning", id, issue.CLIRef(id))
	}
	if issue.MatchClaimant(&recorded, me) != issue.OwnershipMine {
		return nil, false, fmt.Errorf("#%s is owned by %s; planning belongs to its owner", id, describeClaimant(recorded))
	}
	tr := model.TransitionForEvent(status, "start")
	if tr == nil {
		return card, false, nil // already started
	}
	out, err := issue.SetCardField(card, "status", tr.To)
	if err == nil {
		out, err = issue.SetCardField(out, "updated", today)
	}
	if cur, _ := issue.GetField(fm, "started"); err == nil && strings.TrimSpace(cur) == "" {
		out, err = issue.SetCardField(out, "started", started)
	}
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}
