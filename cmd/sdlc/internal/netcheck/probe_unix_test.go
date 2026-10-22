//go:build unix

package netcheck

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/procgroup"
)

func TestNativeUtilityTimeoutAndCancellationOwnDescendants(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(strconv.FormatBool(cancelled), func(t *testing.T) {
			dir := t.TempDir()
			pidfile := filepath.Join(dir, "pid")
			// exec retains the utility PID as the group leader; the background child
			// retains output pipes and must be terminated even if the leader exits.
			script := "#!/bin/sh\necho $$ > '" + pidfile + "'\n/bin/sleep 30 &\nwait\n"
			if err := os.WriteFile(filepath.Join(dir, "dig"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", dir)
			p := &nativeProber{commandTimeout: 150 * time.Millisecond}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, err := p.Lookup(ctx, "example.com", Direct); done <- err }()
			var pid int
			deadline := time.Now().Add(time.Second)
			for time.Now().Before(deadline) {
				b, err := os.ReadFile(pidfile)
				if err == nil {
					pid, _ = strconv.Atoi(strings.TrimSpace(string(b)))
					break
				}
				time.Sleep(time.Millisecond)
			}
			if pid == 0 {
				t.Fatal("utility never started")
			}
			if cancelled {
				cancel()
			}
			select {
			case err := <-done:
				if cancelled && !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
				if !cancelled && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("utility not joined")
			}
			if err := procgroup.Kill(pid, syscall.Signal(0)); !errors.Is(err, os.ErrProcessDone) {
				t.Fatalf("group remains: %v", err)
			}
		})
	}
}
