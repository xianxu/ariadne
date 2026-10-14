package gitx

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

var gitCommandTimeout = 2 * time.Minute

// Even legacy callers without a context have a command ceiling. Contextual
// calls inherit cancellation and the smaller caller budget. The process group
// owns hook, transport and merge-driver descendants, including after the
// direct child exits and leaves inherited pipes open.
func boundedGitCommand(ctx context.Context, env []string, name string, args ...string) (*exec.Cmd, context.Context, context.CancelFunc) {
	// Existing context-free Git APIs pass nil. Normalize only at this single
	// execution boundary; contextual callers retain their operation context.
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, gitCommandTimeout)
	cmd := exec.CommandContext(ctx, name, args...)
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	return cmd, ctx, cancel
}

func finishGitCommand(cmd *exec.Cmd, ctx context.Context, err error) error {
	_ = processgroup.Terminate(cmd, true)
	if ctx.Err() != nil {
		return errors.Join(ctx.Err(), err)
	}
	return err
}

func boundedGitOutput(ctx context.Context, env []string, name string, args ...string) ([]byte, error) {
	cmd, child, cancel := boundedGitCommand(ctx, env, name, args...)
	defer cancel()
	out, err := cmd.Output()
	return out, finishGitCommand(cmd, child, err)
}
