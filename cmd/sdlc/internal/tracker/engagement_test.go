package tracker

import (
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"reflect"
	"testing"
	"time"
)

func engagementTime(h int) time.Time { return time.Date(2026, 10, 1, h, 0, 0, 0, time.UTC) }
func engagementOwner(slot string) *issue.Claimant {
	return &issue.Claimant{Operator: "operator", Machine: issue.MachineFingerprint("machine"), MachineName: "host", Workspace: slot, Worktree: "/repo/" + slot, Repository: "repo"}
}

func TestDeriveEngagementsReleaseAndReclaim(t *testing.T) {
	a, b := engagementOwner("repo:1"), engagementOwner("repo:2")
	events := []OwnershipEvent{{At: engagementTime(0)}, {At: engagementTime(2), Owner: a, Operation: OpClaim, Started: engagementTime(1)}, {At: engagementTime(3), Owner: a}, {At: engagementTime(4)}, {At: engagementTime(6), Owner: a, Operation: OpClaim, Started: engagementTime(1)}, {At: engagementTime(7), Owner: b, Operation: OpReclaim, Started: engagementTime(1)}}
	got, err := DeriveEngagements(events, engagementTime(9))
	if err != nil {
		t.Fatal(err)
	}
	want := []EngagementWindow{{Start: engagementTime(1), End: engagementTime(4), Began: engagementTime(1), Owners: []issue.Claimant{*a}}, {Start: engagementTime(6), End: engagementTime(7), Began: engagementTime(6), Owners: []issue.Claimant{*a}}, {Start: engagementTime(7), End: engagementTime(9), Began: engagementTime(7), Owners: []issue.Claimant{*b}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v; want %#v", got, want)
	}
}
func TestDeriveEngagementsRelocateKeepsSourceRoots(t *testing.T) {
	a, b, c := engagementOwner("repo:1"), engagementOwner("repo:2"), engagementOwner("repo:3")
	got, err := DeriveEngagements([]OwnershipEvent{{At: engagementTime(1), Owner: a, Operation: OpClaim}, {At: engagementTime(3), Owner: b, Operation: OpRelocate}, {At: engagementTime(5), Owner: c, Operation: OpRelocate}}, engagementTime(8))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[2].Began != engagementTime(1) || !reflect.DeepEqual(got[2].Owners, []issue.Claimant{*a, *b, *c}) || got[0].End != engagementTime(3) || len(got[0].Owners) != 1 {
		t.Fatalf("relocate lost roots or aliased old windows: %#v", got)
	}
}
func TestDeriveEngagementsUntilClipsAndIgnoresLaterOwnership(t *testing.T) {
	a := engagementOwner("repo:1")
	events := []OwnershipEvent{{At: engagementTime(1), Owner: a}, {At: engagementTime(4)}, {At: engagementTime(6), Owner: a}}
	for h := 0; h <= 8; h++ {
		got, err := DeriveEngagements(events, engagementTime(h))
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range got {
			if !w.Start.Before(w.End) || w.End.After(engagementTime(h)) {
				t.Fatalf("h=%d: %#v", h, got)
			}
			if w.Start.Before(engagementTime(6)) && w.End.After(engagementTime(4)) {
				t.Fatalf("release gap attributed: %#v", got)
			}
		}
	}
}
func TestDeriveEngagementsRejectsMalformedEvidence(t *testing.T) {
	bad := engagementOwner("repo:1")
	bad.Machine = "raw"
	for name, events := range map[string][]OwnershipEvent{"time": {{Owner: engagementOwner("repo:1")}}, "order": {{At: engagementTime(2)}, {At: engagementTime(1)}}, "owner": {{At: engagementTime(1), Owner: bad}}, "future started": {{At: engagementTime(1), Owner: engagementOwner("repo:1"), Started: engagementTime(2)}}, "cross machine relocate": {{At: engagementTime(1), Owner: engagementOwner("repo:1")}, {At: engagementTime(2), Owner: &issue.Claimant{Operator: "o", Machine: issue.MachineFingerprint("other"), MachineName: "m", Worktree: "/x", Repository: "repo"}, Operation: OpRelocate}}} {
		t.Run(name, func(t *testing.T) {
			if _, err := DeriveEngagements(events, engagementTime(8)); err == nil {
				t.Fatal("accepted malformed evidence")
			}
		})
	}
}
func TestDeriveEngagementsLegacyAndCreationBound(t *testing.T) {
	a := engagementOwner("repo:1")
	for _, events := range [][]OwnershipEvent{{{At: engagementTime(3), Owner: a, Started: engagementTime(1)}}, {{At: engagementTime(2)}, {At: engagementTime(3), Owner: a, Started: engagementTime(1), Operation: OpClaim}}} {
		got, err := DeriveEngagements(events, engagementTime(5))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].Start != engagementTime(3) {
			t.Fatalf("unsupported started reused: %#v", got)
		}
	}
}

// Any valid chronological sequence must leave disjoint, positive intervals
// inside the measurement cutoff, including many release/reclaim/relocate cycles.
func FuzzDeriveEngagements(f *testing.F) {
	f.Add([]byte{1, 2, 0, 1, 3, 0, 1}, uint8(7))
	f.Fuzz(func(t *testing.T, sequence []byte, cut uint8) {
		if len(sequence) > 256 {
			sequence = sequence[:256]
		}
		base := engagementTime(0)
		events := []OwnershipEvent{{At: base}}
		var owner *issue.Claimant
		for i, b := range sequence {
			operation := ""
			switch b % 4 {
			case 0:
				owner = nil
			case 1:
				owner = engagementOwner("repo:1")
				operation = OpClaim
			case 2:
				if owner != nil {
					operation = OpRelocate
				}
				owner = engagementOwner("repo:2")
			case 3:
				owner = engagementOwner("repo:3")
				operation = OpReclaim
			}
			events = append(events, OwnershipEvent{At: base.Add(time.Duration(i+1) * time.Minute), Owner: owner, Operation: operation})
		}
		until := base.Add(time.Duration(cut) * time.Minute)
		windows, err := DeriveEngagements(events, until)
		if err != nil {
			t.Fatal(err)
		}
		var end time.Time
		for _, w := range windows {
			if !w.Start.Before(w.End) || w.End.After(until) || w.Start.Before(end) || w.Began.After(w.Start) || len(w.Owners) == 0 {
				t.Fatalf("invalid windows: %#v", windows)
			}
			end = w.End
		}
	})
}
