package testfix

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// BlockingOrigin returns a bare origin whose pre-receive hook records its pid
// and then blocks, so a push into it is mid-flight on the receive side until
// the caller cancels it. hookPID waits for the hook to start and returns its
// pid. A cancellation test then proves, with AwaitExit, that the far end of the
// local transport died with the cancelled git: one left running keeps writing
// into the origin after the call returns, racing t.TempDir's cleanup (#302).
// The hook gives up by itself after twice HangCeiling, longer than AwaitExit
// waits, so a regression fails rather than passes, and is not left running for
// good.
func BlockingOrigin(t *testing.T) (origin string, hookPID func() int) {
	t.Helper()
	origin = filepath.Join(t.TempDir(), "origin.git")
	Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	hooks := filepath.Join(origin, "hooks")
	Git(t, origin, "config", "core.hooksPath", hooks) // a global core.hooksPath must not hide it
	pidFile := filepath.Join(t.TempDir(), "hook.pid")
	ticks := int(2 * HangCeiling / (100 * time.Millisecond)) // outlives AwaitExit's ceiling
	hook := fmt.Sprintf("#!/bin/sh\necho $$ > %[1]q.tmp && mv %[1]q.tmp %[1]q\ni=0\nwhile [ $i -lt %[2]d ]; do sleep 0.1; i=$((i+1)); done\n", pidFile, ticks)
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooks, "pre-receive"), []byte(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	return origin, func() int {
		t.Helper()
		for deadline := time.Now().Add(HangCeiling); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
			if b, err := os.ReadFile(pidFile); err == nil {
				if pid, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil {
					return pid
				}
			}
		}
		t.Fatalf("the pre-receive hook never started within %v", HangCeiling)
		return 0
	}
}
