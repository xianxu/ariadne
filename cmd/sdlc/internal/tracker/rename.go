// rename.go — moving an issue's card and details to a new slug (#566). The two
// files live on different refs, so the rename is two compare-and-swap commits;
// what both refs hold decides the next one, and a rerun resumes or no-ops.
package tracker

import (
	"errors"
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// RenameObservation is which of the old and new names each ref holds now.
type RenameObservation struct{ TrackerOld, TrackerNew, MainOld, MainNew bool }

type RenameStep int

const (
	RenameCard RenameStep = iota + 1
	RenameDetails
	RenameDone
)

// ErrRenameInconsistent: a ref holds both names or neither; no step can repair it.
var ErrRenameInconsistent = errors.New("each ref must hold exactly one of the old and new names")

// NextRenameStep decides a rename's next ref update. The card moves first, so
// a half-done rename is the card at the new name and the details at the old.
func NextRenameStep(o RenameObservation) (RenameStep, error) {
	if o.TrackerOld == o.TrackerNew || o.MainOld == o.MainNew {
		return 0, ErrRenameInconsistent
	}
	switch {
	case o.TrackerOld:
		return RenameCard, nil
	case o.MainOld:
		return RenameDetails, nil
	}
	return RenameDone, nil
}

// DetailNames splits the details paths on main that carry issue id into the
// new name (present or not) and every other name it is still filed under.
func DetailNames(paths []string, id, newPath string) (olds []string, hasNew bool) {
	for _, p := range paths {
		if fid, _, ok := issue.ParseFilename(p); ok && fid == id {
			if p == newPath {
				hasNew = true
			} else {
				olds = append(olds, p)
			}
		}
	}
	return olds, hasNew
}

// MovePath is the prepare step moving one file's exact bytes to a new path on
// the candidate's pinned tip, refusing when the old is gone or the new taken.
func MovePath(from, to string) func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
	return func(view *gitx.TrunkView) (gitx.TrunkWrite, error) {
		if taken, err := view.Exists(to); err != nil {
			return gitx.TrunkWrite{}, err
		} else if taken {
			return gitx.TrunkWrite{}, fmt.Errorf("%s already exists", to)
		}
		// Read reports an absent path as empty; moving it would publish nothing.
		if present, err := view.Exists(from); err != nil {
			return gitx.TrunkWrite{}, err
		} else if !present {
			return gitx.TrunkWrite{}, fmt.Errorf("%s is no longer there", from)
		}
		b, err := view.Read(from)
		if err != nil {
			return gitx.TrunkWrite{}, fmt.Errorf("read %s: %w", from, err)
		}
		return gitx.TrunkWrite{Write: map[string][]byte{to: b}, Delete: []string{from}, ExactBytes: true}, nil
	}
}

// MoveCard moves card id from one path to another in one tracker commit.
func (r *Repository) MoveCard(id, from, to, token string) error {
	if !tokenPattern.MatchString(token) {
		return errors.New("tracker rename requires an operation token")
	}
	return r.trunk.UpdateMany(cardMessage(id, "rename card", token), MovePath(from, to))
}
