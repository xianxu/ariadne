package judge

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"testing"
	"time"
)

func TestDispatchCancellationCannotBecomeVerdict(t *testing.T) {
	for _, progress := range []bool{false, true} {
		t.Run(map[bool]string{false: "synchronous", true: "heartbeat"}[progress], func(t *testing.T) {
			for _, runErr := range []error{nil, &exec.ExitError{}} {
				t.Run(map[bool]string{false: "clean-exit", true: "exit-error"}[runErr != nil], func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					saved := Run
					t.Cleanup(func() { Run = saved })
					Run = func(context.Context, func(int), string, ...string) (ProcessOutput, error) {
						cancel()
						return ProcessOutput{Stdout: []byte("SHIP"), Stderr: []byte("diagnostic")}, runErr
					}
					var sink io.Writer
					if progress {
						sink = &bytes.Buffer{}
					}
					out, err := Dispatch(ctx, DispatchOptions{Agent: AgentClaude, Stderr: sink})
					if !errors.Is(err, context.Canceled) {
						t.Errorf("cancelled review returned %v; want context cancellation", err)
					}
					if out != "" {
						t.Errorf("cancelled review exposed verdict %q", out)
					}
				})
			}
		})
	}
}

func TestDispatchReviewTimeoutConfiguration(t *testing.T) {
	for _, tc := range []struct {
		value   string
		want    time.Duration
		invalid bool
	}{
		{"", 30 * time.Minute, false}, {"1s", time.Second, false}, {"2h", 2 * time.Hour, false},
		{"0", 0, true}, {"999ms", 0, true}, {"2h1ns", 0, true}, {"forever", 0, true}, {"-1s", 0, true},
	} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("WF_REVIEW_TIMEOUT", tc.value)
			saved := Run
			t.Cleanup(func() { Run = saved })
			called := false
			Run = func(ctx context.Context, _ func(int), _ string, _ ...string) (ProcessOutput, error) {
				called = true
				deadline, ok := ctx.Deadline()
				if !ok {
					t.Error("review has no deadline")
				} else if left := time.Until(deadline); left > tc.want || left < tc.want-time.Second {
					t.Errorf("deadline remaining %v, want about %v", left, tc.want)
				}
				return ProcessOutput{Stdout: []byte("SHIP")}, nil
			}
			_, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude})
			if tc.invalid {
				if err == nil || called {
					t.Errorf("invalid timeout: err=%v called=%v", err, called)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDispatchParentDeadlineCannotBecomeVerdict(t *testing.T) {
	for _, progress := range []bool{false, true} {
		t.Run(map[bool]string{false: "synchronous", true: "heartbeat"}[progress], func(t *testing.T) {
			t.Setenv("WF_REVIEW_TIMEOUT", "2h")
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			saved := Run
			t.Cleanup(func() { Run = saved })
			returned := false
			Run = func(ctx context.Context, _ func(int), _ string, _ ...string) (ProcessOutput, error) {
				<-ctx.Done()
				returned = true
				return ProcessOutput{Stdout: []byte("SHIP")}, &exec.ExitError{}
			}
			var sink io.Writer
			if progress {
				sink = &bytes.Buffer{}
			}
			out, err := Dispatch(ctx, DispatchOptions{Agent: AgentClaude, Stderr: sink})
			if !returned || !errors.Is(err, context.DeadlineExceeded) || out != "" {
				t.Fatalf("deadline result: returned=%v out=%q err=%v", returned, out, err)
			}
		})
	}
}

func TestDispatchAlreadyCanceledDoesNotStartReviewer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	saved := Run
	t.Cleanup(func() { Run = saved })
	Run = func(context.Context, func(int), string, ...string) (ProcessOutput, error) {
		t.Fatal("started cancelled reviewer")
		return ProcessOutput{}, nil
	}
	if out, err := Dispatch(ctx, DispatchOptions{}); out != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("out=%q err=%v", out, err)
	}
}
