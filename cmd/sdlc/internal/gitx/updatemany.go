// updatemany.go — publish N paths to the trunk in ONE commit (ariadne#207).
//
// Update handles one path because its first consumer had one. Issue publishing
// carries every changed issue file and commits them together, so a per-path loop
// would turn one commit into N and make a partial publish reachable where it is
// not today.
package gitx

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ErrTrunkMoved reports that the CAS lost its attempt budget: the trunk moved
// under every attempt. Callers wrap it with what they saw contending.
var ErrTrunkMoved = errors.New("the trunk moved under the publish")

// TrunkWrite is the set of changes one commit applies.
//
// Delete is a SEPARATE FIELD rather than an absent key, because an absent key
// cannot mean "remove this": it is indistinguishable from "unchanged". The arm
// this replaces fails loudly on a deleted issue file; a write-only map would
// have reported success while the file stayed on the trunk. Making the state
// representable is the fix — the same shape as pathPresent's three outcomes.
type TrunkWrite struct {
	Write  map[string][]byte // create or replace
	Delete []string          // remove from the trunk
}

// TrunkView is read access to the base a given attempt will race — the ref as
// of the fetch that just ran, not as of when the caller was constructed.
type TrunkView struct {
	tf  *TrunkFile
	ref string
}

// ViewOf builds a view over an arbitrary ref.
//
// UpdateMany constructs the view itself for each attempt; this exists so a
// CONSUMER's tests can drive their own prepare callback without spinning a real
// CAS cycle. Without it a fake publisher can only pass nil, and every prepare
// that reads the base panics — which would push consumers toward not testing
// the read at all.
func (t *TrunkFile) ViewOf(ref string) *TrunkView { return &TrunkView{tf: t, ref: ref} }

// Read returns a path's bytes on this attempt's base; absent reads as empty.
func (v *TrunkView) Read(path string) ([]byte, error) { return v.tf.readFrom(v.ref, path) }

// Exists reports whether a path is present on this attempt's base.
func (v *TrunkView) Exists(path string) (bool, error) { return v.tf.pathPresent(v.ref, path) }

// Ref is the ref this attempt races, for callers that need to run their own
// query against it (refIDSpace, say) rather than one path at a time.
func (v *TrunkView) Ref() string { return v.ref }

// UpdateMany publishes a set of changes in one commit, DERIVING the set on every
// attempt.
//
// prepare runs after each fetch and before the tree is built, and returns the
// complete change set. Paths are derived per attempt rather than passed in as a
// map, and that is the whole point: a caller whose path depends on trunk state —
// an issue id — must re-decide it against the base the CAS will actually race. A
// fixed map makes the retry re-push the same colliding id and land the duplicate
// as a clean fast-forward, which is ariadne#188's hole one layer up.
func (t *TrunkFile) UpdateMany(msg string, prepare func(*TrunkView) (TrunkWrite, error)) error {
	sign, err := t.signs()
	if err != nil {
		return err
	}
	var lastRejection []byte
	for attempt := 1; attempt <= maxUpdateAttempts; attempt++ {
		if attempt == 1 {
			if out, err := t.fetch(); err != nil {
				return offlineError(t.remote, err, out)
			}
		}
		base, err := t.resolve(t.trackingRef())
		if err != nil {
			return err
		}

		// Pinned to the resolved base SHA, not the tracking-ref NAME: the ref can
		// move under a concurrent fetch, and prepare's decisions must describe the
		// exact tree this attempt's commit is built on.
		set, err := prepare(&TrunkView{tf: t, ref: base})
		if err != nil {
			return err // caller's error, surfaced unwrapped so errors.Is works
		}
		noop, err := t.setMatchesTrunk(set, base)
		if err != nil {
			return err
		}
		// Whole-set, not per-file: skip only when every Write already matches AND
		// every Delete is already absent. A per-file rule would commit whenever any
		// one file differed, breaking the caller's "main already carries this body,
		// publish only" idempotence.
		if noop {
			return nil
		}

		candidate, err := t.commitSet(set, msg, base, sign)
		if err != nil {
			return err
		}
		out, pushErr := t.pushExpected(candidate, base)
		if publicationStep(observedPush(pushErr, out), publicationUnconfirmed) == publicationSucceeded {
			return nil
		}
		// Unlike selected-commit publication, a reservation must establish that
		// THIS write won. An identical peer can construct the same commit, so
		// reachability is evidence only, never ownership after an unknown push.
		if observedPush(pushErr, out) == pushUnknown {
			return fmt.Errorf("%w: candidate %s; inspect %s/%s before retrying: %v\n%s", ErrPublicationUncertain, candidate, t.remote, t.branch, pushErr, out)
		}
		now, err := t.refreshTip()
		switch publicationStep(observedPush(pushErr, out), observedConfirmation(base, now, false, err)) {
		case publicationSucceeded:
			return nil
		case publicationUncertain:
			return fmt.Errorf("%w: candidate %s, remote %s: push %v; confirmation %v\n%s", ErrPublicationUncertain, candidate, now, pushErr, err, out)
		case publicationRefuse:
			return fmt.Errorf("publish set: %v\n%s", pushErr, out)
		case publicationRetry: // fetch pinned the new base; rerun prepare
		}

		lastRejection = out
	}
	return fmt.Errorf("%w after %d attempts; last rejection:\n%s",
		ErrTrunkMoved, maxUpdateAttempts, lastRejection)
}

// setMatchesTrunk reports whether applying this set would change nothing.
func (t *TrunkFile) setMatchesTrunk(set TrunkWrite, base string) (bool, error) {
	for path, want := range set.Write {
		present, err := t.pathPresent(base, path)
		if err != nil {
			return false, err
		}
		if !present {
			return false, nil
		}
		got, err := t.readFrom(base, path)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(got, want) {
			return false, nil
		}
	}
	for _, path := range set.Delete {
		present, err := t.pathPresent(base, path)
		if err != nil {
			return false, err
		}
		if present {
			return false, nil
		}
	}
	return true, nil
}

// commitSet builds one tree carrying every change without changing a caller ref.
func (t *TrunkFile) commitSet(set TrunkWrite, msg, base string, sign bool) (string, error) {
	idx, cleanupIdx, err := tempIndexPath()
	defer cleanupIdx()
	if err != nil {
		return "", err
	}
	env := []string{"GIT_INDEX_FILE=" + idx}

	if _, errOut, err := runGitIn(t.dir, env, "read-tree", base); err != nil {
		return "", fmt.Errorf("git tree construction: %v\n%s", err, errOut)
	}

	// Sorted so the git calls are deterministic — a map's iteration order would
	// make a failure reproduce differently each run.
	paths := make([]string, 0, len(set.Write))
	for p := range set.Write {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, path := range paths {
		blobFile, cleanupBlob, err := writeTemp(set.Write[path])
		if err != nil {
			cleanupBlob()
			return "", err
		}
		out, errOut, err := runGitIn(t.dir, env, "hash-object", "-w", "--path", path, blobFile)
		cleanupBlob()
		if err != nil {
			return "", fmt.Errorf("git tree construction: %v\n%s", err, errOut)
		}
		mode, err := t.modeOf(base, path)
		if err != nil {
			return "", err
		}
		if _, errOut, err := runGitIn(t.dir, env, "update-index", "--add",
			"--cacheinfo", mode+","+strings.TrimSpace(string(out))+","+path); err != nil {
			return "", fmt.Errorf("git tree construction: %v\n%s", err, errOut)
		}
	}
	for _, path := range set.Delete {
		// --force-remove drops the entry whether or not it is in the index, so a
		// delete of an already-absent path is not an error.
		if _, errOut, err := runGitIn(t.dir, env, "update-index", "--force-remove", path); err != nil {
			return "", fmt.Errorf("git tree construction: %v\n%s", err, errOut)
		}
	}

	out, errOut, err := runGitIn(t.dir, env, "write-tree")
	if err != nil {
		return "", fmt.Errorf("git tree construction: %v\n%s", err, errOut)
	}
	tree := strings.TrimSpace(string(out))

	return t.commitTree(tree, base, msg, sign)
}
