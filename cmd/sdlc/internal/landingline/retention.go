package landingline

import "time"

// Retention is the pure rule for what a sweep keeps (ARCH-FUNERAL). Its
// invariant, tested across random logs: a sweep never changes which grants
// are valid, ValidGrants(kept, now) == ValidGrants(all, now), and it never
// keeps an entry dated outside the last Retention.

// fresh reports whether at lies within the last Retention: not older, and not
// in the future (a future date is a forged or skewed entry).
func fresh(at, now time.Time) bool {
	age := now.Sub(at)
	return age >= 0 && age <= Retention
}

// KeepEvents marks the events a sweep keeps.
func KeepEvents(events []Event, now time.Time) []bool {
	keep := make([]bool, len(events))
	for i, e := range events {
		keep[i] = fresh(e.At, now)
	}
	return keep
}

// KeepGrants marks the grant entries a sweep keeps. Only an issue's latest
// entry decides its priority, so the latest is never dropped while it could
// still matter: if it is fresh, it stays, with the issue's other fresh
// entries as history; if it isn't, it confers nothing (a valid grant is under
// MaxGrant old), and the issue's whole history goes with it, so an older
// entry can't become the latest.
func KeepGrants(entries []Grant, now time.Time) []bool {
	latest := map[int]int{}
	for i, g := range entries {
		latest[g.Issue] = i
	}
	keep := make([]bool, len(entries))
	for i, g := range entries {
		keep[i] = fresh(entries[latest[g.Issue]].Granted, now) && (i == latest[g.Issue] || fresh(g.Granted, now))
	}
	return keep
}
