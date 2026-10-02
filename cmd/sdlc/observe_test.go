package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/observe"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// observeIssue runs `sdlc issue show N --json` in-process from the current
// directory and decodes it strictly.
func observeIssue(t *testing.T, n int, repo string) observe.Observation {
	t.Helper()
	var out, errs bytes.Buffer
	if err := runIssueShow(context.Background(), &out, &errs, &issueShowFlags{IssuesDir: "workshop/issues", JSON: true, Repo: repo}, itoa(n)); err != nil {
		t.Fatalf("issue show --json: %v\n%s", err, errs.String())
	}
	var o observe.Observation
	if err := json.Unmarshal(out.Bytes(), &o); err != nil {
		t.Fatalf("observation is not valid contract JSON: %v\n%s", err, out.String())
	}
	return o
}

// localState is what an observation must never change: worktree status and
// every local branch ref.
func localState(t *testing.T, roots ...string) string {
	t.Helper()
	var b strings.Builder
	for _, r := range roots {
		b.WriteString(testfix.Capture(t, r, "status", "--porcelain", "--untracked-files=all"))
		b.WriteString(testfix.Capture(t, r, "for-each-ref", "refs/heads", "refs/sdlc"))
	}
	return b.String()
}

// #279: an issue claimed in a parked slot (no agent, uncommitted work) is
// observed from the primary checkout: its owner, that the owner's worktree
// holds the branch, and that the query changed nothing local.
func TestObserveAParkedSlotFromAnotherCheckout(t *testing.T) {
	r, slot, cardPath, _ := reclaimFixture(t, 390)
	writeRepoFile(t, slot, "wip.txt", "parked, uncommitted\n")
	before := localState(t, r.root, slot)
	o := observeIssue(t, 390, "")
	owner, _ := ownerOf(t, r, cardPath)
	if o.Tracker.State != observe.Present || o.Card.Status != "working" || o.Card.Revision == "" {
		t.Fatalf("card read: %+v %+v", o.Tracker, o.Card)
	}
	if o.Assignment.Claimant == nil || o.Assignment.Claimant.Worktree != owner.Worktree ||
		o.Assignment.Relation != observe.RelationOtherWorkspace || o.Assignment.ClaimantWorktree != observe.FateHoldsBranch {
		t.Fatalf("assignment: %+v", o.Assignment)
	}
	if o.Completion.State != observe.Absent || o.Landing.Outcome != observe.OutcomeNotLanded {
		t.Fatalf("completion/landing: %+v %+v", o.Completion, o.Landing)
	}
	if after := localState(t, r.root, slot); after != before {
		t.Fatalf("an observation changed local state:\nbefore %q\nafter  %q", before, after)
	}
	// The same issue from the slot itself is this workspace's.
	t.Chdir(slot)
	if o := observeIssue(t, 390, ""); o.Assignment.Relation != observe.RelationThisWorkspace {
		t.Fatalf("from the owner's own checkout: %+v", o.Assignment)
	}
}

// #279: with the tracker unreachable the answer comes from the last fetch,
// marked stale with the reason — never absent.
func TestObserveStaleTrackerSaysSo(t *testing.T) {
	r, _, _, _ := reclaimFixture(t, 391)
	if err := os.Rename(r.origin, r.origin+".gone"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Rename(r.origin+".gone", r.origin) })
	o := observeIssue(t, 391, "")
	if o.Tracker.State != observe.Stale || o.Tracker.Error == "" || o.Card.State != observe.Stale || o.Card.Status != "working" {
		t.Fatalf("stale read: %+v %+v", o.Tracker, o.Card)
	}
	if o.Assignment.State != observe.Stale || o.Assignment.Relation != observe.RelationOtherWorkspace {
		t.Fatalf("stale assignment: %+v", o.Assignment)
	}
}

// #279: a landed issue reads as landed at its landed commit, with its close
// evidence; another repository's issue is observed with --repo.
func TestObserveLandedAndAnotherRepository(t *testing.T) {
	r, _, _ := closedAndLanded(t, 392)
	if done, err := publishCodecompleteIssues(context.Background(), "workshop/issues"); err != nil || len(done) != 1 {
		t.Fatalf("flip: %v %v", done, err)
	}
	o := observeIssue(t, 392, "")
	if o.Landing.Outcome != observe.OutcomeLanded || o.Landing.LandedCommit == "" || o.Completion.EvidenceCommit == "" || o.Card.Status != "done" {
		t.Fatalf("landed: %+v %+v", o.Landing, o.Completion)
	}
	elsewhere := t.TempDir()
	t.Chdir(elsewhere)
	if o := observeIssue(t, 392, filepath.Join(r.root, "workshop")); o.Landing.Outcome != observe.OutcomeLanded {
		t.Fatalf("--repo: %+v", o.Landing)
	}
}

// #279 M1 BR-1: the observed repository is the one containing the issues dir,
// never the process cwd. An issues dir outside any repository still shows (its
// details are the record) with no tracker read; one given explicitly while
// standing in a tracked repository observes that dir, not the repository's
// same-numbered issue — and fetches nothing.
func TestObserveAnchorsOnTheIssuesDir(t *testing.T) {
	loose := t.TempDir()
	writeRepoFile(t, loose, "workshop/issues/000393-loose.md", "---\nid: 000393\nstatus: open\n---\n\n# Loose\n\n## Problem\n\nx\n")
	show := func(dir string, json bool) (string, error) {
		var out, errs bytes.Buffer
		err := runIssueShow(context.Background(), &out, &errs, &issueShowFlags{IssuesDir: dir, JSON: json}, "393")
		return out.String() + errs.String(), err
	}
	t.Chdir(t.TempDir()) // not a repository
	if out, err := show(filepath.Join(loose, "workshop/issues"), false); err != nil || !strings.Contains(out, "# Loose") || !strings.Contains(out, "observations") {
		t.Fatalf("outside a repository: %v\n%s", err, out)
	}
	r, _, cardPath, _ := reclaimFixture(t, 393) // a tracked repository with its own #393
	tip := trackerTip(t, r)
	remoteRef := r.git("for-each-ref", "refs/remotes")
	out, err := show(filepath.Join(loose, "workshop/issues"), true)
	if err != nil {
		t.Fatal(err)
	}
	var o observe.Observation
	if err := json.Unmarshal([]byte(out), &o); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if o.Tracker.State != observe.Absent || o.Card.State != observe.Absent || strings.Contains(out, cardPath) {
		t.Fatalf("observed the cwd's repository instead of the given issues dir: %+v %+v", o.Tracker, o.Card)
	}
	if trackerTip(t, r) != tip || r.git("for-each-ref", "refs/remotes") != remoteRef {
		t.Fatal("an observation of a loose issues dir touched the cwd repository")
	}
}

func reviewOf(o observe.Observation, boundary string) (observe.Review, bool) {
	for _, r := range o.Checkpoints.Reviews {
		if r.Boundary == boundary {
			return r, true
		}
	}
	return observe.Review{}, false
}

// #279: one issue through its lifecycle, always observed from a different
// checkout. In progress (flow, no reviews), closed (SHIP from its artifact), its
// close artifact lost (unknown — a close the card records), and landed by a
// squash merge with the branch deleted (the verdict read from main's archive).
func TestObserveCheckpointsAcrossTheLifecycle(t *testing.T) {
	r, cardPath, detailPath := closeReady(t, 394)
	branch := r.git("branch", "--show-current")
	other := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, r.root, "worktree", "add", "-q", "--detach", other, "origin/main")
	look := func() observe.Observation { t.Helper(); t.Chdir(other); defer t.Chdir(r.root); return observeIssue(t, 394, "") }

	o := look()
	if o.Branch.Ref != "refs/heads/"+branch || o.Checkpoints.State != observe.Present || o.Checkpoints.Flow == nil || len(o.Checkpoints.Reviews) != 0 {
		t.Fatalf("in progress: %+v %+v", o.Branch, o.Checkpoints)
	}
	if len(o.Workspaces.Holding) != 1 || o.Workspaces.Holding[0].Path != canonRoot(r.root) || !o.Workspaces.Holding[0].IsClaimant {
		t.Fatalf("holding: %+v", o.Workspaces)
	}

	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "394", "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	o = look()
	if rv, ok := reviewOf(o, "close"); !ok || rv.Verdict != "SHIP" || rv.State != observe.Present || o.Completion.State != observe.Present {
		t.Fatalf("closed: %+v %+v", o.Checkpoints.Reviews, o.Completion)
	}

	review := "workshop/plans/" + strings.TrimSuffix(filepath.Base(detailPath), ".md") + "-close-review.md"
	r.git("rm", "-q", review)
	r.git("commit", "-qm", "#394: lose the close artifact")
	if rv, _ := reviewOf(look(), "close"); rv.State != observe.Unknown {
		t.Fatalf("a recorded close without its artifact: %+v", rv)
	}
	r.git("revert", "--no-edit", "HEAD")

	// Land by squash, delete the branch, complete the card, archive on main.
	card := r.card(cardPath)
	binding, _, _ := issue.CardCompletion([]byte(card))
	r.git("switch", "-q", "main")
	r.git("merge", "-q", "--squash", branch)
	r.git("commit", "-qm", "#394: squash landing")
	r.git("push", "-q", "origin", "main")
	r.git("branch", "-q", "-D", branch)
	landed := r.git("rev-parse", "HEAD")
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.ChangeCard("000394", cardPath, "done", operationToken("done"), func(c []byte) ([]byte, error) {
		return doneCard(c, binding.Token, landed, "2026-10-02")
	}); err != nil {
		t.Fatal(err)
	}
	invalidateIssueRecords(context.Background())
	var stderr bytes.Buffer
	if _, err := archiveDoneIssues(context.Background(), &stderr, "", "workshop/issues", "workshop/history", "workshop/plans"); err != nil {
		t.Fatalf("archive: %v\n%s", err, stderr.String())
	}
	r.git("add", "-A", "workshop")
	r.git("commit", "-qm", "archive completed issues to history")
	r.git("push", "-q", "origin", "main")
	o = look()
	if o.Landing.Outcome != observe.OutcomeLanded || o.Landing.Archived == "" || o.Branch.State != observe.Absent {
		t.Fatalf("landed: %+v %+v", o.Landing, o.Branch)
	}
	if rv, ok := reviewOf(o, "close"); !ok || rv.Verdict != "SHIP" || !strings.Contains(o.Checkpoints.Source, "history/plans") {
		t.Fatalf("verdict after a squash landing: %+v from %s", rv, o.Checkpoints.Source)
	}
}

// #279: where the recorded owner's worktree stands — elsewhere while a third
// worktree holds the branch, missing once removed, another machine never
// probed — and that one query's git work is bounded by the issue, not the fleet.
func TestObserveWorktreeFatesAndBound(t *testing.T) {
	r, slot, cardPath, _ := reclaimFixture(t, 395)
	testfix.Git(t, slot, "switch", "-q", "--detach")
	third := filepath.Join(t.TempDir(), "third")
	testfix.Git(t, r.root, "worktree", "add", "-q", third, "000395-reclaim")
	calls := 0
	prev := observeGit
	observeGit = func(dir string, args ...string) ([]byte, error) { calls++; return prev(dir, args...) }
	o := observeIssue(t, 395, "")
	observeGit = prev
	if o.Assignment.ClaimantWorktree != observe.FateElsewhere || len(o.Workspaces.Holding) != 1 ||
		o.Workspaces.Holding[0].Path != canonRoot(third) || o.Workspaces.Holding[0].IsClaimant {
		t.Fatalf("elsewhere: %+v %+v", o.Assignment, o.Workspaces)
	}
	if calls > 20 {
		t.Fatalf("one observation ran %d git commands; the bound is per issue (holding worktrees + artifacts)", calls)
	}
	testfix.Git(t, r.root, "worktree", "remove", "--force", slot)
	if o := observeIssue(t, 395, ""); o.Assignment.ClaimantWorktree != observe.FateMissing {
		t.Fatalf("removed: %+v", o.Assignment)
	}
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	owner, _ := ownerOf(t, r, cardPath)
	owner.Machine = issue.MachineFingerprint("another-machine")
	if err := env.repo.ChangeCard("000395", cardPath, "owner", operationToken("set"), func(c []byte) ([]byte, error) {
		return issue.SetCardClaimant(c, owner)
	}); err != nil {
		t.Fatal(err)
	}
	invalidateIssueRecords(context.Background())
	if o := observeIssue(t, 395, ""); o.Assignment.ClaimantWorktree != observe.FateOtherMachine {
		t.Fatalf("another machine: %+v", o.Assignment)
	}
}
