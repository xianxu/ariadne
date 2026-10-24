package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

func stagingProofFixture() gatestate.MeasurementInputs {
	m := restingProofFixture()
	m.Since = "2026-10-24T02:55:02Z"
	m.Until = "2026-10-24T03:07:54Z"
	split := "2026-10-24T02:57:54Z"
	m.Engagements[0].Start, m.Engagements[0].Began, m.Engagements[0].End = m.Since, m.Since, m.Until
	m.Sources[0].Started = m.Since
	m.Sources[0].Windows[0].Start, m.Sources[0].Windows[0].End = split, m.Until
	m.Sources[0].DispatchWindows[0].Start, m.Sources[0].DispatchWindows[0].End = m.Since, m.Until
	w := m.Branches[0].Intervals
	w[0].Start, w[0].End, w[0].Branch, w[0].Kind = m.Since, split, "stage/000849-000850", "stage-excluded"
	w[1].Start, w[1].End, w[1].Branch = split, m.Until, "000849-cast"
	return m
}

func TestStageExcludedSavedProofAndVisibleMinutes(t *testing.T) {
	m := stagingProofFixture()
	if err := validateOwnedBranchProof(m, "849"); err != nil {
		t.Fatal(err)
	}
	output := strings.Join(measurementBranchWarnings(m), " ")
	for _, want := range []string{"excluded and unassigned 2.87 branch-window minutes", "stage/000849-000850", "codex/main", "not focused activity minutes"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in %s", want, output)
		}
	}
	for _, mutation := range []string{"credited", "own", "foreign", "unknown", "resting", "kind"} {
		t.Run(mutation, func(t *testing.T) {
			bad := stagingProofFixture()
			w := &bad.Branches[0].Intervals[0]
			switch mutation {
			case "credited":
				bad.Sources[0].Windows[0].Start = bad.Since
			case "own":
				w.Branch = "000849-cast"
			case "foreign":
				w.Branch = "000850-cast"
			case "unknown":
				w.Branch = "feature"
			case "resting":
				w.Kind = "resting-excluded"
			case "kind":
				w.Kind = "branch"
			}
			if err := validateOwnedBranchProof(bad, "849"); err == nil {
				t.Fatal("tampered stage proof accepted")
			}
		})
	}
}
