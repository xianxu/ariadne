package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJudgeAgentDefault_DryRunUsesPairAgent(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")

	var stdout, stderr bytes.Buffer
	f := &judgeFlags{
		Base:          "HEAD",
		Head:          "HEAD",
		DryRun:        true,
		IssuesDir:     "workshop/issues",
		HistoryDir:    "workshop/history",
		AgentExplicit: false,
	}
	if err := runJudge(&stdout, &stderr, "dry", f); err != nil {
		t.Fatalf("runJudge: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "codex exec") {
		t.Fatalf("dry-run command missing codex exec:\n%s", got)
	}
}

func TestJudgeAgentDefault_ExplicitAgentWins(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")

	var stdout, stderr bytes.Buffer
	f := &judgeFlags{
		Base:          "HEAD",
		Head:          "HEAD",
		Agent:         "claude",
		AgentExplicit: true,
		DryRun:        true,
		IssuesDir:     "workshop/issues",
		HistoryDir:    "workshop/history",
	}
	if err := runJudge(&stdout, &stderr, "dry", f); err != nil {
		t.Fatalf("runJudge: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "claude -p") {
		t.Fatalf("dry-run command missing claude -p:\n%s", got)
	}
}

func TestJudgeAgentDefault_AgentCmdWins(t *testing.T) {
	t.Setenv("AGENT_CMD", "gemini")
	t.Setenv("PAIR_AGENT", "codex")

	var stdout, stderr bytes.Buffer
	f := &judgeFlags{
		Base:          "HEAD",
		Head:          "HEAD",
		DryRun:        true,
		IssuesDir:     "workshop/issues",
		HistoryDir:    "workshop/history",
		AgentExplicit: false,
	}
	if err := runJudge(&stdout, &stderr, "dry", f); err != nil {
		t.Fatalf("runJudge: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "gemini -p") {
		t.Fatalf("dry-run command missing gemini -p:\n%s", got)
	}
}

func TestJudgeMilestoneReview_ExplicitRefsRenderPinnedManifest(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := "custom/plans"
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(plansDir, "000069-x-plan.md")
	if err := os.WriteFile(planPath, []byte("# plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	base := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD^"))
	head := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))

	var stdout, stderr bytes.Buffer
	f := &judgeFlags{
		Base: base, Head: "HEAD", DryRun: true,
		IssuesDir: issuesDir, HistoryDir: "custom/history", PlansDir: plansDir,
		Issue: 69, Milestone: "M1",
	}
	if err := runJudge(&stdout, &stderr, "milestone-review", f); err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	for _, want := range []string{
		"mode: committed range", "base: " + base, "head: " + head,
		"issue file: " + filepath.Join(issuesDir, "000069-x.md"),
		"plan file: " + planPath,
		"':!custom/history/'",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("manual milestone prompt missing %q:\n%s", want, got)
		}
	}
}

func TestJudgeMilestoneReview_OmittedHeadKeepsIssueOptional(t *testing.T) {
	_ = closeRepo(t, 69)
	base := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD^"))
	head := strings.TrimSpace(captureGit(t, "rev-parse", "HEAD"))

	var stdout, stderr bytes.Buffer
	f := &judgeFlags{
		Base: base, DryRun: true,
		IssuesDir: "workshop/issues", HistoryDir: "workshop/history", PlansDir: "workshop/plans",
	}
	if err := runJudge(&stdout, &stderr, "milestone-review", f); err != nil {
		t.Fatal(err)
	}
	got := stdout.String()
	for _, want := range []string{"mode: working tree", "ambient HEAD: " + head, "issue:      <unspecified>"} {
		if !strings.Contains(got, want) {
			t.Errorf("issue-less working-tree prompt missing %q:\n%s", want, got)
		}
	}
}

func TestJudgeCmd_RegistersPlansDirForManualMilestoneReview(t *testing.T) {
	if NewJudgeCmd().Flags().Lookup("plans-dir") == nil {
		t.Fatal("judge command does not register --plans-dir")
	}
}
