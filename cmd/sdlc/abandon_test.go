package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const (
	s09Archive = "refs/ariadne/abandoned/000009"
	s09History = "workshop/history/issues/000009-s09.md"
)

func abandon9(as, reason string) error {
	var out, errs bytes.Buffer
	err := runAbandon(context.Background(), &out, &errs, &abandonFlags{Issue: 9, As: as, Reason: reason,
		IssuesDir: "workshop/issues", PlansDir: "workshop/plans", HistoryDir: "workshop/history"})
	if err != nil {
		return errors.Join(err, errors.New(errs.String()))
	}
	return nil
}

func remoteRef(t *testing.T, r *trackerRepo, ref string) string {
	t.Helper()
	if f := strings.Fields(r.git("ls-remote", "origin", ref)); len(f) > 0 {
		return f[0]
	}
	return ""
}

// assertAbandoned checks abandon's whole end state for #9's started work.
func assertAbandoned(t *testing.T, r *trackerRepo, cardPath, as, tipBefore string) {
	t.Helper()
	if r.git("for-each-ref", "refs/heads/"+s09Branch) != "" || remoteTip(t, r, s09Branch) != "" {
		t.Fatal("the issue branch survived")
	}
	if on := r.git("branch", "--show-current"); on != "main" {
		t.Fatalf("the checkout is on %s, not its resting branch", on)
	}
	kept := remoteRef(t, r, s09Archive)
	r.git("fetch", "-q", "origin", s09Archive)
	if kept == "" || !gitSucceeds(r.root, "merge-base", "--is-ancestor", tipBefore, kept) ||
		!strings.Contains(r.git("log", "-1", "--format=%s", kept), "log: abandon ("+as+")") {
		t.Fatalf("the archive ref does not keep the work plus its note: %q", kept)
	}
	card := r.card(cardPath)
	rec, ok, err := issue.CardAbandoned([]byte(card))
	if err != nil || !ok || rec.Ref != s09Archive || rec.Head != kept || rec.Branch != s09Branch || !strings.Contains(card, "status: "+as) || !strings.Contains(card, "claimant:") {
		t.Fatalf("card: %+v %v %v\n%s", rec, ok, err, card)
	}
	r.git("fetch", "-q", "origin")
	tree := r.git("ls-tree", "-r", "--name-only", "origin/main")
	if strings.Contains(tree, handoffDetail) || !strings.Contains(tree, s09History) {
		t.Fatalf("details not archived on main:\n%s", tree)
	}
	archived := r.git("show", "origin/main:"+s09History)
	if !strings.Contains(archived, "status: "+as) || !strings.Contains(archived, "abandoned ("+as+")") {
		t.Fatalf("archived details do not mirror the card or carry the note:\n%s", archived)
	}
}

// #286: abandoning started work keeps it under the archive ref, ends the card,
// archives the details on main and deletes the branch everywhere.
func TestAbandonStartedWork(t *testing.T) {
	r, paths, _ := startedHere(t)
	tip := r.git("rev-parse", "HEAD")
	if err := abandon9("punt", "after the freeze"); err != nil {
		t.Fatal(err)
	}
	assertAbandoned(t, r, paths["000009"], "punt", tip)
	// A rerun finds everything done.
	before := r.git("rev-parse", "origin/main")
	if err := abandon9("punt", "after the freeze"); err != nil {
		t.Fatalf("rerun: %v", err)
	}
	r.git("fetch", "-q", "origin")
	if r.git("rev-parse", "origin/main") != before {
		t.Fatal("a rerun committed to main again")
	}
}

// An open issue has no branch: the card ends with an empty record and the
// details, with the note, are archived.
func TestAbandonOpenIssue(t *testing.T) {
	r, paths := claimSetRepo(t)
	claimFor(t, 9)
	if err := abandon9("wontfix", "out of scope"); err != nil {
		t.Fatal(err)
	}
	rec, ok, _ := issue.CardAbandoned([]byte(r.card(paths["000009"])))
	if !ok || rec.Started() || !strings.Contains(r.card(paths["000009"]), "status: wontfix") {
		t.Fatalf("card: %+v %v", rec, ok)
	}
	r.git("fetch", "-q", "origin")
	if got := r.git("show", "origin/main:"+s09History); !strings.Contains(got, "abandoned (wontfix): out of scope") {
		t.Fatalf("archived details:\n%s", got)
	}
}

// Each refusal names its own check, before any effect.
func TestAbandonRefusals(t *testing.T) {
	for _, c := range []struct {
		name, as, reason, want string
		shape                  func(t *testing.T, r *trackerRepo)
	}{
		{"bad --as", "done", "x", "--as must be", nil},
		{"no reason", "punt", " ", "--reason is required", nil},
		{"dirty tree", "punt", "x", "clean tree", func(t *testing.T, r *trackerRepo) { writeRepoFile(t, r.root, "cmd/nine.go", "package x\n") }},
		{"from rest", "punt", "x", "from its issue branch", func(t *testing.T, r *trackerRepo) { r.git("switch", "-q", "main") }},
		{"not owner", "punt", "x", "is owned by", func(t *testing.T, r *trackerRepo) { withClaimant(t, otherSlot) }},
	} {
		t.Run(c.name, func(t *testing.T) {
			r, _, _ := startedHere(t)
			if c.shape != nil {
				c.shape(t, r)
			}
			head := r.git("rev-parse", "HEAD")
			err := abandon9(c.as, c.reason)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want %q, got %v", c.want, err)
			}
			if r.git("rev-parse", "HEAD") != head || remoteRef(t, r, s09Archive) != "" {
				t.Fatal("a refused abandon had an effect")
			}
		})
	}
	t.Run("already terminal without the record", func(t *testing.T) {
		r, paths := claimSetRepo(t)
		claimFor(t, 9)
		env, err := openTracker(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if err := env.repo.ChangeCard("000009", paths["000009"], "fixture", operationToken("set"), func(c []byte) ([]byte, error) {
			return issue.SetCardField(c, "status", "wontfix")
		}); err != nil {
			t.Fatal(err)
		}
		_ = r
		if err := abandon9("punt", "x"); err == nil || !strings.Contains(err.Error(), "already wontfix") {
			t.Fatalf("got %v", err)
		}
	})
}

// An interrupted abandon is finished by its rerun, from wherever it stopped
// — including from the resting branch once the checkout has left the issue
// branch (the rerun is recognised from the card before the branch checks).
func TestAbandonRerunResumes(t *testing.T) {
	failCard := func() func() {
		prev := cardPublish
		cardPublish = func(*trackerEnv, tracker.Record, []byte, string, []string, func(string, string) error) error {
			return errors.New("interrupted")
		}
		return func() { cardPublish = prev }
	}
	failMain := func() func() {
		prev := mainPublish
		mainPublish = func(*trackerEnv, string, func(*gitx.TrunkView) (gitx.TrunkWrite, error), func(string, string) error) error {
			return errors.New("interrupted")
		}
		return func() { mainPublish = prev }
	}
	for _, c := range []struct {
		name  string
		stub  func() (undo func())
		after func(r *trackerRepo) // the rest of the interrupted state
	}{
		{"after the archive ref, before the card", failCard, nil},
		{"after the card, before main's archive", failMain, nil},
		{"remote branch deleted, still on the branch", failMain, func(r *trackerRepo) {
			r.git("push", "-q", "origin", "--delete", s09Branch)
		}},
		{"back on rest, local branch still present", failMain, func(r *trackerRepo) {
			r.git("push", "-q", "origin", "--delete", s09Branch)
			r.git("switch", "-q", "main")
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			r, paths, _ := startedHere(t)
			if err := pushIssueBranch(boundaryEnv(t), s09Branch); err != nil {
				t.Fatal(err)
			}
			tip := r.git("rev-parse", "HEAD")
			undo := c.stub()
			err := abandon9("punt", "r")
			undo()
			if err == nil {
				t.Fatal("fixture did not interrupt")
			}
			if c.after != nil {
				c.after(r)
			}
			if err := abandon9("punt", "r"); err != nil {
				t.Fatalf("rerun: %v", err)
			}
			assertAbandoned(t, r, paths["000009"], "punt", tip)
			if n := strings.Count(r.git("log", "--format=%s", remoteRef(t, r, s09Archive)), "log: abandon"); n != 1 {
				t.Fatalf("%d abandon notes on the kept tip", n)
			}
		})
	}
}

// abandonDecision follows the lifecycle model: every non-terminal status the
// model lets abandon/defer leave ends there; anything else refuses.
func TestAbandonDecision(t *testing.T) {
	_, card, _, _ := seededIssue(t, "000009", "s09")
	for _, status := range vocab.Issue().AllStatuses() {
		for _, as := range []string{"wontfix", "punt"} {
			in, err := issue.SetCardField([]byte(card), "status", status)
			if err != nil {
				continue // a status a card can't simply be put into (done needs hours)
			}
			out, err := abandonDecision(in, as, "2026-10-08", issue.Abandoned{})
			edge := vocab.Issue().TransitionForEvent(status, abandonEvent[as])
			if (edge != nil && edge.To == as) != (err == nil) {
				t.Fatalf("%s as %s: edge %v, err %v", status, as, edge, err)
			}
			if err == nil && (!strings.Contains(string(out), "status: "+as) || !strings.Contains(string(out), "abandoned:")) {
				t.Fatalf("%s as %s:\n%s", status, as, out)
			}
		}
	}
}

func reopen9() (string, error) {
	stdout, stderr, err := executeSDLCTestCommand("issue", "set-status", "working", "--issue", "9")
	return stdout + stderr, err
}

// abandonedHere is #9's started work abandoned as punt, with main moved on.
func abandonedHere(t *testing.T) (*trackerRepo, map[string]string, string) {
	t.Helper()
	r, paths, _ := startedHere(t)
	if err := abandon9("punt", "after the freeze"); err != nil {
		t.Fatal(err)
	}
	kept := remoteRef(t, r, s09Archive)
	peerCommit(t, r, "unrelated.go")
	return r, paths, kept
}

// assertReopened checks a reopen's whole end state.
func assertReopened(t *testing.T, r *trackerRepo, cardPath, kept string) {
	t.Helper()
	if on := r.git("branch", "--show-current"); on != s09Branch {
		t.Fatalf("on %s, not the restored branch", on)
	}
	r.git("fetch", "-q", "origin")
	if !gitSucceeds(r.root, "merge-base", "--is-ancestor", kept, "HEAD") || !gitSucceeds(r.root, "merge-base", "--is-ancestor", "origin/main", "HEAD") {
		t.Fatal("the branch lacks the kept tip or main")
	}
	if !gitSucceeds(r.root, "cat-file", "-e", "HEAD:"+handoffDetail) || gitSucceeds(r.root, "cat-file", "-e", "HEAD:"+s09History) {
		t.Fatal("the details were not moved back out of the archive")
	}
	if !strings.Contains(r.git("show", "HEAD:"+handoffDetail), "abandoned (punt)") {
		t.Fatal("the restored details lost the abandon note")
	}
	if tip := remoteTip(t, r, s09Branch); tip != r.git("rev-parse", "HEAD") {
		t.Fatalf("origin's branch %q is not HEAD", tip)
	}
	if dirty := r.git("status", "--porcelain"); dirty != "" {
		t.Fatalf("the reopen left changes: %s", dirty)
	}
	if remoteRef(t, r, s09Archive) != "" {
		t.Fatal("the archive ref survived the reopen")
	}
	card := r.card(cardPath)
	if _, ok, _ := issue.CardAbandoned([]byte(card)); ok || !strings.Contains(card, "status: working") {
		t.Fatalf("card:\n%s", card)
	}
	if err := guardTransferredDetails(context.Background()); err != nil {
		t.Fatalf("the reopened branch does not land: %v", err)
	}
}

// #286 D9: reopening an abandoned issue restores its branch at the kept tip,
// merges main and un-archives the details, so it lands cleanly.
func TestReopenRestoresAnAbandonedIssue(t *testing.T) {
	r, paths, kept := abandonedHere(t)
	if out, err := reopen9(); err != nil {
		t.Fatalf("reopen: %v\n%s", err, out)
	}
	assertReopened(t, r, paths["000009"], kept)
}

// A code conflict with main stops before the card changes; resolving it and
// rerunning finishes the reopen.
func TestReopenStopsOnAConflictThenResumes(t *testing.T) {
	r, paths, _ := startedHere(t)
	if err := abandon9("punt", "r"); err != nil {
		t.Fatal(err)
	}
	kept := remoteRef(t, r, s09Archive)
	peerAdd(t, r, "cmd/nine.go", "package nine // main's\n")
	out, err := reopen9()
	if err == nil || !strings.Contains(err.Error()+out, "conflict") {
		t.Fatalf("want a conflict stop: %v\n%s", err, out)
	}
	if !strings.Contains(r.card(paths["000009"]), "status: punt") || !gitSucceeds(r.root, "rev-parse", "-q", "--verify", "MERGE_HEAD") {
		t.Fatal("a conflicted reopen changed the card or abandoned the merge")
	}
	if out, err := reopen9(); err == nil || !strings.Contains(err.Error()+out, "in progress") {
		t.Fatalf("a rerun mid-merge: %v", err)
	}
	if !strings.Contains(err.Error()+out, "keep main's archived copy") {
		t.Fatalf("the stop does not say how to resolve the details: %v", err)
	}
	writeRepoFile(t, r.root, "cmd/nine.go", "package nine // both\n")
	r.git("add", "cmd/nine.go")
	r.git("rm", "-q", handoffDetail) // as the message says: keep main's archived copy
	r.git("commit", "-q", "--no-edit")
	if out, err := reopen9(); err != nil {
		t.Fatalf("rerun after resolving: %v\n%s", err, out)
	}
	assertReopened(t, r, paths["000009"], kept)
}

// A push that fails after the restore commit leaves the card terminal; the
// rerun pushes without restoring twice.
func TestReopenRerunAfterAFailedPush(t *testing.T) {
	r, paths, kept := abandonedHere(t)
	hooks := t.TempDir()
	if err := os.WriteFile(filepath.Join(hooks, "pre-push"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	r.git("config", "core.hooksPath", hooks)
	if out, err := reopen9(); err == nil || !strings.Contains(r.card(paths["000009"]), "status: punt") {
		t.Fatalf("want a refused reopen with the card unchanged: %v\n%s", err, out)
	}
	r.git("config", "--unset", "core.hooksPath")
	if out, err := reopen9(); err != nil {
		t.Fatalf("rerun: %v\n%s", err, out)
	}
	assertReopened(t, r, paths["000009"], kept)
	if n := strings.Count(r.git("log", "--format=%s"), "reopen: restore"); n != 1 {
		t.Fatalf("%d restore commits", n)
	}
}

// An archive ref a reopen failed to delete is overwritten by the next abandon:
// the branch contains its work.
func TestAbandonOverwritesAnOrphanArchiveRef(t *testing.T) {
	r, _, kept := abandonedHere(t)
	if out, err := reopen9(); err != nil {
		t.Fatalf("reopen: %v\n%s", err, out)
	}
	r.git("push", "-q", "origin", kept+":"+s09Archive) // the orphan
	if err := abandon9("wontfix", "for good"); err != nil {
		t.Fatalf("abandon over an orphan: %v", err)
	}
	if now := remoteRef(t, r, s09Archive); now == kept || !gitSucceeds(r.root, "merge-base", "--is-ancestor", kept, now) {
		t.Fatalf("the archive ref was not moved to the new tip: %s", now)
	}
}

// #286 D8: set-status cannot end started work as wontfix/punt, even forced;
// an open issue still can.
func TestSetStatusRedirectsStartedWorkToAbandon(t *testing.T) {
	startedHere(t)
	for _, args := range [][]string{{"issue", "set-status", "punt", "--issue", "9"}, {"issue", "set-status", "punt", "--issue", "9", "--force"}} {
		if _, stderr, err := executeSDLCTestCommand(args...); err == nil || !strings.Contains(err.Error()+stderr, "sdlc abandon --issue N --as punt") {
			t.Fatalf("%v: %v %s", args, err, stderr)
		}
	}
	if _, stderr, err := executeSDLCTestCommand("issue", "set-status", "wontfix", "--issue", "10"); err != nil {
		t.Fatalf("triage of an open issue: %v %s", err, stderr)
	}
}
