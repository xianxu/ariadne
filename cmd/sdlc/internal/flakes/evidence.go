package flakes

import (
	"slices"
	"sort"
	"time"
)

// Attempt is the part of a recorded suite attempt the evidence rules read.
type Attempt struct {
	ID         string
	SHA        string
	Passed     bool
	Baseline   bool // merge's unrecorded diagnostic run on main
	Legacy     bool // a historical log without retry history
	Failures   []string
	Kind       Kind // the gate's recorded outcome; "" for attempts from before #394
	FinishedAt time.Time
}

// Confirmed reports whether a failed attempt has same-exact-head fail→pass
// evidence: a later non-baseline attempt on the same SHA passed (#390's
// intermittency proof). It is the one rule both test-stats and registration
// use.
//
// order must be in Sort's order; i is the failed attempt's index in it.
// "Later" is that order (finish time, then ID), which test-stats also uses.
func Confirmed(order []Attempt, i int) (passedID string, ok bool) {
	failed := order[i]
	if failed.Passed || failed.Legacy || failed.Baseline {
		return "", false
	}
	for _, b := range order[i+1:] {
		if b.SHA == failed.SHA && b.Passed && !b.Legacy && !b.Baseline {
			return b.ID, true
		}
	}
	return "", false
}

// SameHeadFailPass finds registration evidence for test: an attempt where it
// failed, confirmed by a later pass on the same head.
func SameHeadFailPass(all []Attempt, test string) (failedID, passedID string, ok bool) {
	order := Sort(all)
	for i, a := range order {
		if !slices.Contains(a.Failures, test) {
			continue
		}
		if p, ok := Confirmed(order, i); ok {
			return a.ID, p, true
		}
	}
	return "", "", false
}

// Unexplained lists the failed attempts on sha that nothing explains: not a
// recorded known-flake outcome, and not every failure a test registered now.
// A success may not be posted for sha while any remain (#394): that is what
// stops a blind full rerun from turning an unknown failure green.
func Unexplained(all []Attempt, sha string, r Registry) []Attempt {
	var out []Attempt
	for _, a := range Sort(all) {
		if a.SHA != sha || a.Passed || a.Baseline || a.Legacy || a.Kind.Lands() {
			continue
		}
		explained := len(a.Failures) > 0
		for _, f := range a.Failures {
			if _, ok := r.Lookup(f); !ok {
				explained = false
			}
		}
		if a.Kind == Unattributable || !explained {
			out = append(out, a)
		}
	}
	return out
}

// Sort orders attempts by finish time, then ID, without changing all.
func Sort(all []Attempt) []Attempt {
	s := slices.Clone(all)
	sort.SliceStable(s, func(i, j int) bool {
		if s[i].FinishedAt.Equal(s[j].FinishedAt) {
			return s[i].ID < s[j].ID
		}
		return s[i].FinishedAt.Before(s[j].FinishedAt)
	})
	return s
}
