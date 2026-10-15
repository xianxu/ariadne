package usage_test

import (
	"context"
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
					} else if source.MetadataComplete || source.ID == "parent" || total != 132 || coverage.Counts["unknown_metadata"] == 0 || coverage.SpendComplete() {
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
