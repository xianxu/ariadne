package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/observe"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
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

// #279 M2 BR-9: the milestone path end to end through the real gates —
// change-code's plan-quality ledger, a SHIP milestone, and a milestone whose
// review left an open Important finding. Each boundary reads its own verdict
// and its own open blocking findings (the issue-wide ledger, scoped).
func TestObserveMilestonesThroughTheRealGates(t *testing.T) {
	pid := "000396"
	full := "---\nid: " + pid + "\nstatus: open\ndeps: []\ncreated: 2026-09-01\nupdated: 2026-09-01\n---\n\n# milestones\n\n" +
		"## Problem\n\nA gap.\n\n## Spec\n\nA thing.\n\n## Done when\n\n- it works\n\n## Plan\n\n- [ ] M1 — first\n- [ ] M2 — second\n\n## Log\n"
	card, detail, err := issue.SplitCardWithFormat([]byte(full), "sha1")
	if err != nil {
		t.Fatal(err)
	}
	cardPath, detailPath := tracker.CardPath(pid, "ms"), syncIssuesDir+"/"+pid+"-ms.md"
	r := newTrackerRepo(t, map[string]string{cardPath: string(card)}, map[string]string{detailPath: string(detail)})
	run := func(args ...string) error {
		t.Logf("sdlc %v", args)
		var err error
		var stderr string
		msg, died := expectDie(t, func() { _, stderr, err = executeSDLCTestCommand(args...) })
		if died {
			return fmt.Errorf("died: %s", msg)
		}
		if err != nil {
			return fmt.Errorf("%w\n%s", err, stderr)
		}
		return nil
	}
	commitAll := func(msg string) { r.git("add", "-A", "workshop"); r.git("commit", "-qm", msg) }
	for _, args := range [][]string{{"claim", "--issue", "396"}, {"start-plan", "--issue", "396"}} {
		if err := run(args...); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
	if err := run("change-code", "--issue", "396", "--worktree=no", "--flow", "full", "--no-judge", "--no-structural", "--no-estimate", "--no-estimate-recon"); err != nil {
		t.Fatalf("change-code: %v", err)
	}
	// change-code's plan-quality dispatch exits the process (no in-process
	// path, #191), so its ledger is written with the gate's own writer: a
	// finding raised, then addressed. (An undisposed one would be inherited by
	// every boundary, as the gate seeds it — real behavior, not this test's.)
	planLedger := gatestate.Ledger{Gate: "plan-quality", IssueNum: 396, IDPrefix: "PQ", Rounds: []gatestate.Round{
		{N: 1, Timestamp: "2026-10-02T09:00:00-07:00", Agent: "claude",
			New: []gatestate.Finding{{ID: "PQ-1", Severity: "Important", Title: "a plan gap", Family: "plan-gap", Round: 1}}},
		{N: 2, Timestamp: "2026-10-02T09:10:00-07:00", Agent: "claude",
			Dispositions: []gatestate.Disposition{{ID: "PQ-1", State: "addressed", Round: 2}}},
	}}
	if err := writePlanGateLedger(filepath.Join(r.root, "workshop/plans"), pid+"-ms.md", planLedger, "ariadne"); err != nil {
		t.Fatal(err)
	}
	writeRepoFile(t, r.root, "cmd/a.go", "package a\n")
	commitAll("#396: work")
	r.git("add", "cmd/a.go")
	r.git("commit", "-qm", "#396 M1: implement")
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\n```findings\nfindings: []\n```\n")
	if err := run("milestone-close", "--issue", "396", "--milestone", "M1", "--verified", "e2e", "--actual", "0.1", "--no-atlas", "--no-project"); err != nil {
		t.Fatalf("M1: %v", err)
	}
	commitAll("#396 M1: milestone close")
	writeRepoFile(t, r.root, "cmd/b.go", "package a\n")
	r.git("add", "cmd/b.go")
	r.git("commit", "-qm", "#396 M2: implement")
	stubJudge(t, "VERDICT: FIX-THEN-SHIP (confidence: high)\n\n```findings\nfindings:\n  - id: new\n    severity: Important\n    family: test-family\n    title: |\n      a thing to fix\n    detail: |\n      fix it\n```\n")
	_ = run("milestone-close", "--issue", "396", "--milestone", "M2", "--verified", "e2e", "--actual", "0.1", "--no-atlas", "--no-project") // refused: open blocking
	commitAll("#396 M2: review recorded")

	other := filepath.Join(t.TempDir(), "other")
	testfix.Git(t, r.root, "worktree", "add", "-q", "--detach", other, "origin/main")
	t.Chdir(other)
	o := observeIssue(t, 396, "")
	plan, okPlan := reviewOf(o, "plan")
	m1, okM1 := reviewOf(o, "M1")
	m2, okM2 := reviewOf(o, "M2")
	if !okPlan || plan.State != observe.Present || plan.OpenBlocking != 0 {
		t.Errorf("plan boundary: %+v", plan)
	}
	if !okM1 || m1.Verdict != "SHIP" || m1.OpenBlocking != 0 {
		t.Errorf("M1 (scoped away from M2's open finding): %+v", m1)
	}
	if !okM2 || m2.Verdict != "FIX-THEN-SHIP" || m2.OpenBlocking != 1 {
		t.Errorf("M2: %+v", m2)
	}
	if _, reached := reviewOf(o, "close"); reached {
		t.Errorf("an unclosed issue lists a close review: %+v", o.Checkpoints.Reviews)
	}
	if o.Checkpoints.Plan.Ticked != 1 || o.Checkpoints.Flow == nil || o.Checkpoints.Flow.Kind != "full" {
		t.Errorf("plan/flow: %+v %+v", o.Checkpoints.Plan, o.Checkpoints.Flow)
	}
}

// #279 M2 BR-10: the evidence is read at the issues dir the caller gave, not a
// default — a repository keeping details elsewhere is still observed.
func TestObserveEvidenceFollowsTheGivenIssuesDir(t *testing.T) {
	r, _, detailPath := closeReady(t, 397)
	alt := "notes/issues/" + filepath.Base(detailPath)
	writeRepoFile(t, r.root, alt, string(must(os.ReadFile(filepath.Join(r.root, detailPath)))))
	r.git("add", alt)
	r.git("commit", "-qm", "#397: details also kept elsewhere")
	ev := collectEvidence(r.root, "notes/issues", "origin", strings.TrimSuffix(filepath.Base(detailPath), ".md"), "working",
		observe.BranchFacts{Ref: "refs/heads/" + r.git("branch", "--show-current")})
	if ev.Details == nil || ev.DetailsErr != nil {
		t.Fatalf("details at the given dir not read: %+v", ev)
	}
}

// #279 M2 minor: --repo observes an issue of a second tracker repository from
// inside the first.
func TestObserveAcrossTrackerRepositories(t *testing.T) {
	cardA, a, detailA, dA := seededIssue(t, "000398", "alpha")
	ra := newTrackerRepo(t, map[string]string{cardA: a}, map[string]string{detailA: dA})
	cardB, b, detailB, dB := seededIssue(t, "000398", "bravo")
	rb := newTrackerRepo(t, map[string]string{cardB: b}, map[string]string{detailB: dB})
	t.Chdir(ra.root)
	o := observeIssue(t, 398, rb.root)
	if !strings.Contains(o.Card.Source, "bravo") || o.Card.Title == "" {
		t.Fatalf("--repo observed the wrong repository: %+v", o.Card)
	}
	if o := observeIssue(t, 398, ""); !strings.Contains(o.Card.Source, "alpha") {
		t.Fatalf("without --repo: %+v", o.Card)
	}
}
