package bench

import (
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
	r, _ = apply(t, r, CancelConfirmed, "cancel")
	if r.Phase() != PhaseStopped || !r.RuntimeStopped() {
		t.Fatal(r.Phase())
	}
	if _, fx, err := Step(r, Event{ID: "late", RunID: "run-1", Kind: DispatchConfirmed, OperationID: "dispatch", Hub: h, Worker: w}); err == nil || len(fx) != 0 {
		t.Fatal("terminal resurrected")
	}
}
func TestBenchCandidateRequiresProofAndConfirmedShutdown(t *testing.T) {
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
	r, _ = apply(t, r, BeginCancel, "cancel")
	r, _ = apply(t, r, CancelConfirmed, "cancel")
	if r.Phase() != PhaseWouldLand {
		t.Fatal(r.Phase())
	}
	if !r.RuntimeStopped() || r.Candidate() != "candidate" {
		t.Fatal("retirement lost admitted candidate")
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

func TestBenchStepDoesNotMutateItsInput(t *testing.T) {
	r := runningRun(t)
	next, _ := apply(t, r, Timeout, "cancel")
	if r.Phase() != PhaseRunning || next.Phase() != PhaseCancelling {
		t.Fatal("Step mutated input")
	}
	if len(r.seen)+1 != len(next.seen) {
		t.Fatal("event authority unexpectedly shared")
	}
}

// Independent invariants over short adversarial sequences: receipts alone may
// never launch work, unknown outcomes never prove shutdown, and
// erroneous events leave the original state available for reconciliation.
func TestBenchFiniteSequencesPreserveAuthority(t *testing.T) {
	h, w := identities()
	kinds := []EventKind{BeginDispatch, DispatchConfirmed, DispatchUnconfirmed, Observe, ProposeWouldLand, ProofAccepted, Timeout, Stop, CancelConfirmed, CancelUnconfirmed}
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

			}
			if r.Candidate() != "" && next.Candidate() != r.Candidate() {
				t.Fatal("candidate was rewritten")
			}
			walk(next, depth-1)
		}
	}
	walk(preparedRun(t), 4)
}
