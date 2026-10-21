//go:build !unix

package main

import (
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"os/exec"
	"time"
)

const runnerCancelGrace = 30 * time.Second

func configureRunnerCleanup(cmd *exec.Cmd) func() error {
	cmd.WaitDelay = runnerCancelGrace
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, false) }
	return func() error { return processgroup.Terminate(cmd, true) }
}
