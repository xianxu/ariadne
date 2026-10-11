package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/repolock"
)

const repoLockAnnotation = "ariadne.sdlc.repo-lock"
const repoLockWrappedAnnotation = "ariadne.sdlc.repo-lock-wrapped"

type repoLockMode string

const (
	repoLockAuto   repoLockMode = "auto"
	repoLockManual repoLockMode = "manual"
)

type repoLockContextKey struct{}

var repoLockAcquireForCommand = acquireRepoLockForCommand

func markMutatingCommand(cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[repoLockAnnotation] = string(repoLockAuto)
	return cmd
}

func markManualLockCommand(cmd *cobra.Command) *cobra.Command {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[repoLockAnnotation] = string(repoLockManual)
	return cmd
}

func commandNeedsRepoLock(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	mode := repoLockMode(cmd.Annotations[repoLockAnnotation])
	return mode == repoLockAuto || mode == repoLockManual
}

func commandAutoWrapsRepoLock(cmd *cobra.Command) bool {
	if cmd == nil {
		return false
	}
	return repoLockMode(cmd.Annotations[repoLockAnnotation]) == repoLockAuto
}

func wrapRepoLockCommands(root *cobra.Command) {
	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		if commandAutoWrapsRepoLock(cmd) && cmd.RunE != nil && cmd.Annotations[repoLockWrappedAnnotation] != "true" {
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
	return withRequiredRepoTransactionLock(cmd, run)
}

func withRequiredRepoTransactionLock(cmd *cobra.Command, run func() error) error {
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
		GitCommonDir:         gitDir,
		CallerHandlesSignals: cmd.Context().Value(cliSignalContextKey{}) == true,
		Command:              cmd.CommandPath(),
		Args:                 os.Args,
		Hostname:             host,
		PID:                  os.Getpid(),
		CWD:                  cwd,
		ProcessAlive:         processAlive,
		Stderr:               cmd.ErrOrStderr(),
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

// processAlive reports whether pid exists. EPERM means it exists but we may
// not signal it: another user's process, or one outside the Seatbelt sandbox
// slot agents run in. Only ESRCH means gone; reading EPERM as dead let a slot
// reclaim a live peer's lock (side-quest of #310).
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
