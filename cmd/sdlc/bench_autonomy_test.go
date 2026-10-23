package main

import (
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func TestBenchAutonomyNeedsTrustedExactInterval(t *testing.T) {
	runtime := benchRuntimeState{Since: "2026-10-23T20:00:00Z"}
	for _, kind := range []string{"missing", "other-start", "other-end", "operator-input", "no-evidence", "exact"} {
		t.Run(kind, func(t *testing.T) {
			proof := benchRetirement{ConfirmedAt: "2026-10-23T21:00:00Z", Autonomy: &benchAutonomyAttestation{Since: runtime.Since, Until: "2026-10-23T21:00:00Z", NoOperatorInput: true, Evidence: "trusted operator independently confirms no input throughout this exact interval"}}
			switch kind {
			case "missing":
				proof.Autonomy = nil
			case "other-start":
				proof.Autonomy.Since = "2026-10-23T20:01:00Z"
			case "other-end":
				proof.Autonomy.Until = "2026-10-23T20:59:00Z"
			case "operator-input":
				proof.Autonomy.NoOperatorInput = false
			case "no-evidence":
				proof.Autonomy.Evidence = ""
			}
			m := bench.Measurements{Validity: "valid"}
			applyBenchAutonomyMeasurement(runtime, proof, &m)
			if m.Autonomous != (kind == "exact") || m.OperatorWaitNanos != 0 {
				t.Fatal("guessed autonomy or wait", m)
			}
			if kind != "exact" && (m.Validity != "unverified" || len(m.ValidityIssues) == 0) {
				t.Fatal("missing unknown diagnostic", m)
			}
		})
	}
}
