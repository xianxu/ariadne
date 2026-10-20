package flakes

import (
	"slices"
	"sort"
	"time"
)

// Attempt is the part of a recorded suite attempt the evidence rules read.
type Attempt struct {
	ID       string
	SHA      string
	Passed   bool
	Baseline bool // merge's unrecorded diagnostic run on main
	Legacy   bool // a historical log without retry history
	Failures []string
	// Subtests are the failing subtests the attempt's saved log names (#466);
	// records keep top-level failures, so the IO layer reads them from the
	// log. None (an unreadable log) leaves each failure its own leaf, which
	// only a registered whole test explains: fail closed.
	Subtests   []string
	Kind       Kind // the gate's recorded outcome; "" for attempts from before #394
	FinishedAt time.Time
	// Subset marks a run of a selection, not every test (the affected suite,
	// #473): its failures need explaining like any run's, but its pass proves
	// nothing about a test it may not have run, so it never confirms.
	Subset bool
}

// Confirmed reports whether a failed attempt has same-exact-head fail→pass
// evidence: a later non-baseline attempt on the same SHA passed (#390's
// intermittency proof), and was not a subset run (#473). It is the one rule both test-stats and registration
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
		if b.SHA == failed.SHA && b.Passed && !b.Legacy && !b.Baseline && !b.Subset {
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
//
// Registering later explains only a failure the rule never judged (unknown,
// persistent-main, registry unavailable, or an attempt from before #394). A
// failure the rule judged and rejected stays unexplained whatever is
// registered later: a known flake that failed twice, too many at once, the
// rule turned off, or a failure that names no test.
func Unexplained(all []Attempt, sha string, r Registry) []Attempt {
	var out []Attempt
	for _, a := range Sort(all) {
		if a.SHA != sha || a.Passed || a.Baseline || a.Legacy || a.Kind.Lands() {
			continue
		}
		if neverExplained[a.Kind] || !Explained(a.Failures, a.Subtests, r) {
			out = append(out, a)
		}
	}
	return out
}

// Explained reports whether registered names cover every failure, by the
// gate's own rule (explain: each failing leaf covered by itself or an
// ancestor). It is the one coverage answer for the gate, the success guard
// and the flake view (#466); no failures at all is not explained.
func Explained(failures, subtests []string, r Registry) bool {
	if len(failures) == 0 {
		return false
	}
	_, unknown := explain(Result{Failures: failures, Subtests: subtests}, r)
	return len(unknown) == 0
}

// Sort orders attempts by finish time, then ID, without changing all.
// Explainable reports whether registering a test can later explain a failed
// attempt of this kind (#394): never one the rule judged and rejected.
func Explainable(k Kind) bool { return !neverExplained[k] }

var neverExplained = map[Kind]bool{Unattributable: true, KnownFlakeTwice: true, TooManyKnownFlaky: true, RuleOff: true, RerunInconclusive: true}

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
