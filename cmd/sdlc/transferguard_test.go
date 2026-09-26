package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestTransferGuardOverBranchShapes(t *testing.T) {
	for _, c := range []struct {
		name   string
		shape  func(t *testing.T, r *trackerRepo)
		refuse string
	}{
		{"source branch after net-zero handoff", func(*testing.T, *trackerRepo) {}, ""},
		{"source merged main then left details alone", func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
		}, ""},
		{"source re-adds its stale copy", func(t *testing.T, r *trackerRepo) {
			stale := r.git("show", "HEAD~1:"+spinOffDetails)
			writeRepoFile(t, r.root, spinOffDetails, stale+"\n")
			r.git("add", spinOffDetails)
			r.git("commit", "-qm", "re-add")
		}, "conflicts"},
		{"source deletes details after merging main", func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
			r.git("rm", "-q", spinOffDetails)
			r.git("commit", "-qm", "drop details")
		}, "would be deleted"},
		{"source rewrites details after merging main", func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("merge", "-q", "--no-edit", "origin/main")
			writeRepoFile(t, r.root, spinOffDetails, "rewritten\n")
			r.git("commit", "-qam", "rewrite")
		}, "would be overwritten"},
		{"owner's own issue branch edits are exempt", func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("switch", "-q", "-c", "000008-spin-off", "origin/main")
			writeRepoFile(t, r.root, spinOffDetails, "owner's next edit\n")
			r.git("commit", "-qam", "owner works")
		}, ""},
		{"unrelated branch", func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("switch", "-q", "-c", "000011-other", "origin/main")
			writeRepoFile(t, r.root, "other.go", "package other\n")
			r.git("add", "other.go")
			r.git("commit", "-qm", "other work")
		}, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := handedOff(t)
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
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	r.git("push", "-q", "origin", "--delete", "issue-tracker")
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("no tracker: %v", err)
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
	if err := runPublishGate(r.git("merge-base", "HEAD", "origin/main"), "workshop/issues", &stderr); !errors.Is(err, errTransferredDetails) {
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
