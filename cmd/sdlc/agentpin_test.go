package main

import (
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// pinReviewAgent declares the agent CLI a fixture's stubbed review output or
// argv assertion is written for (#349). Without it the test inherits the
// session's agent (PAIR_AGENT=codex in a Codex slot) and production
// resolution correctly picks another CLI. AGENT_CMD is the operator override
// that outranks PAIR_AGENT and the reviewer policy, so it pins the agent
// whatever the ambient AGENT_CMD or PAIR_AGENT holds; only an explicit --agent
// is above it. Production precedence keeps its own tests (agent_resolver_test.go).
func pinReviewAgent(t *testing.T, agent judge.AgentCLI) {
	t.Helper()
	t.Setenv("AGENT_CMD", string(agent))
}
