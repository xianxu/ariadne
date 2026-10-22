//go:build unix

package gitx

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestNetworkInternalContextualProcessResults(t *testing.T) {
	for _, boundary := range []string{"output", "public", "trunk", "bounded-input"} {
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
					switch boundary {
					case "output":
						out, err = boundedGitOutput(ctx, nil, "git", "fetch", "origin")
					case "public":
						out, _, err = RunGitInContext(ctx, dir, nil, "fetch", "origin")
					case "trunk":
						tf, _ := NewTrunkFileContext(ctx, dir, "origin", "main")
						out, _, err = tf.runIn(ctx, nil, "fetch", "origin")
					case "bounded-input":
						out, _, err = runGitBoundedInputContext(ctx, dir, nil, nil, 1024, "fetch", "origin")
					}
					done <- result{out, err}
				}()
				if mode == "cancel" || mode == "deadline" {
					pid := testfix.AwaitPIDFile(ready, testfix.HangCeiling, nil)
					if pid == 0 {
						t.Fatal("diagnostic barrier not reached")
					}
					t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
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
					var network *GitNetworkFailure
					if !errors.Is(got.err, wanted) || !errors.As(got.err, &exit) {
						t.Fatalf("lost context/process error: %v", got.err)
					}
					if mode == "cancel" && errors.As(got.err, &network) {
						t.Fatalf("cancellation classified as network: %v", got.err)
					}
					if err := testfix.AwaitExit(pid, time.Second); err != nil {
						t.Fatal(err)
					}
				} else {
					got := <-done
					if mode == "success" {
						if got.err != nil {
							t.Fatal(got.err)
						}
					} else {
						var network *GitNetworkFailure
						var exit *exec.ExitError
						if !errors.As(got.err, &network) || !errors.As(got.err, &exit) {
							t.Fatalf("lost normal network/process error: %v", got.err)
						}
					}
					if string(got.out) != "parsed-output" {
						t.Fatalf("stdout contaminated: %q", got.out)
					}
				}
			})
		}
	}
}
