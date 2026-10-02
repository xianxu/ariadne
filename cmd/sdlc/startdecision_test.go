package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// #283: start-plan's decision over the model's status × owner product. Only
// the start edge's source and target are admitted (blocked, codecomplete and
// terminal refuse whoever holds them); only the owner starts; an owned open card moves along the model's `start` edge,
// keeping an existing engagement stamp; an owned started card is unchanged.
func TestStartDecision(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/a", Repository: "r"}
	other := me
	other.Operator, other.Worktree = "Them", "/w/b"
	model := vocab.Issue()
	for _, status := range model.AllStatuses() {
		for _, owner := range []*issue.Claimant{nil, &me, &other} {
			raw := []byte("---\nid: 000031\nstatus: " + status + "\ncreated: 2026-10-01\nupdated: 2026-10-01\nstarted: 2026-10-01T09:00:00-07:00\nactual_hours: 1\n---\n\n# t\n\n## Problem\n\nx\n")
			if owner != nil {
				var err error
				if raw, err = issue.SetCardClaimant(raw, *owner); err != nil {
					t.Fatal(err)
				}
			}
			out, changed, err := startDecision(raw, "000031", me, "2026-10-02", "2026-10-02T10:00:00-07:00")
			cell := status + " × " + map[*issue.Claimant]string{nil: "none", &me: "me", &other: "other"}[owner]
			start := model.FirstTransitionForEvent("start")
			switch {
			case status != start.From && status != start.To:
				if err == nil || !strings.Contains(err.Error(), "nothing to plan") {
					t.Errorf("%s: want terminal refusal, got %v", cell, err)
				}
			case owner == nil:
				if err == nil || !strings.Contains(err.Error(), "sdlc claim --issue 31") {
					t.Errorf("%s: want claim-first refusal, got %v", cell, err)
				}
			case owner == &other:
				if err == nil || !strings.Contains(err.Error(), "owned by Them") {
					t.Errorf("%s: want the owner named, got %v", cell, err)
				}
			default:
				fm, _, perr := issue.Parse(string(out))
				got, _ := issue.GetField(fm, "status")
				started, _ := issue.GetField(fm, "started")
				want := status
				if tr := model.TransitionForEvent(status, "start"); tr != nil {
					want = tr.To
				}
				if err != nil || perr != nil || got != want || changed != (want != status) || started != "2026-10-01T09:00:00-07:00" {
					t.Errorf("%s: got status %s changed %v started %s err %v", cell, got, changed, started, errors.Join(err, perr))
				}
			}
		}
	}
	// An open card claimed before the stamp existed is stamped at start.
	raw, _ := issue.SetCardClaimant([]byte("---\nid: 000031\nstatus: open\ncreated: 2026-10-01\nupdated: 2026-10-01\n---\n\n# t\n\n## Problem\n\nx\n"), me)
	out, _, err := startDecision(raw, "000031", me, "2026-10-02", "2026-10-02T10:00:00-07:00")
	if fm, _, _ := issue.Parse(string(out)); err != nil || !strings.Contains(fm, "started: 2026-10-02T10:00:00-07:00") {
		t.Fatalf("unstamped open card not stamped at start: %v\n%s", err, out)
	}
}
