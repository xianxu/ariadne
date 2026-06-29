package judge

import "os"

// AgentDefaultEnv is the explicit environment slice used to derive a default
// judge agent. Avoid wildcard scans: only process/session signals that identify
// the active agent stack belong here.
type AgentDefaultEnv struct {
	AgentCmd      string
	PairAgent     string
	CodexCI       string
	CodexThreadID string
	ClaudeCode    string
}

// CurrentAgentDefaultEnv reads the process environment at the IO boundary.
func CurrentAgentDefaultEnv() AgentDefaultEnv {
	claudeCode := os.Getenv("CLAUDECODE")
	if claudeCode == "" {
		claudeCode = os.Getenv("CLAUDE_CODE_ENTRYPOINT")
	}
	return AgentDefaultEnv{
		AgentCmd:      os.Getenv("AGENT_CMD"),
		PairAgent:     os.Getenv("PAIR_AGENT"),
		CodexCI:       os.Getenv("CODEX_CI"),
		CodexThreadID: os.Getenv("CODEX_THREAD_ID"),
		ClaudeCode:    claudeCode,
	}
}

// ResolveAgentCLI selects the judge agent. Explicit --agent and AGENT_CMD are
// operator/script overrides, so unknown values are returned as-is for BuildArgs
// to reject instead of being silently hidden by auto-detection.
func ResolveAgentCLI(explicit string, explicitSet bool, env AgentDefaultEnv) AgentCLI {
	if explicitSet {
		return AgentCLI(explicit)
	}
	if env.AgentCmd != "" {
		return AgentCLI(env.AgentCmd)
	}
	if agent, ok := knownAgent(env.PairAgent); ok {
		return agent
	}
	if env.CodexCI != "" || env.CodexThreadID != "" {
		return AgentCodex
	}
	if env.ClaudeCode != "" {
		return AgentClaude
	}
	return AgentClaude
}

func knownAgent(s string) (AgentCLI, bool) {
	switch AgentCLI(s) {
	case AgentClaude:
		return AgentClaude, true
	case AgentCodex:
		return AgentCodex, true
	case AgentGemini:
		return AgentGemini, true
	default:
		return "", false
	}
}
