package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
)

func writeReviewConfig(t *testing.T, root, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(reviewConfigPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadReviewerPolicy(t *testing.T) {
	root := t.TempDir()
	if p, err := loadReviewerPolicy(root); err != nil || p.Mode != "" {
		t.Fatalf("absent config: %+v, %v; want the zero (same) policy", p, err)
	}
	writeReviewConfig(t, root, `{"boundary_reviewer": "explicit", "reviewer_pairs": {"claude": "gemini"}}`)
	p, err := loadReviewerPolicy(root)
	if err != nil || p.Reviewer(judge.AgentClaude) != judge.AgentGemini {
		t.Fatalf("explicit config: %+v, %v", p, err)
	}
	for _, bad := range []string{
		`{"boundary_reviewer": "others"}`, // unknown mode
		`{"boundary_reviewr": "other"}`,   // misspelled key
		`{"boundary_reviewer": `,          // malformed
	} {
		writeReviewConfig(t, root, bad)
		if _, err := loadReviewerPolicy(root); err == nil || !strings.Contains(err.Error(), reviewConfigPath) {
			t.Errorf("%s: err = %v, want one naming %s", bad, err, reviewConfigPath)
		}
	}
}

// #310's contract end to end: with `other` configured, a claude-driven boundary
// is reviewed by codex and a codex-driven one by claude, the reviewer runs in
// its non-auto-approve mode, and the ledger round records who reviewed.
func TestBoundaryReviewOtherPolicySwapsReviewer(t *testing.T) {
	for _, tc := range []struct{ main, want, mode string }{
		{"claude", "codex", "--sandbox workspace-write"},
		{"codex", "claude", "--permission-mode dontAsk"},
	} {
		t.Run(tc.main, func(t *testing.T) {
			t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
			t.Setenv("AGENT_CMD", "")
			t.Setenv("CODEX_CI", "")
			t.Setenv("CODEX_THREAD_ID", "")
			t.Setenv("PAIR_AGENT", tc.main)
			p := boundaryWindow(t)
			p.PlansDir = t.TempDir()
			writeReviewConfig(t, ".", `{"boundary_reviewer": "other"}`)
			origProbe := reviewbox.Probe
			t.Cleanup(func() { reviewbox.Probe = origProbe })
			reviewbox.Probe = func() (reviewbox.Fence, string) { return reviewbox.FenceNone, "test" }

			(&fakeReviewer{}).install(t)
			inner := judge.Run
			var agent string
			var agentArgs []string
			judge.Run = func(ctx context.Context, onStart func(int), n string, a ...string) (judge.ProcessOutput, error) {
				agent, agentArgs, _ = unwrapReviewArgv(t, n, a)
				return inner(ctx, onStart, n, a...)
			}
			var errb bytes.Buffer
			res := dispatchBoundaryReview(&bytes.Buffer{}, &errb, p)
			if agent != tc.want || res.Agent != tc.want {
				t.Fatalf("%s main: dispatched %q (result %q), want %q\n%s", tc.main, agent, res.Agent, tc.want, errb.String())
			}
			if warned := strings.Contains(errb.String(), "codex's own sandbox"); warned != (tc.want == "codex" && runtime.GOOS == "darwin") {
				t.Errorf("codex-cannot-nest warning shown=%v for a %s reviewer:\n%s", warned, tc.want, errb.String())
			}
			if !strings.Contains(strings.Join(agentArgs, " "), tc.mode) {
				t.Errorf("%s reviewer argv lacks its permission mode %q: %q", tc.want, tc.mode, agentArgs)
			}
			persistBoundaryRound(&errb, p, res, "2026-10-11T12:00:00-07:00")
			l, err := readBoundaryGateLedger(p.PlansDir, "000069-x.md", 69)
			if err != nil || len(l.Rounds) != 1 || l.Rounds[0].Agent != tc.want {
				t.Fatalf("ledger round agent: %v %+v", err, l.Rounds)
			}
		})
	}
}

// A broken config halts the review with the file named, never a silent `same`.
func TestBoundaryReviewInvalidPolicyDoesNotRun(t *testing.T) {
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	p := boundaryWindow(t)
	writeReviewConfig(t, ".", `{"boundary_reviewer": "others"}`)
	ran := false
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(context.Context, func(int), string, ...string) (judge.ProcessOutput, error) {
		ran = true
		return judge.ProcessOutput{}, nil
	}
	res := dispatchBoundaryReview(&bytes.Buffer{}, &bytes.Buffer{}, p)
	if ran || res.Verdict != judge.VerdictNotRun || !strings.Contains(res.Reason, reviewConfigPath) {
		t.Fatalf("ran=%v verdict=%s reason=%q", ran, res.Verdict, res.Reason)
	}
}
