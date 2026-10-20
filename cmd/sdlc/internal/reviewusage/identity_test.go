package reviewusage

import (
	"strings"
	"testing"
)

const testSession = "11111111-1111-4111-8111-111111111111"

func TestIdentityCurrentCodexStartup(t *testing.T) {
	metadata := "--------\nworkdir: /private/tmp/repo\nmodel: gpt-6.1-sol\nprovider: openai\napproval: never\nsandbox: read-only\nreasoning effort: medium\nreasoning summaries: none\nsession id: " + testSession + "\n--------\nuser\nprompt\n"
	for _, prefix := range []string{"", "Reading additional input from stdin...\n"} {
		raw := prefix + "OpenAI Codex v0.162.1\n" + metadata
		if got, err := CaptureIdentity("codex", nil, []byte(raw)); err != nil || got != testSession {
			t.Fatalf("current startup identity %q: %v", got, err)
		}
	}
	for _, prefix := range []string{"warning\n", "user\n", "Reading additional input from stdin...\nReading additional input from stdin...\n"} {
		if _, err := CaptureIdentity("codex", nil, []byte(prefix+"OpenAI Codex v0.162.1\n"+metadata)); err == nil {
			t.Fatal("arbitrary prefix authenticated startup identity")
		}
	}
}

func TestIdentityStrictProtocol(t *testing.T) {
	init := `{"type":"system","subtype":"init","session_id":"` + testSession + `"}` + "\n"
	result := `{"type":"result","subtype":"success","session_id":"` + testSession + `"}` + "\n"
	for _, raw := range []string{init, result, init + result} {
		got, err := CaptureIdentity("claude", []byte(raw), nil)
		if err != nil || got != testSession {
			t.Fatalf("claude identity %q %v", got, err)
		}
	}
	if _, err := CaptureIdentity("claude", []byte(init+strings.Replace(result, testSession, "22222222-2222-4222-8222-222222222222", 1)), nil); err == nil {
		t.Fatal("conflicting identity admitted")
	}
	// #489: a duplicate key, or one spelled with another case, gives one line two readings.
	for _, dup := range []string{`"session_id":"x","session_id"`, `"session_id":"x","SESSION_ID"`} {
		if _, err := CaptureIdentity("claude", []byte(strings.Replace(init, `"session_id"`, dup, 1)), nil); err == nil {
			t.Fatalf("duplicate key admitted: %s", dup)
		}
	}
	if _, err := CaptureIdentity("claude", []byte(`{"type":"assistant","session_id":"`+testSession+`"}`), nil); err == nil {
		t.Fatal("assistant content impersonated protocol")
	}
	preamble := "OpenAI Codex (v0.1.0)\n--------\nworkdir: /repo\nmodel: gpt-6\nprovider: openai\napproval: never\nsandbox: read-only\nreasoning effort: high\nreasoning summaries: auto\nsession id: " + testSession + "\n--------\nuser\nprompt\n"
	got, err := CaptureIdentity("codex", nil, []byte(preamble))
	if err != nil || got != testSession {
		t.Fatalf("codex identity %q %v", got, err)
	}
	if _, err := CaptureIdentity("codex", []byte(preamble), nil); err == nil {
		t.Fatal("model stdout impersonated startup identity")
	}
	for _, raw := range []string{"user\n" + preamble, "session id: " + testSession + "\n", strings.Replace(preamble, "session id: "+testSession, "session id: not-a-uuid", 1), strings.Repeat("x", IdentityLimit) + preamble, preamble[:strings.Index(preamble, "session id:")] + "--------\nuser\nsession id: " + testSession + "\n"} {
		if _, err := CaptureIdentity("codex", nil, []byte(raw)); err == nil {
			t.Fatal("non-preamble identity admitted")
		}
	}
}
