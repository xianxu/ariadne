package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// This envelope retains the final owner results, including protocol coverage,
// before publishing measurements. These immutable artifacts do not authorize
// resuming a crashed session.
type benchMeasurementEvidence struct {
	Version int                   `json:"version"`
	RunID   string                `json:"run_id"`
	Result  benchCollectionResult `json:"result"`
}

// measureBenchRun requires the caller's run owner lock. Retirement is checked
// even on a repeated in-process call; native sources are captured only once.
func measureBenchRun(ctx context.Context, s *benchSession, d benchCouchDeps) (bench.Measurements, error) {
	var empty bench.Measurements
	if err := verifyBenchRetirement(ctx, s, d); err != nil {
		return empty, err
	}
	r, state, err := s.read()
	if err != nil {
		return empty, err
	}
	if raw := r.Data["measurements"]; len(raw) != 0 {
		var saved bench.Measurements
		if err := decodeBenchRequired(raw, &saved, "gross_nanos", "autonomous", "worker", "tl", "reviewer", "leaks"); err != nil {
			return empty, err
		}
		return saved, nil
	}
	if raw := r.Data["collection"]; len(raw) != 0 {
		var saved benchMeasurementEvidence
		if err := decodeBenchRequired(raw, &saved, "version", "run_id", "result"); err != nil {
			return empty, err
		}
		if saved.Version != 1 || saved.RunID != r.ID || saved.Result.RunID != r.ID {
			return empty, errors.New("collection evidence belongs to another run or version")
		}
		digest := saved.Result.Digest
		if err := sealBenchCollection(&saved.Result); err != nil {
			return empty, err
		}
		if digest != saved.Result.Digest {
			return empty, errors.New("collection evidence digest differs")
		}
		if err := s.putData("measurements", saved.Result.Measurements); err != nil {
			return empty, err
		}
		return saved.Result.Measurements, nil
	}
	runtime, err := benchRetirementSubject(r, state)
	if err != nil {
		return empty, err
	}
	var proof benchRetirement
	if err := decodeBenchRequired(r.Data["retirement"], &proof, "version", "run_id", "confirmed_at", "reviewer_sources"); err != nil {
		return empty, err
	}
	since, err := time.Parse(time.RFC3339Nano, runtime.Since)
	if err != nil {
		return empty, err
	}
	until, err := time.Parse(time.RFC3339Nano, proof.ConfirmedAt)
	if err != nil || until.Before(since) || until.Sub(since) > 24*time.Hour {
		return empty, errors.New("measurement interval is invalid or exceeds 24 hours")
	}
	h, w := state.Bindings()
	undispatched := h == (bench.RuntimeIdentity{}) && w == (bench.RuntimeIdentity{}) && state.DispatchOperation() == ""
	var result benchCollectionResult
	if undispatched {
		unknown := bench.RoleUsage{Coverage: "unknown"}
		result = benchCollectionResult{RunID: r.ID, Valid: true, Captured: []benchCapturedSource{}, Measurements: bench.Measurements{Worker: unknown, TL: unknown, Reviewer: unknown, GrossNanos: until.Sub(since).Nanoseconds(), Leaks: bench.LeakReport{Status: "unknown", Coverage: bench.LeakCoverage{Basis: "no bound native runtime transcripts"}}}}
		benchMetricUnknown(&result.Measurements, "runtime", "run stopped before dispatch; usage, activity and protocol counts are unknown")
		if err := s.putData("captured_sources", result.Captured); err != nil {
			return empty, err
		}
	} else {
		inputs := make([]benchSourceInput, 0, 2+len(proof.ReviewerSources))
		for _, actor := range []struct {
			role  string
			actor benchRuntimeActor
		}{{"worker", runtime.Worker}, {"tl", runtime.Hub}} {
			if len(actor.actor.Transcripts) != 1 {
				return empty, fmt.Errorf("%s needs one unambiguous native-session transcript", actor.role)
			}
			inputs = append(inputs, benchSourceInput{Role: actor.role, Runtime: actor.actor.Identity, NativeSession: actor.actor.NativeSession, File: actor.actor.Transcripts[0]})
		}
		inputs = append(inputs, proof.ReviewerSources...)
		if len(inputs) > 128 {
			return empty, errors.New("measurement exceeds 128 bound sources")
		}
		var sources []benchCapturedSource
		if raw := r.Data["captured_sources"]; len(raw) != 0 {
			if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
				return empty, err
			}
			if err := json.Unmarshal(raw, &sources); err != nil {
				return empty, err
			}
			if len(sources) != len(inputs) {
				return empty, errors.New("capture pins differ from bound sources")
			}
			for i, source := range sources {
				bound := source.benchSourceInput
				bound.File = source.OriginalFile
				if bound != inputs[i] || source.File != "" && !benchPathWithin(filepath.Join(s.root, "observer", "captured-sources"), source.File) {
					return empty, errors.New("capture pins differ from exact source binding or observer storage")
				}
			}
		} else {
			sources, err = captureBenchSources(ctx, filepath.Join(s.root, "observer", "captured-sources"), runtime.Worker.Identity.Root, inputs)
			if err != nil {
				return empty, err
			}
			if err := s.putData("captured_sources", sources); err != nil {
				return empty, err
			}
		}
		focused, focusedErr := loadBenchFocusedInputs(ctx, s, sources)
		if err := ctx.Err(); err != nil {
			return empty, err
		}
		result, err = collectBenchMeasurements(ctx, benchCollectionInput{Task: r.Task, Run: state, Sources: sources, Since: since, Until: until, WorkerRepo: runtime.Worker.Identity.Root, Focused: focused})
		if err != nil {
			return empty, err
		}
		if focusedErr != nil {
			benchMetricUnknown(&result.Measurements, "focused-hours", "retained ownership inputs unavailable: "+focusedErr.Error())
		}
		if err := collectBenchProtocolMetrics(ctx, s, &result.Measurements); err != nil {
			return empty, err
		}
	}
	applyBenchAutonomyMeasurement(runtime, proof, &result.Measurements)
	// Return the same validated representation on repeated in-process calls.
	// The result schema's decoder also normalizes absent extension maps.
	raw, err := json.Marshal(result.Measurements)
	if err != nil {
		return empty, err
	}
	if err := json.Unmarshal(raw, &result.Measurements); err != nil {
		return empty, err
	}
	if err := sealBenchCollection(&result); err != nil {
		return empty, err
	}
	if err := s.putData("collection", benchMeasurementEvidence{Version: 1, RunID: r.ID, Result: result}); err != nil {
		return empty, err
	}
	if err := s.putData("measurements", result.Measurements); err != nil {
		return empty, err
	}
	return result.Measurements, nil
}

func sealBenchCollection(result *benchCollectionResult) error {
	result.Digest = ""
	raw, err := json.Marshal(result)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(raw)
	result.Digest = hex.EncodeToString(hash[:])
	return nil
}
