package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// handedOff runs the spin-off handoff, then lands the owner's edit E on main.
func handedOff(t *testing.T) *trackerRepo {
	t.Helper()
	r := spinOffOnCodeBranch(t)
	var out, errs bytes.Buffer
	if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(8)); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	ownerEdit(t, r)
	return r
}

const spinOffCard = "workshop/issue-cards/000008-spin-off.md"

// The two workspaces of the guard fixtures: this checkout, and another slot.
var (
	thisSlot  = issue.Claimant{Operator: "me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/this", Repository: "r"}
	otherSlot = issue.Claimant{Operator: "peer", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/other", Repository: "r"}
)

// ownSpinOff makes this checkout thisSlot and records owner as #8's claimant.
func ownSpinOff(t *testing.T, r *trackerRepo, owner issue.Claimant) {
	t.Helper()
	withClaimant(t, thisSlot)
	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.ChangeCard("000008", spinOffCard, "fixture owner", operationToken("set"), func(c []byte) ([]byte, error) {
		return issue.SetCardClaimant(c, owner)
	}); err != nil {
		t.Fatal(err)
	}
}

const (
	refuseNotOwner = "sdlc issue restore --issue 8"
	refuseBehind   = "Merge main"
)

func TestTransferGuardOverBranchShapes(t *testing.T) {
	for _, c := range []struct {
		name   string
		owner  bool
		shape  func(t *testing.T, r *trackerRepo)
		refuse string
	}{
		{"source branch after net-zero handoff", false, func(*testing.T, *trackerRepo) {}, ""},
		{"source merged main then left details alone", false, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
		}, ""},
		{"non-owner source re-adds its stale copy", false, reAddStale, refuseNotOwner},
		{"owner's checkout re-adds a stale copy", true, reAddStale, refuseBehind},
		{"non-owner deletes details after merging main", false, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
			r.git("rm", "-q", spinOffDetails)
			r.git("commit", "-qm", "drop details")
		}, refuseNotOwner},
		{"non-owner rewrites details after merging main", false, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
			writeRepoFile(t, r.root, spinOffDetails, "rewritten\n")
			r.git("commit", "-qam", "rewrite")
		}, refuseNotOwner},
		{"owner's edit on the issue branch", true, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000008-spin-off")
		}, ""},
		// pair#365: the issue's branch name was retired; the owner lands anyway.
		{"owner's edit on a renamed branch", true, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000008-spin-off-close")
		}, ""},
		{"owner's local main merges the issue branch for a direct push", true, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000008-spin-off")
			r.git("switch", "-q", "-C", "main", "origin/main")
			r.git("merge", "-q", "--no-ff", "--no-edit", "000008-spin-off")
		}, ""},
		{"owner's local main edits the details after that merge", true, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000008-spin-off")
			r.git("switch", "-q", "-C", "main", "origin/main")
			r.git("merge", "-q", "--no-ff", "--no-edit", "000008-spin-off")
			writeRepoFile(t, r.root, spinOffDetails, "the owner again\n")
			r.git("commit", "-qam", "edit on main")
		}, ""},
		{"non-owner's local main merges its rewrite", false, func(t *testing.T, r *trackerRepo) {
			issueBranchEdit(t, r, "000013-meddler")
			r.git("switch", "-q", "-C", "main", "origin/main")
			r.git("merge", "-q", "--no-ff", "--no-edit", "000013-meddler")
		}, refuseNotOwner},
		{"unrelated branch", false, func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("switch", "-q", "-c", "000011-other", "origin/main")
			writeRepoFile(t, r.root, "other.go", "package other\n")
			r.git("add", "other.go")
			r.git("commit", "-qm", "other work")
		}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := handedOff(t)
			owner := otherSlot
			if c.owner {
				owner = thisSlot
			}
			ownSpinOff(t, r, owner)
			c.shape(t, r)
			err := guardTransferredDetails(context.Background())
			if c.refuse == "" {
				if err != nil {
					t.Fatalf("refused a safe landing: %v", err)
				}
				return
			}
			if !errors.Is(err, errTransferredDetails) || !strings.Contains(err.Error(), c.refuse) {
				t.Fatalf("want refusal %q, got %v", c.refuse, err)
			}
		})
	}
}

// reAddStale keeps the filing branch's pre-handoff copy of #8's details.
func reAddStale(t *testing.T, r *trackerRepo) {
	stale := r.git("show", "HEAD~1:"+spinOffDetails)
	writeRepoFile(t, r.root, spinOffDetails, stale+"\n")
	r.git("add", spinOffDetails)
	r.git("commit", "-qm", "re-add")
}

// An owner whose branch predates main's latest version of the details (another
// slot published edit E) must merge main first, whether or not git could merge
// the two edits cleanly; after the merge it lands.
func TestTransferGuardOwnerBehindMainLatest(t *testing.T) {
	for _, c := range []struct {
		name     string
		edit     func(before string) string
		conflict bool
	}{
		{"edits merge cleanly", func(before string) string { return before + "\nThe owner's closing note.\n" }, false},
		{"edits conflict", func(string) string { return "the owner's design\n" }, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := handedOff(t)
			ownSpinOff(t, r, thisSlot)
			r.git("fetch", "-q", "origin")
			r.git("switch", "-q", "-c", "000008-spin-off", "origin/main~1")
			writeRepoFile(t, r.root, spinOffDetails, c.edit(r.git("show", "HEAD:"+spinOffDetails)))
			r.git("commit", "-qam", "owner edits")
			err := guardTransferredDetails(context.Background())
			if !errors.Is(err, errTransferredDetails) || !strings.Contains(err.Error(), refuseBehind) {
				t.Fatalf("owner behind main's latest version: %v", err)
			}
			out, mergeErr := exec.Command("git", "-C", r.root, "merge", "-q", "--no-edit", "origin/main").CombinedOutput()
			if (mergeErr != nil) != c.conflict {
				t.Fatalf("fixture: merge conflict = %v, want %v: %s", mergeErr != nil, c.conflict, out)
			}
			if c.conflict {
				writeRepoFile(t, r.root, spinOffDetails, "the owner's design, after E\n")
				r.git("commit", "-qam", "merge main")
			}
			if err := guardTransferredDetails(context.Background()); err != nil {
				t.Fatalf("owner after merging main: %v", err)
			}
		})
	}
}

func TestTransferGuardInterruptedRemovalProtectsOwnerEdits(t *testing.T) {
	r := spinOffOnCodeBranch(t)
	// Interrupt before the source removal: the branch still carries B.
	hooks, state := t.TempDir(), t.TempDir()
	hook := "#!/bin/sh\nwhile read l ls rr rs; do\n  case \"$rr\" in refs/heads/issue-tracker)\n" +
		"    n=$(cat '" + state + "/n' 2>/dev/null || echo 0); n=$((n+1)); echo $n > '" + state + "/n'\n" +
		"    if [ $n -ge 2 ]; then exit 1; fi;;\n  esac\ndone\nexit 0\n"
	if err := os.WriteFile(filepath.Join(hooks, "pre-push"), []byte(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	r.git("config", "core.hooksPath", hooks)
	var out, errs bytes.Buffer
	_ = runMoveDetail(context.Background(), &out, &errs, moveFlags(8))
	r.git("config", "--unset", "core.hooksPath")
	// D is on main but the card does not name it yet, and the source still has B.
	// The owner edits main meanwhile; reconcile must still converge to net-zero.
	ownerEdit(t, r)
	var out2, errs2 bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out2, &errs2, 8); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, errs2.String())
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("after reconcile the source branch is net-zero: %v", err)
	}
}

func TestTransferGuardWithoutTrackerIsNotApplicable(t *testing.T) {
	legacyRepo(t)
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("legacy repository: %v", err)
	}
	// A migrated checkout whose tracker vanished is not "no tracker": the
	// guard refuses rather than skip the protection (#252 cutover guard).
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	r.git("push", "-q", "origin", "--delete", "issue-tracker")
	if err := guardTransferredDetails(context.Background()); !errors.Is(err, tracker.ErrCutover) {
		t.Fatalf("marker without tracker: %v", err)
	}
}

// The publish gate (push, merge) and PR creation run the guard for real.
func TestPublishGateAndPRRefuseChangedHandedOffDetails(t *testing.T) {
	r := handedOff(t)
	r.git("fetch", "-q", "origin")
	r.git("merge", "-q", "--no-edit", "origin/main")
	r.git("rm", "-q", spinOffDetails)
	r.git("commit", "-qm", "drop details")
	var stderr bytes.Buffer
	if err := runPublishGate(context.Background(), r.git("merge-base", "HEAD", "origin/main"), "workshop/issues", &stderr); !errors.Is(err, errTransferredDetails) {
		t.Fatalf("publish gate: %v", err)
	}
	var stdout bytes.Buffer
	if err := runPR(&stdout, &stderr, &prFlags{IssuesDir: "workshop/issues"}); !errors.Is(err, errTransferredDetails) {
		t.Fatalf("pr: %v", err)
	}
}

// A fresh clone has none of the originating checkout's recovery refs; the
// handoff record on the card alone must still protect the details.
func TestTransferGuardFromAFreshCloneUsesOnlyTrackerRecords(t *testing.T) {
	r := handedOff(t)
	r.git("push", "-q", "origin", "000007-seven")
	clone := filepath.Join(t.TempDir(), "fresh")
	git(t, "", "clone", "-q", r.origin, clone)
	git(t, clone, "config", "user.name", "c")
	git(t, clone, "config", "user.email", "c@c")
	git(t, clone, "switch", "-q", "000007-seven")
	chdirTo(t, clone)
	if refs := git(t, clone, "for-each-ref", "refs/sdlc/"); strings.TrimSpace(refs) != "" {
		t.Fatalf("fixture: clone has private refs %s", refs)
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("net-zero source branch refused from a fresh clone: %v", err)
	}
	git(t, clone, "merge", "-q", "--no-edit", "origin/main")
	writeRepoFile(t, clone, spinOffDetails, "rewritten in a fresh clone\n")
	git(t, clone, "commit", "-qam", "rewrite")
	if err := guardTransferredDetails(context.Background()); !errors.Is(err, errTransferredDetails) {
		t.Fatalf("fresh clone rewrite not refused: %v", err)
	}
}

// BR-6: between transfer.main and transfer.record the card names no main
// commit, but main already holds the details and may already carry the new
// owner's edits. A landing of the unfinished source branch must be refused.
func TestTransferGuardProtectsAnUnrecordedPublication(t *testing.T) {
	r := spinOffOnCodeBranch(t)
	block := blockTrackerAfterFirstPush(t, r)
	var out, errs bytes.Buffer
	if err := runMoveDetail(context.Background(), &out, &errs, moveFlags(8)); err == nil {
		t.Fatal("interruption fixture did not interrupt")
	}
	_ = os.Remove(block)
	h, ok, _ := issue.CardHandoff([]byte(r.card("workshop/issue-cards/000008-spin-off.md")))
	if !ok || h.MainCommit != "" {
		t.Fatalf("fixture: want an unrecorded handoff, got %+v", h)
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("identical unfinished copy loses nothing: %v", err)
	}
	ownerEdit(t, r)
	if err := guardTransferredDetails(context.Background()); !errors.Is(err, errTransferredDetails) {
		t.Fatalf("unrecorded publication left unprotected: %v", err)
	}
}

// issueBranchEdit is an edit to #8's details on a fresh branch from main.
func issueBranchEdit(t *testing.T, r *trackerRepo, branch string) {
	t.Helper()
	r.git("fetch", "-q", "origin")
	r.git("switch", "-q", "-c", branch, "origin/main")
	writeRepoFile(t, r.root, spinOffDetails, "an edit on "+branch+"\n")
	r.git("commit", "-qam", "edit on "+branch)
}

// #285: the owner is the card's claimant, never a branch. No command help and
// neither guard refusal may say otherwise.
func TestNoHelpCallsABranchTheOwner(t *testing.T) {
	branchOwner := regexp.MustCompile(`(?i)owner'?s? (issue )?branch|owner branch|branch (is|as) the owner`)
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			walk(sub)
		}
		if m := branchOwner.FindString(c.Short + "\n" + c.Long); m != "" {
			t.Errorf("`%s` help calls a branch the owner: %q", c.CommandPath(), m)
		}
	}
	walk(buildRoot())
	env := &trackerEnv{target: gitx.PublicationTarget{Remote: "origin"}}
	for _, v := range []verdict{verdictNotOwner, verdictBehind} {
		if m := branchOwner.FindString(detailsRefusal(env, changedDetail{ID: "000008", Path: spinOffDetails, Verdict: v}).Error()); m != "" {
			t.Errorf("refusal %v calls a branch the owner: %q", v, m)
		}
	}
}
