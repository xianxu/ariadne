//go:build unix

package gitx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestGitDefaultRunnerAndContextRunnerBoundDescendants(t *testing.T) {
	for _, mode := range []string{"default-deadline", "default-exited-parent", "context-exited-parent"} {
		t.Run(mode, func(t *testing.T) {
			pidfile := filepath.Join(t.TempDir(), "child.pid")
			t.Setenv("SDLC_GIT_CHILD_PID", pidfile)
			alias := `!sleep 60 & printf '%s' "$!" > "$SDLC_GIT_CHILD_PID"; exit 0`
			old := gitCommandTimeout
			gitCommandTimeout = 5 * time.Second
			t.Cleanup(func() { gitCommandTimeout = old })
			// The deadline passes on cue once the descendant has reported itself
			// (#434, #407): the bounded runner under the default ceiling derives
			// its timeout from this context, so a clock cannot race the fixture.
			deadline, expire := testfix.DeadlineOnCue(context.Background())
			if mode == "default-deadline" {
				alias = `!sleep 60 & printf '%s' "$!" > "$SDLC_GIT_CHILD_PID"; wait`
			}
			done := make(chan error, 1)
			go func() {
				args := []string{"-c", "alias.hang=" + alias, "hang"}
				switch mode {
				case "context-exited-parent":
					var out, diag bytes.Buffer
					done <- executeGitInputContext(context.Background(), "", nil, nil, &out, &diag, args...)
				case "default-deadline":
					_, err := boundedGitOutput(deadline, nil, "git", args...)
					done <- err
				default:
					_, err := run("git", args...)
					done <- err
				}
			}()
			pid := testfix.AwaitPIDFile(pidfile, testfix.HangCeiling, nil)
			if pid == 0 {
				t.Fatal("git descendant never started")
			}
			expire()
			t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
			select {
			case err := <-done:
				if mode == "default-deadline" && !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("deadline error = %v", err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("git exceeded bound")
			}
			if err := testfix.AwaitExit(pid, time.Second); err != nil {
				t.Fatalf("git left descendant: %v", err)
			}
		})
	}
}

func TestGitHistoryStreamBoundsInheritedPipes(t *testing.T) {
	root := t.TempDir()
	pidfile := filepath.Join(root, "child.pid")
	t.Setenv("SDLC_GIT_CHILD_PID", pidfile)
	t.Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
	script := "#!/bin/sh\nsleep 60 & printf '%s' \"$!\" > \"$SDLC_GIT_CHILD_PID\"; exit 0\n"
	if err := os.WriteFile(filepath.Join(root, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	// The stream's deadline passes on cue after the descendant reports in (#434).
	deadline, expire := testfix.DeadlineOnCue(context.Background())
	trunk := &TrunkFile{dir: root, remote: "origin", branch: "main", ctx: deadline}
	reader, finish, err := trunk.HistoryStream("%H", "workshop/issues", 10)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, _ = io.Copy(io.Discard, reader); done <- finish(false) }()
	pid := testfix.AwaitPIDFile(pidfile, testfix.HangCeiling, nil)
	if pid == 0 {
		t.Fatal("history descendant never started")
	}
	expire()
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("history deadline = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("history reader exceeded bound")
	}
	if err := testfix.AwaitExit(pid, time.Second); err != nil {
		t.Fatal(err)
	}
}
