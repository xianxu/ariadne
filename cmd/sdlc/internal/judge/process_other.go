//go:build !unix

package judge

import (
	"os"
	"os/exec"
)

// Platforms without POSIX process groups retain direct-child cleanup. The
// portable WaitDelay still bounds draining inherited output descriptors.
func configureReviewProcess(cmd *exec.Cmd) {}

func terminateReviewProcess(cmd *exec.Cmd, force bool) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	if force {
		return cmd.Process.Kill()
	}
	return cmd.Process.Signal(os.Interrupt)
}
