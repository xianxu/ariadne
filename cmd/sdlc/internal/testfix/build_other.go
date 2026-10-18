//go:build !unix

package testfix

import "os/exec"

// Without POSIX process groups only the direct child is killed.
func ownGroup(*exec.Cmd) {}

func killGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
