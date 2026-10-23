package main

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func benchScoringFixture(t *testing.T) (*benchStore, benchCouchDeps) {
	t.Helper()
	s, census := benchCleanupFixture(t, t.TempDir(), "run", time.Now().UTC().Add(-time.Hour))
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	r.Events = r.Events[:4] // Synthetic fixture: keep only prepare and dispatch.
	if err := s.write(r); err != nil {
		t.Fatal(err)
	}
	_, state, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	h, w := state.Bindings()
	c := strings.Repeat("a", 40)
	facts := bench.WouldLandFacts{Candidate: c, ObservedHead: c, PushedHead: c, PrivateOrigin: "file:" + filepath.Join(s.root, "public", "origin.git"), SuiteHead: c, SuiteReceiptID: "synthetic-suite", SuiteLogDigest: "synthetic-log", SuitePassed: true, FullSuite: true, CompletionHead: c, ReviewedHead: c, CompletionEvidence: "synthetic-close", CompletionToken: "synthetic-token", ReviewID: "synthetic-review", ReviewVerdict: "SHIP", CompletionBound: true}
	facts.ExpectedOrigin = facts.PrivateOrigin
	for _, e := range []bench.Event{{ID: "proposal", Kind: bench.ProposeWouldLand, OperationID: "proof"}, {ID: "accepted", Kind: bench.ProofAccepted, OperationID: "proof", Facts: facts}, {ID: "stop", Kind: bench.Stop, OperationID: "cancel"}, {ID: "retired", Kind: bench.CancelConfirmed, OperationID: "cancel"}} {
		e.RunID, e.Hub, e.Worker = r.ID, h, w
		if _, _, err := s.apply(e); err != nil {
			t.Fatal(err)
		}
	}
	return s, census.deps()
}

func TestBenchScoringKeepsUnknownChildCustody(t *testing.T) {
	s, census := benchScoringFixture(t)
	calls := 0
	d := benchScoringDeps{Couch: census, Measure: func(context.Context, *benchStore, benchCouchDeps) (bench.Measurements, error) {
		return bench.Measurements{Validity: "unverified"}, nil
	}, Evaluate: func(context.Context, bench.Task, string, string, string) (bench.Acceptance, error) {
		calls++
		_, state, err := s.read()
		if err != nil || state.Phase() != bench.PhaseScoring {
			t.Fatal("execution without persisted authorization", err)
		}
		return bench.Acceptance{Status: "unavailable", Reason: "child retirement unknown"}, &runnerCleanupUnconfirmed{Detail: "synthetic scorer still owned"}
	}}
	if err := scoreBenchRunWith(context.Background(), s, d); err == nil {
		t.Fatal("unknown custody reported complete")
	}
	r, state, err := s.read()
	if err != nil || state.Phase() != bench.PhaseScoring || r.Data["acceptance"] != nil {
		t.Fatal("unknown cleanup became terminal", state.Phase(), err)
	}
	if err := scoreBenchRunWith(context.Background(), s, d); err == nil || calls != 1 {
		t.Fatal("unknown scorer rerun", calls, err)
	}
	if err := cleanBenchRun(context.Background(), s, census); err == nil {
		t.Fatal("unknown scorer cleaned")
	}
}

func TestBenchScoringInvalidProfileSkipsHiddenExecution(t *testing.T) {
	s, census := benchScoringFixture(t)
	d := benchScoringDeps{Couch: census, Measure: func(context.Context, *benchStore, benchCouchDeps) (bench.Measurements, error) {
		return bench.Measurements{Validity: "invalid", ValidityIssues: []string{"worker: actual model mismatched"}}, nil
	}, Evaluate: func(context.Context, bench.Task, string, string, string) (bench.Acceptance, error) {
		t.Fatal("invalid profile executed hidden acceptance")
		return bench.Acceptance{}, errors.New("unreachable")
	}}
	if err := scoreBenchRunWith(context.Background(), s, d); err != nil {
		t.Fatal(err)
	}
	r, state, err := s.read()
	if err != nil || state.Phase() != bench.PhaseAcceptanceUnavailable || r.Data["acceptance"] == nil {
		t.Fatal("invalid outcome missing", state.Phase(), err)
	}
}
