package landingline

import (
	"strings"
	"testing"
)

// #561: the one ramp rule names the literal two-car default, the seven-day
// manual review and its measurements, and never an automatic action.
func TestFormatTrainRampPolicy(t *testing.T) {
	got := FormatTrainRampPolicy()
	for _, want := range []string{"boards 2 cars", "only by hand", "7-day window", "escapes", "denominator",
		"unresolved", "split costs", "Nothing raises a cap or enables trains automatically", ".sdlc/landing.json", "train_interval_minutes to 0"} {
		if !strings.Contains(got, want) {
			t.Errorf("policy lacks %q:\n%s", want, got)
		}
	}
	for _, bad := range []string{"automatically raise", "auto-ramp", "threshold"} {
		if strings.Contains(strings.ToLower(got), bad) {
			t.Errorf("policy promises %q:\n%s", bad, got)
		}
	}
}
