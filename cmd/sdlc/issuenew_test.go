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
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func newIssueFlags() *issueNewFlags {
	return &issueNewFlags{IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
}

// mainFile reads a file from origin's main ("" when absent).
func (r *trackerRepo) mainFile(path string) string {
	r.t.Helper()
	out, err := exec.Command("git", "-C", r.origin, "show", "main:"+path).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

// assertCheckoutUntouched fails when filing changed the checkout: HEAD, the
// index or the working tree (untracked files included).
func assertCheckoutUntouched(t *testing.T, r *trackerRepo, head, status string) {
	t.Helper()
	if now := r.git("rev-parse", "HEAD"); now != head {
		t.Fatalf("filing moved HEAD %s → %s", head, now)
	}
	if now := r.git("status", "--porcelain", "--untracked-files=all"); now != status {
		t.Fatalf("filing wrote the checkout:\n%s\n(was:\n%s)", now, status)
	}
}

// slotClaimant is a workspace identity in slot repo:n of this checkout.
func slotClaimant(t *testing.T, slot string) {
	t.Helper()
	prev := claimantIdentity
	claimantIdentity = func(env *trackerEnv) (issue.Claimant, error) {
		return issue.Claimant{Operator: "Op", Machine: issue.MachineFingerprint("this-machine"), MachineName: "box",
			Worktree: canonRoot(env.root), Repository: env.target.Repository, Workspace: slot}, nil
	}
	t.Cleanup(func() { claimantIdentity = prev })
}

// From the resting branch of :0 the issue is filed on origin, unclaimed, and
// the checkout is left exactly as it was (#308 Done-when 2).
func TestIssueNewPublishesUnclaimedFromRestingZero(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	slotClaimant(t, "ariadne:0")
	mainBefore, head, status := r.originMain(), r.git("rev-parse", "HEAD"), r.git("status", "--porcelain", "--untracked-files=all")
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, newIssueFlags(), []string{"Lift the Subsystem!"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	want := "workshop/issues/000008-lift-the-subsystem.md"
	if strings.TrimSpace(out.String()) != want {
		t.Fatalf("stdout %q, want the details' path on main", out.String())
	}
	card := r.card("workshop/issue-cards/000008-lift-the-subsystem.md")
	parsed, err := issue.ParseCard([]byte(card))
	if err != nil || parsed.ID != "000008" || parsed.Title != "Lift the Subsystem!" {
		t.Fatalf("card %q: %v", card, err)
	}
	if _, claimed, err := issue.CardClaimant([]byte(card)); err != nil || claimed {
		t.Fatalf("filed from :0 but claimed (%v):\n%s", err, card)
	}
	detail := r.mainFile(want)
	for _, section := range []string{"## Spec", "## Plan", "## Log", issue.MirrorField} {
		if !strings.Contains(detail, section) {
			t.Errorf("details on main missing %s", section)
		}
	}
	if r.originMain() == mainBefore {
		t.Fatal("details not published to main")
	}
	if msg := strings.TrimSpace(testfix.Capture(t, r.origin, "log", "-1", "--format=%s", "main")); msg != "#8: issue: new" {
		t.Errorf("main commit subject %q", msg)
	}
	assertCheckoutUntouched(t, r, head, status)
}

// From a numbered slot the creating slot is the claimant, by a claim-kind
// tracker commit (so claim ages and active-time windows read it as a claim);
// the checkout is untouched (#308 Done-when 1).
func TestIssueNewFromSlotClaimsAndLeavesTreeClean(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	slotClaimant(t, "pair:1")
	head, status := r.git("rev-parse", "HEAD"), r.git("status", "--porcelain", "--untracked-files=all")
	var out, errs bytes.Buffer
	if err := runIssueNew(context.Background(), &out, &errs, newIssueFlags(), []string{"Claimed At Birth"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	cardPath := "workshop/issue-cards/000008-claimed-at-birth.md"
	c, claimed, err := issue.CardClaimant([]byte(r.card(cardPath)))
	if err != nil || !claimed || c.Workspace != "pair:1" {
		t.Fatalf("claimant %+v claimed=%v err=%v", c, claimed, err)
	}
	if status, _ := issue.GetField(r.card(cardPath), "status"); status != "open" {
		t.Errorf("claim at creation changed status to %q", status)
	}
	log := testfix.Capture(t, r.origin, "log", "-1", "--format=%B", "issue-tracker", "--", cardPath)
	if !strings.Contains(log, "Tracker-Operation: "+tracker.OpClaim+"-") {
		t.Fatalf("the claim is not a claim-kind tracker commit:\n%s", log)
	}
	if !strings.Contains(log, "\nAgent: "+testfix.FixedAgent) { // #367
		t.Fatalf("the tracker commit lacks the Agent trailer:\n%s", log)
	}
	if r.mainFile("workshop/issues/000008-claimed-at-birth.md") == "" {
		t.Fatal("details not on main")
	}
	assertCheckoutUntouched(t, r, head, status)
}

// On an issue branch nothing is committed there (TL 10-12, option a): the
// details reach the branch with its next main merge.
func TestIssueNewOnIssueBranchDoesNotCommitThere(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	slotClaimant(t, "ariadne:4")
	r.git("switch", "-q", "-c", "000007-seven")
	writeRepoFile(t, r.root, "code.go", "package x\n")
	r.git("add", "code.go") // unrelated staged work stays exactly as it was
	head, status := r.git("rev-parse", "HEAD"), r.git("status", "--porcelain", "--untracked-files=all")
	var out, errs bytes.Buffer
	f := newIssueFlags()
	f.Target, f.Deps = "tgt", []string{"repo#1"}
	if err := runIssueNew(context.Background(), &out, &errs, f, []string{"Spin Off"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	detail := r.mainFile("workshop/issues/000008-spin-off.md")
	if !strings.Contains(detail, "target: tgt") || !strings.Contains(detail, "repo#1") {
		t.Errorf("details on main lost branch-owned fields:\n%s", detail)
	}
	if card := r.card("workshop/issue-cards/000008-spin-off.md"); strings.Contains(card, "target:") || strings.Contains(card, "deps:") {
		t.Errorf("card carries branch-owned fields:\n%s", card)
	}
	assertCheckoutUntouched(t, r, head, status)
}

// C5: a directory with its own issues tree under an enclosing repository (a
// home-directory repository) refuses before anything is filed.
func TestIssueNewRefusesNestedIssuesTreeUnderAnotherRepo(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	nested := filepath.Join(r.root, "projects", "emma")
	if err := os.MkdirAll(filepath.Join(nested, "workshop", "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	chdirTo(t, nested)
	mainBefore := r.originMain()
	var out, errs bytes.Buffer
	err := runIssueNew(context.Background(), &out, &errs, newIssueFlags(), []string{"Wrong Repo"})
	if err == nil || !strings.Contains(err.Error(), "not a repository root") {
		t.Fatalf("nested issues tree: %v", err)
	}
	if r.card("workshop/issue-cards/000008-wrong-repo.md") != "" || r.originMain() != mainBefore {
		t.Fatal("filed into the enclosing repository")
	}
}

// C5 in a legacy repository: filing writes the checkout there, so a toplevel
// with no issues directory (a home-directory repository) refuses rather than
// creating one.
func TestLegacyIssueNewRefusesRepoWithoutIssuesDir(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	var out, errs bytes.Buffer
	err := runIssueNew(context.Background(), &out, &errs, newIssueFlags(), []string{"Home Repo"})
	if err == nil || !strings.Contains(err.Error(), "has no workshop/issues, and issue new will not create it") {
		t.Fatalf("home repository: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "workshop")); !os.IsNotExist(err) {
		t.Fatal("created an issues tree in the home repository")
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
		t.Fatal("reserved a card without a branch to bind its recovery record to")
	}
}
