// Package agentenv names the coding agent a session runs as and reads the
// `Agent:` commit trailers (#367), for every binary that needs them (#441):
// sdlc's internal/agentid re-exports it, and doc-review imports it directly.
// Pure apart from EnvFrom's injected getenv.
package agentenv

import "strings"

// The agent CLIs sdlc knows.
const (
	Claude = "claude"
	Codex  = "codex"
	Gemini = "gemini"
)

// Env is the explicit environment slice that names the active agent stack.
// Avoid wildcard scans: only process/session signals belong here.
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

// TrailerKey is the git trailer every agent-authored or sdlc-made commit carries.
const TrailerKey = "Agent"

// Commit is one commit as a window report sees it.
type Commit struct {
	SHA, Subject string
	Agents       []string // values of its Agent trailers
	Merge        bool
}

// LogFormat is the `git log --format` whose output ParseLog reads: hash,
// parent hashes, subject and the Agent trailer values, one record per commit.
const LogFormat = "%H%x1f%P%x1f%s%x1f%(trailers:key=" + TrailerKey + ",valueonly,separator=%x1e)%x1d"

// ParseLog reads `git log --format=LogFormat` output into commits.
func ParseLog(out string) []Commit {
	var commits []Commit
	for _, rec := range strings.Split(out, "\x1d") {
		f := strings.Split(strings.TrimLeft(rec, "\n"), "\x1f")
		if len(f) != 4 {
			continue
		}
		c := Commit{SHA: f[0], Subject: f[2], Merge: len(strings.Fields(f[1])) > 1}
		for _, v := range strings.Split(f[3], "\x1e") {
			if v = strings.TrimSpace(v); v != "" {
				c.Agents = append(c.Agents, v)
			}
		}
		commits = append(commits, c)
	}
	return commits
}
