package judge

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewprepare"
)

// cappedOutput drains both streams while retaining a fixed diagnostic budget.
type cappedOutput struct {
	mu        sync.Mutex
	data      []byte
	truncated bool
}

func (b *cappedOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	left := reviewprepare.OutputLimit - len(b.data)
	if len(p) > left {
		p = p[:left]
		b.truncated = true
	}
	b.data = append(b.data, p...)
	return n, nil
}

func runPreparation(ctx context.Context, opts DispatchOptions, attempt int) (err error) {
	if opts.Preparation == nil {
		return nil
	}
	cfg := opts.Preparation
	start := time.Now()
	result := reviewprepare.Result{Head: opts.PreparationHead, Command: append([]string(nil), cfg.Command...), Attempt: attempt}
	capture := &cappedOutput{}
	defer func() {
		result.Duration = time.Since(start)
		result.Output = string(capture.data)
		result.Truncated = capture.truncated
		if err != nil {
			result.Error = err.Error()
		}
		if opts.Evidence != nil {
			opts.Evidence.Preparation = append(opts.Evidence.Preparation, result)
		}
	}()
	if opts.Checkout == "" || len(cfg.Command) == 0 {
		return fmt.Errorf("review preparation requires a disposable checkout and command")
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TimeoutMinutes)*time.Minute)
	defer cancel()
	before, err := trackedPreparationState(ctx, opts.Checkout)
	if err != nil {
		return err
	}
	name, args := cfg.Command[0], cfg.Command[1:]
	if opts.Wrap != nil {
		name, args = opts.Wrap(name, args)
	}
	if opts.Stderr != nil {
		fmt.Fprintf(opts.Stderr, "    … preparing review checkout (attempt %d): %q\n", attempt, cfg.Command)
	}
	if err = runCommand(ctx, nil, opts.Checkout, capture, capture, name, args...); err != nil {
		return fmt.Errorf("review preparation attempt %d: %w", attempt, err)
	}
	after, err := trackedPreparationState(ctx, opts.Checkout)
	if err != nil {
		return err
	}
	if !bytes.Equal(before, after) {
		return fmt.Errorf("review preparation changed tracked checkout state; generate ignored or untracked assets only")
	}
	return nil
}

// Include HEAD, the index, and working-tree contents. Existing issue overlays
// may already be dirty; preparation must preserve that starting state exactly.
func trackedPreparationState(ctx context.Context, dir string) ([]byte, error) {
	var state bytes.Buffer
	for _, args := range [][]string{{"rev-parse", "HEAD"}, {"ls-files", "--stage", "-z"}, {"diff", "HEAD", "--binary", "--no-ext-diff", "--no-textconv"}} {
		out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
		if err != nil {
			return nil, fmt.Errorf("inspect prepared tracked tree: %w", err)
		}
		state.Write(out)
		state.WriteByte(0)
	}
	return state.Bytes(), nil
}
