//go:build unix

package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestLandingPublishAndOwnershipReadsHonorCancellation(t *testing.T) {
	for _, mode := range []string{"publish-first-read", "ownership-read", "ownership-workspace", "ownership-fetched-tracker"} {
		t.Run(mode, func(t *testing.T) {
			realGit, err := exec.LookPath("git")
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			git(t, root, "init", "-q")
			if err := os.MkdirAll(filepath.Join(root, "workshop/issues"), 0o755); err != nil {
				t.Fatal(err)
			}
			bin := t.TempDir()
			pidfile, parentfile := filepath.Join(bin, "child.pid"), filepath.Join(bin, "parent.pid")
			t.Setenv("SDLC_GIT_CHILD_PID", pidfile)
			t.Setenv("SDLC_GIT_PARENT_PID", parentfile)
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			condition := "*"
			switch mode {
			case "ownership-read":
				condition = "*' rev-list '*"
			case "ownership-workspace":
				condition = "*' --git-common-dir '*"
			case "ownership-fetched-tracker":
				condition = "*' for-each-ref --count=1 '*"
			}
			script := "#!/bin/sh\ncase \" $* \" in\n" + condition + ") printf '%s' \"$$\" > \"$SDLC_GIT_PARENT_PID\"; sleep 60 & printf '%s' \"$!\" > \"$SDLC_GIT_CHILD_PID\"; wait; exit 1;;\nesac\nexec '" + realGit + "' \"$@\"\n"
			if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pr := landingPR{HeadOID: strings.Repeat("a", 40), BaseOID: strings.Repeat("b", 40)}
			done := make(chan error, 1)
			go func() {
				if mode == "publish-first-read" {
					done <- runLandingPublishGate(ctx, pr, "workshop/issues", "workshop/plans", io.Discard)
					return
				}
				_, err := selectLandingIssues(ctx, root, pr, "workshop/issues")
				done <- err
			}()
			child := awaitLandingGitChild(t, pidfile)
			raw, _ := os.ReadFile(parentfile)
			parent, _ := strconv.Atoi(string(raw))
			t.Cleanup(func() { _ = syscall.Kill(parent, syscall.SIGKILL) })
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation error = %v", err)
				}
			case <-time.After(500 * time.Millisecond):
				_ = syscall.Kill(parent, syscall.SIGKILL)
				_ = syscall.Kill(child, syscall.SIGKILL)
				<-done
				t.Fatal("operation cancellation did not reach Git read")
			}
			if err := testfix.AwaitExit(child, time.Second); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type cancelArchiveReadRunner struct {
	gitRunner
	cancel context.CancelFunc
}

func (r cancelArchiveReadRunner) GitInDir(string, ...string) ([]byte, error) {
	r.cancel()
	return []byte("success\n"), nil
}

func TestArchiveReadRejectsCancelledSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	old := landingArchiveRunner
	landingArchiveRunner = cancelArchiveReadRunner{gitRunner: execGitRunner{}, cancel: cancel}
	t.Cleanup(func() { landingArchiveRunner = old })
	if _, err := archiveReadContext(ctx, t.TempDir(), "rev-parse", "HEAD"); !errors.Is(err, context.Canceled) {
		t.Fatalf("late success = %v", err)
	}
}
