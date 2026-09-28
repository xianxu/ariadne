package tracker

import (
	"bytes"
	"encoding/json"
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

// Snapshot owns a pinned inventory. Accessors never expose its mutable storage.
type Snapshot struct {
	ref       string
	cards     map[string]Record
	maxID     int
	wireBytes int
}

func (s Snapshot) Ref() string { return s.ref }
func (s Snapshot) MaxID() int  { return s.maxID }
func (s Snapshot) Card(id string) (Record, bool) {
	r, ok := s.cards[id]
	r.Raw = bytes.Clone(r.Raw)
	return r, ok
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
	s := Snapshot{ref: ref, cards: map[string]Record{}}
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
		card, err := issue.ParseCard(f.Content)
		if err != nil {
			return Snapshot{}, fmt.Errorf("tracker %s: %w", f.Path, err)
		}
		if card.ID != id {
			return Snapshot{}, fmt.Errorf("tracker %s: filename ID disagrees with card ID %s", f.Path, card.ID)
		}
		if prior, exists := s.cards[id]; exists {
			return Snapshot{}, fmt.Errorf("duplicate tracker ID %s: %s and %s", id, prior.Path, f.Path)
		}
		s.cards[id] = Record{ID: id, Path: f.Path, BlobOID: f.OID, Card: card, Raw: bytes.Clone(f.Content)}
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
	if len(raw) > gitx.SnapshotBlobLimit {
		return fmt.Errorf("%w: tracker blob %s", gitx.ErrOutputLimit, current.Path)
	}
	remaining := s.wireBytes - blobWireBytes(current.BlobOID, len(current.Raw))
	if blobWireBytes(current.BlobOID, len(raw)) > gitx.SnapshotOutputLimit-remaining {
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
