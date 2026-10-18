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
)

// Lands reports whether an outcome is passing landing evidence.
func (k Kind) Lands() bool { return k == Passed || k == KnownFlakePassed || k == KnownFlakeAccept }

// Result is one run's outcome as the classifier reads it.
type Result struct {
	Passed       bool
	Failures     []string // exact "<pkg>.<Test>" names
	Unattributed []string // why some failure has no test name; empty when every failure is named
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
	var known []Entry
	var unknown []string
	for _, f := range a.Failures {
		if id, ok := in.Registry.Lookup(f); ok {
			known = append(known, Entry{Test: f, Issue: id})
		} else {
			unknown = append(unknown, f)
		}
	}
	if len(unknown) > 0 {
		o := Outcome{Kind: Unknown, Known: known, Unknown: unknown, Reason: "unknown failure: " + strings.Join(unknown, ", ")}
		if in.MainBaseline != nil && !in.MainBaseline.Passed {
			var both []string
			for _, u := range unknown {
				if slices.Contains(in.MainBaseline.Failures, u) {
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
	if in.Rerun.Passed {
		return Outcome{Kind: KnownFlakePassed, Known: known, Reason: "known flake: the rerun passed"}
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
