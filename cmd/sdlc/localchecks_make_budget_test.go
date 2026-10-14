package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Exercise the actual Makefile process through the whole CI facts call chain.
// A timeout checked only after make returns cannot satisfy this contract.
func TestWaitCIGateBoundsRealMakeProbe(t *testing.T) {
	for _, mode := range []string{"budget", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			oldNow := ciWaitNow
			t.Cleanup(func() { ciWaitNow = oldNow })
			epoch := time.Unix(0, 0)
			calls := 0
			ciWaitNow = func() time.Time {
				calls++
				if calls == 1 || mode == "cancel" {
					return epoch
				}
				return epoch.Add(ciWaitLimit - 100*time.Millisecond)
			}
			marker := filepath.Join(t.TempDir(), "probe-started")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			stopped := make(chan struct{})
			watcherDone := make(chan struct{})
			if mode == "cancel" {
				go func() {
					defer close(watcherDone)
					tick := time.NewTicker(5 * time.Millisecond)
					defer tick.Stop()
					for {
						select {
						case <-stopped:
							return
						case <-tick.C:
							if _, err := os.Stat(marker); err == nil {
								cancel()
								return
							}
						}
					}
				}()
				defer func() { close(stopped); <-watcherDone }()
			}
			removes, prunes := 0, 0
			var worktree string
			git := func(args ...string) (string, error) {
				if len(args) > 1 && args[0] == "worktree" {
					switch args[1] {
					case "add":
						worktree = args[3]
						return "", os.WriteFile(filepath.Join(worktree, "Makefile"), []byte("SLOW := $(shell touch '"+marker+"'; sleep 1.5)\ntest:; @true\n"), 0600)
					case "remove":
						removes++
					case "prune":
						prunes++
					}
				}
				return "", nil
			}
			start := time.Now()
			err := waitCIGate(ctx, &checksGH{}, "o/r", 3, "b", "h", git, nil, false, io.Discard)
			elapsed := time.Since(start)
			want := context.DeadlineExceeded
			if mode == "cancel" {
				want = context.Canceled
			}
			if !errors.Is(err, want) || elapsed > 700*time.Millisecond {
				t.Errorf("probe ignored %s: elapsed=%v err=%v", mode, elapsed, err)
			}
			if _, err := os.Stat(marker); err != nil {
				t.Errorf("real make probe never started: %v", err)
			}
			if worktree == "" || removes != 1 || prunes != 1 {
				t.Errorf("cleanup: worktree=%q removes=%d prunes=%d", worktree, removes, prunes)
			}
			if _, err := os.Stat(worktree); !os.IsNotExist(err) {
				t.Errorf("probe worktree remains: %v", err)
			}
		})
	}
}
