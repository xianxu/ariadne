package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestActualOwnershipRecoveryRendering(t *testing.T) {
	err := fmt.Errorf("measurement: %w", &unownedActualError{Issue: "5"})
	res := actualErrorResult("5", err)
	for _, close := range []bool{false, true} {
		var out bytes.Buffer
		if close {
			explainActual(&out, "5", "issue", "", res)
		} else {
			printActual(&out, res, "")
		}
		for _, want := range []string{"no owner since cutover: run sdlc claim --issue 5 from the slot doing the work", "work between cutover and claim is unowned and not counted"} {
			if !strings.Contains(out.String(), want) {
				t.Fatalf("close=%v missing %q:\n%s", close, want, out.String())
			}
		}
		for _, wrong := range []string{"restore the named", "reconcile and record"} {
			if strings.Contains(out.String(), wrong) {
				t.Fatalf("close=%v gives wrong recovery:\n%s", close, out.String())
			}
		}
	}
}

func TestActualOwnershipUnrelatedEvidenceStillNeedsRestore(t *testing.T) {
	res := actualErrorResult("5", errors.New("tracker history unavailable"))
	var out bytes.Buffer
	printActual(&out, res, "")
	if !strings.Contains(out.String(), "restore the named measurement evidence") || strings.Contains(out.String(), "sdlc claim") {
		t.Fatal(out.String())
	}
}

// The captured refusal is evidence of the old presentation, not proof that an
// owner is absent. Only the pinned tracker facts may authorize claim recovery.
func TestActualOwnershipCapturedEchoRefusal(t *testing.T) {
	capture, err := os.ReadFile("testdata/actual/echo5-unowned-close.txt")
	if err != nil {
		t.Fatal(err)
	}
	text := string(capture)
	marker := "can't auto-compute ("
	_, diagnostic, ok := strings.Cut(text, marker)
	if !ok {
		t.Fatal("capture lost measurement diagnostic")
	}
	diagnostic, _, ok = strings.Cut(diagnostic, ").\n")
	if !ok {
		t.Fatal("capture lost diagnostic boundary")
	}
	unproven := actualErrorResult("5", errors.New(diagnostic))
	var before bytes.Buffer
	explainActual(&before, "5", "issue", "", unproven)
	if unproven.Recovery != "" || !strings.Contains(before.String(), "restore the named measurement evidence") {
		t.Fatalf("captured prose invented ownership authority: %s", before.String())
	}
	// Real migration/claim fixtures establish the facts that select this error.
	proven := actualErrorResult("5", &unownedActualError{Issue: "5"})
	var after bytes.Buffer
	explainActual(&after, "5", "issue", "", proven)
	for _, old := range []string{"reconcile and record", "restore the named measurement evidence", "recover the evidence below"} {
		if !strings.Contains(text, old) || strings.Contains(after.String(), old) {
			t.Fatalf("captured bad guidance %q was lost or survived recovery: %s", old, after.String())
		}
	}
	if !strings.Contains(after.String(), "run sdlc claim --issue 5 from the slot doing the work") || !strings.Contains(after.String(), "unowned and not counted") {
		t.Fatal(after.String())
	}
}
