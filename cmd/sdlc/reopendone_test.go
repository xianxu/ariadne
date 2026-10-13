package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	s330Branch  = "000330-e2e"
	s330History = "workshop/history/issues/000330-e2e.md"
	s330Plan    = "workshop/plans/000330-e2e-plan.md"
	s330PlanArc = "workshop/history/plans/000330-e2e-plan.md"
)

// doneArchived is #330 landed and reconciled (card done, details archived on
// main), with an archived plan, and this checkout back on its resting branch.
func doneArchived(t *testing.T) (*trackerRepo, string, string) {
	t.Helper()
	r, cardPath, detailPath := externalMergeFixture(t, "closed")
	var out, errs bytes.Buffer
	if err := runRecoveryReconcile(context.Background(), &out, &errs, 330); err != nil {
		t.Fatalf("reconcile: %v\n%s", err, errs.String())
	}
	peerAdd(t, r, s330PlanArc, "# plan\n")
	// The fixture's own notes are dated today; the reopen's entry must be new.
	r.git("fetch", "-q", "origin")
	today := time.Now().Format("2006-01-02")
	peerAdd(t, r, s330History, strings.ReplaceAll(r.git("show", "origin/main:"+s330History), today, "2000-01-01")+"\n")
	r.git("switch", "-q", "main")
	r.git("pull", "-q", "--ff-only", "origin", "main")
	if !strings.Contains(r.card(cardPath), "status: done") {
		t.Fatalf("fixture: card not done\n%s", r.card(cardPath))
	}
	return r, cardPath, detailPath
}

func reopen330() (string, error) {
	stdout, stderr, err := executeSDLCTestCommand("issue", "set-status", "working", "--issue", "330")
	return stdout + stderr, err
}

// writeReopenEntry files today's reopen Log line in the archived copy, as the
// guard's refusal says.
func writeReopenEntry(t *testing.T, r *trackerRepo) {
	t.Helper()
	p := filepath.Join(r.root, s330History)
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	line := "- " + time.Now().Format("2006-01-02") + ": reopened — a regression\n"
	if err := os.WriteFile(p, append(raw, line...), 0o644); err != nil {
		t.Fatal(err)
	}
}

// assertDoneReopened checks a done reopen's whole end state.
func assertDoneReopened(t *testing.T, r *trackerRepo, cardPath, detailPath string) {
	t.Helper()
	if on := r.git("branch", "--show-current"); on != s330Branch {
		t.Fatalf("on %s, not the issue branch", on)
	}
	r.git("fetch", "-q", "origin")
	if !gitSucceeds(r.root, "merge-base", "--is-ancestor", "origin/main", "HEAD") {
		t.Fatal("the branch is not based on main")
	}
	for live, archived := range map[string]string{detailPath: s330History, s330Plan: s330PlanArc} {
		if !gitSucceeds(r.root, "cat-file", "-e", "HEAD:"+live) || gitSucceeds(r.root, "cat-file", "-e", "HEAD:"+archived) {
			t.Fatalf("%s was not moved back to %s", archived, live)
		}
	}
	if !strings.Contains(r.git("show", "HEAD:"+detailPath), "reopened — a regression") {
		t.Fatal("the un-archived details lost the reopen entry")
	}
	if tip := remoteTip(t, r, s330Branch); tip != r.git("rev-parse", "HEAD") {
		t.Fatalf("origin's branch %q is not HEAD", tip)
	}
	if dirty := r.git("status", "--porcelain"); dirty != "" {
		t.Fatalf("the reopen left changes: %s", dirty)
	}
	if card := r.card(cardPath); !strings.Contains(card, "status: working") {
		t.Fatalf("card:\n%s", card)
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("the reopened branch does not land: %v", err)
	}
	if n := strings.Count(r.git("log", "--format=%s"), "reopen: un-archive"); n != 1 {
		t.Fatalf("%d un-archive commits", n)
	}
}

// #305: reopening a done issue un-archives its details and plans onto a pushed
// branch from main, so its next landing un-archives them cleanly. Without the
// reopen's Log entry it refuses, naming the archived copy, and changes nothing.
func TestReopenUnarchivesADoneIssue(t *testing.T) {
	r, cardPath, detailPath := doneArchived(t)
	head, mainBefore := r.git("rev-parse", "HEAD"), r.originMain()
	out, err := reopen330()
	if err == nil || !strings.Contains(err.Error()+out, s330History) {
		t.Fatalf("want a refusal naming the archived copy: %v\n%s", err, out)
	}
	if r.git("rev-parse", "HEAD") != head || r.git("branch", "--show-current") != "main" || r.originMain() != mainBefore ||
		remoteTip(t, r, s330Branch) != "" || !strings.Contains(r.card(cardPath), "status: done") {
		t.Fatal("a refused reopen changed something")
	}
	writeReopenEntry(t, r)
	if out, err := reopen330(); err != nil {
		t.Fatalf("reopen: %v\n%s", err, out)
	}
	assertDoneReopened(t, r, cardPath, detailPath)
}

// A push that fails after the un-archive commit leaves the card done; the
// rerun, from the issue branch, pushes without un-archiving twice.
func TestReopenDoneRerunAfterAFailedPush(t *testing.T) {
	r, cardPath, detailPath := doneArchived(t)
	writeReopenEntry(t, r)
	hooks := t.TempDir()
	if err := os.WriteFile(filepath.Join(hooks, "pre-push"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	r.git("config", "core.hooksPath", hooks)
	if out, err := reopen330(); err == nil || !strings.Contains(r.card(cardPath), "status: done") {
		t.Fatalf("want a refused reopen with the card unchanged: %v\n%s", err, out)
	}
	r.git("config", "--unset", "core.hooksPath")
	if out, err := reopen330(); err != nil {
		t.Fatalf("rerun: %v\n%s", err, out)
	}
	assertDoneReopened(t, r, cardPath, detailPath)
}
