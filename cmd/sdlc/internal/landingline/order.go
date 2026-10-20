// Package landingline orders a repository's landings (#393): one `sdlc merge`
// at a time integrates and lands, so each landing runs one integration suite.
//
// This file is the pure core: who goes next among the live tickets, and which
// logged grants put an issue in the fleet-blocking class. line.go and log.go
// are the IO shell (flock-held ticket files and append-only logs).
package landingline

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
)

// MaxGrant bounds one fleet-blocking grant; a longer need is a renewal.
const MaxGrant = 24 * time.Hour

// Ticket is one waiting or admitted merge. Seq comes from the line's counter,
// so FIFO is a total order that clock skew can't break.
type Ticket struct {
	QueueOperationID string    `json:"queue_operation_id,omitempty"`
	QueuePID         int       `json:"queue_pid,omitempty"`
	QueuePIDStarted  string    `json:"queue_pid_started,omitempty"`
	QueueRoot        string    `json:"queue_root,omitempty"`
	Seq              int64     `json:"seq"`
	Issue            int       `json:"issue"`
	Branch           string    `json:"branch"`
	PID              int       `json:"pid"`
	Enqueued         time.Time `json:"enqueued"`
	// Unreadable marks a held ticket or head.lock whose contents don't parse:
	// present, never reported as absent.
	Unreadable bool `json:"-"`
}

// Grantor is who granted priority: a workspace address (or "operator shell")
// and its slot, nil outside a numbered-slot layout.
type Grantor struct {
	Address string `json:"address"`
	Slot    *int   `json:"slot,omitempty"`
	Name    string `json:"name,omitempty"`
}

// Grant is one entry of the grants log; a revoke is an entry with Revoked set.
type Grant struct {
	Issue   int       `json:"issue"`
	Grantor Grantor   `json:"grantor"`
	Reason  string    `json:"reason"`
	Granted time.Time `json:"granted"`
	Expires time.Time `json:"expires"`
	Revoked bool      `json:"revoked,omitempty"`
}

func (g Grant) boundedExpiry() bool {
	return !g.Expires.IsZero() && g.Expires.After(g.Granted) && g.Expires.Sub(g.Granted) <= MaxGrant
}

// ValidGrants folds the log in order (the latest entry per issue wins) and keeps
// the grants that confer priority at now: granted no later than now, unexpired,
// and at most MaxGrant long, so a future-dated entry confers nothing. Authority is checked here as well as
// when a grant is written, so an entry recorded from a numbered slot confers
// nothing. Trust is cooperative: this guards against agents' mistakes on one
// machine, not against a process that forges the grantor fields.
func ValidGrants(entries []Grant, now time.Time) map[int]Grant {
	latest := map[int]Grant{}
	for _, g := range entries {
		latest[g.Issue] = g
	}
	valid := map[int]Grant{}
	for issue, g := range latest {
		if !g.Revoked && g.Grantor.Address != "" && authorized(g.Grantor.Slot) && g.Reason != "" &&
			!g.Granted.After(now) && g.boundedExpiry() && now.Before(g.Expires) {
			valid[issue] = g
		}
	}
	return valid
}

// Order is the admission order of the waiting tickets: the fleet-blocking ones
// (whose issue holds a valid grant) first, then the rest, FIFO by Seq within
// each class.
func Order(tickets []Ticket, grants map[int]Grant) []Ticket {
	rank := func(t Ticket) int {
		if _, granted := grants[t.Issue]; granted {
			return 0
		}
		return 1
	}
	return slices.SortedFunc(slices.Values(tickets), func(a, b Ticket) int {
		return cmp.Or(cmp.Compare(rank(a), rank(b)), cmp.Compare(a.Seq, b.Seq))
	})
}

// GrantHold explains why a normal ticket cannot acquire a free head. The
// existing valid grants are the hold; no ticket needs to exist yet (#470).
// Sorted output stays stable across polls and names every outstanding recovery.
func GrantHold(issue int, grants map[int]Grant) string {
	if _, granted := grants[issue]; granted || len(grants) == 0 {
		return ""
	}
	var reasons []string
	for _, n := range slices.Sorted(maps.Keys(grants)) {
		g := grants[n]
		reasons = append(reasons, fmt.Sprintf("fleet-blocking grant #%d until %s: %s; in this repository's checkout run: sdlc landing revoke --issue %d", n, g.Expires.Format(time.RFC3339), g.Reason, n))
	}
	return strings.Join(reasons, "; ")
}

// NextHead picks who is admitted next among tickets: Order's first one,
// unless a fleet-blocking grant holds normal admissions. An
// admitted head has left the waiting tickets, so nothing here can preempt it.
func NextHead(tickets []Ticket, grants map[int]Grant) (head Ticket, ok bool) {
	order := Order(tickets, grants)
	if len(order) == 0 || GrantHold(order[0].Issue, grants) != "" {
		return Ticket{}, false
	}
	return order[0], true
}

// Passed is who head went ahead of among the live tickets, for the audit
// log: the older ones (a fleet-blocking head passed them) and the ones
// ordered before it (they were not admissible, #449). The one definition.
func Passed(live []Ticket, grants map[int]Grant, head Ticket) []Ticket {
	var passed []Ticket
	before := true
	for _, t := range Order(live, grants) {
		if t.Seq == head.Seq {
			before = false
			continue
		}
		if before || t.Seq < head.Seq {
			passed = append(passed, t)
		}
	}
	return passed
}

// Authorize is the write-side check for a grant: refused when the caller's
// workspace or the Couch slot it runs in ($COUCH_TREE) is a numbered slot.
func Authorize(callerSlot, treeSlot *int) error {
	if !authorized(callerSlot) || !authorized(treeSlot) {
		return errors.New("a fleet-blocking grant comes from the operator or the TL (:0), not a worker slot")
	}
	return nil
}

func authorized(slot *int) bool { return slot == nil || *slot == 0 }
