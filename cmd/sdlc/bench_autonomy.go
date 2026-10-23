package main

import (
	"strings"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

// M1 trusts the same explicit operator checklist that proves retirement. The
// native Couch API exposes exact known message receipts, but no operator-input
// history query; absence of observed messages is never an autonomy proof.
func applyBenchAutonomyMeasurement(runtime benchRuntimeState, proof benchRetirement, m *bench.Measurements) {
	m.Autonomous = false
	m.OperatorWaitNanos = 0
	a := proof.Autonomy
	if a != nil && a.Since == runtime.Since && a.Until == proof.ConfirmedAt && a.NoOperatorInput && strings.TrimSpace(a.Evidence) != "" && len(a.Evidence) <= 4096 && utf8.ValidString(a.Evidence) && !strings.ContainsRune(a.Evidence, 0) {
		m.Autonomous = true
		return
	}
	benchMetricUnknown(m, "autonomy", "no trusted no-operator-input attestation for the exact run interval; operator-wait coverage unverified")
}
