//go:build unix

// Package processgroup configures and terminates subprocess groups. Callers own
// shutdown timing and must still Wait to reap the direct child.
package processgroup

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
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
	err := syscall.Kill(-cmd.Process.Pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}
