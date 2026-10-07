package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// #284: claim's set decision over the model's status × owner product, for a
// set of two: the second card fixed open and unowned, the first varied. Any
// refusal aborts the set; the owner's repeat is skipped; both repeats is the
// owner's no-op.
func TestClaimSetDecision(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/a", Repository: "r"}
	other := me
	other.Operator, other.Worktree = "Them", "/w/b"
	card := func(id, status string, owner *issue.Claimant) tracker.Record {
		raw := []byte("---\nid: " + id + "\nstatus: " + status + "\ncreated: 2026-10-01\nupdated: 2026-10-01\nactual_hours: 1\n---\n\n# t\n\n## Problem\n\nx\n")
		if owner != nil {
			var err error
			if raw, err = issue.SetCardClaimant(raw, *owner); err != nil {
				t.Fatal(err)
			}
		}
		return tracker.Record{ID: id, Raw: raw}
	}
	ids := []string{"000031", "000032"}
	for _, status := range vocab.Issue().AllStatuses() {
		for _, owner := range []*issue.Claimant{nil, &me, &other} {
			current := map[string]tracker.Record{"000031": card("000031", status, owner), "000032": card("000032", "open", nil)}
			out, err := claimSetDecision(current, ids, "2026-10-07", "2026-10-07T09:00:00-07:00", me)
			cell := status + " × " + map[*issue.Claimant]string{nil: "none", &me: "me", &other: "other"}[owner]
			_, singleErr := claimDecision(current["000031"].Raw, 31, "2026-10-07", "2026-10-07T09:00:00-07:00", &me)
			switch {
			case errors.Is(singleErr, errAlreadyMine):
				if err != nil || len(out) != 1 || out["000032"] == nil {
					t.Errorf("%s: the owner's repeat must be skipped, the other claimed: %v %v", cell, out, err)
				}
			case singleErr != nil:
				if err == nil || out != nil {
					t.Errorf("%s: a refused member must abort the set: %v", cell, out)
				}
			default:
				if err != nil || len(out) != 2 {
					t.Errorf("%s: both must be claimed: %v %v", cell, out, err)
				}
			}
		}
	}
	both := map[string]tracker.Record{"000031": card("000031", "open", &me), "000032": card("000032", "open", &me)}
	if _, err := claimSetDecision(both, ids, "2026-10-07", "2026-10-07T09:00:00-07:00", me); !errors.Is(err, errAlreadyMine) {
		t.Fatalf("all repeats: %v", err)
	}
	if _, err := claimSetDecision(map[string]tracker.Record{"000031": card("000031", "open", nil)}, ids, "2026-10-07", "2026-10-07T09:00:00-07:00", me); err == nil {
		t.Fatal("a missing card must refuse")
	}
}

// claimSetRepo seeds #9, #10 and #11 with details on main.
func claimSetRepo(t *testing.T) (*trackerRepo, map[string]string) {
	t.Helper()
	cards, details, paths := map[string]string{}, map[string]string{}, map[string]string{}
	for _, n := range []string{"000009", "000010", "000011"} {
		cp, c, dp, d := seededIssue(t, n, "s"+n[4:])
		cards[cp], details[dp], paths[n] = c, d, cp
	}
	return newTrackerRepo(t, cards, details), paths
}

func trackerTipOf(t *testing.T, r *trackerRepo) string {
	t.Helper()
	return r.git("rev-parse", "refs/remotes/origin/issue-tracker")
}

// #284: three issues claimed in one tracker commit; all three name this workspace.
func TestClaimSetIsOneTrackerCommit(t *testing.T) {
	r, paths := claimSetRepo(t)
	r.git("fetch", "-q", "origin")
	before := trackerTipOf(t, r)
	var out, errs bytes.Buffer
	if err := runClaim(context.Background(), &out, &errs, &claimFlags{Issues: []int{9, 10, 11}, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil {
		t.Fatalf("%v\n%s", err, errs.String())
	}
	if parent := r.git("rev-parse", "refs/remotes/origin/issue-tracker^"); parent != before {
		t.Fatal("the set took more than one tracker commit")
	}
	if subject := r.git("log", "-1", "--format=%s", "refs/remotes/origin/issue-tracker"); subject != "#9,#10,#11: tracker: update cards" {
		t.Fatalf("subject %q", subject)
	}
	for id, p := range paths {
		if owner, ok := ownerOf(t, r, p); !ok || owner.Worktree != canonRoot(r.root) || !strings.Contains(r.card(p), "status: open") {
			t.Fatalf("#%s not claimed open by this workspace: %+v", id, owner)
		}
	}
}

// #284: a peer claims #10 while the set is in flight — the set fails whole,
// naming #10, and neither #9 nor #11 is claimed. A peer change to a member that
// leaves it claimable (#9's date) is re-decided on the retry, and the set lands.
func TestClaimSetRace(t *testing.T) {
	for _, named := range []bool{true, false} {
		t.Run(map[bool]string{true: "a member claimed by a peer", false: "a member changed, still claimable"}[named], func(t *testing.T) {
			r, paths := claimSetRepo(t)
			peer, err := tracker.NewRepository(context.Background(), r.root, "origin")
			if err != nil {
				t.Fatal(err)
			}
			prev := cardsPublish
			t.Cleanup(func() { cardsPublish = prev })
			cardsPublish = func(env *trackerEnv, ids []string, token string, trailers []string, decide func(map[string]tracker.Record) (map[string][]byte, error), before func(string, string) error) error {
				calls := 0
				return prev(env, ids, token, trailers, decide, func(base, candidate string) error {
					if calls++; calls == 1 {
						target := "000010"
						if !named {
							target = "000009" // the peer only touches #9's date: no ownership change
						}
						snap, err := peer.Snapshot()
						if err != nil {
							return err
						}
						rec, _ := snap.Card(target)
						next, err := issue.SetCardField(rec.Raw, "updated", "2026-01-01")
						if err != nil {
							return err
						}
						if named {
							them := issue.Claimant{Operator: "Them", Machine: issue.MachineFingerprint("other"), MachineName: "box2", Worktree: "/w/them", Repository: "r"}
							if next, err = issue.SetCardClaimant(rec.Raw, them); err != nil {
								return err
							}
						}
						if err := peer.UpdateCard(rec, next, "peer-claim", func(string, string) error { return nil }); err != nil {
							return err
						}
					}
					return before(base, candidate)
				})
			}
			var out, errs bytes.Buffer
			err = runClaim(context.Background(), &out, &errs, &claimFlags{Issues: []int{9, 10, 11}, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"})
			mine := func(id string) bool {
				owner, ok := ownerOf(t, r, paths[id])
				return ok && owner.Worktree == canonRoot(r.root)
			}
			if named {
				if err == nil || !strings.Contains(err.Error(), "#10") || mine("000009") || mine("000011") {
					t.Fatalf("a contested set must fail whole: %v (9 mine %v, 11 mine %v)", err, mine("000009"), mine("000011"))
				}
				return
			}
			if err != nil || !mine("000009") || !mine("000010") || !mine("000011") {
				t.Fatalf("an unrelated race must retry and land: %v", err)
			}
		})
	}
}

// #284: after claiming, a resting branch behind main fast-forwards to it, so
// the lock never protects a stale copy; a dirty file the fast-forward would
// overwrite leaves the rest where it is, with a warning. On an issue branch,
// details that differ from main's are named.
func TestClaimRefreshesTheCheckout(t *testing.T) {
	t.Run("rest behind main fast-forwards", func(t *testing.T) {
		r, _ := claimSetRepo(t)
		peerAdd(t, r, "workshop/issues/000009-s09.md", "edited on main by a peer\n")
		var out, errs bytes.Buffer
		if err := runClaim(context.Background(), &out, &errs, &claimFlags{Issues: []int{10}, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil {
			t.Fatalf("%v\n%s", err, errs.String())
		}
		if r.git("rev-parse", "HEAD") != r.originMain() {
			t.Fatalf("rest not fast-forwarded to main:\n%s", errs.String())
		}
	})
	t.Run("a dirty file in the way stays, warned", func(t *testing.T) {
		r, _ := claimSetRepo(t)
		peerAdd(t, r, "workshop/issues/000009-s09.md", "edited on main by a peer\n")
		head := r.git("rev-parse", "HEAD")
		writeRepoFile(t, r.root, "workshop/issues/000009-s09.md", "local shaping\n")
		var out, errs bytes.Buffer
		if err := runClaim(context.Background(), &out, &errs, &claimFlags{Issues: []int{9}, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil {
			t.Fatalf("%v\n%s", err, errs.String())
		}
		if r.git("rev-parse", "HEAD") != head || !strings.Contains(errs.String(), "not fast-forwarded") {
			t.Fatalf("a blocked fast-forward must leave the rest and warn:\n%s", errs.String())
		}
		if raw := r.git("show", ":workshop/issues/000009-s09.md"); strings.Contains(raw, "local shaping") {
			t.Fatal("fixture: the local edit was staged")
		}
	})
	t.Run("issue branch with differing details warns", func(t *testing.T) {
		r, _ := claimSetRepo(t)
		r.git("switch", "-q", "-c", "000010-s10")
		detail := r.git("show", "HEAD:workshop/issues/000010-s10.md")
		writeRepoFile(t, r.root, "workshop/issues/000010-s10.md", detail+"\nA branch-only edit.\n")
		var out, errs bytes.Buffer
		if err := runClaim(context.Background(), &out, &errs, &claimFlags{Issues: []int{10}, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil {
			t.Fatalf("%v\n%s", err, errs.String())
		}
		if !strings.Contains(errs.String(), "#10's details here differ from main's") {
			t.Fatalf("no warning:\n%s", errs.String())
		}
	})
}

// #284: a set whose publication response is lost asks for a rerun, and the
// rerun is settled by the cards; an unowned started member with no handoff
// branch is taken over within a set, its status kept.
func TestClaimSetLostResponseAndTakeover(t *testing.T) {
	r, paths := claimSetRepo(t)
	restore := loseResponses(t)
	var out, errs bytes.Buffer
	err := runClaim(context.Background(), &out, &errs, &claimFlags{Issues: []int{9, 10}, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"})
	restore()
	if err == nil || !strings.Contains(err.Error(), "rerun the same command (sdlc claim --issue 9,10)") {
		t.Fatalf("a lost response gave no rerun: %v", err)
	}
	landed := r.card(paths["000009"])
	errs.Reset()
	if err := runClaim(context.Background(), &out, &errs, &claimFlags{Issues: []int{9, 10}, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil || !strings.Contains(errs.String(), "already claimed by this workspace") || r.card(paths["000009"]) != landed {
		t.Fatalf("the rerun did not settle it: %v\n%s", err, errs.String())
	}

	env, err := openTrackerAt(context.Background(), r.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := env.repo.ChangeCard("000011", paths["000011"], "status", operationToken("set"), func(c []byte) ([]byte, error) {
		return issue.SetCardField(c, "status", "working")
	}); err != nil {
		t.Fatal(err)
	}
	errs.Reset()
	if err := runClaim(context.Background(), &out, &errs, &claimFlags{Issues: []int{9, 11}, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}); err != nil {
		t.Fatalf("an unowned started member must be taken over in a set: %v\n%s", err, errs.String())
	}
	if owner, ok := ownerOf(t, r, paths["000011"]); !ok || owner.Worktree != canonRoot(r.root) || !strings.Contains(r.card(paths["000011"]), "status: working") {
		t.Fatalf("#11 not taken over with its status kept:\n%s", r.card(paths["000011"]))
	}
}

// #284: a member handed off on a branch is taken over by fetching and checking
// that branch out, which a set cannot do: the set refuses, naming it.
func TestClaimSetRefusesAHandedOffMember(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/a", Repository: "r"}
	them := me
	them.Worktree = "/w/b"
	raw := func(id, status string) []byte {
		return []byte("---\nid: " + id + "\nstatus: " + status + "\ncreated: 2026-10-01\nupdated: 2026-10-01\n---\n\n# t\n\n## Problem\n\nx\n")
	}
	handed, err := issue.SetCardRelease(raw("000032", "working"), &issue.Release{By: issue.ReleasedBy(them), Branch: "000032-x", Head: strings.Repeat("a", 40)})
	if err != nil {
		t.Fatal(err)
	}
	current := map[string]tracker.Record{"000031": {ID: "000031", Raw: raw("000031", "open")}, "000032": {ID: "000032", Raw: handed}}
	if _, err := claimSetDecision(current, []string{"000031", "000032"}, "2026-10-07", "2026-10-07T09:00:00-07:00", me); err == nil || !strings.Contains(err.Error(), "claim it alone") {
		t.Fatalf("a handed-off member in a set: %v", err)
	}
	if _, err := claimSetDecision(current, []string{"000032"}, "2026-10-07", "2026-10-07T09:00:00-07:00", me); err != nil {
		t.Fatalf("alone, the decision takes it over (the shell fetches): %v", err)
	}
}

// #284: a claim takes the lock and spends any release that left the card open.
func TestClaimClearsARelease(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/a", Repository: "r"}
	them := me
	them.Operator, them.Worktree = "Them", "/w/b"
	raw := []byte("---\nid: 000031\nstatus: open\ncreated: 2026-10-01\nupdated: 2026-10-01\n---\n\n# t\n\n## Problem\n\nx\n")
	released, err := issue.SetCardRelease(raw, &issue.Release{By: issue.ReleasedBy(them)})
	if err != nil {
		t.Fatal(err)
	}
	out, err := claimDecision(released, 31, "2026-10-07", "2026-10-07T09:00:00-07:00", &me)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := issue.CardRelease(out); ok {
		t.Fatalf("the release survived the claim:\n%s", out)
	}
	if got, ok, _ := issue.CardClaimant(out); !ok || got != me {
		t.Fatal("not claimed")
	}
}
