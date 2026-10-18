package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/pkg/sidedoc"
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

	// The whole-file view (#424): the frontmatter's lines and the sections in
	// order (sorted by heading when unset), rendered by text(). The
	// frontmatter scope (#426, pair#643) reads and replaces only front; its
	// hash covers the block alone, so a section write never makes it stale.
	front string
	order []string
	// shadow is couch's diff baseline: what the agent last read or wrote,
	// per section ("" key: the frontmatter). Diff reports what differs from
	// it; a whole-file read resets all of it, a section write only its
	// section, a frontmatter read or write only "". nil until first used:
	// the starting state counts as seen.
	shadow          map[string]string
	editFrontBefore map[int]string // WriteFrontmatter call number → the frontmatter someone else writes just before it
	frontWriteCalls int
	frontWrites     int

	writeCalls int // every WriteSection call
	writes     int // successful writes
	seenAll    bool
}

// newFake is a session-sync file: couch serves only those, which it knows by
// their frontmatter's `type: session-sync`.
func newFake(sections map[string]string) *fakeSessionSync {
	return &fakeSessionSync{sections: sections, front: "type: session-sync", editBefore: map[int]map[string]string{}}
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
	if f.shadow != nil {
		f.shadow[heading] = f.sections[heading]
	}
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

func (f *fakeSessionSync) Diff(string) (string, error) {
	if f.diff != "" || f.shadow == nil {
		return f.diff, nil
	}
	var b strings.Builder
	if f.shadow[""] != f.front {
		b.WriteString("frontmatter changed\n")
	}
	for _, h := range f.headings() {
		if f.shadow[h] != f.sections[h] {
			b.WriteString("## " + h + " changed\n")
		}
	}
	return b.String(), nil
}

// seen starts the baseline at the current state (call before simulating an
// operator edit), and seeAll resets it to the current state.
func (f *fakeSessionSync) seen() { f.seeAll() }

func (f *fakeSessionSync) seeAll() {
	f.shadow = map[string]string{"": f.front}
	for h, b := range f.sections {
		f.shadow[h] = b
	}
}

func (f *fakeSessionSync) MarkAllSeen(string) error {
	f.seenAll = true
	f.seeAll()
	return nil
}

func (f *fakeSessionSync) headings() []string {
	if f.order != nil {
		return f.order
	}
	hs := make([]string, 0, len(f.sections))
	for h := range f.sections {
		hs = append(hs, h)
	}
	slices.Sort(hs)
	return hs
}

// text renders the whole file the way couch stores it.
func (f *fakeSessionSync) text() string {
	var b strings.Builder
	if f.front != "" {
		b.WriteString("---\n" + f.front + "\n---\n")
	}
	// Sections render byte for byte, so a frontmatter write leaves every
	// section's hash alone, as in couch.
	for _, h := range f.headings() {
		b.WriteString("\n## " + h + "\n" + f.sections[h])
	}
	return b.String()
}

// block is the frontmatter as couch's scope projects it.
func (f *fakeSessionSync) block() string {
	b, _, _ := sidedoc.FrontBlock(f.text())
	return b
}

func (f *fakeSessionSync) ReadFrontmatter(string) (string, string, error) {
	b := f.block()
	if b == "" {
		return "", "", errors.New("couch --session-sync read refused (session-sync): expected a closed leading frontmatter block")
	}
	if f.shadow != nil {
		f.shadow[""] = f.front
	}
	return b, fakeHash(b), nil
}

func (f *fakeSessionSync) WriteFrontmatter(_, bodyPath, hash string) error {
	f.frontWriteCalls++
	if fm, ok := f.editFrontBefore[f.frontWriteCalls]; ok {
		f.front = fm
	}
	if f.holder == "operator" {
		return ErrLocked
	}
	if hash == "" {
		return errors.New("couch --session-sync write: --if-hash is required")
	}
	if fakeHash(f.block()) != hash {
		return ErrStale
	}
	raw, err := os.ReadFile(bodyPath)
	if err != nil {
		return err
	}
	next, rest, ok := sidedoc.FrontBlock(string(raw))
	if !ok || strings.TrimSpace(rest) != "" {
		return errors.New("couch --session-sync write refused: the payload must be one frontmatter block and nothing else")
	}
	f.front = strings.TrimSuffix(strings.TrimPrefix(next, "---\n"), "\n---")
	f.frontWrites++
	if f.shadow != nil {
		f.shadow[""] = f.front
	}
	return nil
}

func (f *fakeSessionSync) Peek(string) (string, error) { return f.text(), nil }
