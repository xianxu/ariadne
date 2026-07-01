package judge

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestHeartbeatLine pins the pure progress-line wording (#140): elapsed is
// rounded + humanized, the agent name is shown (harness-agnostic), and the PID —
// once known — comes with a copy-pasteable `ps -p` inspection hint. pid <= 0
// renders the "pending" window before the child is observed starting.
func TestHeartbeatLine(t *testing.T) {
	cases := []struct {
		name    string
		elapsed time.Duration
		agent   string
		pid     int
		want    []string
		absent  []string
	}{
		{"claude with pid", 30 * time.Second, "claude", 4242,
			[]string{"30s elapsed via claude", "pid 4242", "ps -p 4242", "still working"}, nil},
		{"codex minutes", 90 * time.Second, "codex", 111,
			[]string{"1m30s elapsed via codex", "pid 111", "ps -p 111"}, nil},
		{"gemini pending pid", 5 * time.Second, "gemini", 0,
			[]string{"5s elapsed via gemini", "pid pending"}, []string{"ps -p"}},
		{"empty agent falls back", 5 * time.Second, "", 7,
			[]string{"via agent", "pid 7"}, nil},
		{"sub-second elapsed rounds", 1200 * time.Millisecond, "claude", 9,
			[]string{"1s elapsed"}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := heartbeatLine(c.elapsed, c.agent, c.pid)
			for _, w := range c.want {
				if !strings.Contains(got, w) {
					t.Errorf("heartbeatLine = %q, missing %q", got, w)
				}
			}
			for _, a := range c.absent {
				if strings.Contains(got, a) {
					t.Errorf("heartbeatLine = %q, should not contain %q", got, a)
				}
			}
		})
	}
}

// writerFunc adapts a func to io.Writer so the test can observe each heartbeat
// write as it happens (the handshake signal).
type writerFunc func(p []byte) (int, error)

func (w writerFunc) Write(p []byte) (int, error) { return w(p) }

// TestDispatch_HeartbeatWhileWaiting drives Dispatch's progress path (#140) with
// a fake long-running Run: the fake reports pid 4242 at launch, then blocks until
// released, standing in for a multi-minute review. A hand-driven ticker + a
// synchronous tick→observe handshake makes the assertion deterministic (no real
// sleeping, no select-race flakiness): each tick produces exactly one heartbeat
// line before the next tick is sent. The final captured output and nil error
// must survive unchanged, so downstream Classify/ParseVerdict are untouched.
func TestDispatch_HeartbeatWhileWaiting(t *testing.T) {
	const canned = "VERDICT: SHIP (confidence: high)\n\nlooks good.\n"

	origRun := Run
	t.Cleanup(func() { Run = origRun })
	release := make(chan struct{})
	started := make(chan struct{}) // closed once onStart has stored the pid
	var sawOnStart bool
	Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		if onStart != nil {
			onStart(4242)
			sawOnStart = true
		}
		close(started)
		<-release // block like a real review in flight
		return []byte(canned), nil
	}

	origTicker := newHeartbeatTicker
	t.Cleanup(func() { newHeartbeatTicker = origTicker })
	tickCh := make(chan time.Time)
	var stopped bool
	newHeartbeatTicker = func(d time.Duration) (<-chan time.Time, func()) {
		return tickCh, func() { stopped = true }
	}

	// Deterministic elapsed: 30s on the first beat, 60s on the second, …
	origSince := sinceStart
	t.Cleanup(func() { sinceStart = origSince })
	var sinceCalls int
	sinceStart = func(time.Time) time.Duration {
		sinceCalls++
		return time.Duration(sinceCalls) * 30 * time.Second
	}

	wrote := make(chan string, 8)
	stderr := writerFunc(func(p []byte) (int, error) { wrote <- string(p); return len(p), nil })

	type res struct {
		out string
		err error
	}
	resCh := make(chan res, 1)
	go func() {
		out, err := Dispatch(context.Background(), DispatchOptions{
			Agent: AgentClaude, Prompt: "review please", Stderr: stderr,
		})
		resCh <- res{out, err}
	}()

	<-started // ensure onStart has stored the pid before the first tick reads it

	const beats = 3
	var lines []string
	for i := 0; i < beats; i++ {
		tickCh <- time.Time{} // blocks until the select loop takes the tick
		lines = append(lines, <-wrote)
	}

	close(release) // let the fake review finish
	r := <-resCh

	if r.err != nil {
		t.Fatalf("Dispatch returned error: %v", r.err)
	}
	if r.out != canned {
		t.Errorf("output altered by heartbeat path:\n got %q\nwant %q", r.out, canned)
	}
	if !sawOnStart {
		t.Error("Run's onStart was never invoked — PID never surfaced")
	}
	if len(lines) != beats {
		t.Fatalf("got %d heartbeat lines, want %d:\n%v", len(lines), beats, lines)
	}
	wantElapsed := []string{"30s elapsed", "1m0s elapsed", "1m30s elapsed"}
	for i, ln := range lines {
		if !strings.Contains(ln, "pid 4242") {
			t.Errorf("beat %d missing child pid: %q", i, ln)
		}
		if !strings.Contains(ln, "via claude") {
			t.Errorf("beat %d missing agent name: %q", i, ln)
		}
		if !strings.Contains(ln, wantElapsed[i]) {
			t.Errorf("beat %d elapsed = %q, want %q", i, ln, wantElapsed[i])
		}
	}
	if !stopped {
		t.Error("heartbeat ticker was not stopped when Dispatch returned")
	}
}

// TestRun_RealSubprocess exercises the reimplemented production Run (#140)
// against a real OS process — not a stub — so the Start→onStart→Wait→combined-
// buffer path is proven end-to-end (the lessons file: pure tests can't see IO
// bugs). It asserts a real PID reaches onStart, stdout+stderr are combined like
// the old CombinedOutput, and a non-zero exit surfaces as *exec.ExitError with
// output intact (the shape Dispatch's classifyRunResult depends on).
func TestRun_RealSubprocess(t *testing.T) {
	t.Run("captures pid and combines streams", func(t *testing.T) {
		var gotPID int
		out, err := Run(context.Background(), func(pid int) { gotPID = pid },
			"sh", "-c", "echo to-stdout; echo to-stderr 1>&2")
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		if gotPID <= 0 {
			t.Errorf("onStart pid = %d, want a real pid > 0", gotPID)
		}
		if s := string(out); !strings.Contains(s, "to-stdout") || !strings.Contains(s, "to-stderr") {
			t.Errorf("combined output missing a stream: %q", s)
		}
	})

	t.Run("non-zero exit surfaces ExitError with output", func(t *testing.T) {
		out, err := Run(context.Background(), nil, "sh", "-c", "echo boom 1>&2; exit 3")
		if _, ok := err.(*exec.ExitError); !ok {
			t.Fatalf("err = %v (%T), want *exec.ExitError", err, err)
		}
		if !strings.Contains(string(out), "boom") {
			t.Errorf("output not captured on non-zero exit: %q", out)
		}
	})
}

// TestDispatch_NoStderrNoHeartbeat guards the fast path: with opts.Stderr nil,
// Dispatch runs synchronously (no ticker goroutine) and returns the output +
// exit-code semantics exactly as before — the existing callers/tests rely on it.
func TestDispatch_NoStderrNoHeartbeat(t *testing.T) {
	origRun := Run
	t.Cleanup(func() { Run = origRun })
	Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		return []byte("VERDICT: SHIP (confidence: high)\n"), nil
	}
	// If Dispatch touched a nil Stderr it would panic; a clean return proves the
	// gate holds.
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "x"})
	if err != nil {
		t.Fatalf("Dispatch: %v", err)
	}
	if !strings.Contains(out, "VERDICT: SHIP") {
		t.Errorf("output = %q", out)
	}
}
