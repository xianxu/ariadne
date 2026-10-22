package judge

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestDependencyRetryReceivesDispatchCancellation(t *testing.T) {
	source := testfix.Repo(t, testfix.InitialCommit())
	head := strings.TrimSpace(testfix.Capture(t, source, "rev-parse", "HEAD"))
	owner, err := reviewbox.Create(t.TempDir(), source, head, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	tools := t.TempDir()
	ready := filepath.Join(tools, "child.pid")
	if err := os.WriteFile(filepath.Join(tools, "git"), []byte("#!/bin/sh\nsleep 60 &\nprintf '%s' \"$!\" > \"$REVIEW_RETRY_READY\"\nwait\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tools+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("REVIEW_RETRY_READY", ready)
	original := Run
	t.Cleanup(func() { Run = original })
	launched := 0
	Run = func(context.Context, func(int), string, ...string) (ProcessOutput, error) {
		launched++
		return ProcessOutput{Stdout: []byte("response missing verdict")}, nil
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := Dispatch(ctx, DispatchOptions{Agent: AgentCodex, Checkout: owner.Dir,
			BeforeRetryContext: owner.ResetContext,
			BeforeRetry:        func() error { return errors.New("context-free retry must not run") },
		})
		done <- err
	}()
	pid := testfix.AwaitPIDFile(ready, testfix.HangCeiling, nil)
	if pid == 0 {
		cancel()
		t.Fatal("retry Git descendant never became ready")
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("retry lost cancellation: %v", err)
	}
	if launched != 1 {
		t.Fatalf("cancelled retry launched another reviewer: %d", launched)
	}
	if err := testfix.AwaitExit(pid, testfix.HangCeiling); err != nil {
		t.Fatal(err)
	}
	if err := owner.Close(); err != nil {
		t.Fatal(err)
	}
	if entries, err := os.ReadDir(filepath.Dir(owner.Dir)); err != nil || len(entries) != 0 {
		t.Fatalf("retry ownership debris: %v, %v", entries, err)
	}
}
