package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func trackerChangeCodeFlags() *changeCodeFlags {
	return &changeCodeFlags{Issue: 9, IssuesDir: "workshop/issues", PlansDir: "workshop/plans", Worktree: "no",
		NoStructural: true, NoJudge: true, NoEstimate: true, NoEstimateRecon: true}
}

// claimedOnBranch seeds #9, claims it and prepares its branch, as the verbs do.
func claimedOnBranch(t *testing.T) (*trackerRepo, string, string) {
	t.Helper()
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatal(err)
	}
	if err := startPlanBranch(context.Background(), &out, 9); err != nil {
		t.Fatal(err)
	}
	return r, cardPath, detailPath
}

func TestChangeCodeRefusesTrackerIssueOnRest(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	head := r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	err := runChangeCode(strings.NewReader(""), &out, &errs, trackerChangeCodeFlags())
	if err == nil || !strings.Contains(err.Error(), "start-plan --issue 9") {
		t.Fatalf("change-code on rest: %v", err)
	}
	if r.git("rev-parse", "HEAD") != head || r.git("status", "--porcelain") != "" {
		t.Fatal("refusal changed the resting checkout")
	}
}

// #283: an owned card that start-plan never started refuses toward start-plan,
// even from the issue's branch (made by hand, or left by a lost start write).
func TestChangeCodeRefusesUnstartedClaim(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatal(err)
	}
	r.git("switch", "-q", "-c", "000009-nine")
	err := runChangeCode(strings.NewReader(""), &out, &errs, trackerChangeCodeFlags())
	if err == nil || !strings.Contains(err.Error(), "not started") || !strings.Contains(err.Error(), "start-plan --issue 9") {
		t.Fatalf("change-code on an unstarted claim: %v", err)
	}
}

func TestChangeCodeRefreshesCardFieldsAndCheckpointsLocally(t *testing.T) {
	r, cardPath, detailPath := claimedOnBranch(t)
	// The estimate is set on the card (as `issue set-estimate` does), not in details.
	repo, err := tracker.NewRepository(context.Background(), r.root, "origin")
	if err != nil {
		t.Fatal(err)
	}
	snap, err := repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	current, _ := snap.Card("000009")
	fm, body, _ := issue.Parse(string(current.Raw))
	withEstimate := issue.Compose(issue.SetField(fm, "estimate_hours", "3.5"), body)
	if err := repo.UpdateCard(current, []byte(withEstimate), "test-estimate", func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, r.root, "workshop/plans/000009-nine-plan.md", "# plan\n")
	mainBefore, cardBefore := r.originMain(), r.card(cardPath)
	var out, errs bytes.Buffer
	if err := runChangeCode(strings.NewReader(""), &out, &errs, trackerChangeCodeFlags()); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if r.originMain() != mainBefore || r.card(cardPath) != cardBefore {
		t.Fatal("change-code published")
	}
	if got := r.git("log", "-1", "--format=%s"); got != "#9: plan: accepted design at change-code" {
		t.Fatalf("checkpoint subject %q", got)
	}
	files := r.git("show", "--name-only", "--format=", "HEAD")
	if !strings.Contains(files, detailPath) || !strings.Contains(files, "workshop/plans/000009-nine-plan.md") {
		t.Fatalf("checkpoint carried %q", files)
	}
	if !strings.Contains(r.git("show", "HEAD:"+detailPath), "estimate_hours: 3.5") {
		t.Fatal("gates did not see the card's estimate")
	}
	if r.git("branch", "--show-current") != "000009-nine" {
		t.Fatal("change-code left the issue branch")
	}
}

func TestChangeCodeRefusesHandEditedCardField(t *testing.T) {
	r, _, detailPath := claimedOnBranch(t)
	raw, err := os.ReadFile(filepath.Join(r.root, detailPath))
	if err != nil {
		t.Fatal(err)
	}
	fm, body, _ := issue.Parse(string(raw))
	edited := issue.Compose(issue.SetField(fm, "estimate_hours", "9"), body)
	writeRepoFile(t, r.root, detailPath, edited)
	head := r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	err = runChangeCode(strings.NewReader(""), &out, &errs, trackerChangeCodeFlags())
	if err == nil || !strings.Contains(err.Error(), "owned by the card") || !strings.Contains(err.Error(), "set-estimate") {
		t.Fatalf("hand-edited estimate: %v", err)
	}
	if r.git("rev-parse", "HEAD") != head {
		t.Fatal("refused change-code committed")
	}
}

// Regression: malformed mirrored details (a duplicated card key) once read as
// "pre-tracker" and skipped every ownership check.
func TestChangeCodeRefusesMalformedMirroredDetails(t *testing.T) {
	r, _, detailPath := claimedOnBranch(t)
	raw, err := os.ReadFile(filepath.Join(r.root, detailPath))
	if err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, r.root, detailPath, strings.Replace(string(raw), "status: working", "status: working\nstatus: done", 1))
	head := r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	if err := runChangeCode(strings.NewReader(""), &out, &errs, trackerChangeCodeFlags()); err == nil {
		t.Fatal("malformed mirrored details passed change-code")
	}
	if r.git("rev-parse", "HEAD") != head {
		t.Fatal("refused change-code committed")
	}
}

func TestChangeCodeDryRunLeavesStaleMirrorOnDisk(t *testing.T) {
	r, _, detailPath := claimedOnBranch(t)
	retitleElsewhere(t, r, "000009", "Renamed Elsewhere")
	abs := filepath.Join(r.root, detailPath)
	before, _ := os.ReadFile(abs)
	f := trackerChangeCodeFlags()
	f.DryRun = true
	var out, errs bytes.Buffer
	if err := runChangeCode(strings.NewReader(""), &out, &errs, f); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if got, _ := os.ReadFile(abs); !bytes.Equal(got, before) {
		t.Fatal("dry run rewrote the details")
	}
}

// BR-13: a gate refusal (the estimate gate here, with no estimate on the card)
// happens after the mirror is refreshed in memory; the file must be untouched.
func TestChangeCodeGateRefusalLeavesStaleMirrorOnDisk(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	r, _, detailPath := claimedOnBranch(t)
	retitleElsewhere(t, r, "000009", "Renamed Elsewhere")
	abs := filepath.Join(r.root, detailPath)
	before, _ := os.ReadFile(abs)
	head := r.git("rev-parse", "HEAD")
	cmd := exec.Command(binary, "change-code", "--issue", "9", "--worktree", "no", "--no-judge", "--no-structural", "--no-estimate-recon", "--flow", "full")
	cmd.Dir = r.root
	if out, err := cmd.CombinedOutput(); err == nil || !strings.Contains(string(out), "estimate gate failed") {
		t.Fatalf("expected the estimate gate to refuse: %v\n%s", err, out)
	}
	if got, _ := os.ReadFile(abs); !bytes.Equal(got, before) {
		t.Fatal("refused change-code rewrote the details")
	}
	if r.git("rev-parse", "HEAD") != head {
		t.Fatal("refused change-code committed")
	}
}
