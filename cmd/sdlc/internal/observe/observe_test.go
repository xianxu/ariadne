package observe

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

var at = time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC)

func card(t testing.TB, status string, owner *issue.Claimant, landed bool) []byte {
	t.Helper()
	extra := ""
	if status == "codecomplete" || status == "done" {
		extra = "actual_hours: 1\n"
	}
	raw := []byte("---\nid: 000279\nstatus: " + status + "\ncreated: 2026-10-01\nupdated: 2026-10-02\n" + extra + "---\n\n# Observe\n\n## Problem\n\nx\n")
	var err error
	if owner != nil {
		if raw, err = issue.SetCardClaimant(raw, *owner); err != nil {
			t.Fatal(err)
		}
	}
	if status == "codecomplete" || status == "done" {
		b := issue.Completion{Token: "close-aaaaaaaaaaaa", Repository: "github.com/x/r", ReviewedHEAD: strings.Repeat("a", 40), EvidenceCommit: strings.Repeat("b", 40)}
		if landed {
			b.LandedCommit = strings.Repeat("c", 40)
		}
		if raw, err = issue.SetCardCompletion(raw, b); err != nil {
			t.Fatal(err)
		}
	}
	return raw
}

var (
	me    = issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Worktree: "/w/a", Repository: "github.com/x/r"}
	slot  = issue.Claimant{Operator: "Me", Machine: issue.MachineFingerprint("m1"), MachineName: "box", Workspace: "r:2", Worktree: "/w/slot2", Repository: "github.com/x/r"}
	other = issue.Claimant{Operator: "Them", Machine: issue.MachineFingerprint("m9"), MachineName: "far", Worktree: "/w/a", Repository: "github.com/x/r"}
)

func base(raw []byte) Inputs {
	return Inputs{Issue: "000279", ObservedAt: at, Tracked: true, TrackerRef: strings.Repeat("e", 40), Card: raw,
		CardPath: "workshop/issue-cards/000279-observe.md", CardBlob: strings.Repeat("f", 40), Me: &me,
		Worktrees: []LocalWorktree{{"/w/a", "main"}, {"/w/slot2", "000279-observe"}}}
}

// #279: the assignment and landing judgments across relation × owner worktree
// × tracker freshness. State is read quality; the value lives in its own field.
func TestAssembleTrackerSections(t *testing.T) {
	for _, c := range []struct {
		name     string
		in       Inputs
		card     State
		relation Relation
		fate     WorktreeFate
		outcome  Outcome
	}{
		{"mine", base(card(t, "working", &me, false)), Present, RelationThisWorkspace, FateElsewhere, OutcomeNotLanded},
		{"parked slot holds it", base(card(t, "working", &slot, false)), Present, RelationOtherWorkspace, FateHoldsBranch, OutcomeNotLanded},
		{"unattributed", base(card(t, "working", nil, false)), Present, RelationUnattributed, "", OutcomeNotLanded},
		{"other machine (never probed)", base(card(t, "working", &other, false)), Present, RelationOtherWorkspace, FateOtherMachine, OutcomeNotLanded},
		{"owner worktree gone", func() Inputs { in := base(card(t, "working", &slot, false)); in.Worktrees = in.Worktrees[:1]; return in }(), Present, RelationOtherWorkspace, FateMissing, OutcomeNotLanded},
		{"worktree list failed", func() Inputs {
			in := base(card(t, "working", &slot, false))
			in.WorktreesErr = errors.New("git worktree list: boom")
			return in
		}(), Present, RelationOtherWorkspace, FateUnknown, OutcomeNotLanded},
		{"landed", base(card(t, "done", &me, true)), Present, RelationThisWorkspace, FateElsewhere, OutcomeLanded},
		{"stale but landed", func() Inputs {
			in := base(card(t, "done", &me, true))
			in.TrackerStale, in.TrackerErr = true, errors.New("dial: no route")
			return in
		}(), Stale, RelationThisWorkspace, FateElsewhere, OutcomeLanded},
		{"tracker read failed", func() Inputs {
			in := base(nil)
			in.TrackerErr = errors.New("snapshot: boom")
			return in
		}(), Unknown, "", "", ""},
		{"untracked repository", func() Inputs { in := base(nil); in.Tracked = false; return in }(), Absent, "", "", ""},
		{"no such card", base(nil), Absent, "", "", ""},
		{"unreadable card", base([]byte("---\nstatus: [\n---\n")), Unknown, "", "", ""},
	} {
		o := Assemble(c.in)
		if err := o.Validate(); err != nil {
			t.Errorf("%s: invalid observation: %v", c.name, err)
			continue
		}
		if o.Card.State != c.card || o.Assignment.Relation != c.relation || o.Assignment.ClaimantWorktree != c.fate || o.Landing.Outcome != c.outcome {
			t.Errorf("%s: card %s relation %q fate %q outcome %q; want %s %q %q %q", c.name, o.Card.State, o.Assignment.Relation,
				o.Assignment.ClaimantWorktree, o.Landing.Outcome, c.card, c.relation, c.fate, c.outcome)
		}
		if c.card == Unknown && o.Card.Error == "" || c.card == Stale && !strings.Contains(o.Card.Error, "dial: no route") {
			t.Errorf("%s: the card read does not say why: %+v", c.name, o.Card.Read)
		}
	}
	landed := Assemble(base(card(t, "done", &me, true)))
	if landed.Completion.State != Present || landed.Completion.EvidenceCommit != strings.Repeat("b", 40) || landed.Landing.LandedCommit != strings.Repeat("c", 40) {
		t.Fatalf("completion/landing values: %+v %+v", landed.Completion, landed.Landing)
	}
	if w := Assemble(base(card(t, "working", &me, false))); w.Completion.State != Absent {
		t.Fatalf("a working card has no completion: %+v", w.Completion)
	}
}

func TestObservationJSONIsStrict(t *testing.T) {
	o := Assemble(base(card(t, "done", &slot, true)))
	raw, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var back Observation
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("round trip: %v\n%s", err, raw)
	}
	again, _ := json.Marshal(back)
	if string(again) != string(raw) {
		t.Fatalf("round trip changed the document:\n%s\n%s", raw, again)
	}
	for name, doc := range map[string]string{
		"unknown key":   strings.Replace(string(raw), `"issue":`, `"surprise":1,"issue":`, 1),
		"duplicate key": strings.Replace(string(raw), `"issue":"000279"`, `"issue":"000279","issue":"000279"`, 1),
		"wrong version": strings.Replace(string(raw), `"schema_version":1`, `"schema_version":2`, 1),
		"unknown state": strings.Replace(string(raw), `"state":"present"`, `"state":"fine"`, 1),
		"null holding":  strings.Replace(string(raw), `"holding":[]`, `"holding":null`, 1),
	} {
		if doc == string(raw) {
			t.Fatalf("%s: fixture did not change the document", name)
		}
		if err := json.Unmarshal([]byte(doc), &back); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	bad := o
	bad.Card.Read = Read{State: Unknown}
	if _, err := json.Marshal(bad); err == nil {
		t.Fatal("an unknown read without a reason was emitted")
	}
}

// Assemble never panics and always yields a valid observation, whatever the
// card bytes are.
func FuzzAssemble(f *testing.F) {
	f.Add(card(f, "working", &me, false))
	f.Add(card(f, "done", &slot, true))
	f.Add([]byte("---\nid: 000279\nstatus: working\nclaimant: {x: 1}\n---\n# t\n## Problem\n"))
	f.Add([]byte(""))
	f.Fuzz(func(t *testing.T, raw []byte) {
		for _, stale := range []bool{false, true} {
			in := base(raw)
			if stale {
				in.TrackerStale, in.TrackerErr = true, errors.New("offline")
			}
			if err := Assemble(in).Validate(); err != nil {
				t.Fatalf("invalid observation for %q (stale=%v): %v", raw, stale, err)
			}
		}
	})
}
