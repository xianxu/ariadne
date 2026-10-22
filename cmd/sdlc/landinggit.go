package main

import (
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

const landingGitTimeout = 2 * time.Minute
const landingRecoveryTimeout = 30 * time.Second

// Every Git command in durable landing, including transport, hooks, merge
// drivers, probes and checkout cleanup, has a per-command ceiling. Suite time
// is owned separately by the local-check runner. Recovery uses fresh bounded
// IO because the operation context may already have expired.
type boundedLandingGitRunner struct {
	gitRunner
	ctx     context.Context
	timeout time.Duration
}

func boundedLandingGit(ctx context.Context, r gitRunner) gitRunner {
	return newBoundedLandingGit(ctx, r, landingGitTimeout)
}

func newBoundedLandingGit(ctx context.Context, r gitRunner, timeout time.Duration) gitRunner {
	ctx = commandContext(ctx)
	if old, ok := r.(boundedLandingGitRunner); ok {
		r = old.gitRunner
	}
	return boundedLandingGitRunner{gitRunner: r, ctx: ctx, timeout: timeout}
}

// withLandingGitContext lets polling impose its smaller remaining IO budget
// while preserving the command ceiling and the injected Git seam.
func withLandingGitContext(ctx context.Context, r gitRunner) gitRunner {
	limit := landingGitTimeout
	if old, ok := r.(boundedLandingGitRunner); ok {
		limit = old.timeout
	}
	return newBoundedLandingGit(ctx, r, limit)
}

func recoveryLandingGit(r gitRunner) gitRunner {
	ctx := commandContext(nil)
	if old, ok := r.(boundedLandingGitRunner); ok {
		ctx = context.WithoutCancel(old.ctx)
	}
	return newBoundedLandingGit(ctx, r, landingRecoveryTimeout)
}

func (r boundedLandingGitRunner) Git(args ...string) ([]byte, error) {
	return r.run("", true, args...)
}

func (r boundedLandingGitRunner) GitInDir(dir string, args ...string) ([]byte, error) {
	return r.run(dir, false, args...)
}

func (r boundedLandingGitRunner) run(dir string, currentDir bool, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(r.ctx, r.timeout)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	switch r.gitRunner.(type) {
	case execGitRunner, *execGitRunner:
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = dir
		processgroup.Configure(cmd)
		cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
		cmd.WaitDelay = time.Second
		out, diagnostic, err := runGitCmdRaw(cmd)
		// WaitDelay closes inherited pipes; it does not kill their owners.
		// Also reclaim descendants whose parent exited successfully.
		_ = processgroup.Terminate(cmd, true)
		if ctx.Err() != nil {
			err = errors.Join(ctx.Err(), err)
		}
		if err != nil {
			return append(out, diagnostic...), gitx.ClassifyNetworkFailure(args, diagnostic, err)
		}
		return out, nil
	default:
		// This injected seam supplies raw results, including raw process causes;
		// it must not preclassify errors before this owner's final context check.
		// All production callers use the concrete subprocess path above (nested
		// bounded runners are unwrapped by the constructor).
		var out []byte
		var err error
		if currentDir {
			out, err = r.gitRunner.Git(args...)
		} else {
			out, err = r.gitRunner.GitInDir(dir, args...)
		}
		if ctx.Err() != nil {
			err = errors.Join(ctx.Err(), err)
		}
		return out, err
	}
}
