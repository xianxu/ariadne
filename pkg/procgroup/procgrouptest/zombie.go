//go:build unix

// Package procgrouptest builds process groups in the states procgroup reports
// on, for tests in packages that depend on those states (#361).
package procgrouptest

import (
	"errors"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"
)

// ZombieGroup starts a process group whose only member exits and stays
// unreaped, and returns its command (the group leader; its pid is the pgid). On macOS it waits until the kernel answers
// EPERM for the group (killpg skips zombies), the state #361 is about;
// elsewhere a zombie still accepts signals, so it waits a short while. The
// member is reaped at cleanup.
func ZombieGroup(t testing.TB) *exec.Cmd {
	t.Helper()
	cmd := exec.Command("sh", "-c", "exit 0")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Wait() })
	pgid := cmd.Process.Pid
	if runtime.GOOS != "darwin" {
		time.Sleep(200 * time.Millisecond)
		return cmd
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		// Signal 0 probes without reaping: EPERM once the member is a zombie.
		if err := syscall.Kill(-pgid, 0); errors.Is(err, syscall.EPERM) {
			return cmd
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("group %d never became zombie-only", pgid)
	return nil
}
