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
