package bench

import (
	"encoding/json"
	"fmt"
	"testing"
)

func identities() (RuntimeIdentity, RuntimeIdentity) {
	h := RuntimeIdentity{Slot: "ops-bench:0", Session: "hub-session", Generation: "1", Root: "/run/hub", Agent: "claude", Model: "opus", Effort: "high"}
	w := RuntimeIdentity{Slot: "work:0", Session: "worker-session", Generation: "2", Root: "/run/work", Agent: "codex", Model: "sol", Effort: "medium"}
	return h, w
}
func apply(t *testing.T, r Run, kind EventKind, op string) (Run, []Effect) {
	t.Helper()
	h, w := identities()
	next, effects, err := Step(r, Event{ID: string(kind) + op, RunID: "run-1", Kind: kind, OperationID: op, Hub: h, Worker: w})
	if err != nil {
		t.Fatal(err)
	}
	return next, effects
}
func preparedRun(t *testing.T) Run {
	t.Helper()
	r, err := NewRun("run-1")
	if err != nil {
		t.Fatal(err)
	}
	r, e := apply(t, r, BeginPrepare, "prepare")
	if len(e) != 1 || e[0].Kind != Prepare {
		t.Fatal(e)
	}
	r, _ = apply(t, r, Prepared, "prepare")
	return r
}
func runningRun(t *testing.T) Run {
	t.Helper()
	r := preparedRun(t)
	r, _ = apply(t, r, BeginDispatch, "dispatch")
	r, _ = apply(t, r, DispatchConfirmed, "dispatch")
	return r
}

func TestBenchPreparedRunCanStopWithoutInventingRuntimeBindings(t *testing.T) {
	r := preparedRun(t)
	n, effects, err := Step(r, Event{ID: "cancel-before-dispatch", RunID: r.ID(), Kind: Stop, OperationID: "cancel", Reason: "no dispatch authorized; prepared roots independently retired"})
	if err != nil || len(effects) != 0 || n.Phase() != PhaseStopped || !n.RuntimeStopped() {
		t.Fatalf("prepared cancellation: %s quiet=%t %v %v", n.Phase(), n.RuntimeStopped(), effects, err)
	}
	h, w := n.Bindings()
	if h != (RuntimeIdentity{}) || w != (RuntimeIdentity{}) {
		t.Fatal("unlaunched cancellation invented actors")
	}
}

func TestBenchUnavailableAcceptanceIsTerminalForRun(t *testing.T) {
	r := runningRun(t)
	r, _ = apply(t, r, ProposeWouldLand, "proof")
	h, w := identities()
	var err error
	r, _, err = Step(r, Event{ID: "accepted", RunID: r.ID(), Kind: ProofAccepted, OperationID: "proof", Hub: h, Worker: w, Facts: admittedFacts()})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range []struct {
		kind EventKind
		op   string
	}{{BeginCancel, "cancel"}, {CancelConfirmed, "cancel"}, {BeginScore, "score"}, {ScoreUnavailable, "score"}} {
		r, _ = apply(t, r, e.kind, e.op)
	}
	if _, fx, err := Step(r, Event{ID: "score-again", RunID: r.ID(), Kind: BeginScore, OperationID: "new-score", Hub: h, Worker: w}); err == nil || len(fx) != 0 {
		t.Fatal("confirmed unavailable acceptance was re-executed")
	}
}
func TestBenchUnknownDeliveryNeverRedispatches(t *testing.T) {
	r := preparedRun(t)
	r, e := apply(t, r, BeginDispatch, "dispatch")
	if e[0].Kind != Dispatch {
		t.Fatal(e)
	}
	r, _ = apply(t, r, DispatchUnconfirmed, "dispatch")
	h, w := identities()
	_, e, err := Step(r, Event{ID: "retry", RunID: "run-1", Kind: BeginDispatch, OperationID: "second", Hub: h, Worker: w})
	if err == nil || len(e) != 0 {
		t.Fatal("unknown dispatch retried", e, err)
	}
	r, e = apply(t, r, Observe, "dispatch")
	if e[0].Kind != Watch || r.Phase() != PhaseDispatchUnknown {
		t.Fatal(r.Phase(), e)
	}
	r, _ = apply(t, r, DispatchConfirmed, "dispatch")
	if r.Phase() != PhaseRunning {
		t.Fatal(r.Phase())
	}
}
func TestBenchExactIdentityDuplicatesAndTerminalMonotonicity(t *testing.T) {
	r := runningRun(t)
	h, w := identities()
	for _, bad := range []Event{{ID: "foreign", RunID: "other", Kind: Observe, OperationID: "dispatch", Hub: h, Worker: w}, {ID: "generation", RunID: "run-1", Kind: Observe, OperationID: "dispatch", Hub: h, Worker: RuntimeIdentity{Slot: w.Slot}}, {ID: "operation", RunID: "run-1", Kind: Observe, OperationID: "other", Hub: h, Worker: w}} {
		if _, fx, err := Step(r, bad); err == nil || len(fx) != 0 {
			t.Fatal("foreign fact accepted")
		}
	}
	e := Event{ID: "timeout", RunID: "run-1", Kind: Timeout, OperationID: "cancel", Hub: h, Worker: w}
	next, fx, err := Step(r, e)
	if err != nil || len(fx) != 1 || fx[0].Kind != Cancel {
		t.Fatal(fx, err)
	}
	r = next
	_, fx, err = Step(r, e)
	if err != nil || len(fx) != 0 {
		t.Fatal("duplicate repeated effect")
	}
	e.Kind = Stop
	if _, _, err = Step(r, e); err == nil {
		t.Fatal("conflicting duplicate accepted")
	}
	r, _ = apply(t, r, CancelUnconfirmed, "cancel")
	if r.Phase() != PhaseReconciliation || r.RuntimeStopped() {
		t.Fatal(r.Phase())
	}
	if _, _, err := Step(r, Event{ID: "unsafe-score", RunID: "run-1", Kind: BeginScore, OperationID: "score", Hub: h, Worker: w}); err == nil {
		t.Fatal("scored uncertain runtime")
	}
	r, _ = apply(t, r, CancelConfirmed, "cancel")
	if r.Phase() != PhaseStopped || !r.RuntimeStopped() {
		t.Fatal(r.Phase())
	}
	if _, fx, err := Step(r, Event{ID: "late", RunID: "run-1", Kind: DispatchConfirmed, OperationID: "dispatch", Hub: h, Worker: w}); err == nil || len(fx) != 0 {
		t.Fatal("terminal resurrected")
	}
}
func TestBenchScoreAndCleanupRequireProofAndConfirmedShutdown(t *testing.T) {
	r := runningRun(t)
	h, w := identities()
	r, fx := apply(t, r, ProposeWouldLand, "proof")
	if fx[0].Kind != CheckProof {
		t.Fatal(fx)
	}
	e := Event{ID: "proof-valid", RunID: "run-1", Kind: ProofAccepted, OperationID: "proof", Hub: h, Worker: w, Facts: admittedFacts()}
	var err error
	r, _, err = Step(r, e)
	if err != nil || r.Candidate() != "candidate" {
		t.Fatal(err)
	}
	if _, _, err = Step(r, Event{ID: "early-score", RunID: "run-1", Kind: BeginScore, OperationID: "score", Hub: h, Worker: w}); err == nil {
		t.Fatal("scored before shutdown")
	}
	r, _ = apply(t, r, BeginCancel, "cancel")
	r, _ = apply(t, r, CancelConfirmed, "cancel")
	if r.Phase() != PhaseWouldLand {
		t.Fatal(r.Phase())
	}
	r, fx = apply(t, r, BeginScore, "score")
	if fx[0].Kind != Score || fx[0].Candidate != "candidate" {
		t.Fatal(fx)
	}
	r, _ = apply(t, r, ScoreCompleted, "score")
	if r.CanClean() {
		t.Fatal("clean before recording")
	}
	r, fx = apply(t, r, BeginRecord, "record")
	if fx[0].Kind != Record {
		t.Fatal(fx)
	}
	r, _ = apply(t, r, RecordConfirmed, "record")
	if !r.CanClean() {
		t.Fatal("confirmed recorded run cannot clean")
	}
	_, fx = apply(t, r, BeginClean, "clean")
	if fx[0].Kind != Clean {
		t.Fatal(fx)
	}
}

func TestBenchCleanupEffectAdmittedOnce(t *testing.T) {
	r := runningRun(t)
	r, _ = apply(t, r, Stop, "cancel")
	r, _ = apply(t, r, CancelConfirmed, "cancel")
	r, _ = apply(t, r, BeginRecord, "record")
	r, _ = apply(t, r, RecordConfirmed, "record")
	r, _ = apply(t, r, BeginClean, "clean")
	h, w := identities()
	if _, fx, err := Step(r, Event{ID: "another-clean", RunID: "run-1", Kind: BeginClean, OperationID: "other-clean", Hub: h, Worker: w}); err == nil || len(fx) != 0 {
		t.Fatal("cleanup authorized twice")
	}
}

func TestBenchInvalidProofCannotFreezeCandidate(t *testing.T) {
	r := runningRun(t)
	r, _ = apply(t, r, ProposeWouldLand, "proof")
	h, w := identities()
	facts := admittedFacts()
	facts.ReviewVerdict = "REWORK"
	e := Event{ID: "invalid-proof", RunID: "run-1", Kind: ProofAccepted, OperationID: "proof", Hub: h, Worker: w, Facts: facts}
	next, fx, err := Step(r, e)
	if err == nil || next.Candidate() != "" || next.Phase() != PhaseProving || len(fx) != 0 {
		t.Fatal("invalid proof acquired authority")
	}
	r, _ = apply(t, r, Timeout, "cancel")
	e.Facts = admittedFacts()
	e.ID = "late-proof"
	if _, fx, err = Step(r, e); err == nil || len(fx) != 0 {
		t.Fatal("late proof undid timeout")
	}
}

func TestBenchEventReplayAndInputAreIndependent(t *testing.T) {
	h, w := identities()
	events := []Event{
		{ID: "prepare", RunID: "run-1", Kind: BeginPrepare, OperationID: "p"},
		{ID: "prepared", RunID: "run-1", Kind: Prepared, OperationID: "p"},
		{ID: "dispatch", RunID: "run-1", Kind: BeginDispatch, OperationID: "d", Hub: h, Worker: w},
		{ID: "lost", RunID: "run-1", Kind: DispatchUnconfirmed, OperationID: "d", Hub: h, Worker: w},
		{ID: "confirmed", RunID: "run-1", Kind: DispatchConfirmed, OperationID: "d", Hub: h, Worker: w},
	}
	r, err := NewRun("run-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		previous := r
		next, _, err := Step(r, event)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := previous.seen[event.ID]; ok {
			t.Fatal("Step mutated its input")
		}
		r = next
	}
	var decoded []Event
	bytes, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatal(err)
	}
	replay, _ := NewRun("run-1")
	for _, event := range decoded {
		var err error
		replay, _, err = Step(replay, event)
		if err != nil {
			t.Fatal(err)
		}
	}
	if replay.Phase() != r.Phase() || replay.DispatchOperation() != r.DispatchOperation() {
		t.Fatal("replay lost authority")
	}
}

// Independent invariants over short adversarial sequences: receipts alone may
// never launch work, unknown outcomes never enable scoring or cleanup, and
// erroneous events leave the original state available for reconciliation.
func TestBenchFiniteSequencesPreserveAuthority(t *testing.T) {
	h, w := identities()
	kinds := []EventKind{BeginDispatch, DispatchConfirmed, DispatchUnconfirmed, Observe, ProposeWouldLand, ProofAccepted, Timeout, Stop, CancelConfirmed, CancelUnconfirmed, BeginScore, ScoreCompleted, BeginRecord, RecordConfirmed, BeginClean}
	var walk func(Run, int)
	walk = func(r Run, depth int) {
		if depth == 0 {
			return
		}
		for index, kind := range kinds {
			op := "dispatch"
			if kind == Timeout || kind == Stop || kind == CancelConfirmed || kind == CancelUnconfirmed {
				op = "cancel"
			}
			e := Event{ID: fmt.Sprintf("%d-%d", depth, index), RunID: "run-1", Kind: kind, OperationID: op, Hub: h, Worker: w, Facts: admittedFacts()}
			next, fx, err := Step(r, e)
			if err != nil {
				if next.Phase() != r.Phase() || next.Candidate() != r.Candidate() || len(fx) != 0 {
					t.Fatal("rejected event changed authority")
				}
				continue
			}
			for _, effect := range fx {
				if effect.Kind == Dispatch && r.Phase() != PhasePrepared {
					t.Fatal("dispatch without preparation")
				}
				if effect.Kind == Score && (!r.RuntimeStopped() || r.Candidate() == "") {
					t.Fatal("score without proof and shutdown")
				}
				if effect.Kind == Clean && !r.CanClean() {
					t.Fatal("cleanup without recorded safe terminal state")
				}
			}
			if r.Candidate() != "" && next.Candidate() != r.Candidate() {
				t.Fatal("candidate was rewritten")
			}
			walk(next, depth-1)
		}
	}
	walk(preparedRun(t), 4)
}
