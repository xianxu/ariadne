// landingintegrate.go — integrate a moved main before testing and landing (#332).
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
)

const integrationTrailer = "Integrated-Main"
const integrationLimit = 3

// Mutable durations let tests exercise the real polling boundary without waiting
// minutes. CI's independent budget is owned by waitCIGate.
var integrationPollInterval = 5 * time.Second
var integrationSyncTimeout = 2 * time.Minute

type integrationStep string

const (
	integrationMerge      integrationStep = "merge"
	integrationAbort      integrationStep = "abort"
	integrationReset      integrationStep = "reset"
	integrationRetryReset integrationStep = "retry-reset"
	integrationPush       integrationStep = "push"
	integrationReplay     integrationStep = "replay"
	integrationGates      integrationStep = "gates"
	integrationLand       integrationStep = "land"
	integrationRefuse     integrationStep = "refuse"
)

func integrationAncestor(r gitRunner, root, ancestor, head string) (bool, error) {
	_, err := landingGit(r, root, "merge-base", "--is-ancestor", ancestor, head)
	if err == nil {
		return true, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) && ee.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// integrationResume validates the commit, not just its message. Only an exact
// clean two-parent merge on top of the observed remote branch may be published
// automatically. Ordinary unpublished edits keep the existing push refusal.
func integrationResume(r gitRunner, t landingTarget, head, remote, main string) (bool, error) {
	if head == "" || remote == "" || head == remote {
		return false, nil
	}
	parents, err := landingGit(r, t.Root, "show", "-s", "--format=%P", head)
	if err != nil {
		return false, err
	}
	p := strings.Fields(parents)
	if len(p) != 2 || p[0] != remote {
		return false, nil
	}
	trailer, err := landingGit(r, t.Root, "show", "-s", "--format=%(trailers:key="+integrationTrailer+",valueonly)", head)
	if err != nil {
		return false, err
	}
	integrated := strings.TrimSpace(trailer)
	if !landingOIDValid(integrated) || integrated != p[1] {
		return false, nil
	}
	included, err := integrationAncestor(r, t.Root, integrated, main)
	if err != nil || !included {
		return false, err
	}
	tree, err := landingGit(r, t.Root, "merge-tree", "--write-tree", p[0], p[1])
	if err != nil {
		return false, fmt.Errorf("cannot verify interrupted integration's clean merge: %w", err)
	}
	actual, err := landingGit(r, t.Root, "rev-parse", head+"^{tree}")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(tree) == actual, nil
}

// selectIntegrationPR preserves strict ordinary selection. The sole exception
// is a validated integration whose push (or GitHub registration) was interrupted.
func selectIntegrationPR(ctx context.Context, r gitRunner, gh landingGH, t landingTarget, branch, head, current string, prs []landingPR, lock func(func() error) error, stderr io.Writer) (landingPR, bool, error) {
	pr, selectionErr := selectLandingPR(prs, t, branch, head)
	if selectionErr == nil {
		return pr, false, nil
	}
	candidate, err := selectLandingCandidate(prs, t, branch, head)
	if err != nil || candidate.State != "OPEN" || current != branch {
		return landingPR{}, false, selectionErr
	}
	var main, remote string
	var valid bool
	err = lock(func() error {
		if err := revalidateLanding(r, t, branch, head, current); err != nil {
			return err
		}
		var err error
		main, err = t.fetchMain(r)
		if err != nil {
			return err
		}
		remote, err = remoteRefTip(func(args ...string) (string, error) { return landingGit(r, t.Root, args...) }, t.Remote, "refs/heads/"+branch)
		if err != nil {
			return err
		}
		if remote != candidate.HeadOID && remote != head {
			return selectionErr
		}
		valid, err = integrationResume(r, t, head, candidate.HeadOID, main)
		return err
	})
	if err != nil {
		return landingPR{}, false, err
	}
	if !valid {
		return landingPR{}, false, selectionErr
	}
	if remote == head {
		pr, err = syncIntegrationPR(ctx, gh, t, branch, head, candidate, stderr)
		return pr, false, err
	}
	return candidate, true, nil
}

func syncIntegrationPR(ctx context.Context, gh landingGH, t landingTarget, branch, head string, previous landingPR, stderr io.Writer) (landingPR, error) {
	ctx, cancel := context.WithTimeout(ctx, integrationSyncTimeout)
	defer cancel()
	for {
		prs, err := gh.LandingPRs(ctx, t.Repo, branch)
		if err != nil {
			return landingPR{}, err
		}
		pr, err := selectLandingCandidate(prs, t, branch, head)
		if err != nil {
			return landingPR{}, err
		}
		if pr.Number != previous.Number || pr.State != "OPEN" {
			return landingPR{}, errors.New("PR identity or state changed while registering integration; preserve checkout and retry")
		}
		if pr.HeadOID == head {
			return selectLandingPR(prs, t, branch, head)
		}
		if pr.HeadOID != previous.HeadOID {
			return landingPR{}, errors.New("PR head changed concurrently while registering integration; preserve checkout and retry")
		}
		fmt.Fprintf(stderr, "waiting for GitHub to register pushed head %s on PR #%d (limit %v)\n", shortOID(head), pr.Number, integrationSyncTimeout)
		timer := time.NewTimer(integrationPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return landingPR{}, fmt.Errorf("GitHub has not registered the pushed head; re-run sdlc merge: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

// integrateLandingMain makes the merge, or resumes the already validated one;
// replay must pass before a leased push can make either commit public.
func integrateLandingMain(ctx context.Context, r gitRunner, gh landingGH, t landingTarget, branch string, s integrationState, lock func(func() error) error, noJudge bool, issuesDir string, stderr io.Writer) (integrationState, error) {
	if err := lock(func() error {
		if err := revalidateLanding(r, t, branch, s.head, branch); err != nil {
			return err
		}
		if s.phase != integrationMerge {
			return nil
		}
		latest, err := t.fetchMain(r)
		if err != nil {
			return err
		}
		observation, err := observeIntegrationMain(r, t, s, latest, stderr)
		if err != nil {
			return err
		}
		s, _, err = transitionIntegration(s, observation)
		if err != nil || s.phase != integrationMerge {
			return err
		}
		msg := fmt.Sprintf("Merge %s into %s (sdlc merge, #332)\n\n%s: %s", s.main, branch, integrationTrailer, s.main)
		_, mergeErr := landingGit(r, t.Root, "-c", "merge.autoStash=false", "-c", "rerere.enabled=false", "merge", "--no-ff", "--no-edit", "-m", agentid.StampCurrent(msg, ctx), s.main)
		if mergeErr != nil {
			var effect integrationStep
			s, effect, _ = transitionIntegration(s, integrationEvent{kind: integrationMergeFailed})
			recovery := recoveryLandingGit(r)
			paths, pathErr := landingGit(recovery, t.Root, "diff", "--name-only", "--diff-filter=U", "-z")
			mergeHead, probeErr := landingOptional(recovery, t.Root, "rev-parse", "--verify", "--quiet", "MERGE_HEAD")
			if probeErr != nil {
				return probeErr
			}
			if mergeHead != "" && effect == integrationAbort {
				if _, err := landingGit(recovery, t.Root, "merge", "--abort"); err != nil {
					return fmt.Errorf("integration failed and abort failed: %w; original: %v", err, mergeErr)
				}
			}
			if pathErr != nil {
				return pathErr
			}
			s, _, _ = transitionIntegration(s, integrationEvent{kind: integrationRecoveryCompleted})
			return fmt.Errorf("main integration conflict or merge failure in %s (branch preserved): %w\n  → merge %s/main manually, resolve the paths, re-close with --fixed-then-ship --full-suite, then retry sdlc merge", strings.Join(splitNUL(paths), ", "), mergeErr, t.Remote)
		}
		head, err := landingGit(r, t.Root, "rev-parse", "HEAD")
		if err != nil {
			return err
		}
		s, _, err = transitionIntegration(s, integrationEvent{kind: integrationMerged, head: head})
		return err
	}); err != nil {
		return s, err
	}
	if s.phase == integrationGates {
		return s, nil
	}
	if s.phase != integrationReplay {
		return s, fmt.Errorf("integration expected replay, got %s", s.phase)
	}
	oldHead, head, pr := s.oldHead, s.head, s.pr
	main := s.main
	integrated := pr
	integrated.HeadOID, integrated.BaseOID = head, main
	var replayErr error
	if !noJudge {
		replayErr = runLandingPublishGate(ctx, integrated, issuesDir, stderr)
	}
	event := integrationReplayAccepted
	if replayErr != nil {
		event = integrationReplayRejected
	}
	var effect integrationStep
	var err error
	s, effect, err = transitionIntegration(s, integrationEvent{kind: event})
	if err != nil {
		return s, err
	}
	switch effect {
	case integrationReset:
		if err := lock(func() error {
			recovery := recoveryLandingGit(r)
			if err := revalidateLanding(recovery, t, branch, head, branch); err != nil {
				return fmt.Errorf("replay refused (%v); preserving changed checkout: %w", replayErr, err)
			}
			if _, err := t.fetchMain(recovery); err != nil {
				return err
			}
			if _, err := landingGit(recovery, t.Root, "reset", "--keep", oldHead); err != nil {
				return fmt.Errorf("replay refused (%v); rollback failed: %w", replayErr, err)
			}
			return nil
		}); err != nil {
			return s, err
		}
		s, _, _ = transitionIntegration(s, integrationEvent{kind: integrationRecoveryCompleted})
		return s, fmt.Errorf("integration replay refused; restored branch before pushing: %w", replayErr)
	case integrationPush:
		if err := lock(func() error {
			if err := revalidateLanding(r, t, branch, head, branch); err != nil {
				return err
			}
			latest, err := t.fetchMain(r)
			if err != nil {
				return err
			}
			observation, err := observeIntegrationMain(r, t, s, latest, stderr)
			if err != nil {
				return err
			}
			s, effect, err = transitionIntegration(s, observation)
			if err != nil {
				return err
			}
			if effect == integrationRetryReset {
				recovery := recoveryLandingGit(r)
				if _, err := landingGit(recovery, t.Root, "reset", "--keep", oldHead); err != nil {
					return fmt.Errorf("main moved before integration push; rollback failed (branch preserved): %w", err)
				}
				s, _, err = transitionIntegration(s, integrationEvent{kind: integrationRecoveryCompleted})
				return err
			}
			if _, err := landingGit(r, t.Root, "push", "--force-with-lease=refs/heads/"+branch+":"+oldHead, t.Remote, head+":refs/heads/"+branch); err != nil {
				s, _, _ = transitionIntegration(s, integrationEvent{kind: integrationPushFailed})
				return fmt.Errorf("integration lease push refused or completion uncertain; concurrent remote work is preserved (local integration retained); re-run sdlc merge to reconcile: %w", err)
			}
			return nil
		}); err != nil {
			return s, err
		}
	}
	if s.phase == integrationGates {
		return s, nil // Main moved during replay; retry from the original remote head.
	}
	synced, err := syncIntegrationPR(ctx, gh, t, branch, head, pr, stderr)
	if err != nil {
		return s, err
	}
	s, _, err = transitionIntegration(s, integrationEvent{kind: integrationPushRegistered, pr: synced})
	return s, err
}
