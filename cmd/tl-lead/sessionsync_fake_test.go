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

	// The whole-file view (#424): the frontmatter's lines and the sections in
	// order (sorted by heading when unset), rendered by text() and replaced by
	// WriteAll, so a section write and a whole-file write conflict as in couch.
	front         string
	order         []string
	// shadow is couch's diff baseline: what the agent last read or wrote,
	// per section ("" key: the frontmatter). Diff reports what differs from
	// it; a whole-file read or write resets all of it, a section write only
	// its section. nil until first used: the starting state counts as seen.
	shadow map[string]string
	editAllBefore map[int]map[string]string // WriteAll call number → heading → new body
	allWriteCalls int
	allWrites     int

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
	// Exactly invertible: WriteAll's parse returns each body byte for byte,
	// so a section the whole-file write leaves alone keeps its hash, as in couch.
	for _, h := range f.headings() {
		b.WriteString("\n## " + h + "\n" + f.sections[h])
	}
	return b.String()
}

func (f *fakeSessionSync) ReadAll(string) (string, string, error) {
	t := f.text()
	if f.shadow != nil {
		f.seeAll()
	}
	return t, fakeHash(t), nil
}

func (f *fakeSessionSync) WriteAll(_, bodyPath, hash string) error {
	f.allWriteCalls++
	for h, b := range f.editAllBefore[f.allWriteCalls] {
		f.sections[h] = b
	}
	if f.holder == "operator" {
		return ErrLocked
	}
	if hash == "" {
		return errors.New("couch --session-sync write: --if-hash is required")
	}
	if fakeHash(f.text()) != hash {
		return ErrStale
	}
	raw, err := os.ReadFile(bodyPath)
	if err != nil {
		return err
	}
	text := string(raw)
	f.front = ""
	if rest, ok := strings.CutPrefix(text, "---\n"); ok {
		fm, after, found := strings.Cut(rest, "\n---\n")
		if !found {
			return errors.New("fake: unterminated frontmatter")
		}
		f.front, text = fm, after
	}
	f.sections, f.order = map[string]string{}, nil
	parts := strings.Split("\n"+text, "\n## ")
	for _, part := range parts[1:] {
		h, body, _ := strings.Cut(part, "\n")
		f.sections[h] = body
		f.order = append(f.order, h)
	}
	f.allWrites++
	if f.shadow != nil {
		f.seeAll()
	}
	return nil
}

func (f *fakeSessionSync) Peek(string) (string, error) { return f.text(), nil }
