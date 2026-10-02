package observe

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
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
		{"owner worktree gone", func() Inputs {
			in := base(card(t, "working", &slot, false))
			in.Worktrees = in.Worktrees[:1]
			return in
		}(), Present, RelationOtherWorkspace, FateMissing, OutcomeNotLanded},
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

var update = flag.Bool("update", false, "rewrite golden files")

// The schema_version 1 wire format, pinned: any drift of keys, order or
// encoding fails here (rewrite deliberately with -update and bump the version
// for an incompatible change).
func TestObservationGolden(t *testing.T) {
	in := base(card(t, "done", &slot, true))
	in.MainArchive = "workshop/history/issues/000279-observe.md"
	in.Branch = BranchFacts{Ref: "refs/heads/000279-observe", Head: strings.Repeat("1", 40), LastCommitAt: "2026-10-02T08:00:00Z", AheadOfMain: 3}
	in.Holding = []HoldingFacts{{Path: "/w/slot2", Address: "r:2", Branch: "000279-observe", Head: strings.Repeat("1", 40), DirtyCount: 2, Ahead: 3}}
	in.Evidence = Evidence{Source: "refs/remotes/origin/main:workshop/history/plans", Details: details("- [x] M1 — contract\n- [x] M2 — the rest\n"),
		PlanGate: &ArtifactFacts{Found: true},
		Artifacts: map[string]ArtifactFacts{
			"M1":    {Found: true, Sidecar: sidecar("SHIP", "a..b")},
			"M2":    {Found: true, Sidecar: sidecar("FIX-THEN-SHIP", "b..c"), OpenBlocking: 1},
			"close": {Found: true, Sidecar: sidecar("SHIP", "a..d")},
		}}
	got, err := json.MarshalIndent(Assemble(in), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "observation-v1.golden.json")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, append(got, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if string(want) != string(got)+"\n" {
		t.Fatalf("schema_version 1 drifted from %s:\n%s", golden, got)
	}
	var back Observation
	if err := json.Unmarshal(want, &back); err != nil {
		t.Fatalf("the golden is not a valid observation: %v", err)
	}
}

func details(plan string) []byte {
	return []byte("---\nid: 000279\nstatus: done\nflow: {kind: full, provenance: inferred}\n---\n\n# Observe\n\n## Problem\n\nx\n\n## Plan\n\n" + plan + "\n## Log\n")
}

func sidecar(verdict, window string) string {
	return "# Boundary Review\n\n| field | value |\n|-------|-------|\n| window | " + window + " |\n| timestamp | 2026-10-02T08:30:00-07:00 |\n| verdict | " + verdict + " |\n\n## Review\n\nfine\n"
}

// #279: checkpoints read from the evidence location. Reviews follow plan order
// and close; a boundary the record says closed, without its artifact, is
// unknown, never absent; a boundary not reached is omitted.
func TestAssembleCheckpoints(t *testing.T) {
	in := base(card(t, "codecomplete", &me, false))
	in.Evidence = Evidence{Source: "refs/heads/000279-observe:workshop/plans", Details: details("- [x] M1 — a\n- [ ] M2 — b\n"),
		Artifacts: map[string]ArtifactFacts{"M1": {Found: true, Sidecar: sidecar("SHIP", "a..b"), OpenBlocking: 0}}}
	o := Assemble(in)
	if err := o.Validate(); err != nil {
		t.Fatal(err)
	}
	cp := o.Checkpoints
	if cp.State != Present || cp.Flow == nil || cp.Flow.Kind != "full" || cp.Plan.Total != 2 || cp.Plan.Ticked != 1 {
		t.Fatalf("checkpoints: %+v", cp)
	}
	got := map[string]Review{}
	for _, r := range cp.Reviews {
		got[r.Boundary] = r
	}
	if got["M1"].Verdict != "SHIP" || got["M1"].Window != "a..b" {
		t.Fatalf("M1: %+v", got["M1"])
	}
	if _, reached := got["M2"]; reached {
		t.Fatal("an unreached milestone was listed")
	}
	if got["close"].State != Unknown || !strings.Contains(got["close"].Error, "recorded as closed") {
		t.Fatalf("a codecomplete card's missing close artifact: %+v", got["close"])
	}
	// Unreadable evidence and no evidence at all.
	in.Evidence = Evidence{Source: "x", Err: errors.New("git show: boom")}
	if o := Assemble(in); o.Checkpoints.State != Unknown {
		t.Fatalf("unreadable evidence: %+v", o.Checkpoints.Read)
	}
	in.Evidence = Evidence{}
	if o := Assemble(in); o.Checkpoints.State != Unknown {
		t.Fatalf("a closed card with nowhere holding its evidence must be unknown: %+v", o.Checkpoints.Read)
	}
	working := base(card(t, "working", &me, false))
	if o := Assemble(working); o.Checkpoints.State != Absent || o.Branch.State != Absent || o.Workspaces.State != Absent {
		t.Fatalf("a fresh claim: %+v %+v %+v", o.Checkpoints.Read, o.Branch.Read, o.Workspaces.Read)
	}
}

// #279 BR-4: cards read but their tracker commit unnamed — the reason is kept.
func TestAssembleKeepsTheRefError(t *testing.T) {
	in := base(card(t, "working", &me, false))
	in.TrackerRef, in.TrackerRefErr = "", errors.New("rev-parse: bad ref")
	o := Assemble(in)
	if o.Tracker.State != Present || o.Tracker.RefError != "rev-parse: bad ref" || o.Validate() != nil {
		t.Fatalf("ref error: %+v", o.Tracker)
	}
}

// #279 BR-5: every enum-typed contract field rejects a value outside its set.
func TestValidateRejectsEveryUnknownEnum(t *testing.T) {
	good := Assemble(base(card(t, "done", &slot, true)))
	if err := good.Validate(); err != nil {
		t.Fatal(err)
	}
	review := func(r Review) func(*Observation) {
		return func(o *Observation) { o.Checkpoints.Reviews = append(o.Checkpoints.Reviews, r) }
	}
	// Each row spoils one field and must be refused BY THAT FIELD's check: the
	// error names it, so a row can't pass on some other invariant (a nil
	// collection once made every row here vacuous).
	for name, c := range map[string]struct {
		spoil func(*Observation)
		names string
	}{
		"state":                     {func(o *Observation) { o.Branch.State = "fine" }, "branch: unknown state"},
		"card authority":            {func(o *Observation) { o.Card.Authority = AuthorityWorktree }, "card: authority"},
		"workspaces authority":      {func(o *Observation) { o.Workspaces.Authority = AuthorityTracker }, "workspaces: authority"},
		"checkpoints authority":     {func(o *Observation) { o.Checkpoints.Authority = "" }, "checkpoints: authority"},
		"relation":                  {func(o *Observation) { o.Assignment.Relation = "friend" }, "unknown relation"},
		"claimant_worktree":         {func(o *Observation) { o.Assignment.ClaimantWorktree = "nearby" }, "unknown claimant_worktree"},
		"outcome":                   {func(o *Observation) { o.Landing.Outcome = "shipped" }, "unknown outcome"},
		"review boundary empty":     {review(Review{Read: Read{State: Absent}}), "review boundary"},
		"review boundary grammar":   {review(Review{Read: Read{State: Absent}, Boundary: "M"}), "review boundary"},
		"review boundary lowercase": {review(Review{Read: Read{State: Absent}, Boundary: "m1"}), "review boundary"},
		"verdict":                   {review(Review{Read: Read{State: Present}, Boundary: "close", Verdict: "MAYBE"}), "not a review verdict"},
		"flow kind":                 {func(o *Observation) { o.Checkpoints.Flow = &Flow{Kind: "huge", Provenance: "inferred"} }, "flow kind"},
		"flow provenance":           {func(o *Observation) { o.Checkpoints.Flow = &Flow{Kind: "full", Provenance: "x"} }, "flow provenance"},
	} {
		o := good
		o.Checkpoints.Reviews = append([]Review{}, good.Checkpoints.Reviews...) // non-nil, independent
		c.spoil(&o)
		if err := o.Validate(); err == nil || !strings.Contains(err.Error(), c.names) {
			t.Errorf("%s: want a refusal naming %q, got %v", name, c.names, err)
		}
	}
}

// #279 M2 BR-7: a value read from the details never falls back to its zero
// value on a failed read — the checkpoints section degrades, with the reason.
// A review artifact whose verdict is not a verdict is unknown, not a verdict.
func TestCheckpointsDegradeOnFailedReads(t *testing.T) {
	ok := Evidence{Source: "refs/heads/x:workshop/plans", Details: details("- [x] M1 — a\n"),
		Artifacts: map[string]ArtifactFacts{"M1": {Found: true, Sidecar: sidecar("SHIP", "a..b")}}}
	for name, c := range map[string]struct {
		ev   func(Evidence) Evidence
		want string
	}{
		"details unreadable": {func(e Evidence) Evidence { e.Details, e.DetailsErr = nil, errors.New("git show: boom"); return e }, "unreadable"},
		"details missing":    {func(e Evidence) Evidence { e.Details = nil; return e }, "not its details"},
		"details unparsable": {func(e Evidence) Evidence { e.Details = []byte("no frontmatter"); return e }, "do not parse"},
		"flow malformed": {func(e Evidence) Evidence {
			e.Details = []byte(strings.Replace(string(e.Details), "flow: {kind: full, provenance: inferred}", "flow: {kind: huge}", 1))
			return e
		}, "flow record is malformed"},
	} {
		in := base(card(t, "working", &me, false))
		in.Evidence = c.ev(ok)
		o := Assemble(in)
		if o.Checkpoints.State != Unknown || !strings.Contains(o.Checkpoints.Error, c.want) || o.Validate() != nil {
			t.Errorf("%s: %+v", name, o.Checkpoints.Read)
		}
	}
	in := base(card(t, "codecomplete", &me, false))
	in.Evidence = ok
	in.Evidence.Artifacts = map[string]ArtifactFacts{"close": {Found: true, Sidecar: sidecar("unknown", "a..b")}}
	if rv, _ := reviewFor(Assemble(in), "close"); rv.State != Unknown || rv.Verdict != "" {
		t.Fatalf("a recorded non-verdict: %+v", rv)
	}
	quoted := sidecar("SHIP", "a..b") + "\n| verdict | REWORK |\n"
	in.Evidence.Artifacts = map[string]ArtifactFacts{"close": {Found: true, Sidecar: quoted}}
	if rv, _ := reviewFor(Assemble(in), "close"); rv.Verdict != "SHIP" {
		t.Fatalf("a table quoted in the review body overrode the metadata: %+v", rv)
	}
}

func reviewFor(o Observation, boundary string) (Review, bool) {
	for _, r := range o.Checkpoints.Reviews {
		if r.Boundary == boundary {
			return r, true
		}
	}
	return Review{}, false
}
