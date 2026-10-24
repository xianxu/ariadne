package bench

import "time"

// ProjectBenchMetrics supplies only evidenced numeric observations to Metis.
// The accompanying artifacts retain unknown coverage, exact credits, and cache
// counters. A successful command is not evidence of hidden per-case pass rate.
func ProjectBenchMetrics(m Measurements, a *Acceptance) map[string]float64 {
	metrics := map[string]float64{"wall_seconds": float64(m.GrossNanos) / float64(time.Second)}
	// Only an exact-interval autonomy attestation establishes known session wait.
	// Human grading owns its separately measured operator wait in the grade step.
	if m.Autonomous {
		metrics["human_wait_seconds"] = float64(m.OperatorWaitNanos) / float64(time.Second)
		metrics["automated_wall_seconds"] = float64(m.GrossNanos-m.OperatorWaitNanos) / float64(time.Second)
	}
	for _, role := range []struct {
		name  string
		usage RoleUsage
	}{{"worker", m.Worker}, {"tl", m.TL}, {"reviewer", m.Reviewer}} {
		if role.usage.InputTokens != nil {
			metrics[role.name+"_input_tokens"] = float64(*role.usage.InputTokens)
		}
		if role.usage.OutputTokens != nil {
			metrics[role.name+"_output_tokens"] = float64(*role.usage.OutputTokens)
		}
	}
	if m.ReviewRounds != nil {
		metrics["review_rounds"] = float64(*m.ReviewRounds)
	}
	if m.GateRefusals != nil {
		metrics["gate_refusals"] = float64(*m.GateRefusals)
	}
	if a != nil && a.Status == "cases" && ValidateAcceptance(*a) == nil {
		metrics["hidden_pass_rate"] = float64(*a.Passed) / float64(*a.Total)
	}
	return metrics
}
