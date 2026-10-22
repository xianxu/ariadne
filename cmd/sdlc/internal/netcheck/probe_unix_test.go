//go:build unix

package netcheck

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/procgroup"
)

func TestNativeUtilityTimeoutAndCancellationOwnDescendants(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(strconv.FormatBool(cancelled), func(t *testing.T) {
			dir := t.TempDir()
			pidfile := filepath.Join(dir, "pid")
			// exec retains the utility PID as the group leader; the background child
			// retains output pipes and must be terminated even if the leader exits.
			script := "#!/bin/sh\n/bin/sleep 30 &\necho $$ > '" + pidfile + "'\nwait\n"
			if err := os.WriteFile(filepath.Join(dir, "dig"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir)
			// Trigger the deadline only after the utility and its child exist.
			// The command's own bound is only a hang guard in this fixture.
			p := &nativeProber{commandTimeout: testfix.HangCeiling}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			stop := cancel
			if !cancelled {
				ctx, stop = testfix.DeadlineOnCue(ctx)
			}
			done := make(chan error, 1)
			go func() { _, err := p.Lookup(ctx, "example.com", Direct); done <- err }()
			pid := testfix.AwaitPIDFile(pidfile, testfix.HangCeiling, nil)
			if pid == 0 {
				t.Fatal("utility never started")
			}
			stop()
			select {
			case err := <-done:
				if cancelled && !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
				if !cancelled && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
			case <-time.After(testfix.HangCeiling):
				t.Fatal("utility not joined")
			}
			if err := procgroup.Kill(pid, syscall.Signal(0)); !errors.Is(err, os.ErrProcessDone) {
				t.Fatalf("group remains: %v", err)
			}
		})
	}
}
