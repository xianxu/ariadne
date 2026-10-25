package affected

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// Miss is one failing Go test whose package a selection did not run (#616): evidence of a missing declaration.
type Miss struct {
	Test string
	Pkg  string
}

// Missed returns the failures sel would not have run. Only a narrowed, group-free selection is evidence: a full one
// ran everything, and a group's command may run any test, so its reach is unknown (as splitImpact treats it). A name
// with no Go package, and a failure known explains (a registered flake), is never a miss.
func Missed(sel Selection, failures []string, known func(string) bool) []Miss {
	if sel.Full || len(sel.Groups) != 0 {
		return nil
	}
	selected := map[string]bool{}
	for _, p := range sel.GoPkgs {
		selected[p] = true
	}
	var out []Miss
	seen := map[Miss]bool{}
	for _, f := range failures {
		n, ok := flakes.ParseName(f)
		if !ok || selected[n.Pkg] || (known != nil && known(f)) {
			continue
		}
		m := Miss{Test: f, Pkg: n.Pkg}
		if !seen[m] {
			seen[m] = true
			out = append(out, m)
		}
	}
	return out
}

const (
	// MissedLedgerVersion is the one version a writer understands: an entry's members decide dedupe and pruning, so
	// a later version's file is refused rather than rewritten.
	MissedLedgerVersion = 1
	missedLedgerCap     = 200
	missedPathCap       = 20
)

// MissedLedger is the curation list of missed mappings: one entry per (package, changed-path set). Unknown members
// survive every rewrite, at both levels (ARCH-EVOLVE).
type MissedLedger struct {
	Version int                        `json:"version"`
	Entries []MissedEntry              `json:"entries"`
	Extra   map[string]json.RawMessage `json:"-"`
}

// MissedEntry is one missed mapping and its newest evidence.
type MissedEntry struct {
	ID         string                     `json:"id"`
	Package    string                     `json:"package"`
	Test       string                     `json:"test"`
	Paths      []string                   `json:"paths"`       // the first missedPathCap changed paths, sorted
	PathsTotal int                        `json:"paths_total"` // how many the change had
	Head       string                     `json:"head"`
	Main       string                     `json:"main"`
	Source     string                     `json:"source"` // "merge" or "train"
	FirstSeen  time.Time                  `json:"first_seen"`
	LastSeen   time.Time                  `json:"last_seen"`
	Hits       int                        `json:"hits"`
	Extra      map[string]json.RawMessage `json:"-"`
}

// MissEvidence is what one observation adds: the change's paths and where it was seen.
type MissEvidence struct {
	Paths  []string
	Head   string
	Main   string
	Source string
	At     time.Time
}

var missID = regexp.MustCompile(`^[0-9a-f]{12}$`)

// MissID keys an entry by its package and its full, sorted changed-path set (storage keeps only the first few).
func MissID(pkg string, paths []string) string {
	sorted := slices.Sorted(slices.Values(paths))
	sum := sha256.Sum256([]byte(pkg + "\n" + strings.Join(slices.Compact(sorted), "\n")))
	return hex.EncodeToString(sum[:])[:12]
}

// Record adds misses under ev: a known ID gains a hit and the newest evidence, keeping its first sighting and its
// unknown members; a new one is appended. At the cap the oldest entries (by last sighting) go; dropped says how many.
func (l *MissedLedger) Record(misses []Miss, ev MissEvidence) (recorded, dropped int) {
	if l.Version == 0 {
		l.Version = MissedLedgerVersion
	}
	paths := slices.Compact(slices.Sorted(slices.Values(ev.Paths)))
	kept := paths
	if len(kept) > missedPathCap {
		kept = kept[:missedPathCap]
	}
	for _, m := range misses {
		id := MissID(m.Pkg, paths)
		e := MissedEntry{ID: id, Package: m.Pkg, Test: m.Test, Paths: append([]string{}, kept...), PathsTotal: len(paths),
			Head: ev.Head, Main: ev.Main, Source: ev.Source, FirstSeen: ev.At, LastSeen: ev.At, Hits: 1}
		if i := slices.IndexFunc(l.Entries, func(x MissedEntry) bool { return x.ID == id }); i >= 0 {
			prev := l.Entries[i]
			e.Hits, e.Extra = prev.Hits+1, prev.Extra
			if prev.FirstSeen.Before(e.FirstSeen) {
				e.FirstSeen = prev.FirstSeen
			}
			l.Entries[i] = e
		} else {
			l.Entries = append(l.Entries, e)
		}
		recorded++
	}
	if over := len(l.Entries) - missedLedgerCap; over > 0 {
		slices.SortStableFunc(l.Entries, func(a, b MissedEntry) int { return a.LastSeen.Compare(b.LastSeen) })
		l.Entries = slices.Delete(l.Entries, 0, over)
		dropped = over
	}
	return recorded, dropped
}

// Dismiss removes the entry with id.
func (l *MissedLedger) Dismiss(id string) (MissedEntry, error) {
	i := slices.IndexFunc(l.Entries, func(x MissedEntry) bool { return x.ID == id })
	if i < 0 {
		return MissedEntry{}, fmt.Errorf("no missed mapping %q", id)
	}
	e := l.Entries[i]
	l.Entries = slices.Delete(l.Entries, i, i+1)
	return e, nil
}

// Prune removes the entries covered reports (main's mapping now selects them) and returns them.
func (l *MissedLedger) Prune(covered func(MissedEntry) bool) []MissedEntry {
	var gone []MissedEntry
	l.Entries = slices.DeleteFunc(l.Entries, func(e MissedEntry) bool {
		if covered(e) {
			gone = append(gone, e)
			return true
		}
		return false
	})
	return gone
}

func ledgerFields(l *MissedLedger, entries *[]json.RawMessage) map[string]any {
	return map[string]any{"version": &l.Version, "entries": entries}
}

func entryFields(e *MissedEntry) map[string]any {
	return map[string]any{"id": &e.ID, "package": &e.Package, "test": &e.Test, "paths": &e.Paths, "paths_total": &e.PathsTotal,
		"head": &e.Head, "main": &e.Main, "source": &e.Source, "first_seen": &e.FirstSeen, "last_seen": &e.LastSeen, "hits": &e.Hits}
}

// DecodeMissedLedger parses the ledger strictly where its schema is known and keeps every unknown member.
func DecodeMissedLedger(data []byte) (MissedLedger, error) {
	var l MissedLedger
	var raw []json.RawMessage
	extra, err := jsonstrict.DecodeObject(data, ledgerFields(&l, &raw))
	if err != nil {
		return MissedLedger{}, fmt.Errorf("missed ledger: %w", err)
	}
	if l.Version != MissedLedgerVersion {
		return MissedLedger{}, fmt.Errorf("missed ledger: version %d, want %d", l.Version, MissedLedgerVersion)
	}
	l.Extra = extra
	ids := map[string]bool{}
	for i, r := range raw {
		var e MissedEntry
		ex, err := jsonstrict.DecodeObject(r, entryFields(&e))
		if err != nil {
			return MissedLedger{}, fmt.Errorf("missed ledger: entries[%d]: %w", i, err)
		}
		if err := e.valid(); err != nil {
			return MissedLedger{}, fmt.Errorf("missed ledger: entries[%d]: %w", i, err)
		}
		if ids[e.ID] {
			return MissedLedger{}, fmt.Errorf("missed ledger: duplicate entry %s", e.ID)
		}
		ids[e.ID] = true
		e.Extra = ex
		l.Entries = append(l.Entries, e)
	}
	return l, nil
}

func (e MissedEntry) valid() error {
	switch {
	case !missID.MatchString(e.ID):
		return fmt.Errorf("bad id %q", e.ID)
	case e.Package == "" || e.Test == "" || e.Paths == nil:
		return errors.New("package, test and paths are required")
	case e.Source != "merge" && e.Source != "train":
		return fmt.Errorf("bad source %q", e.Source)
	case e.Hits < 1 || e.PathsTotal < len(e.Paths) || len(e.Paths) > missedPathCap:
		return errors.New("bad hits or paths")
	case e.FirstSeen.IsZero() || e.LastSeen.Before(e.FirstSeen):
		return errors.New("bad sighting times")
	}
	return nil
}

func (l MissedLedger) MarshalJSON() ([]byte, error) {
	entries := l.Entries
	if entries == nil {
		entries = []MissedEntry{}
	}
	base, err := json.Marshal(struct {
		Version int           `json:"version"`
		Entries []MissedEntry `json:"entries"`
	}{l.Version, entries})
	if err != nil {
		return nil, err
	}
	return jsonstrict.EncodeWithExtra(base, l.Extra, func(k string) bool { return k == "version" || k == "entries" }), nil
}

func (e MissedEntry) MarshalJSON() ([]byte, error) {
	type known MissedEntry
	base, err := json.Marshal(known(e))
	if err != nil {
		return nil, err
	}
	fields := entryFields(&e)
	return jsonstrict.EncodeWithExtra(base, e.Extra, func(k string) bool { _, ok := fields[k]; return ok }), nil
}
