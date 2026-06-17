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

func TestActualCmd_Registered(t *testing.T) {
	cmd := NewActualCmd()
	for _, flag := range []string{"issue", "brain-dir"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("actual command missing flag: --%s", flag)
		}
	}
}
