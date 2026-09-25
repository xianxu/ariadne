//go:build unix

package processgroup

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
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
		case <-time.After(2 * time.Second):
			t.Error("fixture parent did not finish after cleanup")
		}
	})
	deadline := time.Now().Add(3 * time.Second)
	child := 0
	for time.Now().Before(deadline) {
		raw, _ := os.ReadFile(ready)
		child, _ = strconv.Atoi(string(raw))
		if child > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
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
	case <-time.After(4 * time.Second):
		t.Fatal("cancellation exceeded shutdown bound")
	}
	if time.Since(start) >= 5*time.Second {
		t.Error("shutdown exceeded five seconds")
	}
	if err := syscall.Kill(cmd.Process.Pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Errorf("parent not reaped: %v", err)
	}
	// An orphan zombie may await the system reaper, but must not be running.
	state, err := exec.Command("ps", "-o", "stat=", "-p", strconv.Itoa(child)).Output()
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			t.Fatalf("inspect child: %v", err)
		}
	}
	if stat := strings.TrimSpace(string(state)); stat != "" && !strings.HasPrefix(stat, "Z") {
		t.Errorf("descendant survived cancellation: %s", stat)
	}
}

func TestTerminateUnstarted(t *testing.T) {
	for _, force := range []bool{false, true} {
		if err := Terminate(exec.Command("unused"), force); !errors.Is(err, os.ErrProcessDone) {
			t.Errorf("force=%v: %v", force, err)
		}
	}
}
