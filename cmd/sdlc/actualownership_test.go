package main

import (
	"bytes"
	"errors"
	"fmt"
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
