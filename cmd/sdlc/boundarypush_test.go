package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

const s09Branch = "000009-s09"

// remoteTip is the publication remote's commit for branch, "" when absent.
func remoteTip(t *testing.T, r *trackerRepo, branch string) string {
	t.Helper()
	if f := strings.Fields(r.git("ls-remote", "--heads", "origin", "refs/heads/"+branch)); len(f) > 0 {
		return f[0]
	}
	return ""
}

func boundaryEnv(t *testing.T) *trackerEnv {
	t.Helper()
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return env
}

// #286: the boundary push leases on the last-fetched copy, so the owner's own
// rewrite is force-pushed while a tip it never saw is left alone.
func TestBoundaryPushLeases(t *testing.T) {
	r, _, owner := startedHere(t)
	var errs bytes.Buffer
	boundaryPush(boundaryEnv(t), &errs, "test")
	if remoteTip(t, r, s09Branch) != r.git("rev-parse", "HEAD") {
		t.Fatalf("first push: origin does not hold HEAD\n%s", errs.String())
	}
	r.git("commit", "-q", "--amend", "-m", "#9: rewritten") // the owner rebases
	boundaryPush(boundaryEnv(t), &errs, "test")
	if remoteTip(t, r, s09Branch) != r.git("rev-parse", "HEAD") {
		t.Fatalf("rewrite: origin does not hold the rewritten HEAD\n%s", errs.String())
	}
	// A lost response: origin has the tip, this checkout's tracking ref is stale.
	r.git("update-ref", "refs/remotes/origin/"+s09Branch, "HEAD~1")
	if err := pushIssueBranch(boundaryEnv(t), s09Branch); err != nil {
		t.Fatalf("repeat after a lost response: %v", err)
	}
	// Someone else pushed a tip this checkout never fetched.
	peer, _ := anotherMachine(t, r, owner)
	testfix.Git(t, peer, "fetch", "-q", "origin", s09Branch)
	testfix.Git(t, peer, "switch", "-q", s09Branch)
	writeRepoFile(t, peer, "rogue.txt", "x\n")
	testfix.Git(t, peer, "add", "rogue.txt")
	testfix.Git(t, peer, "commit", "-qm", "#9: rogue")
	testfix.Git(t, peer, "push", "-q", "origin", s09Branch)
	rogue := testfix.Capture(t, peer, "rev-parse", "HEAD")
	writeRepoFile(t, r.root, "mine.txt", "x\n")
	r.git("add", "mine.txt")
	r.git("commit", "-qm", "#9: mine")
	errs.Reset()
	boundaryPush(boundaryEnv(t), &errs, "test")
	if !strings.Contains(errs.String(), "last fetched") || remoteTip(t, r, s09Branch) != strings.TrimSpace(rogue) {
		t.Fatalf("an unseen tip must be refused and kept:\n%s", errs.String())
	}
}

func TestBoundaryPushSkipsTheRestingBranch(t *testing.T) {
	r, _, _ := startedHere(t)
	r.git("switch", "-q", "main")
	var errs bytes.Buffer
	boundaryPush(boundaryEnv(t), &errs, "test")
	if errs.Len() != 0 || remoteTip(t, r, s09Branch) != "" {
		t.Fatalf("pushed from rest: %q", errs.String())
	}
}

func TestDeleteRemoteBranchLeasesOnItsHead(t *testing.T) {
	r, _, _ := startedHere(t)
	if err := pushIssueBranch(boundaryEnv(t), s09Branch); err != nil {
		t.Fatal(err)
	}
	head := r.git("rev-parse", "HEAD")
	if err := deleteRemoteBranch(boundaryEnv(t), s09Branch, r.git("rev-parse", "HEAD~1")); err == nil || remoteTip(t, r, s09Branch) != head {
		t.Fatalf("deleting a branch past its expected head: %v", err)
	}
	if err := deleteRemoteBranch(boundaryEnv(t), s09Branch, head); err != nil || remoteTip(t, r, s09Branch) != "" {
		t.Fatalf("delete at its head: %v", err)
	}
	if err := deleteRemoteBranch(boundaryEnv(t), s09Branch, head); err != nil {
		t.Fatalf("a branch already gone: %v", err)
	}
}

// #286: after each boundary verb origin's issue branch equals local HEAD, and
// a rebase followed by the next boundary force-pushes with the lease.
func TestBoundaryVerbsPushTheIssueBranch(t *testing.T) {
	r, _, detailPath := closeReady(t, 320)
	branch := r.git("branch", "--show-current")
	if remoteTip(t, r, branch) == "" {
		t.Fatal("start-plan did not push the issue branch")
	}
	run := func(args ...string) {
		t.Helper()
		if _, stderr, err := executeSDLCTestCommand(args...); err != nil {
			t.Fatalf("sdlc %s: %v\n%s", strings.Join(args, " "), err, stderr)
		}
	}
	atHead := func(after string) {
		t.Helper()
		if tip, head := remoteTip(t, r, branch), r.git("rev-parse", "HEAD"); tip != head {
			t.Fatalf("after %s: origin has %s, HEAD is %s", after, tip, head)
		}
	}
	raw := r.git("show", "HEAD:"+detailPath)
	writeRepoFile(t, r.root, detailPath, strings.Replace(raw, "- [x] do it", "- [x] do it\n- [ ] M1 — part one", 1)+"\n")
	r.git("commit", "-qam", "#320: plan: M1")
	run("milestone-close", "--issue", "320", "--milestone", "M1", "--verified", "e2e", "--actual", "0.1", "--no-atlas", "--no-project", "--no-judge")
	atHead("milestone-close")
	r.git("commit", "-qam", "#320 M1: milestone close")

	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	run("close", "--issue", "320", "--verified", "e2e", "--actual", "1", "--no-atlas", "--no-ledger", "--no-project")
	atHead("close")

	peerCommit(t, r, "other.go")
	r.git("fetch", "-q", "origin")
	r.git("rebase", "-q", "origin/main")
	run("issue", "set-status", "working", "--issue", "320")
	run("close", "--issue", "320", "--verified", "e2e again", "--actual", "1", "--no-atlas", "--no-ledger", "--no-project")
	atHead("a close after a rebase")
}
