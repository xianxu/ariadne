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

// The failure is judged on the channel and never over a verdict: a review
// quoting the signatures is a review, a verdict followed by an error result
// keeps its verdict (BR-1), and only a network cause names the sandbox fix,
// with the agent's own host.
func TestRunFailure(t *testing.T) {
	quoting := AgentRun{Stream: true, Messages: []string{"The fix detects `API Error` and ERR_PROXY_TUNNEL.\n\nVERDICT: SHIP"}, Result: &Result{Text: "VERDICT: SHIP"}}
	maxTurns := AgentRun{Stream: true, Messages: []string{"Still reading the diff."}, Result: &Result{Subtype: "error_max_turns", IsError: true, Text: "reached max turns"}}
	for _, c := range []struct {
		name    string
		agent   AgentCLI
		run     AgentRun
		stderr  string
		nonZero bool
		want    error // nil: not a failure
		host    string
	}{
		{"stream error result", AgentClaude, ReadStream(fixture(t, "api_error.jsonl")), "", true, ErrAPIUnreachable, "api.anthropic.com"},
		{"verdict, then an error result", AgentClaude, ReadStream(fixture(t, "verdict_then_error.jsonl")), "", true, nil, ""},
		{"error result without a network cause", AgentClaude, maxTurns, "", true, ErrReviewDidNotRun, ""},
		{"review quoting the strings", AgentClaude, quoting, "", false, nil, ""},
		{"clean stream", AgentClaude, ReadStream(fixture(t, "clean.jsonl")), "", false, nil, ""},
		{"codex, stderr signature, non-zero exit", AgentCodex, AgentRun{Tail: "error"}, "fetch failed: ENOTFOUND api.example", true, ErrAPIUnreachable, "api.openai.com"},
		{"stderr signature on a clean exit is left to the retry", AgentCodex, AgentRun{Tail: "error"}, "ECONNREFUSED", false, nil, ""},
		{"gemini, output signature, non-zero exit", AgentGemini, AgentRun{Tail: "API Error: 403 forbidden"}, "", true, ErrAPIUnreachable, "generativelanguage.googleapis.com"},
		{"output signature on a clean exit", AgentGemini, AgentRun{Tail: "API Error: 403 forbidden"}, "", false, nil, ""},
		{"text agent with a verdict", AgentCodex, AgentRun{Tail: "API Error mentioned\nVERDICT: CLEAN"}, "ECONNREFUSED", true, nil, ""},
	} {
		err := RunFailure(c.agent, c.run, []byte(c.stderr), c.nonZero)
		switch {
		case c.want == nil && err != nil:
			t.Errorf("%s: %v, want no failure", c.name, err)
		case c.want != nil && !errors.Is(err, c.want):
			t.Errorf("%s: %v, want %v", c.name, err, c.want)
		case c.want == ErrReviewDidNotRun && errors.Is(err, ErrAPIUnreachable):
			t.Errorf("%s: %v names the network for a non-network cause", c.name, err)
		case c.host != "" && !strings.Contains(err.Error(), "allow "+c.host):
			t.Errorf("%s: %v, want the %s remedy", c.name, err, c.host)
		}
	}
}

// BR-2: codex and gemini prose is never parsed as stream events, so a
// quoted event line can't fail their review.
func TestDispatchParsesOnlyClaudeAsAStream(t *testing.T) {
	quoted := "The fixture has this line:\n{\"type\":\"result\",\"is_error\":true,\"result\":\"API Error\"}\n\nVERDICT: CLEAN"
	fakeRuns(t, ProcessOutput{Stdout: []byte(quoted)})
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Prompt: "review"})
	if err != nil || !strings.Contains(out, `{"type":"result"`) {
		t.Fatalf("codex review read as a stream: %v\n%s", err, out)
	}
}

// A retry that runs out of time keeps the first attempt as evidence and
// leaves the fail-safe (no verdict) to the caller.
func TestDispatchRetryDeadlineKeepsTheFirstAttempt(t *testing.T) {
	orig := Run
	t.Cleanup(func() { Run = orig })
	n := 0
	Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (ProcessOutput, error) {
		n++
		if n == 1 {
			return streamOut(t, "background_wait.jsonl"), nil
		}
		<-ctx.Done()
		return ProcessOutput{}, ctx.Err()
	}
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "review", Timeout: 300 * time.Millisecond})
	if err != nil || !strings.Contains(out, "I'll wait") || !strings.Contains(out, "## Attempt 2") || HasVerdict(out) {
		t.Fatalf("err %v, out:\n%s", err, out)
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

// Any error from a retry carries the first attempt's text (the evidence),
// beside the error.
func TestDispatchRetryErrorKeepsTheFirstAttempt(t *testing.T) {
	fakeRuns(t, streamOut(t, "background_wait.jsonl"), streamOut(t, "api_error.jsonl"))
	out, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentClaude, Prompt: "review"})
	if !errors.Is(err, ErrAPIUnreachable) || !strings.Contains(out, "I'll wait") || !strings.Contains(out, "## Attempt 2") {
		t.Fatalf("err %v, out:\n%s", err, out)
	}
}
