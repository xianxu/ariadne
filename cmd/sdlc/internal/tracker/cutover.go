// cutover.go — the marker that makes a repository's tracker era explicit (#252).
//
// The migration's main commit adds CutoverMarkerPath naming the tracker's root
// commit. A checkout that sees a tracker but no marker, a marker but no tracker,
// or a marker whose root is not the tracker's, is on the wrong side of the
// cutover: it refuses rather than read cards against legacy details (or legacy
// details as if no tracker existed).
package tracker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// CutoverMarkerPath is the marker's repository-relative path, beside the
// details home.
var CutoverMarkerPath = path.Join(path.Dir(vocab.Issue().Discovery().Home), ManifestPath)

// ErrCutover marks a checkout on the wrong side of the tracker cutover.
var ErrCutover = errors.New("issue tracker cutover mismatch")

var markerOID = regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`)

type cutoverMarker struct {
	Version     int    `json:"version"`
	TrackerRoot string `json:"tracker_root"`
}

// CutoverMarkerBytes renders the marker naming the tracker's root commit.
func CutoverMarkerBytes(trackerRoot string) []byte {
	raw, _ := json.Marshal(cutoverMarker{Version: FormatVersion, TrackerRoot: trackerRoot})
	return append(raw, '\n')
}

// ParseCutoverMarker validates a marker (untrusted: hand-editable, possibly
// written by another version) and returns the tracker root it names.
func ParseCutoverMarker(raw []byte) (string, error) {
	var m cutoverMarker
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return "", fmt.Errorf("cutover marker %s: %w", CutoverMarkerPath, err)
	}
	if m.Version != FormatVersion {
		return "", fmt.Errorf("cutover marker %s: unsupported version %d (this sdlc understands %d)", CutoverMarkerPath, m.Version, FormatVersion)
	}
	if !markerOID.MatchString(m.TrackerRoot) {
		return "", fmt.Errorf("cutover marker %s: tracker_root must be a full object ID", CutoverMarkerPath)
	}
	return m.TrackerRoot, nil
}

// ReadCutoverMarker reads the marker in the checkout at root ("" when absent).
func ReadCutoverMarker(checkoutRoot string) (string, bool, error) {
	raw, err := os.ReadFile(filepath.Join(checkoutRoot, filepath.FromSlash(CutoverMarkerPath)))
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	trackerRoot, err := ParseCutoverMarker(raw)
	return trackerRoot, err == nil, err
}

// GuardCutover makes every read and card preparation of r prove that the
// checkout at checkoutRoot is on the tracker's side of the cutover. Commands
// open repositories guarded; package tests of the tracker itself need not.
func (r *Repository) GuardCutover(checkoutRoot string) *Repository {
	r.checkout = checkoutRoot
	return r
}

// read is the single snapshot read of a guarded repository.
func (r *Repository) read(view *gitx.TrunkView) (Snapshot, error) {
	if r.checkout != "" {
		if err := r.checkCutover(view.Ref()); err != nil {
			return Snapshot{}, err
		}
	}
	return readSnapshot(view)
}

func (r *Repository) checkCutover(ref string) error {
	trackerRoot, present, err := ReadCutoverMarker(r.checkout)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCutover, err)
	}
	if !present {
		return fmt.Errorf("%w: the issue tracker exists but this checkout has no %s.\n"+
			"      On a resting branch: `git pull` (the migration's main commit carries it).\n"+
			"      On a branch from before the cutover: `sdlc issue migrate --reconcile`.\n"+
			"      If main lacks it too, the migration is unfinished: `sdlc issue migrate` (dry run), then `--apply`",
			ErrCutover, CutoverMarkerPath)
	}
	ok, err := r.trunk.HasRoot(trackerRoot, ref)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: %s names tracker root %s, which is not where the issue tracker's history starts; the tracker was re-created — stop and reconcile by hand", ErrCutover, CutoverMarkerPath, trackerRoot)
	}
	return nil
}

// ErrLegacyDetails refuses a legacy-era action on unmirrored details in a
// checkout of a repository that has cut over to the tracker.
var ErrLegacyDetails = errors.New("details predate the issue tracker cutover")

// RefuseLegacyDetails is the one decision behind every "unmirrored details take
// the legacy path" site: in a repository that has cut over (the checkout
// carries the marker, or has fetched an issue tracker), the legacy path would
// write card fields into details, so it refuses with the reconcile action.
func RefuseLegacyDetails(checkoutRoot, detailsPath string) error {
	cut, err := CutOver(checkoutRoot)
	if err != nil || !cut {
		return err
	}
	return fmt.Errorf("%w: %s has no card_mirror, but this repository has an issue tracker; run `sdlc issue migrate --reconcile` on this branch first", ErrLegacyDetails, detailsPath)
}

// CutOver reports, offline, whether the checkout's repository has cut over to
// the issue tracker: the checkout carries the marker, or its clone has fetched
// a tracker. Legacy-only entrypoints refuse when it holds.
func CutOver(checkoutRoot string) (bool, error) {
	_, marked, err := ReadCutoverMarker(checkoutRoot)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrCutover, err)
	}
	if marked {
		return true, nil
	}
	return FetchedTracker(checkoutRoot)
}

// FetchedTracker reports whether the clone at root has fetched an issue
// tracker from any remote.
func FetchedTracker(root string) (bool, error) {
	out, err := gitx.RunGit("-C", root, "for-each-ref", "--count=1", "--format=%(refname)", "refs/remotes/*/"+vocab.Issue().Discovery().Tracker)
	if err != nil {
		return false, fmt.Errorf("list fetched issue trackers: %w", err)
	}
	return len(bytes.TrimSpace(out)) > 0, nil
}

// checkAbsentTracker refuses a guarded checkout whose marker names a tracker
// the remote does not have: falling back to legacy details would be wrong.
func (r *Repository) checkAbsentTracker() error {
	if r.checkout == "" {
		return nil
	}
	return MarkerWithoutTracker(r.checkout)
}

// MarkerWithoutTracker refuses a checkout that carries the cutover marker when
// no issue tracker is reachable (absent, or no publication target).
func MarkerWithoutTracker(checkoutRoot string) error {
	trackerRoot, present, err := ReadCutoverMarker(checkoutRoot)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCutover, err)
	}
	if present {
		return fmt.Errorf("%w: %s names issue tracker root %s, but no issue tracker is reachable; fix the publication remote (it must carry the issue-tracker branch) rather than fall back to legacy details", ErrCutover, CutoverMarkerPath, trackerRoot)
	}
	return nil
}
