package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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

// BR-1: a tracking ref this checkout kept for the landed branch, which origin
// has since deleted, does not make every rerun's leased push refuse.
func TestReopenDoneWithAStaleTrackingRef(t *testing.T) {
	r, cardPath, detailPath := doneArchived(t)
	writeReopenEntry(t, r)
	r.git("update-ref", "refs/remotes/origin/"+s330Branch, r.git("rev-parse", "refs/heads/"+s330Branch))
	if out, err := reopen330(); err != nil {
		t.Fatalf("reopen: %v\n%s", err, out)
	}
	assertDoneReopened(t, r, cardPath, detailPath)
}

// BR-2: a finish lost after the card write (the mirror commit and its push)
// is resumed by rerunning: the remote branch is gone, as if never pushed.
func TestReopenDoneRerunFinishesALostFinish(t *testing.T) {
	r, cardPath, detailPath := doneArchived(t)
	writeReopenEntry(t, r)
	if out, err := reopen330(); err != nil {
		t.Fatalf("reopen: %v\n%s", err, out)
	}
	if strings.Contains(r.git("log", "-1", "--format=%s"), "mirror working card") {
		r.git("reset", "-q", "HEAD~1") // the finish never committed its mirror
	}
	r.git("push", "-q", "origin", ":"+s330Branch) // nor pushed
	if out, err := reopen330(); err != nil {
		t.Fatalf("rerun: %v\n%s", err, out)
	}
	assertDoneReopened(t, r, cardPath, detailPath)
}

// BR-3, BR-4: each refusal fires before anything changes; --force skips only
// the Log guard.
func TestReopenDoneRefusals(t *testing.T) {
	for name, tc := range map[string]struct {
		setup func(t *testing.T, r *trackerRepo, cardPath string)
		want  string
	}{
		"another workspace owns it": {func(t *testing.T, r *trackerRepo, cardPath string) {
			owner, _ := ownerOf(t, r, cardPath)
			_, them := anotherMachine(t, r, owner)
			env, err := openTrackerAt(context.Background(), r.root)
			if err != nil {
				t.Fatal(err)
			}
			if err := env.repo.ChangeCard("000330", cardPath, "fixture", operationToken("set"), func(c []byte) ([]byte, error) {
				return issue.SetCardClaimant(c, them)
			}); err != nil {
				t.Fatal(err)
			}
		}, "only that workspace reopens it"},
		"another branch": {func(t *testing.T, r *trackerRepo, _ string) {
			r.git("switch", "-q", "-c", "elsewhere")
		}, "reopen from the resting branch"},
		"a dirty tree": {func(t *testing.T, r *trackerRepo, _ string) {
			writeRepoFile(t, r.root, "cmd/a.go", "package a // edited\n")
		}, "cmd/a.go"},
		"an unlanded local branch without the archive": {func(t *testing.T, r *trackerRepo, _ string) {
			r.git("branch", "-f", s330Branch, "HEAD~3")
			r.git("switch", "-q", s330Branch)
			r.git("commit", "-q", "--allow-empty", "-m", "#330: stray")
			r.git("switch", "-q", "main")
		}, "predates main's archive"},
	} {
		t.Run(name, func(t *testing.T) {
			r, cardPath, _ := doneArchived(t)
			tc.setup(t, r, cardPath)
			writeReopenEntry(t, r)
			branch, head := r.git("branch", "--show-current"), r.git("rev-parse", "HEAD")
			out, err := reopen330()
			if err == nil || !strings.Contains(err.Error()+out, tc.want) {
				t.Fatalf("want a refusal naming %q: %v\n%s", tc.want, err, out)
			}
			if r.git("branch", "--show-current") != branch || r.git("rev-parse", "HEAD") != head ||
				remoteTip(t, r, s330Branch) != "" || !strings.Contains(r.card(cardPath), "status: done") {
				t.Fatal("a refused reopen changed something")
			}
		})
	}
	t.Run("--force skips the Log guard", func(t *testing.T) {
		r, cardPath, detailPath := doneArchived(t)
		stdout, stderr, err := executeSDLCTestCommand("issue", "set-status", "working", "--issue", "330", "--force")
		if err != nil {
			t.Fatalf("forced reopen: %v\n%s%s", err, stdout, stderr)
		}
		if !gitSucceeds(r.root, "cat-file", "-e", "HEAD:"+detailPath) || !strings.Contains(r.card(cardPath), "status: working") {
			t.Fatal("a forced reopen did not un-archive")
		}
	})
}
