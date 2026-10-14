//go:build unix

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The quick-flow admission check reaches a window planner and a numstat read.
// Cancellation must stop the latter even if the former completed normally.
func TestQuickGrewPastReviewCancelsSlowNumstat(t *testing.T) {
	git, _ := publishRepo(t)
	publishFlowIssue(t, git, quickRecord, map[string]int{"cmd/a.go": 1})
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	marker := filepath.Join(bin, "entered")
	script := "#!/bin/sh\ncase \"$*\" in *'diff --numstat'*) touch \"" + marker + "\"; sleep 30;; esac\nexec \"" + realGit + "\" \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- quickGrewPastReview(ctx, []string{issuePathFor(69)}) }()
	until := time.Now().Add(4 * time.Second)
	for time.Now().Before(until) {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("numstat was never reached")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil || !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "canceled") {
			t.Fatalf("cancellation = %v", err)
		}
	// Allow bounded process-group/pipe cleanup and shared-suite scheduling.
	// The uncancelled fixture lasts thirty seconds, so natural completion
	// cannot satisfy this three-second cancellation assertion.
	case <-time.After(3 * time.Second):
		t.Fatal("numstat ignored cancellation")
	}
}
