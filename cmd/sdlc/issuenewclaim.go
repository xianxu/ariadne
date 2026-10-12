// issuenewclaim.go — who holds a freshly filed issue (#308). Filing from a
// numbered slot (`repo:N`, N ≥ 1) is that slot taking the work, so the issue
// is claimed by it at creation. Filing from `:0` (the operator's resting
// brainstorm) or from a checkout outside the slot layout leaves it unclaimed:
// whichever slot later works it claims it.
package main

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// creatorClaims decides from the creator's slot label (Claimant.Workspace,
// "" outside the slot layout) whether a new issue is claimed by its creator.
// Pure.
func creatorClaims(slot string) bool {
	if slot == "" {
		return false
	}
	a, err := workspace.ParseAddress(slot)
	return err == nil && a.Repo != "" && a.Slot >= 1
}

// claimAtCreation claims a just-published issue for the creating slot through
// claim's own card decision and token, so claim ages and active-time windows
// read it as a claim. Unlike `sdlc claim` it refreshes nothing local: filing
// writes no checkout. The issue exists either way, so a failure here is
// reported with the next action rather than failing the filing.
func claimAtCreation(env *trackerEnv, stderr io.Writer, id string, me issue.Claimant) {
	today, started := time.Now().Format("2006-01-02"), startedClock()
	decide := func(current map[string]tracker.Record) (map[string][]byte, error) {
		return claimSetDecision(current, []string{id}, today, started, me)
	}
	err := cardsPublish(env, []string{id}, operationToken(tracker.OpClaim), nil, decide, nil)
	invalidateIssueRecords(env.ctx)
	if errors.Is(err, errAlreadyMine) {
		err = nil
	}
	if err = uncertainCardWrite(err, "sdlc claim --issue "+issue.CLIRef(id)); err != nil {
		cwarn(stderr, fmt.Sprintf("#%s is filed but not claimed: %v\n      claim it with `sdlc claim --issue %s`", issue.CLIRef(id), err, issue.CLIRef(id)))
		return
	}
	cok(stderr, fmt.Sprintf("#%s claimed by this slot (%s); status stays open until `sdlc start-plan --issue %s` starts it", issue.CLIRef(id), claimantWhere(me), issue.CLIRef(id)))
}
