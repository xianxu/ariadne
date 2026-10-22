package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// Local absence is independent of the aggregate server verdict: a waiting
// car needs the action even while GitHub is pending, red, or not registered.
func TestDecideCIGateMissingLocalEvidence(t *testing.T) {
	decl := localCheckDecl{Version: 1, Checks: []localCheck{{Name: "suite", Required: true}}}
	for _, tc := range []struct {
		name      string
		checks    []prCheck
		workflows bool
		reason    ciReason
	}{
		{"green", []prCheck{{Name: "build", Bucket: "pass"}}, true, ciLocalNeeded},
		{"pending", []prCheck{{Name: "build", Bucket: "pending"}}, true, ciPending},
		{"red", []prCheck{{Name: "build", Bucket: "fail"}}, true, ciRed},
		{"unknown", []prCheck{{Name: "build", Bucket: "new-state"}}, true, ciUnknown},
		{"not registered", nil, true, ciNotRegistered},
		{"no workflows", nil, false, ciLocalNeeded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := decideCIGate(ciInputs{Checks: tc.checks, HeadHasWorkflows: tc.workflows, Local: decl})
			if v.Reason != tc.reason || !reflect.DeepEqual(v.LocalNeeded, []string{"suite"}) {
				t.Fatalf("reason=%v needed=%v; want reason=%v needed=[suite]", v.Reason, v.LocalNeeded, tc.reason)
			}
			verdict := missingLocalGateError(t, v.refusal(7), tc.reason, []string{"suite"})
			// Errors own their metadata; mutating a reusable verdict must not
			// change the action already handed to its consumer.
			v.LocalNeeded[0] = "changed"
			missingLocalGateError(t, verdict, tc.reason, []string{"suite"})
		})
	}
}

func missingLocalGateError(t *testing.T, err error, reason ciReason, want []string) *ciGateError {
	t.Helper()
	var verdict *ciGateError
	if !errors.As(fmt.Errorf("wrapped: %w", err), &verdict) || verdict.Reason != reason {
		t.Fatalf("error=%v; want typed reason %v", err, reason)
	}
	if !reflect.DeepEqual(verdict.LocalNeeded, want) {
		t.Fatalf("error=%v lacks missing-local evidence %v", err, want)
	}
	return verdict
}

// The stateful fake keeps status records on their exact PR head. A passing
// record appearing between two observations clears only the local absence;
// it does not change the independently failing/pending server check.
func TestRunCIGateMissingLocalEvidence(t *testing.T) {
	ctx := context.Background()
	const head = "550head"
	decl := localCheckDecl{Version: 1, Checks: []localCheck{{Name: "suite", Required: true}}}
	for _, tc := range []struct {
		name            string
		checks          []prCheck
		workflows       bool
		required        []string
		missingReason   ciReason
		satisfiedReason ciReason
	}{
		{"green", []prCheck{{Name: "build", Bucket: "pass"}}, true, nil, ciLocalNeeded, ciGreen},
		{"pending", []prCheck{{Name: "build", Bucket: "pending"}}, true, nil, ciPending, ciPending},
		{"red", []prCheck{{Name: "build", Bucket: "fail"}}, true, nil, ciRed, ciRed},
		{"unknown", []prCheck{{Name: "build", Bucket: "new-state"}}, true, nil, ciUnknown, ciUnknown},
		{"not registered", nil, true, nil, ciNotRegistered, ciGreen},
		{"no workflows", nil, false, nil, ciLocalNeeded, ciGreen},
		{"guard missing", []prCheck{{Name: "build", Bucket: "pass"}}, true, []string{localGatesVerdict}, ciMissing, ciMissing},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gh := newCIFakeGH()
			gh.setHead(7, head, tc.checks...)
			facts := &ciGateFacts{decl: decl, hasWorkflows: tc.workflows, required: tc.required}
			git := func(args ...string) (string, error) {
				switch args[0] {
				case "merge-base":
					return "main", nil
				case "rev-list":
					return head, nil // root: no inheritable record
				default:
					return "", fmt.Errorf("unexpected git %v", args)
				}
			}
			observe := func() error {
				return runPreparedCIGate(ctx, gh, "o/r", 7, "origin/main", head, git, nil, false, io.Discard, facts)
			}
			missingLocalGateError(t, observe(), tc.missingReason, []string{"suite"})
			if err := gh.PostStatus(ctx, "o/r", head, localStatusContext("suite"), "success", "passed"); err != nil {
				t.Fatal(err)
			}
			if err := observe(); tc.satisfiedReason.passes() {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				missingLocalGateError(t, err, tc.satisfiedReason, nil)
			}
		})
	}
}

func TestRunCIGateMissingLocalNeedsConfirmedHead(t *testing.T) {
	for _, mode := range []string{"head lag", "unread checks", "unread records"} {
		t.Run(mode, func(t *testing.T) {
			gh := newCIFakeGH()
			gh.setHead(7, "head", prCheck{Name: "build", Bucket: "pass"})
			caller := ghCaller(gh)
			git := func(...string) (string, error) { return "", errors.New("records unavailable") }
			switch mode {
			case "head lag":
				gh.setHead(7, "other", prCheck{Name: "build", Bucket: "pass"})
			case "unread checks":
				caller = &missingLocalUnreadGH{ciFakeGH: gh}
			}
			facts := &ciGateFacts{decl: localCheckDecl{Version: 1, Checks: []localCheck{{Name: "suite", Required: true}}}, hasWorkflows: true}
			err := runPreparedCIGate(context.Background(), caller, "o/r", 7, "origin/main", "head", git, nil, false, io.Discard, facts)
			if err == nil {
				t.Fatal("unconfirmed reads passed")
			}
			var verdict *ciGateError
			if errors.As(err, &verdict) {
				missingLocalGateError(t, err, ciHeadLag, nil)
			} else if !strings.Contains(err.Error(), "could not read") {
				t.Fatal(err)
			}
		})
	}
}

type missingLocalUnreadGH struct{ *ciFakeGH }

func (g *missingLocalUnreadGH) PRChecks(context.Context, string, int) ([]prCheck, error) {
	return nil, errors.New("checks unavailable")
}
