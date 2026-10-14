//go:build unix

package gitx

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestRebasedReviewedBaseContextCancelsReplayCommands(t *testing.T) {
	for _, command := range []string{"version", "merge-base", "merge-tree", "commit-tree"} {
		t.Run(command, func(t *testing.T) {
			f := patchRepo(t)
			realGit, err := exec.LookPath("git")
			if err != nil {
				t.Fatal(err)
			}
			bin := t.TempDir()
			pidfile, parentfile := filepath.Join(bin, "child.pid"), filepath.Join(bin, "parent.pid")
			t.Setenv("SDLC_GIT_CHILD_PID", pidfile)
			t.Setenv("SDLC_GIT_PARENT_PID", parentfile)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			script := "#!/bin/sh\ncase \" $* \" in\n*' " + command + " '*) printf '%s' \"$$\" > \"$SDLC_GIT_PARENT_PID\"; sleep 60 & printf '%s' \"$!\" > \"$SDLC_GIT_CHILD_PID\"; wait; exit 1;;\nesac\nexec '" + realGit + "' \"$@\"\n"
			if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, _, err := RebasedReviewedBaseContext(ctx, "main", f.h1); done <- err }()
			pid := 0
			until := time.Now().Add(testfix.HangCeiling)
			for time.Now().Before(until) {
				raw, _ := os.ReadFile(pidfile)
				pid, _ = strconv.Atoi(string(raw))
				if pid > 0 {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
			if pid == 0 {
				t.Fatal("replay command never started")
			}
			raw, _ := os.ReadFile(parentfile)
			parent, _ := strconv.Atoi(string(raw))
			t.Cleanup(func() { _ = syscall.Kill(parent, syscall.SIGKILL); _ = syscall.Kill(pid, syscall.SIGKILL) })
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation = %v", err)
				}
			case <-time.After(500 * time.Millisecond):
				_ = syscall.Kill(parent, syscall.SIGKILL)
				_ = syscall.Kill(pid, syscall.SIGKILL)
				<-done
				t.Fatal("replay lost operation cancellation")
			}
			if err := testfix.AwaitExit(pid, time.Second); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRunGitContextRejectsCancelledSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	old := runGitInContext
	runGitInContext = func(context.Context, string, []string, ...string) ([]byte, []byte, error) {
		cancel()
		return []byte("success\n"), nil, nil
	}
	t.Cleanup(func() { runGitInContext = old })
	if _, _, err := RunGitContext(ctx, "rev-parse", "HEAD"); !errors.Is(err, context.Canceled) {
		t.Fatalf("late success = %v", err)
	}
}
