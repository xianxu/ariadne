package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func publish(t *testing.T, ids ...int) (string, error) {
	t.Helper()
	var out, errs bytes.Buffer
	err := runTrackedPublish(context.Background(), &out, &errs, ids, "workshop/issues", false)
	return out.String() + errs.String(), err
}

func claimFor(t *testing.T, ids ...int) {
	t.Helper()
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, &claimFlags{Issues: ids, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil {
		t.Fatalf("claim %v: %v\n%s", ids, err, errs.String())
	}
}

// appendDetail adds a line to an issue's local details.
func appendDetail(t *testing.T, r *trackerRepo, rel, line string) {
	t.Helper()
	writeRepoFile(t, r.root, rel, r.git("show", "HEAD:"+rel)+"\n"+line+"\n")
}

// #284: claim → edit on the resting branch → publish lands the edit on main in
// one narrow commit, and the rest ends equal to main with a clean tree.
func TestPublishOwnedEditsFromRest(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9, 10)
	appendDetail(t, r, "workshop/issues/000009-s09.md", "Shaped nine.")
	appendDetail(t, r, "workshop/issues/000010-s10.md", "Shaped ten.")
	before := r.originMain()
	if out, err := publish(t, 9, 10); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if parent := r.git("rev-parse", r.originMain()+"^"); parent != before {
		t.Fatal("the edits took more than one main commit")
	}
	if subject := r.git("log", "-1", "--format=%s", r.originMain()); subject != "#9,#10: issue: publish details" {
		t.Fatalf("subject %q", subject)
	}
	for rel, line := range map[string]string{"workshop/issues/000009-s09.md": "Shaped nine.", "workshop/issues/000010-s10.md": "Shaped ten."} {
		if !strings.Contains(r.git("show", r.originMain()+":"+rel), line) {
			t.Fatalf("%s lacks %q on main", rel, line)
		}
	}
	if r.git("rev-parse", "HEAD") != r.originMain() || r.git("status", "--porcelain") != "" {
		t.Fatalf("rest not left equal to main and clean:\n%s", r.git("status", "--porcelain"))
	}
}

// #284: publishing from an issue branch writes main and commits the same bytes
// narrowly on the branch, so merging main later changes nothing for that file.
func TestPublishFromAnIssueBranch(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	r.git("switch", "-q", "-c", "000009-s09")
	appendDetail(t, r, "workshop/issues/000009-s09.md", "Shaped on the branch.")
	if out, err := publish(t, 9); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	r.git("fetch", "-q", "origin")
	if diff := r.git("diff", "HEAD", "origin/main", "--", "workshop/issues/000009-s09.md"); diff != "" {
		t.Fatalf("the branch and main disagree on the published file:\n%s", diff)
	}
	if subject := r.git("log", "-1", "--format=%s"); subject != "#9: issue: details as published" {
		t.Fatalf("branch commit %q", subject)
	}
}

// #284: refusals publish nothing — not owned, unowned, and a main copy that
// moved since this checkout's base.
func TestPublishRefusals(t *testing.T) {
	r, paths := claimSetRepo(t)
	appendDetail(t, r, "workshop/issues/000009-s09.md", "Unclaimed edit.")
	before := r.originMain()
	if out, err := publish(t, 9); err == nil || !strings.Contains(err.Error(), "claim it first") {
		t.Fatalf("unowned: %v\n%s", err, out)
	}
	r.git("checkout", "--", "workshop/issues/000009-s09.md")
	claimFor(t, 10)
	owner, _ := ownerOf(t, r, paths["000010"])
	elsewhere := owner
	elsewhere.Worktree = "/elsewhere/ariadne"
	withClaimant(t, elsewhere)
	appendDetail(t, r, "workshop/issues/000010-s10.md", "Not my edit to publish.")
	if out, err := publish(t, 10); err == nil || !strings.Contains(err.Error(), "only its owner republishes") {
		t.Fatalf("foreign: %v\n%s", err, out)
	}
	withClaimant(t, owner)
	peerAdd(t, r, "workshop/issues/000010-s10.md", r.git("show", "HEAD:workshop/issues/000010-s10.md")+"\nA peer's edit on main.\n")
	if out, err := publish(t, 10); err == nil || !strings.Contains(err.Error(), "main's details changed since this checkout's base") {
		t.Fatalf("moved main: %v\n%s", err, out)
	}
	if r.originMain() == before {
		return // the peer's commit is the only one; nothing of ours published
	}
	if log := r.git("log", "--format=%s", before+".."+r.originMain()); strings.Contains(log, "publish details") {
		t.Fatalf("a refused publish wrote main:\n%s", log)
	}
}

// #284: a first publication — details only on the filing branch — is
// move-detail's transfer: the card records the handoff and needs no owner.
func TestPublishFirstPublicationIsMoveDetail(t *testing.T) {
	cardPath, card, detailPath, detail := seededIssue(t, "000012", "new")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, nil)
	writeRepoFile(t, r.root, detailPath, detail)
	if out, err := publish(t, 12); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(r.git("show", r.originMain()+":"+detailPath), "Seeded new") {
		t.Fatal("details not on main")
	}
	if h, ok, err := issue.CardHandoff([]byte(r.card(cardPath))); err != nil || !ok || h.MainCommit == "" {
		t.Fatalf("no handoff record: %+v %v %v", h, ok, err)
	}
}

// #284: a lost publication response asks for a rerun; the rerun finds main
// already holding the details and only finishes the checkout.
func TestPublishRerunAfterALostResponse(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	appendDetail(t, r, "workshop/issues/000009-s09.md", "Shaped nine.")
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var out, errs bytes.Buffer
	// Publish through a trunk writer whose push lands but reports a lost
	// response: the outcome an agent cannot see.
	err = republishOwnedLosing(t, env, &out, &errs)
	if err == nil || !strings.Contains(err.Error(), "rerun the same command (sdlc issue publish --issue 9)") {
		t.Fatalf("lost response: %v", err)
	}
	if !strings.Contains(r.git("show", r.originMain()+":workshop/issues/000009-s09.md"), "Shaped nine.") {
		t.Fatal("fixture: the lost push did not land")
	}
	if o, err := publish(t, 9); err != nil {
		t.Fatalf("rerun: %v\n%s", err, o)
	}
	if r.git("rev-parse", "HEAD") != r.originMain() || r.git("status", "--porcelain") != "" {
		t.Fatal("the rerun did not finish the checkout")
	}
}

// republishOwnedLosing runs republishOwned with main's push landing and its
// response lost.
func republishOwnedLosing(t *testing.T, env *trackerEnv, stdout, stderr *bytes.Buffer) error {
	t.Helper()
	prev := mainPublish
	defer func() { mainPublish = prev }()
	mainPublish = func(env *trackerEnv, msg string, prepare func(*gitx.TrunkView) (gitx.TrunkWrite, error), before func(string, string) error) error {
		if err := prev(env, msg, prepare, before); err != nil {
			return err
		}
		return fmt.Errorf("%w: push response lost", gitx.ErrPublicationUncertain)
	}
	return republishOwned(env, stdout, stderr, []string{"000009"}, "workshop/issues", false)
}

// #284: ownership is checked again against a fresh tracker read just before
// main's push — a reclaim that lands after the decision refuses the publish,
// and main is untouched.
func TestPublishRefusesAReclaimBeforeThePush(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	appendDetail(t, r, "workshop/issues/000009-s09.md", "Shaped nine.")
	before := r.originMain()
	prev := mainPublish
	t.Cleanup(func() { mainPublish = prev })
	mainPublish = func(env *trackerEnv, msg string, prepare func(*gitx.TrunkView) (gitx.TrunkWrite, error), before func(string, string) error) error {
		snap, err := env.repo.Snapshot()
		if err != nil {
			return err
		}
		card, _ := snap.Card("000009")
		them := issue.Claimant{Operator: "Them", Machine: issue.MachineFingerprint("other"), MachineName: "box2", Worktree: "/w/them", Repository: "r"}
		next, err := issue.SetCardClaimant(card.Raw, them)
		if err != nil {
			return err
		}
		if err := env.repo.UpdateCard(card, next, "peer-reclaim", func(string, string) error { return nil }); err != nil {
			return err
		}
		return prev(env, msg, prepare, before)
	}
	if out, err := publish(t, 9); err == nil || !strings.Contains(err.Error(), "only its owner republishes") {
		t.Fatalf("a reclaim before the push must refuse: %v\n%s", err, out)
	}
	if r.originMain() != before {
		t.Fatal("main was written after ownership moved")
	}
}

// #284 BR-8: main moved under a resting branch's edit. Publishing brings main
// in itself — the refusal's remedy runs from the refused state. A clean merge
// publishes both changes; a conflict leaves markers, and after they are
// resolved the rerun publishes.
func TestPublishBringsAMovedMainIn(t *testing.T) {
	t.Run("clean merge publishes", func(t *testing.T) {
		r, _ := claimSetRepo(t)
		claimFor(t, 9)
		rel := "workshop/issues/000009-s09.md"
		orig := r.git("show", "HEAD:"+rel) + "\n" // r.git trims; keep the file's final newline
		peerAdd(t, r, rel, strings.Replace(orig, "## Problem", "## Problem\n\nA peer's framing.", 1))
		writeRepoFile(t, r.root, rel, orig+"\nShaped nine.\n")
		if out, err := publish(t, 9); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
		onMain := r.git("show", r.originMain()+":"+rel)
		if !strings.Contains(onMain, "A peer's framing.") || !strings.Contains(onMain, "Shaped nine.") {
			t.Fatalf("both changes must be on main:\n%s", onMain)
		}
		if r.git("rev-parse", "HEAD") != r.originMain() || r.git("status", "--porcelain") != "" {
			t.Fatal("the rest is not left equal to main")
		}
	})
	t.Run("conflict leaves markers; resolving and rerunning publishes", func(t *testing.T) {
		r, _ := claimSetRepo(t)
		claimFor(t, 9)
		rel := "workshop/issues/000009-s09.md"
		orig := r.git("show", "HEAD:"+rel) + "\n" // r.git trims; keep the file's final newline
		peerAdd(t, r, rel, orig+"\nThe peer's ending.\n")
		writeRepoFile(t, r.root, rel, orig+"\nMy ending.\n")
		before := r.originMain()
		out, err := publish(t, 9)
		if err == nil || !strings.Contains(err.Error(), "resolve the markers, then rerun `sdlc issue publish --issue 9`") {
			t.Fatalf("a conflict must refuse with the runnable step: %v\n%s", err, out)
		}
		local := r.git("show", ":"+rel) // the index is main's copy now
		body, _ := readFileString(r.root, rel)
		if !strings.Contains(body, "<<<<<<< yours") || r.git("rev-parse", "HEAD") != before || !strings.Contains(local, "The peer's ending.") {
			t.Fatalf("markers missing or rest not at main:\n%s", body)
		}
		writeRepoFile(t, r.root, rel, orig+"\nThe peer's ending.\nMy ending.\n")
		if out, err := publish(t, 9); err != nil {
			t.Fatalf("the rerun after resolving: %v\n%s", err, out)
		}
		if onMain := r.git("show", r.originMain()+":"+rel); !strings.Contains(onMain, "My ending.") || !strings.Contains(onMain, "The peer's ending.") {
			t.Fatalf("resolved edit not on main:\n%s", onMain)
		}
	})
}

func readFileString(root, rel string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, rel))
	return string(b), err
}

// #284: publishing #9 from #10's branch commits nothing there (one issue, one
// branch, #272): the branch's copy is taken back, main has the edit.
func TestPublishFromAnotherIssuesBranchCommitsNothingThere(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	r.git("switch", "-q", "-c", "000010-s10")
	head := r.git("rev-parse", "HEAD")
	appendDetail(t, r, "workshop/issues/000009-s09.md", "Shaped from the wrong branch.")
	if out, err := publish(t, 9); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if r.git("rev-parse", "HEAD") != head || r.git("status", "--porcelain") != "" {
		t.Fatalf("another issue's branch was changed:\n%s", r.git("status", "--porcelain"))
	}
	if !strings.Contains(r.git("show", r.originMain()+":workshop/issues/000009-s09.md"), "Shaped from the wrong branch.") {
		t.Fatal("the edit did not reach main")
	}
}

// #284 BR-13: after a conflicted merge-in, a rerun without resolving the
// markers refuses; main is untouched.
func TestPublishRefusesConflictMarkers(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	rel := "workshop/issues/000009-s09.md"
	orig := r.git("show", "HEAD:"+rel) + "\n"
	peerAdd(t, r, rel, orig+"\nThe peer's ending.\n")
	writeRepoFile(t, r.root, rel, orig+"\nMy ending.\n")
	if _, err := publish(t, 9); err == nil {
		t.Fatal("fixture: the merge-in must conflict")
	}
	before := r.originMain()
	if out, err := publish(t, 9); err == nil || !strings.Contains(err.Error(), "still has merge conflict markers") {
		t.Fatalf("unresolved markers must refuse: %v\n%s", err, out)
	}
	if r.originMain() != before {
		t.Fatal("conflict markers reached main")
	}
}

// #284 BR-14: bringing main in fails at the fast-forward (the rest carries a
// commit main lacks): every edit is put back exactly as it was.
func TestPublishBringInPutsEditsBackOnFailure(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	rel := "workshop/issues/000009-s09.md"
	orig := r.git("show", "HEAD:"+rel) + "\n"
	writeRepoFile(t, r.root, "local-only.txt", "x\n")
	r.git("add", "local-only.txt")
	r.git("commit", "-qm", "a commit main lacks")
	peerAdd(t, r, rel, strings.Replace(orig, "## Problem", "## Problem\n\nA peer's framing.", 1))
	edit := orig + "\nShaped nine.\n"
	writeRepoFile(t, r.root, rel, edit)
	out, err := publish(t, 9)
	if err == nil || !strings.Contains(err.Error(), "the edits are as they were") {
		t.Fatalf("a failed bring-in must say the edits are kept: %v\n%s", err, out)
	}
	if got, _ := readFileString(r.root, rel); got != edit {
		t.Fatalf("the edit was not put back exactly:\n%s", got)
	}
}

// #284 BR-8: off a resting branch, the main-moved refusal names steps that run
// from the refused state: on the issue's own branch, commit, merge main, rerun.
func TestPublishMovedMainOnTheIssueBranchRemedy(t *testing.T) {
	r, _ := claimSetRepo(t)
	claimFor(t, 9)
	rel := "workshop/issues/000009-s09.md"
	r.git("switch", "-q", "-c", "000009-s09")
	orig := r.git("show", "HEAD:"+rel) + "\n"
	peerAdd(t, r, rel, strings.Replace(orig, "## Problem", "## Problem\n\nA peer's framing.", 1))
	writeRepoFile(t, r.root, rel, orig+"\nShaped on the branch.\n")
	out, err := publish(t, 9)
	if err == nil || !strings.Contains(err.Error(), "Commit the edit") || !strings.Contains(err.Error(), "git merge origin/main") {
		t.Fatalf("the refusal must name the runnable steps: %v\n%s", err, out)
	}
	r.git("commit", "-qm", "#9: plan: shaped", "--", rel)
	r.git("fetch", "-q", "origin")
	r.git("merge", "-q", "--no-edit", "origin/main")
	if out, err := publish(t, 9); err != nil {
		t.Fatalf("the named steps did not lead to a publish: %v\n%s", err, out)
	}
	onMain := r.git("show", r.originMain()+":"+rel)
	if !strings.Contains(onMain, "A peer's framing.") || !strings.Contains(onMain, "Shaped on the branch.") {
		t.Fatalf("both changes must be on main:\n%s", onMain)
	}
}
