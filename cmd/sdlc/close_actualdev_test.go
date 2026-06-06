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
		{"small abs gap below floor", 0.11, 0.30, devOK},      // 0.19h apart
		{"tiny values, high ratio but below floor", 0.5, 0.05, devOK}, // 0.45h apart
		{"the nous#42 fabrication", 13.5, 0.30, devRefuse},    // 45×
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
	if err := checkActualDeviation(&buf, "99999", 13.5); err != nil {
		t.Fatalf("expected nil (skip) when unmeasurable, got: %v", err)
	}
	if out := strings.TrimSpace(buf.String()); out != "" {
		t.Fatalf("expected no output when unmeasurable, got: %q", out)
	}
}
