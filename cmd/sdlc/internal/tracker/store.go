package tracker

import (
	"errors"
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// RecoveryReceipts keeps receipts in checkout-local recovery refs, one per
// operation token, pinning the objects a resumed operation must re-read (the
// card and source blobs and any outstanding candidate). Every write is a CAS on
// the generation this process last wrote or loaded.
type RecoveryReceipts struct {
	store      *gitx.RecoveryStore
	repository string
	gen        map[string]string
}

func NewRecoveryReceipts(store *gitx.RecoveryStore, repository string) *RecoveryReceipts {
	return &RecoveryReceipts{store: store, repository: repository, gen: map[string]string{}}
}

func (s *RecoveryReceipts) Save(r Receipt) error {
	raw, err := MarshalReceipt(r)
	if err != nil {
		return err
	}
	b := r.binding()
	var objects []string
	for _, oid := range []string{b.CandidateOID, r.wire.Spec.SourceBlob, r.wire.Spec.CardOID} {
		if oid != "" {
			objects = append(objects, oid)
		}
	}
	token := r.wire.Spec.Token
	oid, err := s.store.Save(token, s.gen[token], gitx.RecoveryData{Document: raw, Objects: objects})
	if err != nil {
		return err
	}
	s.gen[token] = oid
	return nil
}

func (s *RecoveryReceipts) Delete(r Receipt) error {
	token := r.wire.Spec.Token
	if s.gen[token] == "" {
		return fmt.Errorf("receipt %s was never persisted by this run", token)
	}
	if err := s.store.Delete(token, s.gen[token]); err != nil {
		return err
	}
	delete(s.gen, token)
	return nil
}

// Discard drops a receipt that published nothing (see Discardable). A receipt
// this run never persisted needs no cleanup.
func (s *RecoveryReceipts) Discard(r Receipt) error {
	if !r.Discardable() {
		return fmt.Errorf("receipt %s guards published or outstanding work", r.wire.Spec.Token)
	}
	if s.gen[r.wire.Spec.Token] == "" {
		return nil
	}
	return s.Delete(r)
}

// Load reads one receipt and adopts its generation for later CAS writes.
func (s *RecoveryReceipts) Load(token string) (Receipt, error) {
	entry, err := s.store.Load(token)
	if err != nil {
		return Receipt{}, err
	}
	r, err := ParseReceipt(entry.Data.Document, s.repository)
	if err != nil {
		return Receipt{}, fmt.Errorf("recovery record %s: %w", token, err)
	}
	if r.wire.Spec.Token != token {
		return Receipt{}, fmt.Errorf("recovery record %s names another operation", token)
	}
	s.gen[token] = entry.OID
	return r, nil
}

// List loads every receipt for this repository. A record that cannot be parsed
// is reported, never skipped: it may guard an uncertain publication.
func (s *RecoveryReceipts) List() ([]Receipt, error) {
	entries, err := s.store.List()
	if err != nil {
		return nil, err
	}
	var receipts []Receipt
	var bad []error
	for _, e := range entries {
		r, err := s.Load(e.Token)
		if err != nil {
			bad = append(bad, err)
			continue
		}
		receipts = append(receipts, r)
	}
	return receipts, errors.Join(bad...)
}

// Discardable reports that nothing was published or left outstanding: no stage
// is proven and no candidate awaits an observation. Deleting it loses nothing.
func (r Receipt) Discardable() bool {
	switch r.wire.Phase {
	case phaseQueued, phasePreparing, phaseRefreshing:
		return len(r.wire.Proofs) == 0
	}
	return false
}

// Stage names the operation's current stage ("finalize" once all are proven).
func (r Receipt) Stage() string { return r.binding().Stage }
