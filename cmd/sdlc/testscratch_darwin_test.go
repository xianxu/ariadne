package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A folder removeTree cannot take (an immutable file: chflags uchg, which no
// chmod lifts) is logged as not removed, and the sweep goes on to the next
// candidate (#631).
func TestSweepStaleTempLogsAFailedRemovalAndContinues(t *testing.T) {
	dir := t.TempDir()
	stuck := filepath.Join(dir, "sdlc-run-suite-100-1")
	next := filepath.Join(dir, "sdlc-run-suite-101-2")
	for _, d := range []string{stuck, next} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	pinned := filepath.Join(stuck, "pinned")
	if err := os.WriteFile(pinned, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("chflags", "uchg", pinned).CombinedOutput(); err != nil {
		t.Fatalf("chflags: %v %s", err, out)
	}
	t.Cleanup(func() { _ = exec.Command("chflags", "nouchg", pinned).Run() })
	var log bytes.Buffer
	sweepStaleTemp(dir, true, time.Now(), func(int) bool { return false }, &log)
	if _, err := os.Stat(stuck); err != nil {
		t.Fatalf("the immutable folder vanished: %v", err)
	}
	if _, err := os.Stat(next); !os.IsNotExist(err) {
		t.Fatalf("the sweep stopped at the failed removal: %v\n%s", err, log.String())
	}
	if !strings.Contains(log.String(), "[scratch sweep] could not remove "+stuck+" after ") {
		t.Fatalf("no failure line:\n%s", log.String())
	}
}
