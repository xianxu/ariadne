package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func TestQueueRuntimeLifetimeAndPause(t *testing.T) {
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	ctx, q, err := startQueueOperation(context.Background(), t.TempDir(), "head", []string{"merge", "landing"})
	if err != nil {
		t.Fatal(err)
	}
	defer q.finish("")
	if queueRuntimeFrom(ctx) != q || isQueueAdmitted(ctx) {
		t.Fatal("lost waiting runtime")
	}
	s := &queuegate.Store{Root: q.store.Root}
	if x, err := s.Snapshot(ctx); err != nil || len(x.Active) != 0 {
		t.Fatalf("waiting became active: %+v %v", x, err)
	}
	if err := s.Pause(ctx, "one", "hold", false); err != nil {
		t.Fatal(err)
	}
	called := false
	_, err = q.admit(ctx, func() (bool, error) { called = true; return true, nil })
	var paused *queuegate.PausedError
	if !errors.As(err, &paused) || called {
		t.Fatalf("paused admission called resource: %v %v", called, err)
	}
	if err := s.Resume(ctx, "one"); err != nil {
		t.Fatal(err)
	}
	if err := q.waitAdmission(ctx, io.Discard); err != nil {
		t.Fatal(err)
	}
	if !isQueueAdmitted(withQueueAdmitted(ctx)) {
		t.Fatal("admission marker missing")
	}
	if x, err := s.Snapshot(ctx); err != nil || len(x.Active) != 2 {
		t.Fatalf("active kinds absent: %+v %v", x, err)
	}
	if err := q.finish(""); err != nil {
		t.Fatal(err)
	}
	if x, err := s.Snapshot(ctx); err != nil || len(x.Active) != 0 {
		t.Fatalf("completed lifetime leaked: %+v %v", x, err)
	}
}

func TestQueueRealHelperPause(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", "../../scripts/test/test-lock-pause.test.py")
	cmd.WaitDelay = time.Second
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("private real helper: %v\n%s", err, out)
	}
}

func TestQueueRuntimeAdmissionNeverScansLegacy(t *testing.T) {
	for _, n := range []int{8, 16} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
			var ops []*queueRuntime
			for i := 0; i < n; i++ {
				_, q, err := startQueueOperation(context.Background(), t.TempDir(), "head", []string{"merge", "landing"})
				if err != nil {
					t.Fatal(err)
				}
				ops = append(ops, q)
				defer q.finish("")
			}
			spies := t.TempDir()
			calls := filepath.Join(spies, "calls")
			for _, name := range []string{"git", "ps", "pgrep", "lsof"} {
				if err := os.WriteFile(filepath.Join(spies, name), []byte("#!/bin/sh\necho scan >> '"+calls+"'\nexit 1\n"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", spies)
			for _, q := range ops {
				if ok, err := q.admit(context.Background(), nil); err != nil || !ok {
					t.Fatalf("admission: %v %v", ok, err)
				}
			}
			if data, err := os.ReadFile(calls); !os.IsNotExist(err) {
				t.Fatalf("%d admissions inspected legacy processes/topology: %q %v", n, data, err)
			}
		})
	}
}

func TestQueueRuntimeConsumesExactCancellation(t *testing.T) {
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	ctx, q, err := startQueueOperation(context.Background(), t.TempDir(), "head", []string{"suite"})
	if err != nil {
		t.Fatal(err)
	}
	defer q.finish("cancelled")
	if _, err := q.admit(ctx, nil); err != nil {
		t.Fatal(err)
	}
	owned, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := consumeQueueCancellation(owned, q, cancel)
	defer stop()
	if err := q.store.Pause(ctx, "fast", "drain", true); err != nil {
		t.Fatal(err)
	}
	select {
	case <-owned.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("owner did not consume durable intent")
	}
}
