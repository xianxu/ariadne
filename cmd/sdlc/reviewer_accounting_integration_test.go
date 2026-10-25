package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

// These tests enter the production gate, never supply UsageIntent or a recorder,
// and use an actual fake CLI process. Removing a gate's accounting attachment
// must therefore break both its persisted receipt and its issue attribution.
func accountingCLI(t *testing.T, verdict string) {
	t.Helper()
	dir := t.TempDir()
	text, err := json.Marshal("VERDICT: " + verdict + emptyBoundaryFindings)
	if err != nil {
		t.Fatal(err)
	}
	stream := `{"type":"system","subtype":"init","session_id":"` + "00000000-0000-4000-8000-SESSION" + `"}` + "\n" +
		`{"type":"assistant","message":{"content":[{"type":"text","text":` + string(text) + `}]}}` + "\n" +
		`{"type":"result","subtype":"success","session_id":"` + "00000000-0000-4000-8000-SESSION" + `"}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "claude"), []byte("#!/bin/sh\nsid=$(printf '%012d' \"$$\")\ncat <<'REPLY' | sed \"s/SESSION/$sid/g\"\n"+stream+"REPLY\n"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, agent := range []string{"codex", "gemini"} {
		if err := os.WriteFile(filepath.Join(dir, agent), []byte("#!/bin/sh\necho unexpected reviewer agent >&2\nexit 99\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
}

func assertGateUsage(t *testing.T, root, recipe string, issue int) reviewusage.Record {
	t.Helper()
	ctx := context.Background()
	records, coverage, err := loadUsageReviewers(ctx, []string{root}, reviewerTestScope(t, root, issue))
	if err != nil {
		t.Fatal(err)
	}
	var selected []reviewusage.Record
	for _, r := range records {
		if r.Intent.Recipe == recipe {
			selected = append(selected, r)
		}
	}
	if len(selected) == 0 {
		t.Fatalf("production %s dispatch has no usage binding: records=%+v coverage=%+v", recipe, records, coverage)
	}
	r := selected[0]
	if r.Phase != reviewusage.Finished || r.PID <= 0 || r.SessionID == "" {
		t.Fatalf("actual CLI identity/progress lost: %+v", r)
	}
	at, err := time.Parse(time.RFC3339Nano, r.StartedAt)
	if err != nil {
		t.Fatal(err)
	}
	end, err := time.Parse(time.RFC3339Nano, r.FinishedAt)
	if err != nil {
		t.Fatal(err)
	}
	at = at.Add(end.Sub(at) / 2)
	sessions := []transcripts.SessionEvidence{{Harness: r.Intent.Agent, ID: r.SessionID, CWD: r.Intent.Root, Started: at}}
	bindings := bindUsageReviewers(sessions, records, end.Add(time.Second))
	response := usage.Response{ID: "production-review-response", SourceID: r.SessionID, Agent: r.Intent.Agent, Model: "fixture", At: at, Tokens: &usage.Counters{Input: 17, Output: 3}}
	report, err := usage.AggregateUsage([]usage.Response{response}, usage.Window{Since: at.Add(-time.Second), Until: end.Add(time.Second)}, usage.Window{Since: at.Add(-time.Second), Until: end.Add(time.Second)}, func(response usage.Response) usage.Attribution {
		return usageDetailAttribution(response, usage.Attribution{Role: "worker", Slot: "unknown-slot", Include: true}, bindings, usageAuthority{Repository: r.Intent.Repository}, issue, true)
	})
	if err != nil {
		t.Fatal(err)
	}
	parts := usageIssueParts(report)
	if parts[0].Tokens != nil || usageTokenValue(parts[1].Tokens) != 20 || parts[1].Rounds != 1 || parts[1].Attempts != 1 || usageTokenValue(parts[2].Tokens) != 20 {
		t.Fatalf("production binding did not reach issue worker/reviewer/total: %+v records=%+v", parts, records)
	}
	return r
}

func TestProductionBoundaryReviewUsage(t *testing.T) {
	p := boundaryWindow(t)
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	accountingCLI(t, "SHIP")
	p.Context = context.Background()
	p.Agent = "claude"
	p.AgentExplicit = true
	p.PlansDir = "workshop/plans"
	var output bytes.Buffer
	result := dispatchBoundaryReview(&output, &output, p)
	if result.DispatchError != nil || result.Round == nil {
		t.Fatalf("production boundary dispatch: %+v\n%s", result, output.String())
	}
	record := assertGateUsage(t, root, "milestone-review", 69)
	if !record.Intent.DisposableRoot || record.Intent.Root == root {
		t.Fatalf("boundary checkout identity lost: %+v", record)
	}
	sidecar, err := os.ReadFile(result.SidecarPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sidecar), record.Intent.ID) {
		t.Fatal("boundary sidecar omitted real dispatch receipt")
	}
}

func TestProductionPlanningReviewUsage(t *testing.T) {
	repo, _ := syncRepo(t)
	content := "---\nid: 000206\nstatus: working\n---\n## Spec\nDesign\n## Plan\n- [ ] Implement\n## Estimate\nReview this estimate\n"
	writeSyncIssue(t, repo, filepath.Base(issuePath206), content)
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-qm", "review inputs")
	accountingCLI(t, "CLEAN")
	root := buildRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"change-code", "--issue", "206", "--flow", "full", "--no-structural", "--no-estimate", "--no-estimate-recon", "--agent", "claude", "--worktree", "no", "--force", "fixture"})
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("production change-code: %v\n%s", err, out.String())
	}
	plan := assertGateUsage(t, repo, "plan-quality", 206)
	estimate := assertGateUsage(t, repo, "estimate-quality", 206)
	ledger, err := os.ReadFile(filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan-gate.md"))
	if err != nil {
		t.Fatal(err)
	}
	copied := reviewerRecordsInArtifact(string(ledger))
	for _, want := range []reviewusage.Record{plan, estimate} {
		found := false
		for _, record := range copied {
			if record == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("production ledger omitted exact %s receipt: %+v", want.Intent.Recipe, copied)
		}
	}
}

func TestProductionPlanningCancellationRetainsAccountingOnly(t *testing.T) {
	repo, _ := syncRepo(t)
	writeSyncIssue(t, repo, filepath.Base(issuePath206), "---\nid: 000206\nstatus: working\n---\n## Spec\nDesign\n## Plan\n- [ ] Implement\n")
	before := git(t, repo, "rev-parse", "HEAD")
	dir := t.TempDir()
	// The real CLI remains alive until command cancellation kills its group.
	if err := os.WriteFile(filepath.Join(dir, "claude"), []byte("#!/bin/sh\nexec sleep 60\n"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, agent := range []string{"codex", "gemini"} {
		if err := os.WriteFile(filepath.Join(dir, agent), []byte("#!/bin/sh\necho unexpected reviewer agent >&2\nexit 99\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	root := buildRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"change-code", "--issue", "206", "--flow", "full", "--no-structural", "--no-estimate", "--no-estimate-recon", "--agent", "claude", "--worktree", "no", "--force", "cannot waive interruption"})
	done := make(chan error, 1)
	go func() { done <- root.ExecuteContext(ctx) }()
	finished := false
	defer func() {
		cancel()
		if !finished {
			select {
			case <-done:
			case <-time.After(testfix.HangCeiling):
				t.Error("canceled production command was not reaped")
			}
		}
	}()
	var started reviewusage.Record
	testfix.AwaitCondition(t, "the production dispatch persisted its started receipt", func() bool {
		records, _, err := loadUsageReviewers(context.Background(), []string{repo}, reviewerTestScope(t, repo, 206))
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range records {
			if r.Intent.Recipe == "plan-quality" && r.Phase == reviewusage.Started {
				started = r
			}
		}
		return started.Phase == reviewusage.Started
	}, []chan error{done})
	ledger := filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan-gate.md")
	if _, err := os.Stat(ledger); !os.IsNotExist(err) {
		t.Fatalf("launched receipt advanced semantic ledger: %v", err)
	}
	cancel()
	select {
	case err := <-done:
		finished = true
		if err == nil {
			t.Fatalf("cancellation accepted review: %s", out.String())
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatal("production cancellation did not return")
	}
	records, _, err := loadUsageReviewers(context.Background(), []string{repo}, reviewerTestScope(t, repo, 206))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range records {
		if r.Intent.ID == started.Intent.ID {
			found = true
			if r.Phase != reviewusage.Finished || r.Outcome != "canceled" || r.PID != started.PID {
				t.Fatalf("cancellation lost confirmed receipt: %+v", r)
			}
		}
	}
	if !found {
		t.Fatal("interrupted dispatch receipt disappeared")
	}
	if _, err := os.Stat(ledger); !os.IsNotExist(err) {
		t.Fatalf("cancellation advanced semantic ledger: %v", err)
	}
	if after := git(t, repo, "rev-parse", "HEAD"); after != before {
		t.Fatal("canceled accounting committed gate progress")
	}
}
