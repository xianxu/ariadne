package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
)

// #68 M2: cwd → Claude Code transcript-folder encoding ('/' and '.' → '-').
func TestCwdToTranscriptDir(t *testing.T) {
	cases := map[string]string{
		"/Users/x/workspace/nous":    "-Users-x-workspace-nous",
		"/Users/x/workspace/brain":   "-Users-x-workspace-brain",
		"/Users/x/.claude/projects":  "-Users-x--claude-projects", // leading '/.' → '--'
		"/w/worktree/ariadne-000040": "-w-worktree-ariadne-000040",
	}
	for in, want := range cases {
		if got := cwdToTranscriptDir(in); got != want {
			t.Errorf("cwdToTranscriptDir(%q) = %q, want %q", in, got, want)
		}
	}
}

// #68 M2: dir-selection is brain + repo, existing folders only, never unrelated.
func TestSelectActualDirs(t *testing.T) {
	root := t.TempDir()
	prev := transcriptsRoot
	transcriptsRoot = root
	t.Cleanup(func() { transcriptsRoot = prev })

	repo := "/w/nous"
	brain := "/w/brain"
	mk := func(p string) {
		if err := os.Mkdir(filepath.Join(root, cwdToTranscriptDir(p)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mk(repo)
	mk(brain)
	mk("/w/pair") // unrelated — present on disk but NOT passed in → must be excluded

	got := selectActualDirs(repo, brain)
	want := []string{
		filepath.Join(root, cwdToTranscriptDir(brain)), // brain first
		filepath.Join(root, cwdToTranscriptDir(repo)),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("selectActualDirs = %v, want %v (brain+repo only, no unrelated)", got, want)
	}

	// A repo whose folder doesn't exist is silently skipped (not invented).
	got2 := selectActualDirs("/w/does-not-exist", brain)
	if len(got2) != 1 || got2[0] != filepath.Join(root, cwdToTranscriptDir(brain)) {
		t.Errorf("missing repo folder should be skipped; got %v", got2)
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
