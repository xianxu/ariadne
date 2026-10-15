package usage

import (
	"fmt"
	"testing"
)

func TestCodexModernReconcilesEveryCumulativeSource(t *testing.T) {
	header := `{"type":"session_meta","payload":{"id":"thread"}}` + "\n"
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
		{"later legacy exceeds modern", header + base + legacy(500, 100), 100, 20, true, false},
		{"copied history cannot hide ambiguous legacy spend", header + "{\"type\":\"session_meta\",\"payload\":{\"id\":\"thread\"}}\n" + "{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-14T10:00:00Z\",\"payload\":{\"response_id\":\"parent-response\",\"thread_id\":\"parent\",\"turn_id\":\"t\",\"usage\":{\"input_tokens\":900,\"output_tokens\":200}}}\n" + base + legacy(500, 100), 100, 20, true, false},
		{"earlier legacy exceeds modern", header + legacy(500, 100) + base, 100, 20, true, false},
		{"reset cannot erase legacy high water", header + base + legacy(500, 100) + legacy(0, 0), 100, 20, true, true},
		{"legacy high water without thread counter", header + legacy(500, 100) + legacy(0, 0) + "{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-14T10:00:00Z\",\"payload\":{\"response_id\":\"r\",\"thread_id\":\"thread\",\"turn_id\":\"t\",\"usage\":{\"input_tokens\":100,\"output_tokens\":20}}}\n", 100, 20, true, true},
		{"thread input high water cannot be overwritten by output", header + modern("r", 100, 20, 500, 10) + modern("s", 0, 0, 100, 20), 100, 20, true, false},
		{"thread output high water cannot be overwritten by input", header + modern("r", 100, 20, 10, 100) + modern("s", 0, 0, 100, 20), 100, 20, true, false},
		{"reconciled snapshots do not add spend", header + base + legacy(0, 0) + legacy(100, 20) + legacy(100, 20), 100, 20, false, false},
		{"reset with no excess remains uncertain", header + base + legacy(100, 20) + legacy(0, 0), 100, 20, false, true},
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

func TestCodexTurnCumulativeCoverage(t *testing.T) {
	header := `{"type":"session_meta","payload":{"id":"thread"}}` + "\n"
	record := func(id, turn string, input, output int64, cumulative string) string {
		field := ""
		if cumulative != "" {
			field = `,"turn_token_usage":` + cumulative
		}
		return fmt.Sprintf("{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-14T10:00:00Z\",\"payload\":{\"response_id\":%q,\"thread_id\":\"thread\",\"turn_id\":%q,\"usage\":{\"input_tokens\":%d,\"output_tokens\":%d}%s}}\n", id, turn, input, output, field)
	}
	total := func(input, output int64) string {
		return fmt.Sprintf(`{"input_tokens":%d,"output_tokens":%d}`, input, output)
	}
	base := record("r", "t1", 100, 20, total(500, 100))
	matching := record("s", "t2", 1000, 200, total(1000, 200))
	for _, tc := range []struct {
		name, text    string
		input, output int64
		gap           bool
	}{
		{"missing spend with no thread counter", header + base, 100, 20, true},
		{"another turn cannot fill missing spend", header + base + matching, 1100, 220, true},
		{"duplicate cannot fill missing spend", header + base + base, 100, 20, true},
		{"turn input high water survives output increase", header + record("r", "t1", 100, 20, total(500, 10)) + record("s", "t1", 0, 0, total(100, 20)), 100, 20, true},
		{"turn output high water survives input increase", header + record("r", "t1", 100, 20, total(10, 100)) + record("s", "t1", 0, 0, total(100, 20)), 100, 20, true},
		{"invalid negative turn counter", header + record("r", "t1", 100, 20, total(-1, 20)), 100, 20, true},
		{"invalid turn counter shape", header + record("r", "t1", 100, 20, `{"input_tokens":"unknown","output_tokens":20}`), 100, 20, true},
		{"missing turn counter dimension", header + record("r", "t1", 100, 20, `{"input_tokens":100}`), 100, 20, true},
		{"matching multi response and multi turn totals", header + record("r", "t1", 100, 20, total(100, 20)) + record("q", "t1", 200, 40, total(300, 60)) + matching, 1300, 260, false},
		{"matching duplicate has no extra charge", header + matching + matching, 1000, 200, false},
		{"absent turn counter", header + record("r", "t1", 100, 20, ""), 100, 20, false},
		{"null turn counter", header + record("r", "t1", 100, 20, "null"), 100, 20, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			responses, coverage := decode(t, "codex", tc.text)
			responses, normalized, err := NormalizeResponses(responses)
			if err != nil {
				t.Fatal(err)
			}
			coverage.Merge(normalized)
			var input, output int64
			for _, response := range responses {
				input += response.Tokens.Input
				output += response.Tokens.Output
			}
			if input != tc.input || output != tc.output {
				t.Fatalf("cumulative evidence changed response spend: %d/%d", input, output)
			}
			if (coverage.Counts["uncovered_segments"] > 0) != tc.gap || coverage.SpendComplete() == tc.gap {
				t.Fatalf("turn coverage lost cumulative evidence: %+v", coverage)
			}
		})
	}
}
