package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #282 (B2 d, e): a worktree mirror refresh lands only on the issue's own
// branch. A claim or setter run from the resting branch, another issue's code
// branch or a detached HEAD changes the card and leaves the checkout clean.
func TestMirrorRefreshesOnlyOnTheIssueBranch(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	otherCard, other, otherDetailPath, otherDetail := seededIssue(t, "000010", "ten")
	r := newTrackerRepo(t, map[string]string{cardPath: card, otherCard: other}, map[string]string{detailPath: detail, otherDetailPath: otherDetail})
	clean := func(where string) {
		t.Helper()
		if st := r.git("status", "--porcelain"); st != "" {
			t.Fatalf("%s: the checkout was edited:\n%s", where, st)
		}
	}
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatalf("claim on the resting branch: %v\n%s", err, errs.String())
	}
	clean("claim on the resting branch")
	r.git("switch", "-q", "-c", "000010-ten")
	setTitle(t, 9, "Retitled on another branch")
	clean("set-title on another issue's branch")
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatalf("claim on another issue's branch: %v\n%s", err, errs.String())
	}
	clean("claim on another issue's branch")
	r.git("switch", "-q", "--detach")
	setTitle(t, 9, "Retitled detached")
	clean("set-title on a detached HEAD")
	r.git("switch", "-q", "-c", "000009-nine")
	setTitle(t, 9, "Retitled at home")
	if local, _ := os.ReadFile(filepath.Join(r.root, detailPath)); !strings.Contains(string(local), "# Retitled at home") {
		t.Fatalf("the issue's own branch was not refreshed:\n%s", local)
	}
}

// #282 (B2 c): a revert that rolls the mirror back leaves an older, untouched
// projection; re-running claim on the issue branch converges it, so the
// mirror is never restored by hand.
func TestRevertedMirrorConvergesOnTheNextRefresh(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000009", "nine")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	r.git("switch", "-q", "-c", "000009-nine")
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatalf("claim: %v\n%s", err, errs.String())
	}
	r.git("commit", "-qam", "claim mirror")
	setTitle(t, 9, "Second title")
	r.git("commit", "-qam", "retitle mirror")
	r.git("revert", "--no-edit", "HEAD")
	errs.Reset()
	if err := runClaim(context.Background(), &out, &errs, claimFlagsFor(9)); err != nil {
		t.Fatalf("re-claim: %v\n%s", err, errs.String())
	}
	local, _ := os.ReadFile(filepath.Join(r.root, detailPath))
	if !strings.Contains(string(local), "# Second title") || strings.Contains(errs.String(), "not refreshed") {
		t.Fatalf("the reverted mirror did not converge:\n%s\n%s", local, errs.String())
	}
}

func setTitle(t *testing.T, id int, title string) {
	t.Helper()
	root := buildRoot()
	root.SetArgs([]string{"issue", "set-title", "--issue", itoa(id), title})
	var errs bytes.Buffer
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&errs)
	if err := root.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("set-title: %v\n%s", err, errs.String())
	}
	if strings.Contains(errs.String(), "not refreshed") {
		t.Fatalf("set-title warned: %s", errs.String())
	}
}
