// reclaim.go — `sdlc reclaim` (#278): operator-directed transfer of an issue's
// recorded responsibility (#277's claimant) to the workspace running it, after
// the operator has coordinated out of band. Two steps: inspect (read-only)
// shows the current and proposed owner and the card revision; confirm, pinned
// to that revision with --expect and justified with --reason, publishes the
// transfer by compare-and-swap. The tracker commit's trailers are the record.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

const (
	reclaimFromTrailer   = "Reclaim-From"
	reclaimToTrailer     = "Reclaim-To"
	reclaimReasonTrailer = "Reclaim-Reason"
)

// reclaimDecision is reclaim's pure core: the card with me as its claimant, and
// the claimant it replaces. The card must be owned work (working, blocked or
// codecomplete) with a recorded owner; rev is the card revision read now and
// expect the one the operator inspected. The owner's repeat is errAlreadyMine —
// so an identical retry, or one after a lost publication response, is decided
// by the card itself.
func reclaimDecision(card []byte, rev, expect, reason string, me issue.Claimant) ([]byte, issue.Claimant, error) {
	fm, _, err := issue.Parse(string(card))
	if err != nil {
		return nil, issue.Claimant{}, err
	}
	id, _ := issue.GetField(fm, "id")
	switch status, _ := issue.GetField(fm, "status"); status {
	case "working", "blocked", "codecomplete":
	case "open":
		return nil, issue.Claimant{}, fmt.Errorf("#%s is open; nobody holds it — `sdlc claim --issue %s` takes it", id, issue.CLIRef(id))
	default:
		return nil, issue.Claimant{}, fmt.Errorf("#%s is %s; there is no live responsibility to reclaim", id, status)
	}
	recorded, has, err := issue.CardClaimant(card)
	if err != nil {
		return nil, issue.Claimant{}, err
	}
	if !has {
		return nil, issue.Claimant{}, fmt.Errorf("#%s has no recorded owner (claimed before #277); record one with `sdlc claim --issue %s --adopt`", id, issue.CLIRef(id))
	}
	if issue.MatchClaimant(&recorded, me) == issue.OwnershipMine {
		return nil, recorded, errAlreadyMine
	}
	if expect == "" {
		return nil, recorded, errors.New("--expect is required: the card revision you inspected (run `sdlc reclaim --issue N` to see it)")
	}
	if rev != expect {
		return nil, recorded, fmt.Errorf("#%s changed since you inspected it (now %s, you saw %s); inspect it again before reclaiming", id, shortOID(rev), shortOID(expect))
	}
	reason = strings.TrimSpace(reason)
	if reason == "" || strings.ContainsAny(reason, "\r\n") {
		return nil, recorded, errors.New("--reason is required: one line saying why responsibility moves (the operator's out-of-band decision)")
	}
	next, err := issue.SetCardClaimant(card, me)
	return next, recorded, err
}

// reclaimTrailers records a transfer in its tracker commit.
func reclaimTrailers(from, to issue.Claimant, reason string) []string {
	return []string{
		reclaimFromTrailer + ": " + describeClaimant(from),
		reclaimToTrailer + ": " + describeClaimant(to),
		reclaimReasonTrailer + ": " + strings.TrimSpace(reason),
	}
}

// reclaimEvent is one past transfer read back from tracker history.
type reclaimEvent struct{ Commit, Date, From, To, Reason string }

// parseReclaimTrailers reads the trailers reclaimTrailers wrote, from one
// commit message; ok is false for a commit that is not a reclaim.
func parseReclaimTrailers(message string) (from, to, reason string, ok bool) {
	for _, line := range strings.Split(message, "\n") {
		key, value, found := strings.Cut(line, ": ")
		if !found {
			continue
		}
		switch key {
		case reclaimFromTrailer:
			from = value
		case reclaimToTrailer:
			to = value
		case reclaimReasonTrailer:
			reason = value
		}
	}
	return from, to, reason, from != "" && to != "" && reason != ""
}
