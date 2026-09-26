package tracker

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// Evidence performs a close's local effect: one commit on the source branch
// recording the listed files. Prepare builds the commit without touching the
// branch, index or worktree; Apply moves the branch to it (compare-and-swap on
// its parent) and refreshes only those index entries; Applied observes whether
// the branch contains it (false is a completed "not applied", never a guess).
type Evidence interface {
	Prepare(spec ReceiptSpec) (string, error)
	Apply(spec ReceiptSpec, commit string) error
	Applied(spec ReceiptSpec, commit string) (bool, error)
}

// CompletionOp adapts the completion receipt: commit the evidence, then publish
// codecomplete with its binding on the card. The card change is re-derived from
// the current card and the evidence commit message, so a later process — a
// FIX-THEN-SHIP resume or a recovery — needs nothing but the receipt.
type CompletionOp struct {
	diagnostics
	ctx      context.Context
	repo     *Repository
	branch   string // refs/heads/<name> of the checkout driving this run
	evidence Evidence
	today    string
}

func NewCompletionOp(ctx context.Context, repo *Repository, branch string, evidence Evidence, today string) *CompletionOp {
	return &CompletionOp{ctx: ctx, repo: repo, branch: branch, evidence: evidence, today: today}
}

// EvidenceCommit is the confirmed evidence commit, or "" before it exists.
func EvidenceCommit(r Receipt) string {
	for _, c := range r.Confirmations() {
		if c.Stage == "completion.evidence" {
			return c.CandidateOID
		}
	}
	return ""
}

var closeActualTrailer = regexp.MustCompile(`(?m)^Close-Actual: (\S+)$`)

// CloseActual reads the actual hours a close recorded in its evidence message.
func CloseActual(message string) (string, error) {
	m := closeActualTrailer.FindAllStringSubmatch(message, -1)
	if len(m) != 1 {
		return "", errors.New("evidence message must carry exactly one Close-Actual trailer")
	}
	return m[0][1], nil
}

// CodecompleteCard is the pure card change of a close: status codecomplete,
// measured (or N/A) actual hours, today's date and the completion binding. A
// card that is done, or not in a closable state, refuses rather than being
// overwritten.
func CodecompleteCard(current []byte, spec ReceiptSpec, evidenceCommit, today string) ([]byte, error) {
	fm, _, err := issue.Parse(string(current))
	if err != nil {
		return nil, err
	}
	if status, _ := issue.GetField(fm, "status"); status != "working" && status != "codecomplete" {
		return nil, fmt.Errorf("card #%s is %s; only working (or a re-close of codecomplete) work closes", spec.IssueID, status)
	}
	actual, err := CloseActual(spec.EvidenceMessage)
	if err != nil {
		return nil, err
	}
	// Actual first: every setter revalidates the card, and codecomplete requires it.
	next, err := issue.SetCardField(current, "actual_hours", actual)
	if err == nil {
		next, err = issue.SetCardField(next, "updated", today)
	}
	if err == nil {
		next, err = issue.SetCardField(next, "status", "codecomplete")
	}
	if err != nil {
		return nil, err
	}
	return issue.SetCardCompletion(next, issue.Completion{Token: spec.Token, Repository: spec.Repository,
		ReviewedHEAD: spec.ReviewedHEAD, EvidenceCommit: evidenceCommit})
}

func (op *CompletionOp) Prepare(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	switch e.Stage {
	case "completion.evidence":
		if err := requireSourceCheckout(spec, op.branch); err != nil {
			return Event{}, err
		}
		commit, err := op.evidence.Prepare(spec)
		if err != nil {
			return Event{}, err
		}
		return Event{Kind: EventCandidatePrepared, Binding: e.Expected, CandidateOID: commit}, nil
	case "completion.codecomplete":
		evidence := e.Expected.EvidenceOID
		c, err := op.repo.PrepareCardChange(spec.IssueID, spec.CardPath, "codecomplete", spec.Token, func(current []byte) ([]byte, error) {
			return CodecompleteCard(current, spec, evidence, op.today)
		})
		if err != nil {
			return Event{}, err
		}
		return Event{Kind: EventCandidatePrepared, Binding: e.Expected, CandidateOID: c.OID, CandidateBaseOID: c.Base}, nil
	}
	return Event{}, fmt.Errorf("completion cannot prepare stage %s", e.Stage)
}

func (op *CompletionOp) Apply(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	switch e.Kind {
	case WriteEvidence:
		if err := requireSourceCheckout(spec, op.branch); err != nil {
			return Event{}, err
		}
		if err := op.evidence.Apply(spec, e.CandidateOID); err != nil {
			op.note(err)
			// The ref CAS either happened or not; observe instead of guessing.
			return op.probeEvidence(e, spec, true)
		}
		return Event{Kind: EventConfirmed, Binding: e.Expected, CandidateOID: e.CandidateOID}, nil
	case PublishCard:
		return publishCardEvent(op.repo, &op.diagnostics, e, spec.CardPath)
	}
	return Event{}, fmt.Errorf("completion cannot apply %s", e.Kind)
}

func (op *CompletionOp) Probe(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	switch e.Stage {
	case "completion.evidence":
		if err := requireSourceCheckout(spec, op.branch); err != nil {
			return Event{}, err
		}
		return op.probeEvidence(e, spec, false)
	case "completion.codecomplete":
		return probeCardEvent(op.repo, &op.diagnostics, e, spec.CardPath)
	}
	return Event{}, fmt.Errorf("completion cannot probe stage %s", e.Stage)
}

func (op *CompletionOp) probeEvidence(e Effect, spec ReceiptSpec, applying bool) (Event, error) {
	applied, err := op.evidence.Applied(spec, e.CandidateOID)
	switch {
	case err != nil:
		op.note(err)
		if applying {
			return Event{Kind: EventUnknown, Binding: e.Expected}, nil
		}
		return Event{Kind: EventProbeUnknown, Binding: e.Expected}, nil
	case applied:
		return Event{Kind: EventConfirmed, Binding: e.Expected, CandidateOID: e.CandidateOID}, nil
	case applying:
		// A local effect is not a race: the apply failed and the branch does not
		// hold the commit. Stop with the receipt; recovery re-prepares.
		return Event{Kind: EventUnknown, Binding: e.Expected}, nil
	}
	return Event{Kind: EventNotApplied, Binding: e.Expected}, nil
}

// Refresh after a proven race or a not-applied evidence commit: the intent is
// unchanged; the next preparation reads the current branch and card.
func (op *CompletionOp) Refresh(e Effect, r Receipt) (Event, error) {
	spec := r.Spec()
	return Event{Kind: EventRevalidated, Binding: e.Expected, Replacement: &spec}, nil
}
