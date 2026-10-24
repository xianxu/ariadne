package main

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
)

// #599: a check's deadline bounds its commands alone. Setup before them is not
// charged to it, and bookkeeping after them cannot turn a finished run into a
// timeout. The phase seam stands in for a loaded machine, so no test here
// depends on how fast this one is.
func TestLocalCheckDeadlineBoundsTheCommandsOnly(t *testing.T) {
	for _, tc := range []struct {
		name    string
		slow    string
		command []string
		timeout bool
	}{
		{name: "slow setup", slow: "setup", command: []string{"true"}},
		{name: "slow bookkeeping", slow: "ran", command: []string{"true"}},
		{name: "real overrun", command: []string{"sleep", "30"}, timeout: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, sha, _, out := localRunRepo(t, "acquired")
			d.minute = 300 * time.Millisecond
			d.phaseHook = func(phase string) {
				if phase == tc.slow {
					time.Sleep(3 * d.minute)
				}
			}
			r, err := runLocalCheck(context.Background(), d, localCheck{Name: "suite", Command: tc.command, TimeoutMinutes: 1}, sha)
			if err != nil || r.TimedOut != tc.timeout || r.passed() == tc.timeout {
				t.Fatalf("timed out %v, want %v: %+v %v\n%s", r.TimedOut, tc.timeout, r, err, out)
			}
			if tc.timeout && r.Duration < d.minute {
				t.Fatalf("stopped after %v, before its %v deadline", r.Duration, d.minute)
			}
		})
	}
}

func TestDeadlineStopped(t *testing.T) {
	failed := &exec.ExitError{}
	for _, tc := range []struct {
		runErr, ctxErr error
		want           bool
	}{
		{nil, context.DeadlineExceeded, false}, // finished, then the deadline passed
		{failed, context.DeadlineExceeded, true},
		{context.DeadlineExceeded, context.DeadlineExceeded, true}, // never started
		{failed, nil, false},
		{failed, context.Canceled, false},
		{nil, nil, false},
	} {
		if got := deadlineStopped(tc.runErr, tc.ctxErr); got != tc.want {
			t.Errorf("deadlineStopped(%v, %v) = %v, want %v", tc.runErr, tc.ctxErr, got, tc.want)
		}
	}
	if deadlineStopped(errors.New("x"), errors.Join(context.DeadlineExceeded)) != true {
		t.Error("a wrapped deadline is still the deadline")
	}
}

// #599: a graph preparation that succeeded stands, even when its deadline
// passes during the teardown after it.
func TestAffectedGraphPreparationSuccessOutlivesItsDeadline(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "acquired")
	d.minute = 200 * time.Millisecond
	d.phaseHook = func(phase string) {
		if phase == "ran" {
			time.Sleep(3 * d.minute)
		}
	}
	p := &affected.Preparation{Command: []string{"true"}, TimeoutMinutes: 1}
	if err := prepareAffectedGraph(context.Background(), d, t.TempDir(), p); err != nil {
		t.Fatalf("a successful preparation became %v", err)
	}
	p.Command = []string{"sleep", "30"}
	if err := prepareAffectedGraph(context.Background(), d, t.TempDir(), p); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("a real overrun is not the deadline: %v", err)
	}
}
