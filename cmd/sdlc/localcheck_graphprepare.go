package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

// prepareAffectedGraph owns generation only for the graph checkout's lifetime.
// Main's declaration is already parsed and pinned by runLocalCheckCmd.
func prepareAffectedGraph(ctx context.Context, d localRunDeps, dir string, p *affected.Preparation) error {
	if p == nil {
		return ctx.Err()
	}
	minute := d.minute
	if minute == 0 {
		minute = time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.TimeoutMinutes)*minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, p.Command[0], p.Command[1:]...)
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = d.Stderr, d.Stderr
	fmt.Fprintf(d.Stderr, "==> affected graph preparation: %q\n", p.Command)
	err := cmd.Run()
	_ = processgroup.Terminate(cmd, true)
	// A command that succeeded stands, even if the deadline passed after it (#599).
	if err != nil && ctx.Err() != nil {
		err = ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("graph preparation: %w", err)
	}
	return nil
}
