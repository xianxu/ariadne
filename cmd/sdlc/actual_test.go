package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
)

// Transcript-source selection (cwd encoding, brain+repo dir selection, Codex
// cwd matching) moved to internal/transcripts in #134 and is tested there
// (claude_test.go / codex_test.go / transcripts_test.go). actual_test.go keeps
// the engine-glue contract tests below.

// nonEmpty drops "" cwds (an unresolved brain/repo path) while preserving order
// — the brain+repo list computeActual hands the harness registry.
func TestNonEmpty(t *testing.T) {
	cases := []struct {
		in   []string
		want []string
	}{
		{[]string{"/brain", "/repo"}, []string{"/brain", "/repo"}},
		{[]string{"", "/repo"}, []string{"/repo"}},
		{[]string{"/brain", ""}, []string{"/brain"}},
		{[]string{"", ""}, nil},
		{nil, nil},
	}
	for _, c := range cases {
		if got := nonEmpty(c.in...); !reflect.DeepEqual(got, c.want) {
			t.Errorf("nonEmpty(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// #110: the engine-result → outcome contract (pure; replaces the old classifyV3
// exit-code mapping now that the engine is in-process).
func TestStatusFromResult(t *testing.T) {
	cases := []struct {
		name    string
		res     activetime.Result
		want    actualStatus
		wantHrs float64
	}{
		{
			"telemetry gap → judgment",
			activetime.Result{Status: activetime.TelemetryGap},
			actualTelemetryGap, 0,
		},
		{
			"measured + issue present → measured (minutes→hours)",
			activetime.Result{Status: activetime.Measured, PerIssue: map[string]float64{"68": 90}},
			actualMeasured, 1.5,
		},
		{
			"measured but issue absent → empty window",
			activetime.Result{Status: activetime.Measured, PerIssue: map[string]float64{"5": 30}},
			actualEmptyWindow, 0,
		},
		{
			"empty window → empty window",
			activetime.Result{Status: activetime.EmptyWindow, PerIssue: map[string]float64{}},
			actualEmptyWindow, 0,
		},
	}
	for _, c := range cases {
		st, hrs := statusFromResult(c.res, "68")
		if st != c.want || hrs != c.wantHrs {
			t.Errorf("%s: statusFromResult = (%d,%v), want (%d,%v)", c.name, st, hrs, c.want, c.wantHrs)
		}
	}
}

func TestPrintActualWarnings(t *testing.T) {
	var out bytes.Buffer
	printActual(&out, actualResult{
		Status: actualMeasured,
		Issue:  "8",
		Hours:  1.25,
		Window: "abc12345 → HEAD",
		Warnings: []string{
			"#8 10.0m/100% mention fallback without issue commit boundary",
		},
	})
	s := out.String()
	if !strings.Contains(s, "attribution warning:") || !strings.Contains(s, "fallback") {
		t.Fatalf("actual output missing warning:\n%s", s)
	}
	if !strings.Contains(s, "close with") {
		t.Fatalf("actual output should still include close suggestion:\n%s", s)
	}
}

// TestWindowStart pins the #113 window-start picker: the EARLIER of the
// parent-of-first-#N anchor and the claim's working-transition, with either
// empty falling back to the other.
func TestWindowStart(t *testing.T) {
	const (
		early = "2026-06-10T09:00:00-07:00"
		late  = "2026-06-14T17:00:00-07:00"
	)
	cases := []struct {
		name             string
		parent, wt, want string
	}{
		{"claim-early: wt earlier wins", late, early, early},
		{"late claim: parent earlier wins", early, late, early},
		{"no working-transition: parent", late, "", late},
		{"no commit anchor: wt", "", early, early},
		{"both empty", "", "", ""},
		{"equal: either (parent)", early, early, early},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := windowStart(c.parent, c.wt); got != c.want {
				t.Errorf("windowStart(%q, %q) = %q, want %q", c.parent, c.wt, got, c.want)
			}
		})
	}
}

// TestResolveWindowStart pins the #116 three-anchor preference: explicit
// `started:` → WorkingTransitionISO heuristic → commit-parent, all delegated to
// windowStart's earlier-of pick (so a `started:` later than the commit window
// never regresses the start).
func TestResolveWindowStart(t *testing.T) {
	const (
		t0 = "2026-06-10T08:00:00-07:00" // earliest
		t1 = "2026-06-11T09:00:00-07:00"
		t2 = "2026-06-12T12:00:00-07:00" // latest
	)
	cases := []struct {
		name                      string
		parent, started, wt, want string
	}{
		{"started supersedes wt (earliest wins)", t2, t0, t1, t0},
		{"no started: falls back to wt heuristic", t2, "", t1, t1},
		{"neither: commit-parent default", t2, "", "", t2},
		{"started later than parent: no regression", t0, t2, t1, t0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := resolveWindowStart(c.parent, c.started, c.wt); got != c.want {
				t.Errorf("resolveWindowStart(%q, %q, %q) = %q, want %q", c.parent, c.started, c.wt, got, c.want)
			}
		})
	}
}

func TestActualCmd_Registered(t *testing.T) {
	cmd := NewActualCmd()
	for _, flag := range []string{"issue", "brain-dir"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("actual command missing flag: --%s", flag)
		}
	}
}
