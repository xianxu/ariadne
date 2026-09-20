package weavefs

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
)

func TestOwnedRunnerBoundsBufferedDescendants(t *testing.T) {
	for _, cancelRun := range []bool{false, true} {
		t.Run(map[bool]string{false: "normal-exit", true: "cancel"}[cancelRun], func(t *testing.T) {
			fixture := t.TempDir()
			stage, err := staging.New(filepath.Join(fixture, "generation"))
			if err != nil {
				t.Fatal(err)
			}
			if err := syscall.Mkfifo(filepath.Join(fixture, "gate"), 0600); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				lease, err := staging.Exclusive(stage)
				if err == nil {
					lease.Close()
				} else if errors.Is(err, staging.ErrInUse) {
					b, _ := os.ReadFile(filepath.Join(fixture, "pgid"))
					pgid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
					if pgid > 0 {
						_ = syscall.Kill(-pgid, syscall.SIGKILL)
					}
				}
				_ = staging.Remove(stage)
			})
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var output bytes.Buffer
			runner := ExecRunner{Context: ctx, Stdout: &output, Stderr: &output}
			script := `echo $$ > "$1/pgid"
(sh -c 'echo ready > "$1/ready"; read token < "$1/gate"; echo late > "$1/late"' sh "$1") &
if [ "$2" = cancel ]; then wait; else
  while [ ! -f "$1/ready" ]; do sleep 0.01; done
fi`
			mode := "normal"
			if cancelRun {
				mode = "cancel"
			}
			done := make(chan error, 1)
			go func() { done <- runner.RunOwned(fixture, []string{"sh", "-c", script, "sh", fixture, mode}, stage) }()
			for {
				if _, err := os.Stat(filepath.Join(fixture, "ready")); err == nil {
					break
				}
				select {
				case err := <-done:
					t.Fatal("runner ended before child readiness", err)
				case <-ctx.Done():
					t.Fatal("child readiness timeout")
				case <-time.After(5 * time.Millisecond):
				}
			}
			if cancelRun {
				cancel()
			}
			select {
			case err := <-done:
				if cancelRun && err == nil {
					t.Fatal("cancellation succeeded")
				}
				if !cancelRun && err != nil && !errors.Is(err, exec.ErrWaitDelay) {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("buffered output kept Wait blocked")
			}
			// Positive kernel proof: a live inherited writer would prevent this lock.
			lease, err := staging.Exclusive(stage)
			if err != nil {
				t.Fatalf("returned with descendant lease alive: %v", err)
			}
			lease.Close()
			if err := staging.Remove(stage); err != nil {
				t.Fatal(err)
			}
		})
	}
}
