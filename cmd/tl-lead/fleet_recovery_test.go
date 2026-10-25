package main

import (
	"bytes"
	"testing"
)

// A claim is pending-work evidence, never evidence of a shutdown asleep mark.
func TestFleetResumeDoesNotInferAsleepFromClaims(t *testing.T) {
	d, c, cwd := fleetFixture(t)
	var out bytes.Buffer
	_ = fleetResume(&out, d, cwd, true)
	if got := c.attempts(); len(got) != 0 {
		t.Fatalf("resumed without positive asleep evidence: %v\n%s", got, out.String())
	}
}
