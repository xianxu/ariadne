// Package tracker provides authoritative issue-card snapshots and conditional
// writes on the dedicated tracker branch, without changing the caller checkout.
package tracker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

var ErrCardChanged = errors.New("tracker card changed; reread before retrying")
var ErrNoChange = errors.New("tracker write makes no change; content equality does not establish operation ownership")

// Repository always fetches from the caller's explicit publication remote.
// Construct it with the command context; never reuse it after cancellation.
type Repository struct {
	trunk    *gitx.TrunkFile
	checkout string // guarded checkout root (GuardCutover); "" for unguarded
	// markerAt, when set (GuardCutoverAt), is the commit whose tree carries the
	// marker the guard reads, instead of the checkout's files.
	markerAt string
	// verified is the last (marker root, tracker ref) pair the guard proved, so
	// repeated reads of one generation (a CAS retry, a second reader) do not
	// respawn the root listing.
	verified [2]string
}

func NewRepository(ctx context.Context, root, remote string) (*Repository, error) {
	if root == "" || remote == "" {
		return nil, errors.New("tracker requires explicit repository root and publication remote")
	}
	tf, err := gitx.NewTrunkFileContext(ctx, root, remote, vocab.Issue().Discovery().Tracker)
	if err != nil {
		return nil, err
	}
	return &Repository{trunk: tf}, nil
}

func (r *Repository) Snapshot() (Snapshot, error) {
	view, err := r.trunk.Snapshot()
	if err != nil {
		return Snapshot{}, err
	}
	return r.read(view)
}

var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// UpdateCard writes one existing card only if the expected path and blob still
// identify its current version. The caller supplies a unique durable operation
// token and records each candidate before publication. An uncertain publication
// remains uncertain; this API never infers success from equal content.
func (r *Repository) UpdateCard(expected Record, raw []byte, operationToken string, beforePush func(base, candidate string) error) error {
	return r.UpdateCardWithTrailers(expected, raw, operationToken, nil, beforePush)
}

// UpdateCardWithTrailers is UpdateCard whose tracker commit also carries the
// given trailer lines ("Key: value") after its Tracker-Operation — the record
// of why the card changed, where the change itself is the record (#278).
func (r *Repository) UpdateCardWithTrailers(expected Record, raw []byte, operationToken string, trailers []string, beforePush func(base, candidate string) error) error {
	for _, t := range trailers {
		if strings.ContainsAny(t, "\r\n") || !strings.Contains(t, ": ") {
			return fmt.Errorf("malformed tracker trailer %q", t)
		}
	}
	if !tokenPattern.MatchString(operationToken) || beforePush == nil {
		return errors.New("tracker update requires an operation token and receipt callback")
	}
	if len(raw) > gitx.SnapshotBlobLimit {
		return fmt.Errorf("%w: tracker replacement blob", gitx.ErrOutputLimit)
	}
	card, err := issue.ParseCard(raw)
	if err != nil {
		return err
	}
	if expected.ID == "" || expected.BlobOID == "" || card.ID != expected.ID {
		return errors.New("tracker update must preserve expected card identity")
	}
	content := bytes.Clone(raw)
	message := cardMessage(card.ID, "update card", operationToken, trailers...)
	return r.trunk.UpdateManyPrepared(message, func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		snapshot, err := r.read(view)
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
	}, beforePush)
}

// LocalSnapshot reads the last-fetched tracker without network IO; ok is false
// when this clone has never fetched it. Its content may be stale: callers label
// it so and never authorize a write from it.
func (r *Repository) LocalSnapshot() (Snapshot, bool, error) {
	view, ok, err := r.trunk.LocalView()
	if err != nil || !ok {
		return Snapshot{}, ok, err
	}
	s, err := r.read(view)
	return s, err == nil, err
}

// TrackingRef is the local remote-tracking ref of the tracker branch, for
// readers that include its history (the active-time window).
func (r *Repository) TrackingRef() string { return r.trunk.TrackingRef() }
