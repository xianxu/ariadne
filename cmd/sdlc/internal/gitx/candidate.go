package gitx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrCandidateNoChange reports that a prepared set would not change the tip.
var ErrCandidateNoChange = errors.New("candidate makes no change on the destination tip")

// Candidate is one prepared commit and the exact destination tip it extends.
// Its message must carry a unique operation token: an operation proves
// ownership by the reachability of this exact commit, never by equal content.
type Candidate struct{ Base, OID string }

type PushOutcome uint8

const (
	PushUnknown PushOutcome = iota
	PushAccepted
	PushRejected
)

type ProbeOutcome uint8

const (
	// ProbeUnknown: no completed observation (offline, cancelled, query failure).
	ProbeUnknown ProbeOutcome = iota
	// ProbeReachable: the candidate is an ancestor of the fresh destination tip.
	ProbeReachable
	// ProbeAbsentMoved: the tip moved past the candidate's base without it, so
	// the candidate's lease can never succeed later. It is proven not applied.
	ProbeAbsentMoved
	// ProbeAbsentSame: the tip still equals the base. A delayed receive-pack
	// could yet apply the candidate; only re-pushing the identical candidate
	// under the same lease settles the outcome.
	ProbeAbsentSame
)

// PrepareCandidate fetches the destination, pins its tip, and builds exactly one
// commit on it. It never pushes and never changes a caller ref, index or worktree.
func (t *TrunkFile) PrepareCandidate(msg string, prepare func(*TrunkView) (TrunkWrite, error)) (Candidate, error) {
	if strings.TrimSpace(msg) == "" || prepare == nil {
		return Candidate{}, errors.New("candidate requires a message and prepare callback")
	}
	sign, err := t.signs()
	if err != nil {
		return Candidate{}, err
	}
	if out, err := t.fetch(); err != nil {
		return Candidate{}, offlineError(t.remote, err, out)
	}
	base, err := t.resolve(t.trackingRef())
	if err != nil {
		return Candidate{}, err
	}
	set, err := prepare(&TrunkView{tf: t, ref: base})
	if err != nil {
		return Candidate{}, err
	}
	noop, err := t.setMatchesTrunk(set, base)
	if err != nil {
		return Candidate{}, err
	}
	if noop {
		return Candidate{}, ErrCandidateNoChange
	}
	oid, err := t.commitSet(set, msg, base, sign)
	if err != nil {
		return Candidate{}, err
	}
	return Candidate{Base: base, OID: oid}, nil
}

// PushCandidate attempts the candidate once, leased on its base. A transport or
// cancellation failure is Unknown, never a rejection.
func (t *TrunkFile) PushCandidate(c Candidate) (PushOutcome, error) {
	if _, err := parseObjectID([]byte(c.Base)); err != nil {
		return PushUnknown, err
	}
	if _, err := parseObjectID([]byte(c.OID)); err != nil {
		return PushUnknown, err
	}
	if err := t.operationContext().Err(); err != nil {
		return PushUnknown, err
	}
	out, err := t.pushExpected(c.OID, c.Base)
	if t.operationContext().Err() != nil {
		return PushUnknown, t.operationContext().Err()
	}
	switch observedPush(err, out) {
	case pushAccepted:
		return PushAccepted, nil
	case pushRejected:
		return PushRejected, nil
	}
	return PushUnknown, fmt.Errorf("%w: candidate %s: %v\n%s", ErrPublicationUncertain, c.OID, err, out)
}

// ProbeCandidate observes the fresh destination tip. It performs no mutation.
func (t *TrunkFile) ProbeCandidate(c Candidate) (ProbeOutcome, string, error) {
	now, err := t.refreshTip()
	if err != nil {
		return ProbeUnknown, "", err
	}
	ctx, cancel := context.WithTimeout(t.operationContext(), 30*time.Second)
	defer cancel()
	reachable, err := t.ancestor(ctx, c.OID, now)
	if err != nil {
		return ProbeUnknown, now, err
	}
	switch {
	case reachable:
		return ProbeReachable, now, nil
	case now == c.Base:
		return ProbeAbsentSame, now, nil
	}
	return ProbeAbsentMoved, now, nil
}

// ReadAt returns a path's bytes at an exact commit (absent reads as empty), for
// callers that must describe a confirmed candidate rather than the moving tip.
func (t *TrunkFile) ReadAt(commit, path string) ([]byte, error) {
	if _, err := parseObjectID([]byte(commit)); err != nil {
		return nil, err
	}
	return t.readFrom(commit, path)
}

// BlobAt returns a path's blob OID at an exact commit, or "" when absent.
func (t *TrunkFile) BlobAt(commit, path string) (string, error) {
	if _, err := parseObjectID([]byte(commit)); err != nil {
		return "", err
	}
	out, diag, err := t.run(nil, "ls-tree", "--object-only", commit, "--", path)
	if err != nil {
		return "", fmt.Errorf("read blob identity %s:%s: %v\n%s", commit, path, err, diag)
	}
	if strings.TrimSpace(string(out)) == "" {
		return "", nil
	}
	return parseObjectID(out)
}

// RemoteExists reports whether the destination branch exists on the remote.
// Absent is a completed observation (ls-remote exit 2), never a transport error.
func (t *TrunkFile) RemoteExists() (bool, error) {
	out, diag, err := t.run(nil, "ls-remote", "--refs", "--exit-code", "--", t.remote, t.localRef())
	if err == nil {
		return len(strings.TrimSpace(string(out))) > 0, nil
	}
	if gitExitCode(err) == 2 && len(out) == 0 {
		return false, nil
	}
	return false, offlineError(t.remote, err, diag)
}

// LocalView pins the last-fetched tracking ref without network IO. ok is false
// when this clone has never fetched the branch.
func (t *TrunkFile) LocalView() (*TrunkView, bool, error) {
	present, err := t.refPresent(t.trackingRef())
	if err != nil || !present {
		return nil, false, err
	}
	tip, err := t.resolve(t.trackingRef())
	if err != nil {
		return nil, false, err
	}
	return t.ViewOf(tip), true, nil
}
