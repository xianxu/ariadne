package gitx

import (
	"bytes"
	"context"
	"errors"
	"io"
)

var ErrOutputLimit = errors.New("Git snapshot output exceeds safety budget")

// RunGitBoundedContext shares the contextual process lifetime and output bound
// with callers that inspect metadata before materializing a disposable tree.
func RunGitBoundedContext(ctx context.Context, dir string, env []string, limit int, args ...string) ([]byte, []byte, error) {
	return runGitBoundedInputContext(ctx, dir, env, nil, limit, args...)
}

// These are input-safety ceilings, not performance targets. Re-measure with the
// 10,000-card tracker benchmark before raising them. Diagnostics need only the
// first failure, and are never parsed as file content.
const (
	SnapshotOutputLimit   = 32 << 20
	SnapshotBlobLimit     = 1 << 20
	SnapshotEntryLimit    = 10001
	snapshotOutputLimit   = SnapshotOutputLimit
	diagnosticOutputLimit = 64 << 10
)

type boundedOutput struct {
	buffer   bytes.Buffer
	limit    int
	cancel   context.CancelFunc
	exceeded bool
}

func (w *boundedOutput) Len() int      { return w.buffer.Len() }
func (w *boundedOutput) Bytes() []byte { return w.buffer.Bytes() }

func (w *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > w.limit-w.Len() {
		w.exceeded = true
		w.cancel()
		return 0, ErrOutputLimit
	}
	return w.buffer.Write(p)
}

func runGitBoundedInputContext(ctx context.Context, dir string, env []string, input io.Reader, limit int, args ...string) ([]byte, []byte, error) {
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	out := &boundedOutput{limit: limit, cancel: cancel}
	diag := &boundedOutput{limit: diagnosticOutputLimit, cancel: cancel}
	err := executeGitInputContext(child, dir, env, input, out, diag, args...)
	if ctx.Err() != nil {
		err = errors.Join(ctx.Err(), err)
	} else if out.exceeded || diag.exceeded {
		err = ErrOutputLimit
	}
	return out.Bytes(), diag.Bytes(), ClassifyNetworkFailure(args, diag.Bytes(), err)
}
