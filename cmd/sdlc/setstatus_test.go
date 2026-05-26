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

func TestCheckTransitionGuards_WorkingNeedsEstimate(t *testing.T) {
	fmNoEst := "id: 000001\nstatus: open\n"
	body := "# Title\n"
	err := checkTransitionGuards("open", "working", fmNoEst, body)
	if err == nil {
		t.Fatal("expected refusal for missing estimate_hours")
	}
	if !strings.Contains(err.Error(), "estimate_hours") {
		t.Errorf("error message should mention estimate_hours: %q", err.Error())
	}

	fmEmpty := "id: 000001\nstatus: open\nestimate_hours:\n"
	if err := checkTransitionGuards("open", "working", fmEmpty, body); err == nil {
		t.Error("expected refusal for empty estimate_hours value")
	}

	fmOK := "id: 000001\nstatus: open\nestimate_hours: 3.5\n"
	if err := checkTransitionGuards("open", "working", fmOK, body); err != nil {
		t.Errorf("expected nil error for estimate_hours = 3.5, got: %v", err)
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
