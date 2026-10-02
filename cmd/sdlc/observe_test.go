package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
