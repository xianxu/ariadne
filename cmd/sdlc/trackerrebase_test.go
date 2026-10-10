package main

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// closedThenRebased closes issue id (SHIP), lets main move, and rebases the issue
// branch onto it: the evidence commit is rewritten off the branch (#304 B3).
func closedThenRebased(t *testing.T, id int) (*trackerRepo, string, issue.Completion) {
	t.Helper()
	r, cardPath, _ := closeReady(t, id)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", itoa(id), "--verified", "e2e", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("close: %v\n%s", err, stderr)
	}
	b, ok, _ := issue.CardCompletion([]byte(r.card(cardPath)))
	if !ok {
		t.Fatal("close bound nothing")
	}
	if msg := r.git("log", "-1", "--format=%B", b.EvidenceCommit); !strings.Contains(msg, "Close-Token: "+b.Token) {
		t.Fatalf("the evidence message must carry the binding token:\n%s", msg)
	}
	peerCommit(t, r, "other.go")
	r.git("fetch", "-q", "origin")
	r.git("rebase", "-q", "origin/main")
	if gitSucceeds(r.root, "merge-base", "--is-ancestor", b.EvidenceCommit, "HEAD") {
		t.Fatal("fixture: the rebase did not rewrite the evidence commit")
	}
	return r, cardPath, b
}

func ownedIDs(t *testing.T) []string {
	t.Helper()
	owned, _, err := branchOwnedCompletions(context.Background(), "workshop/issues")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, oc := range owned {
		ids = append(ids, oc.ID)
	}
	return ids
}

// #304 B3: a rebase used to orphan the close — ownedCompletions dropped it and the
// publish gate reported "nothing to verify". The Close-Token keeps it owned and gated.
func TestRebasedCloseIsStillOwnedAndGated(t *testing.T) {
	r, _, _ := closedThenRebased(t, 330)
	if ids := ownedIDs(t); len(ids) != 1 || ids[0] != "000330" {
		t.Fatalf("the rebased close must still be the branch's: %v", ids)
	}
	if err := runPublishGate(context.Background(), "origin/main", "workshop/issues", io.Discard); err != nil {
		t.Fatalf("an unchanged rebased close must publish: %v", err)
	}
	writeRepoFile(t, r.root, "cmd/late.go", "package a\n")
	r.git("add", "cmd/late.go")
	r.git("commit", "-qm", "#330: late")
	err := runPublishGate(context.Background(), "origin/main", "workshop/issues", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "cmd/late.go") {
		t.Fatalf("post-rebase code must refuse, naming it (it used to fail open): %v", err)
	}
}

// A squash drops the token: the close is no longer carried, and the publish refuses
// LOUDLY rather than skipping the card.
func TestSquashedCloseRefusesLoudly(t *testing.T) {
	r, _, _ := closedThenRebased(t, 331)
	branch := r.git("branch", "--show-current")
	r.git("switch", "-q", "-c", "squashed", "origin/main")
	r.git("merge", "-q", "--squash", branch)
	r.git("commit", "-qm", "#331: everything")
	if ids := ownedIDs(t); len(ids) != 0 {
		t.Fatalf("a squashed close carries no token: %v", ids)
	}
	err := runPublishGate(context.Background(), "origin/main", "workshop/issues", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "#331 is codecomplete but this branch carries no close") {
		t.Fatalf("want the loud unowned refusal, got: %v", err)
	}
	assertGatesigAttributes(t, err.Error(), "no-judge", "merge", "push")
}

// Rebase → land → done, for a merge-commit landing, a squash landing, and the settle
// path (no PR: the rebased evidence reached main).
func TestRebasedCloseLandsToDone(t *testing.T) {
	land := func(t *testing.T, squash bool) {
		r, cardPath, _ := closedThenRebased(t, 332)
		head, branch := r.git("rev-parse", "HEAD"), r.git("branch", "--show-current")
		base := r.git("merge-base", "HEAD", "origin/main")
		merge := head
		if squash {
			peer := t.TempDir()
			testfix.Git(t, "", "clone", "-q", r.origin, peer)
			testfix.Git(t, peer, "fetch", "-q", "origin", branch)
			testfix.Git(t, peer, "-c", "user.name=g", "-c", "user.email=g@g", "merge", "-q", "--squash", "FETCH_HEAD")
			testfix.Git(t, peer, "-c", "user.name=g", "-c", "user.email=g@g", "commit", "-qm", "Squashed PR")
			testfix.Git(t, peer, "push", "-q", "origin", "main")
			merge = strings.TrimSpace(testfix.Capture(t, peer, "rev-parse", "HEAD"))
		} else {
			r.git("push", "-q", "origin", "HEAD:main")
		}
		pr := landingPR{Number: 332, State: "MERGED", Repo: "test/repo", HeadRef: branch, HeadOID: head, BaseRef: "main", BaseOID: base, MergeOID: merge}
		if err := completeLandingPR(context.Background(), r.root, "workshop/issues", pr); err != nil {
			t.Fatal(err)
		}
		if card := r.card(cardPath); !strings.Contains(card, "status: done") {
			t.Fatalf("a landed rebased close must complete the card:\n%s", card)
		}
		if pins := r.git("for-each-ref", "refs/sdlc/reviewed/000332/"); pins != "" {
			t.Fatalf("done must end the issue's reviewed-head pins (#304 D4):\n%s", pins)
		}
	}
	t.Run("merge commit", func(t *testing.T) { land(t, false) })
	t.Run("squash", func(t *testing.T) { land(t, true) })
	t.Run("settle", func(t *testing.T) {
		r, cardPath, _ := closedThenRebased(t, 333)
		r.git("push", "-q", "origin", "HEAD:main")
		// A pin whose issue has no card (landed and settled elsewhere, or archived by
		// hand) is swept by settle; the per-site unpins never saw it.
		r.git("update-ref", "refs/sdlc/reviewed/000999/close", "HEAD")
		env := boundaryEnv(t)
		if _, err := settleLandedCompletions(context.Background(), env, "workshop/issues"); err != nil {
			t.Fatal(err)
		}
		if card := r.card(cardPath); !strings.Contains(card, "status: done") {
			t.Fatalf("settle must find the rebased close on main by its token:\n%s", card)
		}
		if pins := r.git("for-each-ref", "refs/sdlc/reviewed/"); pins != "" {
			t.Fatalf("settle must sweep pins of done or cardless issues:\n%s", pins)
		}
	})
}

// #283/#301 supersession: an older close's token never claims a newer close. A branch
// carrying only the first close, after a reopen-and-close bound the card to a second
// token, owns nothing — and the publish refuses loudly.
func TestOlderCloseTokenNeverClaimsANewerClose(t *testing.T) {
	r, cardPath, first := closedThenRebased(t, 334)
	firstRebased := r.git("rev-parse", "HEAD")
	if _, stderr, err := executeSDLCTestCommand("issue", "set-status", "working", "--issue", "334"); err != nil {
		t.Fatalf("reopen: %v\n%s", err, stderr)
	}
	writeRepoFile(t, r.root, "cmd/b.go", "package a\n")
	r.git("add", "cmd/b.go")
	r.git("commit", "-qm", "#334: follow-up")
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	if _, stderr, err := executeSDLCTestCommand("close", "--issue", "334", "--verified", "again", "--actual", "1", "--no-atlas"); err != nil {
		t.Fatalf("re-close: %v\n%s", err, stderr)
	}
	second, _, _ := issue.CardCompletion([]byte(r.card(cardPath)))
	if second.Token == first.Token {
		t.Fatal("fixture: the re-close must mint a new token")
	}
	r.git("reset", "-q", "--hard", firstRebased)
	if ids := ownedIDs(t); len(ids) != 0 {
		t.Fatalf("the first close's token claimed the second close: %v", ids)
	}
	if err := runPublishGate(context.Background(), "origin/main", "workshop/issues", io.Discard); err == nil {
		t.Fatal("a branch carrying only a superseded close must refuse")
	}
}

// The done-site unpin, alone: completeOnCard ends the pins itself, without the settle
// sweep that otherwise follows it and would mask a missing unpin (lesson #286 BR-2).
func TestCompleteOnCardEndsPins(t *testing.T) {
	r, _, _ := closedThenRebased(t, 335)
	owned, env, err := branchOwnedCompletions(context.Background(), "workshop/issues")
	if err != nil || len(owned) != 1 {
		t.Fatalf("owned: %v %v", owned, err)
	}
	if pins := r.git("for-each-ref", "refs/sdlc/reviewed/000335/"); pins == "" {
		t.Fatal("fixture: the close must have pinned its reviewed head")
	}
	if err := completeOnCard(env, owned[0], r.git("rev-parse", "HEAD")); err != nil {
		t.Fatal(err)
	}
	if pins := r.git("for-each-ref", "refs/sdlc/reviewed/000335/"); pins != "" {
		t.Fatalf("completeOnCard must end the pins itself:\n%s", pins)
	}
}
