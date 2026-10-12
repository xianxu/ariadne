package main

import (
	"context"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// #302: a cancelled tracker push takes the receive side down with it. Killing
// only the direct git left receive-pack (and its hook) writing into the origin
// after the call returned, and under sharded load that late write raced
// t.TempDir's cleanup: "unlinkat …/origin.git: directory not empty".
func TestTrackerGitCancelEndsTheReceiveSide(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	origin, hookPID := testfix.BlockingOrigin(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	env := &trackerEnv{ctx: ctx, root: root}
	done := make(chan error, 1)
	go func() { _, err := env.gitRaw(nil, "push", "-q", origin, "HEAD:refs/heads/main"); done <- err }()
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
