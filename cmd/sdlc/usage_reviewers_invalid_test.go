package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

func TestUsageReviewerInvalidRetainedEvidenceCannotBecomeWorker(t *testing.T) {
	for _, format := range []string{"ledger", "comment", "private"} {
		t.Run(format, func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit())
			record := usageAttributionRecord(t, "invalid-version", usageAttributionReviewerID, root, 1, 1)
			record.Version = 99
			var raw, suffix string
			if format == "private" {
				store, err := gitx.NewRecoveryStoreFor(context.Background(), root, "review-usage")
				if err != nil {
					t.Fatal(err)
				}
				encoded, err := json.Marshal(record)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := store.Save(record.Intent.ID, "", gitx.RecoveryData{Document: encoded}); err != nil {
					t.Fatal(err)
				}
			} else if format == "ledger" {
				raw = gatestate.Render(gatestate.Ledger{Gate: "boundary-review", IssueNum: 365, Rounds: []gatestate.Round{{N: 1, ReviewerRuns: []reviewusage.Record{record}}}}, "repo")
				suffix = "gate"
			} else {
				encoded, err := json.Marshal([]reviewusage.Record{record})
				if err != nil {
					t.Fatal(err)
				}
				raw = "# Review\n<!-- sdlc-review-usage " + string(encoded) + " -->\n"
				suffix = "review"
			}
			if format != "private" {
				path := filepath.Join(root, "workshop/plans/000365-usage-close-"+suffix+".md")
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
					t.Fatal(err)
				}
			}
			records, coverage, err := loadUsageReviewers(context.Background(), []string{root}, 365)
			if err != nil {
				t.Fatal(err)
			}
			if coverage.Complete() || coverage.Counts["unavailable_reviewer_bindings"] == 0 {
				t.Errorf("invalid retained evidence reported complete: %+v", coverage)
			}
			sessions := []transcripts.SessionEvidence{usageAttributionSession(usageAttributionReviewerID, root, "", false, 0)}
			bindings := bindUsageReviewers(sessions, records, usageAttributionAt(30))
			response := usage.Response{Agent: "codex", SourceID: usageAttributionReviewerID, At: usageAttributionAt(15)}
			got := usageDetailAttribution(response, usage.Attribution{Slot: "repo:3", Role: "worker", Include: true}, bindings, usageAuthority{Repository: record.Intent.Repository}, 365, true)
			if got.Role != "reviewer" || !got.Conflict || got.Include {
				t.Fatalf("invalid reviewer became ordinary worker: %+v bindings=%+v", got, bindings)
			}
		})
	}
}

func TestUsageReviewerMalformedArtifactsAreVisible(t *testing.T) {
	for name, raw := range map[string]string{
		"ledger":              "---\nrounds: [\n---\n",
		"missing-frontmatter": "broken ledger",
		"comment":             "# Review\n<!-- sdlc-review-usage [broken -->\n",
		"comment-terminator":  "# Review\n<!-- sdlc-review-usage []\n",
	} {
		t.Run(name, func(t *testing.T) {
			root := testfix.Repo(t, testfix.InitialCommit())
			suffix := "gate"
			if strings.HasPrefix(name, "comment") {
				suffix = "review"
			}
			path := filepath.Join(root, "workshop/plans/000365-usage-close-"+suffix+".md")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			_, coverage, err := loadUsageReviewers(context.Background(), []string{root}, 365)
			if err != nil {
				t.Fatal(err)
			}
			if coverage.Complete() || coverage.Counts["unavailable_reviewer_bindings"] == 0 {
				t.Fatalf("malformed evidence silently complete: %+v", coverage)
			}
		})
	}
}

func TestUsageReviewerInvalidIdentityQuarantine(t *testing.T) {
	for _, mode := range []string{"invalid-time-exact", "invalid-version-root", "malformed-comment"} {
		t.Run(mode, func(t *testing.T) {
			record := usageAttributionRecord(t, "quarantine", usageAttributionReviewerID, "/disposable", 1, 1)
			session := usageAttributionSession(usageAttributionReviewerID, "/disposable", "", false, 12)
			var records []reviewusage.Record
			if mode == "invalid-time-exact" {
				record.Intent.RequestedAt = "bad"
				records = []reviewusage.Record{record}
			} else if mode == "invalid-version-root" {
				record.Version = 99
				record.SessionID = ""
				record.Intent.DisposableRoot = true
				records = []reviewusage.Record{record}
			} else {
				encoded, err := json.Marshal([]reviewusage.Record{record})
				if err != nil {
					t.Fatal(err)
				}
				var coverage usage.Coverage
				records, coverage = decodeReviewerAccountingArtifact("<!-- sdlc-review-usage " + string(encoded))
				if coverage.Complete() {
					t.Fatal("malformed comment complete")
				}
			}
			got := bindUsageReviewers([]transcripts.SessionEvidence{session}, records, usageAttributionAt(30))
			if b, ok := got["codex:"+session.ID]; !ok || !b.Conflict {
				t.Fatalf("recoverable invalid identity disappeared: %+v", got)
			}
		})
	}
}

func TestUsageReviewerInvalidAccountingCannotCollectPrivateReceipt(t *testing.T) {
	record := usageAttributionRecord(t, "quarantine", usageAttributionReviewerID, "/repo", 1, 1)
	encoded, err := json.Marshal([]reviewusage.Record{record})
	if err != nil {
		t.Fatal(err)
	}
	raw := "<!-- sdlc-review-usage " + string(encoded)
	if got := reviewerRecordsInArtifact(raw); len(got) != 0 {
		t.Fatalf("malformed comment authorized cleanup: %+v", got)
	}
	record.Version = 99
	raw = gatestate.Render(gatestate.Ledger{ReviewerRuns: []reviewusage.Record{record}}, "repo")
	if got := reviewerRecordsInArtifact(raw); len(got) != 0 {
		t.Fatalf("invalid ledger authorized cleanup: %+v", got)
	}
}
