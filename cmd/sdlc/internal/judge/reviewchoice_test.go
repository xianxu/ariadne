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

func firstOf(c ReviewChoice) Candidate {
	f, _ := c.First()
	return f
}

func TestResolveReview(t *testing.T) {
	other := ReviewerPolicy{Mode: ReviewerOther, Models: map[AgentCLI]string{AgentClaude: "opus"}}
	claudeEnv := AgentDefaultEnv{ClaudeCode: "1"}
	c := ResolveReview("", false, claudeEnv, other, []Family{FamilyAnthropic}, KindBounded, nil)
	if c.Forced || firstOf(c).Agent != AgentCodex {
		t.Errorf("planned choice: %+v", c)
	}
	c = ResolveReview("claude", true, claudeEnv, other, []Family{FamilyAnthropic}, KindBounded, nil)
	if !c.Forced || len(c.Candidates) != 1 || firstOf(c).Agent != AgentClaude || firstOf(c).Model != "opus" {
		t.Errorf("--agent choice: %+v", c)
	}
	c = ResolveReview("", false, AgentDefaultEnv{AgentCmd: "gemini"}, other, nil, KindBounded, nil)
	if !c.Forced || firstOf(c).Agent != AgentGemini {
		t.Errorf("AGENT_CMD choice: %+v", c)
	}
	// An empty choice reports no first candidate, and its refusal names the skips.
	none := ResolveReview("", false, claudeEnv, other, []Family{FamilyAnthropic}, KindBounded, func(a AgentCLI) (string, bool) { return string(a) + " missing", false })
	if _, ok := none.First(); ok {
		t.Error("an empty choice offered a first candidate")
	}
	if err := none.Refusal(); !errors.Is(err, ErrReviewDidNotRun) || !strings.Contains(err.Error(), "codex: codex missing") {
		t.Errorf("refusal %v", err)
	}
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

func TestReviewerRecordRoundTrip(t *testing.T) {
	codex := Candidate{Agent: AgentCodex, Model: "gpt-6-astra", Family: FamilyOpenAI}
	gemini := Candidate{Agent: AgentGemini, Family: FamilyGoogle}
	claude := Candidate{Agent: AgentClaude, Family: FamilyAnthropic}
	for _, tc := range []struct {
		name   string
		used   Candidate
		choice ReviewChoice
		skips  []Skip
		want   string
	}{
		{"plain", codex, ReviewChoice{Authors: []Family{FamilyAnthropic}}, nil, "codex/gpt-6-astra (openai; author anthropic)"},
		{"fallback", gemini, ReviewChoice{Authors: []Family{FamilyOpenAI, FamilyAnthropic}}, []Skip{{Agent: AgentCodex, Reason: "usage limit\nmore"}}, "gemini (google; author anthropic,openai); degraded: codex unavailable: usage limit"},
		{"forced same family", claude, ReviewChoice{Authors: []Family{FamilyAnthropic}, Forced: true}, nil, "claude (anthropic; author anthropic); degraded: same-family by choice"},
		{"unknown authors", codex, ReviewChoice{}, nil, "codex/gpt-6-astra (openai; author unknown)"},
	} {
		got := ReviewerRecord(tc.used, tc.choice, tc.skips)
		if got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
			continue
		}
		e, ok := ParseReviewerRecord(got)
		if !ok || e.Agent != string(tc.used.Agent) || e.Family != tc.used.Family || e.Degraded != strings.Contains(tc.want, "degraded") {
			t.Errorf("%s: parsed %+v %v", tc.name, e, ok)
		}
	}
	tally := TallyReviewers([]string{
		"codex/gpt-6-astra (openai; author anthropic)",
		"codex (openai; author anthropic)",
		"claude (anthropic; author openai)",
		"gemini (google; author anthropic); degraded: codex unavailable: quota",
		"not a record",
	})
	want := map[string]int{"anthropic → openai": 2, "openai → anthropic": 1, "anthropic → google (degraded)": 1, "unparsed": 1}
	for k, n := range want {
		if tally[k] != n {
			t.Errorf("tally[%q] = %d, want %d (all %v)", k, tally[k], n, tally)
		}
	}
}
// ReviewerEntry is one parsed Review-Reviewer trailer.
type ReviewerEntry struct {
	Agent, Model string
	Family       Family
	Authors      []Family
	Degraded     bool
}

// ParseReviewerRecord reads a Review-Reviewer value back; ok is false for a
// value not in ReviewerRecord's form.
func ParseReviewerRecord(v string) (ReviewerEntry, bool) {
	head, _, degraded := strings.Cut(v, "; degraded:")
	open := strings.Index(head, " (")
	if open < 0 || !strings.HasSuffix(head, ")") {
		return ReviewerEntry{}, false
	}
	who, inner := head[:open], head[open+2:len(head)-1]
	fam, authors, ok := strings.Cut(inner, "; author ")
	if !ok {
		return ReviewerEntry{}, false
	}
	e := ReviewerEntry{Family: Family(fam), Degraded: degraded}
	e.Agent, e.Model, _ = strings.Cut(who, "/")
	if authors != "unknown" {
		for _, a := range strings.Split(authors, ",") {
			e.Authors = append(e.Authors, Family(a))
		}
	}
	return e, e.Agent != ""
}

// TallyReviewers counts Review-Reviewer values by "author → reviewer" family
// pair, for a week's retro; values not in the form are counted as "unparsed".
func TallyReviewers(values []string) map[string]int {
	out := map[string]int{}
	for _, v := range values {
		e, ok := ParseReviewerRecord(strings.TrimSpace(v))
		if !ok {
			out["unparsed"]++
			continue
		}
		key := valueOrDefault(familyList(e.Authors), "unknown") + " → " + valueOrDefault(string(e.Family), "unknown")
		if e.Degraded {
			key += " (degraded)"
		}
		out[key]++
	}
	return out
}

