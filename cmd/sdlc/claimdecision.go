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

// claimDecision changes only reservation metadata on the observed remote record:
// open → working, stamped with the claiming workspace (#277) in the same card
// write. A working card is the owner's (errAlreadyMine), someone else's (refused,
// naming them), or unattributed (refused toward --adopt). A pre-tracker repository
// passes nil: its details are the whole record and carry no ownership.
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
	if !vocab.Issue().IsOpen(status) {
		if status == "working" && me != nil {
			if err := ownedBy(raw, id, *me); err != nil {
				return nil, err
			}
		}
		return nil, fmt.Errorf("remote issue #%d is not open (status %q); already-working issues are taken; continue existing work without claiming again", id, status)
	}
	fm = issue.SetField(fm, "status", "working")
	fm = issue.SetField(fm, "updated", today)
	if value, _ := issue.GetField(fm, "started"); value == "" {
		fm = issue.SetField(fm, "started", started)
	}
	out := []byte(issue.Compose(fm, body))
	if me == nil {
		return out, nil
	}
	return issue.SetCardClaimant(out, *me)
}

// ownedBy explains a working card's ownership to a would-be claimant: nil never
// (the caller refuses anyway), errAlreadyMine for the owner, else a refusal that
// names the owner or the adoption path.
func ownedBy(raw []byte, id int, me issue.Claimant) error {
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
		return fmt.Errorf("remote issue #%d is working, claimed by %s; coordinate with its owner — reassignment is operator-directed reclaim (#278), never a repeat claim", id, describeClaimant(recorded))
	default:
		return fmt.Errorf("remote issue #%d is working with no recorded owner (claimed before ownership, #277); if this workspace holds that work, record it with `sdlc claim --issue %d --adopt`", id, id)
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
// a working or blocked card that has none. The owner's repeat is
// errAlreadyMine; an owned card refuses — adoption never reassigns.
func adoptDecision(raw []byte, id string, me issue.Claimant) ([]byte, error) {
	fm, _, err := issue.Parse(string(raw))
	if err != nil {
		return nil, err
	}
	if status, _ := issue.GetField(fm, "status"); status != "working" && status != "blocked" {
		return nil, fmt.Errorf("#%s is %s; --adopt records the owner of working work claimed before #277 — an open issue is claimed with plain `sdlc claim --issue %s`", id, status, issue.CLIRef(id))
	}
	recorded, has, err := issue.CardClaimant(raw)
	if err != nil {
		return nil, fmt.Errorf("card #%s: %w", id, err)
	}
	if has {
		if issue.MatchClaimant(&recorded, me) == issue.OwnershipMine {
			return nil, errAlreadyMine
		}
		return nil, fmt.Errorf("#%s is owned by %s; --adopt never reassigns — that is operator-directed reclaim (#278)", id, describeClaimant(recorded))
	}
	return issue.SetCardClaimant(raw, me)
}
