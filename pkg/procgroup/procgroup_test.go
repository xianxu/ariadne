//go:build unix

package procgroup

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/procgroup/procgrouptest"
)

// A group whose only member has exited but is not yet reaped (macOS: EPERM;
// Linux: the zombie accepts the signal) and a reaped group (ESRCH) both read as
// done, and neither is an error a caller must stop on.
func TestKillReportsADyingGroupAsDone(t *testing.T) {
	pgid := procgrouptest.ZombieGroup(t).Process.Pid
	if runtime.GOOS == "darwin" {
		if raw := syscall.Kill(-pgid, syscall.SIGKILL); !errors.Is(raw, syscall.EPERM) {
			t.Fatalf("precondition: the kernel answered %v, not EPERM, for a zombie-only group", raw)
		}
	}
	if err := Kill(pgid, syscall.SIGKILL); err != nil && !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("zombie-only group: %v", err)
	}
}

func TestKillReportsAnEmptyGroupAsDone(t *testing.T) {
	cmd := exec.Command("sh", "-c", "exit 0")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Run(); err != nil { // reaped: the group is empty
		t.Fatal(err)
	}
	if err := Kill(cmd.Process.Pid, syscall.SIGKILL); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("reaped group: %v", err)
	}
}

func TestKillSignalsALiveGroup(t *testing.T) {
	cmd := exec.Command("sh", "-c", "sleep 30 & wait")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if err := Kill(cmd.Process.Pid, syscall.SIGKILL); err != nil {
		t.Fatalf("live group: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the group survived SIGKILL")
	}
}
