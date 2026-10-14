//go:build unix

package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestRunLocalCheckTerminatesDescendants(t *testing.T) {
	for _, mode := range []string{"timeout", "exited-parent"} {
		t.Run(mode, func(t *testing.T) {
			d, sha, gh, _ := localRunRepo(t, "skipped")
			d.minute = 500 * time.Millisecond
			pidfile := filepath.Join(t.TempDir(), "child.pid")
			script := `sleep 60 & printf '%s' "$!" > "$1"; wait`
			timeout := 1
			if mode == "exited-parent" {
				script = `sleep 60 & printf '%s' "$!" > "$1"; exit 0`
				timeout = 30 // pipe drain, rather than timeout, must terminate the child
			}
			c := localCheck{Name: "suite", Command: []string{"sh", "-c", script, "fixture", pidfile}, TimeoutMinutes: timeout}
			type result struct {
				rec localRunRecord
				err error
			}
			done := make(chan result, 1)
			go func() { rec, err := runLocalCheck(context.Background(), d, c, sha); done <- result{rec, err} }()
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
						t.Error("runner did not finish after fixture cleanup")
					}
				}
			})
			deadline := time.Now().Add(testfix.HangCeiling)
			for time.Now().Before(deadline) {
				raw, _ := os.ReadFile(pidfile)
				child, _ = strconv.Atoi(string(raw))
				if child > 0 {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
			if child == 0 {
				t.Fatal("descendant never started")
			}
			select {
			case got := <-done:
				finished = true
				if got.err != nil || got.rec.passed() || gh.statuses[sha]["sdlc-local/suite"] != "failure" {
					t.Fatalf("record = %+v, err = %v, statuses = %v", got.rec, got.err, gh.statuses[sha])
				}
				if got.rec.TimedOut != (mode == "timeout") {
					t.Errorf("TimedOut = %v for %s", got.rec.TimedOut, mode)
				}
			case <-time.After(8 * time.Second):
				t.Fatal("runner exceeded timeout plus pipe-drain bound")
			}
			// The process cleanup is synchronous with return; allow the OS a
			// small reaping interval, but do not wait for the 60-second sleeper.
			if err := testfix.AwaitExit(child, time.Second); err != nil {
				t.Errorf("runner left a live descendant: %v", err)
			}
		})
	}
}
