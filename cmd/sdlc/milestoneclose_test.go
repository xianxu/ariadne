package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func TestMilestoneCloseCmd_Registered(t *testing.T) {
	cmd := NewMilestoneCloseCmd()
	if cmd.Use != "milestone-close" {
		t.Errorf("Use = %q want milestone-close", cmd.Use)
	}
	for _, want := range []string{"issue", "milestone", "actual", "verified", "force", "dry-run", "no-judge", "agent"} {
		if cmd.Flags().Lookup(want) == nil {
			t.Errorf("flag --%s not registered", want)
		}
	}
}

func TestDispatchMilestoneReview_PromptBuildAndDispatch(t *testing.T) {
	// Replace judge.Run with a fake that captures the args, returns a
	// "clean" response. Lets us verify the dispatch path without
	// invoking a real agent.
	orig := judge.Run
	defer func() { judge.Run = orig }()

	var seenName string
	var seenArgs []string
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
		seenName = name
		seenArgs = args
		return judge.ProcessOutput{Stdout: []byte("No DRY violations found.\n")}, nil
	}

	prompt := judge.BuildPrompt(judge.MilestoneReview, judge.PromptInput{
		Diff:     "+ new code\n",
		Base:     "abc123",
		Head:     "HEAD",
		IssueRef: "ariadne#31 M6",
	})
	if !strings.Contains(prompt, "ariadne#31 M6") {
		t.Errorf("prompt missing issue ref:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Base: abc123") {
		t.Errorf("prompt missing base line:\n%s", prompt)
	}

	out, err := judge.Dispatch(context.Background(), judge.DispatchOptions{
		Agent:        judge.AgentClaude,
		Prompt:       prompt,
		AllowedTools: judge.MilestoneReview.AllowedTools(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if judge.Classify(out) != judge.Clean {
		t.Errorf("expected Clean classification, got %s", judge.Classify(out))
	}
	if seenName != "claude" {
		t.Errorf("dispatch agent = %q want claude", seenName)
	}
	if seenArgs[len(seenArgs)-1] != prompt {
		t.Errorf("dispatch prompt not last arg")
	}
}

func TestDispatchBoundaryReview_AgentDefaultUsesPairAgent(t *testing.T) {
	issuesDir := closeRepo(t, 31)
	head := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")
	orig := judge.Run
	defer func() { judge.Run = orig }()

	var seenName string
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
		seenName = name
		return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n")}, nil
	}

	res := dispatchBoundaryReview(io.Discard, io.Discard, boundaryReviewParams{
		Label:     "#31 M1",
		Base:      shortSHA(head),
		BaseLong:  head,
		Head:      head,
		IssuesDir: issuesDir,
		IssueNum:  31,
		Milestone: "M1",
	})
	if res.Verdict != judge.VerdictShip {
		t.Fatalf("verdict = %s, want SHIP", res.Verdict)
	}
	if seenName != "codex" {
		t.Fatalf("milestone boundary review agent = %q, want codex", seenName)
	}
}

// TestEmitTrailerBlock_Shape exercises the format the milestone-close
// commit's trailer block must take. The block is what `sdlc close`
// later greps for via `git log --grep "Review-Verdict:"`, so any drift
// from this shape breaks the verifier in close.go.
func TestEmitTrailerBlock_Shape(t *testing.T) {
	tests := []struct {
		name     string
		result   reviewResult
		contains []string
		absent   []string
	}{
		{
			name: "ship verdict",
			result: reviewResult{
				Verdict: judge.VerdictShip,
				Base:    "abc1234",
				Head:    "HEAD",
			},
			contains: []string{
				"── milestone-close trailers",
				"Review-Verdict: SHIP",
				"Review-Window: abc1234..HEAD",
			},
			absent: []string{"Review-Reason:"},
		},
		{
			name: "not-run carries reason",
			result: reviewResult{
				Verdict: judge.VerdictNotRun,
				Reason:  "--no-judge",
				Base:    "def5678",
				Head:    "HEAD",
			},
			contains: []string{
				"Review-Verdict: not-run",
				"Review-Window: def5678..HEAD",
				"Review-Reason: --no-judge",
			},
		},
		{
			name: "fix-then-ship verdict",
			result: reviewResult{
				Verdict: judge.VerdictFixThenShip,
				Base:    "111aaaa",
				Head:    "HEAD",
			},
			contains: []string{"Review-Verdict: FIX-THEN-SHIP"},
		},
		{
			name: "rework verdict",
			result: reviewResult{
				Verdict: judge.VerdictRework,
				Base:    "222bbbb",
				Head:    "HEAD",
			},
			contains: []string{"Review-Verdict: REWORK"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			emitTrailerBlock(&buf, tt.result, "milestone-close")
			out := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(out, want) {
					t.Errorf("emitted block missing %q:\n%s", want, out)
				}
			}
			for _, bad := range tt.absent {
				if strings.Contains(out, bad) {
					t.Errorf("emitted block should NOT contain %q:\n%s", bad, out)
				}
			}
			// Blank line before the trailers (git interpret-trailers
			// expects a blank separating body from trailer block).
			if !strings.Contains(out, "\n\nReview-Verdict:") {
				t.Errorf("expected blank line before Review-Verdict trailer:\n%s", out)
			}
		})
	}
}

// TestAppendVerdictSuffix exercises the post-judge mutation that mirrors
// the verdict into the issue file's log line for human grep.
func TestAppendVerdictSuffix(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		milestone string
		verdict   judge.Verdict
		want      string
		wantOK    bool
	}{
		{
			name:      "appends to matching milestone line",
			body:      "## Log\n\n- 2026-05-26: closed M3 — tests pass; live `sdlc state` shows M1-M3\n",
			milestone: "M3",
			verdict:   judge.VerdictShip,
			want:      "## Log\n\n- 2026-05-26: closed M3 — tests pass; live `sdlc state` shows M1-M3; review verdict: SHIP\n",
			wantOK:    true,
		},
		{
			name:      "picks the right milestone when multiple lines exist",
			body:      "- 2026-05-26: closed M2 — older\n- 2026-05-26: closed M3 — newer\n",
			milestone: "M3",
			verdict:   judge.VerdictFixThenShip,
			want:      "- 2026-05-26: closed M2 — older\n- 2026-05-26: closed M3 — newer; review verdict: FIX-THEN-SHIP\n",
			wantOK:    true,
		},
		{
			name:      "idempotent on re-run (already has suffix)",
			body:      "- 2026-05-26: closed M3 — tests pass; review verdict: SHIP\n",
			milestone: "M3",
			verdict:   judge.VerdictShip,
			want:      "- 2026-05-26: closed M3 — tests pass; review verdict: SHIP\n",
			wantOK:    true,
		},
		{
			name:      "returns false when no matching line",
			body:      "- 2026-05-26: closed M2 — only this one\n",
			milestone: "M3",
			verdict:   judge.VerdictShip,
			want:      "- 2026-05-26: closed M2 — only this one\n",
			wantOK:    false,
		},
		{
			name:      "writes not-run verdict",
			body:      "- 2026-05-26: closed M4 — docs only\n",
			milestone: "M4",
			verdict:   judge.VerdictNotRun,
			want:      "- 2026-05-26: closed M4 — docs only; review verdict: not-run\n",
			wantOK:    true,
		},
		{
			// #69: whole-issue close has no milestone → "closed — ..." line.
			name:      "issue close (empty milestone) matches 'closed — '",
			body:      "- 2026-06-03: closed — all tests pass\n",
			milestone: "",
			verdict:   judge.VerdictShip,
			want:      "- 2026-06-03: closed — all tests pass; review verdict: SHIP\n",
			wantOK:    true,
		},
		{
			// Guard: empty-milestone matcher must NOT match a milestone line.
			name:      "issue-close matcher skips a milestone close line",
			body:      "- 2026-06-03: closed M2 — slice done\n",
			milestone: "",
			verdict:   judge.VerdictShip,
			want:      "- 2026-06-03: closed M2 — slice done\n",
			wantOK:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := appendVerdictSuffix(tt.body, tt.milestone, tt.verdict)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

// TestDispatchMilestoneReview_VerdictCapture verifies the verdict is
// parsed from the fake judge's output and surfaced in the reviewResult.
// Mirrors TestDispatchMilestoneReview_PromptBuildAndDispatch above but
// asserts on the verdict-tracking shape rather than the dispatch shape.
func TestDispatchMilestoneReview_VerdictCapture(t *testing.T) {
	orig := judge.Run
	defer func() { judge.Run = orig }()

	cases := []struct {
		name        string
		agentOutput string
		want        judge.Verdict
	}{
		{"ship", "SHIP (confidence: high)\n\nLooks good.\n", judge.VerdictShip},
		{"fix-then-ship", "FIX-THEN-SHIP (confidence: medium)\nOne tweak.\n", judge.VerdictFixThenShip},
		{"rework", "REWORK (confidence: low)\nNeeds restructure.\n", judge.VerdictRework},
		{"unknown", "Looks fine, ship it.\n", judge.VerdictUnknown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
				return judge.ProcessOutput{Stdout: []byte(tc.agentOutput)}, nil
			}
			out, err := judge.Dispatch(context.Background(), judge.DispatchOptions{
				Agent:        judge.AgentClaude,
				Prompt:       "p",
				AllowedTools: judge.MilestoneReview.AllowedTools(),
			})
			if err != nil {
				t.Fatal(err)
			}
			if got := judge.ParseVerdict(out); got != tc.want {
				t.Errorf("ParseVerdict() = %s want %s", got, tc.want)
			}
		})
	}
}

func TestMilestoneCloseFlagSurface(t *testing.T) {
	// Flag validation lives in runMilestoneClose (not via cobra
	// MarkFlagRequired) so error format matches die()'s red prefix —
	// same posture as close, judge, set-status, etc. Confirm we
	// didn't accidentally mark --milestone required at the cobra layer.
	cmd := NewMilestoneCloseCmd()
	f := cmd.Flags().Lookup("milestone")
	if f == nil {
		t.Fatal("--milestone not registered")
	}
	if _, ok := f.Annotations["cobra_annotation_bash_completion_one_required_flag"]; ok {
		t.Errorf("--milestone should NOT use MarkFlagRequired (use runtime die() instead)")
	}
}
