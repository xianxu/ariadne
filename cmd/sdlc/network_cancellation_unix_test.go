//go:build unix

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestNetworkContextualRunnerFinalizesBeforeClassification(t *testing.T) {
	for _, boundary := range []string{"landing", "weekly"} {
		for _, mode := range []string{"cancel", "deadline", "network", "success"} {
			t.Run(boundary+"/"+mode, func(t *testing.T) {
				dir := t.TempDir()
				script := `#!/bin/sh
printf 'parsed-output'
printf '%s\n' 'ssh: Could not resolve hostname git.example: Temporary failure in name resolution' >&2
case "$SDLC_NETWORK_MODE" in
 success) exit 0 ;;
 network) exit 128 ;;
esac
printf '%s' "$$" > "$SDLC_NETWORK_READY"
exec sleep 60
`
				if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0755); err != nil {
					t.Fatal(err)
				}
				ready := filepath.Join(dir, "ready.pid")
				t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
				t.Setenv("SDLC_NETWORK_READY", ready)
				t.Setenv("SDLC_NETWORK_MODE", mode)
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				var expire func()
				if mode == "deadline" {
					ctx, expire = testfix.DeadlineOnCue(ctx)
				}
				type result struct {
					out []byte
					err error
				}
				done := make(chan result, 1)
				go func() {
					var out []byte
					var err error
					if boundary == "landing" {
						out, err = boundedLandingGit(ctx, execGitRunner{}).GitInDir(dir, "fetch", "origin")
					} else {
						out, err = weeklyGit(ctx, dir, "fetch", "origin")
					}
					done <- result{out, err}
				}()
				if mode == "cancel" || mode == "deadline" {
					pid := awaitLandingGitChild(t, ready)
					if mode == "cancel" {
						cancel()
					} else {
						expire()
					}
					got := <-done
					wanted := error(context.Canceled)
					if mode == "deadline" {
						wanted = context.DeadlineExceeded
					}
					var exit *exec.ExitError
					if !errors.Is(got.err, wanted) || !errors.As(got.err, &exit) {
						t.Fatalf("lost context/process identity: %v", got.err)
					}
					var network *gitx.GitNetworkFailure
					if mode == "cancel" && errors.As(got.err, &network) {
						t.Fatalf("cancelled command classified as infrastructure: %v", got.err)
					}
					// A deadline with explicit transport evidence may remain classified.
					if err := testfix.AwaitExit(pid, time.Second); err != nil {
						t.Fatal(err)
					}
				} else {
					got := <-done
					if mode == "success" {
						if got.err != nil || string(got.out) != "parsed-output" {
							t.Fatalf("parsed success=%q,%v", got.out, got.err)
						}
					} else {
						var network *gitx.GitNetworkFailure
						var exit *exec.ExitError
						if !errors.As(got.err, &network) || !errors.As(got.err, &exit) {
							t.Fatalf("ordinary transport lost cause: %v", got.err)
						}
					}
				}
			})
		}
	}
}

// The fallback is a raw-result injected seam; production runners are concrete
// execGitRunner values (or unwrapped bounded runners), handled above it.
type networkRawInjectedRunner struct {
	gitRunner
	cancel context.CancelFunc
	cause  error
}

func (r networkRawInjectedRunner) Git(_ ...string) ([]byte, error) {
	r.cancel()
	return []byte("ssh: Could not resolve hostname git.example: Temporary failure in name resolution"), r.cause
}
func (r networkRawInjectedRunner) GitInDir(_ string, args ...string) ([]byte, error) {
	return r.Git(args...)
}

func TestNetworkInjectedLandingRunnerRetainsRawCancellation(t *testing.T) {
	processErr := exec.Command("sh", "-c", "exit 128").Run()
	if processErr == nil {
		t.Fatal("fixture should return raw exit error")
	}
	for _, inDir := range []bool{false, true} {
		for _, cause := range []error{nil, processErr} {
			ctx, cancel := context.WithCancel(t.Context())
			runner := boundedLandingGit(ctx, networkRawInjectedRunner{cancel: cancel, cause: cause})
			var err error
			if inDir {
				_, err = runner.GitInDir(t.TempDir(), "fetch", "origin")
			} else {
				_, err = runner.Git("fetch", "origin")
			}
			cancel()
			var network *gitx.GitNetworkFailure
			if !errors.Is(err, context.Canceled) || errors.As(err, &network) {
				t.Fatalf("inDir=%v raw cancellation=%v", inDir, err)
			}
			if cause != nil {
				var exit *exec.ExitError
				if !errors.Is(err, cause) || !errors.As(err, &exit) {
					t.Fatalf("lost raw process identity: %v", err)
				}
			}
		}
	}
}
