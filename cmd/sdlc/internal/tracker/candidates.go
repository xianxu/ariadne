package tracker

import (
	"bytes"
	"errors"
	"fmt"
	"path"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// ErrIDTaken reports that a creation candidate's ID exists on the fresh tip.
var ErrIDTaken = errors.New("tracker ID already allocated")

// The candidate-step API serves receipt-driven operations: each call is one
// declared effect (prepare, push, probe), so the pure receipt engine — not a
// fused retry loop — decides retries, refresh and recovery.

func cardMessage(id, what, token string) string {
	return fmt.Sprintf("#%s: tracker: %s\n\nTracker-Operation: %s", id, what, token)
}

// CardPath is the tracker path for a card, derived from its detail filename.
func CardPath(id, slug string) string {
	return path.Join(vocab.Issue().Discovery().Cards, id+"-"+slug+".md")
}

// PrepareCreate builds a candidate adding one new card. It refuses when the ID
// (under any slug) or the path is already present on the fresh tip.
func (r *Repository) PrepareCreate(cardPath string, raw []byte, token string) (gitx.Candidate, error) {
	card, err := r.validCandidateCard(cardPath, raw, token)
	if err != nil {
		return gitx.Candidate{}, err
	}
	content := bytes.Clone(raw)
	return r.trunk.PrepareCandidate(cardMessage(card.ID, "new card", token), func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		snapshot, err := readSnapshot(view)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		if prior, taken := snapshot.Card(card.ID); taken {
			return gitx.TrunkWrite{}, fmt.Errorf("%w: #%s is %s", ErrIDTaken, card.ID, prior.Path)
		}
		if err := snapshot.validateAddition(cardPath, content); err != nil {
			return gitx.TrunkWrite{}, err
		}
		return gitx.TrunkWrite{Write: map[string][]byte{cardPath: content}, ExactBytes: true}, nil
	})
}

// PrepareUpdate builds a candidate replacing one card at its expected blob.
func (r *Repository) PrepareUpdate(expected Record, raw []byte, what, token string) (gitx.Candidate, error) {
	card, err := r.validCandidateCard(expected.Path, raw, token)
	if err != nil {
		return gitx.Candidate{}, err
	}
	if expected.ID == "" || expected.BlobOID == "" || card.ID != expected.ID {
		return gitx.Candidate{}, errors.New("tracker update must preserve expected card identity")
	}
	content := bytes.Clone(raw)
	return r.trunk.PrepareCandidate(cardMessage(card.ID, what, token), func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		snapshot, err := readSnapshot(view)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		current, ok := snapshot.Card(expected.ID)
		if !ok || current.Path != expected.Path || current.BlobOID != expected.BlobOID {
			return gitx.TrunkWrite{}, ErrCardChanged
		}
		if bytes.Equal(current.Raw, content) {
			return gitx.TrunkWrite{}, ErrNoChange
		}
		if err := snapshot.validateReplacement(current, content); err != nil {
			return gitx.TrunkWrite{}, err
		}
		return gitx.TrunkWrite{Write: map[string][]byte{current.Path: content}, ExactBytes: true}, nil
	})
}

// PrepareCardChange builds a candidate applying mutate to the card as it is on
// the candidate's own pinned tip. For monotone changes an operation owns (its
// handoff record), this re-derives from current content instead of replacing a
// newer card with stale bytes; mutate refuses when the change no longer applies.
func (r *Repository) PrepareCardChange(id, cardPath, what, token string, mutate func(current []byte) ([]byte, error)) (gitx.Candidate, error) {
	if !tokenPattern.MatchString(token) || mutate == nil {
		return gitx.Candidate{}, errors.New("card change requires an operation token and mutation")
	}
	return r.trunk.PrepareCandidate(cardMessage(id, what, token), func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		snapshot, err := readSnapshot(view)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		current, ok := snapshot.Card(id)
		if !ok || current.Path != cardPath {
			return gitx.TrunkWrite{}, ErrCardChanged
		}
		next, err := mutate(current.Raw)
		if err != nil {
			return gitx.TrunkWrite{}, err
		}
		if _, err := r.validCandidateCard(cardPath, next, token); err != nil {
			return gitx.TrunkWrite{}, err
		}
		if bytes.Equal(current.Raw, next) {
			return gitx.TrunkWrite{}, ErrNoChange
		}
		if err := snapshot.validateReplacement(current, next); err != nil {
			return gitx.TrunkWrite{}, err
		}
		return gitx.TrunkWrite{Write: map[string][]byte{cardPath: next}, ExactBytes: true}, nil
	})
}

func (r *Repository) validCandidateCard(cardPath string, raw []byte, token string) (issue.Card, error) {
	if !tokenPattern.MatchString(token) {
		return issue.Card{}, errors.New("tracker candidate requires an operation token")
	}
	if len(raw) > gitx.SnapshotBlobLimit {
		return issue.Card{}, fmt.Errorf("%w: tracker candidate blob", gitx.ErrOutputLimit)
	}
	card, err := issue.ParseCard(raw)
	if err != nil {
		return issue.Card{}, err
	}
	id, slug, ok := issue.ParseFilename(path.Base(cardPath))
	if !ok || slug == "" || id != card.ID || path.Dir(cardPath) != vocab.Issue().Discovery().Cards {
		return issue.Card{}, fmt.Errorf("invalid tracker card path %q for #%s", cardPath, card.ID)
	}
	return card, nil
}

// validateAddition checks the envelope after adding one new card.
func (s Snapshot) validateAddition(cardPath string, raw []byte) error {
	if len(s.cards)+2 > gitx.SnapshotEntryLimit { // cards + manifest + new card
		return fmt.Errorf("%w: tracker entry count", gitx.ErrOutputLimit)
	}
	for _, r := range s.cards {
		if r.Path == cardPath {
			return fmt.Errorf("%w: %s", ErrIDTaken, cardPath)
		}
	}
	// The pinned commit OID has the repository's object-ID length.
	if blobWireBytes(s.ref, len(raw)) > gitx.SnapshotOutputLimit-s.wireBytes {
		return fmt.Errorf("%w: tracker batch after addition", gitx.ErrOutputLimit)
	}
	return nil
}

func (r *Repository) Push(c gitx.Candidate) (gitx.PushOutcome, error) {
	return r.trunk.PushCandidate(c)
}
func (r *Repository) Probe(c gitx.Candidate) (gitx.ProbeOutcome, string, error) {
	return r.trunk.ProbeCandidate(c)
}

// CardBlobAt names the card generation a confirmed candidate produced.
func (r *Repository) CardBlobAt(commit, cardPath string) (string, error) {
	oid, err := r.trunk.BlobAt(commit, cardPath)
	if err == nil && oid == "" {
		err = fmt.Errorf("candidate %s carries no card %s", commit, cardPath)
	}
	return oid, err
}

// ReadCardBlob reads an exact historical card generation (a mirror baseline).
// Missing and unreadable are errors: a baseline is never guessed.
func (r *Repository) ReadCardBlob(oid string) ([]byte, error) {
	view, err := r.trunk.Snapshot()
	if err != nil {
		return nil, err
	}
	return view.ReadBlob(oid)
}
