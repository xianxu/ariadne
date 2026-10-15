package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

func coverageClaudeSource(t *testing.T, root, id, mode string, tokens int) {
	t.Helper()
	path := filepath.Join(os.Getenv("HOME"), ".claude/projects/fixture", id+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	stamp := "2026-10-14T00:10:00Z"
	if mode == "invalid-metadata" {
		stamp = "not-a-timestamp"
	}
	data := fmt.Sprintf("{\"type\":\"user\",\"sessionId\":%q,\"cwd\":%q,\"timestamp\":%q}\n", id, root, stamp)
	if mode == "response" {
		data += fmt.Sprintf("{\"type\":\"assistant\",\"timestamp\":\"2026-10-14T00:15:00Z\",\"message\":{\"id\":%q,\"model\":\"fixture\",\"stop_reason\":\"end_turn\",\"usage\":{\"input_tokens\":%d,\"output_tokens\":0,\"cache_read_input_tokens\":0,\"cache_creation_input_tokens\":0}}}\n", "message-"+id, tokens)
	}
	if mode == "malformed-source" {
		data += "this is not JSON\n"
	}
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func coverageReviewerFixture(t *testing.T, format, mode string) (string, reviewusage.Record) {
	t.Helper()
	ctx := context.Background()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("GIT_AUTHOR_DATE", "2026-10-14T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-14T00:00:00Z")
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, root, "workshop/issues/000365-usage.md", "---\nid: 365\nstatus: working\n---\n# Usage\n")
	testfix.Git(t, root, "add", "workshop")
	testfix.Git(t, root, "commit", "-qm", "#365: fixture")
	record := usageAttributionRecord(t, "collector-retained", usageAttributionReviewerID, root, 1, 1)
	record.Intent.Agent = "claude"
	record.Intent.Recipe = "plan-quality"
	repository, err := resolveUsageRepository(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	record.Intent.Repository = repository
	record.Intent.ReviewedHEAD = strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	// Both account readings are usable: the independent worker response below
	// makes a forecast possible unless the retained reviewer proves a spend gap.
	meterPath, err := usage.MeterPath()
	if err != nil {
		t.Fatal(err)
	}
	for i, at := range []int{5, 25} {
		if _, err := usage.RecordMeter(ctx, meterPath, usage.Observation{Agent: "claude", Unit: "credits", Remaining: float64(100 - i*10), At: usageAttributionAt(at)}); err != nil {
			t.Fatal(err)
		}
	}
	coverageClaudeSource(t, root, usageAttributionOtherID, "response", 20)
	switch mode {
	case "missing-id":
		record.SessionID = ""
	case "unavailable-exact":
	case "invalid-metadata", "metadata-only", "malformed-source":
		coverageClaudeSource(t, root, record.SessionID, mode, 0)
	case "launch-failed":
		record.PID = 0
		record.StartedAt = ""
		record.SessionID = ""
		record.Outcome = "launch-failed"
	case "zero-response", "finished-prefix":
		coverageClaudeSource(t, root, record.SessionID, "response", 0)
	case "positive-response":
		coverageClaudeSource(t, root, record.SessionID, "response", 10)
	default:
		t.Fatalf("unknown fixture mode %s", mode)
	}
	if err := reviewusage.Validate(record); err != nil {
		t.Fatalf("fixture must remain schema-valid: %v", err)
	}
	store, err := reviewusage.NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.SaveRequested(ctx, record.Intent); err != nil {
		t.Fatal(err)
	}
	if mode != "finished-prefix" {
		if err := store.Record(ctx, record); err != nil {
			t.Fatal(err)
		}
	}
	if format != "private" {
		var raw string
		if format == "ledger" {
			raw = gatestate.Render(gatestate.Ledger{Gate: "plan-quality", IssueNum: 365, Rounds: []gatestate.Round{{N: 1, ReviewerRuns: []reviewusage.Record{record}}}}, "fixture")
		} else {
			encoded, err := json.Marshal([]reviewusage.Record{record})
			if err != nil {
				t.Fatal(err)
			}
			raw = "<!-- sdlc-review-usage " + string(encoded) + " -->\n"
		}
		suffix := "gate"
		if format == "comment" {
			suffix = "review"
		}
		writeRepoFile(t, root, "workshop/plans/000365-usage-plan-"+suffix+".md", raw)
		testfix.Git(t, root, "add", "workshop")
		testfix.Git(t, root, "commit", "-qm", "persist retained reviewer")
		if err := cleanupReviewerAccounting(ctx, root, "origin", store); err != nil {
			t.Fatal(err)
		}
		entries, err := store.List(ctx)
		if mode == "finished-prefix" {
			if err != nil || len(entries) != 1 || entries[0].Record.Phase != reviewusage.Requested {
				t.Fatalf("fixture requested prefix lost: %v %+v", err, entries)
			}
		} else if err != nil || len(entries) != 0 {
			t.Fatalf("fixture private receipt was not collected: %v %+v", err, entries)
		}
	}
	return root, record
}

func TestUsageCollectorRetainedReviewerCoverage(t *testing.T) {
	for _, format := range []string{"private", "ledger", "comment"} {
		for _, mode := range []string{"missing-id", "unavailable-exact", "invalid-metadata", "metadata-only", "malformed-source", "launch-failed", "zero-response", "positive-response", "finished-prefix"} {
			if format == "private" && mode == "finished-prefix" {
				continue
			}
			t.Run(format+"/"+mode, func(t *testing.T) {
				_, _ = coverageReviewerFixture(t, format, mode)
				incomplete := mode != "launch-failed" && mode != "zero-response" && mode != "positive-response" && mode != "finished-prefix"
				for _, issue := range []int{365, 0} {
					out, err := collectUsage(context.Background(), usageFlags{Issue: issue, Since: "1h", Burn: 3 * time.Hour}, usageAttributionAt(30))
					if err != nil {
						t.Fatal(err)
					}
					var claude, codex *agentGlance
					for i := range out.AgentGlance {
						switch out.AgentGlance[i].Agent {
						case "claude":
							claude = &out.AgentGlance[i]
						case "codex":
							codex = &out.AgentGlance[i]
						}
					}
					if claude == nil || codex == nil {
						t.Fatalf("missing fleet glance: %+v", out.AgentGlance)
					}
					if incomplete {
						if out.Report.Coverage.Counts["unresolved_reviewer_usage"] != 1 || out.Report.Coverage.SpendComplete() {
							t.Errorf("issue=%d report discarded unresolved receipt: %+v", issue, out.Report.Coverage)
						}
						if out.Coverage.Counts["unresolved_reviewer_usage"] != 1 || out.Coverage.SpendComplete() || out.Coverage.Complete() {
							t.Errorf("issue=%d retained launched spend disappeared: %+v", issue, out.Coverage)
						}
						if claude.Coverage.Counts["unresolved_reviewer_usage"] != 1 || claude.Coverage.SpendComplete() || claude.Projection.Status != "incomplete_evidence" || claude.Projection.HoursRemaining != nil {
							t.Errorf("issue=%d unresolved spend escaped forecast input: %+v", issue, claude)
						}
					} else {
						if !out.Coverage.SpendComplete() || out.Coverage.Counts["unresolved_reviewer_usage"] != 0 || out.Coverage.Counts["unfinished_reviewer_bindings"] != 0 {
							t.Errorf("issue=%d usable/no-launch control marked incomplete: %+v", issue, out.Coverage)
						}
						if !claude.Coverage.SpendComplete() || claude.Projection.Status != "calibrated" || claude.Projection.HoursRemaining == nil {
							t.Errorf("issue=%d valid control lost calibrated forecast: %+v", issue, claude)
						}
						if issue > 0 {
							want := int64(0)
							if mode == "positive-response" {
								want = 10
							}
							if out.IssueParts[0].Tokens != 20 || out.IssueParts[1].Tokens != want || out.IssueParts[2].Tokens != 20+want {
								t.Errorf("usable reviewer/worker union: %+v", out.IssueParts)
							}
						}
					}
					if !codex.Coverage.SpendComplete() || codex.Coverage.Counts["unresolved_reviewer_usage"] != 0 {
						t.Errorf("unrelated agent lost spend coverage: %+v", codex)
					}
				}
			})
		}
	}
}
