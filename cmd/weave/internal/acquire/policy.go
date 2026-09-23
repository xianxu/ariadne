package acquire

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// Policy restricts source acquisition to independent siblings of a verified host.
// Callers obtain these paths from workspace environment discovery.
type Policy struct{ EnvironmentRoot, HostRoot, HostCommonDir string }

// validate receives lexical and physical evidence separately: canonicalization
// must never conceal an unsupported manifest-relative destination.
func (p Policy) validate(lexical, physical string, source Source) error {
	if filepath.Dir(lexical) != p.EnvironmentRoot || filepath.Dir(physical) != p.EnvironmentRoot || lexical != physical {
		return fmt.Errorf("source destination %s must be an ordinary direct child of environment %s; compose from the dependency primary", lexical, p.EnvironmentRoot)
	}
	if physical == p.HostRoot {
		return fmt.Errorf("source destination %s collides with environment host", physical)
	}
	if strings.HasPrefix(source.Identity, "file:") {
		return fmt.Errorf("numbered environment requires a recorded remote source, not a local/file source")
	}
	return nil
}

func (c Client) privateExisting(ctx context.Context, dir string) error {
	top, err := c.git(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	common, err := c.git(ctx, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return err
	}
	gitDir, err := c.git(ctx, dir, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return err
	}
	if canonical(top) != dir || canonical(common) != filepath.Join(dir, ".git") || canonical(gitDir) != canonical(common) || canonical(common) == canonical(c.Policy.HostCommonDir) {
		return fmt.Errorf("dependency %s must be an independent ordinary checkout, without host aliases or linked worktrees", dir)
	}
	return nil
}

// The transition core accepts only evidence in acquisition order. No state is
// persisted: the filesystem and the existing owned-stage ledger remain truth.
type acquisitionState uint8

const (
	uninspected acquisitionState = iota
	existingComplete
	absent
	stagingClone
	checking
	stagedVerified
	publishing
	published
	failed
	unconfirmed
)

type acquisitionEvent uint8

const (
	existingVerified acquisitionEvent = iota
	absentRemote
	stageCreated
	cloneSucceeded
	checksSucceeded
	destinationAbsent
	publishSucceeded
	operationFailed
	publicationUncertain
	destinationAppeared
)

func transition(state acquisitionState, event acquisitionEvent) (acquisitionState, error) {
	if state == existingComplete || state == published || state == failed || state == unconfirmed {
		return state, fmt.Errorf("acquisition already complete or requires reinspection")
	}
	if event == operationFailed || event == destinationAppeared {
		return failed, nil
	}
	if state == publishing && event == publicationUncertain {
		return unconfirmed, nil
	}
	switch {
	case state == uninspected && event == existingVerified:
		return existingComplete, nil
	case state == uninspected && event == absentRemote:
		return absent, nil
	case state == absent && event == stageCreated:
		return stagingClone, nil
	case state == stagingClone && event == cloneSucceeded:
		return checking, nil
	case state == checking && event == checksSucceeded:
		return stagedVerified, nil
	case state == stagedVerified && event == destinationAbsent:
		return publishing, nil
	case state == publishing && event == publishSucceeded:
		return published, nil
	}
	return state, fmt.Errorf("invalid acquisition event %d in state %d", event, state)
}
