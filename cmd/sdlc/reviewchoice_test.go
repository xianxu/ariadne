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

func TestAuthorFamiliesFromLog(t *testing.T) {
	rec := func(sha, parents, subject string, agents ...string) string {
		return sha + "\x1f" + parents + "\x1f" + subject + "\x1f" + strings.Join(agents, "\x1e") + "\x1d"
	}
	families := func(fs []judge.Family) string {
		var s []string
		for _, f := range fs {
			s = append(s, string(f))
		}
		return strings.Join(s, ",")
	}
	for _, tc := range []struct {
		name string
		log  string
		want string // the session family is always first
	}{
		{"codex trailer with model and effort", rec("a1", "p0", "x", "codex/gpt-6-astra/medium"), "anthropic,openai"},
		{"two commits, two families", rec("a1", "p0", "x", "codex/gpt/med") + "\n" + rec("a2", "a1", "y", "gemini/g/high"), "anthropic,openai,google"},
		{"a merge contributes nothing", rec("m1", "p0 p1", "merge", "codex/gpt/med"), "anthropic"},
		{"an unknown agent contributes nothing", rec("a1", "p0", "x", "qoder/q/low"), "anthropic"},
		{"a bare agent name counts", rec("a1", "p0", "x", "gemini"), "anthropic,google"},
		{"no trailers", rec("a1", "p0", "x"), "anthropic"},
	} {
		if got := families(authorFamiliesFromLog(tc.log, judge.FamilyAnthropic)); got != tc.want {
			t.Errorf("%s: %s, want %s", tc.name, got, tc.want)
		}
	}
}

// The git path end to end: a codex-trailered commit in the window counts, and
// a window that cannot be read is recorded as degraded rather than ignored.
func TestReviewAuthorFamiliesReadsTheWindow(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("CODEX_CI", "")
	t.Setenv("CODEX_THREAD_ID", "")
	t.Setenv("PAIR_AGENT", "claude")
	boundaryWindow(t) // a committed repository, the working directory
	run := func(args ...string) string {
		t.Helper()
		out, err := exec.Command("git", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	base := run("rev-parse", "HEAD")
	run("commit", "-q", "--allow-empty", "-m", "codex wrote this", "--trailer", "Agent: codex/gpt-6-astra/medium")
	fams, unread := reviewAuthorFamilies(context.Background(), base, "HEAD")
	if unread != "" || len(fams) != 2 || fams[0] != judge.FamilyAnthropic || fams[1] != judge.FamilyOpenAI {
		t.Fatalf("families %v (unread %q), want anthropic,openai", fams, unread)
	}
	if _, unread := reviewAuthorFamilies(context.Background(), "no-such-base", "HEAD"); !strings.HasPrefix(unread, "authors unread") {
		t.Errorf("an unreadable window was not reported: %q", unread)
	}
	choice := resolveReviewChoiceWith(context.Background(), judge.ReviewerPolicy{Mode: judge.ReviewerOther}, "", "HEAD", "", false, judge.KindBounded, judge.ContainNone)
	first, _ := choice.First()
	if rec := judge.ReviewerRecord(first, choice, nil); !strings.Contains(rec, "degraded: authors unread") {
		t.Errorf("unread authors not recorded as degraded: %q", rec)
	}
}
