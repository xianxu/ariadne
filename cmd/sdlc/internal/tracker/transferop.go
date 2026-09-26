package tracker

import (
	"context"
	"errors"
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// ErrDestinationExists refuses a handoff whose details already exist on main:
// creation is complete there, and another thread may own the issue.
var ErrDestinationExists = errors.New("details already exist on main")

// Remover relinquishes the local source after the transfer is confirmed on
// main and the card: a narrow deletion commit on a feature branch, or a
// fast-forward on the resting branch. It must be idempotent — called again
// after an interruption it reports success when the removal already happened —
// and must refuse when the source no longer has the pinned bytes.
type Remover func(spec ReceiptSpec, mainCommit string) error

// TransferOp adapts the transfer receipt: record the handoff on the card, publish
// the details as a main-native commit, record that commit on the card, then
// remove the local source. Card stages re-derive from the current card, so an
// unrelated card change (a claim, once details are on main) never strands it.
type TransferOp struct {
	diagnostics
	ctx    context.Context
	repo   *Repository
	main   *gitx.TrunkFile
	root   string
	branch string // refs/heads/<name> of the checkout driving this run
	remove Remover
}

func NewTransferOp(ctx context.Context, repo *Repository, main *gitx.TrunkFile, root, branch string, remove Remover) *TransferOp {
	return &TransferOp{ctx: ctx, repo: repo, main: main, root: root, branch: branch, remove: remove}
}

// removeSource runs the remover only in the source checkout.
func (op *TransferOp) removeSource(r Receipt) error {
	spec := r.Spec()
	if err := requireSourceCheckout(spec, op.branch); err != nil {
		return err
	}
	return op.remove(spec, MainCommit(r))
}

func (op *TransferOp) handoff(spec ReceiptSpec) issue.Handoff {
	return issue.Handoff{Token: spec.Token, Repository: spec.Repository, SourceBranch: spec.SourceBranch,
		SourceBase: spec.SourceBase, SourceHEAD: spec.SourceHEAD, SourcePath: spec.SourcePath,
		SourceBlob: spec.SourceBlob, Destination: spec.DestinationPath}
}

// MainCommit is the confirmed main publication, or "" before transfer.main.
func MainCommit(r Receipt) string {
	for _, c := range r.Confirmations() {
		if c.Stage == "transfer.main" {
			return c.CandidateOID
		}
	}
	return ""
}

func (op *TransferOp) Prepare(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	var c gitx.Candidate
	var err error
	switch e.Stage {
	case "transfer.handoff":
		h := op.handoff(spec)
		c, err = op.repo.PrepareCardChange(spec.IssueID, spec.CardPath, "initial-details handoff", spec.Token, func(current []byte) ([]byte, error) {
			return issue.SetCardHandoff(current, h)
		})
	case "transfer.main":
		var detail []byte
		if detail, err = gitx.ReadBlob(op.ctx, op.root, spec.SourceBlob); err != nil {
			return Event{}, err
		}
		msg := fmt.Sprintf("#%s: issue: initial details\n\nTracker-Operation: %s\nDetail-Handoff: #%s", spec.IssueID, spec.Token, spec.IssueID)
		c, err = op.main.PrepareCandidate(msg, func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
			present, err := view.Exists(spec.DestinationPath)
			if err != nil {
				return gitx.TrunkWrite{}, err
			}
			if present {
				return gitx.TrunkWrite{}, fmt.Errorf("%w: %s", ErrDestinationExists, spec.DestinationPath)
			}
			return gitx.TrunkWrite{Write: map[string][]byte{spec.DestinationPath: detail}, ExactBytes: true}, nil
		})
	case "transfer.record":
		h := op.handoff(spec)
		h.MainCommit = MainCommit(r)
		c, err = op.repo.PrepareCardChange(spec.IssueID, spec.CardPath, "record initial-details publication", spec.Token, func(current []byte) ([]byte, error) {
			return issue.SetCardHandoff(current, h)
		})
	case "transfer.remove":
		// A local effect's candidate is the exact source blob it relinquishes.
		return Event{Kind: EventCandidatePrepared, Binding: e.Expected, CandidateOID: spec.SourceBlob}, nil
	default:
		return Event{}, fmt.Errorf("transfer cannot prepare stage %s", e.Stage)
	}
	if err != nil {
		return Event{}, err
	}
	return Event{Kind: EventCandidatePrepared, Binding: e.Expected, CandidateOID: c.OID, CandidateBaseOID: c.Base}, nil
}

func (op *TransferOp) Apply(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	switch e.Kind {
	case PublishCard:
		return publishCardEvent(op.repo, &op.diagnostics, e, spec.CardPath)
	case PublishMain:
		return publishEvent(mainPublisher{op.main}, &op.diagnostics, e, e.Expected.MainBase, op.confirmMain(e))
	case RemoveSource:
		if err := op.removeSource(r); err != nil {
			return Event{}, err
		}
		return Event{Kind: EventConfirmed, Binding: e.Expected, CandidateOID: e.CandidateOID}, nil
	}
	return Event{}, fmt.Errorf("transfer cannot apply %s", e.Kind)
}

func (op *TransferOp) Probe(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	switch e.Stage {
	case "transfer.handoff", "transfer.record":
		return probeCardEvent(op.repo, &op.diagnostics, e, spec.CardPath)
	case "transfer.main":
		return probeEvent(mainPublisher{op.main}, &op.diagnostics, e, e.Expected.MainBase, op.confirmMain(e))
	case "transfer.remove":
		// The remover is idempotent: it confirms an earlier removal or finishes it.
		if err := op.removeSource(r); err != nil {
			return Event{Kind: EventProbeUnknown, Binding: e.Expected}, err
		}
		return Event{Kind: EventConfirmed, Binding: e.Expected, CandidateOID: e.CandidateOID}, nil
	}
	return Event{}, fmt.Errorf("transfer cannot probe stage %s", e.Stage)
}

func (op *TransferOp) confirmMain(e Effect) func() (Event, error) {
	return func() (Event, error) {
		return Event{Kind: EventConfirmed, Binding: e.Expected, CandidateOID: e.CandidateOID}, nil
	}
}

// Refresh after a proven race: the intent is unchanged (the engine re-pins
// the destination base when the next candidate is prepared).
func (op *TransferOp) Refresh(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	return Event{Kind: EventRevalidated, Binding: e.Expected, Replacement: &spec}, nil
}

// mainPublisher exposes main's candidate steps through the shared publisher.
type mainPublisher struct{ t *gitx.TrunkFile }

func (m mainPublisher) Push(c gitx.Candidate) (gitx.PushOutcome, error) { return m.t.PushCandidate(c) }
func (m mainPublisher) Probe(c gitx.Candidate) (gitx.ProbeOutcome, string, error) {
	return m.t.ProbeCandidate(c)
}
