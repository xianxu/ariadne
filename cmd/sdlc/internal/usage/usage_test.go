package usage

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

func stamp(s string) time.Time { t, _ := time.Parse(time.RFC3339, s); return t }
func decode(t *testing.T, agent, text string) ([]Response, Coverage) {
	t.Helper()
	r, c, e := Decode(context.Background(), agent, "source", strings.NewReader(text))
	if e != nil {
		t.Fatal(e)
	}
	return r, c
}
func p(n int64) *int64 { return &n }

func TestCodexModernCompactionAndModel(t *testing.T) {
	r, c := decode(t, "codex", `{"type":"session_meta","payload":{"id":"thread"}}
{"type":"turn_context","payload":{"turn_id":"t","model":"model-a"}}
{"type":"token_usage_record","timestamp":"2026-10-14T10:00:00Z","payload":{"response_id":"r","thread_id":"thread","turn_id":"t","usage":{"input_tokens":100,"cached_input_tokens":80,"output_tokens":20,"reasoning_output_tokens":5,"total_tokens":120}}}
{"type":"compacted","timestamp":"2026-10-14T10:01:00Z","payload":{"latest_token_usage_record":{"response_id":"r","thread_id":"thread","turn_id":"t","usage":{"input_tokens":100,"cached_input_tokens":80,"output_tokens":20,"reasoning_output_tokens":5,"total_tokens":120}}}}
{"type":"event_msg","timestamp":"2026-10-14T10:02:00Z","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":500,"output_tokens":100},"last_token_usage":{"input_tokens":500,"output_tokens":100}}}}
`)
	r, n, e := NormalizeResponses(r)
	if e != nil {
		t.Fatal(e)
	}
	c.Merge(n)
	if len(r) != 1 || r[0].Model != "model-a" || r[0].Tokens.Input != 100 || r[0].Tokens.Output != 20 || *r[0].Tokens.CacheRead != 80 || r[0].Tokens.CacheWrite != nil {
		t.Fatalf("responses=%+v coverage=%+v", r, c)
	}
	if c.Counts["duplicates"] != 1 {
		t.Fatalf("coverage=%+v", c)
	}
	if c.Counts["uncovered_segments"] != 1 {
		t.Fatalf("modern mixed-file counter gap disappeared: %+v", c)
	}
	total, e := r[0].Tokens.Total()
	if e != nil || total != 120 {
		t.Fatalf("total=%d %v", total, e)
	}
}

func TestCodexLegacyBaselineDuplicateReset(t *testing.T) {
	r, c := decode(t, "codex", `{"type":"session_meta","payload":{"id":"thread"}}
{"type":"turn_context","payload":{"model":"legacy"}}
{"type":"event_msg","timestamp":"2026-10-14T09:00:00Z","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":100,"output_tokens":10},"last_token_usage":{"input_tokens":100,"output_tokens":10}}}}
{"type":"event_msg","timestamp":"2026-10-14T10:00:00Z","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":140,"output_tokens":20},"last_token_usage":{"input_tokens":999,"output_tokens":999}}}}
{"type":"event_msg","timestamp":"2026-10-14T10:01:00Z","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":140,"output_tokens":20},"last_token_usage":{"input_tokens":999,"output_tokens":999}}}}
{"type":"event_msg","timestamp":"2026-10-14T10:02:00Z","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":5,"output_tokens":1}}}}
{"type":"event_msg","timestamp":"2026-10-14T10:03:00Z","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":15,"output_tokens":3}}}}
`)
	if len(r) != 2 || r[0].Tokens.Input != 40 || r[0].Tokens.Output != 10 || r[1].Tokens.Input != 10 || r[1].Tokens.Output != 2 {
		t.Fatalf("responses=%+v", r)
	}
	if c.Counts["baseline_gaps"] != 1 || c.Counts["reset_gaps"] != 1 || c.Complete() {
		t.Fatalf("coverage=%+v", c)
	}
}

func TestClaudeRepeatedBlocksRevisionAndCache(t *testing.T) {
	r, c := decode(t, "claude", `{"type":"assistant","timestamp":"2026-10-14T10:00:00Z","requestId":"req","message":{"id":"m","model":"claude-model","stop_reason":null,"usage":{"input_tokens":10,"output_tokens":1,"cache_read_input_tokens":80,"cache_creation_input_tokens":20,"cache_creation":{"ephemeral_5m_input_tokens":20}}}}
{"type":"assistant","timestamp":"2026-10-14T10:01:00Z","requestId":"req","message":{"id":"m","model":"claude-model","stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":80,"cache_creation_input_tokens":20}}}
{"type":"assistant","timestamp":"2026-10-14T10:01:00Z","requestId":"req","message":{"id":"m","model":"claude-model","stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":5,"cache_read_input_tokens":80,"cache_creation_input_tokens":20}}}
`)
	r, n, e := NormalizeResponses(r)
	if e != nil {
		t.Fatal(e)
	}
	c.Merge(n)
	if len(r) != 1 || r[0].Provisional || r[0].Tokens.Input != 110 || r[0].Tokens.Output != 5 || *r[0].Tokens.CacheWrite != 20 || r[0].Tokens.Reasoning != nil || !r[0].At.Equal(stamp("2026-10-14T10:01:00Z")) {
		t.Fatalf("responses=%+v", r)
	}
	if c.Counts["revisions"] != 1 || c.Counts["duplicates"] != 1 {
		t.Fatalf("coverage=%+v", c)
	}
}

func TestNormalizeConflictBeforeWindow(t *testing.T) {
	base := Response{ID: "m", SourceID: "s", Agent: "claude", Model: "a", RequestID: "req", At: stamp("2026-10-14T09:00:00Z"), Tokens: Counters{Input: 10, Output: 5}}
	for _, kind := range []string{"terminal", "model", "request", "source"} {
		t.Run(kind, func(t *testing.T) {
			next := base
			next.At = stamp("2026-10-14T10:00:00Z")
			switch kind {
			case "terminal":
				next.Tokens.Output = 6
			case "model":
				next.Model = "b"
			case "request":
				next.RequestID = "other"
			case "source":
				next.SourceID = "other"
			}
			got, e := AggregateUsage([]Response{base, next}, Window{stamp("2026-10-14T10:00:00Z"), stamp("2026-10-14T11:00:00Z")}, Window{stamp("2026-10-14T10:00:00Z"), stamp("2026-10-14T11:00:00Z")}, nil)
			if e != nil {
				t.Fatal(e)
			}
			if got.Totals.Input != 0 || got.Coverage.Counts["conflicts"] != 1 {
				t.Fatalf("report=%+v", got)
			}
		})
	}
}

func TestAggregateHoursBurnUnknownAndOverflow(t *testing.T) {
	since := stamp("2026-10-14T10:30:00Z")
	until := stamp("2026-10-14T12:00:00Z")
	rr := []Response{{ID: "a", SourceID: "s", Agent: "codex", Model: "a", At: since, Tokens: Counters{Input: 100, Output: 20, CacheRead: p(80)}}, {ID: "b", SourceID: "s", Agent: "codex", Model: "b", At: stamp("2026-10-14T11:00:00Z"), Tokens: Counters{Input: 30, Output: 0}}, {ID: "c", Agent: "codex", At: until, Tokens: Counters{Input: 999}}}
	got, e := AggregateUsage(rr, Window{since, until}, Window{since, until}, nil)
	if e != nil {
		t.Fatal(e)
	}
	if got.Totals.Input != 130 || got.Totals.Output != 20 || got.BurnTokensPerHour != 100 || len(got.Groups) != 2 || len(got.Hourly) != 2 || got.Totals.CacheRead != nil || got.Coverage.Counts["unknown_slot"] != 2 {
		t.Fatalf("report=%+v", got)
	}
	rr[0].Tokens.Input = math.MaxInt64
	if _, e = AggregateUsage(rr, Window{since, until}, Window{since, until}, nil); e == nil {
		t.Fatal("overflow admitted")
	}
}

func TestDecodeMalformedUnsupportedCancellation(t *testing.T) {
	_, c := decode(t, "codex", "{bad}\n{\"type\":\"token_usage_record\",\"timestamp\":\"bad\",\"payload\":{}}\n")
	if c.Counts["malformed_records"] != 1 || c.Counts["unknown_metadata"] == 0 || c.Counts["uncovered_segments"] == 0 || c.Complete() {
		t.Fatalf("coverage=%+v", c)
	}
	_, c = decode(t, "other", "{}\n")
	if c.Counts["unsupported_formats"] != 1 {
		t.Fatalf("coverage=%+v", c)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, e := Decode(ctx, "codex", "s", strings.NewReader("{}\n")); e == nil {
		t.Fatal("cancel ignored")
	}
}

func TestUsageSpendCompletenessIsNotSlotAttribution(t *testing.T) {
	var c Coverage
	c.Add("unknown_slot", "")
	c.Add("unknown_model", "")
	c.Add("attribution_conflicts", "")
	if c.Complete() || !c.SpendComplete() {
		t.Fatalf("attribution uncertainty lost known spending: %+v", c)
	}
	c.Add("baseline_gaps", "")
	if c.SpendComplete() {
		t.Fatal("missing spending baseline is complete")
	}
}

func TestClaudeProvisionalConflictingFinalAndOverflow(t *testing.T) {
	r, _ := decode(t, "claude", `{"type":"assistant","timestamp":"2026-10-14T10:00:00Z","message":{"id":"m","model":"a","stop_reason":null,"usage":{"input_tokens":10,"output_tokens":1}}}`)
	window := Window{stamp("2026-10-14T09:00:00Z"), stamp("2026-10-14T11:00:00Z")}
	report, err := AggregateUsage(r, window, window, nil)
	if err != nil || report.Totals.Input != 10 || report.Coverage.Counts["provisional"] != 1 || report.Coverage.SpendComplete() {
		t.Fatalf("partial response lost or presented complete: %+v %v", report, err)
	}
	final := r[0]
	final.Provisional = false
	final.Tokens.Output = 5
	conflict := final
	conflict.Tokens.Output = 6
	normal, coverage, err := NormalizeResponses([]Response{final, conflict, final})
	if err != nil || len(normal) != 0 || coverage.Counts["conflicts"] != 1 {
		t.Fatalf("conflicting final was revived: %+v %+v %v", normal, coverage, err)
	}
	_, coverage = decode(t, "codex", `{"type":"session_meta","payload":{"id":"thread"}}
{"type":"event_msg","timestamp":"2026-10-14T10:00:00Z","payload":{"type":"token_count","info":{"total_token_usage":{"total_tokens":100}}}}`)
	if coverage.Counts["malformed_records"] != 1 || coverage.SpendComplete() {
		t.Fatalf("unsupported input/output invented: %+v", coverage)
	}
	if _, err := (Counters{Input: math.MaxInt64, Output: 1}).Total(); err == nil {
		t.Fatal("total overflow admitted")
	}
}

func TestModernCopiedHistoryAndReconstructedCompaction(t *testing.T) {
	r, c := decode(t, "codex", `{"type":"session_meta","payload":{"id":"child"}}
{"type":"session_meta","payload":{"id":"parent"}}
{"type":"turn_context","payload":{"turn_id":"t","model":"a"}}
{"type":"token_usage_record","timestamp":"2026-10-14T10:00:00Z","payload":{"response_id":"parent-response","thread_id":"parent","turn_id":"t","usage":{"input_tokens":100,"output_tokens":20}}}
{"type":"compacted","timestamp":"2026-10-14T10:01:00Z","payload":{"latest_token_usage_record":{"response_id":"child-response","thread_id":"child","turn_id":"t","usage":{"input_tokens":10,"output_tokens":2}}}}
`)
	if len(r) != 1 || r[0].ID != "child-response" || c.Counts["copied_history"] != 1 || c.Counts["reconstructed_compaction"] != 1 || c.SpendComplete() {
		t.Fatalf("copied parent spend admitted: %+v %+v", r, c)
	}
}

func TestAggregateFiltersAfterRevisionAndPreservesMeasuredSubsets(t *testing.T) {
	base := Response{ID: "a", SourceID: "s", Agent: "claude", Model: "a", At: stamp("2026-10-14T09:00:00Z"), Tokens: Counters{Input: 10, Output: 1, CacheRead: p(5)}, Provisional: true}
	final := base
	final.Provisional = false
	final.At = stamp("2026-10-14T10:30:00Z")
	final.Tokens.Output = 2
	other := final
	other.ID = "b"
	other.Model = "b"
	other.Tokens.CacheRead = nil
	w := Window{stamp("2026-10-14T10:00:00Z"), stamp("2026-10-14T11:00:00Z")}
	resolve := func(Response) Attribution {
		return Attribution{Include: true, Slot: "repo:1", Role: "reviewer", Round: "round1", Attempt: "attempt1"}
	}
	r, err := AggregateUsage([]Response{base, final, other}, w, w, resolve)
	if err != nil || r.Totals.Input != 20 || r.Totals.Output != 4 || r.Totals.CacheRead != nil || r.MeasuredDimensions.CacheRead == nil || *r.MeasuredDimensions.CacheRead != 5 || r.Groups[0].Role != "reviewer" || r.Groups[0].Round != "round1" || r.Coverage.Counts["provisional"] != 0 {
		t.Fatalf("report=%+v err=%v", r, err)
	}
	if r.BurnTokensPerHour != 24 {
		t.Fatalf("burn=%v", r.BurnTokensPerHour)
	}
}

func FuzzNormalizeDuplicates(f *testing.F) {
	f.Add(int64(10), int64(5))
	f.Fuzz(func(t *testing.T, in, out int64) {
		if in < 0 || out < 0 || in > math.MaxInt64-out {
			return
		}
		r := Response{ID: "r", SourceID: "s", Agent: "codex", Model: "m", At: stamp("2026-10-14T10:00:00Z"), Tokens: Counters{Input: in, Output: out}}
		got, c, err := NormalizeResponses([]Response{r, r, r})
		if err != nil || len(got) != 1 || !got[0].Tokens.equal(r.Tokens) || c.Counts["duplicates"] != 2 {
			t.Fatalf("dedup=%+v %+v %v", got, c, err)
		}
	})
}

func TestSanitizedRecordedSchemas(t *testing.T) {
	for _, agent := range []string{"codex", "claude"} {
		t.Run(agent, func(t *testing.T) {
			file, err := os.Open("testdata/" + agent + ".jsonl")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			responses, c, err := Decode(context.Background(), agent, "recorded", file)
			if err != nil {
				t.Fatal(err)
			}
			responses, normalized, err := NormalizeResponses(responses)
			c.Merge(normalized)
			if err != nil || len(responses) == 0 || c.Counts["malformed_records"] != 0 {
				t.Fatalf("recorded decode: %+v %+v %v", responses, c, err)
			}
			if agent == "claude" && (len(responses) != 1 || responses[0].Tokens.Input != 53466 || responses[0].Tokens.Output != 236 || *responses[0].Tokens.CacheWrite != 25846 || responses[0].Tokens.Reasoning == nil || *responses[0].Tokens.Reasoning != 20) {
				t.Fatalf("TTL subdivisions or partial blocks double-counted: %+v", responses)
			}
			if agent == "codex" && c.Counts["duplicates"] != 1 {
				t.Fatalf("recorded compaction copy not deduplicated: %+v", c)
			}
		})
	}
}

func TestCompactionCopyDoesNotChangeOriginalTimestamp(t *testing.T) {
	r, _ := decode(t, "codex", `{"type":"session_meta","payload":{"id":"thread"}}
{"type":"turn_context","payload":{"turn_id":"t","model":"a"}}
{"type":"compacted","timestamp":"2026-10-14T09:00:00Z","payload":{"latest_token_usage_record":{"response_id":"r","thread_id":"thread","turn_id":"t","usage":{"input_tokens":10,"output_tokens":2}}}}
{"type":"token_usage_record","timestamp":"2026-10-14T10:00:00Z","payload":{"response_id":"r","thread_id":"thread","turn_id":"t","usage":{"input_tokens":10,"output_tokens":2}}}
`)
	r, _, err := NormalizeResponses(r)
	if err != nil || len(r) != 1 || !r[0].At.Equal(stamp("2026-10-14T10:00:00Z")) {
		t.Fatalf("compaction timestamp replaced actual response timestamp: %+v %v", r, err)
	}
}

func TestDecodePartialTailAndLineBound(t *testing.T) {
	_, c := decode(t, "codex", `{"type":"token_usage_record"`)
	if c.Counts["truncated_tails"] != 1 || c.SpendComplete() {
		t.Fatalf("partial tail coverage=%+v", c)
	}
	_, c, err := Decode(context.Background(), "codex", "s", strings.NewReader(strings.Repeat("x", MaxLineBytes+1)))
	if err == nil || c.Counts["unavailable_sources"] != 1 {
		t.Fatalf("oversized record admitted: %+v %v", c, err)
	}
}

func TestAttributionConflictPreservesIssueExclusionAndFleetUnknown(t *testing.T) {
	w := Window{stamp("2026-10-14T10:00:00Z"), stamp("2026-10-14T11:00:00Z")}
	r := Response{ID: "r", SourceID: "s", Agent: "codex", Model: "a", At: w.Since, Tokens: Counters{Input: 10, Output: 2}}
	for _, include := range []bool{false, true} {
		report, err := AggregateUsage([]Response{r}, w, w, func(Response) Attribution {
			return Attribution{Include: include, Conflict: true, Slot: "repo:1", Role: "worker"}
		})
		if err != nil {
			t.Fatal(err)
		}
		if report.Coverage.Counts["attribution_conflicts"] != 1 {
			t.Fatalf("lost conflict %+v", report)
		}
		if !include {
			if report.Totals.Input != 0 || len(report.Groups) != 0 {
				t.Fatalf("issue exclusion lost %+v", report)
			}
		} else if len(report.Groups) != 1 || report.Groups[0].Slot != "unknown-slot" || report.Groups[0].Role != "unknown" || report.Coverage.Counts["unknown_slot"] != 1 {
			t.Fatalf("fleet unknown spend lost %+v", report)
		}
	}
	report, err := AggregateUsage([]Response{r}, w, w, func(Response) Attribution { return Attribution{Include: true, Slot: "unknown-slot"} })
	if err != nil || report.Coverage.Counts["unknown_slot"] != 1 || !report.Coverage.SpendComplete() {
		t.Fatalf("explicit unknown slot coverage %+v %v", report, err)
	}
}

func TestDiscoveryCoverageSeparatesSpendAndLineage(t *testing.T) {
	for _, kind := range []string{"unavailable_source", "candidate_limit", "unsupported_harness"} {
		var c Coverage
		c.Add(kind, "")
		if c.SpendComplete() {
			t.Errorf("%s missing spend accepted", kind)
		}
	}
	for _, kind := range []string{"unknown_metadata", "invalid_lineage", "dispatch_scan_incomplete"} {
		var c Coverage
		c.Add(kind, "")
		if !c.SpendComplete() || c.Complete() {
			t.Errorf("%s attribution confused with spending", kind)
		}
	}
}

func TestEmptyDetailWindowKeepsIndependentBurn(t *testing.T) {
	at := stamp("2026-10-14T10:00:00Z")
	r := Response{ID: "r", SourceID: "s", Agent: "codex", Model: "m", At: at, Tokens: Counters{Input: 10, Output: 2}}
	empty := Window{Since: at, Until: at}
	burn := Window{Since: at, Until: at.Add(time.Hour)}
	report, err := AggregateUsage([]Response{r}, empty, burn, nil)
	if err != nil || report.Totals.Input != 0 || report.Totals.Output != 0 || len(report.Hourly) != 0 || report.BurnTokensPerHour != 12 {
		t.Fatalf("empty detail obscured independent burn: %+v %v", report, err)
	}
	if _, err := AggregateUsage([]Response{r}, Window{Since: at.Add(time.Second), Until: at}, burn, nil); err == nil {
		t.Fatal("reversed detail admitted")
	}
	if _, err := AggregateUsage([]Response{r}, empty, empty, nil); err == nil {
		t.Fatal("zero burn duration admitted")
	}
}

func TestClaudeStructuredAggregateIsCoverageOnly(t *testing.T) {
	for _, text := range []string{
		`{"type":"result","modelUsage":{"claude-model":{"inputTokens":100,"outputTokens":20,"cacheReadInputTokens":80}}}`,
		`{"type":"result","totalTokens":120}`,
		`{"type":"cost-state","modelUsage":{"claude-model":{"inputTokens":100,"outputTokens":20}},"totalTokens":120}`,
	} {
		responses, c := decode(t, "claude", text)
		if len(responses) != 0 || c.Counts["aggregate_usage_unreconciled"] != 1 || c.SpendComplete() {
			t.Fatalf("aggregate invented response spend: %+v %+v", responses, c)
		}
	}
	text := `{"type":"assistant","timestamp":"2026-10-14T10:00:00Z","requestId":"req","message":{"id":"m","model":"model","stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":2}}}
{"type":"result","modelUsage":{"model":{"inputTokens":10,"outputTokens":2}},"totalTokens":12}`
	responses, c := decode(t, "claude", text)
	w := Window{Since: stamp("2026-10-14T09:00:00Z"), Until: stamp("2026-10-14T11:00:00Z")}
	report, err := AggregateUsage(responses, w, w, nil)
	total, _ := report.Totals.Total()
	if err != nil || total != 12 || c.Counts["aggregate_usage_unreconciled"] != 1 {
		t.Fatalf("response billed with its aggregate: %+v %+v %v", report, c, err)
	}
	for _, invalid := range []string{`{"type":"result","totalTokens":-1}`, `{"type":"result","totalTokens":9223372036854775808}`, `{"type":"result","modelUsage":{"model":{"inputTokens":"unknown"}}}`} {
		responses, c := decode(t, "claude", invalid)
		if len(responses) != 0 || c.Counts["malformed_records"] != 1 || c.SpendComplete() {
			t.Fatalf("invalid aggregate unreported: %+v %+v", responses, c)
		}
	}
	_, c = decode(t, "claude", `{"type":"user","message":{"content":"totalTokens: 5000 modelUsage inputTokens 4000"}}`)
	if c.Counts["aggregate_usage_unreconciled"] != 0 {
		t.Fatalf("user prose parsed as usage %+v", c)
	}
}

func TestCompletenessIncludesAllSpendingAndReviewerGaps(t *testing.T) {
	for _, kind := range []string{"source_mutations", "aggregate_usage_unreconciled", "unresolved_reviewer_usage"} {
		var c Coverage
		c.Add(kind, "")
		if c.Complete() || c.SpendComplete() {
			t.Errorf("%s not incomplete %+v", kind, c)
		}
	}
	for _, kind := range []string{"legacy_reviewer_binding_gaps", "unavailable_reviewer_bindings", "unfinished_reviewer_bindings", "reviewer_binding_bound", "excluded_unbound_responses"} {
		var c Coverage
		c.Add(kind, "")
		if c.Complete() || !c.SpendComplete() {
			t.Errorf("%s attribution/spend confused %+v", kind, c)
		}
	}
}

func TestClaudeAggregateModelBoundAndBoundedExamples(t *testing.T) {
	var entries []string
	for i := 0; i < MaxAggregateModels+1; i++ {
		entries = append(entries, fmt.Sprintf("\"model-%d\":{\"inputTokens\":1}", i))
	}
	responses, c := decode(t, "claude", `{"type":"result","modelUsage":{`+strings.Join(entries, ",")+`}}`)
	if len(responses) != 0 || c.Counts["malformed_records"] != 1 || c.SpendComplete() {
		t.Fatalf("aggregate model bound bypassed %+v %+v", responses, c)
	}
	_, c = decode(t, "claude", strings.Repeat("{\"type\":\"result\",\"totalTokens\":10}\n", 10))
	if c.Counts["aggregate_usage_unreconciled"] != 10 || len(c.Examples["aggregate_usage_unreconciled"]) != 3 {
		t.Fatalf("unbounded diagnostic %+v", c)
	}
}
