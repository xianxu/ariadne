package flakes

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Policy is what a suite does with a failure that only registered flakes
// explain (#394 D2): the operator's choice, in `.sdlc/local-checks.json`.
type Policy string

const (
	PolicyRerun  Policy = "rerun"  // one targeted rerun of the registered tests (the default)
	PolicyAccept Policy = "accept" // record the pass with the issue link and no rerun
	PolicyOff    Policy = "off"    // the rule is off: every failure blocks
)

// ParsePolicy reads a declared policy; "" is the default.
func ParsePolicy(s string) (Policy, error) {
	switch Policy(s) {
	case "":
		return PolicyRerun, nil
	case PolicyRerun, PolicyAccept, PolicyOff:
		return Policy(s), nil
	}
	return "", fmt.Errorf("known_flakes %q: want rerun, accept or off", s)
}

// MaxRerun bounds a targeted rerun: more registered failures in one attempt
// is not a flake pattern but something to investigate (tl-worker: a second
// unrelated flake on the same head means stop).
const MaxRerun = 5

// Kind is the gate's decision for one suite attempt on one head.
type Kind string

const (
	Passed            Kind = "passed"
	NeedRerun         Kind = "need-rerun"           // effect: rerun Known's tests, then classify again
	KnownFlakePassed  Kind = "known-flake-rerun"    // the targeted rerun passed: landing evidence
	KnownFlakeAccept  Kind = "known-flake-accepted" // policy accept: landing evidence without a rerun
	KnownFlakeTwice   Kind = "known-flake-failed-twice"
	Unknown           Kind = "unknown-failure"
	Unattributable    Kind = "unattributable-failure"
	PersistentMain    Kind = "persistent-main-failure"
	RegistryUnusable  Kind = "known-flake-rule-unavailable"
	RuleOff           Kind = "failed" // policy off: the failure blocks as it always did
	TooManyKnownFlaky Kind = "too-many-known-flakes"
	RerunInconclusive Kind = "known-flake-rerun-inconclusive" // timed out, or a named test did not run
)

// Lands reports whether an outcome is passing landing evidence.
func (k Kind) Lands() bool { return k == Passed || k == KnownFlakePassed || k == KnownFlakeAccept }

// Result is one run's outcome as the classifier reads it.
type Result struct {
	Passed       bool
	Failures     []string // exact "<pkg>.<Test>" names
	Unattributed []string // why some failure has no test name; empty when every failure is named
	// Subtests are the failing subtests the log names (#466), as
	// "<pkg>.<Test>/<sub>…"; a failing top-level test is explained by its
	// failing leaves (Leaves), each covered by a registered name.
	Subtests []string
}

// explain splits a run's failures into the registered names that cover them
// (known, unique) and the leaves nothing covers (unknown): a failing
// top-level test is judged by its failing leaf subtests, so a registered
// subtest explains its own failure and never its sibling's.
func explain(r Result, reg Registry) (known []Entry, unknown []string) {
	for _, f := range r.Failures {
		for _, leaf := range Leaves(f, r.Subtests) {
			e, ok := reg.Cover(leaf)
			if !ok {
				unknown = append(unknown, leaf)
				continue
			}
			if !slices.ContainsFunc(known, func(k Entry) bool { return k.Test == e.Test }) {
				known = append(known, e)
			}
		}
	}
	return known, unknown
}

// topOf is a failing name's top-level test.
func topOf(name string) string {
	if n, ok := ParseName(name); ok {
		return n.Top()
	}
	return name
}

// Input is everything the gate decides from. Rerun is nil until the targeted
// rerun has run; MainBaseline is nil unless merge ran its diagnostic baseline.
type Input struct {
	Attempt      Result
	Registry     Registry
	RegistryErr  error // the registry could not be read: never an empty registry
	Policy       Policy
	Rerun        *Result
	MainBaseline *Result
}

// Outcome is the decision: a kind, the registered failures (Known) and the
// rest (Unknown), and a one-line reason for the status and the error.
type Outcome struct {
	Kind    Kind
	Known   []Entry
	Unknown []string
	Reason  string
}

// Issues lists the flake issues an outcome names, sorted and unique.
func (o Outcome) Issues() []string {
	var ids []string
	for _, e := range o.Known {
		if !slices.Contains(ids, e.Issue) {
			ids = append(ids, e.Issue)
		}
	}
	sort.Strings(ids)
	return ids
}

// Classify is the gate's whole decision (#394), pure. A passing attempt
// passes. A failure lands only when every failure is a named, registered
// test and either the policy accepts it or one targeted rerun of exactly
// those tests passes. Anything unnamed, unregistered, or unreadable blocks.
func Classify(in Input) Outcome {
	a := in.Attempt
	if a.Passed {
		return Outcome{Kind: Passed}
	}
	if in.Policy == PolicyOff {
		return Outcome{Kind: RuleOff, Unknown: a.Failures, Reason: "the known-flake rule is off"}
	}
	if len(a.Unattributed) > 0 || len(a.Failures) == 0 {
		why := strings.Join(a.Unattributed, "; ")
		if why == "" {
			why = "the suite failed without naming a test"
		}
		return Outcome{Kind: Unattributable, Unknown: a.Failures, Reason: why}
	}
	if in.RegistryErr != nil {
		return Outcome{Kind: RegistryUnusable, Unknown: a.Failures, Reason: in.RegistryErr.Error()}
	}
	known, unknown := explain(a, in.Registry)
	if len(unknown) > 0 {
		o := Outcome{Kind: Unknown, Known: known, Unknown: unknown, Reason: "unknown failure: " + strings.Join(unknown, ", ")}
		if in.MainBaseline != nil && !in.MainBaseline.Passed {
			var both []string
			for _, u := range unknown {
				if slices.Contains(in.MainBaseline.Failures, topOf(u)) {
					both = append(both, u)
				}
			}
			if len(both) > 0 {
				o.Kind, o.Reason = PersistentMain, "persistent main failure: "+strings.Join(both, ", ")
			}
		}
		return o
	}
	if len(known) > MaxRerun {
		return Outcome{Kind: TooManyKnownFlaky, Known: known, Reason: fmt.Sprintf("%d registered tests failed in one attempt (more than %d): investigate", len(known), MaxRerun)}
	}
	if in.Policy == PolicyAccept {
		return Outcome{Kind: KnownFlakeAccept, Known: known, Reason: "known flake accepted by policy"}
	}
	if in.Rerun == nil {
		return Outcome{Kind: NeedRerun, Known: known, Reason: "known flake: rerunning " + strings.Join(tests(known), ", ")}
	}
	if len(in.Rerun.Unattributed) > 0 {
		return Outcome{Kind: RerunInconclusive, Known: known, Reason: "the known-flake rerun proved nothing: " + strings.Join(in.Rerun.Unattributed, "; ")}
	}
	if in.Rerun.Passed {
		return Outcome{Kind: KnownFlakePassed, Known: known, Reason: "known flake: the rerun passed"}
	}
	// The rerun runs whole top-level tests (RerunTests): a sibling that
	// failed there is no known flake, and says so.
	if _, rerunUnknown := explain(*in.Rerun, in.Registry); len(rerunUnknown) > 0 {
		return Outcome{Kind: Unknown, Known: known, Unknown: rerunUnknown, Reason: "the known-flake rerun failed elsewhere: " + strings.Join(rerunUnknown, ", ")}
	}
	return Outcome{Kind: KnownFlakeTwice, Known: known, Unknown: in.Rerun.Failures, Reason: "known flake failed twice on this head: " + strings.Join(tests(known), ", ")}
}

// Tests lists the tests a registry entry set names, in order.
func Tests(es []Entry) []string { return tests(es) }

func tests(es []Entry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Test
	}
	return out
}

// Revalidate rechecks a known-flake landing just before it is posted, or
// reposted from a saved attempt (#394 M2 review): the policy must still allow
// it, and every test it relied on must still be registered to an open issue.
// A rerun can take minutes, and a saved attempt can be days old, so the flake
// issue may have closed or the operator may have turned the rule off. Any
// other outcome passes through unchanged.
func Revalidate(o Outcome, r Registry, regErr error, policy Policy) Outcome {
	if o.Kind != KnownFlakePassed && o.Kind != KnownFlakeAccept {
		return o
	}
	switch {
	case policy == PolicyOff:
		return Outcome{Kind: RuleOff, Known: o.Known, Reason: "the known-flake rule is now off"}
	case o.Kind == KnownFlakeAccept && policy != PolicyAccept:
		return Outcome{Kind: Unknown, Known: o.Known, Reason: "known flakes are no longer accepted without a rerun (policy " + string(policy) + ")"}
	case regErr != nil:
		return Outcome{Kind: RegistryUnusable, Known: o.Known, Reason: regErr.Error()}
	}
	for _, e := range o.Known {
		if _, ok := r.Lookup(e.Test); !ok {
			return Outcome{Kind: Unknown, Known: o.Known, Unknown: []string{e.Test},
				Reason: e.Test + " is no longer a known flake (issue #" + e.Issue + " closed or dropped it)"}
		}
	}
	return o
}

// RerunMissing lists the tests a rerun log does not show passing. The rerun
// runs `go test -v`, so each named test or subtest prints
// "--- PASS: <Test>[/<sub>…] (…)"; a run that matched no test exits 0 without
// it, which proves nothing.
func RerunMissing(log string, tests []string) []string {
	var missing []string
	for _, t := range tests {
		n, ok := ParseName(t)
		if !ok || !strings.Contains(log, "--- PASS: "+n.Run()+" (") {
			missing = append(missing, t)
		}
	}
	return missing
}

// RerunTests are the whole top-level tests a known set reruns (#466): a
// registered subtest reruns its top-level test, siblings and all, so one go
// test invocation covers every entry and a sibling that fails there blocks.
func RerunTests(known []Entry) []string {
	var out []string
	for _, e := range known {
		if top := topOf(e.Test); !slices.Contains(out, top) {
			out = append(out, top)
		}
	}
	return out
}
