package main

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
)

// boundaryUnder runs a boundary review of a claude-driven window under `other`
// with only the agents in up available, recording which CLIs ran.
func boundaryUnder(t *testing.T, up map[judge.AgentCLI]bool) (reviewResult, []string, string) {
	t.Helper()
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	t.Setenv("AGENT_CMD", "")
	t.Setenv("CODEX_CI", "")
	t.Setenv("CODEX_THREAD_ID", "")
	t.Setenv("PAIR_AGENT", "claude")
	p := boundaryWindow(t)
	p.PlansDir = t.TempDir()
	writeReviewConfig(t, ".", `{"boundary_reviewer": "other"}`)
	origProbe := reviewbox.Probe
	t.Cleanup(func() { reviewbox.Probe = origProbe })
	reviewbox.Probe = func() (reviewbox.Fence, string) { return reviewbox.FenceNone, "test" }
	origAvail := reviewAvailability
	t.Cleanup(func() { reviewAvailability = origAvail })
	reviewAvailability = func(judge.Containment) judge.Availability {
		return func(a judge.AgentCLI) (string, bool) {
			if up[a] {
				return "", true
			}
			return string(a) + " is not on PATH", false
		}
	}
	(&fakeReviewer{}).install(t)
	inner := judge.Run
	var ran []string
	judge.Run = func(ctx context.Context, onStart func(int), n string, a ...string) (judge.ProcessOutput, error) {
		agent, _, _ := unwrapReviewArgv(t, n, a)
		ran = append(ran, agent)
		return inner(ctx, onStart, n, a...)
	}
	var errb bytes.Buffer
	res := dispatchBoundaryReview(&bytes.Buffer{}, &errb, p)
	return res, ran, errb.String()
}

// #409 end to end at the boundary: a reviewer that cannot launch here is
// skipped visibly, the next family reviews, and the trailer and ledger say so.
func TestBoundaryReviewSkipsAnUnavailableReviewerVisibly(t *testing.T) {
	res, ran, stderr := boundaryUnder(t, map[judge.AgentCLI]bool{judge.AgentGemini: true})
	// Only gemini ran (the same-agent response retry may run it twice).
	if res.Agent != "gemini" || len(ran) == 0 || strings.Trim(strings.ReplaceAll(strings.Join(ran, ","), "gemini", ""), ",") != "" {
		t.Fatalf("reviewed by %q (ran %v), want gemini\n%s", res.Agent, ran, stderr)
	}
	if !strings.Contains(stderr, "reviewer codex (openai) unavailable: codex is not on PATH") {
		t.Errorf("the skip was not announced:\n%s", stderr)
	}
	want := "gemini (google; author anthropic); degraded: codex unavailable: codex is not on PATH"
	if res.Reviewer != want {
		t.Errorf("reviewer record %q, want %q", res.Reviewer, want)
	}
	var trailer string
	for _, l := range reviewTrailers(res) {
		if strings.HasPrefix(l, "Review-Reviewer: ") {
			trailer = l
		}
	}
	if trailer != "Review-Reviewer: "+want {
		t.Errorf("trailer %q", trailer)
	}
}

// No reviewer of another family can run: the review refuses, naming each,
// and never falls back to the author's own family.
func TestBoundaryReviewRefusesWithNoOtherFamily(t *testing.T) {
	res, ran, stderr := boundaryUnder(t, map[judge.AgentCLI]bool{judge.AgentClaude: true})
	if res.Verdict != judge.VerdictNotRun || len(ran) != 0 {
		t.Fatalf("verdict %s, ran %v; want not-run with nothing launched\n%s", res.Verdict, ran, stderr)
	}
	for _, want := range []string{"no reviewer available", "codex: codex is not on PATH", "gemini: gemini is not on PATH"} {
		if !strings.Contains(res.Reason, want) {
			t.Errorf("refusal %q lacks %q", res.Reason, want)
		}
	}
	if res.Agent == "claude" {
		t.Error("the author's own family reviewed")
	}
}

// The trailer is a git trailer the week's tally can read back (Done-when).
func TestReviewReviewerTrailerParsesAsAGitTrailer(t *testing.T) {
	msg := "#409: close\n\nbody\n\nReview-Verdict: SHIP\nReview-Reviewer: codex/gpt-6-astra (openai; author anthropic)\n" +
		"Review-Reviewer: gemini (google; author anthropic); degraded: codex unavailable: usage limit\n"
	cmd := exec.Command("git", "interpret-trailers", "--parse")
	cmd.Stdin = strings.NewReader(msg)
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	var values []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if v, ok := strings.CutPrefix(line, "Review-Reviewer: "); ok {
			values = append(values, v)
		}
	}
	want := []string{"codex/gpt-6-astra (openai; author anthropic)", "gemini (google; author anthropic); degraded: codex unavailable: usage limit"}
	if strings.Join(values, "\n") != strings.Join(want, "\n") {
		t.Fatalf("parsed %q, want %q", values, want)
	}
}
