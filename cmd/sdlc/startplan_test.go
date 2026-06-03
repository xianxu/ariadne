package main

import (
	"strings"
	"testing"
)

func TestStartPlanCmd_Registered(t *testing.T) {
	cmd := NewStartPlanCmd()
	if cmd.Flags().Lookup("issue") == nil {
		t.Error("start-plan command missing --issue flag")
	}
}

// #75 M2: start-plan delivers the at-plan architecture lens (the forward
// injection) to the main thread, labeled with the issue.
func TestRunStartPlan_RendersAtPlanLens(t *testing.T) {
	var b strings.Builder
	runStartPlan(&b, 75)
	out := b.String()
	for _, want := range []string{"#75", "ARCH-DRY", "at-plan", "change-code"} {
		if !strings.Contains(out, want) {
			t.Errorf("start-plan output missing %q:\n%s", want, out)
		}
	}
	// No --issue → generic label, still renders the principles.
	var b2 strings.Builder
	runStartPlan(&b2, 0)
	if !strings.Contains(b2.String(), "ARCH-PURE") {
		t.Error("start-plan with no issue should still render the principles")
	}
}
