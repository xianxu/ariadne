package queuegate

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func testOperation() Operation {
	return Operation{SchemaVersion: 1, ID: "op-1", Workspace: "pair:5", Kinds: []string{"merge", "landing"}, Phase: "waiting", PID: 123, PIDStarted: "incarnation", StartedAt: time.Date(2026, 10, 20, 1, 0, 0, 0, time.UTC), Root: "/private/fixture", BootID: "boot"}
}

func TestOwnerCapacityAndUnknownOwnershipAreConservative(t *testing.T) {
	c := NewControl()
	for i := 0; i < MaxOwners; i++ {
		var err error
		c, err = Transition(c, Action{Kind: "pause", Owner: fmt.Sprintf("owner-%d", i)})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Transition(c, Action{Kind: "pause", Owner: "overflow"}); err == nil {
		t.Fatal("owner capacity overflow accepted")
	}
	if _, err := Transition(c, Action{Kind: "pause", Owner: "owner-1", CancelRunning: true}); err != nil {
		t.Fatal("idempotent owner rejected at capacity", err)
	}
	op := testOperation()
	op.Phase = "running"
	op.Workspace = ""
	s := BuildSnapshot(NewControl(), []Operation{op}, nil)
	if s.Complete || s.Detail == "" || s.Active[0].State != "unknown" {
		t.Fatalf("unresolved workspace %+v", s)
	}
	if err := ValidateSnapshot(s); err != nil {
		t.Fatal(err)
	}
	ops := make([]Operation, 300)
	for i := range ops {
		ops[i] = testOperation()
		ops[i].ID = fmt.Sprintf("operation-%d", i)
		ops[i].Phase = "running"
		ops[i].Workspace = strings.Repeat("w", MaxText)
	}
	s = BuildSnapshot(NewControl(), ops, nil)
	if s.Complete || s.State != "unknown" || len(s.Active) != 0 {
		t.Fatal("oversized serialized status did not fall back safely")
	}
}

func TestDirectModelRejectsInvalidUTF8WithoutChangingIdentity(t *testing.T) {
	bad := string([]byte{0xff})
	if _, err := Transition(NewControl(), Action{Kind: "pause", Owner: bad}); err == nil {
		t.Fatal("invalid UTF8 owner accepted")
	}
	if _, err := Transition(NewControl(), Action{Kind: "pause", Owner: "A", Reason: bad}); err == nil {
		t.Fatal("invalid UTF8 reason accepted")
	}
	c := NewControl()
	c.PauseOwners = []PauseOwner{{Owner: "A", Reason: bad}}
	if _, err := EncodeControl(c); err == nil {
		t.Fatal("invalid UTF8 persisted owner accepted")
	}
	for _, mutate := range []func(*Operation){func(o *Operation) { o.Workspace = bad }, func(o *Operation) { o.Head = bad }, func(o *Operation) { o.Root = bad }, func(o *Operation) { o.PIDStarted = bad }, func(o *Operation) { o.BootID = bad }, func(o *Operation) { o.Detail = bad }} {
		op := testOperation()
		mutate(&op)
		if _, err := EncodeOperation(op); err == nil {
			t.Fatalf("invalid UTF8 operation accepted %+v", op)
		}
	}
	s := BuildSnapshot(NewControl(), nil, nil)
	s.Detail = bad
	if err := ValidateSnapshot(s); err == nil {
		t.Fatal("invalid UTF8 snapshot detail accepted")
	}
	s = BuildSnapshot(NewControl(), nil, nil)
	s.Active = []Activity{{ID: "a", Workspace: bad, Kind: "suite", State: "running"}}
	if err := ValidateSnapshot(s); err == nil {
		t.Fatal("invalid UTF8 activity accepted")
	}
}

func TestOwnerTransitionsComposeAndRemainIdempotent(t *testing.T) {
	c := NewControl()
	var err error
	c.Extra = map[string]json.RawMessage{"future": json.RawMessage(`1e9999`)}
	for _, a := range []Action{{Kind: "pause", Owner: "A", Reason: "first"}, {Kind: "pause", Owner: "B", Reason: "second"}, {Kind: "pause", Owner: "A", Reason: "replacement", CancelRunning: true}, {Kind: "pause", Owner: "A"}, {Kind: "resume", Owner: "missing"}} {
		c, err = Transition(c, a)
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(c.PauseOwners) != 2 || c.PauseOwners[0].Reason != "first" || !c.PauseOwners[0].CancelRunning {
		t.Fatalf("owners: %+v", c)
	}
	before := c
	c, err = Transition(c, Action{Kind: "resume", Owner: "A"})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.PauseOwners) != 1 || c.PauseOwners[0].Owner != "B" || len(before.PauseOwners) != 2 || string(c.Extra["future"]) != "1e9999" {
		t.Fatalf("resume: %+v before %+v", c, before)
	}
	c.PauseOwners[0].Reason = "mutated"
	if before.PauseOwners[1].Reason != "second" {
		t.Fatal("transition aliases its input")
	}
	for _, a := range []Action{{Kind: "pause"}, {Kind: "resume"}, {Kind: "bad", Owner: "A"}} {
		if _, err := Transition(c, a); err == nil {
			t.Fatalf("accepted %+v", a)
		}
	}
}

func TestOperationTransitionTable(t *testing.T) {
	valid := map[string]map[string]string{"waiting": {"admit": "running", "cancel": "interrupted", "unknown": "unknown"}, "running": {"cancel": "interrupting", "finish": "finished", "unknown": "unknown"}, "interrupting": {"cancel": "interrupting", "finish": "interrupted", "interrupted": "interrupted", "unknown": "unknown"}, "unknown": {"unknown": "unknown"}}
	for _, phase := range []string{"waiting", "running", "interrupting", "unknown", "finished", "interrupted"} {
		for _, event := range []string{"admit", "cancel", "finish", "interrupted", "unknown", "bad"} {
			t.Run(phase+"/"+event, func(t *testing.T) {
				op := testOperation()
				op.Phase = phase
				got, err := TransitionOperation(op, event)
				want, ok := valid[phase][event]
				if !ok {
					if err == nil {
						t.Fatalf("accepted transition to %s", got.Phase)
					}
					return
				}
				if err != nil || got.Phase != want || op.Phase != phase {
					t.Fatalf("got %+v %v want %s", got, err, want)
				}
			})
		}
	}
}

func TestSnapshotProvesOnlyAdmittedActivity(t *testing.T) {
	op := testOperation()
	waiting := op
	waiting.ID = "waiting"
	op.Phase = "running"
	s := BuildSnapshot(NewControl(), []Operation{waiting, op}, nil)
	if s.State != "running" || !s.Complete || len(s.Active) != 2 || len(s.PauseOwners) != 0 || s.Detail != "" {
		t.Fatalf("snapshot %+v", s)
	}
	if err := ValidateSnapshot(s); err != nil {
		t.Fatal(err)
	}
	c, _ := Transition(NewControl(), Action{Kind: "pause", Owner: "A"})
	op.Phase = "interrupting"
	s = BuildSnapshot(c, []Operation{op}, nil)
	if s.State != "paused" || !s.Complete || s.Active[0].State != "interrupting" {
		t.Fatalf("snapshot %+v", s)
	}
	op.Phase = "unknown"
	op.Workspace = ""
	s = BuildSnapshot(c, []Operation{op}, []string{"old runner"})
	if s.Complete || s.Detail == "" || s.Active[0].State != "unknown" {
		t.Fatalf("snapshot %+v", s)
	}
	op.Phase = "finished"
	s = BuildSnapshot(c, []Operation{op}, nil)
	if len(s.Active) != 0 {
		t.Fatal(s)
	}
}

func TestSnapshotRejectsInvalidStatesAndBounds(t *testing.T) {
	base := BuildSnapshot(NewControl(), nil, nil)
	if !reflect.DeepEqual(base.Active, []Activity{}) || !reflect.DeepEqual(base.PauseOwners, []PauseOwner{}) {
		t.Fatal("arrays must be nonnull")
	}
	for _, mutate := range []func(*Snapshot){func(s *Snapshot) { s.State = "paused" }, func(s *Snapshot) { s.SchemaVersion = 2 }, func(s *Snapshot) { s.Complete = false }, func(s *Snapshot) { s.State = "unknown" }, func(s *Snapshot) {
		s.Active = []Activity{{ID: "x", Workspace: "pair:5", Kind: "bad", State: "running"}}
	}, func(s *Snapshot) { s.Active = []Activity{{ID: "x", Workspace: "", Kind: "suite", State: "running"}} }, func(s *Snapshot) { s.PauseOwners = nil }} {
		s := base
		mutate(&s)
		if err := ValidateSnapshot(s); err == nil {
			t.Fatalf("accepted %+v", s)
		}
	}
	ops := make([]Operation, MaxOperations+1)
	for i := range ops {
		ops[i] = testOperation()
		ops[i].Phase = "running"
	}
	s := BuildSnapshot(NewControl(), ops, nil)
	if s.Complete || s.State != "unknown" || s.Detail == "" {
		t.Fatalf("overflow %+v", s)
	}
}
