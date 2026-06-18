package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsValidStatus(t *testing.T) {
	for _, s := range validStatuses {
		if !isValidStatus(s) {
			t.Errorf("isValidStatus(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "WORKING", "done!", "completed", "in-progress"} {
		if isValidStatus(s) {
			t.Errorf("isValidStatus(%q) = true, want false", s)
		}
	}
}

func TestCheckTransitionGuards_RefusesDone(t *testing.T) {
	fm := "id: 000001\nstatus: working\nestimate_hours: 2\n"
	body := "# Title\n"
	err := checkTransitionGuards("working", "done", fm, body)
	if err == nil {
		t.Fatal("expected refusal")
	}
	if !strings.Contains(err.Error(), "sdlc close") {
		t.Errorf("error should redirect to sdlc close: %q", err.Error())
	}
}

// TestCheckTransitionGuards_WorkingNoLongerNeedsEstimate pins the #113
// decoupling: → working is now estimate-free (claim is a cheap lock; the
// estimate gate moved to `sdlc change-code`). All three shapes — missing,
// empty, and present estimate_hours — must flip cleanly.
func TestCheckTransitionGuards_WorkingNoLongerNeedsEstimate(t *testing.T) {
	body := "# Title\n"
	cases := []struct {
		name string
		fm   string
	}{
		{"missing estimate", "id: 000001\nstatus: open\n"},
		{"empty estimate", "id: 000001\nstatus: open\nestimate_hours:\n"},
		{"present estimate", "id: 000001\nstatus: open\nestimate_hours: 3.5\n"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkTransitionGuards("open", "working", tt.fm, body); err != nil {
				t.Errorf("open → working should be estimate-free now, got: %v", err)
			}
		})
	}
}

func TestCheckTransitionGuards_ReopenNeedsLogEntry(t *testing.T) {
	fm := "id: 000001\nstatus: done\n"
	today := todayIso()

	// No Log section at all → refused.
	if err := checkTransitionGuards("done", "open", fm, "# T\n"); err == nil {
		t.Error("expected refusal: no Log section at all")
	}

	// Log section exists but no today entry → refused.
	bodyOld := "# T\n\n## Log\n\n- 2025-01-01: started\n"
	if err := checkTransitionGuards("done", "open", fm, bodyOld); err == nil {
		t.Error("expected refusal: Log lacks today's entry")
	}

	// Log section with today's entry → OK.
	bodyOK := "# T\n\n## Log\n\n- " + today + ": reopened — found regression\n"
	if err := checkTransitionGuards("done", "open", fm, bodyOK); err != nil {
		t.Errorf("expected ok for today-entry, got: %v", err)
	}

	// Log entry as ### Header instead of bullet — still recognized.
	bodyHeader := "# T\n\n## Log\n\n### " + today + "\nreopened — context\n"
	if err := checkTransitionGuards("done", "open", fm, bodyHeader); err != nil {
		t.Errorf("expected ok for today-header, got: %v", err)
	}
}

func TestCheckTransitionGuards_NormalTransitions(t *testing.T) {
	// open → blocked: allowed.
	fm := "id: 000001\nstatus: open\n"
	if err := checkTransitionGuards("open", "blocked", fm, "# T\n"); err != nil {
		t.Errorf("open → blocked should be ok, got: %v", err)
	}
	// working → blocked: allowed.
	fm = "id: 000001\nstatus: working\nestimate_hours: 1\n"
	if err := checkTransitionGuards("working", "blocked", fm, "# T\n"); err != nil {
		t.Errorf("working → blocked should be ok, got: %v", err)
	}
	// working → punt: allowed.
	if err := checkTransitionGuards("working", "punt", fm, "# T\n"); err != nil {
		t.Errorf("working → punt should be ok, got: %v", err)
	}
}

func TestLogHasEntryToday_VariousShapes(t *testing.T) {
	today := "2026-05-25"
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"bullet entry", "## Log\n\n- 2026-05-25: thing\n", true},
		{"date header", "## Log\n\n### 2026-05-25\n", true},
		{"date in middle of paragraph", "## Log\n\nWork done on 2026-05-25.\n", true},
		{"no log section", "## Plan\n- [ ] x\n", false},
		{"log without today", "## Log\n\n- 2025-12-31: old\n", false},
		// today appears only after another ## section → must not count.
		{"date past next section", "## Log\n\n- old: stuff\n\n## Plan\n- 2026-05-25 in plan\n", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := logHasEntryToday(c.body, today); got != c.want {
				t.Errorf("logHasEntryToday(...) = %v, want %v", got, c.want)
			}
		})
	}
}

// TestRunSetStatus_DryRunHonored exercises the dry-run path end-to-end:
// valid working → blocked transition, no file mutation, summary printed.
func TestRunSetStatus_DryRunHonored(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "000042-foo.md")
	original := "---\nid: 000042\nstatus: working\nestimate_hours: 2\nupdated: 2026-04-01\n---\n# Foo\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	f := &setStatusFlags{
		Issue:     42,
		Status:    "blocked",
		IssuesDir: dir,
		DryRun:    true,
	}
	if err := runSetStatus(&stdout, &stderr, f); err != nil {
		t.Fatalf("runSetStatus err: %v", err)
	}
	// File contents preserved.
	got, _ := os.ReadFile(path)
	if string(got) != original {
		t.Errorf("dry-run mutated file:\n--- got ---\n%s\n--- want ---\n%s", got, original)
	}
	if !strings.Contains(stdout.String(), "Would update") {
		t.Errorf("dry-run stdout missing summary: %q", stdout.String())
	}
}

// TestRunSetStatus_WritesNewStatus tests the happy path: valid
// transition writes the file with the new status + today's updated.
// TestApplyStatus_StampsStarted pins #116: the open→working flip stamps an
// idempotent `started:` engagement anchor (injected clock for determinism); an
// existing stamp is never overwritten.
func TestApplyStatus_StampsStarted(t *testing.T) {
	orig := startedClock
	t.Cleanup(func() { startedClock = orig })
	dir := t.TempDir()

	// Case A: open without started → stamped.
	pathA := filepath.Join(dir, "000008-a.md")
	if err := os.WriteFile(pathA, []byte("---\nid: 000008\nstatus: open\nestimate_hours: 1\n---\n# A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	startedClock = func() string { return "2026-06-18T10:00:00-07:00" }
	if _, _, _, err := applyStatus(dir, 8, "working", false, false); err != nil {
		t.Fatalf("applyStatus A: %v", err)
	}
	if a, _ := os.ReadFile(pathA); !strings.Contains(string(a), "started: 2026-06-18T10:00:00-07:00") {
		t.Errorf("open→working should stamp started:\n%s", a)
	}

	// Case B: open WITH an existing started → not overwritten (idempotent).
	pathB := filepath.Join(dir, "000009-b.md")
	if err := os.WriteFile(pathB, []byte("---\nid: 000009\nstatus: open\nestimate_hours: 1\nstarted: 2025-01-01T00:00:00-07:00\n---\n# B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	startedClock = func() string { return "2099-12-31T23:59:59-07:00" }
	if _, _, _, err := applyStatus(dir, 9, "working", false, false); err != nil {
		t.Fatalf("applyStatus B: %v", err)
	}
	b, _ := os.ReadFile(pathB)
	if strings.Contains(string(b), "2099") {
		t.Errorf("existing started: must not be overwritten:\n%s", b)
	}
	if !strings.Contains(string(b), "started: 2025-01-01T00:00:00-07:00") {
		t.Errorf("original started: lost:\n%s", b)
	}
}

func TestRunSetStatus_WritesNewStatus(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "000007-foo.md")
	if err := os.WriteFile(path, []byte("---\nid: 000007\nstatus: open\nestimate_hours: 1\n---\n# Foo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	f := &setStatusFlags{
		Issue:     7,
		Status:    "working",
		IssuesDir: dir,
	}
	if err := runSetStatus(&stdout, &stderr, f); err != nil {
		t.Fatalf("runSetStatus err: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "status: working") {
		t.Errorf("file does not contain new status:\n%s", got)
	}
	if !strings.Contains(string(got), "updated: "+todayIso()) {
		t.Errorf("file missing today's updated:\n%s", got)
	}
}

// todayIso returns time.Now() formatted as YYYY-MM-DD — same format
// the production code uses. (Time injection would be cleaner; for M4
// we match the existing close.go posture of calling time.Now()
// directly.)
func todayIso() string {
	return time.Now().Format("2006-01-02")
}
