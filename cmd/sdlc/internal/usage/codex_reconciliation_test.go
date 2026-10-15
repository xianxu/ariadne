package usage

import (
	"fmt"
	"testing"
)

func TestCodexModernReconcilesEveryCumulativeSource(t *testing.T) {
	modern := func(id string, input, output, threadInput, threadOutput int64) string {
		return fmt.Sprintf("{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-14T10:00:00Z\",\"payload\":{\"response_id\":%q,\"thread_id\":\"thread\",\"turn_id\":\"t\",\"usage\":{\"input_tokens\":%d,\"output_tokens\":%d},\"thread_token_usage\":{\"input_tokens\":%d,\"output_tokens\":%d}}}\n", id, input, output, threadInput, threadOutput)
	}
	legacy := func(input, output int64) string {
		return fmt.Sprintf("{\"type\":\"event_msg\",\"timestamp\":\"2026-10-14T10:01:00Z\",\"payload\":{\"type\":\"token_count\",\"info\":{\"total_token_usage\":{\"input_tokens\":%d,\"output_tokens\":%d}}}}\n", input, output)
	}
	base := modern("r", 100, 20, 100, 20)
	for _, tc := range []struct {
		name, text    string
		input, output int64
		gap, reset    bool
	}{
		{"later legacy exceeds modern", base + legacy(500, 100), 100, 20, true, false},
		{"copied history cannot hide ambiguous legacy spend", "{\"type\":\"session_meta\",\"payload\":{\"id\":\"thread\"}}\n" + "{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-14T10:00:00Z\",\"payload\":{\"response_id\":\"parent-response\",\"thread_id\":\"parent\",\"turn_id\":\"t\",\"usage\":{\"input_tokens\":900,\"output_tokens\":200}}}\n" + base + legacy(500, 100), 100, 20, true, false},
		{"earlier legacy exceeds modern", legacy(500, 100) + base, 100, 20, true, false},
		{"reset cannot erase legacy high water", base + legacy(500, 100) + legacy(0, 0), 100, 20, true, true},
		{"legacy high water without thread counter", legacy(500, 100) + legacy(0, 0) + "{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-14T10:00:00Z\",\"payload\":{\"response_id\":\"r\",\"thread_id\":\"thread\",\"turn_id\":\"t\",\"usage\":{\"input_tokens\":100,\"output_tokens\":20}}}\n", 100, 20, true, true},
		{"thread input high water cannot be overwritten by output", modern("r", 100, 20, 500, 10) + modern("s", 0, 0, 100, 20), 100, 20, true, false},
		{"thread output high water cannot be overwritten by input", modern("r", 100, 20, 10, 100) + modern("s", 0, 0, 100, 20), 100, 20, true, false},
		{"reconciled snapshots do not add spend", base + legacy(0, 0) + legacy(100, 20) + legacy(100, 20), 100, 20, false, false},
		{"reset with no excess remains uncertain", base + legacy(100, 20) + legacy(0, 0), 100, 20, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			responses, coverage := decode(t, "codex", tc.text)
			var input, output int64
			for _, response := range responses {
				input += response.Tokens.Input
				output += response.Tokens.Output
			}
			if input != tc.input || output != tc.output {
				t.Fatalf("cumulative evidence changed response spend: %d/%d", input, output)
			}
			if (coverage.Counts["uncovered_segments"] > 0) != tc.gap || (coverage.Counts["reset_gaps"] > 0) != tc.reset || coverage.SpendComplete() != (!tc.gap && !tc.reset) {
				t.Fatalf("coverage lost cumulative evidence: %+v", coverage)
			}
		})
	}
}
