package usage_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

func TestCodexDiscoveryDecodeExecutingIdentityContract(t *testing.T) {
	metadata := func(payload string) string {
		return `{"type":"session_meta","timestamp":"2026-10-14T10:00:00Z","payload":` + payload + "}\n"
	}
	response := func(id, thread string, input, output int64) string {
		return fmt.Sprintf("{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-14T10:01:00Z\",\"payload\":{\"response_id\":%q,\"thread_id\":%q,\"turn_id\":\"turn\",\"usage\":{\"input_tokens\":%d,\"output_tokens\":%d}}}\n", id, thread, input, output)
	}
	parentMeta := metadata(`{"id":"parent","cwd":"/repo"}`)
	for _, tc := range []struct {
		name, first string
		valid       bool
	}{
		{"id", `{"id":"child","cwd":"/repo"}`, true},
		{"thread_id fallback", `{"thread_id":"child","cwd":"/repo"}`, true},
		{"equal aliases", `{"id":"child","thread_id":"child","cwd":"/repo"}`, true},
		{"session_id is copied ancestor", `{"thread_id":"child","session_id":"parent","cwd":"/repo"}`, true},
		{"empty first metadata", `{}`, false},
		{"missing identity", `{"cwd":"/repo"}`, false},
		{"invalid first identity type", `{"id":4,"cwd":"/repo"}`, false},
		{"invalid payload", `"broken"`, false},
		{"malformed first metadata JSON", `{"id":`, false},
		{"conflicting aliases", `{"id":"child","thread_id":"parent","cwd":"/repo"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, copiedMetadata := range []string{"", parentMeta} {
				t.Run(fmt.Sprintf("copied_metadata_%t", copiedMetadata != ""), func(t *testing.T) {
					root := t.TempDir()
					text := metadata(tc.first) + copiedMetadata + response("child-response", "child", 10, 2) + response("parent-response", "parent", 100, 20)
					path := filepath.Join(root, "session.jsonl")
					if err := os.WriteFile(path, []byte(text), 0600); err != nil {
						t.Fatal(err)
					}
					discovery, err := transcripts.DiscoverUsageEvidence(context.Background(), []transcripts.Harness{transcripts.CodexHarness(root)}, time.Time{})
					if err != nil || len(discovery.Sources) != 1 {
						t.Fatalf("discovery=%+v err=%v", discovery, err)
					}
					source := discovery.Sources[0]
					responses, coverage, err := usage.Decode(context.Background(), "codex", source.ID, strings.NewReader(text))
					if err != nil {
						t.Fatal(err)
					}
					var total int64
					for _, r := range responses {
						n, err := r.Tokens.Total()
						if err != nil {
							t.Fatal(err)
						}
						total += n
					}
					actual, warnings, actualErr := transcripts.DiscoverEvidence(context.Background(), []string{"/repo"}, []transcripts.Harness{transcripts.CodexHarness(root)})
					if tc.valid {
						if !source.MetadataComplete || source.ID != "child" || total != 12 || coverage.Counts["copied_history"] != 1 || !coverage.Complete() {
							t.Fatalf("discovery and decoder disagree: source=%+v total=%d coverage=%+v", source, total, coverage)
						}
						if actualErr != nil || len(actual) != 1 || actual[0].ID != "child" {
							t.Fatalf("actual discovery identity disagrees: %+v %v", actual, actualErr)
						}
					} else if source.MetadataComplete || source.ID == "parent" || total != 0 || coverage.Counts["unknown_metadata"] == 0 || coverage.SpendComplete() {
						t.Fatalf("later copied metadata repaired invalid first identity: source=%+v total=%d coverage=%+v", source, total, coverage)
					}
					if !tc.valid && (len(actual) != 0 || actualErr == nil && len(warnings) == 0) {
						t.Fatalf("actual discovery repaired or hid invalid first identity: %+v warnings=%v err=%v", actual, warnings, actualErr)
					}
				})
			}
		})
	}
}

func TestCodexExecutingIdentityOrdering(t *testing.T) {
	idHeader := `{"type":"session_meta","timestamp":"2026-10-14T08:00:00Z","payload":{"id":"child","cwd":"/repo"}}` + "\n"
	threadHeader := `{"type":"session_meta","timestamp":"2026-10-14T08:00:00Z","payload":{"thread_id":"child","cwd":"/repo"}}` + "\n"
	invalidHeader := `{"type":"session_meta","payload":{"id":"child","thread_id":"parent"}}` + "\n"
	ordinary := func(thread, at string, input, output int64) string {
		return fmt.Sprintf("{\"type\":\"token_usage_record\",\"timestamp\":%q,\"payload\":{\"response_id\":\"same-response\",\"thread_id\":%q,\"turn_id\":\"turn\",\"usage\":{\"input_tokens\":%d,\"output_tokens\":%d},\"thread_token_usage\":{\"input_tokens\":%d,\"output_tokens\":%d},\"turn_token_usage\":{\"input_tokens\":%d,\"output_tokens\":%d}}}\n", at, thread, input, output, input, output, input, output)
	}
	compacted := func(ordinary string) string {
		var record map[string]json.RawMessage
		if err := json.Unmarshal([]byte(ordinary), &record); err != nil {
			t.Fatal(err)
		}
		record["type"] = json.RawMessage(`"compacted"`)
		payload, err := json.Marshal(map[string]json.RawMessage{"latest_token_usage_record": record["payload"]})
		if err != nil {
			t.Fatal(err)
		}
		record["payload"] = payload
		out, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		return string(out) + "\n"
	}

	decodeDiscovered := func(t *testing.T, text string) ([]usage.Response, usage.Coverage, transcripts.UsageSource) {
		t.Helper()
		root := t.TempDir()
		path := filepath.Join(root, "session.jsonl")
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
		discovered, err := transcripts.DiscoverUsageEvidence(context.Background(), []transcripts.Harness{transcripts.CodexHarness(root)}, time.Time{})
		if err != nil || len(discovered.Sources) != 1 {
			t.Fatalf("discovery=%+v err=%v", discovered, err)
		}
		source := discovered.Sources[0]
		r, c, err := usage.Decode(context.Background(), "codex", source.ID, strings.NewReader(text))
		if err != nil {
			t.Fatal(err)
		}
		return r, c, source
	}
	parent := ordinary("parent", "2026-10-14T09:00:00Z", 100, 20)
	child := ordinary("child", "2026-10-14T10:00:00Z", 10, 2)
	for _, header := range []struct{ name, text string }{{"id", idHeader}, {"thread_id", threadHeader}} {
		for _, parentKind := range []struct{ name, text string }{{"ordinary", parent}, {"compacted", compacted(parent)}} {
			for _, late := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/late_%t", header.name, parentKind.name, late), func(t *testing.T) {
					text := header.text + parentKind.text + child
					if late {
						text = parentKind.text + header.text + child
					}
					responses, c, source := decodeDiscovered(t, text)
					if !source.MetadataComplete || source.ID != "child" {
						t.Fatalf("executing identity not verified: %+v", source)
					}
					if len(responses) != 1 || responses[0].Tokens.Input != 10 || responses[0].Tokens.Output != 2 || !responses[0].At.Equal(time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)) {
						t.Fatalf("parent spend or timestamps tainted child: %+v %+v", responses, c)
					}
					if c.SpendComplete() == late || c.Counts["reconstructed_compaction"] != 0 || c.Counts["uncovered_segments"] != map[bool]int64{false: 0, true: 1}[late] {
						t.Fatalf("identity ordering/cumulative bookkeeping incorrect: %+v", c)
					}
				})
			}
		}
	}
	for _, tc := range []struct{ name, text string }{
		{"absent ordinary", parent + child},
		{"absent compacted", compacted(parent) + compacted(child)},
		{"invalid early", invalidHeader + parent + child},
		{"invalid late", parent + invalidHeader + child},
		{"empty", ""},
		{"legacy only", `{"type":"event_msg","timestamp":"2026-10-14T10:00:00Z","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":10,"output_tokens":2}}}}` + "\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, c, source := decodeDiscovered(t, tc.text)
			if source.MetadataComplete || len(r) != 0 || c.SpendComplete() || c.Counts["unknown_metadata"] == 0 || c.Counts["uncovered_segments"] == 0 {
				t.Fatalf("unidentified source confidently counted: %+v %+v source=%+v", r, c, source)
			}
		})
	}
	legacy := `{"type":"event_msg","timestamp":"2026-10-14T09:00:00Z","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":500,"output_tokens":100}}}}` + "\n"
	t.Run("legacy prefix does not establish cumulative baseline", func(t *testing.T) {
		r, c, err := usage.Decode(context.Background(), "codex", "fixture", strings.NewReader(legacy+idHeader+strings.ReplaceAll(strings.ReplaceAll(legacy, "500", "10"), "100", "2")+child))
		if err != nil || len(r) != 1 || r[0].Tokens.Input != 10 || c.Counts["uncovered_segments"] != 1 || c.Counts["baseline_gaps"] != 0 || c.Counts["reset_gaps"] != 0 || c.SpendComplete() {
			t.Fatalf("preidentity cumulative evidence tainted owned response: %+v %+v %v", r, c, err)
		}
	})
	t.Run("valid identity without cwd remains complete", func(t *testing.T) {
		text := `{"type":"session_meta","payload":{"id":"child"}}` + "\n" + child
		r, c, source := decodeDiscovered(t, text)
		if source.MetadataComplete || len(r) != 1 || !c.SpendComplete() {
			t.Fatalf("missing cwd confused with missing identity: %+v %+v source=%+v", r, c, source)
		}
	})
	for _, before := range []struct{ name, text string }{{"ordinary", child}, {"compacted", compacted(child)}} {
		t.Run("owned prefix rejected/"+before.name, func(t *testing.T) {
			r, c, _ := decodeDiscovered(t, before.text+idHeader)
			if len(r) != 0 || c.SpendComplete() || c.Counts["uncovered_segments"] != 1 {
				t.Fatalf("owned prefix guessed before identity: %+v %+v", r, c)
			}
		})
	}
	t.Run("parent original cannot repair child compaction", func(t *testing.T) {
		r, c, err := usage.Decode(context.Background(), "codex", "fixture", strings.NewReader(parent+idHeader+compacted(child)))
		if err != nil || len(r) != 1 || !r[0].At.Equal(time.Date(2026, 10, 14, 10, 0, 0, 0, time.UTC)) || c.Counts["reconstructed_compaction"] != 1 || c.SpendComplete() {
			t.Fatalf("preidentity original repaired child compaction: %+v %+v %v", r, c, err)
		}
	})
}
