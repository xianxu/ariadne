//go:build unix

package testfix

import (
	"os/exec"
	"testing"
	"time"
)

func TestAwaitExit(t *testing.T) {
	sleeper := exec.Command("sleep", "60")
	if err := sleeper.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sleeper.Process.Kill(); _ = sleeper.Wait() })
	if err := AwaitExit(sleeper.Process.Pid, 50*time.Millisecond); err == nil {
		t.Fatal("live sleeper reported exited")
	}

	// An exited, unreaped child is a zombie: it exists but no longer runs.
	exited := exec.Command("true")
	if err := exited.Start(); err != nil {
		t.Fatal(err)
	}
	if err := AwaitExit(exited.Process.Pid, 10*time.Second); err != nil {
		t.Fatalf("zombie: %v", err)
	}
	if err := exited.Wait(); err != nil {
		t.Fatal(err)
	}
	if err := AwaitExit(exited.Process.Pid, 0); err != nil {
		t.Fatalf("reaped: %v", err)
	}
}
