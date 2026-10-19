package agentid

import (
	"path/filepath"

	"github.com/xianxu/ariadne/pkg/agentenv"
)

// The agent CLIs sdlc knows, its environment slice and its detection live in
// pkg/agentenv (#441, so doc-review shares them); these re-export them.
const (
	Claude = agentenv.Claude
	Codex  = agentenv.Codex
	Gemini = agentenv.Gemini
)

// Env is agentenv.Env; judge.AgentDefaultEnv is this type.
type Env = agentenv.Env

// EnvFrom reads the agent signals through getenv (os.Getenv in production).
func EnvFrom(getenv func(string) string) Env { return agentenv.EnvFrom(getenv) }

// Known reports whether s names an agent CLI sdlc supports.
func Known(s string) bool { return agentenv.Known(s) }

// ClaudeProjects is Claude Code's per-cwd transcript store under home.
func ClaudeProjects(home string) string { return filepath.Join(home, ".claude", "projects") }

// CodexSessions is Codex's date-sharded transcript store under home.
func CodexSessions(home string) string { return filepath.Join(home, ".codex", "sessions") }
