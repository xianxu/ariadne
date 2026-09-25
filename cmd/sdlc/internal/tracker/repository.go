// Package tracker provides authoritative issue-card snapshots and conditional
// writes on the dedicated tracker branch, without changing the caller checkout.
package tracker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

var ErrCardChanged = errors.New("tracker card changed; reread before retrying")
var ErrNoChange = errors.New("tracker write makes no change; content equality does not establish operation ownership")

// Repository always fetches from the caller's explicit publication remote.
// Construct it with the command context; never reuse it after cancellation.
type Repository struct{ trunk *gitx.TrunkFile }

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
	return readSnapshot(view)
}

var tokenPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

// UpdateCard writes one existing card only if the expected path and blob still
// identify its current version. The caller supplies a unique durable operation
// token and records each candidate before publication. An uncertain publication
// remains uncertain; this API never infers success from equal content.
func (r *Repository) UpdateCard(expected Record, raw []byte, operationToken string, beforePush func(base, candidate string) error) error {
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
	message := fmt.Sprintf("#%s: tracker: update card\n\nTracker-Operation: %s", card.ID, operationToken)
	return r.trunk.UpdateManyPrepared(message, func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		snapshot, err := readSnapshot(view)
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
