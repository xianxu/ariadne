package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
	"go.yaml.in/yaml/v3"
)

// errAlreadyMine is claimDecision's answer to an owner claiming its own
// working issue again: nothing to write, and not a failure (#277).
var errAlreadyMine = errors.New("already claimed by this workspace")

// claimDecision changes only reservation metadata on the observed remote record.
// With an identity (an issue tracker repository) it is the ownership event
// `claim` (#283): an unowned open card gets this workspace as its owner and
// keeps its status — starting the lifecycle is start-plan's. A card already
// held is the owner's (errAlreadyMine) or someone else's (refused, naming them);
// unowned started work is refused toward --adopt. A pre-tracker repository
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
		if has || !model.IsOpen(status) {
			return nil, ownedBy(raw, id, status, *me)
		}
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
	return issue.SetCardClaimant(out, *me)
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
		return fmt.Errorf("remote issue #%d is %s with no recorded owner (claimed before #277); if this workspace holds that work, record it with `sdlc claim --issue %d --adopt`", id, status, id)
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

// adoptDecision is claim --adopt's pure core (#277): record me as the owner of
// a started (active, per the model) card that has none. The owner's repeat is
// errAlreadyMine; an owned card refuses — adoption never reassigns.
func adoptDecision(raw []byte, id string, me issue.Claimant) ([]byte, error) {
	fm, _, err := issue.Parse(string(raw))
	if err != nil {
		return nil, err
	}
	if status, _ := issue.GetField(fm, "status"); !vocab.Issue().IsActive(status) {
		return nil, fmt.Errorf("#%s is %s; --adopt records the owner of started work claimed before #277 — an open issue is claimed with plain `sdlc claim --issue %s`", id, status, issue.CLIRef(id))
	}
	recorded, has, err := issue.CardClaimant(raw)
	if err != nil {
		return nil, fmt.Errorf("card #%s: %w", id, err)
	}
	if has {
		if issue.MatchClaimant(&recorded, me) == issue.OwnershipMine {
			return nil, errAlreadyMine
		}
		return nil, fmt.Errorf("#%s is owned by %s; --adopt never reassigns — that is the operator-directed `sdlc reclaim --issue %s`", id, describeClaimant(recorded), issue.CLIRef(id))
	}
	return issue.SetCardClaimant(raw, me)
}
