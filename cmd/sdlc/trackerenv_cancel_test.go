package main

import (
	"context"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// #302: a cancelled tracker push takes the receive side down with it. Killing
// only the direct git left receive-pack (and its hook) writing into the origin
// after the call returned, and under sharded load that late write raced
// t.TempDir's cleanup: "unlinkat …/origin.git: directory not empty".
func TestTrackerGitCancelEndsTheReceiveSide(t *testing.T) {
	testfix.AssertCancelEndsReceiveSide(t, func(ctx context.Context, root, origin string) error {
		_, err := (&trackerEnv{ctx: ctx, root: root}).gitRaw(nil, "push", "-q", origin, "HEAD:refs/heads/main")
		return err
	})
}
