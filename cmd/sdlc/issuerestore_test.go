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

// The guard's offered restore makes a stale filing branch land: one commit
// puts main's version of #8's details back, and nothing else changes.
func TestIssueRestoreMakesAStaleFilingBranchLand(t *testing.T) {
	r := handedOff(t)
	ownSpinOff(t, r, otherSlot)
	reAddStale(t, r)
	if err := guardTransferredDetails(context.Background()); !errors.Is(err, errTransferredDetails) || !strings.Contains(err.Error(), refuseNotOwner) {
		t.Fatalf("fixture: want the restore offer, got %v", err)
	}
	before := r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	if err := runIssueRestore(context.Background(), &out, &errs, []int{8}); err != nil {
		t.Fatalf("restore: %v\n%s", err, errs.String())
	}
	if parent := r.git("rev-parse", "HEAD~1"); parent != before {
		t.Fatalf("restore made %s, want one commit on %s", r.git("log", "--oneline", before+"..HEAD"), before)
	}
	if subject := r.git("log", "-1", "--format=%s"); subject != "#8: issue: restore main's details" {
		t.Fatalf("subject %q", subject)
	}
	if files := r.git("diff", "--name-only", "HEAD~1", "HEAD"); files != spinOffDetails {
		t.Fatalf("restore touched %q", files)
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("after restore: %v", err)
	}
	// A rerun has nothing left to restore and commits nothing.
	out.Reset()
	if err := runIssueRestore(context.Background(), &out, &errs, []int{8}); err != nil || !strings.Contains(out.String(), "nothing to restore") {
		t.Fatalf("rerun: %v %q", err, out.String())
	}
	if r.git("rev-parse", "HEAD~1") != before {
		t.Fatal("rerun committed")
	}
}

// Main archived #8 (its details left workshop/issues/); a stale branch that
// still carries them would re-add them. Restore removes the branch's copy.
func TestIssueRestoreRemovesWhatMainArchived(t *testing.T) {
	r := handedOff(t)
	ownSpinOff(t, r, otherSlot)
	r.git("fetch", "-q", "origin")
	r.git("merge", "-q", "--no-edit", "origin/main")
	writeRepoFile(t, r.root, spinOffDetails, "a stale edit\n")
	r.git("commit", "-qam", "edit #8")
	archiveOnMain(t, r)
	if err := guardTransferredDetails(context.Background()); !errors.Is(err, errTransferredDetails) {
		t.Fatalf("fixture: want a refusal, got %v", err)
	}
	var out, errs bytes.Buffer
	if err := runIssueRestore(context.Background(), &out, &errs, []int{8}); err != nil {
		t.Fatalf("restore: %v\n%s", err, errs.String())
	}
	if gitSucceeds(r.root, "cat-file", "-e", "HEAD:"+spinOffDetails) {
		t.Fatal("restore left the archived details on the branch")
	}
	if got, want := r.git("rev-parse", "HEAD:workshop/history/issues/000008-spin-off.md"), r.git("rev-parse", "origin/main:workshop/history/issues/000008-spin-off.md"); got != want {
		t.Fatal("restore left the branch's archived copy different from main's")
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("after restore: %v", err)
	}
}

func TestIssueRestoreRefusesADirtyDetailsFile(t *testing.T) {
	r := handedOff(t)
	ownSpinOff(t, r, otherSlot)
	reAddStale(t, r)
	writeRepoFile(t, r.root, spinOffDetails, "uncommitted\n")
	before := r.git("rev-parse", "HEAD")
	var out, errs bytes.Buffer
	err := runIssueRestore(context.Background(), &out, &errs, []int{8})
	if err == nil || !strings.Contains(err.Error(), "uncommitted") || !strings.Contains(err.Error(), spinOffDetails) {
		t.Fatalf("dirty details: %v", err)
	}
	if r.git("rev-parse", "HEAD") != before {
		t.Fatal("a refused restore committed")
	}
}

// archiveOnMain moves #8's details to history on origin's main, from a peer.
func archiveOnMain(t *testing.T, r *trackerRepo) {
	t.Helper()
	peer := t.TempDir()
	git(t, "", "clone", "-q", r.origin, peer)
	git(t, peer, "config", "user.name", "peer")
	git(t, peer, "config", "user.email", "p@p")
	if err := os.MkdirAll(filepath.Join(peer, "workshop/history/issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, peer, "mv", spinOffDetails, "workshop/history/issues/000008-spin-off.md")
	git(t, peer, "commit", "-qm", "archive completed issues to history")
	git(t, peer, "push", "-q", "origin", "main")
}
