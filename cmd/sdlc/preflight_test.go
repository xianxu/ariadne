package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

func TestPreflightAgentDefault_UsesPairAgent(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")
	seenName := stubPreflightJudgeName(t)

	opts := preflightOptions{
		Categories: []judge.Category{judge.DRY},
		Base:       "HEAD",
		Head:       "HEAD",
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	}
	if err := runPreflightJudges(opts); err != nil {
		t.Fatalf("runPreflightJudges: %v", err)
	}
	if *seenName != "codex" {
		t.Fatalf("preflight agent = %q, want codex", *seenName)
	}
}

func TestPreflightAgentDefault_ExplicitAgentWins(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")
	seenName := stubPreflightJudgeName(t)

	opts := preflightOptions{
		Categories: []judge.Category{judge.DRY},
		Base:       "HEAD",
		Head:       "HEAD",
		Agent:      judge.AgentClaude,
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
	}
	if err := runPreflightJudges(opts); err != nil {
		t.Fatalf("runPreflightJudges: %v", err)
	}
	if *seenName != "claude" {
		t.Fatalf("preflight agent = %q, want claude", *seenName)
	}
}

func stubPreflightJudgeName(t *testing.T) *string {
	t.Helper()
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	seenName := ""
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		seenName = name
		return []byte("VERDICT: CLEAN (confidence: high)\n"), nil
	}
	return &seenName
}
