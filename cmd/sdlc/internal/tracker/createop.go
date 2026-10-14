package tracker

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// Draft is one rendered issue for a candidate ID: the card for the tracker and
// its mirrored details for main. Paths are repository-relative.
type Draft struct {
	CardPath, DetailPath string
	Card, Detail         []byte
}

// Render produces the draft for an ID. It is called again with a new ID after a
// lost allocation race, so it must derive everything from the ID it is given.
type Render func(id string) (Draft, error)

// Checkout pins the caller's repository identity and branch for an operation.
type Checkout struct {
	Root         string // worktree root: its object database holds the pinned blobs
	Repository   string // publication identity (gitx.PublicationTarget.Repository)
	Branch       string // refs/heads/<name>
	HEAD         string
	MainBase     string // fetched main tip the operation observed
	ObjectFormat string
}

// CreationOp adapts the creation receipt to Git: publish the card, then the
// details as a main-native commit (#308). No checkout is written. Bytes are
// read back from the object database by the OIDs the receipt pins, so a later
// process can resume from the receipt.
type CreationOp struct {
	diagnostics
	ctx    context.Context
	repo   *Repository
	main   *gitx.TrunkFile
	co     Checkout
	render Render
}

func NewCreationOp(ctx context.Context, repo *Repository, main *gitx.TrunkFile, co Checkout, render Render) *CreationOp {
	return &CreationOp{ctx: ctx, repo: repo, main: main, co: co, render: render}
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
	return validateDetailBundle(d.DetailPath, d.Detail, nil)
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
	case "creation.main":
		detail, err := op.main.ViewOf(spec.MainBase).ReadBlob(spec.SourceBlob)
		if err != nil {
			return Event{}, err
		}
		msg := fmt.Sprintf("#%s: issue: new\n\nTracker-Operation: %s", issue.CLIRef(spec.IssueID), spec.Token)
		c, err := DetailPublication(op.main, spec.DestinationPath, detail, msg)
		if err != nil {
			return Event{}, err
		}
		return Event{Kind: EventCandidatePrepared, Binding: e.Expected, CandidateOID: c.OID, CandidateBaseOID: c.Base}, nil
	}
	return Event{}, fmt.Errorf("creation cannot prepare stage %s", e.Stage)
}

func (op *CreationOp) Apply(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	if e.Stage == "creation.reserve" {
		if _, _, err := pinnedDetail(op.main.ViewOf(spec.MainBase), spec); err != nil {
			return Event{}, blockedMainReceipt(op.co.Root, spec, e.CandidateOID, err)
		}
	}

	switch e.Stage {
	case "creation.reserve":
		return publishCardEvent(op.repo, &op.diagnostics, e, r.Spec().CardPath)
	case "creation.main":
		p, err := checkedMainPublisher(op.main, op.co.Root, spec, e.CandidateOID, nil)
		if err != nil {
			return Event{}, err
		}
		return publishEvent(p, &op.diagnostics, e, e.Expected.MainBase, confirmedMain(e))
	}
	return Event{}, fmt.Errorf("creation cannot apply stage %s", e.Stage)
}

func (op *CreationOp) Probe(e Effect, r Receipt) (Event, error) {
	if e.Stage == "creation.reserve" {
		if _, _, err := pinnedDetail(op.main.ViewOf(r.Spec().MainBase), r.Spec()); err != nil {
			return Event{}, blockedMainReceipt(op.co.Root, r.Spec(), e.CandidateOID, err)
		}
	}
	switch e.Stage {
	case "creation.reserve":
		return probeCardEvent(op.repo, &op.diagnostics, e, r.Spec().CardPath)
	case "creation.main":
		p, err := checkedMainPublisher(op.main, op.co.Root, r.Spec(), e.CandidateOID, nil)
		if err != nil {
			return Event{}, err
		}
		return probeEvent(p, &op.diagnostics, e, e.Expected.MainBase, confirmedMain(e))
	}
	return Event{}, fmt.Errorf("creation cannot probe stage %s", e.Stage)
}

// Refresh follows a lost race. A lost reservation allocates a fresh ID on the
// new tip (the engine accepts a new identity only before the reservation is
// proven); a moved main keeps the intent, re-pinned at the next prepare.
func (op *CreationOp) Refresh(e Effect, r Receipt) (Event, error) {
	if e.Stage != "creation.reserve" {
		spec := r.Spec()
		return Event{Kind: EventRevalidated, Binding: e.Expected, Replacement: &spec}, nil
	}
	next, err := op.allocate(r.Spec().Token)
	if err != nil {
		return Event{}, err
	}
	return Event{Kind: EventRevalidated, Binding: e.Expected, Replacement: &next}, nil
}

// publisher is one destination ref's candidate steps (tracker or main).
type publisher interface {
	Push(gitx.Candidate) (gitx.PushOutcome, error)
	Probe(gitx.Candidate) (gitx.ProbeOutcome, string, error)
}

// diagnostics keeps the last transport error an observation absorbed, so an
// uncertain stop can say why (auth, hook, network) instead of only "unknown".
type diagnostics struct{ last error }

func (d *diagnostics) note(err error) {
	if err != nil {
		d.last = err
	}
}

// Diagnostic implements the optional interface Drive reports on an uncertain stop.
func (d *diagnostics) Diagnostic() error { return d.last }

// publishEvent pushes a prepared candidate once and reports the observation:
// accepted, a proven race, or unknown. confirm builds the Confirmed event.
func publishEvent(p publisher, d *diagnostics, e Effect, base string, confirm func() (Event, error)) (Event, error) {
	c := gitx.Candidate{Base: base, OID: e.CandidateOID}
	outcome, err := p.Push(c)
	d.note(err)
	switch outcome {
	case gitx.PushAccepted:
		return confirm()
	case gitx.PushRejected:
		// A rejection alone is not a race: the identical candidate may already be
		// there (a lost earlier acknowledgement). Probe before calling it lost.
		return probeEvent(p, d, e, base, confirm)
	}
	return Event{Kind: EventUnknown, Binding: e.Expected}, nil
}

// probeEvent observes an outstanding candidate. From Apply (a remote effect)
// absence is a race; from Probe it is a completed not-applied observation.
func probeEvent(p publisher, d *diagnostics, e Effect, base string, confirm func() (Event, error)) (Event, error) {
	c := gitx.Candidate{Base: base, OID: e.CandidateOID}
	probe, _, err := p.Probe(c)
	d.note(err)
	if probe == gitx.ProbeAbsentSame {
		// Settle a possibly delayed push by re-pushing the identical candidate
		// under the same lease: it lands once, or it already has. Then observe.
		out, err := p.Push(c)
		d.note(err)
		if out == gitx.PushAccepted {
			return confirm()
		}
		probe, _, err = p.Probe(c)
		d.note(err)
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

func publishCardEvent(repo *Repository, d *diagnostics, e Effect, cardPath string) (Event, error) {
	return publishEvent(repo, d, e, e.Expected.TrackerBase, func() (Event, error) { return confirmedCard(repo, e, cardPath) })
}

func probeCardEvent(repo *Repository, d *diagnostics, e Effect, cardPath string) (Event, error) {
	return probeEvent(repo, d, e, e.Expected.TrackerBase, func() (Event, error) { return confirmedCard(repo, e, cardPath) })
}

func confirmedCard(repo *Repository, e Effect, cardPath string) (Event, error) {
	blob, err := repo.CardBlobAt(e.CandidateOID, cardPath)
	if err != nil {
		return Event{}, err
	}
	return Event{Kind: EventConfirmed, Binding: e.Expected, CandidateOID: e.CandidateOID, ResultCardOID: blob}, nil
}
