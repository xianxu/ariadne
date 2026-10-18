//go:build unix

package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestMakeProbeBoundsDescendantPipes(t *testing.T) {
	for _, mode := range []string{"hanging-shell", "exited-parent"} {
		t.Run(mode, func(t *testing.T) {
			// The probe's deadline passes on cue once the descendant has reported
			// itself (#434); its own timeout stays at production length.
			ctx, expire := testfix.DeadlineOnCue()
			dir := t.TempDir()
			procedureWrite(t, dir, "Makefile", "HANG := $(shell sh ./hang.sh)\ntest:; @true\n")
			procedureWrite(t, dir, "hang.sh", "sleep 60 &\nprintf '%s' \"$!\" > child.pid\nwait\n")
			if mode == "exited-parent" {
				// A make process may exit while its children retain its pipes.
				procedureWrite(t, dir, "make", "#!/bin/sh\nsleep 60 &\nprintf '%s' \"$!\" > child.pid\nexit 0\n")
				if err := os.Chmod(filepath.Join(dir, "make"), 0755); err != nil {
					t.Fatal(err)
				}
				t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
			}
			done := make(chan error, 1)
			go func() { _, _, err := probeMakeTestFactsContext(ctx, dir); done <- err }()
			child := 0
			finished := false
			t.Cleanup(func() {
				if child > 0 {
					_ = syscall.Kill(child, syscall.SIGKILL)
				}
				if !finished {
					select {
					case <-done:
					case <-time.After(testfix.HangCeiling):
						t.Error("probe did not finish after fixture cleanup")
					}
				}
			})
			deadline := time.Now().Add(testfix.HangCeiling)
			for time.Now().Before(deadline) {
				raw, _ := os.ReadFile(filepath.Join(dir, "child.pid"))
				child, _ = strconv.Atoi(string(raw))
				if child > 0 {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
			if child == 0 {
				t.Fatal("descendant never started")
			}
			expire()
			select {
			case err := <-done:
				finished = true
				if err == nil {
					t.Error("probe with hanging descendant succeeded")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("probe exceeded deadline plus bounded pipe-drain interval")
			}
			// This bound must be shorter than the child's 60-second sleep,
			// otherwise natural exit can conceal broken group termination.
			if err := testfix.AwaitExit(child, time.Second); err != nil {
				t.Errorf("probe left running descendant: %v", err)
			}
		})
	}
}

func TestMakeProbeStatFailureIsNotAbsence(t *testing.T) {
	dir := t.TempDir()
	if err := os.Symlink("Makefile", filepath.Join(dir, "Makefile")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := probeMakeTestFacts(dir); err == nil || !strings.Contains(err.Error(), "Makefile") {
		t.Fatalf("stat failure must name Makefile, got %v", err)
	}
}
