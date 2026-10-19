package main

import (
	"crypto/sha256"
	"encoding/json"
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
	pre   string // the text between the frontmatter and the first section (#455)
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

	// The registry (#455): the paths couch tracks, any that the operator
	// holds locked, every MarkAllSeen by path, and successful whole-file
	// applies to the session file.
	registered  map[string]bool
	lockedFiles map[string]bool
	file        string // the session file's path; Apply to any other edits it on disk, as couch does
	seenFiles   []string
	applies     int
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

func (f *fakeSessionSync) Status(file string) (SyncStatus, error) {
	if f.gap {
		return SyncStatus{}, fmt.Errorf("couch --session-sync status: %w", ErrDeliveryGap)
	}
	if f.lockedFiles[file] {
		return SyncStatus{Holder: "operator"}, nil
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

func (f *fakeSessionSync) MarkAllSeen(file string) error {
	f.seenFiles = append(f.seenFiles, file)
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
	b.WriteString(f.pre)
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

// Apply is couch's whole-file edit batch: refused under the lock, stale when
// the file's hash moved, and all or none, every record resolved against the
// file as it stands before anything is written (the occurrence-th plain
// match of a non-empty old, no two records overlapping).
func (f *fakeSessionSync) Apply(file, editsPath, hash string) error {
	if f.holder == "operator" || f.lockedFiles[file] {
		return ErrLocked
	}
	onDisk := f.file != "" && file != f.file
	text := f.text()
	if onDisk {
		b, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		text = string(b)
	}
	if hash != fakeHash(text) {
		return ErrStale
	}
	raw, err := os.ReadFile(editsPath)
	if err != nil {
		return err
	}
	var recs []editRecord
	if err := json.Unmarshal(raw, &recs); err != nil {
		return fmt.Errorf("couch --session-sync apply refused (edits-invalid): %v", err)
	}
	next, err := applyRecords(text, recs)
	if err != nil {
		return fmt.Errorf("couch --session-sync apply refused (edits-invalid): %v", err)
	}
	if onDisk {
		f.applies++
		return os.WriteFile(file, []byte(next), 0o644)
	}
	if err := f.load(next); err != nil {
		return err
	}
	f.applies++
	return nil
}

// load replaces the fake's file with text, in text()'s shape.
func (f *fakeSessionSync) load(text string) error {
	front, rest, ok := sidedoc.FrontBlock(text)
	if !ok {
		return errors.New("fake: no frontmatter")
	}
	f.front = strings.TrimSuffix(strings.TrimPrefix(front, "---\n"), "\n---")
	rest = strings.TrimPrefix(rest, "\n")
	parts := strings.Split(rest, "\n## ")
	f.pre, f.sections, f.order = parts[0], map[string]string{}, nil
	for _, part := range parts[1:] {
		h, body, _ := strings.Cut(part, "\n")
		f.sections[h] = body
		f.order = append(f.order, h)
	}
	return nil
}

func (f *fakeSessionSync) Registrations() ([]string, error) {
	var out []string
	for p := range f.registered {
		out = append(out, p)
	}
	slices.Sort(out)
	return out, nil
}

// Rename is couch's: the new path must exist, and a rename already done
// (old gone from the registry, new in it) succeeds again.
func (f *fakeSessionSync) Rename(old, new string) error {
	if !f.registered[old] {
		if f.registered[new] {
			return nil
		}
		return fmt.Errorf("couch --session-sync rename refused (not-registered): %s", old)
	}
	if _, err := os.Stat(new); err != nil {
		return fmt.Errorf("couch --session-sync rename: %v", err)
	}
	delete(f.registered, old)
	f.registered[new] = true
	return nil
}
