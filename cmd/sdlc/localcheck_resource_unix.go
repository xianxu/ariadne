//go:build unix

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/pkg/procgroup"
)

const runnerCancelGrace = 30 * time.Second

// configureRunnerCleanup gives the entire owned group one shared grace budget.
// Wait reaps only the leader; group members can continue cleanup without pipes.
// No detached timer or goroutine outlives the returned join function.
func configureRunnerCleanup(cmd *exec.Cmd) func() error {
	return configureRunnerCleanupWithin(cmd, runnerCancelGrace)
}

func configureRunnerCleanupWithin(cmd *exec.Cmd, grace time.Duration) func() error {
	started := make(chan time.Time, 1)
	cmd.WaitDelay = grace
	cmd.Cancel = func() error {
		select {
		case started <- time.Now():
		default:
		}
		return processgroup.Terminate(cmd, false)
	}
	return func() error {
		if cmd.Process == nil {
			return nil
		}
		gone := func() bool { return errors.Is(procgroup.Kill(cmd.Process.Pid, 0), os.ErrProcessDone) }
		if gone() {
			return nil
		}
		deadline := time.Now().Add(grace)
		select {
		case at := <-started:
			deadline = at.Add(grace)
		default:
			_ = processgroup.Terminate(cmd, false)
		}
		for !gone() && time.Now().Before(deadline) {
			time.Sleep(20 * time.Millisecond)
		}
		if gone() {
			return nil
		}
		if err := processgroup.Terminate(cmd, true); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return fmt.Errorf("runner group cleanup unconfirmed: %w", err)
		}
		return nil
	}
}
