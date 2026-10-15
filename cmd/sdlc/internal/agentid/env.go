package agentid

import "path/filepath"

// The agent CLIs sdlc knows. judge's AgentCLI constants are these names.
const (
	Claude = "claude"
	Codex  = "codex"
	Gemini = "gemini"
)

// Env is the explicit environment slice that names the active agent stack.
// Avoid wildcard scans: only process/session signals belong here. It lives in
// this leaf package so the commit primitives (gitx) can stamp without
// importing the judge; judge.AgentDefaultEnv is this type.
type Env struct {
	AgentCmd      string
	PairAgent     string
	CodexCI       string
	CodexThreadID string
	ClaudeCode    string
}

// EnvFrom reads the agent signals through getenv (os.Getenv in production).
func EnvFrom(getenv func(string) string) Env {
	claudeCode := getenv("CLAUDECODE")
	if claudeCode == "" {
		claudeCode = getenv("CLAUDE_CODE_ENTRYPOINT")
	}
	return Env{
		AgentCmd:      getenv("AGENT_CMD"),
		PairAgent:     getenv("PAIR_AGENT"),
		CodexCI:       getenv("CODEX_CI"),
		CodexThreadID: getenv("CODEX_THREAD_ID"),
		ClaudeCode:    claudeCode,
	}
}

// Known reports whether s names an agent CLI sdlc supports.
func Known(s string) bool { return s == Claude || s == Codex || s == Gemini }

// Main is the agent driving this session: PAIR_AGENT when it names a known
// agent, Codex when its session signals are set, Claude otherwise.
func (e Env) Main() string {
	if Known(e.PairAgent) {
		return e.PairAgent
	}
	if e.CodexCI != "" || e.CodexThreadID != "" {
		return Codex
	}
	return Claude
}

// InSession reports whether any agent session signal is present. A plain
// shell (an operator typing sdlc commands) has none, and so no agent to name.
func (e Env) InSession() bool {
	return e.CodexCI != "" || e.CodexThreadID != "" || e.ClaudeCode != ""
}

// ClaudeProjects is Claude Code's per-cwd transcript store under home.
func ClaudeProjects(home string) string { return filepath.Join(home, ".claude", "projects") }

// CodexSessions is Codex's date-sharded transcript store under home.
func CodexSessions(home string) string { return filepath.Join(home, ".codex", "sessions") }
