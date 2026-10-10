package judge

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// fakeRuns makes Dispatch's process seam replay outputs in order, counting
// the dispatches it saw.
func fakeRuns(t *testing.T, outs ...ProcessOutput) *int {
	t.Helper()
	orig := Run
	t.Cleanup(func() { Run = orig })
	n := 0
	Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (ProcessOutput, error) {
		out := outs[min(n, len(outs)-1)]
		n++
		return out, nil
	}
	return &n
}

func streamOut(t *testing.T, file string) ProcessOutput {
	return ProcessOutput{Stdout: fixture(t, file)}
}

// #300 Done-when: a verdict followed by a block-less postscript parses to the
// earlier verdict and findings, with no retry.
func TestDispatchKeepsAVerdictBeforeAPostscript(t *testing.T) {
	calls := fakeRuns(t, streamOut(t, "postscript.jsonl"))
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "review"})
	if err != nil || *calls != 1 {
		t.Fatalf("err %v, %d dispatches", err, *calls)
	}
	if v := ParseVerdict(out); v != VerdictFixThenShip || !strings.Contains(out, "```findings") {
		t.Fatalf("verdict %q from:\n%s", v, out)
	}
}

// #271: a run that ends without a verdict is dispatched once more, and the
// second run's verdict is the answer.
func TestDispatchRetriesARunWithoutAVerdict(t *testing.T) {
	var errs bytes.Buffer
	calls := fakeRuns(t, streamOut(t, "background_wait.jsonl"), streamOut(t, "clean.jsonl"))
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "review", Stderr: &errs})
	if err != nil || *calls != 2 || ParseVerdict(out) != VerdictFixThenShip {
		t.Fatalf("err %v, %d dispatches, out:\n%s", err, *calls, out)
	}
	if !strings.Contains(errs.String(), "once more") {
		t.Fatalf("the retry was not reported: %q", errs.String())
	}
}

// Two runs without a verdict keep the fail-safe (unknown) and return both
// runs' text, labelled, for the sidecar.
func TestDispatchTwoRunsWithoutAVerdictStayUnknown(t *testing.T) {
	calls := fakeRuns(t, streamOut(t, "background_wait.jsonl"))
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "review"})
	if err != nil || *calls != 2 || ParseVerdict(out) != VerdictUnknown {
		t.Fatalf("err %v, %d dispatches, verdict %q", err, *calls, ParseVerdict(out))
	}
	if strings.Count(out, "I'll wait") != 2 || !strings.Contains(out, "## Attempt 2") {
		t.Fatalf("both runs not kept:\n%s", out)
	}
}

// D2: a run whose stream reports an API error is "review did not run", with
// the sandbox action, and isn't retried (the block would repeat).
func TestDispatchAPIErrorIsReviewDidNotRun(t *testing.T) {
	calls := fakeRuns(t, streamOut(t, "api_error.jsonl"))
	_, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "review"})
	if !errors.Is(err, ErrAPIUnreachable) || !strings.Contains(err.Error(), "ERR_PROXY_TUNNEL") || !strings.Contains(err.Error(), "allow api.anthropic.com") || *calls != 1 {
		t.Fatalf("err %v, %d dispatches", err, *calls)
	}
}

// The failure is judged on the channel: a review quoting the signatures is a
// review; a text-mode agent's stderr signature without a verdict is not.
func TestAPIFailure(t *testing.T) {
	quoting := AgentRun{Stream: true, Messages: []string{"The fix detects `API Error` and ERR_PROXY_TUNNEL.\n\nVERDICT: SHIP"}, Result: &Result{Text: "VERDICT: SHIP"}}
	for _, c := range []struct {
		name    string
		run     AgentRun
		stderr  string
		nonZero bool
		want    bool
	}{
		{"stream error result", ReadStream(fixture(t, "api_error.jsonl")), "", true, true},
		{"review quoting the strings", quoting, "", false, false},
		{"clean stream", ReadStream(fixture(t, "clean.jsonl")), "", false, false},
		{"text agent, stderr signature, no verdict", AgentRun{Tail: "error"}, "fetch failed: ENOTFOUND api.example", true, true},
		{"text agent, stdout signature, non-zero exit", AgentRun{Tail: "API Error: 403 forbidden"}, "", true, true},
		{"text agent, stdout signature, clean exit", AgentRun{Tail: "API Error: 403 forbidden"}, "", false, false},
		{"text agent with a verdict", AgentRun{Tail: "API Error mentioned\nVERDICT: CLEAN"}, "ECONNREFUSED", true, false},
	} {
		if _, got := APIFailure(c.run, []byte(c.stderr), c.nonZero); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// D3: WF_REVIEW_TIMEOUT wins, then the caller's sized limit, then 30 minutes.
func TestDispatchTimeout(t *testing.T) {
	for _, c := range []struct {
		env   string
		sized time.Duration
		want  time.Duration
	}{
		{"", 0, 30 * time.Minute},
		{"", 45 * time.Minute, 45 * time.Minute},
		{"75m", 45 * time.Minute, 75 * time.Minute},
	} {
		if got, err := dispatchTimeout(c.env, c.sized); err != nil || got != c.want {
			t.Errorf("%+v: %v %v", c, got, err)
		}
	}
}
