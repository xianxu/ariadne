package usage

import (
	"strings"
	"testing"
	"time"
)

func TestCodexExecutingIdentityOwnsTurnModels(t *testing.T) {
	context := `{"type":"turn_context","payload":{"turn_id":"turn","model":"foreign"}}` + "\n"
	header := `{"type":"session_meta","payload":{"id":"child"}}` + "\n"
	response := `{"type":"token_usage_record","timestamp":"2026-10-15T10:00:00Z","payload":{"response_id":"response","thread_id":"child","turn_id":"turn","usage":{"input_tokens":10,"output_tokens":2}}}` + "\n"
	for _, tc := range []struct {
		name, text, model string
		complete          bool
	}{
		{"foreign context before late header", context + header + response, "", true},
		{"owned context after early header", header + context + response, "foreign", true},
		{"owned context replaces ignored foreign prefix", context + header + strings.ReplaceAll(context, "foreign", "local") + response, "local", true},
		{"absent identity", context + response, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, c := decode(t, "codex", tc.text)
			if c.SpendComplete() != tc.complete {
				t.Fatalf("coverage=%+v", c)
			}
			if !tc.complete {
				if len(r) != 0 {
					t.Fatalf("unidentified response admitted: %+v", r)
				}
				return
			}
			if len(r) != 1 || r[0].Model != tc.model {
				t.Fatalf("foreign model context crossed identity boundary: %+v coverage=%+v", r, c)
			}
			at := stamp("2026-10-15T10:00:00Z")
			window := Window{Since: at, Until: at.Add(time.Hour)}
			report, err := AggregateUsage(r, window, window, nil)
			if err != nil || (report.Coverage.Counts["unknown_model"] > 0) != (tc.model == "") || !report.Coverage.SpendComplete() {
				t.Fatalf("model coverage=%+v err=%v", report.Coverage, err)
			}
		})
	}
}
