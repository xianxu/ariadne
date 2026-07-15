package main

import (
	"bytes"
	"strings"
	"testing"
)

// #178: the omit-path ADOPTS a measured actual instead of refusing with a
// "→ close with: --actual N" suggestion the agent copies verbatim (the spine's
// second-largest refusal volume, #172). Only actualMeasured adopts; every other
// status keeps the explainActual refusal.
func TestResolveOmittedActual(t *testing.T) {
	cases := []struct {
		name   string
		res    actualResult
		want   string
		wantOK bool
	}{
		{"measured adopts, %.2f pinned", actualResult{Status: actualMeasured, Hours: 1.234}, "1.23", true},
		{"measured adopts whole hours", actualResult{Status: actualMeasured, Hours: 7.0}, "7.00", true},
		{"telemetry gap refuses", actualResult{Status: actualTelemetryGap}, "", false},
		{"empty window refuses", actualResult{Status: actualEmptyWindow}, "", false},
		{"no window refuses", actualResult{Status: actualNoWindow}, "", false},
		{"engine error refuses", actualResult{Status: actualError}, "", false},
	}
	for _, tc := range cases {
		got, ok := resolveOmittedActual(tc.res)
		if ok != tc.wantOK || got != tc.want {
			t.Errorf("%s: resolveOmittedActual = (%q, %v), want (%q, %v)", tc.name, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestFormatAdoptLine(t *testing.T) {
	res := actualResult{Status: actualMeasured, Hours: 1.23,
		Window: "a1b2c3d4 → HEAD", Peers: []string{"172", "173"}}

	line := formatAdoptLine(res)
	for _, want := range []string{"1.23", "a1b2c3d4 → HEAD", "#172", "#173", "measured"} {
		if !strings.Contains(line, want) {
			t.Errorf("adopt line missing %q: %q", want, line)
		}
	}
}

// Milestone mode NEVER adopts (#178 close-review Important #1): computeActual's
// window is issue-scoped, so the cumulative value would double-count across
// per-milestone project detail blocks (lessons.md increments rule). The
// measurement is still returned — it feeds the suggest flow unchanged.
func TestAdoptOmittedActualMilestoneKeepsSuggestFlow(t *testing.T) {
	orig := computeActualForCloseFn
	computeActualForCloseFn = func(string) actualResult {
		return actualResult{Status: actualMeasured, Hours: 6.57, Window: "abcd1234 → HEAD"}
	}
	t.Cleanup(func() { computeActualForCloseFn = orig })

	var stderr bytes.Buffer
	f := &closeFlags{Issue: 172}
	res, ok := adoptOmittedActual(&stderr, f, "172", "milestone")
	if ok || f.Actual != "" {
		t.Fatalf("milestone mode must not adopt, got ok=%v f.Actual=%q", ok, f.Actual)
	}
	if res.Status != actualMeasured || res.Hours != 6.57 {
		t.Errorf("measurement must be returned for the suggest flow, got %+v", res)
	}
	if stderr.Len() != 0 {
		t.Errorf("no side effects on the milestone arm, got %q", stderr.String())
	}
}

// The wiring arm: adoptOmittedActual measures via the computeActualForCloseFn
// seam exactly once, adopts into f.Actual, and prints the info line.
func TestAdoptOmittedActual(t *testing.T) {
	calls := 0
	orig := computeActualForCloseFn
	computeActualForCloseFn = func(issueStr string) actualResult {
		calls++
		return actualResult{Status: actualMeasured, Hours: 0.65, Window: "abcd1234 → HEAD", Issue: issueStr}
	}
	t.Cleanup(func() { computeActualForCloseFn = orig })

	var stderr bytes.Buffer
	f := &closeFlags{Issue: 178}
	_, ok := adoptOmittedActual(&stderr, f, "178", "issue")
	if !ok || f.Actual != "0.65" {
		t.Fatalf("want adoption of 0.65, got ok=%v f.Actual=%q", ok, f.Actual)
	}
	if calls != 1 {
		t.Errorf("engine must run exactly once (deviation check is skipped for adopted values), got %d", calls)
	}
	if out := stderr.String(); !strings.Contains(out, "0.65") || !strings.Contains(out, "measured") {
		t.Errorf("adopt info line not printed: %q", out)
	}

	// unmeasurable → no adoption, no output side effects (the caller then runs
	// explainActual + exit; that arm's decision is pinned by TestResolveOmittedActual)
	computeActualForCloseFn = func(string) actualResult { return actualResult{Status: actualTelemetryGap} }
	var stderr2 bytes.Buffer
	f2 := &closeFlags{Issue: 178}
	res, ok2 := adoptOmittedActual(&stderr2, f2, "178", "issue")
	if ok2 {
		t.Fatal("telemetry gap must not adopt")
	}
	if f2.Actual != "" {
		t.Errorf("f.Actual must stay empty on refusal, got %q", f2.Actual)
	}
	if res.Status != actualTelemetryGap {
		t.Errorf("the measurement must be returned for the caller's explainActual (no re-measure), got %v", res.Status)
	}
}

// #172-instrument guard: the adopt info line must classify as NEITHER a bypass
// ACK nor a refusal under any gate signature — else the friction report would
// count routine adoptions as gate events.
func TestAdoptLineNoGatesigCollision(t *testing.T) {
	res := actualResult{Status: actualMeasured, Hours: 1.23,
		Window: "a1b2c3d4 → HEAD", Peers: []string{"172"}}
	// as rendered: cinfo prefixes "==> " with ANSI + the reset marker
	assertNoGatesigCollision(t, "\x1b[1;36m==>\x1b[0m "+formatAdoptLine(res))
}
