package tracker

import (
	"errors"
	"fmt"
)

// ErrOperationUncertain reports a stopped operation whose last effect has no
// completed observation. Its receipt is durable; recovery probes before replay.
var ErrOperationUncertain = errors.New("operation outcome uncertain; receipt retained for recovery")

// Adapter performs one declared effect and reports what it observed. Each
// method returns the next event for the pure engine; an error means the adapter
// could not observe anything and the operation stops with its receipt intact.
type Adapter interface {
	// Prepare builds the candidate for e.Stage (EventCandidatePrepared).
	Prepare(e Effect, r Receipt) (Event, error)
	// Apply performs the stage's mutation (Confirmed, RefRace or Unknown).
	Apply(e Effect, r Receipt) (Event, error)
	// Probe settles an uncertain candidate (Confirmed, NotApplied or ProbeUnknown).
	Probe(e Effect, r Receipt) (Event, error)
	// Refresh re-reads inputs after a race and returns EventRevalidated.
	Refresh(e Effect, r Receipt) (Event, error)
}

// ReceiptStore persists receipts durably before any effect depends on them.
type ReceiptStore interface {
	Save(Receipt) error
	Delete(Receipt) error
}

// Stepper advances one operation kind; see StepCreation/StepTransfer/StepCompletion.
type Stepper func(Receipt, Event) (Receipt, []Effect, error)

func CreationStepper(r Receipt, e Event) (Receipt, []Effect, error) {
	s, effects, err := StepCreation(Creation{r}, e)
	return s.receipt, effects, err
}
func TransferStepper(r Receipt, e Event) (Receipt, []Effect, error) {
	s, effects, err := StepTransfer(Transfer{r}, e)
	return s.receipt, effects, err
}
func CompletionStepper(r Receipt, e Event) (Receipt, []Effect, error) {
	s, effects, err := StepCompletion(Completion{r}, e)
	return s.receipt, effects, err
}

// maxDriveTicks bounds a run: every stage takes a handful of events, and the
// engine itself bounds publication retries. Exceeding it is a model bug.
const maxDriveTicks = 256

// Drive runs an operation from its current receipt until it finishes or stops.
// It is the only place effects are dispatched, so persistence always precedes
// the mutation it protects and an uncertain observation always stops the run.
func Drive(r Receipt, step Stepper, a Adapter, store ReceiptStore) (Receipt, error) {
	r, effects, err := step(r, Event{Kind: EventBegin, Binding: r.binding()})
	if err != nil {
		return r, err
	}
	for tick := 0; len(effects) > 0; tick++ {
		if tick > maxDriveTicks {
			return r, errors.New("operation exceeded its step bound")
		}
		e := effects[0]
		var event Event
		switch e.Kind {
		case PersistReceipt:
			if err := store.Save(r); err != nil {
				return r, fmt.Errorf("persist receipt: %w", err)
			}
			event = Event{Kind: EventReceiptSaved, Binding: r.binding()}
		case CleanupReceipt:
			if err := store.Delete(r); err != nil {
				return r, fmt.Errorf("release receipt: %w", err)
			}
			event = Event{Kind: EventCleanupConfirmed, Binding: r.binding()}
		case PrepareCandidate:
			event, err = a.Prepare(e, r)
		case ProbePublication:
			event, err = a.Probe(e, r)
		case RefreshInputs:
			event, err = a.Refresh(e, r)
		default:
			event, err = a.Apply(e, r)
		}
		if err != nil {
			return r, err
		}
		next, nextEffects, err := step(r, event)
		if err != nil {
			return r, err
		}
		r, effects = next, nextEffects
		if event.Kind == EventUnknown || event.Kind == EventProbeUnknown {
			if err := store.Save(r); err != nil {
				return r, fmt.Errorf("%w; persisting it also failed: %v", ErrOperationUncertain, err)
			}
			cause := ""
			if d, ok := a.(interface{ Diagnostic() error }); ok && d.Diagnostic() != nil {
				cause = "; last Git error: " + d.Diagnostic().Error()
			}
			return r, fmt.Errorf("%w at %s (#%s, operation %s)%s", ErrOperationUncertain, e.Stage, r.wire.Spec.IssueID, r.wire.Spec.Token, cause)
		}
	}
	return r, nil
}
