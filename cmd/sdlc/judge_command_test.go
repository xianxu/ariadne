package main

import (
	"bytes"
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
