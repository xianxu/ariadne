package agentid

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Record shapes as Claude Code and Codex write them (trimmed to the fields read).
const claudeTail = `ature":"cut mid-line by the tail read"}
{"type":"assistant","message":{"model":"claude-sonnet-5-5"},"effort":"low"}
{"type":"user","message":{"content":"x"}}
{"type":"assistant","message":{"model":"claude-opus-5-5"},"effort":"high","perTurnEffort":"high"}
{"type":"system","subtype":"turn_duration"}
`

const codexTail = `{"type":"turn_context","payload":{"model":"gpt-6","effort":"low"}}
{"type":"response_item","payload":{"type":"message"}}
{"type":"turn_context","payload":{"cwd":"/w","model":"gpt-6-astra","effort":"medium","collaboration_mode":{"settings":{"reasoning_effort":"medium"}}}}
{"type":"event_msg","payload":{"type":"token_count"}}
`

func TestTranscriptTails(t *testing.T) {
	if m, e := claudeFromTail([]byte(claudeTail)); m != "claude-opus-5-5" || e != "high" {
		t.Errorf("claude tail = %q %q", m, e)
	}
	if m, e := codexFromTail([]byte(codexTail)); m != "gpt-6-astra" || e != "medium" {
		t.Errorf("codex tail = %q %q", m, e)
	}
	if m, e := claudeFromTail([]byte("not json\n{\"type\":\"user\"}\n")); m != "" || e != "" {
		t.Errorf("unknown records yielded %q %q", m, e)
	}
}

func TestDetect(t *testing.T) {
	home := t.TempDir()
	write := func(rel, body string) {
		p := filepath.Join(home, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".claude/projects/-w-repo/sid-1.jsonl", claudeTail)
	write(".codex/sessions/2026/10/14/rollout-2026-10-14T01-02-03-tid-9.jsonl", codexTail)
	for _, tc := range []struct {
		name string
		env  map[string]string
		want string
	}{
		{"plain shell", nil, "sdlc"},
		{"claude session", map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_SESSION_ID": "sid-1"}, "claude/claude-opus-5-5/high"},
		{"claude, transcript missing", map[string]string{"CLAUDECODE": "1", "CLAUDE_CODE_SESSION_ID": "gone", "CLAUDE_EFFORT": "max"}, "claude/-/max"},
		{"codex session", map[string]string{"CODEX_THREAD_ID": "tid-9"}, "codex/gpt-6-astra/medium"},
		{"codex, transcript missing", map[string]string{"CODEX_THREAD_ID": "nope"}, "codex/-/-"},
		{"override wins", map[string]string{"CLAUDECODE": "1", OverrideEnv: "codex/m/e"}, "codex/m/e"},
		{"malformed override ignored", map[string]string{OverrideEnv: "junk"}, "sdlc"},
	} {
		d := Detector{Home: home, Getenv: func(k string) string { return tc.env[k] }}
		if got := d.Detect().String(); got != tc.want {
			t.Errorf("%s: Detect = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestTailIsBounded(t *testing.T) {
	home := t.TempDir()
	p := filepath.Join(home, "big.jsonl")
	body := strings.Repeat("{\"type\":\"user\"}\n", tailBytes/8) + `{"type":"assistant","message":{"model":"m"},"effort":"e"}` + "\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Detector{Home: home}.tail(p)
	if len(got) != tailBytes {
		t.Fatalf("tail read %d bytes of %d, want %d", len(got), len(body), tailBytes)
	}
	if m, _ := claudeFromTail(got); m != "m" {
		t.Fatalf("the newest record was not in the tail")
	}
}
