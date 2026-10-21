package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/pkg/testgate"
)

func TestLandingCompletionBudget(t *testing.T) {
	now := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name        string
		remaining   time.Duration
		hasDeadline bool
		want        time.Duration
	}{
		{"short supervisor", 20 * time.Second, true, 10 * time.Second},
		{"loaded supervisor", 10 * time.Minute, true, 5 * time.Minute},
		{"no supervisor", 0, false, testgate.HangCeiling},
		{"expired", -time.Second, true, 0},
		{"at deadline", 0, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := landingCompletionBudget(now.Add(tc.remaining), tc.hasDeadline, now); got != tc.want {
				t.Fatalf("budget = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestLandingCompletionPreservesErrors(t *testing.T) {
	for _, want := range []error{nil, errors.New("merge failed"), context.DeadlineExceeded} {
		done := make(chan error, 1)
		done <- want
		if got := finished(t, done); got != want {
			t.Fatalf("error = %v, want %v", got, want)
		}
	}
}

func TestLandingCompletionDiagnostics(t *testing.T) {
	const childEnv = "ARIADNE_LANDING_WAIT_DIAGNOSTIC"
	if helper := os.Getenv(childEnv); helper != "" {
		done := make(chan error)
		stop, ready := make(chan struct{}), make(chan struct{})
		go landingCompletionDiagnosticBlocked(ready, stop)
		<-ready
		defer close(stop)
		if helper == "finished" {
			finished(t, done)
		} else {
			wantSettled(t, done, landingline.OutcomeLanded)
		}
		t.Fatal("completion helper returned without an event")
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, helper := range []string{"finished", "wantSettled"} {
		t.Run(helper, func(t *testing.T) {
			// Real testing.T deadline supervises the actual helper. Its named failure
			// must precede Go's outer alarm; no artificial clock or short elapsed bound.
			cmd := exec.Command(executable, "-test.run=^TestLandingCompletionDiagnostics$", "-test.timeout=8s")
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), childEnv+"="+helper)
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("hung helper unexpectedly passed: %s", out)
			}
			output := string(out)
			for _, want := range []string{"completion wait exhausted", "candidate", "runDurableMerge", "repository-lock cleanup", "archive confirmation", "checkout revalidation", "goroutine ", "[running]:", "TestLandingCompletionDiagnostics", "landingCompletionDiagnosticBlocked"} {
				if !strings.Contains(output, want) {
					t.Errorf("missing %q in diagnostic:\n%s", want, output)
				}
			}
			if strings.Contains(output, "panic: test timed out") {
				t.Fatalf("outer supervisor fired before helper:\n%s", output)
			}
		})
	}
}

// landingCompletionBudget reserves half of the actual remaining Go supervisor
// deadline for diagnostics and cleanup, including any time spent at testgate.
func landingCompletionBudget(deadline time.Time, hasDeadline bool, now time.Time) time.Duration {
	if !hasDeadline {
		return testgate.HangCeiling
	}
	return max(time.Duration(0), deadline.Sub(now)/2)
}

// awaitLandingCompletion only observes the completion channel: diagnostics must
// not acquire fixture, fake GitHub, or repository locks that the merge may hold.
func awaitLandingCompletion(t *testing.T, done <-chan error, failure string) error {
	t.Helper()
	deadline, hasDeadline := t.Deadline()
	budget := landingCompletionBudget(deadline, hasDeadline, time.Now())
	fail := func() {
		t.Helper()
		var stacks bytes.Buffer
		_ = pprof.Lookup("goroutine").WriteTo(&stacks, 2)
		t.Fatalf("%s: completion wait exhausted after %s; candidate: runDurableMerge final repository-lock cleanup (archive confirmation and checkout revalidation)\nall goroutine stacks:\n%s", failure, budget, stacks.String())
	}
	if budget <= 0 {
		fail()
		return nil
	}
	timer := time.NewTimer(budget)
	defer timer.Stop()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		fail()
		return nil
	}
}

// This independent goroutine proves the diagnostic includes stacks beyond the
// failing test's own stack. Readiness avoids depending on its scheduling.
func landingCompletionDiagnosticBlocked(ready chan<- struct{}, stop <-chan struct{}) {
	close(ready)
	<-stop
}
