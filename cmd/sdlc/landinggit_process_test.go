//go:build unix

package main

import (
	"context"
	"errors"
	"github.com/xianxu/ariadne/pkg/testgate"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func awaitLandingGitChild(t *testing.T, pidfile string) int {
	t.Helper()
	pid := testfix.AwaitPIDFile(pidfile, testfix.HangCeiling, nil)
	if pid == 0 {
		t.Fatal("git fixture descendant never started")
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	return pid
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
			// The deadline passes on cue once the descendant has reported itself
			// (#434): a clock would race the fixture's startup under load.
			var expire func()
			if mode == "deadline" {
				ctx, expire = testfix.DeadlineOnCue(ctx)
			}
			r := newBoundedLandingGit(ctx, execGitRunner{}, 5*time.Second)
			done := make(chan error, 1)
			go func() { _, err := r.GitInDir(root, "-c", "alias.hang="+alias, "hang"); done <- err }()
			child := awaitLandingGitChild(t, pidfile)
			switch mode {
			case "cancel":
				cancel()
			case "deadline":
				expire()
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
	type valueKey struct{}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), valueKey{}, "command scope"))
	r := boundedLandingGit(ctx, execGitRunner{})
	cancel()
	if _, err := r.GitInDir(root, "rev-parse", "--git-dir"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled probe = %v", err)
	}
	recovery := recoveryLandingGit(r)
	if got := recovery.(boundedLandingGitRunner).ctx.Value(valueKey{}); got != "command scope" {
		t.Fatalf("recovery dropped command scope: %v", got)
	}
	if _, err := recovery.GitInDir(root, "rev-parse", "--git-dir"); err != nil {
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

// A local check's cleanup Git outlives the run's cancellation but not its own
// budget (#434: asserted here directly, so the signal test need not race a
// shrunk deadline against fixture startup).
func TestLocalCheckCleanupGitIsIndependentAndBounded(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init", "-q")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := localCheckCleanupGit(ctx, execGitRunner{}, root)("rev-parse", "--git-dir"); err != nil {
		t.Fatalf("cleanup Git refused after the run's cancellation: %v", err)
	}
	prev := localCheckCleanupTimeout
	//lint:allow test-shrunk-deadline — asserts only that the hung cleanup Git is bounded (#434); no readiness is awaited
	localCheckCleanupTimeout = 200 * time.Millisecond
	t.Cleanup(func() { localCheckCleanupTimeout = prev })
	// The alias outlives the hang ceiling unkilled, so only the budget lets
	// the call return (#622).
	var err error
	testgate.Returns(t, "the budgeted cleanup git", func() {
		_, err = localCheckCleanupGit(ctx, execGitRunner{}, root)("-c", "alias.hang=!sleep 600", "hang")
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("hung cleanup Git: %v", err)
	}
}
