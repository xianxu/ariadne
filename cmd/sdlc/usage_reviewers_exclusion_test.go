package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

func exclusionReport(t *testing.T, roots []string, selectedIssue int, record reviewusage.Record) usage.Report {
	t.Helper()
	records, coverage, err := loadUsageReviewers(context.Background(), roots, selectedIssue)
	if err != nil {
		t.Fatal(err)
	}
	if !coverage.Complete() {
		t.Fatalf("valid accounting became incomplete: %+v", coverage)
	}
	sessions := []transcripts.SessionEvidence{usageAttributionSession(record.SessionID, record.Intent.Root, "", false, 0)}
	bindings := bindUsageReviewers(sessions, records, usageAttributionAt(30))
	response := usage.Response{ID: "response", Agent: "codex", SourceID: record.SessionID, Model: "fixture", At: usageAttributionAt(15), Tokens: usage.Counters{Input: 17, Output: 3}}
	report, err := usage.AggregateUsage([]usage.Response{response}, usage.Window{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}, usage.Window{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}, func(r usage.Response) usage.Attribution {
		return usageDetailAttribution(r, usage.Attribution{Role: "worker", Slot: "repo:3", Include: true}, bindings, usageAuthority{Repository: record.Intent.Repository}, selectedIssue, true)
	})
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func TestUsageReviewerOtherIssueExclusionSurvivesReceiptCleanup(t *testing.T) {
	for _, format := range []string{"ledger", "comment"} {
		t.Run(format, func(t *testing.T) {
			ctx := context.Background()
			root := testfix.Repo(t, testfix.InitialCommit())
			record := usageAttributionRecord(t, "other-issue", usageAttributionReviewerID, root, 1, 1)
			record.Intent.IssueID = "000366"
			store, err := reviewusage.NewStore(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := store.SaveRequested(ctx, record.Intent); err != nil {
				t.Fatal(err)
			}
			if err := store.Record(ctx, record); err != nil {
				t.Fatal(err)
			}
			assertTotals := func(stage string) {
				t.Helper()
				excluded := usageIssueParts(exclusionReport(t, []string{root}, 365, record))
				if excluded[0].Tokens != 0 || excluded[1].Tokens != 0 || excluded[2].Tokens != 0 {
					t.Fatalf("%s: other issue reviewer became selected worker: %+v", stage, excluded)
				}
				included := usageIssueParts(exclusionReport(t, []string{root}, 366, record))
				if included[0].Tokens != 0 || included[1].Tokens != 20 || included[1].Rounds != 1 || included[1].Attempts != 1 || included[2].Tokens != 20 {
					t.Fatalf("%s: reviewer lost or counted twice: %+v", stage, included)
				}
			}
			assertTotals("private-only")
			var raw, suffix string
			if format == "ledger" {
				raw = gatestate.Render(gatestate.Ledger{Gate: "plan-quality", IssueNum: 366, Rounds: []gatestate.Round{{N: 1, ReviewerRuns: []reviewusage.Record{record}}}}, "repo")
				suffix = "gate"
			} else {
				encoded, err := json.Marshal([]reviewusage.Record{record})
				if err != nil {
					t.Fatal(err)
				}
				raw = "# Review\n<!-- sdlc-review-usage " + string(encoded) + " -->\n"
				suffix = "review"
			}
			path := filepath.Join(root, "workshop/plans/000366-other-plan-"+suffix+".md")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
				t.Fatal(err)
			}
			assertTotals("receipt-and-uncommitted-artifact")
			testfix.Git(t, root, "add", "workshop")
			testfix.Git(t, root, "commit", "-qm", "persist other issue reviewer")
			if err := cleanupReviewerAccounting(ctx, root, "origin", store); err != nil {
				t.Fatal(err)
			}
			entries, err := store.List(ctx)
			if err != nil || len(entries) != 0 {
				t.Fatalf("committed receipt not collected: %+v %v", entries, err)
			}
			assertTotals("artifact-after-actual-cleanup")
		})
	}
}

func TestUsageReviewerExclusionIncludesOtherRepositories(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	peer := testfix.Repo(t, testfix.InitialCommit())
	record := usageAttributionRecord(t, "other-repo", usageAttributionReviewerID, root, 1, 1)
	record.Intent.Repository = "github.com/example/other-repo"
	record.Intent.IssueID = "000366"
	path := filepath.Join(peer, "workshop/plans/000366-other-plan-gate.md")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	raw := gatestate.Render(gatestate.Ledger{Gate: "plan-quality", IssueNum: 366, ReviewerRuns: []reviewusage.Record{record}}, "repo")
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	report := exclusionReport(t, reviewRepositories(root, []string{peer}), 365, record)
	if parts := usageIssueParts(report); parts[0].Tokens != 0 || parts[2].Tokens != 0 {
		t.Fatalf("other-repo reviewer became worker: %+v", parts)
	}
}

func TestUsageReviewerExclusionKeepsSelectedLegacyDiagnostics(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	for _, issue := range []int{365, 366} {
		path := filepath.Join(root, fmt.Sprintf("workshop/plans/%06d-legacy-plan-gate.md", issue))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		raw := gatestate.Render(gatestate.Ledger{Gate: "plan-quality", IssueNum: issue, Rounds: []gatestate.Round{{N: 1}}}, "repo")
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
	}
	_, coverage, err := loadUsageReviewers(context.Background(), []string{root}, 365)
	if err != nil {
		t.Fatal(err)
	}
	if coverage.Counts["legacy_reviewer_binding_gaps"] != 1 {
		t.Fatalf("selected issue inherited unrelated legacy gaps: %+v", coverage)
	}
	_, fleetCoverage, err := loadUsageReviewers(context.Background(), []string{root}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if fleetCoverage.Counts["legacy_reviewer_binding_gaps"] != 2 {
		t.Fatalf("fleet legacy gaps disappeared: %+v", fleetCoverage)
	}
}

func TestUsageReviewerUnfilteredEvidenceBoundIsExplicit(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	dir := filepath.Join(root, "workshop/plans")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4097; i++ {
		path := filepath.Join(dir, fmt.Sprintf("000366-bound-%04d-gate.md", i))
		if err := os.WriteFile(path, []byte("---\ngate: plan-quality\n---\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	_, coverage, err := loadUsageReviewers(context.Background(), []string{root}, 365)
	if err != nil {
		t.Fatal(err)
	}
	if coverage.Complete() || coverage.Counts["reviewer_binding_bound"] != 1 {
		t.Fatalf("cross-issue exclusion scan silently complete at bound: %+v", coverage)
	}
}

func TestUsageReviewerUnselectedUnreadableEvidenceRemainsVisible(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	record := usageAttributionRecord(t, "invalid-other", usageAttributionReviewerID, root, 1, 1)
	record.Intent.IssueID = "000366"
	record.Version = 99
	raw := gatestate.Render(gatestate.Ledger{Gate: "plan-quality", IssueNum: 366, ReviewerRuns: []reviewusage.Record{record}}, "repo")
	path := filepath.Join(root, "workshop/plans/000366-invalid-plan-gate.md")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	records, coverage, err := loadUsageReviewers(context.Background(), []string{root}, 365)
	if err != nil {
		t.Fatal(err)
	}
	if coverage.Complete() || coverage.Counts["unavailable_reviewer_bindings"] != 1 {
		t.Fatalf("other issue's unreadable exclusion evidence disappeared: %+v", coverage)
	}
	sessions := []transcripts.SessionEvidence{usageAttributionSession(record.SessionID, root, "", false, 0)}
	bindings := bindUsageReviewers(sessions, records, usageAttributionAt(30))
	if binding, ok := bindings["codex:"+record.SessionID]; !ok || !binding.Conflict {
		t.Fatalf("other issue's invalid reviewer identity became worker: %+v", bindings)
	}
}
