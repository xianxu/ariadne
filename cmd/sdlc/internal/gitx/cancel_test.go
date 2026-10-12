package gitx

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// #302: a cancelled trunk push takes the receive side down with it; see
// cmd/sdlc's TestTrackerGitCancelEndsTheReceiveSide.
func TestGitCancelEndsTheReceiveSide(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	origin, hookPID := testfix.BlockingOrigin(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		var out, errs bytes.Buffer
		done <- executeGitInputContext(ctx, root, nil, nil, &out, &errs, "push", "-q", origin, "HEAD:refs/heads/main")
	}()
	pid := hookPID()
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("a cancelled push reported success")
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatalf("the cancelled push did not return within %v", testfix.HangCeiling)
	}
	if err := testfix.AwaitExit(pid, testfix.HangCeiling); err != nil {
		t.Fatalf("the receive side outlived the cancelled push: %v", err)
	}
}
