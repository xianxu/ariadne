//go:build unix

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func awaitLandingGitChild(t *testing.T, pidfile string) int {
	t.Helper()
	until := time.Now().Add(testfix.HangCeiling)
	for time.Now().Before(until) {
		raw, _ := os.ReadFile(pidfile)
		pid, _ := strconv.Atoi(string(raw))
		if pid > 0 {
			t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
			return pid
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("git fixture descendant never started")
	return 0
}

func TestBoundedLandingGitTerminatesDescendants(t *testing.T) {
	for _, mode := range []string{"cancel", "deadline", "exited-parent"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			git(t, root, "init", "-q")
			pidfile := filepath.Join(t.TempDir(), "child.pid")
			t.Setenv("SDLC_GIT_CHILD_PID", pidfile)
			alias := `!sleep 60 & printf '%s' "$!" > "$SDLC_GIT_CHILD_PID"; wait`
			if mode == "exited-parent" {
				alias = `!sleep 60 & printf '%s' "$!" > "$SDLC_GIT_CHILD_PID"; exit 0`
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			limit := 5 * time.Second
			if mode == "deadline" {
				limit = 500 * time.Millisecond
			}
			r := newBoundedLandingGit(ctx, execGitRunner{}, limit)
			done := make(chan error, 1)
			go func() { _, err := r.GitInDir(root, "-c", "alias.hang="+alias, "hang"); done <- err }()
			child := awaitLandingGitChild(t, pidfile)
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				if mode == "cancel" && !errors.Is(err, context.Canceled) {
					t.Fatalf("cancel error = %v", err)
				}
				if mode == "deadline" && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("deadline error = %v", err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("git runner exceeded cancellation/drain bound")
			}
			if err := testfix.AwaitExit(child, time.Second); err != nil {
				t.Fatalf("git left live descendant: %v", err)
			}
		})
	}
}

func TestBoundedLandingGitRecoveryUsesFreshContext(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init", "-q")
	ctx, cancel := context.WithCancel(context.Background())
	r := boundedLandingGit(ctx, execGitRunner{})
	cancel()
	if _, err := r.GitInDir(root, "rev-parse", "--git-dir"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled probe = %v", err)
	}
	if _, err := recoveryLandingGit(r).GitInDir(root, "rev-parse", "--git-dir"); err != nil {
		t.Fatalf("fresh recovery probe = %v", err)
	}
}

func TestBoundedLandingGitUncertainPushPreservesPublishedHead(t *testing.T) {
	root, remote := t.TempDir(), t.TempDir()
	git(t, remote, "init", "--bare", "-q")
	git(t, root, "init", "-q")
	git(t, root, "config", "user.name", "fixture")
	git(t, root, "config", "user.email", "fixture@example.com")
	git(t, root, "commit", "--allow-empty", "-qm", "old")
	old := git(t, root, "rev-parse", "HEAD")
	git(t, root, "push", "-q", remote, "HEAD:refs/heads/issue")
	git(t, root, "commit", "--allow-empty", "-qm", "integration")
	head := git(t, root, "rev-parse", "HEAD")
	pidfile := filepath.Join(t.TempDir(), "child.pid")
	t.Setenv("SDLC_GIT_CHILD_PID", pidfile)
	hook := "#!/bin/sh\nsleep 60 & printf '%s' \"$!\" > \"$SDLC_GIT_CHILD_PID\"; wait\n"
	if err := os.WriteFile(filepath.Join(remote, "hooks", "post-receive"), []byte(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := boundedLandingGit(ctx, execGitRunner{})
	done := make(chan error, 1)
	go func() {
		_, err := r.GitInDir(root, "push", "--force-with-lease=refs/heads/issue:"+old, remote, head+":refs/heads/issue")
		done <- err
	}()
	child := awaitLandingGitChild(t, pidfile)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("uncertain push error = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("push did not return on cancellation")
	}
	if err := testfix.AwaitExit(child, time.Second); err != nil {
		t.Fatal(err)
	}
	// Post-receive runs after publication. Cancellation cannot establish
	// refusal: retain the local head and observe the remote with fresh IO.
	recovery := recoveryLandingGit(r)
	got, err := recovery.GitInDir(remote, "rev-parse", "refs/heads/issue")
	if err != nil || string(got) != head+"\n" {
		t.Fatalf("published remote = %q, error = %v", got, err)
	}
	got, err = recovery.GitInDir(root, "rev-parse", "HEAD")
	if err != nil || string(got) != head+"\n" {
		t.Fatalf("retained local = %q, error = %v", got, err)
	}
}
