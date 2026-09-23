package gitx

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

// CommitChange includes deleted paths and both endpoints of renames. Empty modes
// are represented by Git's 000000, so callers can validate the entire selection.
type CommitChange struct{ Path, OldMode, NewMode string }
type SelectedCommit struct {
	Source, Parent string
	message        string
	Changes        []CommitChange
}
type CommitPublicationOutcome string

const (
	CommitPublished      CommitPublicationOutcome = "published"
	CommitAlreadyApplied CommitPublicationOutcome = "already-applied"
	CommitNoChange       CommitPublicationOutcome = "no-change"
	CommitUncertain      CommitPublicationOutcome = "uncertain"
)

type CommitPublicationResult struct {
	Outcome                      CommitPublicationOutcome
	Source, Candidate, RemoteTip string
}

var ErrPublicationUncertain = errors.New("publication outcome uncertain")

// SelectCommit pins the source once. No source ancestors become part of the
// publication unit. Eligibility of workflow paths belongs to the command layer.
func (t *TrunkFile) SelectCommit(rev string) (SelectedCommit, error) {
	out, diag, err := runGitIn(t.dir, nil, "rev-parse", "--verify", "--end-of-options", rev+"^{commit}")
	if err != nil {
		return SelectedCommit{}, fmt.Errorf("resolve selected commit: %v\n%s", err, diag)
	}
	source, err := parseObjectID(out)
	if err != nil {
		return SelectedCommit{}, err
	}
	out, diag, err = runGitIn(t.dir, nil, "rev-list", "--parents", "-n", "1", source, "--")
	if err != nil {
		return SelectedCommit{}, fmt.Errorf("read selected parent: %v\n%s", err, diag)
	}
	fields := strings.Fields(string(out))
	if len(fields) != 2 {
		return SelectedCommit{}, errors.New("select one non-merge commit with exactly one parent; root and merge commits cannot be published")
	}
	parent, err := parseObjectID([]byte(fields[1]))
	if err != nil {
		return SelectedCommit{}, err
	}
	selected := SelectedCommit{Source: source, Parent: parent}
	out, diag, err = runGitIn(t.dir, nil, "diff-tree", "--raw", "--no-abbrev", "--no-renames", "-r", "-z", selected.Parent, source, "--")
	if err != nil {
		return SelectedCommit{}, fmt.Errorf("read selected paths: %v\n%s", err, diag)
	}
	parts := strings.Split(string(out), "\x00")
	for i := 0; i < len(parts)-1; i += 2 {
		if i+1 >= len(parts)-1 {
			return SelectedCommit{}, errors.New("malformed selected path list")
		}
		h := strings.Fields(parts[i])
		if len(h) != 5 || !strings.HasPrefix(h[0], ":") {
			return SelectedCommit{}, errors.New("malformed selected tree entry")
		}
		change := CommitChange{Path: parts[i+1], OldMode: strings.TrimPrefix(h[0], ":"), NewMode: h[1]}
		for _, mode := range []string{change.OldMode, change.NewMode} {
			if mode != "000000" && mode != "100644" {
				return SelectedCommit{}, fmt.Errorf("selected path %q has ineligible mode %s; only ordinary non-executable files may be published", change.Path, mode)
			}
		}
		selected.Changes = append(selected.Changes, change)
	}
	out, diag, err = runGitIn(t.dir, nil, "show", "-s", "--format=%B", source, "--")
	if err != nil {
		return SelectedCommit{}, fmt.Errorf("read selected message: %v\n%s", err, diag)
	}
	selected.message = strings.TrimRight(string(out), "\n")
	return selected, nil
}

// PublishCommit applies one already-selected patch with its explicit parent as
// the merge base. Git constructs objects only; caller refs/index/files stay put.
func (t *TrunkFile) PublishCommit(selected SelectedCommit) (CommitPublicationResult, error) {
	result := CommitPublicationResult{Source: selected.Source}
	if _, err := parseObjectID([]byte(selected.Source)); err != nil {
		return result, err
	}
	if _, err := parseObjectID([]byte(selected.Parent)); err != nil {
		return result, err
	}

	sign, err := t.signs()
	if err != nil {
		return result, err
	}
	for attempt := 1; attempt <= maxUpdateAttempts; attempt++ {
		if diag, err := t.fetch(); err != nil {
			return result, offlineError(t.remote, err, diag)
		}
		base, err := t.resolve(t.trackingRef())
		if err != nil {
			return result, err
		}
		result.RemoteTip = base
		applied, err := t.sourceApplied(selected.Source, base)
		if err != nil {
			return result, err
		}
		if applied {
			result.Outcome = CommitAlreadyApplied
			return result, nil
		}
		out, diag, err := runGitIn(t.dir, nil, "merge-tree", "--write-tree", "--name-only", "-z", "--merge-base="+selected.Parent, base, selected.Source)
		if err != nil {
			if gitExitCode(err) == 1 {
				return result, fmt.Errorf("selected commit conflicts; publish prerequisites first if needed: %s", strings.ReplaceAll(string(out), "\x00", " "))
			}
			return result, fmt.Errorf("three-way publication requires git merge-tree --write-tree --merge-base support: %v\n%s", err, diag)
		}
		tree := strings.SplitN(string(out), "\x00", 2)[0]
		tree, err = parseObjectID([]byte(tree))
		if err != nil {
			return result, err
		}
		existing, err := t.resolve(base + "^{tree}")
		if err != nil {
			return result, err
		}
		if tree == existing {
			result.Outcome = CommitNoChange
			return result, nil
		}
		msg := selected.message + "\n\nSource-Commit: " + selected.Source
		candidate, err := t.commitTree(tree, base, msg, sign)
		if err != nil {
			return result, err
		}
		result.Candidate = candidate
		pushOut, pushErr := t.pushExpected(candidate, base)
		if publicationStep(observedPush(pushErr, pushOut), publicationUnconfirmed) == publicationSucceeded {
			result.Outcome = CommitPublished
			return result, nil
		}
		now, confirmed, err := t.confirmPush(candidate)
		if err == nil && !confirmed {
			confirmed, err = t.sourceApplied(selected.Source, now)
		}
		switch publicationStep(observedPush(pushErr, pushOut), observedConfirmation(base, now, confirmed, err)) {
		case publicationSucceeded:
			result.RemoteTip = now
			result.Outcome = CommitPublished
			return result, nil
		case publicationUncertain:
			result.Outcome = CommitUncertain
			return result, fmt.Errorf("%w: source %s candidate %s; inspect %s/%s: push %v; confirmation %v\n%s", ErrPublicationUncertain, selected.Source, candidate, t.remote, t.branch, pushErr, err, pushOut)
		case publicationRefuse:
			return result, fmt.Errorf("publish selected commit: %v\n%s", pushErr, pushOut)
		case publicationRetry: // remerge the selected patch against the new tip
		}

	}
	return result, fmt.Errorf("%w after %d attempts", ErrTrunkMoved, maxUpdateAttempts)
}

// sourceApplied shares one deadline across reachability and trailer lookup. A
// failed query is never interpreted as absence. WaitDelay bounds pipe draining.
func (t *TrunkFile) sourceApplied(source, base string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ancestor, err := t.ancestor(ctx, source, base)
	if err != nil || ancestor {
		return ancestor, err
	}
	out, diag, err := runGitInContext(ctx, t.dir, nil, "log", "--fixed-strings", "--grep=Source-Commit: "+source, "--format=%(trailers:key=Source-Commit,valueonly)", base, "--")
	if err != nil {
		return false, fmt.Errorf("publication provenance query: %v\n%s", err, diag)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == source {
			return true, nil
		}
	}
	return false, nil
}
func (t *TrunkFile) ancestor(ctx context.Context, source, base string) (bool, error) {
	_, diag, err := runGitInContext(ctx, t.dir, nil, "merge-base", "--is-ancestor", source, base)
	if err == nil {
		return true, nil
	}
	if gitExitCode(err) == gitAbsentExit {
		return false, nil
	}
	return false, fmt.Errorf("publication reachability query: %v\n%s", err, diag)
}
func (t *TrunkFile) commitTree(tree, base, msg string, sign bool) (string, error) {
	args := []string{"commit-tree", tree, "-p", base, "-m", msg}
	if sign {
		args = append([]string{"commit-tree", "-S"}, args[1:]...)
	}
	out, diag, err := runGitIn(t.dir, nil, args...)
	if err != nil {
		return "", fmt.Errorf("create publication commit: %v\n%s", err, diag)
	}
	return parseObjectID(out)
}
func (t *TrunkFile) pushExpected(candidate, base string) ([]byte, error) {
	out, diag, err := runGitIn(t.dir, nil, "push", "--porcelain", "--force-with-lease=refs/heads/"+t.branch+":"+base, t.remote, candidate+":refs/heads/"+t.branch)
	return append(out, diag...), err
}

// pushRejectedExplicitly recognizes Git porcelain refusal records for our sole
// pushed ref. Receive-pack can reject a simultaneous CAS after negotiation, so
// that refusal appears as remote rejected rather than client-side stale info.
func pushRejectedExplicitly(out []byte) bool {
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "!\t") && (strings.Contains(line, "[rejected] (stale info)") || strings.Contains(line, "[remote rejected]")) {
			return true
		}
	}
	return false
}
func (t *TrunkFile) refreshTip() (string, error) {
	if diag, err := t.fetch(); err != nil {
		return "", offlineError(t.remote, err, diag)
	}
	return t.resolve(t.trackingRef())
}
func (t *TrunkFile) confirmPush(candidate string) (string, bool, error) {
	now, err := t.refreshTip()
	if err != nil {
		return "", false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	confirmed, err := t.ancestor(ctx, candidate, now)
	return now, confirmed, err
}

// parseObjectID closes the Git object grammar before responses become argv.
func parseObjectID(out []byte) (string, error) {
	oid := strings.TrimSpace(string(out))
	if len(oid) != 40 && len(oid) != 64 {
		return "", fmt.Errorf("malformed Git object ID %q", oid)
	}
	if _, err := hex.DecodeString(oid); err != nil || strings.Trim(oid, "0") == "" {
		return "", fmt.Errorf("malformed Git object ID %q", oid)
	}
	return oid, nil
}
