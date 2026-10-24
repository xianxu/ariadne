package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

type benchScoringDeps struct {
	SuiteHelper benchSuiteHelper
	Evaluate    func(context.Context, bench.Task, string, string, string) (bench.Acceptance, error)
}

// The Metis step is the execution owner. Its exclusive intent fences all hidden
// evaluation, including a child whose result was lost. There is no run replay.
func scoreBenchEvidence(ctx context.Context, task bench.Task, e benchSessionEvidence, workRoot, output string, d benchScoringDeps) (bench.Acceptance, error) {
	a := bench.Acceptance{Status: "unavailable", Reason: "session stopped before an admitted candidate"}
	if err := writeBenchJSONExclusive(filepath.Join(output, "score-intent.json"), map[string]any{"version": 1, "metis_run": e.MetisRun, "candidate": e.Candidate, "patch_sha256": task.Acceptance.PatchSHA256}); err != nil {
		return a, fmt.Errorf("hidden evaluation already intended or output unavailable; reconcile scorer custody, then use a fresh Metis repeat: %w", err)
	}
	if e.Measurements.Validity == "invalid" {
		a.Reason = "invalid session profile; hidden evaluation not executed"
	} else if e.Candidate != "" {
		patch, err := readBenchPatch(task.Acceptance.Patch)
		if err != nil {
			return a, err
		}
		digest := sha256.Sum256(patch)
		if hex.EncodeToString(digest[:]) != task.Acceptance.PatchSHA256 {
			return a, errors.New("hidden patch changed before capture")
		}
		if err := writeBenchExclusive(filepath.Join(output, "acceptance.patch"), patch, 0600); err != nil {
			return a, err
		}
		evaluate := d.Evaluate
		if evaluate == nil {
			evaluate = func(ctx context.Context, task bench.Task, work, candidate, output string) (bench.Acceptance, error) {
				return evaluateBenchAcceptanceWith(ctx, task, work, candidate, output, benchScoreDeps{SuiteHelper: d.SuiteHelper})
			}
		}
		var errEval error
		a, errEval = evaluate(ctx, task, workRoot, e.Candidate, output)
		var unconfirmed *runnerCleanupUnconfirmed
		if errors.As(errEval, &unconfirmed) {
			return a, errEval
		}
		if err := bench.ValidateAcceptance(a); err != nil {
			return a, errors.Join(errEval, err)
		}
		if err := writeBenchJSONExclusive(filepath.Join(output, "acceptance.json"), a); err != nil {
			return a, errors.Join(errEval, err)
		}
		// A confirmed unavailable result is a terminal finding. Keep private diagnostics
		// as an artifact, rather than asking Metis to launch another hidden evaluation.
		if errEval != nil {
			if err := writeBenchExclusive(filepath.Join(output, "finding.txt"), []byte(errEval.Error()), 0600); err != nil {
				return a, err
			}
		}
		return a, nil
	}
	if err := writeBenchJSONExclusive(filepath.Join(output, "acceptance.json"), a); err != nil {
		return a, err
	}
	return a, nil
}
