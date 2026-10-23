package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

type benchScoringDeps struct {
	Couch    benchCouchDeps
	Measure  func(context.Context, *benchStore, benchCouchDeps) (bench.Measurements, error)
	Evaluate func(context.Context, bench.Task, string, string, string) (bench.Acceptance, error)
}

func scoreBenchRun(ctx context.Context, s *benchStore, d benchCouchDeps) error {
	return scoreBenchRunWith(ctx, s, benchScoringDeps{Couch: d})
}

func scoreBenchRunWith(ctx context.Context, s *benchStore, d benchScoringDeps) error {
	release, err := s.acquireRun()
	if err != nil {
		return err
	}
	defer release()
	r, state, err := s.read()
	if err != nil {
		return err
	}
	if err := verifyBenchRetirement(ctx, s, d.Couch); err != nil {
		return err
	}
	if state.Phase() == bench.PhaseScoring {
		// A lost result may still own a scorer. Retirement of the agents does
		// not authorize a second scorer or deletion of its locked worktree.
		return errors.New("bench scoring outcome unconfirmed; inspect retained scorer logs and native worktree leases before any retry or cleanup")
	}
	if state.Phase() == bench.PhaseScored || state.Phase() == bench.PhaseAcceptanceUnavailable || state.Phase() == bench.PhaseStopped {
		if raw, ok := r.Data["acceptance"]; ok {
			var a bench.Acceptance
			return json.Unmarshal(raw, &a)
		}
	}
	if state.Phase() != bench.PhaseWouldLand && state.Phase() != bench.PhaseStopped {
		return fmt.Errorf("bench score needs a retired would-land or stopped run; retained phase %s", state.Phase())
	}
	measure := d.Measure
	if measure == nil {
		measure = measureBenchRun
	}
	m, err := measure(ctx, s, d.Couch)
	if err != nil {
		return err
	}
	a := bench.Acceptance{Status: "unavailable", Reason: "run stopped before an admitted would-land candidate"}
	if state.Phase() == bench.PhaseStopped {
		return s.putData("acceptance", a)
	}
	h, w := state.Bindings()
	op := "score-" + r.ID
	_, effects, err := s.apply(bench.Event{ID: op + "-begin", RunID: r.ID, Kind: bench.BeginScore, OperationID: op, Hub: h, Worker: w})
	if err != nil {
		return err
	}
	if len(effects) != 1 || effects[0].Kind != bench.Score {
		return errors.New("bench score effect not authorized")
	}
	var scoreErr error
	if m.Validity == "invalid" {
		a.Reason = "run invalid: actual transcript profile differs from trusted provisioning; see private collection evidence"
	} else {
		evaluate := d.Evaluate
		if evaluate == nil {
			evaluate = evaluateBenchAcceptance
		}
		a, scoreErr = evaluate(ctx, r.Task, filepath.Join(r.Root, "public", "work"), state.Candidate(), filepath.Join(r.Root, "observer"))
		var unknown *runnerCleanupUnconfirmed
		if errors.As(scoreErr, &unknown) {
			return scoreErr
		}
	}
	if err := bench.ValidateAcceptance(a); err != nil {
		return errors.Join(scoreErr, err)
	}
	if err := s.putData("acceptance", a); err != nil {
		return errors.Join(scoreErr, err)
	}
	kind := bench.ScoreCompleted
	if a.Status == "unavailable" {
		kind = bench.ScoreUnavailable
	}
	_, _, err = s.apply(bench.Event{ID: op + "-" + string(kind), RunID: r.ID, Kind: kind, OperationID: op, Hub: h, Worker: w, Reason: a.Reason})
	return errors.Join(scoreErr, err)
}
