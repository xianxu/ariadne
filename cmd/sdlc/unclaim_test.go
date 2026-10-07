package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// #284: unclaimDecision over every status × owner {none, me, other} × release
// {none, by me, by other}. Only the owner releases, on a status that holds the
// lock; the status never moves; an unowned card this workspace released is the
// rerun's no-op.
func TestUnclaimDecision(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/a", Repository: "r"}
	other := me
	other.Operator, other.Worktree = "Them", "/w/b"
	for _, status := range vocab.Issue().AllStatuses() {
		for _, owner := range []*issue.Claimant{nil, &me, &other} {
			for _, released := range []*issue.Claimant{nil, &me, &other} {
				raw := []byte("---\nid: 000031\nstatus: " + status + "\ncreated: 2026-10-01\nupdated: 2026-10-01\nactual_hours: 1\n---\n\n# t\n\n## Problem\n\nx\n")
				var err error
				if released != nil {
					if raw, err = issue.SetCardRelease(raw, &issue.Release{By: issue.ReleasedBy(*released)}); err != nil {
						t.Fatal(err)
					}
				}
				if owner != nil {
					if raw, err = issue.SetCardClaimant(raw, *owner); err != nil {
						t.Fatal(err)
					}
				}
				name := func(c *issue.Claimant) string {
					return map[*issue.Claimant]string{nil: "none", &me: "me", &other: "other"}[c]
				}
				cell := status + " × owner " + name(owner) + " × released " + name(released)
				out, err := unclaimDecision(raw, "000031", me, "", "")
				switch {
				case !vocab.Issue().CanHoldOwner(status):
					if err == nil {
						t.Errorf("%s: released a card that holds no lock", cell)
					}
				case owner == &me:
					rel, ok, _ := issue.CardRelease(out)
					_, held, _ := issue.CardClaimant(out)
					fm, _, _ := issue.Parse(string(out))
					got, _ := issue.GetField(fm, "status")
					if err != nil || held || !ok || rel.By.Claimant() != me || got != status {
						t.Errorf("%s: want released by me, status kept: %v held=%v release=%+v status=%s", cell, err, held, rel, got)
					}
				case owner == nil && released == &me:
					if !errors.Is(err, errAlreadyReleased) {
						t.Errorf("%s: want the rerun's no-op, got %v", cell, err)
					}
				default:
					if err == nil || errors.Is(err, errAlreadyReleased) {
						t.Errorf("%s: want a refusal, got %v", cell, err)
					}
				}
			}
		}
	}
}

func unclaim(t *testing.T, note string, ids ...int) (string, error) {
	t.Helper()
	var out, errs bytes.Buffer
	err := runUnclaim(context.Background(), &out, &errs, &unclaimFlags{Issues: ids, Note: note, IssuesDir: "workshop/issues"})
	return out.String() + errs.String(), err
}

// #284: claim → shape → unclaim publishes the edits, then releases the claim in
// one tracker commit for the set; status stays open; the rest is clean and
// equal to main; the note is in each Log on main.
func TestUnclaimPublishesThenReleases(t *testing.T) {
	r, paths := claimSetRepo(t)
	claimFor(t, 9, 10)
	appendDetail(t, r, "workshop/issues/000009-s09.md", "Shaped nine.")
	tracker := func() string { return r.git("rev-parse", "refs/remotes/origin/issue-tracker") }
	r.git("fetch", "-q", "origin")
	before := tracker()
	if out, err := unclaim(t, "split #9 out; #10 needs nothing", 9, 10); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	r.git("fetch", "-q", "origin")
	if parent := r.git("rev-parse", tracker()+"^"); parent != before {
		t.Fatal("the release took more than one tracker commit")
	}
	for _, id := range []string{"000009", "000010"} {
		card := r.card(paths[id])
		if _, held, _ := issue.CardClaimant([]byte(card)); held || !strings.Contains(card, "status: open") {
			t.Fatalf("#%s not released open:\n%s", id, card)
		}
	}
	nine := r.git("show", r.originMain()+":workshop/issues/000009-s09.md")
	if !strings.Contains(nine, "Shaped nine.") || !strings.Contains(nine, "unclaimed: split #9 out; #10 needs nothing") {
		t.Fatalf("edits or note not on main:\n%s", nine)
	}
	if r.git("rev-parse", "HEAD") != r.originMain() || r.git("status", "--porcelain") != "" {
		t.Fatal("the rest is not clean and equal to main")
	}
}

// #284: another workspace's claim is refused, and so is a card nobody holds.
func TestUnclaimRefusals(t *testing.T) {
	r, paths := claimSetRepo(t)
	if out, err := unclaim(t, "", 9); err == nil || !strings.Contains(err.Error(), "nothing to release") {
		t.Fatalf("unowned: %v\n%s", err, out)
	}
	claimFor(t, 10)
	owner, _ := ownerOf(t, r, paths["000010"])
	elsewhere := owner
	elsewhere.Worktree = "/elsewhere/ariadne"
	withClaimant(t, elsewhere)
	if out, err := unclaim(t, "", 10); err == nil || !strings.Contains(err.Error(), "only its owner releases it") {
		t.Fatalf("foreign: %v\n%s", err, out)
	}
}

// #284: a lost response from the release asks for a rerun; the rerun finds the
// cards released by this workspace and only finishes.
func TestUnclaimRerunAfterALostResponse(t *testing.T) {
	r, paths := claimSetRepo(t)
	claimFor(t, 9)
	restore := loseResponses(t)
	out, err := unclaim(t, "", 9)
	restore()
	if err == nil || !strings.Contains(err.Error(), "rerun the same command (sdlc unclaim --issue 9)") {
		t.Fatalf("lost response: %v\n%s", err, out)
	}
	if out, err := unclaim(t, "", 9); err != nil || !strings.Contains(out, "already released by this workspace") {
		t.Fatalf("rerun: %v\n%s", err, out)
	}
	if _, held, _ := issue.CardClaimant([]byte(r.card(paths["000009"]))); held {
		t.Fatal("still held")
	}
}
