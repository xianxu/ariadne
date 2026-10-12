package gitx

import (
	"bytes"
	"context"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// #302: a cancelled trunk push takes the receive side down with it; see
// cmd/sdlc's TestTrackerGitCancelEndsTheReceiveSide.
func TestGitCancelEndsTheReceiveSide(t *testing.T) {
	testfix.AssertCancelEndsReceiveSide(t, func(ctx context.Context, root, origin string) error {
		var out, errs bytes.Buffer
		return executeGitInputContext(ctx, root, nil, nil, &out, &errs, "push", "-q", origin, "HEAD:refs/heads/main")
	})
}
