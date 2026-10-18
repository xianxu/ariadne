package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/agentenv"
	"github.com/xianxu/ariadne/pkg/reviewpolicy"
)

func TestChooseReviewers(t *testing.T) {
	all := func(reviewpolicy.Agent) (string, bool) { return "", true }
	noCodex := func(a reviewpolicy.Agent) (string, bool) {
		if a == AgentCodex {
			return "codex is not on PATH", false
		}
		return "", true
	}
	anth, oai := reviewpolicy.Anthropic, reviewpolicy.OpenAI
	for _, tc := range []struct {
		name     string
		policy   reviewpolicy.Policy
		explicit string
		authors  []reviewpolicy.Family
		avail    reviewpolicy.Availability
		env      agentenv.Env
		want     string // candidate agents in order
	}{
		{"default (same) still cross-reviews a claude doc", reviewpolicy.Policy{}, "", []reviewpolicy.Family{anth}, all, agentenv.Env{}, "codex,gemini"},
		{"same is other for doc-review", reviewpolicy.Policy{Mode: reviewpolicy.Same}, "", []reviewpolicy.Family{oai}, all, agentenv.Env{}, "claude,gemini"},
		{"other, mixed authors leave gemini", reviewpolicy.Policy{Mode: reviewpolicy.Other}, "", []reviewpolicy.Family{anth, oai}, all, agentenv.Env{}, "gemini"},
		{"explicit pairs are honored", reviewpolicy.Policy{Mode: reviewpolicy.Explicit, Pairs: map[reviewpolicy.Agent]reviewpolicy.Agent{AgentClaude: AgentGemini}}, "", []reviewpolicy.Family{anth}, all, agentenv.Env{}, "gemini"},
		{"a missing CLI is a skip", reviewpolicy.Policy{}, "", []reviewpolicy.Family{anth}, noCodex, agentenv.Env{}, "gemini"},
		{"the command-line agent wins", reviewpolicy.Policy{}, "claude", []reviewpolicy.Family{anth}, all, agentenv.Env{}, "claude"},
		{"AGENT_CMD is sdlc's, not doc-review's", reviewpolicy.Policy{}, "", []reviewpolicy.Family{anth}, all, agentenv.Env{AgentCmd: "claude"}, "codex,gemini"},
		{"explicit, the session has no pair: no self-review", reviewpolicy.Policy{Mode: reviewpolicy.Explicit, Pairs: map[reviewpolicy.Agent]reviewpolicy.Agent{AgentCodex: AgentClaude}}, "", []reviewpolicy.Family{anth}, all, agentenv.Env{PairAgent: "claude"}, ""},
		{"explicit pair back into an author family is skipped", reviewpolicy.Policy{Mode: reviewpolicy.Explicit, Pairs: map[reviewpolicy.Agent]reviewpolicy.Agent{AgentClaude: AgentClaude}}, "", []reviewpolicy.Family{anth}, all, agentenv.Env{PairAgent: "claude"}, ""},
	} {
		c := chooseReviewers(tc.policy, tc.explicit, tc.explicit != "", tc.env, tc.authors, tc.avail)
		var got []string
		for _, cand := range c.Candidates {
			got = append(got, string(cand.Agent))
			if !c.Forced {
				for _, f := range tc.authors {
					if cand.Family == f {
						t.Errorf("%s: %s shares the author family %s", tc.name, cand.Agent, f)
					}
				}
			}
		}
		if strings.Join(got, ",") != tc.want {
			t.Errorf("%s: candidates %v, want %s", tc.name, got, tc.want)
		}
	}
}

// docRepo commits a document authored by agent (an Agent: trailer) in a fresh
// repository, with the session running as session.
func docRepo(t *testing.T, agent, session, reviewJSON string) string {
	t.Helper()
	for _, k := range []string{"AGENT_CMD", "CODEX_CI", "CODEX_THREAD_ID", "CLAUDECODE", "CLAUDE_CODE_ENTRYPOINT"} {
		t.Setenv(k, "")
	}
	t.Setenv("PAIR_AGENT", session)
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "T")
	run("config", "commit.gpgsign", "false")
	if reviewJSON != "" {
		if err := os.MkdirAll(filepath.Join(dir, "workshop"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "workshop", "review.json"), []byte(reviewJSON), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	doc := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(doc, []byte("# Doc\nclaim [1].\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	run("commit", "-q", "-m", "write the doc", "--trailer", "Agent: "+agent+"/model/medium")
	return doc
}

// agentsBehave makes each reviewer CLI review, fail to launch, or crash, and
// records which ran; missing CLIs are not on PATH.
func agentsBehave(t *testing.T, behaviour map[string]string, missing ...string) *[]string {
	t.Helper()
	var ran []string
	prevRun, prevLook := runAgent, lookPath
	t.Cleanup(func() { runAgent, lookPath = prevRun, prevLook })
	lookPath = func(name string) (string, error) {
		for _, m := range missing {
			if m == name {
				return "", exec.ErrNotFound
			}
		}
		return "/bin/" + name, nil
	}
	runAgent = func(_ context.Context, name string, args ...string) ([]byte, error) {
		ran = append(ran, name)
		switch behaviour[name] {
		case "launch":
			return nil, errors.New("exec: " + name + ": executable file not found")
		case "crash":
			return []byte("panic: gpt-image-2"), &exec.ExitError{}
		}
		for i, a := range args {
			if a == "-o" && i+1 < len(args) {
				_ = os.WriteFile(args[i+1], []byte("## Summary\nfine"), 0o644)
			}
		}
		return []byte("## Summary\nfine"), nil
	}
	return &ran
}

func TestRunReviewChoosesFallsBackAndRecords(t *testing.T) {
	read := func(t *testing.T, path string) string {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("report %s: %v", path, err)
		}
		return string(b)
	}
	t.Run("a codex-authored doc is reviewed by claude", func(t *testing.T) {
		doc := docRepo(t, "codex", "codex", "")
		ran := agentsBehave(t, nil)
		var out bytes.Buffer
		if err := runReview(&out, &bytes.Buffer{}, &reviewFlags{File: doc}); err != nil {
			t.Fatal(err)
		}
		if strings.Join(*ran, ",") != "claude" {
			t.Fatalf("ran %v, want claude", *ran)
		}
		if r := read(t, strings.TrimSuffix(doc, ".md")+"-claude-check.md"); !strings.Contains(r, `reviewer_record: "claude (anthropic; author openai)"`) {
			t.Errorf("report record:\n%s", r)
		}
	})
	t.Run("codex missing: gemini reviews, recorded as degraded", func(t *testing.T) {
		doc := docRepo(t, "claude", "claude", "")
		ran := agentsBehave(t, nil, "codex")
		var errb bytes.Buffer
		if err := runReview(&bytes.Buffer{}, &errb, &reviewFlags{File: doc}); err != nil {
			t.Fatal(err)
		}
		if strings.Join(*ran, ",") != "gemini" || !strings.Contains(errb.String(), "reviewer codex (openai) unavailable: codex is not on PATH") {
			t.Fatalf("ran %v\n%s", *ran, errb.String())
		}
		if r := read(t, strings.TrimSuffix(doc, ".md")+"-gemini-check.md"); !strings.Contains(r, "degraded: codex unavailable: codex is not on PATH") {
			t.Errorf("report record:\n%s", r)
		}
	})
	t.Run("a failed launch moves on", func(t *testing.T) {
		doc := docRepo(t, "claude", "claude", "")
		ran := agentsBehave(t, map[string]string{"codex": "launch"})
		if err := runReview(&bytes.Buffer{}, &bytes.Buffer{}, &reviewFlags{File: doc}); err != nil {
			t.Fatal(err)
		}
		if strings.Join(*ran, ",") != "codex,gemini" {
			t.Fatalf("ran %v", *ran)
		}
	})
	t.Run("a crash does not hand over", func(t *testing.T) {
		doc := docRepo(t, "claude", "claude", "")
		ran := agentsBehave(t, map[string]string{"codex": "crash"})
		err := runReview(&bytes.Buffer{}, &bytes.Buffer{}, &reviewFlags{File: doc})
		if err == nil || !strings.Contains(err.Error(), "doc-review gemini") || strings.Join(*ran, ",") != "codex" {
			t.Fatalf("ran %v, err %v", *ran, err)
		}
	})
	t.Run("nobody of another family: refuse, naming each", func(t *testing.T) {
		doc := docRepo(t, "claude", "claude", "")
		ran := agentsBehave(t, nil, "codex", "gemini")
		err := runReview(&bytes.Buffer{}, &bytes.Buffer{}, &reviewFlags{File: doc})
		if !errors.Is(err, reviewpolicy.ErrNoReviewer) || len(*ran) != 0 || !strings.Contains(err.Error(), "codex: codex is not on PATH") || !strings.Contains(err.Error(), "gemini: gemini is not on PATH") {
			t.Fatalf("ran %v, err %v", *ran, err)
		}
	})
	t.Run("a repo's same still cross-reviews", func(t *testing.T) {
		doc := docRepo(t, "claude", "claude", `{"boundary_reviewer": "same"}`)
		ran := agentsBehave(t, nil)
		if err := runReview(&bytes.Buffer{}, &bytes.Buffer{}, &reviewFlags{File: doc}); err != nil {
			t.Fatal(err)
		}
		if strings.Join(*ran, ",") != "codex" {
			t.Fatalf("ran %v, want codex", *ran)
		}
	})
}

// A refused explicit choice says why: the self-review candidate is a skip.
func TestChooseReviewersRecordsTheSelfReviewSkip(t *testing.T) {
	c := chooseReviewers(reviewpolicy.Policy{Mode: reviewpolicy.Explicit, Pairs: map[reviewpolicy.Agent]reviewpolicy.Agent{AgentCodex: AgentClaude}}, "", false,
		agentenv.Env{PairAgent: "claude"}, []reviewpolicy.Family{reviewpolicy.Anthropic}, func(reviewpolicy.Agent) (string, bool) { return "", true })
	if len(c.Candidates) != 0 || len(c.Skips) != 1 || !strings.Contains(c.Skips[0].Reason, "never self-reviews") {
		t.Fatalf("choice %+v", c)
	}
}

// A named agent that cannot start reports its own launch failure, not advice
// to name one.
func TestRunReviewForcedLaunchFailure(t *testing.T) {
	doc := docRepo(t, "claude", "claude", "")
	agentsBehave(t, map[string]string{"gemini": "launch"})
	err := runReview(&bytes.Buffer{}, &bytes.Buffer{}, &reviewFlags{Agent: AgentGemini, AgentSet: true, File: doc})
	if err == nil || !strings.Contains(err.Error(), "gemini failed to run") || strings.Contains(err.Error(), "name one explicitly") {
		t.Fatalf("err %v", err)
	}
}

// The refusal's remedy follows its skips: a policy skip is not cured by
// waiting, a missing CLI is.
func TestRefusalRemedyFollowsTheSkips(t *testing.T) {
	policyOnly := (&reviewpolicy.NoReviewerError{Skips: []reviewpolicy.Skip{{Agent: AgentClaude, Reason: "shares", Policy: true}}}).Error()
	if strings.Contains(policyOnly, "retry later") || !strings.Contains(policyOnly, "reviewer_pairs, or name one explicitly") {
		t.Errorf("policy-only refusal %q", policyOnly)
	}
	transient := (&reviewpolicy.NoReviewerError{Skips: []reviewpolicy.Skip{{Agent: AgentCodex, Reason: "codex is not on PATH"}}}).Error()
	if !strings.Contains(transient, "retry later, or name one explicitly") || strings.Contains(transient, "reviewer_pairs") {
		t.Errorf("transient refusal %q", transient)
	}
}
