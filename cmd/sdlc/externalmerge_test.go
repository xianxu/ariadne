package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// #287: every combination of the facts. Landed evidence settles, whoever
// owns the card; a merged branch without it needs a close from the owner, or
// a claim first; anything else is not an outside merge.
func TestExternalMergeVerdict(t *testing.T) {
	for _, evidence := range []bool{false, true} {
		for _, merged := range []bool{false, true} {
			for _, own := range []issue.Ownership{issue.OwnershipMine, issue.OwnershipForeign, issue.OwnershipUnknown} {
				f := mergeFacts{EvidenceOnMain: evidence, BranchMerged: merged, Owner: own}
				want := mergeNone
				switch {
				case evidence:
					want = mergeSettle
				case merged && own == issue.OwnershipUnknown:
					want = mergeClaimNeeded
				case merged:
					want = mergeCloseNeeded
				}
				if got := externalMergeVerdict(f); got != want {
					t.Errorf("%+v: got %v, want %v", f, got, want)
				}
			}
		}
	}
}

// outsideMerge merges branch into origin's main from another clone, as the
// GitHub button does (a merge commit), leaving the card alone.
func outsideMerge(t *testing.T, r *trackerRepo, branch string) {
	t.Helper()
	peer := t.TempDir()
	testfix.Git(t, "", "clone", "-q", r.origin, peer)
	testfix.Git(t, peer, "config", "user.name", "web")
	testfix.Git(t, peer, "config", "user.email", "w@w")
	testfix.Git(t, peer, "merge", "-q", "--no-ff", "--no-edit", "origin/"+branch)
	testfix.Git(t, peer, "push", "-q", "origin", "main")
}

// externalMergeFixture is issue 330's branch merged outside sdlc: after its
// close (closed), before it (not closed), or with no owner either.
func externalMergeFixture(t *testing.T, variant string) (*trackerRepo, string, string) {
	t.Helper()
	r, cardPath, detailPath := closeReady(t, 330)
	branch := r.git("branch", "--show-current")
	if variant == "closed" {
		stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
		if _, stderr, err := executeSDLCTestCommand("close", "--issue", "330", "--verified", "e2e", "--actual", "1", "--no-atlas", "--no-ledger", "--no-project"); err != nil {
			t.Fatalf("close: %v\n%s", err, stderr)
		}
	} else if err := pushIssueBranch(boundaryEnv(t), branch); err != nil {
		t.Fatal(err)
	}
	if variant == "unowned" {
		dropClaimant(t, r, "000330", cardPath)
	}
	outsideMerge(t, r, branch)
	r.git("fetch", "-q", "origin")
	return r, cardPath, detailPath
}

func stateDrift(t *testing.T) []DriftFinding {
	t.Helper()
	out, stderr, err := executeSDLCTestCommand("state", "--json")
	if err != nil {
		t.Fatalf("state: %v\n%s", err, stderr)
	}
	var s State
	if err := json.Unmarshal([]byte(out), &s); err != nil {
		t.Fatalf("state json: %v\n%s", err, out)
	}
	return s.Drift
}

// #287: `sdlc state` reports a merge done outside sdlc with the exact next
// action, and changes nothing.
func TestStateReportsExternalMerges(t *testing.T) {
	for variant, want := range map[string]string{
		"closed":     "sdlc issue recovery reconcile --issue 330",
		"not closed": "`sdlc close --issue 330` on a branch from main",
		"unowned":    "`sdlc claim --issue 330`, then `sdlc close --issue 330`",
	} {
		t.Run(variant, func(t *testing.T) {
			r, _, _ := externalMergeFixture(t, variant)
			mainBefore, trackerBefore := r.originMain(), r.git("ls-remote", "origin", "refs/heads/issue-tracker")
			var found []string
			for _, d := range stateDrift(t) {
				if d.Issue == "000330" {
					found = append(found, d.Message)
				}
			}
			if len(found) != 1 || !strings.Contains(found[0], want) {
				t.Fatalf("findings for #330: %q", found)
			}
			if r.originMain() != mainBefore || r.git("ls-remote", "origin", "refs/heads/issue-tracker") != trackerBefore {
				t.Fatal("state changed the remote")
			}
		})
	}
}

// A started branch with no work of its own sits on main's first-parent line
// and is not mistaken for a merge.
func TestStateIgnoresAFreshBranch(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	var out bytes.Buffer
	if err := startPlanBranch(context.Background(), &out, 9); err != nil { // pushes the branch, at main
		t.Fatal(err)
	}
	if remoteTip(t, r, s09Branch) != r.originMain() {
		t.Fatal("fixture: want a pushed branch with no work of its own")
	}
	for _, d := range stateDrift(t) {
		if strings.Contains(d.Message, "outside sdlc") {
			t.Fatalf("a fresh branch read as merged: %+v", d)
		}
	}
}

// #287: reconcile finishes a closed issue merged outside sdlc — card done,
// details archived on main mirroring it, the branch gone from the remote — and
// a rerun changes nothing. Without the close it names the next action and
// changes nothing.
func TestReconcileFinishesAnExternalMerge(t *testing.T) {
	t.Run("closed", func(t *testing.T) {
		r, cardPath, detailPath := externalMergeFixture(t, "closed")
		branch := r.git("branch", "--show-current")
		var out, errs bytes.Buffer
		if err := runRecoveryReconcile(context.Background(), &out, &errs, 330); err != nil {
			t.Fatalf("reconcile: %v\n%s", err, errs.String())
		}
		if !strings.Contains(r.card(cardPath), "status: done") {
			t.Fatalf("card:\n%s", r.card(cardPath))
		}
		r.git("fetch", "-q", "origin")
		archived := "workshop/history/issues/" + filepath.Base(detailPath)
		tree := r.git("ls-tree", "-r", "--name-only", "origin/main")
		if strings.Contains(tree, detailPath) || !strings.Contains(tree, archived) || !strings.Contains(r.git("show", "origin/main:"+archived), "status: done") {
			t.Fatalf("not archived on main:\n%s", tree)
		}
		if remoteTip(t, r, branch) != "" {
			t.Fatal("the merged branch is still on the remote")
		}
		before := r.originMain()
		errs.Reset()
		if err := runRecoveryReconcile(context.Background(), &out, &errs, 330); err != nil || r.originMain() != before {
			t.Fatalf("rerun: %v (main moved: %v)\n%s", err, r.originMain() != before, errs.String())
		}
	})
	t.Run("not closed", func(t *testing.T) {
		r, cardPath, _ := externalMergeFixture(t, "not closed")
		mainBefore, card := r.originMain(), r.card(cardPath)
		var out, errs bytes.Buffer
		if err := runRecoveryReconcile(context.Background(), &out, &errs, 330); err != nil {
			t.Fatalf("reconcile: %v\n%s", err, errs.String())
		}
		if !strings.Contains(errs.String(), "`sdlc close --issue 330`") {
			t.Fatalf("no next action:\n%s", errs.String())
		}
		if r.originMain() != mainBefore || r.card(cardPath) != card {
			t.Fatal("reconcile changed an unclosed outside merge")
		}
	})
}
