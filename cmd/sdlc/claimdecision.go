package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
	"go.yaml.in/yaml/v3"
)

// errAlreadyMine is claimDecision's answer to an owner claiming its own
// working issue again: nothing to write, and not a failure (#277).
var errAlreadyMine = errors.New("already claimed by this workspace")


// claimDecision changes only reservation metadata on the observed remote record.
// With an identity (an issue tracker repository) it is the ownership event
// `claim` (#283): an unowned card on a status that holds the lock gets this
// workspace as its owner and keeps its status — starting the lifecycle is
// start-plan's; unowned started work is a takeover (#284), which spends the
// release that left it unowned. A card already held is the owner's
// (errAlreadyMine) or someone else's (refused, naming them). A pre-tracker repository
// passes nil: it has no claimant, so its claim performs `start` in one step.
func claimDecision(raw []byte, id int, today, started string, me *issue.Claimant) ([]byte, error) {
	fm, body, err := issue.Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("remote issue #%d is malformed: %w", id, err)
	}
	var fields map[string]interface{}
	if err := yaml.Unmarshal([]byte(fm), &fields); err != nil {
		return nil, fmt.Errorf("remote issue #%d is malformed: %w", id, err)
	}
	rawID, _ := issue.GetField(fm, "id")
	parsedID, idErr := strconv.Atoi(rawID)
	if idErr != nil || parsedID != id {
		return nil, fmt.Errorf("remote issue #%d has missing or mismatched id", id)
	}
	status, ok := fields["status"].(string)
	if !ok {
		return nil, fmt.Errorf("remote issue #%d has no status", id)
	}
	model := vocab.Issue()
	if me != nil && model.CanHoldOwner(status) {
		_, has, err := issue.CardClaimant(raw)
		if err != nil {
			return nil, fmt.Errorf("remote issue #%d: %w", id, err)
		}
		if has {
			return nil, ownedBy(raw, id, status, *me)
		}
		// Unowned started work — released (#284) or claimed before #277 — is
		// taken over by a plain claim; its status stays.
	} else if !model.IsOpen(status) {
		return nil, fmt.Errorf("remote issue #%d is not open (status %q); already-working issues are taken; continue existing work without claiming again", id, status)
	}
	if me == nil {
		fm = issue.SetField(fm, "status", "working") // legacy: claim is start
	}
	fm = issue.SetField(fm, "updated", today)
	// The engagement stamp stays at the first claim: shaping time counts toward
	// the issue (#283 D1); gap-truncation bounds an idle claim→start interval.
	if value, _ := issue.GetField(fm, "started"); value == "" {
		fm = issue.SetField(fm, "started", started)
	}
	out := []byte(issue.Compose(fm, body))
	if me == nil {
		return out, nil
	}
	if out, err = issue.SetCardClaimant(out, *me); err != nil {
		return nil, err
	}
	// A claim takes the lock; a release that left it open is spent (#284).
	if _, released, err := issue.CardRelease(out); err != nil {
		return nil, err
	} else if released {
		return issue.SetCardRelease(out, nil)
	}
	return out, nil
}

// claimSetDecision is claim's decision over a set (#284), made on the bytes one
// attempt read: each card is judged by claimDecision. Any refusal aborts the
// whole set; the owner's own repeats are skipped; a set of nothing but repeats
// is errAlreadyMine. A card missing from current refuses.
func claimSetDecision(current map[string]tracker.Record, ids []string, today, started string, me issue.Claimant) (map[string][]byte, error) {
	out := map[string][]byte{}
	for _, id := range ids {
		rec, ok := current[id]
		if !ok {
			return nil, fmt.Errorf("no readable card #%s on the tracker", id)
		}
		n, err := strconv.Atoi(id)
		if err != nil {
			return nil, fmt.Errorf("card id %q is not numeric", id)
		}
		if len(ids) > 1 {
			if rel, ok, err := issue.CardRelease(rec.Raw); err != nil {
				return nil, fmt.Errorf("card #%s: %w", id, err)
			} else if ok && rel.Branch != "" {
				return nil, fmt.Errorf("#%s was handed off on branch %s; taking it over fetches and checks that out — claim it alone (`sdlc claim --issue %s`)", issue.CLIRef(id), rel.Branch, issue.CLIRef(id))
			}
		}
		next, err := claimDecision(rec.Raw, n, today, started, &me)
		if errors.Is(err, errAlreadyMine) {
			continue
		}
		if err != nil {
			return nil, err
		}
		out[id] = next
	}
	if len(out) == 0 {
		return nil, errAlreadyMine
	}
	return out, nil
}

// ownedBy explains a held card to a would-be claimant: errAlreadyMine for the
// owner, else a refusal that names the owner or the adoption path.
func ownedBy(raw []byte, id int, status string, me issue.Claimant) error {
	recorded, has, err := issue.CardClaimant(raw)
	if err != nil {
		return fmt.Errorf("remote issue #%d: %w", id, err)
	}
	var rec *issue.Claimant
	if has {
		rec = &recorded
	}
	switch issue.MatchClaimant(rec, me) {
	case issue.OwnershipMine:
		return errAlreadyMine
	case issue.OwnershipForeign:
		return fmt.Errorf("remote issue #%d is %s, claimed by %s; coordinate with its owner — reassignment is the operator-directed `sdlc reclaim --issue %d`, never a repeat claim", id, status, describeClaimant(recorded), id)
	default:
		return fmt.Errorf("remote issue #%d is %s with an unreadable owner", id, status)
	}
}

// describeClaimant names an owner for humans: operator, machine, slot, path.
func describeClaimant(c issue.Claimant) string {
	where := c.Worktree
	if c.Workspace != "" {
		where = c.Workspace + " (" + c.Worktree + ")"
	}
	return fmt.Sprintf("%s on %s at %s", c.Operator, c.MachineName, where)
}
