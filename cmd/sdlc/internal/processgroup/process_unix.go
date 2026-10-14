//go:build unix

// Package processgroup configures and terminates subprocess groups. Callers own
// shutdown timing and must still Wait to reap the direct child.
package processgroup

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/pkg/procgroup"
)

// Configure places the child in its own group. Call before Start.
func Configure(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// Terminate signals the entire configured group, including descendants.
// force selects SIGKILL instead of graceful SIGTERM.
func Terminate(cmd *exec.Cmd, force bool) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	signal := syscall.SIGTERM
	if force {
		signal = syscall.SIGKILL
	}
	// procgroup reads the kernel's answer: an empty group, or (macOS) one
	// whose members are all zombies or exiting, is done (#361).
	return terminateGroup(cmd.Process.Pid, signal, procgroup.Kill)
}

func terminateGroup(pgid int, signal syscall.Signal, kill func(int, syscall.Signal) error) error {
	if err := kill(pgid, signal); err != nil || signal != syscall.SIGKILL {
		return err
	}
	// A successful killpg is one traversal, not a fork barrier: a child born
	// during that traversal can still hold our output pipes. Forceful teardown
	// repeats until the kernel reports no runnable members. Bound the retry
	// work to 100ms (inside the probe's existing 700ms cancellation contract);
	// Linux zombie-only groups may accept signals until their reaper runs.
	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
		err := kill(pgid, signal)
		if errors.Is(err, os.ErrProcessDone) {
			// A signal was sent: preserve exec.Cmd.Cancel's nil result so
			// context cancellation cannot become a successful command.
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}
