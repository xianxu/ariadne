package tracker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

const ManifestPath = "issue-tracker.json"
const FormatVersion = 1

// ManifestBytes returns the version marker for an initialized tracker.
func ManifestBytes() []byte { return []byte(fmt.Sprintf("{\"version\":%d}\n", FormatVersion)) }

// Record is a detached value projection of an authoritative tracker card.
type Record struct {
	ID, Path, BlobOID string
	Card              issue.Card
	Raw               []byte
}

// UnreadableCard is a card the snapshot holds but cannot parse (#288): one
// malformed card is quarantined here instead of failing every tracker read.
// It keeps its ID and path (no reuse) and is never returned as a Record.
type UnreadableCard struct {
	ID, Path, BlobOID string
	Err               error
}

var (
	// ErrNoCard: the tracker has no card for the ID.
	ErrNoCard = errors.New("no card on the tracker")
	// ErrUnreadableCard: the tracker's card for the ID cannot be parsed.
	ErrUnreadableCard = errors.New("unreadable card on the tracker")
)

// Snapshot owns a pinned inventory. Accessors never expose its mutable storage.
type Snapshot struct {
	ref        string
	cards      map[string]Record
	unreadable map[string]UnreadableCard
	maxID      int
	wireBytes  int
}

func (s Snapshot) Ref() string { return s.ref }
func (s Snapshot) MaxID() int  { return s.maxID }
func (s Snapshot) Card(id string) (Record, bool) {
	r, ok := s.cards[id]
	r.Raw = bytes.Clone(r.Raw)
	return r, ok
}

// Unreadable lists the quarantined cards, by ID.
func (s Snapshot) Unreadable() []UnreadableCard {
	out := make([]UnreadableCard, 0, len(s.unreadable))
	for _, u := range s.unreadable {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Require is the lookup for a verb acting on one card: the record, or an error
// that says whether the card is missing (ErrNoCard) or unreadable
// (ErrUnreadableCard, naming its path and cause).
func (s Snapshot) Require(id string) (Record, error) {
	if r, ok := s.Card(id); ok {
		return r, nil
	}
	if u, ok := s.unreadable[id]; ok {
		return Record{}, fmt.Errorf("card #%s (%s): %w: %v", id, u.Path, ErrUnreadableCard, u.Err)
	}
	return Record{}, fmt.Errorf("card #%s: %w", id, ErrNoCard)
}

// taken reports the path holding id, readable or not.
func (s Snapshot) taken(id string) (string, bool) {
	if r, ok := s.cards[id]; ok {
		return r.Path, true
	}
	u, ok := s.unreadable[id]
	return u.Path, ok
}

func (s Snapshot) Records() []Record {
	ids := make([]string, 0, len(s.cards))
	for id := range s.cards {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	records := make([]Record, 0, len(ids))
	for _, id := range ids {
		r, _ := s.Card(id)
		records = append(records, r)
	}
	return records
}

func readSnapshot(view *gitx.TrunkView) (Snapshot, error) {
	files, err := view.Files("")
	if err != nil {
		return Snapshot{}, err
	}
	return parseSnapshot(view.Ref(), files)
}

func parseSnapshot(ref string, files []gitx.TreeFile) (Snapshot, error) {
	s := Snapshot{ref: ref, cards: map[string]Record{}, unreadable: map[string]UnreadableCard{}}
	if len(files) > gitx.SnapshotEntryLimit {
		return Snapshot{}, fmt.Errorf("%w: tracker entry count", gitx.ErrOutputLimit)
	}
	// Validate the storage envelope before parsing/copying any card. cat-file
	// emits a frame for every entry, even entries sharing the same blob OID.
	for _, f := range files {
		if len(f.Content) > gitx.SnapshotBlobLimit {
			return Snapshot{}, fmt.Errorf("%w: tracker blob %s", gitx.ErrOutputLimit, f.Path)
		}
		frame := blobWireBytes(f.OID, len(f.Content))
		if frame > gitx.SnapshotOutputLimit-s.wireBytes {
			return Snapshot{}, fmt.Errorf("%w: tracker batch", gitx.ErrOutputLimit)
		}
		s.wireBytes += frame
	}
	manifest := false
	home := vocab.Issue().Discovery().Cards
	for _, f := range files {
		if f.Mode != "100644" {
			return Snapshot{}, fmt.Errorf("tracker %s: ordinary 100644 file required (mode %s)", f.Path, f.Mode)
		}
		if f.Path == ManifestPath {
			if manifest {
				return Snapshot{}, fmt.Errorf("duplicate tracker manifest")
			}
			manifest = true
			if err := validateManifest(f.Content); err != nil {
				return Snapshot{}, err
			}
			continue
		}
		if path.Dir(f.Path) != home {
			return Snapshot{}, fmt.Errorf("unexpected tracker path %q", f.Path)
		}
		id, slug, ok := issue.ParseFilename(path.Base(f.Path))
		if !ok || slug == "" || strings.ContainsAny(slug, "\\\x00\r\n") {
			return Snapshot{}, fmt.Errorf("invalid tracker card path %q", f.Path)
		}
		if prior, exists := s.taken(id); exists {
			return Snapshot{}, fmt.Errorf("duplicate tracker ID %s: %s and %s", id, prior, f.Path)
		}
		// A card's own content is quarantined, not fatal (#288); everything
		// above is the tracker's structure and still fails the read.
		card, err := issue.ParseCard(f.Content)
		if err == nil && card.ID != id {
			err = fmt.Errorf("filename ID disagrees with card ID %s", card.ID)
		}
		if err != nil {
			s.unreadable[id] = UnreadableCard{ID: id, Path: f.Path, BlobOID: f.OID, Err: fmt.Errorf("tracker %s: %w", f.Path, err)}
		} else {
			s.cards[id] = Record{ID: id, Path: f.Path, BlobOID: f.OID, Card: card, Raw: bytes.Clone(f.Content)}
		}
		n, _ := strconv.Atoi(id)
		if n > s.maxID {
			s.maxID = n
		}
	}
	if !manifest {
		return Snapshot{}, fmt.Errorf("tracker manifest %s missing; complete migration before use", ManifestPath)
	}
	return s, nil
}

func blobWireBytes(oid string, size int) int {
	return len(oid) + len(" blob ") + len(strconv.Itoa(size)) + 1 + size + 1
}

// validateReplacement checks the exact response size after replacing one
// existing card. The object format and entry count do not change on update.
func (s Snapshot) validateReplacement(current Record, raw []byte) error {
	return s.validateReplacements([]replacement{{current, raw}})
}

// replacement is one card's new bytes over the record it replaces.
type replacement struct {
	current Record
	raw     []byte
}

// validateReplacements checks every replacement's blob limit and that the
// snapshot, with all of them applied, stays within the read budget (#284:
// a set changes several cards in one commit).
func (s Snapshot) validateReplacements(rs []replacement) error {
	wire := s.wireBytes
	for _, r := range rs {
		if len(r.raw) > gitx.SnapshotBlobLimit {
			return fmt.Errorf("%w: tracker blob %s", gitx.ErrOutputLimit, r.current.Path)
		}
		wire += blobWireBytes(r.current.BlobOID, len(r.raw)) - blobWireBytes(r.current.BlobOID, len(r.current.Raw))
	}
	if wire > gitx.SnapshotOutputLimit {
		return fmt.Errorf("%w: tracker replacement exceeds batch budget", gitx.ErrOutputLimit)
	}
	return nil
}

func validateManifest(raw []byte) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	start, err := d.Token()
	if err != nil || start != json.Delim('{') {
		return fmt.Errorf("invalid tracker manifest: expected object")
	}
	seen := false
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return fmt.Errorf("invalid tracker manifest: %w", err)
		}
		if key != "version" || seen {
			return fmt.Errorf("invalid tracker manifest field %v", key)
		}
		seen = true
		var version int
		if err := d.Decode(&version); err != nil || version != FormatVersion {
			return fmt.Errorf("unsupported tracker format; require version %d", FormatVersion)
		}
	}
	end, err := d.Token()
	if err != nil || end != json.Delim('}') || !seen {
		return fmt.Errorf("invalid tracker manifest: missing version or closing object")
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("invalid tracker manifest: trailing content")
	}
	return nil
}
