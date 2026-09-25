package tracker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// Draft is one rendered issue for a candidate ID: the card for the tracker and
// its mirrored details for the checkout. Paths are repository-relative.
type Draft struct {
	CardPath, DetailPath string
	Card, Detail         []byte
}

// Render produces the draft for an ID. It is called again with a new ID after a
// lost allocation race, so it must derive everything from the ID it is given.
type Render func(id string) (Draft, error)

// Checkout pins the caller's repository identity and branch for an operation.
type Checkout struct {
	Root         string // worktree root; details are written beneath it
	Repository   string // publication identity (gitx.PublicationTarget.Repository)
	Branch       string // refs/heads/<name>
	HEAD         string
	MainBase     string // fetched main tip the operation observed
	ObjectFormat string
}

// CreationOp adapts the creation receipt to Git: publish the card, then
// materialize the details locally. Bytes are read back from the object database
// by the OIDs the receipt pins, so a later process can resume from the receipt.
type CreationOp struct {
	ctx    context.Context
	repo   *Repository
	co     Checkout
	render Render
}

func NewCreationOp(ctx context.Context, repo *Repository, co Checkout, render Render) *CreationOp {
	return &CreationOp{ctx: ctx, repo: repo, co: co, render: render}
}

// Start allocates max(id)+1 from a fresh snapshot and returns the queued receipt.
func (op *CreationOp) Start(token string) (Receipt, error) {
	spec, err := op.allocate(token)
	if err != nil {
		return Receipt{}, err
	}
	s, err := NewCreation(spec)
	return s.receipt, err
}

func (op *CreationOp) allocate(token string) (ReceiptSpec, error) {
	snap, err := op.repo.Snapshot()
	if err != nil {
		return ReceiptSpec{}, err
	}
	id := fmt.Sprintf("%06d", snap.MaxID()+1)
	d, err := op.render(id)
	if err != nil {
		return ReceiptSpec{}, err
	}
	if err := op.checkDraft(id, d); err != nil {
		return ReceiptSpec{}, err
	}
	cardOID, err := gitx.WriteBlob(op.ctx, op.co.Root, d.Card)
	if err != nil {
		return ReceiptSpec{}, err
	}
	detailOID, err := gitx.WriteBlob(op.ctx, op.co.Root, d.Detail)
	if err != nil {
		return ReceiptSpec{}, err
	}
	return ReceiptSpec{
		Token: token, Repository: op.co.Repository, IssueID: id,
		CardPath: d.CardPath, SourcePath: d.DetailPath, DestinationPath: d.DetailPath,
		SourceBranch: op.co.Branch, SourceBase: op.co.HEAD, SourceHEAD: op.co.HEAD,
		SourceBlob: detailOID, CardOID: cardOID, TrackerBase: snap.Ref(), MainBase: op.co.MainBase,
		Source: LocalSource,
	}, nil
}

// checkDraft verifies the details mirror exactly the card they were split from.
func (op *CreationOp) checkDraft(id string, d Draft) error {
	card, err := issue.ParseCard(d.Card)
	if err != nil {
		return err
	}
	if card.ID != id || path.Base(d.CardPath) != path.Base(d.DetailPath) {
		return fmt.Errorf("draft for #%s names another issue or file", id)
	}
	baseline, err := issue.MirrorBaselineOID(d.Detail)
	if err != nil {
		return err
	}
	want, err := issue.CardBlobOID(d.Card, op.co.ObjectFormat)
	if err != nil || baseline != want {
		return errors.New("draft details do not mirror the draft card")
	}
	return nil
}

func (op *CreationOp) Prepare(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	switch e.Stage {
	case "creation.reserve":
		card, err := gitx.ReadBlob(op.ctx, op.co.Root, spec.CardOID)
		if err != nil {
			return Event{}, err
		}
		c, err := op.repo.PrepareCreate(spec.CardPath, card, spec.Token)
		if err != nil {
			return Event{}, err
		}
		return Event{Kind: EventCandidatePrepared, Binding: e.Expected, CandidateOID: c.OID, CandidateBaseOID: c.Base}, nil
	case "creation.detail":
		// A local effect's candidate is the exact detail blob it materializes.
		return Event{Kind: EventCandidatePrepared, Binding: e.Expected, CandidateOID: spec.SourceBlob}, nil
	}
	return Event{}, fmt.Errorf("creation cannot prepare stage %s", e.Stage)
}

func (op *CreationOp) Apply(e Effect, r Receipt) (Event, error) {
	switch e.Stage {
	case "creation.reserve":
		return publishCardEvent(op.repo, e, r.Spec().CardPath)
	case "creation.detail":
		return op.materialize(e, r.Spec())
	}
	return Event{}, fmt.Errorf("creation cannot apply stage %s", e.Stage)
}

func (op *CreationOp) Probe(e Effect, r Receipt) (Event, error) {
	switch e.Stage {
	case "creation.reserve":
		return probeCardEvent(op.repo, e, r.Spec().CardPath)
	case "creation.detail":
		ev, err := op.materialize(e, r.Spec())
		if err != nil {
			return Event{Kind: EventProbeUnknown, Binding: e.Expected}, err
		}
		return ev, nil
	}
	return Event{}, fmt.Errorf("creation cannot probe stage %s", e.Stage)
}

// Refresh follows a lost reservation: allocate a fresh ID on the new tip. The
// engine accepts a new identity only before the reservation is proven.
func (op *CreationOp) Refresh(e Effect, r Receipt) (Event, error) {
	next, err := op.allocate(r.Spec().Token)
	if err != nil {
		return Event{}, err
	}
	return Event{Kind: EventRevalidated, Binding: e.Expected, Replacement: &next}, nil
}

// materialize writes the details exactly once. An existing file with other
// bytes is someone's work and is never overwritten.
func (op *CreationOp) materialize(e Effect, spec ReceiptSpec) (Event, error) {
	want, err := gitx.ReadBlob(op.ctx, op.co.Root, e.CandidateOID)
	if err != nil {
		return Event{}, err
	}
	dest := filepath.Join(op.co.Root, filepath.FromSlash(spec.DestinationPath))
	info, err := os.Lstat(dest)
	switch {
	case err == nil:
		if !info.Mode().IsRegular() {
			return Event{}, fmt.Errorf("%s exists and is not an ordinary file", spec.DestinationPath)
		}
		have, err := os.ReadFile(dest)
		if err != nil {
			return Event{}, err
		}
		if !bytes.Equal(have, want) {
			return Event{}, fmt.Errorf("%s exists with other content; card #%s is reserved — reconcile the file, then `sdlc issue recovery reconcile --issue %s`", spec.DestinationPath, spec.IssueID, spec.IssueID)
		}
	case errors.Is(err, os.ErrNotExist):
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return Event{}, err
		}
		f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return Event{}, err
		}
		_, werr := f.Write(want)
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			return Event{}, werr
		}
	default:
		return Event{}, err
	}
	return Event{Kind: EventConfirmed, Binding: e.Expected, CandidateOID: e.CandidateOID}, nil
}

// publisher is one destination ref's candidate steps (tracker or main).
type publisher interface {
	Push(gitx.Candidate) (gitx.PushOutcome, error)
	Probe(gitx.Candidate) (gitx.ProbeOutcome, string, error)
}

// publishEvent pushes a prepared candidate once and reports the observation:
// accepted, a proven race, or unknown. confirm builds the Confirmed event.
func publishEvent(p publisher, e Effect, base string, confirm func() (Event, error)) (Event, error) {
	c := gitx.Candidate{Base: base, OID: e.CandidateOID}
	outcome, _ := p.Push(c)
	switch outcome {
	case gitx.PushAccepted:
		return confirm()
	case gitx.PushRejected:
		// A rejection alone is not a race: the identical candidate may already be
		// there (a lost earlier acknowledgement). Probe before calling it lost.
		return probeEvent(p, e, base, confirm)
	}
	return Event{Kind: EventUnknown, Binding: e.Expected}, nil
}

// probeEvent observes an outstanding candidate. From Apply (a remote effect)
// absence is a race; from Probe it is a completed not-applied observation.
func probeEvent(p publisher, e Effect, base string, confirm func() (Event, error)) (Event, error) {
	c := gitx.Candidate{Base: base, OID: e.CandidateOID}
	probe, _, _ := p.Probe(c)
	if probe == gitx.ProbeAbsentSame {
		// Settle a possibly delayed push by re-pushing the identical candidate
		// under the same lease: it lands once, or it already has. Then observe.
		if out, _ := p.Push(c); out == gitx.PushAccepted {
			return confirm()
		}
		probe, _, _ = p.Probe(c)
	}
	applying := e.Kind != ProbePublication
	switch probe {
	case gitx.ProbeReachable:
		return confirm()
	case gitx.ProbeAbsentMoved:
		if applying {
			return Event{Kind: EventRefRace, Binding: e.Expected}, nil
		}
		return Event{Kind: EventNotApplied, Binding: e.Expected}, nil
	}
	if applying {
		return Event{Kind: EventUnknown, Binding: e.Expected}, nil
	}
	return Event{Kind: EventProbeUnknown, Binding: e.Expected}, nil
}

func publishCardEvent(repo *Repository, e Effect, cardPath string) (Event, error) {
	return publishEvent(repo, e, e.Expected.TrackerBase, func() (Event, error) { return confirmedCard(repo, e, cardPath) })
}

func probeCardEvent(repo *Repository, e Effect, cardPath string) (Event, error) {
	return probeEvent(repo, e, e.Expected.TrackerBase, func() (Event, error) { return confirmedCard(repo, e, cardPath) })
}

func confirmedCard(repo *Repository, e Effect, cardPath string) (Event, error) {
	blob, err := repo.CardBlobAt(e.CandidateOID, cardPath)
	if err != nil {
		return Event{}, err
	}
	return Event{Kind: EventConfirmed, Binding: e.Expected, CandidateOID: e.CandidateOID, ResultCardOID: blob}, nil
}
