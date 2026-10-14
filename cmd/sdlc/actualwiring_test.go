package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

func measuredInputsFixture() *gatestate.MeasurementInputs {
	return &gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodLegacyV1,
		Repository: "github.com/example/repo", CodeHead: strings.Repeat("a", 40),
		Since: "2026-10-13T18:00:00Z", Until: "2026-10-13T19:00:00Z", GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true}
}

func TestActualAdoptionRetainsInputsAndExplicitBrain(t *testing.T) {
	original := computeActualForCloseFn
	t.Cleanup(func() { computeActualForCloseFn = original })
	inputs := measuredInputsFixture()
	computeActualForCloseFn = func(_ context.Context, _, _, _ string, brain ...string) actualResult {
		if len(brain) != 1 || brain[0] != "/tmp/explicit-brain" {
			t.Fatalf("brain override lost: %v", brain)
		}
		return actualResult{Status: actualMeasured, Hours: 1, Inputs: inputs}
	}
	f := &closeFlags{Issue: 335, BrainDir: "/tmp/explicit-brain"}
	if _, ok := adoptOmittedActual(&bytes.Buffer{}, f, "335", "issue"); !ok || f.Inputs != inputs {
		t.Fatalf("admitted evidence lost: %+v", f)
	}
}

func TestBoundaryRetainsActualInputs(t *testing.T) {
	issues := closeRepo(t, 69)
	plans := t.TempDir()
	inputs := measuredInputsFixture()
	var stderr bytes.Buffer
	p := boundaryReviewParams{IssuesDir: issues, IssueNum: 69, PlansDir: plans, ActualInputs: inputs}
	d := persistBoundaryRound(&stderr, p, reviewResult{Agent: "codex", Head: inputs.CodeHead, Round: &gatestate.RoundReport{}}, "2026-10-13T20:00:00Z")
	if d.Block {
		t.Fatalf("persist refused: %s", stderr.String())
	}
	l, err := readBoundaryGateLedger(plans, filepath.Base(mustIssuePath(t, issues, 69)), 69)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Rounds) != 1 || l.Rounds[0].ActualInputs == nil || l.Rounds[0].ActualInputs.CodeHead != inputs.CodeHead {
		t.Fatalf("measurement inputs not retained: %+v", l.Rounds)
	}
	// Malformed authority must refuse before writing a seemingly complete round.
	inputs.CodeHead = "bad"
	if d := persistBoundaryRound(&stderr, p, reviewResult{Round: &gatestate.RoundReport{}}, "2026-10-13T20:01:00Z"); !d.Block {
		t.Fatal("malformed measurement accepted")
	}
}

func TestActualEstimateWarning(t *testing.T) {
	for _, tc := range []struct {
		actual, estimate float64
		warn             bool
	}{{.05, 4.69, true}, {23.76, 1, true}, {.04, .3, false}, {2, 4, false}, {1, 0, false}} {
		if got := actualEstimateWarning(tc.actual, tc.estimate); (got != "") != tc.warn {
			t.Fatalf("actual=%v estimate=%v warning=%q", tc.actual, tc.estimate, got)
		}
	}
}

func TestActualExclusionDiagnosticsAreBounded(t *testing.T) {
	warnings := []string{"historical authority reconstructed"}
	for i := 0; i < 1000; i++ {
		warnings = append(warnings, "excluded codex session "+strings.Repeat("x", 40)+": no verified ownership provenance")
	}
	got := actualExclusionWarnings(warnings)
	if len(got) != 2 || !strings.Contains(got[1], "1000 sessions") || len(got[1]) > 600 || got[0] != warnings[0] {
		t.Fatalf("unbounded or lost diagnostic: %v", got)
	}
}
