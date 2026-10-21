package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The helper is a real subprocess so deadline/exit handling exercises os/exec.
func TestMoveCouchHelperProcess(t *testing.T) {
	mode := os.Getenv("TL_LEAD_MOVE_HELPER")
	if mode == "" {
		return
	}
	switch mode {
	case "hang":
		//lint:allow test-fixed-sleep — this helper simulates an unresponsive external process; the parent deadline, not this delay, synchronizes the test
		time.Sleep(700 * time.Millisecond)
	case "rename-hang":
		if err := os.WriteFile(os.Getenv("TL_LEAD_MOVE_EFFECT"), []byte("renamed"), 0600); err != nil {
			os.Exit(2)
		}
		//lint:allow test-fixed-sleep — this helper simulates an unresponsive external process; the parent deadline, not this delay, synchronizes the test
		time.Sleep(700 * time.Millisecond)
	case "pipe-parent":
		child := exec.Command(os.Args[0], "-test.run=^TestMoveCouchHelperProcess$")
		child.Env = append(os.Environ(), "TL_LEAD_MOVE_HELPER=hang")
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
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
	c := couchSessionSync{bin: moveCouchHelper(t, "hang"), timeout: 100 * time.Millisecond}
	start := time.Now()
	_, err := c.run("list", "--json")
	for _, want := range []string{"cannot reach Couch", "list", "100ms", "outside", "sandbox"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("error %v missing %q", err, want)
		}
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("deadline did not bound child")
	}
}

func TestMoveCouchRenameDeadlineUnconfirmed(t *testing.T) {
	bin := moveCouchHelper(t, "rename-hang")
	effect := filepath.Join(t.TempDir(), "effect")
	t.Setenv("TL_LEAD_MOVE_EFFECT", effect)
	c := couchSessionSync{bin: bin, timeout: 300 * time.Millisecond}
	err := c.Rename("old", "new")
	if err == nil || !strings.Contains(err.Error(), "outcome unconfirmed") {
		t.Fatalf("rename error = %v", err)
	}
	if _, err := os.Stat(effect); err != nil {
		t.Fatalf("rename never published: %v", err)
	}
}

func TestMoveCouchKnownRefusals(t *testing.T) {
	for _, tc := range []struct {
		mode string
		want error
	}{{"locked", ErrLocked}, {"stale", ErrStale}} {
		t.Run(tc.mode, func(t *testing.T) {
			c := couchSessionSync{bin: moveCouchHelper(t, tc.mode), timeout: time.Second}
			if err := c.Rename("old", "new"); !errors.Is(err, tc.want) {
				t.Fatalf("got %v; want %v", err, tc.want)
			}
		})
	}
}

func TestMoveCouchPermissionFailsClearly(t *testing.T) {
	c := couchSessionSync{bin: moveCouchHelper(t, "permission"), timeout: time.Second}
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
	c := couchSessionSync{bin: moveCouchHelper(t, "pipe-parent"), timeout: 2 * time.Second}
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
