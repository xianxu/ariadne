package agentid

import "testing"

func TestFromTranscriptExplicitSource(t *testing.T) {
	for _, tc := range []struct{ agent, raw, model, effort string }{
		{"codex", `{"type":"turn_context","payload":{"model":"model-x","effort":"high"}}`, "model-x", "high"},
		{"claude", `{"type":"assistant","effort":"medium","message":{"model":"model-y"}}`, "model-y", "medium"},
		{"claude", `{"type":"assistant","message":{"model":"model-y"}}`, "model-y", ""},
		{"unknown", `{}`, "", ""},
	} {
		got := FromTranscript(tc.agent, []byte(tc.raw))
		if got.Model != tc.model || got.Effort != tc.effort {
			t.Fatalf("%s: %+v", tc.agent, got)
		}
	}
}
