//go:build unix

package judge

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Reexec the test binary, never a real agent. Both processes intentionally
// ignore graceful shutdown; the descendant keeps inherited output pipes open.
func TestJudgeControlledProcess(t *testing.T) {
	mode := os.Getenv("JUDGE_TEST_PROCESS")
	if mode == "" {
		return
	}
	signal.Ignore(syscall.SIGTERM, syscall.SIGINT)
	if mode == "parent" || mode == "exiting-parent" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestJudgeControlledProcess$")
		cmd.Env = append(os.Environ(), "JUDGE_TEST_PROCESS=descendant")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			os.Exit(2)
		}
	} else {
		if err := os.WriteFile(os.Getenv("JUDGE_TEST_READY"), []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
			os.Exit(3)
		}
	}
	if mode == "exiting-parent" {
		deadline := time.Now().Add(testfix.HangCeiling)
		for time.Now().Before(deadline) {
			if raw, err := os.ReadFile(os.Getenv("JUDGE_TEST_READY")); err == nil && len(raw) > 0 {
				os.Exit(0)
			}
			time.Sleep(time.Millisecond)
		}
		os.Exit(4)
	}
	fmt.Fprintln(os.Stdout, "SHIP")
	for {
		time.Sleep(time.Hour)
	}
}

func TestReviewProcessCancellationKillsGroupAndReaps(t *testing.T) {
	oldGrace := reviewShutdownGrace
	reviewShutdownGrace = 50 * time.Millisecond
	t.Cleanup(func() { reviewShutdownGrace = oldGrace })
	ready := filepath.Join(t.TempDir(), "ready")
	t.Setenv("GORACE", os.Getenv("GORACE")+" atexit_sleep_ms=0")
	t.Setenv("JUDGE_TEST_PROCESS", "parent")
	t.Setenv("JUDGE_TEST_READY", ready)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	started := make(chan int, 1)
	go func() {
		_, err := Run(ctx, func(pid int) { started <- pid }, os.Args[0], "-test.run=^TestJudgeControlledProcess$")
		done <- err
	}()
	parent := <-started
	child := waitReviewChild(t, ready, parent)
	// Emergency cleanup also makes the regression safe against the broken runner.
	defer syscall.Kill(parent, syscall.SIGKILL)
	defer syscall.Kill(child, syscall.SIGKILL)
	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Error("cancelled process reported success")
		}
	case <-time.After(testfix.HangCeiling):
		t.Error("review cancellation left inherited pipes/processes alive beyond shutdown bound")
		_ = syscall.Kill(child, syscall.SIGKILL)
		_ = syscall.Kill(parent, syscall.SIGKILL)
		select {
		case <-done:
		case <-time.After(testfix.HangCeiling):
			t.Fatal("runner did not return after emergency cleanup")
		}
	}
	// Under the production grace proves the injected 50 ms one governed; a
	// tighter bound measures the machine's load, not the runner (#325).
	if elapsed := time.Since(start); elapsed >= oldGrace {
		t.Errorf("shutdown took %v, not bounded by the injected grace", elapsed)
	}
	if err := syscall.Kill(parent, 0); err != syscall.ESRCH {
		t.Errorf("direct reviewer was not reaped: %v", err)
	}
	// A dead orphan may briefly remain a zombie until its system reaper collects it.
	// It must stop executing (and holding the pipe) once Run has returned.
	if err := testfix.AwaitExit(child, testfix.HangCeiling); err != nil {
		t.Errorf("descendant survived shutdown: %v", err)
	}
}

func waitReviewChild(t *testing.T, path string, parent int) int {
	t.Helper()
	deadline := time.NewTimer(testfix.HangCeiling)
	defer deadline.Stop()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-deadline.C:
			_ = syscall.Kill(parent, syscall.SIGKILL)
			t.Fatal("controlled reviewer did not become ready")
		case <-tick.C:
			if raw, err := os.ReadFile(path); err == nil && len(raw) > 0 {
				pid, err := strconv.Atoi(string(raw))
				if err != nil {
					t.Fatal(err)
				}
				return pid
			}
		}
	}
}

func TestReviewProcessBoundsPipeDrainAfterNormalExit(t *testing.T) {
	oldGrace := reviewShutdownGrace
	reviewShutdownGrace = 50 * time.Millisecond
	t.Cleanup(func() { reviewShutdownGrace = oldGrace })
	ready := filepath.Join(t.TempDir(), "ready")
	t.Setenv("GORACE", os.Getenv("GORACE")+" atexit_sleep_ms=0")
	t.Setenv("JUDGE_TEST_PROCESS", "exiting-parent")
	t.Setenv("JUDGE_TEST_READY", ready)
	// ErrWaitDelay is the evidence that the drain bound fired; the descendant
	// never closes its pipes, so without it Run would hang to this ceiling. The
	// window includes two process starts, so no elapsed bound (#325).
	ctx, cancel := context.WithTimeout(context.Background(), testfix.HangCeiling)
	defer cancel()
	_, err := Run(ctx, nil, os.Args[0], "-test.run=^TestJudgeControlledProcess$")
	if err != exec.ErrWaitDelay {
		t.Errorf("unclosed inherited pipes: err=%v, want ErrWaitDelay", err)
	}
	raw, readErr := os.ReadFile(ready)
	if readErr != nil {
		t.Fatal(readErr)
	}
	child, parseErr := strconv.Atoi(string(raw))
	if parseErr != nil {
		t.Fatal(parseErr)
	}
	defer syscall.Kill(child, syscall.SIGKILL)
	if err := testfix.AwaitExit(child, testfix.HangCeiling); err != nil {
		t.Errorf("descendant survived pipe-drain cleanup: %v", err)
	}
}
