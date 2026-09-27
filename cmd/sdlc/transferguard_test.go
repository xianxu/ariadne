package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
		{"main merges the owner's branch for a direct push", func(t *testing.T, r *trackerRepo) {
			ownerBranchEdit(t, r)
			r.git("switch", "-q", "-C", "main", "origin/main")
			r.git("merge", "-q", "--no-ff", "--no-edit", "000008-spin-off")
		}, ""},
		{"a branch stacked on the owner's", func(t *testing.T, r *trackerRepo) {
			ownerBranchEdit(t, r)
			r.git("switch", "-q", "-c", "000012-stacked")
			writeRepoFile(t, r.root, "stacked.go", "package stacked\n")
			r.git("add", "stacked.go")
			r.git("commit", "-qm", "stacked work")
		}, ""},
		{"main merges a non-owner rewrite", func(t *testing.T, r *trackerRepo) {
			r.git("fetch", "-q", "origin")
			r.git("switch", "-q", "-c", "000013-meddler", "origin/main")
			writeRepoFile(t, r.root, spinOffDetails, "rewritten by someone else\n")
			r.git("commit", "-qam", "meddle")
			r.git("switch", "-q", "-C", "main", "origin/main")
			r.git("merge", "-q", "--no-ff", "--no-edit", "000013-meddler")
		}, "would be overwritten"},
		{"main edits the details itself after merging the owner's work", func(t *testing.T, r *trackerRepo) {
			ownerBranchEdit(t, r)
			r.git("switch", "-q", "-C", "main", "origin/main")
			r.git("merge", "-q", "--no-ff", "--no-edit", "000008-spin-off")
			writeRepoFile(t, r.root, spinOffDetails, "not the owner's\n")
			r.git("commit", "-qam", "edit on main")
		}, "would be overwritten"},
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

func TestTransferGuardRefusesMalformedHandoffRepoWide(t *testing.T) {
	cardPath, card, _, _ := seededIssue(t, "000009", "nine")
	oid := strings.Repeat("a", 40)
	bad, err := issue.SetCardHandoff([]byte(card), issue.Handoff{Token: "move-x", Repository: "file:/x", SourceBranch: "refs/heads/x",
		SourceBase: oid, SourceHEAD: oid, SourceBlob: oid, SourcePath: "../../etc/000009-nine.md", Destination: "../../etc/000009-nine.md", MainCommit: oid})
	if err != nil {
		t.Fatal(err)
	}
	newTrackerRepo(t, map[string]string{cardPath: string(bad)}, nil)
	err = guardTransferredDetails(context.Background())
	if err == nil || !strings.Contains(err.Error(), "#000009 is malformed") || !strings.Contains(err.Error(), "every PR, push and merge") {
		t.Fatalf("malformed destination: %v", err)
	}
}

// ownerBranchEdit is the claimed owner's design edit on the issue's own branch.
func ownerBranchEdit(t *testing.T, r *trackerRepo) {
	t.Helper()
	r.git("fetch", "-q", "origin")
	r.git("switch", "-q", "-c", "000008-spin-off", "origin/main")
	writeRepoFile(t, r.root, spinOffDetails, "owner's design\n")
	r.git("commit", "-qam", "owner designs")
}

