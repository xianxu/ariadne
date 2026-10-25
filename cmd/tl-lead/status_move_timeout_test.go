package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/testgate"
)

// The helper is a real subprocess so deadline/exit handling exercises os/exec.
func TestMoveCouchHelperProcess(t *testing.T) {
	mode := os.Getenv("TL_LEAD_MOVE_HELPER")
	if mode == "" {
		return
	}
	switch mode {
	case "hang":
		// Unresponsive input, not a synchronization delay. The parent kills it.
		for {
			time.Sleep(time.Hour)
		}
	case "rename-hang":
		if gate := os.Getenv("TL_LEAD_MOVE_GATE"); gate != "" {
			if err := os.WriteFile(gate+".ready", nil, 0600); err != nil {
				os.Exit(2)
			}
			awaitMoveEffect(t, t.Context(), gate, nil)
		}
		if err := os.WriteFile(os.Getenv("TL_LEAD_MOVE_EFFECT"), []byte("renamed"), 0600); err != nil {
			os.Exit(2)
		}
		// The parent observes the effect before cancelling. Never return on a
		// timer: slow scheduling must not turn interruption into success.
		for {
			time.Sleep(time.Hour)
		}
	case "pipe-parent":
		child := exec.Command(os.Args[0], "-test.run=^TestMoveCouchHelperProcess$")
		child.Env = append(os.Environ(), "TL_LEAD_MOVE_HELPER=pipe-child")
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		awaitMoveEffect(t, t.Context(), os.Getenv("TL_LEAD_MOVE_PIPE")+".ready", nil)
	case "pipe-child":
		gate := os.Getenv("TL_LEAD_MOVE_PIPE")
		if err := os.WriteFile(gate+".ready", nil, 0600); err != nil {
			os.Exit(2)
		}
		for {
			if _, err := os.Stat(gate); err == nil {
				break
			}
			// A failed parent assertion still removes this t.TempDir.
			if _, err := os.Stat(filepath.Dir(gate)); err != nil {
				os.Exit(0)
			}
			time.Sleep(5 * time.Millisecond)
		}
		if err := os.WriteFile(gate+".done", nil, 0600); err != nil {
			os.Exit(2)
		}
	case "locked":
		os.Exit(3)
	case "stale":
		os.Exit(6)
	case "permission":
		fmt.Fprintln(os.Stderr, "connect: operation not permitted")
		os.Exit(1)
	}
	os.Exit(0)
}

func moveCouchHelper(t *testing.T, mode string) string {
	t.Helper()
	t.Setenv("TL_LEAD_MOVE_HELPER", mode)
	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(t.TempDir(), "couch")
	body := "#!/bin/sh\nexec '" + strings.ReplaceAll(bin, "'", "'\\''") + "' -test.run='^TestMoveCouchHelperProcess$' -- \"$@\"\n"
	if err := os.WriteFile(script, []byte(body), 0700); err != nil {
		t.Fatal(err)
	}
	return script
}

func TestMoveCouchDeadline(t *testing.T) {
	//lint:allow test-shrunk-deadline — asserts only the real elapsed bound, no helper startup effect
	c := couchSessionSync{bin: moveCouchHelper(t, "hang"), timeout: 100 * time.Millisecond}
	var err error
	// The helper hangs forever, so only the deadline lets run return (#622).
	testgate.Returns(t, "the deadline-bounded couch call", func() { _, err = c.run("list", "--json") })
	for _, want := range []string{"cannot reach Couch", "list", "100ms", "outside", "sandbox"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("error %v missing %q", err, want)
		}
	}
}

func TestMoveCouchRenameDeadlineUnconfirmed(t *testing.T) {
	for _, beforeEffect := range []bool{false, true} {
		name := "after effect"
		if beforeEffect {
			name = "before effect"
		}
		t.Run(name, func(t *testing.T) {
			bin := moveCouchHelper(t, "rename-hang")
			root := t.TempDir()
			effect, gate := filepath.Join(root, "effect"), filepath.Join(root, "gate")
			t.Setenv("TL_LEAD_MOVE_EFFECT", effect)
			t.Setenv("TL_LEAD_MOVE_GATE", gate)
			ctx, cancel := moveTestContext(t)
			c := couchSessionSync{bin: bin, ctx: ctx}
			result := make(chan error, 1)
			done := make(chan struct{})
			go func() { defer close(done); result <- c.Rename("old", "new") }()
			t.Cleanup(func() { cancel(); <-done })
			// Hold the child before its effect, independently of startup latency.
			awaitMoveEffect(t, ctx, gate+".ready", result)
			if _, err := os.Stat(effect); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("effect before release: %v", err)
			}
			if !beforeEffect {
				if err := os.WriteFile(gate, nil, 0600); err != nil {
					t.Fatal(err)
				}
				awaitMoveEffect(t, ctx, effect, result)
			}
			cancel()
			err := <-result
			if err == nil || !strings.Contains(err.Error(), "outcome unconfirmed") || !strings.Contains(err.Error(), "context canceled") {
				t.Fatalf("rename error = %v", err)
			}
			_, err = os.Stat(effect)
			if beforeEffect && !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("cancelled gated child published effect: %v", err)
			}
			if !beforeEffect && err != nil {
				t.Fatalf("published effect lost: %v", err)
			}
		})
	}
}

// Startup is bounded by the enclosing test, never an assumed launch latency.
// Cleanup cancellation reaps the child even after an assertion failure.
func moveTestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	var ctx context.Context
	var cancel context.CancelFunc
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(context.Background(), deadline)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	t.Cleanup(cancel)
	return ctx, cancel
}

func awaitMoveEffect(t *testing.T, ctx context.Context, path string, result <-chan error) {
	t.Helper()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		if _, err := os.Stat(path); err == nil {
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
		select {
		case err := <-result:
			t.Fatalf("helper ended before publishing readiness: %v", err)
		case <-ctx.Done():
			t.Fatalf("helper not ready before test deadline: %v", ctx.Err())
		case <-tick.C:
		}
	}
}

func TestMoveCouchKnownRefusals(t *testing.T) {
	for _, tc := range []struct {
		mode string
		want error
	}{{"locked", ErrLocked}, {"stale", ErrStale}} {
		t.Run(tc.mode, func(t *testing.T) {
			ctx, _ := moveTestContext(t)
			c := couchSessionSync{bin: moveCouchHelper(t, tc.mode), ctx: ctx}
			if err := c.Rename("old", "new"); !errors.Is(err, tc.want) {
				t.Fatalf("got %v; want %v", err, tc.want)
			}
		})
	}
}

func TestMoveCouchPermissionFailsClearly(t *testing.T) {
	ctx, _ := moveTestContext(t)
	c := couchSessionSync{bin: moveCouchHelper(t, "permission"), ctx: ctx}
	_, err := c.run("list", "--json")
	if err == nil || !strings.Contains(err.Error(), "cannot reach Couch") || !strings.Contains(err.Error(), "operation not permitted") {
		t.Fatalf("error = %v", err)
	}
}

func TestMoveCouchBoundIsScoped(t *testing.T) {
	original := couchSessionSync{bin: "couch"}
	bounded, ok := boundedMoveSync(original).(couchSessionSync)
	if !ok || bounded.timeout != 5*time.Second || original.timeout != 0 {
		t.Fatalf("bounded = %#v, original = %#v", bounded, original)
	}
	fake := newFake(nil)
	if got := boundedMoveSync(fake); got != fake {
		t.Fatal("stateful fake replaced")
	}
}

func TestMoveCouchInheritedPipeIsBounded(t *testing.T) {
	gate := filepath.Join(t.TempDir(), "pipe")
	t.Setenv("TL_LEAD_MOVE_PIPE", gate)
	ctx, _ := moveTestContext(t)
	t.Cleanup(func() {
		if err := os.WriteFile(gate, nil, 0600); err != nil {
			t.Error(err)
			return
		}
		awaitMoveEffect(t, ctx, gate+".done", nil)
	})
	c := couchSessionSync{bin: moveCouchHelper(t, "pipe-parent"), ctx: ctx}
	_, err := c.run("list", "--json")
	if err == nil || !strings.Contains(err.Error(), "cannot reach Couch") || !strings.Contains(err.Error(), "WaitDelay") {
		t.Fatalf("inherited pipe error = %v", err)
	}
}

func TestMoveCouchUnreachableHasNoMigrationEffects(t *testing.T) {
	root, old, _ := sessionMoveFixture(t)
	c := couchSessionSync{bin: moveCouchHelper(t, "permission")}
	err := moveSession(deps{ss: c}, &bytes.Buffer{}, root, old, true, nil)
	if err == nil || !strings.Contains(err.Error(), "cannot reach Couch") {
		t.Fatal(err)
	}
	for _, rel := range []string{".gitignore", "session.md", "session.side"} {
		if _, err := os.Stat(filepath.Join(root, rel)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("preflight published %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, old)); err != nil {
		t.Fatal("source removed")
	}
}
