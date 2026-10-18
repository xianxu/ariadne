//go:build unix

package testfix

import (
	"os/exec"
	"syscall"
)

// ownGroup starts cmd in its own process group, so killGroup reaches its
// descendants too.
func ownGroup(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} }

func killGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
