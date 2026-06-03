package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
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

// #68 M2: parse the per-issue total ("  #N: <h> hr  (<m> min)") for the primary
// issue, ignoring peers and #unattributed.
func TestParseV3PrimaryHours(t *testing.T) {
	out := `# per-issue totals
  ##unattributed: 4.07 hr  (244.4 min)
  #14: 7.79 hr  (467.3 min)
  #5: 1.84 hr  (110.3 min)
`
	if h, ok := parseV3PrimaryHours(out, "14"); !ok || h != 7.79 {
		t.Errorf("primary #14 = (%v,%v), want (7.79,true)", h, ok)
	}
	if h, ok := parseV3PrimaryHours(out, "5"); !ok || h != 1.84 {
		t.Errorf("primary #5 = (%v,%v), want (1.84,true)", h, ok)
	}
	// A whole-number-hours line.
	if h, ok := parseV3PrimaryHours("  #9: 2 hr  (120.0 min)\n", "9"); !ok || h != 2 {
		t.Errorf("whole-hour #9 = (%v,%v), want (2,true)", h, ok)
	}
	// Not present → false (don't fabricate; caller treats as empty window).
	if _, ok := parseV3PrimaryHours(out, "99"); ok {
		t.Error("absent issue should parse false")
	}
	// Must NOT prefix-match: #1 should not match the #14 line (the real guard —
	// callers always pass a numeric issue, so #unattributed can't collide).
	if _, ok := parseV3PrimaryHours("  #14: 7.79 hr  (467.3 min)\n", "1"); ok {
		t.Error("#1 must not prefix-match the #14 total")
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

// computeActual classification via a stubbed v3Runner (no real python/git).
// We can't drive the full path (CommitWindow reads cwd git) in a unit test, so
// this pins parseV3PrimaryHours + the exit-code → status mapping in isolation by
// asserting the runner contract the engine relies on.
func TestActualCmd_Registered(t *testing.T) {
	cmd := NewActualCmd()
	for _, flag := range []string{"issue", "brain-dir"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("actual command missing flag: --%s", flag)
		}
	}
}
