package processmanual

import (
	"encoding/json"
	"strings"
	"testing"
)

// Fixtures mirror real ~/.codex/sessions rollout lines (shapes verified against
// the live corpus; atlas/workflow/introspect.md → "Codex transcript format" is
// the contract these derive from — NOT re-discovered format knowledge).

const (
	codexRootMeta = `{"timestamp":"2026-07-01T10:00:00.000Z","type":"session_meta","payload":{"id":"019e-root","timestamp":"2026-07-01T10:00:00.000Z","cwd":"/Users/x/workspace/pair","originator":"codex-tui","cli_version":"0.134.0","source":"cli","thread_source":"user"}}`
	// fork-replay: forked_from_id set on the FIRST meta; the file replays the
	// parent's transcript and carries the parent's meta later — SKIP, else the
	// parent's events double-count (the spec's 66%-inflation trap).
	codexForkMeta = `{"timestamp":"2026-07-01T11:00:00.000Z","type":"session_meta","payload":{"session_id":"019e-root","id":"019e-fork","forked_from_id":"019e-root","parent_thread_id":"019e-root","cwd":"/Users/x/workspace/pair","originator":"codex-tui"}}`
	// sub-agent thread: parent_thread_id/agent_nickname WITHOUT forked_from_id —
	// own content, no replay → processed like any session.
	codexSubMeta = `{"timestamp":"2026-07-01T12:00:00.000Z","type":"session_meta","payload":{"id":"019e-sub","parent_thread_id":"019e-root","cwd":"/Users/x/workspace/pair","thread_source":"subagent","agent_nickname":"Helmholtz"}}`
)

// codexCall builds a response_item/function_call line the way codex emits it:
// `arguments` is a JSON-ENCODED STRING (double encoding) carrying the command
// under `cmd` (a plain string — the shape agent_codex.py reads).
func codexCall(t *testing.T, callID, cmd string) string {
	t.Helper()
	args, _ := json.Marshal(map[string]any{"cmd": cmd, "workdir": "/Users/x/workspace/pair", "yield_time_ms": 1000, "max_output_tokens": 16000})
	b, _ := json.Marshal(map[string]any{
		"timestamp": "2026-07-01T10:01:00.000Z", "type": "response_item",
		"payload": map[string]any{"type": "function_call", "name": "exec_command", "arguments": string(args), "call_id": callID},
	})
	return string(b)
}

// codexOutput builds the linked function_call_output with the real exec_command
// wrapper (Chunk ID / Wall time / Process exited / Output:) around the payload.
func codexOutput(t *testing.T, callID, exitCode, body string) string {
	t.Helper()
	out := "Chunk ID: 04b0cc\nWall time: 0.1 seconds\nProcess exited with code " + exitCode + "\nOriginal token count: 100\nOutput:\n" + body
	b, _ := json.Marshal(map[string]any{
		"timestamp": "2026-07-01T10:01:01.000Z", "type": "response_item",
		"payload": map[string]any{"type": "function_call_output", "call_id": callID, "output": out},
	})
	return string(b)
}

func TestCodexMetaKind(t *testing.T) {
	cases := []struct {
		name string
		data string
		want codexMetaKind
		cwd  string
	}{
		{"plain root", codexRootMeta, codexRoot, "/Users/x/workspace/pair"},
		{"fork-replay (forked_from_id on FIRST meta)", codexForkMeta + "\n" + codexRootMeta, codexForkReplay, "/Users/x/workspace/pair"},
		{"sub-agent thread (no forked_from_id)", codexSubMeta, codexSubAgent, "/Users/x/workspace/pair"},
	}
	for _, c := range cases {
		kind, cwd, ok := codexMeta([]byte(c.data))
		if !ok || kind != c.want || cwd != c.cwd {
			t.Errorf("%s: got (kind=%d cwd=%q ok=%v), want (kind=%d cwd=%q ok=true)", c.name, kind, cwd, ok, c.want, c.cwd)
		}
	}
	if _, _, ok := codexMeta([]byte(`{"timestamp":"t","type":"event_msg","payload":{}}`)); ok {
		t.Error("file without session_meta must yield ok=false")
	}
}

func TestParseCodexInvocations(t *testing.T) {
	ack := "  [1;33m[!][0m --no-atlas (or --force): skipping atlas/ change check — rationale in --verified"
	root := strings.Join([]string{
		codexRootMeta,
		codexCall(t, "call_1", "sdlc close --issue 42 --no-atlas --actual 1.0 --verified 'x'"),
		codexOutput(t, "call_1", "0", "closing…\n"+ack+"\ndone."),
		codexCall(t, "call_2", "rg --files"), // non-sdlc → excluded
		codexOutput(t, "call_2", "0", "a.md"),
		`{"timestamp":"2026-07-01T10:02:00.000Z","type":"event_msg","payload":{"type":"user_message","message":"hi"}}`,
	}, "\n")
	verbs := map[string]bool{"close": true}

	invs := parseCodexInvocations([]byte(root), verbs)
	if len(invs) != 1 {
		t.Fatalf("root: want 1 sdlc invocation, got %+v", invs)
	}
	inv := invs[0]
	if inv.Verb != "close" || inv.IssueID != "42" || inv.Agent != "codex" {
		t.Errorf("inv = {verb:%q issue:%q agent:%q}, want {close 42 codex}", inv.Verb, inv.IssueID, inv.Agent)
	}
	if !strings.Contains(inv.Output, "--no-atlas (or --force): skipping") {
		t.Errorf("linked output missing the ACK: %q", inv.Output)
	}
	// end-to-end: the SAME classifier reads codex output (ANSI survives the wrapper)
	if evs := invocationGateEvents(inv); len(evs) != 1 || evs[0].Kind != GateBypass || evs[0].Gate != "no-atlas" {
		t.Errorf("want one no-atlas bypass from the codex output, got %+v", evs)
	}

	// fork-replay: same events, but the FIRST meta carries forked_from_id → SKIP ALL
	fork := codexForkMeta + "\n" + strings.Join(strings.Split(root, "\n")[1:], "\n") + "\n" + codexRootMeta
	if got := parseCodexInvocations([]byte(fork), verbs); got != nil {
		t.Errorf("fork-replay rollout must be skipped entirely, got %+v", got)
	}

	// sub-agent: own content, processed
	sub := codexSubMeta + "\n" + strings.Join(strings.Split(root, "\n")[1:], "\n")
	if got := parseCodexInvocations([]byte(sub), verbs); len(got) != 1 {
		t.Errorf("sub-agent rollout must be processed, got %+v", got)
	}
}

// A refused sdlc invocation under codex carries a non-zero wrapper exit — Failed
// derives from `Process exited with code N` (N≠0). This is NOT the atlas spec's
// taste-friction `is_error` gate (which additionally requires a friction hint);
// Failed answers "did the command complete?", for the firing-order ladder.
func TestParseCodexInvocationsFailed(t *testing.T) {
	refusal := "[1;31mError: pair#75 is already status: done — pass --no-reclose-guard (or --force) to re-close intentionally[0m"
	data := strings.Join([]string{
		codexRootMeta,
		codexCall(t, "call_1", "sdlc close --issue 75"),
		codexOutput(t, "call_1", "71", refusal),
	}, "\n")
	invs := parseCodexInvocations([]byte(data), map[string]bool{"close": true})
	if len(invs) != 1 || !invs[0].Failed {
		t.Fatalf("want one Failed invocation (exit 71), got %+v", invs)
	}
	if evs := invocationGateEvents(invs[0]); len(evs) != 1 || evs[0].Kind != GateRefusal || evs[0].Gate != "no-reclose-guard" {
		t.Errorf("want the reclose refusal classified, got %+v", evs)
	}
}
