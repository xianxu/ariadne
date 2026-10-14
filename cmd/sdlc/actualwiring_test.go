package main

import (
	"bytes"
	"context"
	"io"
	"os"
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

// BR-3: exercise the production close callers rather than just the comparator.
// Removing computeClose's warning call must fail the whole-issue cases; a
// milestone's increment must never be compared with the whole issue's estimate.
func TestOmittedActualCloseEstimateWarningWiring(t *testing.T) {
	for _, tc := range []struct {
		name, milestone, diagnostic string
		hours                       float64
	}{
		{"implausibly low issue actual", "", "measured actual 0.05h is 0.05× estimate 1.00h", .05},
		{"implausibly high issue actual", "", "measured actual 23.76h is 23.76× estimate 1.00h", 23.76},
		{"milestone increment is not whole issue actual", "M1", "", .05},
	} {
		t.Run(tc.name, func(t *testing.T) {
			issuesDir := closeRepo(t, 335)
			plans := t.TempDir()
			t.Setenv("WF_CALIB_LEDGER", filepath.Join(t.TempDir(), "calibration.tsv"))
			original := computeActualForCloseFn
			t.Cleanup(func() { computeActualForCloseFn = original })
			measurements := 0
			computeActualForCloseFn = func(_ context.Context, issue, milestone, _ string, _ ...string) actualResult {
				measurements++
				if issue != "335" || milestone != tc.milestone {
					t.Fatalf("measurement request issue=%q milestone=%q", issue, milestone)
				}
				return actualResult{Status: actualMeasured, Issue: issue, Hours: tc.hours, Window: "fixture window"}
			}
			calls, _ := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nFine.\n"+emptyBoundaryFindings)
			var stderr bytes.Buffer
			var err error
			if tc.milestone == "" {
				err = runCloseWithReview(io.Discard, &stderr, &closeFlags{
					Issue: 335, Verified: "fixture behavior verified", NoAtlas: true, NoProject: true,
					IssuesDir: issuesDir, PlansDir: plans, BrainDir: "../nonexistent-brain",
				})
			} else {
				err = runMilestoneClose(io.Discard, &stderr, &milestoneCloseFlags{
					Issue: 335, Milestone: tc.milestone, Verified: "fixture increment verified", NoAtlas: true, NoProject: true,
					IssuesDir: issuesDir, PlansDir: plans, BrainDir: "../nonexistent-brain",
				})
			}
			if err != nil {
				t.Fatalf("close: %v\n%s", err, stderr.String())
			}
			if measurements != 1 || *calls != 1 {
				t.Fatalf("measurements=%d reviews=%d; want one measurement and review", measurements, *calls)
			}
			output := stderr.String()
			if !strings.Contains(output, "using measured actual:") {
				t.Fatalf("omitted actual was not adopted:\n%s", output)
			}
			if tc.diagnostic != "" {
				if !strings.Contains(output, tc.diagnostic) || !strings.Contains(output, "inspect ownership/source provenance before adopting this calibration sample") {
					t.Fatalf("missing estimate plausibility diagnostic:\n%s", output)
				}
				data, err := os.ReadFile(mustIssuePath(t, issuesDir, 335))
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(data), "status: codecomplete") {
					t.Fatalf("warning prevented the reviewed issue close: %s", data)
				}
			} else if strings.Contains(output, "inspect ownership/source provenance") {
				t.Fatalf("milestone increment compared with the whole issue estimate:\n%s", output)
			}
		})
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
