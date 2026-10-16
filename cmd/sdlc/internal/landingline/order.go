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
	"slices"
	"time"
)

// MaxGrant bounds one fleet-blocking grant; a longer need is a renewal.
const MaxGrant = 24 * time.Hour

// Ticket is one waiting or admitted merge. Seq comes from the line's counter,
// so FIFO is a total order that clock skew can't break.
type Ticket struct {
	Seq      int64     `json:"seq"`
	Issue    int       `json:"issue"`
	Branch   string    `json:"branch"`
	PID      int       `json:"pid"`
	Enqueued time.Time `json:"enqueued"`
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

// ValidGrants folds the log in order (the latest entry per issue wins) and keeps
// the grants that confer priority at now. Authority is checked here as well as
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
			now.Before(g.Expires) && g.Expires.Sub(g.Granted) <= MaxGrant {
			valid[issue] = g
		}
	}
	return valid
}

// NextHead picks who is admitted next: the oldest ticket whose issue holds a
// valid grant, else the oldest ticket. displaced lists the older tickets a
// fleet-blocking head goes ahead of, for the audit log. An admitted head is
// never among the tickets passed in, so nothing here can preempt it.
func NextHead(tickets []Ticket, grants map[int]Grant) (head Ticket, displaced []Ticket, ok bool) {
	if len(tickets) == 0 {
		return Ticket{}, nil, false
	}
	sorted := slices.SortedFunc(slices.Values(tickets), func(a, b Ticket) int { return cmp.Compare(a.Seq, b.Seq) })
	for i, t := range sorted {
		if _, granted := grants[t.Issue]; granted {
			return t, sorted[:i:i], true
		}
	}
	return sorted[0], nil, true
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
