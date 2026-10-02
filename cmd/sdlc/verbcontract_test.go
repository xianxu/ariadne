package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// #277: the plan's verb contract (workshop/plans/000277-claim-ownership-plan.md,
// "Verb contract") as one table — every situation × verb cell, over the pure
// decisions claim, claim --adopt and set-status → working run. A cell is either
// "stamp" (the result records me as the claimant), "mine" (the owner's no-op),
// or a fragment the refusal must carry.
func TestVerbContractTable(t *testing.T) {
	me := issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/a", Repository: "r"}
	other := me
	other.Operator, other.Worktree = "Them", "/w/b"
	card := func(status string, owner *issue.Claimant) []byte {
		raw := []byte("---\nid: 000031\nstatus: " + status + "\ncreated: 2026-10-01\nupdated: 2026-10-01\nstarted: 2026-10-01T09:00:00-07:00\n---\n\n# t\n\n## Problem\n\nx\n")
		if status == "open" {
			raw = []byte(strings.Replace(string(raw), "started: 2026-10-01T09:00:00-07:00\n", "", 1))
		}
		if owner != nil {
			var err error
			if raw, err = issue.SetCardClaimant(raw, *owner); err != nil {
				t.Fatal(err)
			}
		}
		return raw
	}
	verbs := map[string]func(raw []byte) ([]byte, error){
		"claim": func(raw []byte) ([]byte, error) {
			return claimDecision(raw, 31, "2026-10-01", "2026-10-01T09:00:00-07:00", &me)
		},
		"adopt": func(raw []byte) ([]byte, error) { return adoptDecision(raw, "000031", me) },
		// --force: the lifecycle guards are not what this table is about.
		"set-status working": func(raw []byte) ([]byte, error) {
			out, _, err := statusDecision(raw, "", "working", true, "2026-10-01", "2026-10-01T09:00:00-07:00", &me)
			return out, err
		},
	}
	for _, row := range []struct {
		situation string
		raw       []byte
		want      map[string]string
	}{
		{"open", card("open", nil), map[string]string{"claim": "stamp", "adopt": "plain `sdlc claim", "set-status working": "stamp"}},
		{"working, mine", card("working", &me), map[string]string{"claim": "mine", "adopt": "mine", "set-status working": "stamp"}},
		{"working, foreign", card("working", &other), map[string]string{"claim": "claimed by Them", "adopt": "never reassigns", "set-status working": "`sdlc reclaim`"}},
		{"working, unknown", card("working", nil), map[string]string{"claim": "--adopt", "adopt": "stamp", "set-status working": "--adopt"}},
		{"blocked, unknown (reopen)", card("blocked", nil), map[string]string{"claim": "not open", "adopt": "stamp", "set-status working": "stamp"}},
		{"blocked, foreign (reopen)", card("blocked", &other), map[string]string{"claim": "not open", "adopt": "never reassigns", "set-status working": "`sdlc reclaim`"}},
	} {
		for verb, decide := range verbs {
			want := row.want[verb]
			out, err := decide(row.raw)
			switch want {
			case "stamp":
				got, ok, cerr := issue.CardClaimant(out)
				if err != nil || cerr != nil || !ok || got != me {
					t.Errorf("%s × %s: want me stamped, got %+v %v %v", row.situation, verb, got, err, cerr)
				}
			case "mine":
				if !errors.Is(err, errAlreadyMine) {
					t.Errorf("%s × %s: want the owner's no-op, got %v", row.situation, verb, err)
				}
			default:
				if err == nil || !strings.Contains(err.Error(), want) {
					t.Errorf("%s × %s: want refusal %q, got %v", row.situation, verb, want, err)
				}
			}
		}
	}
}
