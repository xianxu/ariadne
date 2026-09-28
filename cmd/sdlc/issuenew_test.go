package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func newIssueFlags() *issueNewFlags {
	return &issueNewFlags{IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
}

func TestIssueNewReservesCardAndLeavesRestUnpublished(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	mainBefore, headBefore := r.originMain(), r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, newIssueFlags(), []string{"Lift the Subsystem!"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	want := "workshop/issues/000008-lift-the-subsystem.md"
	if strings.TrimSpace(out.String()) != want {
		t.Fatalf("stdout %q, want the created path", out.String())
	}
	card := r.card("workshop/issue-cards/000008-lift-the-subsystem.md")
	parsed, err := issue.ParseCard([]byte(card))
	if err != nil || parsed.ID != "000008" || parsed.Title != "Lift the Subsystem!" {
		t.Fatalf("card %q: %v", card, err)
	}
	detail, err := os.ReadFile(filepath.Join(r.root, want))
	if err != nil {
		t.Fatal(err)
	}
	for _, section := range []string{"## Spec", "## Plan", "## Log", issue.MirrorField} {
		if !strings.Contains(string(detail), section) {
			t.Errorf("details missing %s", section)
		}
	}
	// On the resting branch: nothing published to main, no checkpoint commit.
	if r.originMain() != mainBefore || r.git("rev-parse", "HEAD") != headBefore {
		t.Fatal("card-only filing moved main or committed on rest")
	}
	if !strings.Contains(errs.String(), "move-detail --issue 8`") {
		t.Errorf("no handoff next action:\n%s", errs.String())
	}
}

func TestIssueNewOnFeatureBranchCommitsOnlyItsDetails(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	r.git("switch", "-q", "-c", "000007-seven")
	writeRepoFile(t, r.root, "code.go", "package x\n")
	r.git("add", "code.go") // unrelated staged work must stay staged
	mainBefore := r.originMain()
	var out, errs bytes.Buffer
	f := newIssueFlags()
	f.Target, f.Deps = "tgt", []string{"repo#1"}
	if err := runIssueNew(context.Background(), &out, &errs, f, []string{"Spin Off"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if files := r.git("show", "--name-only", "--format=", "HEAD"); files != "workshop/issues/000008-spin-off.md" {
		t.Fatalf("commit carried %q", files)
	}
	if staged := r.git("diff", "--cached", "--name-only"); staged != "code.go" {
		t.Fatalf("unrelated staged work disturbed: %q", staged)
	}
	detail := r.git("show", "HEAD:workshop/issues/000008-spin-off.md")
	if !strings.Contains(detail, "target: tgt") || !strings.Contains(detail, "repo#1") {
		t.Errorf("details lost branch-owned fields:\n%s", detail)
	}
	if card := r.card("workshop/issue-cards/000008-spin-off.md"); strings.Contains(card, "target:") || strings.Contains(card, "deps:") {
		t.Errorf("card carries branch-owned fields:\n%s", card)
	}
	if r.originMain() != mainBefore {
		t.Fatal("spin-off published to main")
	}
}

func TestIssueNewDryRunReservesNothing(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	var out, errs bytes.Buffer
	f := newIssueFlags()
	f.DryRun = true
	if err := runIssueNew(context.Background(), &out, &errs, f, []string{"Some title"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if !strings.Contains(out.String(), "000008-some-title.md") {
		t.Fatalf("dry run output:\n%s", out.String())
	}
	if r.card("workshop/issue-cards/000008-some-title.md") != "" {
		t.Fatal("dry run reserved a card")
	}
	if _, err := os.Stat(filepath.Join(r.root, "workshop/issues")); !os.IsNotExist(err) {
		t.Fatal("dry run wrote details")
	}
}

func TestIssueNewFromGitHubLinksCard(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	prevGH, prevRepo := ghClient, detectRepo
	ghClient = stubGH{title: "Fix the thing", body: "GH body text."}
	detectRepo = func() (string, error) { return "xianxu/ariadne", nil }
	t.Cleanup(func() { ghClient, detectRepo = prevGH, prevRepo })
	var out, errs bytes.Buffer
	f := newIssueFlags()
	f.FromGitHub = 99
	if err := runIssueNew(context.Background(), &out, &errs, f, nil); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	card := r.card("workshop/issue-cards/000008-fix-the-thing.md")
	if !strings.Contains(card, "github_issue: 99") || !strings.Contains(card, "GH body text.") {
		t.Fatalf("card lost GitHub linkage or original report:\n%s", card)
	}
}

func TestIssueNewRefusesDetachedHead(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	r.git("switch", "-q", "--detach")
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, newIssueFlags(), []string{"Detached"}); err == nil || !strings.Contains(err.Error(), "check out a branch") {
		t.Fatalf("detached: %v", err)
	}
	if r.card("workshop/issue-cards/000008-detached.md") != "" {
		t.Fatal("reserved a card it could not materialize")
	}
}
