package judge

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// fakeAgents answers each agent CLI by name: "launch" fails to start, "quota"
// fails before reviewing, "began" fails after starting, anything else reviews.
func fakeAgents(t *testing.T, behaviour map[string]string) *[]string {
	t.Helper()
	var called []string
	prev := Run
	Run = func(ctx context.Context, onStart func(int), name string, args ...string) (ProcessOutput, error) {
		called = append(called, name)
		switch behaviour[name] {
		case "launch":
			return ProcessOutput{}, exec.ErrNotFound
		case "quota":
			if name == "claude" {
				return ProcessOutput{Stdout: []byte(`{"type":"result","subtype":"error","is_error":true,"result":"API Error: 429 rate limit"}`)}, &exec.ExitError{}
			}
			return ProcessOutput{Stderr: []byte(`ERROR: {"type":"error","status":429,"error":{"message":"You've hit your usage limit."}}`)}, &exec.ExitError{}
		case "began":
			return ProcessOutput{Stdout: []byte("partial review, then the connection dropped")}, &exec.ExitError{}
		}
		return ProcessOutput{Stdout: []byte("VERDICT: SHIP\n")}, nil
	}
	t.Cleanup(func() { Run = prev })
	return &called
}

func choiceOf(agents ...AgentCLI) ReviewChoice {
	c := ReviewChoice{Authors: []Family{FamilyAnthropic}}
	for _, a := range agents {
		c.Candidates = append(c.Candidates, Candidate{Agent: a, Family: FamilyOf(string(a))})
	}
	return c
}

func TestDispatchChoice(t *testing.T) {
	opts := func(stderr *bytes.Buffer) DispatchOptions {
		return DispatchOptions{Prompt: "review", AllowedTools: "Read", Stderr: stderr}
	}
	t.Run("an unavailable first candidate moves to the next, visibly", func(t *testing.T) {
		called := fakeAgents(t, map[string]string{"codex": "quota"})
		var stderr bytes.Buffer
		out, used, skips, err := DispatchChoice(context.Background(), opts(&stderr), choiceOf(AgentCodex, AgentGemini))
		if err != nil || used.Agent != AgentGemini || !HasVerdict(out) {
			t.Fatalf("used %s, err %v, out %q", used.Agent, err, out)
		}
		if len(skips) != 1 || skips[0].Agent != AgentCodex || !strings.Contains(skips[0].Reason, "usage limit") {
			t.Errorf("skips %+v", skips)
		}
		if !strings.Contains(stderr.String(), "reviewer codex (openai) unavailable") {
			t.Errorf("the skip was not announced:\n%s", stderr.String())
		}
		if strings.Join(*called, ",") != "codex,gemini" {
			t.Errorf("calls %v", *called)
		}
	})
	t.Run("a launch failure is unavailability too", func(t *testing.T) {
		fakeAgents(t, map[string]string{"codex": "launch"})
		_, used, skips, err := DispatchChoice(context.Background(), opts(&bytes.Buffer{}), choiceOf(AgentCodex, AgentGemini))
		if err != nil || used.Agent != AgentGemini || len(skips) != 1 {
			t.Fatalf("used %s, skips %+v, err %v", used.Agent, skips, err)
		}
		// The recorded reason leads with the cause and keeps the operator's PATH
		// out of the trailer and ledger; the full diagnosis is only in Detail.
		if r := skips[0].Reason; !strings.HasPrefix(r, "could not launch: "+exec.ErrNotFound.Error()) || strings.Contains(r, "PATH=") {
			t.Errorf("launch reason %q", r)
		}
	})
	t.Run("a launch failure's error keeps the full diagnosis", func(t *testing.T) {
		fakeAgents(t, map[string]string{"codex": "launch"})
		_, err := Dispatch(context.Background(), DispatchOptions{Agent: AgentCodex, Prompt: "review", AllowedTools: "Read"})
		var unavailable *UnavailableError
		if !errors.As(err, &unavailable) || !strings.Contains(err.Error(), "PATH=") || strings.Contains(unavailable.Reason, "PATH=") {
			t.Fatalf("err %v (reason %q)", err, unavailable)
		}
	})
	t.Run("every candidate unavailable refuses, naming each", func(t *testing.T) {
		fakeAgents(t, map[string]string{"codex": "quota", "gemini": "launch"})
		choice := choiceOf(AgentCodex, AgentGemini)
		choice.Skips = []Skip{{Agent: AgentClaude, Reason: "claude is not on PATH"}}
		_, _, skips, err := DispatchChoice(context.Background(), opts(&bytes.Buffer{}), choice)
		var none *NoReviewerError
		if !errors.As(err, &none) || !errors.Is(err, ErrReviewDidNotRun) {
			t.Fatalf("err %v, want a NoReviewerError", err)
		}
		if len(skips) != 3 || !strings.Contains(err.Error(), "codex:") || !strings.Contains(err.Error(), "gemini:") || !strings.Contains(err.Error(), "claude:") {
			t.Errorf("refusal does not name every candidate: %v (skips %+v)", err, skips)
		}
	})
	t.Run("a reviewer that began never hands over to another", func(t *testing.T) {
		called := fakeAgents(t, map[string]string{"codex": "began"})
		_, used, skips, _ := DispatchChoice(context.Background(), opts(&bytes.Buffer{}), choiceOf(AgentCodex, AgentGemini))
		// The existing missing-verdict retry stays on the same agent.
		if used.Agent != AgentCodex || len(skips) != 0 || strings.Contains(strings.Join(*called, ","), "gemini") {
			t.Fatalf("used %s, skips %+v, calls %v", used.Agent, skips, *called)
		}
	})
	t.Run("no candidates at all refuses", func(t *testing.T) {
		fakeAgents(t, nil)
		_, _, _, err := DispatchChoice(context.Background(), opts(&bytes.Buffer{}), ReviewChoice{})
		if !errors.As(err, new(*NoReviewerError)) {
			t.Fatalf("err %v", err)
		}
	})
}

// A reviewer whose first attempt ran and whose retry could not run did review:
// the review ends there and never hands over to another family (#409 BR-9).
func TestUnavailableRetryNeverHandsOver(t *testing.T) {
	var called []string
	prev := Run
	Run = func(ctx context.Context, onStart func(int), name string, args ...string) (ProcessOutput, error) {
		called = append(called, name)
		if len(called) == 1 { // attempt 1 reviewed, but without a verdict
			return ProcessOutput{Stdout: []byte("reviewed some files, no verdict")}, nil
		}
		return ProcessOutput{Stderr: []byte(`ERROR: {"type":"error","status":429}`)}, &exec.ExitError{}
	}
	t.Cleanup(func() { Run = prev })
	_, used, skips, err := DispatchChoice(context.Background(), DispatchOptions{Prompt: "review", AllowedTools: "Read"}, choiceOf(AgentCodex, AgentGemini))
	if strings.Join(called, ",") != "codex,codex" || used.Agent != AgentCodex || len(skips) != 0 {
		t.Fatalf("called %v, used %s, skips %+v, err %v", called, used.Agent, skips, err)
	}
	if err == nil || errors.Is(err, ErrReviewerUnavailable) {
		t.Errorf("the retry's failure should end the review as did-not-run, got %v", err)
	}
}

// The refusal every sdlc site reports keeps judge's contract across the move
// to pkg/reviewpolicy (#441 BR-1): a NoReviewerError, a review that did not
// run, and the --agent remedy.
func TestRefusalKeepsJudgesContract(t *testing.T) {
	choice := ResolveReview("", false, AgentDefaultEnv{ClaudeCode: "1"}, ReviewerPolicy{Mode: ReviewerOther}, []Family{FamilyAnthropic}, KindBounded,
		func(a AgentCLI) (string, bool) { return string(a) + " missing", false })
	err := Refusal(choice)
	var none *NoReviewerError
	if !errors.As(err, &none) || !errors.Is(err, ErrReviewDidNotRun) {
		t.Fatalf("refusal %v does not match both", err)
	}
	if !strings.Contains(err.Error(), "name one with --agent") || !strings.Contains(err.Error(), "codex: codex missing") {
		t.Errorf("refusal text %q", err)
	}
}
