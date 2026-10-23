package bench

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Phase string

const (
	PhaseNew                   Phase = "new"
	PhasePreparing             Phase = "preparing"
	PhasePrepared              Phase = "prepared"
	PhaseDispatching           Phase = "dispatching"
	PhaseDispatchUnknown       Phase = "dispatch-unconfirmed"
	PhaseRunning               Phase = "running"
	PhaseProving               Phase = "evidence-required"
	PhaseWouldLand             Phase = "would-land"
	PhaseCancelling            Phase = "cancelling"
	PhaseReconciliation        Phase = "needs-reconciliation"
	PhaseStopped               Phase = "stopped"
	PhaseScoring               Phase = "scoring"
	PhaseAcceptanceUnavailable Phase = "acceptance-unavailable"
	PhaseScored                Phase = "scored"
	PhaseRecording             Phase = "recording"
	PhaseRecordUnknown         Phase = "recording-unconfirmed"
	PhaseRecorded              Phase = "recorded"
	PhaseCleaning              Phase = "cleaning"
	PhaseCleaned               Phase = "cleaned"
)

type EventKind string

const (
	BeginPrepare        EventKind = "begin-prepare"
	Prepared            EventKind = "prepared"
	PrepareFailed       EventKind = "prepare-failed"
	BeginDispatch       EventKind = "begin-dispatch"
	DispatchConfirmed   EventKind = "dispatch-confirmed"
	DispatchUnconfirmed EventKind = "dispatch-unconfirmed"
	Observe             EventKind = "observe"
	ProposeWouldLand    EventKind = "propose-would-land"
	ProofAccepted       EventKind = "proof-accepted"
	ProofRejected       EventKind = "proof-rejected"
	BeginCancel         EventKind = "begin-cancel"
	Timeout             EventKind = "timeout"
	Stop                EventKind = "stop"
	RuntimeLost         EventKind = "runtime-lost"
	CancelConfirmed     EventKind = "cancel-confirmed"
	CancelUnconfirmed   EventKind = "cancel-unconfirmed"
	BeginScore          EventKind = "begin-score"
	ScoreCompleted      EventKind = "score-completed"
	ScoreUnavailable    EventKind = "score-unavailable"
	BeginRecord         EventKind = "begin-record"
	RecordConfirmed     EventKind = "record-confirmed"
	RecordUnconfirmed   EventKind = "record-unconfirmed"
	BeginClean          EventKind = "begin-clean"
	CleanConfirmed      EventKind = "clean-confirmed"
	CleanUnconfirmed    EventKind = "clean-unconfirmed"
)

type EffectKind string

const (
	Prepare    EffectKind = "prepare"
	Dispatch   EffectKind = "dispatch"
	Watch      EffectKind = "watch"
	CheckProof EffectKind = "check-proof"
	Cancel     EffectKind = "cancel"
	Score      EffectKind = "score"
	Record     EffectKind = "record"
	Clean      EffectKind = "clean"
)

// RuntimeIdentity names one exact provisioned conversation, not just a slot.
type RuntimeIdentity struct{ Slot, Session, Generation, Root, Agent, Model, Effort string }

func (i RuntimeIdentity) Validate() error {
	for _, v := range []string{i.Slot, i.Session, i.Generation, i.Root, i.Agent, i.Model, i.Effort} {
		if strings.TrimSpace(v) == "" || strings.ContainsAny(v, "\r\n\x00") {
			return fmt.Errorf("incomplete runtime identity")
		}
	}
	if !filepath.IsAbs(i.Root) || filepath.Clean(i.Root) != i.Root {
		return fmt.Errorf("runtime root must be canonical and absolute")
	}
	return nil
}

// Event is a durable fact or requested action. Every accepted event can be
// replayed through Step, discarding historical effects, to reconstruct Run.
type Event struct {
	ID, RunID   string
	Kind        EventKind
	OperationID string
	Hub, Worker RuntimeIdentity
	Facts       WouldLandFacts
	Reason      string
}
type Effect struct {
	Kind                          EffectKind
	RunID, OperationID, Candidate string
	Hub, Worker                   RuntimeIdentity
}

// Run's authority is private: only Step can advance it. Store accepted Events,
// not an independently editable phase snapshot. Step never mutates its input.
type Run struct {
	id                                                                   string
	phase                                                                Phase
	hub, worker                                                          RuntimeIdentity
	bound, quiet                                                         bool
	candidate                                                            string
	prepareOp, dispatchOp, proofOp, cancelOp, scoreOp, recordOp, cleanOp string
	seen                                                                 map[string]Event
}

func NewRun(id string) (Run, error) {
	if strings.TrimSpace(id) == "" || len(id) > 128 || strings.ContainsAny(id, "/\\\r\n\x00") {
		return Run{}, fmt.Errorf("invalid run ID")
	}
	return Run{id: id, phase: PhaseNew, seen: map[string]Event{}}, nil
}
func (r Run) ID() string                                   { return r.id }
func (r Run) Phase() Phase                                 { return r.phase }
func (r Run) Candidate() string                            { return r.candidate }
func (r Run) Bindings() (RuntimeIdentity, RuntimeIdentity) { return r.hub, r.worker }
func (r Run) RuntimeStopped() bool                         { return r.quiet }
func (r Run) CanClean() bool                               { return r.phase == PhaseRecorded && r.quiet }
func (r Run) DispatchOperation() string                    { return r.dispatchOp }
func (r Run) CancelOperation() string                      { return r.cancelOp }
func (r Run) ProofOperation() string                       { return r.proofOp }

func Step(r Run, e Event) (Run, []Effect, error) {
	fail := func(message string) (Run, []Effect, error) {
		return r, nil, fmt.Errorf("bench %s: %s in %s", e.Kind, message, r.phase)
	}
	if r.id == "" || e.RunID != r.id || e.ID == "" || e.OperationID == "" {
		return fail("missing or mismatched run/event/operation identity")
	}
	if previous, ok := r.seen[e.ID]; ok {
		if previous != e {
			return fail("conflicting duplicate event")
		}
		return r, nil, nil
	}
	if r.bound && (e.Hub != r.hub || e.Worker != r.worker) {
		return fail("runtime identity changed; reconcile exact original actors")
	}
	n := r
	var effect EffectKind
	match := func(op string) bool { return op != "" && e.OperationID == op }
	switch e.Kind {
	case BeginPrepare:
		if r.phase != PhaseNew {
			return fail("preparation already started")
		}
		n.phase = PhasePreparing
		n.prepareOp = e.OperationID
		effect = Prepare
	case Prepared, PrepareFailed:
		if r.phase != PhasePreparing || !match(r.prepareOp) {
			return fail("unexpected preparation result")
		}
		if e.Kind == Prepared {
			n.phase = PhasePrepared
		} else {
			n.phase = PhaseStopped
			n.quiet = true
		}
	case BeginDispatch:
		if r.phase != PhasePrepared || r.dispatchOp != "" {
			return fail("dispatch already started or not prepared")
		}
		if e.Hub.Validate() != nil || e.Worker.Validate() != nil || e.Hub.Slot == e.Worker.Slot || e.Hub.Root == e.Worker.Root {
			return fail("invalid distinct runtime bindings")
		}
		n.hub, n.worker, n.bound = e.Hub, e.Worker, true
		n.phase = PhaseDispatching
		n.dispatchOp = e.OperationID
		effect = Dispatch
	case DispatchConfirmed, DispatchUnconfirmed:
		if (r.phase != PhaseDispatching && r.phase != PhaseDispatchUnknown) || !match(r.dispatchOp) {
			return fail("unexpected dispatch result")
		}
		if e.Kind == DispatchConfirmed {
			n.phase = PhaseRunning
		} else {
			n.phase = PhaseDispatchUnknown
		}
	case Observe:
		if r.phase == PhaseDispatching || r.phase == PhaseDispatchUnknown || r.phase == PhaseRunning || r.phase == PhaseProving {
			if !match(r.dispatchOp) {
				return fail("wrong dispatch observation")
			}
		} else if r.phase == PhaseCancelling || r.phase == PhaseReconciliation {
			if !match(r.cancelOp) {
				return fail("wrong cancellation observation")
			}
		} else if r.phase == PhaseRecordUnknown {
			if !match(r.recordOp) {
				return fail("wrong recording observation")
			}
		} else {
			return fail("nothing to observe")
		}
		effect = Watch
	case ProposeWouldLand:
		if r.phase != PhaseRunning {
			return fail("proposal requires running actors")
		}
		n.phase = PhaseProving
		n.proofOp = e.OperationID
		effect = CheckProof
	case ProofAccepted, ProofRejected:
		if r.phase != PhaseProving || !match(r.proofOp) {
			return fail("unexpected proof result")
		}
		if e.Kind == ProofRejected {
			n.phase = PhaseRunning
		} else {
			if err := ValidateWouldLand(e.Facts); err != nil {
				return fail(err.Error())
			}
			n.candidate = e.Facts.Candidate
			n.phase = PhaseWouldLand
		}
	case BeginCancel, Timeout, Stop, RuntimeLost:
		if r.phase == PhasePrepared && e.Kind == Stop {
			if e.Hub != (RuntimeIdentity{}) || e.Worker != (RuntimeIdentity{}) {
				return fail("undispatched stop cannot invent runtime bindings")
			}
			n.phase = PhaseStopped
			n.quiet = true // No external dispatch effect was ever authorized.
			break
		}
		if r.phase != PhaseDispatching && r.phase != PhaseDispatchUnknown && r.phase != PhaseRunning && r.phase != PhaseProving && r.phase != PhaseWouldLand {
			return fail("no cancellable run")
		}
		if r.cancelOp != "" {
			return fail("cancellation already requested")
		}
		n.cancelOp = e.OperationID
		n.phase = PhaseCancelling
		effect = Cancel
		if e.Kind == RuntimeLost {
			n.phase = PhaseReconciliation
		}
	case CancelConfirmed, CancelUnconfirmed:
		if (r.phase != PhaseCancelling && r.phase != PhaseReconciliation) || !match(r.cancelOp) {
			return fail("unexpected cancellation result")
		}
		if e.Kind == CancelUnconfirmed {
			n.phase = PhaseReconciliation
		} else {
			n.quiet = true
			n.phase = PhaseStopped
			if r.candidate != "" {
				n.phase = PhaseWouldLand
			}
		}
	case BeginScore:
		if r.phase != PhaseWouldLand || !r.quiet || r.candidate == "" {
			return fail("scoring needs admitted candidate and confirmed shutdown")
		}
		n.phase = PhaseScoring
		n.scoreOp = e.OperationID
		effect = Score
	case ScoreCompleted, ScoreUnavailable:
		if r.phase != PhaseScoring || !match(r.scoreOp) {
			return fail("unexpected scoring result")
		}
		if e.Kind == ScoreCompleted {
			n.phase = PhaseScored
		} else {
			n.phase = PhaseAcceptanceUnavailable
		}
	case BeginRecord:
		if (r.phase != PhaseScored && r.phase != PhaseStopped && r.phase != PhaseAcceptanceUnavailable) || !r.quiet {
			return fail("recording needs confirmed terminal run")
		}
		n.phase = PhaseRecording
		n.recordOp = e.OperationID
		effect = Record
	case RecordConfirmed, RecordUnconfirmed:
		if (r.phase != PhaseRecording && r.phase != PhaseRecordUnknown) || !match(r.recordOp) {
			return fail("unexpected recording result")
		}
		if e.Kind == RecordConfirmed {
			n.phase = PhaseRecorded
		} else {
			n.phase = PhaseRecordUnknown
		}
	case BeginClean:
		if !r.CanClean() {
			return fail("cleanup needs recording and confirmed shutdown")
		}
		n.phase = PhaseCleaning
		n.cleanOp = e.OperationID
		effect = Clean
	case CleanConfirmed, CleanUnconfirmed:
		if r.phase != PhaseCleaning || !match(r.cleanOp) {
			return fail("unexpected cleanup result")
		}
		if e.Kind == CleanConfirmed {
			n.phase = PhaseCleaned
		}
	default:
		return fail("unknown event kind")
	}
	n.seen = make(map[string]Event, len(r.seen)+1)
	for id, event := range r.seen {
		n.seen[id] = event
	}
	n.seen[e.ID] = e
	if effect == "" {
		return n, nil, nil
	}
	return n, []Effect{{Kind: effect, RunID: r.id, OperationID: e.OperationID, Candidate: n.candidate, Hub: n.hub, Worker: n.worker}}, nil
}
