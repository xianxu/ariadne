package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

// fakeSessionSync models a session-sync file the way couch's help states it:
// per-section bodies with a content hash, the operator's lock, the owed-event
// ledger, compare-and-replace writes (stale on a hash mismatch, refused under
// the lock or without a hash), and a fold that refuses an ID not owed. It is
// stateful across calls; editBefore simulates someone else editing a section
// just before write call k, which makes that write go stale for real.
type fakeSessionSync struct {
	sections   map[string]string
	holder     string
	owed       []string
	diff       string
	editBefore map[int]map[string]string // write call number (1-based) → heading → new body

	gap bool // status refuses with a delivery gap

	writeCalls int // every WriteSection call
	writes     int // successful writes
	seenAll    bool
}

func newFake(sections map[string]string) *fakeSessionSync {
	return &fakeSessionSync{sections: sections, editBefore: map[int]map[string]string{}}
}

func fakeHash(body string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(body)))[:16]
}

func (f *fakeSessionSync) ReadSection(_, heading string) (string, string, error) {
	b, ok := f.sections[heading]
	if !ok {
		return "", "", fmt.Errorf("couch --session-sync read refused (session-sync): expected exactly one `## %s` heading line, found 0", heading)
	}
	return b, fakeHash(b), nil
}

func (f *fakeSessionSync) WriteSection(_, heading, bodyPath, hash string, fold []string) error {
	f.writeCalls++
	for h, b := range f.editBefore[f.writeCalls] {
		f.sections[h] = b
	}
	if f.holder == "operator" {
		return ErrLocked
	}
	if hash == "" {
		return errors.New("couch --session-sync write: --if-hash is required")
	}
	cur, ok := f.sections[heading]
	if !ok {
		return fmt.Errorf("couch --session-sync write refused (no-section): no ## %s", heading)
	}
	if fakeHash(cur) != hash {
		return ErrStale
	}
	for _, id := range fold {
		if !slices.Contains(f.owed, id) {
			return fmt.Errorf("couch --session-sync write refused (not-owed): %s", id)
		}
	}
	raw, err := os.ReadFile(bodyPath)
	if err != nil {
		return err
	}
	f.sections[heading] = strings.TrimSuffix(string(raw), "\n")
	f.owed = without(f.owed, fold)
	f.writes++
	return nil
}

func (f *fakeSessionSync) Status(string) (SyncStatus, error) {
	if f.gap {
		return SyncStatus{}, fmt.Errorf("couch --session-sync status: %w", ErrDeliveryGap)
	}
	return SyncStatus{Holder: f.holder, Owed: slices.Clone(f.owed)}, nil
}

func (f *fakeSessionSync) Diff(string) (string, error) { return f.diff, nil }

func (f *fakeSessionSync) MarkAllSeen(string) error {
	f.seenAll = true
	return nil
}
