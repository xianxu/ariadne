package flakes

import (
	"slices"
	"sort"
	"strings"
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
	// Targeted marks a flake rerun. Its pass retains legacy Go confirmation,
	// while typed registration requires a later complete-suite pass.
	Targeted bool
	// TypedFacts are ephemeral facts from the attempt's bounded, digest-verified
	// saved log. They are never inferred from a passing exit or persisted as a
	// case inventory. Nil means the typed execution has not been proved.
	TypedFacts *UnitFacts
	// RunnerContractSHA256 binds typed fail and pass evidence to the same
	// declared runner contract; legacy Go evidence does not need it.
	RunnerContractSHA256 string
}

// Confirmed reports whether a failed attempt has same-exact-head fail→pass
// evidence: a later non-baseline attempt on the same SHA passed (#390's
// intermittency proof), and was not a subset run (#473). It is the one rule both test-stats and registration
// use. Typed failures additionally need exact saved-log failure and pass proof
// under the same declared runner contract from a complete-suite pass (never
// a targeted flake child); every typed failure in the row must be covered.
//
// order must be in Sort's order; i is the failed attempt's index in it.
// "Later" is that order (finish time, then ID), which test-stats also uses.
func Confirmed(order []Attempt, i int) (passedID string, ok bool) {
	failed := order[i]
	if failed.Passed || failed.Legacy || failed.Baseline || failed.Kind == CouldNotRun {
		return "", false
	}
	for _, b := range order[i+1:] {
		if !confirmationCandidate(failed, b) {
			continue
		}
		covered := true
		for _, name := range failed.Failures {
			if needsTypedProof(name) && !passCoversTyped(failed, b, name) {
				covered = false
				break
			}
		}
		if covered {
			return b.ID, true
		}
	}
	return "", false
}

func confirmationCandidate(failed, passed Attempt) bool {
	return passed.SHA == failed.SHA && passed.Passed && !passed.Legacy && !passed.Baseline && !passed.Subset
}

func needsTypedProof(name string) bool {
	return strings.HasPrefix(name, "script:") || strings.HasPrefix(name, "lua:")
}

// passCoversTyped requires an observed exact failure and positive execution
// proof from a complete-suite pass. A targeted child's exact proof supports
// resume, but cannot establish typed registration's full-suite intermittency.
func passCoversTyped(failed, passed Attempt, name string) bool {
	if passed.Targeted || failed.TypedFacts == nil || passed.TypedFacts == nil || failed.RunnerContractSHA256 == "" || failed.RunnerContractSHA256 != passed.RunnerContractSHA256 {
		return false
	}
	if len(failed.TypedFacts.Unattributed) != 0 || len(passed.TypedFacts.Failures) != 0 || !slices.Contains(failed.TypedFacts.Failures, name) {
		return false
	}
	return len(RequiredUnitProof(*passed.TypedFacts, []string{name})) == 0
}

// confirmedName applies the same timing and attempt eligibility as Confirmed,
// but only proves the requested failure. Another typed failure in that run
// must not erase this name's proof or tighten legacy Go registration.
func confirmedName(order []Attempt, i int, name string) (string, bool) {
	failed := order[i]
	if failed.Passed || failed.Legacy || failed.Baseline || failed.Kind == CouldNotRun {
		return "", false
	}
	for _, passed := range order[i+1:] {
		if confirmationCandidate(failed, passed) && (!needsTypedProof(name) || passCoversTyped(failed, passed, name)) {
			return passed.ID, true
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
		if p, ok := confirmedName(order, i, test); ok {
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
		if a.SHA != sha || a.Passed || a.Baseline || a.Legacy || a.Kind.Lands() || a.Kind == CouldNotRun {
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
	return len(failures) > 0 && len(Uncovered(failures, subtests, r)) == 0
}

// Uncovered is the failing leaves no registered name covers, by the gate's rule
// (explain): what a later name, an open flake issue's or a fixed one's, has to
// account for (#479).
func Uncovered(failures, subtests []string, r Registry) []string {
	_, unknown := explain(Result{Failures: failures, Subtests: subtests}, r)
	return unknown
}

// Sort orders attempts by finish time, then ID, without changing all.
// Explainable reports whether registering a test can later explain a failed
// attempt of this kind (#394): never one the rule judged and rejected.
func Explainable(k Kind) bool { return !neverExplained[k] }

var neverExplained = map[Kind]bool{Unattributable: true, KnownFlakeTwice: true, TooManyKnownFlaky: true, RuleOff: true, RerunInconclusive: true, CouldNotRun: true}

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
