//go:build !unix

package processgroup

import (
	"os"
	"os/exec"
)

// Configure is a no-op on platforms without POSIX process groups.
func Configure(cmd *exec.Cmd) {}

// Terminate retains direct-child cleanup on platforms without process groups.
// Callers should use WaitDelay to bound inherited output-pipe draining.
func Terminate(cmd *exec.Cmd, force bool) error {
	if cmd.Process == nil {
		return os.ErrProcessDone
	}
	if force {
		return cmd.Process.Kill()
	}
	return cmd.Process.Signal(os.Interrupt)
}
