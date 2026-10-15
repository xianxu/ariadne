//go:build unix

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestUsageLateActualCancellationIsAnError(t *testing.T) {
	r, _, _, _, _ := ownedResolverFixture(t)
	chdirTo(t, r.root)
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	path := filepath.Join(r.root, "workshop/plans/000335-owned-close-gate.md")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Pairing a writer proves the collector reached actual's final ledger read.
	// The earlier accounting scan ignores non-regular files.
	paired := make(chan *os.File, 1)
	go func() { f, _ := os.OpenFile(path, os.O_WRONLY, 0600); paired <- f }()
	done := make(chan error, 1)
	go func() {
		_, err := collectUsage(ctx, usageFlags{Issue: 335, Burn: 3 * time.Hour}, time.Date(2026, 10, 14, 0, 20, 0, 0, time.UTC))
		done <- err
	}()
	var writer *os.File
	select {
	case writer = <-paired:
	case err := <-done:
		t.Fatalf("collector exited before late read: %v", err)
	case <-time.After(testfix.HangCeiling):
		t.Fatal("late actual read never reached")
	}
	if writer == nil {
		t.Fatal("FIFO writer did not open")
	}
	cancel()
	writer.Close()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("late actual cancellation returned %v, want context.Canceled", err)
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatal("collector did not return")
	}
}

type usageCancelWriter struct {
	bytes.Buffer
	cancel context.CancelFunc
}

func (w *usageCancelWriter) Write(p []byte) (int, error) { w.cancel(); return w.Buffer.Write(p) }
func TestUsageFinalOutputCancellationIsAnError(t *testing.T) {
	for _, format := range []string{"text", "json"} {
		t.Run(format, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("XDG_STATE_HOME", t.TempDir())
			testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			w := &usageCancelWriter{cancel: cancel}
			cmd := NewUsageCmd()
			cmd.SetOut(w)
			cmd.SetErr(&bytes.Buffer{})
			if format == "json" {
				cmd.SetArgs([]string{"--json"})
			} else {
				cmd.SetArgs([]string{})
			}
			err := cmd.ExecuteContext(ctx)
			if w.Len() == 0 {
				t.Fatal("report writer never reached")
			}
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation during final %s output returned %v", format, err)
			}
		})
	}
}
