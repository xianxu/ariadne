package main

import (
	"context"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewerAccountingCleanupRequiresCommittedExactRecord(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	ctx := context.Background()
	store, e := reviewusage.NewStore(root)
	if e != nil {
		t.Fatal(e)
	}
	r := usageAttributionRecord(t, "retained", usageAttributionReviewerID, "/repo", 1, 1)
	r.Intent.ReviewedHEAD = strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	requested, e := reviewusage.NewRequested(r.Intent)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = store.SaveRequested(ctx, r.Intent); e != nil {
		t.Fatal(e)
	}
	if e = store.Record(ctx, r); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(root, "workshop", "plans", "000365-usage-close-gate.md")
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		t.Fatal(e)
	}
	ledger := gatestate.Ledger{Gate: "boundary-review", IssueNum: 365, IDPrefix: "BR", Rounds: []gatestate.Round{{N: 1, ReviewerRuns: []reviewusage.Record{r}}}}
	if e = os.WriteFile(path, []byte(gatestate.Render(ledger, "repo")), 0600); e != nil {
		t.Fatal(e)
	}
	if e = cleanupReviewerAccounting(ctx, root, "origin", store); e != nil {
		t.Fatal(e)
	}
	if _, e = store.Load(ctx, r.Intent.ID); e != nil {
		t.Fatalf("uncommitted copy deleted receipt: %v", e)
	}
	ledger.Rounds[0].ReviewerRuns = []reviewusage.Record{requested}
	if e = os.WriteFile(path, []byte(gatestate.Render(ledger, "repo")), 0600); e != nil {
		t.Fatal(e)
	}
	testfix.Git(t, root, "add", "workshop")
	testfix.Git(t, root, "commit", "-qm", "incomplete accounting copy")
	if e = cleanupReviewerAccounting(ctx, root, "origin", store); e != nil {
		t.Fatal(e)
	}
	if _, e = store.Load(ctx, r.Intent.ID); e != nil {
		t.Fatalf("partial committed copy deleted receipt: %v", e)
	}
	ledger.Rounds[0].ReviewerRuns = []reviewusage.Record{r}
	if e = os.WriteFile(path, []byte(gatestate.Render(ledger, "repo")), 0600); e != nil {
		t.Fatal(e)
	}
	testfix.Git(t, root, "add", "workshop")
	testfix.Git(t, root, "commit", "-qm", "exact accounting copy")
	if e = cleanupReviewerAccounting(ctx, root, "origin", store); e != nil {
		t.Fatal(e)
	}
	entries, e := store.List(ctx)
	if e != nil || len(entries) != 0 {
		t.Fatalf("committed exact copy retained: %+v %v", entries, e)
	}
}
func TestReviewerBindingArtifactsRoundTripAndLegacyBytes(t *testing.T) {
	r := usageAttributionRecord(t, "retained", usageAttributionReviewerID, "/repo", 1, 1)
	ledger := gatestate.Ledger{Gate: "plan-quality", IssueNum: 365, IDPrefix: "PQ", ReviewerRuns: []reviewusage.Record{r}, Rounds: []gatestate.Round{{N: 1, ReviewerRuns: []reviewusage.Record{r}}}}
	raw := gatestate.Render(ledger, "repo")
	got := reviewerRecordsInArtifact(raw)
	if len(got) != 2 || got[0] != r || got[1] != r {
		t.Fatal(got)
	}
	m := sidecarMeta{Repo: "repo", IssueNum: 365, ReviewerRuns: []reviewusage.Record{r}}
	got = reviewerRecordsInArtifact(renderReviewEntry(m, false))
	if len(got) != 1 || got[0] != r {
		t.Fatal(got)
	}
	legacy := gatestate.Render(gatestate.Ledger{Gate: "plan-quality", IssueNum: 365, IDPrefix: "PQ"}, "repo")
	if strings.Contains(legacy, "reviewer_runs") {
		t.Fatal(legacy)
	}
	m.ReviewerRuns = nil
	if strings.Contains(renderReviewEntry(m, false), "sdlc-review-usage") {
		t.Fatal("legacy sidecar changed")
	}
}
