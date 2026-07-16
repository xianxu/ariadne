package main

import (
	"bytes"
	"strings"
	"testing"
)

// TestActualDeviation pins the pure comparator's ratio/floor policy (#87).
func TestActualDeviation(t *testing.T) {
	cases := []struct {
		name             string
		passed, measured float64
		want             devVerdict
	}{
		{"exact match", 0.30, 0.30, devOK},
		{"small abs gap below floor", 0.11, 0.30, devOK},              // 0.19h apart
		{"tiny values, high ratio but below floor", 0.5, 0.05, devOK}, // 0.45h apart
		{"the nous#42 fabrication", 13.5, 0.30, devRefuse},            // 45×
		{"warn boundary 3x", 3.0, 1.0, devWarn},
		{"just under warn", 2.99, 1.0, devOK},
		{"refuse boundary 10x", 10.0, 1.0, devRefuse},
		{"just under refuse", 9.9, 1.0, devWarn},
		{"symmetric: measured >> passed", 0.30, 13.5, devRefuse},
		{"overclaim vs near-zero measured", 1.0, 0.05, devRefuse}, // ~1h apart, 20×
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ratio := actualDeviation(tc.passed, tc.measured)
			if got != tc.want {
				t.Fatalf("actualDeviation(%.2f, %.2f) = %v (ratio %.2f), want %v",
					tc.passed, tc.measured, got, ratio, tc.want)
			}
		})
	}
}

// TestCheckActualDeviation_SkipsWhenUnmeasurable: when the engine can't measure
// (here: a bogus issue with no commit window), the check must NOT block — it
// returns nil and emits nothing, so a legitimate close is never gated on an
// unavailable measurement.
func TestCheckActualDeviation_SkipsWhenUnmeasurable(t *testing.T) {
	var buf bytes.Buffer
	// #99999 has no commits referencing it → computeActual → actualNoWindow.
	if err := checkActualDeviation(&buf, "99999", 13.5, "issue"); err != nil {
		t.Fatalf("expected nil (skip) when unmeasurable, got: %v", err)
	}
	if out := strings.TrimSpace(buf.String()); out != "" {
		t.Fatalf("expected no output when unmeasurable, got: %q", out)
	}
}

// Milestone actuals are per-boundary increments, while the active-time engine
// currently returns a cumulative claim→HEAD issue measurement. Those values are
// not comparable: checking 0.37h M2 against 5.14h cumulative falsely refuses as
// a 14× deviation. Until the engine has a milestone window, the pass-path must
// skip this issue-close-only backstop.
func TestCheckActualDeviation_MilestoneSkipsCumulativeMeasurement(t *testing.T) {
	orig := computeActualForCloseFn
	calls := 0
	computeActualForCloseFn = func(string) actualResult {
		calls++
		return actualResult{Status: actualMeasured, Hours: 5.14}
	}
	t.Cleanup(func() { computeActualForCloseFn = orig })

	var buf bytes.Buffer
	if err := checkActualDeviation(&buf, "180", 0.37, "milestone"); err != nil {
		t.Fatalf("milestone increment must not be compared with cumulative actual: %v", err)
	}
	if calls != 0 {
		t.Fatalf("milestone mode ran cumulative measurement %d time(s), want 0", calls)
	}
}
