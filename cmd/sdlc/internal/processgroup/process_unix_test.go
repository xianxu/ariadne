//go:build unix

package processgroup

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/procgroup"
	"github.com/xianxu/ariadne/pkg/procgroup/procgrouptest"
)

func TestCancellationKillsDescendants(t *testing.T) {
	ready := filepath.Join(t.TempDir(), "child")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// The child ignores graceful termination and inherits the output pipes.
	cmd := exec.CommandContext(ctx, "sh", "-c", `trap '' TERM; sleep 60 & child=$!; printf '%s' "$child" > "$1"; wait`, "fixture", ready)
	Configure(cmd)
	cmd.Cancel = func() error { return Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	finished := make(chan struct{})
	go func() {
		done <- cmd.Wait()
		close(finished)
	}()
	t.Cleanup(func() {
		_ = Terminate(cmd, true)
		_ = cmd.Process.Kill() // Emergency cleanup if group configuration regresses.
		select {
		case <-finished:
		case <-time.After(testfix.HangCeiling):
			t.Error("fixture parent did not finish after cleanup")
		}
	})
	child := testfix.AwaitPIDFile(ready, testfix.HangCeiling, nil)
	if child == 0 {
		t.Fatal("child did not become ready")
	}
	t.Cleanup(func() { _ = syscall.Kill(child, syscall.SIGKILL) })
	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Error("cancelled command succeeded")
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatal("cancellation exceeded shutdown bound")
	}
	// Five seconds is the shutdown contract itself (WaitDelay 1 s plus the
	// group kill), not a hang guard; it is the one elapsed bound kept (#325).
	//lint:allow test-tight-wallclock — the 5s shutdown bound is processgroup's contract (WaitDelay 1s plus the group kill), not a hang guard (#325)
	if time.Since(start) >= 5*time.Second {
		t.Error("shutdown exceeded five seconds")
	}
	if err := syscall.Kill(cmd.Process.Pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Errorf("parent not reaped: %v", err)
	}
	// An orphan zombie may await the system reaper, but must not be running.
	if err := testfix.AwaitExit(child, testfix.HangCeiling); err != nil {
		t.Errorf("descendant survived cancellation: %v", err)
	}
}

func TestTerminateUnstarted(t *testing.T) {
	for _, force := range []bool{false, true} {
		if err := Terminate(exec.Command("unused"), force); !errors.Is(err, os.ErrProcessDone) {
			t.Errorf("force=%v: %v", force, err)
		}
	}
}

// #361: a group whose members are all zombies or exiting (macOS killpg: EPERM)
// is done, not a failure to cancel.
func TestTerminateReportsADyingGroupAsDone(t *testing.T) {
	cmd := procgrouptest.ZombieGroup(t)
	for _, force := range []bool{false, true} {
		if err := Terminate(cmd, force); err != nil && !errors.Is(err, os.ErrProcessDone) {
			t.Fatalf("force=%v: %v", force, err)
		}
	}
}

// The kernel's successful group signal can miss a descendant born during its
// traversal. Reproduce that ordering by signalling only the ready parent on
// the first pass, leaving its pipe-holding descendant in the original group.
func TestForcedTerminationRetiresMemberMissedByFirstSignal(t *testing.T) {
	ready := filepath.Join(t.TempDir(), "child")
	cmd := exec.Command("sh", "-c", `sleep 60 & child=$!; printf '%s' "$child" > "$1"; wait`, "fixture", ready)
	Configure(cmd)
	cmd.WaitDelay = time.Second
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	child := testfix.AwaitPIDFile(ready, testfix.HangCeiling, nil)
	if child == 0 {
		cmd.Process.Kill()
		cmd.Wait()
		t.Fatal("child never ready")
	}
	t.Cleanup(func() { _ = syscall.Kill(child, syscall.SIGKILL); _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) })
	calls := 0
	start := time.Now()
	err := terminateGroup(cmd.Process.Pid, syscall.SIGKILL, func(pgid int, sig syscall.Signal) error {
		calls++
		if calls == 1 {
			return syscall.Kill(cmd.Process.Pid, sig)
		}
		return procgroup.Kill(pgid, sig)
	})
	if err != nil && !errors.Is(err, os.ErrProcessDone) {
		t.Errorf("terminate: %v", err)
	}
	_ = cmd.Wait()
	//lint:allow test-tight-wallclock — 700ms separates a group kill from waiting out the holder's 1s drain bound; no event marks it
	if elapsed := time.Since(start); elapsed >= 700*time.Millisecond {
		t.Errorf("pipe holder survived first signal: %v", elapsed)
	}
	if calls < 2 {
		t.Error("successful first signal was treated as group retirement")
	}
	if err := testfix.AwaitExit(child, 100*time.Millisecond); err != nil {
		t.Errorf("missed descendant survived: %v", err)
	}
}

func TestForcedTerminationPreservesSuccessfulCancelResult(t *testing.T) {
	calls := 0
	err := terminateGroup(123, syscall.SIGKILL, func(int, syscall.Signal) error {
		calls++
		if calls == 1 {
			return nil
		}
		return os.ErrProcessDone
	})
	// exec.Cmd.Cancel needs nil once our signal was sent: returning ProcessDone
	// instead would suppress context cancellation when the parent exits zero.
	if err != nil {
		t.Fatalf("successful cancellation became %v", err)
	}
}
