package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/repolock"
)

const repoLockAnnotation = "ariadne.sdlc.repo-lock"
const repoLockWrappedAnnotation = "ariadne.sdlc.repo-lock-wrapped"

type repoLockContextKey struct{}

var repoLockAcquireForCommand = acquireRepoLockForCommand

func markMutatingCommand(cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[repoLockAnnotation] = "true"
	return cmd
}

func commandNeedsRepoLock(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	return cmd.Annotations[repoLockAnnotation] == "true"
}

func wrapRepoLockCommands(root *cobra.Command) {
	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd.RunE != nil && cmd.Annotations[repoLockWrappedAnnotation] != "true" {
			orig := cmd.RunE
			cmd.RunE = func(c *cobra.Command, args []string) error {
				return withRepoTransactionLock(c, func() error {
					return orig(c, args)
				})
			}
			if cmd.Annotations == nil {
				cmd.Annotations = map[string]string{}
			}
			cmd.Annotations[repoLockWrappedAnnotation] = "true"
		}
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)
}

func withRepoTransactionLock(cmd *cobra.Command, run func() error) error {
	if !commandNeedsRepoLock(cmd) {
		return run()
	}
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	if held, _ := ctx.Value(repoLockContextKey{}).(bool); held {
		return run()
	}
	release, err := repoLockAcquireForCommand(cmd)
	if err != nil {
		return err
	}
	var releaseOnce sync.Once
	releaseFn := func() {
		releaseOnce.Do(func() {
			_ = release()
		})
	}
	unregisterDieCleanup := registerDieCleanup(releaseFn)
	lockedCtx := context.WithValue(ctx, repoLockContextKey{}, true)
	cmd.SetContext(lockedCtx)
	defer cmd.SetContext(ctx)
	defer unregisterDieCleanup()
	defer releaseFn()
	return run()
}

func acquireRepoLockForCommand(cmd *cobra.Command) (func() error, error) {
	gitDir, err := repoLockGitCommonDir()
	if err != nil {
		return nil, err
	}
	host, _ := os.Hostname()
	cwd, _ := os.Getwd()
	lock, err := repolock.Acquire(cmd.Context(), repolock.Options{
		GitCommonDir:  gitDir,
		Command:       cmd.CommandPath(),
		Args:          os.Args,
		Hostname:      host,
		PID:           os.Getpid(),
		CWD:           cwd,
		ProcessAlive:  processAlive,
		Stderr:        cmd.ErrOrStderr(),
		WaitTimeout:   30 * time.Minute,
		StaleDuration: 30 * time.Minute,
	})
	if err != nil {
		return nil, err
	}
	return lock.Release, nil
}

func repoLockGitCommonDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	common := strings.TrimSpace(gitx.Capture("rev-parse", "--git-common-dir"))
	if common == "" {
		top := strings.TrimSpace(gitx.Capture("rev-parse", "--show-toplevel"))
		if top == "" {
			common = ".git"
		} else {
			common = filepath.Join(top, ".git")
		}
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(cwd, common)
	}
	abs, err := filepath.Abs(common)
	if err != nil {
		return "", fmt.Errorf("resolve git common dir %q: %w", common, err)
	}
	return abs, nil
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}
