package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func reclaimCard(t *testing.T, status string, owner *issue.Claimant) []byte {
	t.Helper()
	raw := []byte("---\nid: 000031\nstatus: " + status + "\ncreated: 2026-10-01\nupdated: 2026-10-01\n" +
		map[bool]string{true: "actual_hours: 1\n", false: ""}[status == "codecomplete" || status == "done"] +
		"---\n\n# t\n\n## Problem\n\nx\n")
	if owner != nil {
		var err error
		if raw, err = issue.SetCardClaimant(raw, *owner); err != nil {
			t.Fatal(err)
		}
	}
	return raw
}

// #278: reclaim's decision over status × ownership × expect × reason. Only an
// owned card of another workspace, inspected at this revision, with a one-line
// reason, transfers; the owner's repeat is the no-op that reconciles a retry.
func TestReclaimDecision(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m2"), MachineName: "box2", Worktree: "/w/new", Repository: "r"}
	old := issue.Claimant{Operator: "Them", Machine: issue.MachineFingerprint("m1"), MachineName: "box1", Workspace: "r:1", Worktree: "/w/old", Repository: "r"}
	const rev = "1111111111111111111111111111111111111111"
	for _, c := range []struct {
		name, status   string
		owner          *issue.Claimant
		expect, reason string
		want           string // "transfer", "mine", or a refusal fragment
	}{
		{"working foreign", "working", &old, rev, "box1 died; operator moved the work", "transfer"},
		{"blocked foreign", "blocked", &old, rev, "r", "transfer"},
		{"codecomplete foreign", "codecomplete", &old, rev, "r", "transfer"},
		{"already mine (retry)", "working", &me, rev, "r", "mine"},
		{"mine, stale expect still no-op", "working", &me, "2222222222222222222222222222222222222222", "", "mine"},
		{"open", "open", nil, rev, "r", "sdlc claim --issue 31"},
		{"done", "done", &old, rev, "r", "no live responsibility"},
		{"unattributed", "working", nil, rev, "r", "--adopt"},
		{"no expect", "working", &old, "", "r", "--expect is required"},
		{"stale expect", "working", &old, "2222222222222222222222222222222222222222", "r", "changed since you inspected"},
		{"empty reason", "working", &old, rev, "  ", "--reason is required"},
		{"multi-line reason", "working", &old, rev, "a\nb", "--reason is required"},
	} {
		next, from, err := reclaimDecision(reclaimCard(t, c.status, c.owner), rev, c.expect, c.reason, me)
		switch c.want {
		case "transfer":
			got, ok, cerr := issue.CardClaimant(next)
			if err != nil || cerr != nil || !ok || got != me || from != old {
				t.Errorf("%s: want transfer from %+v, got %+v (from %+v) %v %v", c.name, old, got, from, err, cerr)
			}
		case "mine":
			if !errors.Is(err, errAlreadyMine) {
				t.Errorf("%s: want the owner's no-op, got %v", c.name, err)
			}
		default:
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("%s: want refusal %q, got %v", c.name, c.want, err)
			}
		}
	}
}

func TestReclaimTrailersRoundTrip(t *testing.T) {
	from := issue.Claimant{Operator: "Them", MachineName: "box1", Workspace: "r:1", Worktree: "/w/old"}
	to := issue.Claimant{Operator: "Me", MachineName: "box2", Worktree: "/w/new"}
	msg := "#31: tracker: update card\n\nTracker-Operation: reclaim-x\n" + strings.Join(reclaimTrailers(from, to, " box1: disk failed ✓ "), "\n")
	f, tt, reason, ok := parseReclaimTrailers(msg)
	if !ok || f != describeClaimant(from) || tt != describeClaimant(to) || reason != "box1: disk failed ✓" {
		t.Fatalf("round trip: %q %q %q %v", f, tt, reason, ok)
	}
	if _, _, _, ok := parseReclaimTrailers("#31: tracker: update card\n\nTracker-Operation: claim-x"); ok {
		t.Fatal("a claim commit parsed as a reclaim")
	}
}
