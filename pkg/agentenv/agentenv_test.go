package agentenv

import (
	"strings"
	"testing"
)

func TestEnvMain(t *testing.T) {
	env := func(kv ...string) Env {
		m := map[string]string{}
		for i := 0; i+1 < len(kv); i += 2 {
			m[kv[i]] = kv[i+1]
		}
		return EnvFrom(func(k string) string { return m[k] })
	}
	for _, tc := range []struct {
		name string
		e    Env
		main string
		in   bool
	}{
		{"no signals: claude, not in a session", env(), Claude, false},
		{"PAIR_AGENT names a known agent", env("PAIR_AGENT", "gemini"), Gemini, false},
		{"an unknown PAIR_AGENT is ignored", env("PAIR_AGENT", "bard", "CODEX_CI", "1"), Codex, true},
		{"codex's thread id", env("CODEX_THREAD_ID", "t"), Codex, true},
		{"claude code's entrypoint counts", env("CLAUDE_CODE_ENTRYPOINT", "cli"), Claude, true},
	} {
		if got := tc.e.Main(); got != tc.main {
			t.Errorf("%s: Main() = %q, want %q", tc.name, got, tc.main)
		}
		if got := tc.e.InSession(); got != tc.in {
			t.Errorf("%s: InSession() = %v, want %v", tc.name, got, tc.in)
		}
	}
	if env("AGENT_CMD", "codex").AgentCmd != "codex" {
		t.Error("AGENT_CMD not read")
	}
}

func TestParseLog(t *testing.T) {
	rec := func(sha, parents, subject string, agents ...string) string {
		return sha + "\x1f" + parents + "\x1f" + subject + "\x1f" + strings.Join(agents, "\x1e") + "\x1d"
	}
	out := rec("a1", "p0", "one", "codex/gpt/med", " claude/opus/high ") + "\n" + rec("m1", "a1 b1", "merge") + "bad\x1fx\x1d"
	commits := ParseLog(out)
	if len(commits) != 2 {
		t.Fatalf("commits %+v", commits)
	}
	if c := commits[0]; c.SHA != "a1" || c.Subject != "one" || c.Merge || strings.Join(c.Agents, ",") != "codex/gpt/med,claude/opus/high" {
		t.Errorf("first commit %+v", c)
	}
	if c := commits[1]; !c.Merge || len(c.Agents) != 0 {
		t.Errorf("merge commit %+v", c)
	}
	if !strings.Contains(LogFormat, "trailers:key="+TrailerKey) {
		t.Errorf("LogFormat does not read the %s trailer: %s", TrailerKey, LogFormat)
	}
}
