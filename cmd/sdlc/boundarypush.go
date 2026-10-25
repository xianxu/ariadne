// boundarypush.go — the issue branch on the publication remote (#286). The
// owner pushes it at every sdlc boundary (start-plan, milestone-close, close,
// unclaim) so started work survives a lost machine, and merge and abandon
// delete it when the work ends. The claim plus no stacking (#272) make the
// owner its only writer, which is what makes a leased force-push safe.
package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// boundaryPushTimeout bounds one boundary push.
var boundaryPushTimeout = 2 * time.Minute

// boundaryPush pushes the checkout's issue branch after a boundary verb's own
// effect has landed. It never fails the verb (#286 D2): the next boundary or
// `sdlc pr` pushes again. A resting branch, main or a detached HEAD has
// nothing to push.
func boundaryPush(env *trackerEnv, stderr io.Writer, verb string) {
	boundaryPushAt(env, stderr, verb, nil, nil)
}

// closeBranchPublication is captured under the finalization lock and never
// refreshed by its unlocked publication tail.
type closeBranchPublication struct {
	root                 string
	target               gitx.PublicationTarget
	branch, head, expect string
}

func boundaryPushAt(env *trackerEnv, stderr io.Writer, verb string, publication *closeBranchPublication, lock func(func() error) error) {
	branch := env.branch
	if branch == "" || env.onRest() || branch == "main" {
		return
	}
	// Bounded, so a hanging network delays the verb by at most this much
	// instead of blocking it (D2: the push never holds the boundary hostage).
	ctx, cancel := context.WithTimeout(env.ctx, boundaryPushTimeout)
	defer cancel()
	bounded := *env
	bounded.ctx = ctx
	var err error
	if publication == nil {
		err = leasedBranchPush(bounded.git, bounded.target.Remote, branch)
	} else if publication.root != bounded.root {
		err = fmt.Errorf("publication repository changed")
	} else {
		err = pushClosePublication(bounded.git, *publication, lock)
	}
	if err != nil {
		if publication == nil {
			cwarn(stderr, fmt.Sprintf("%s: %s was not pushed: %v — the %s itself is done; the next boundary or `sdlc pr` pushes it", verb, branch, err, verb))
		} else {
			cwarn(stderr, fmt.Sprintf("%s: %s publication: %v — the %s itself is done; inspect before the next boundary or `sdlc pr`", verb, branch, err, verb))
		}
		return
	}
	cok(stderr, fmt.Sprintf("%s pushed to %s", branch, env.target.Remote))
}

// gitFn runs one git command in a checkout, returning trimmed stdout. Both
// the tracker environment and the landing runner supply one, so the issue
// branch has one leased push and one leased delete (ARCH-DRY).
type gitFn func(args ...string) (string, error)

// pushIssueBranch publishes the issue branch from a tracker checkout.
func pushIssueBranch(env *trackerEnv, branch string) error {
	return leasedBranchPush(env.git, env.target.Remote, branch)
}

// leasedBranchPush publishes branch to remote. The owner is its only writer
// (the claim, and no stacking, #272), so a rewritten branch may replace the
// remote's — but only the copy this checkout last saw: the lease is the
// last-fetched remote-tracking ref (absent: the branch must not exist there
// yet), never a fresh read that would turn it into a blind force (#284 BR-24).
// The explicit refspec publishes without configuring an issue-branch upstream.
func leasedBranchPush(git gitFn, remote, branch string) error {
	tracking := gitx.RemoteTrackingRef(remote, branch)
	expect, err := git("for-each-ref", "--format=%(objectname)", tracking)
	if err != nil {
		return err
	}
	ref := "refs/heads/" + branch

	if _, err := git("push", "-q", "--force-with-lease="+ref+":"+expect, remote, ref+":"+ref); err != nil {
		if strings.Contains(err.Error(), "stale info") {
			return fmt.Errorf("%w (the remote's %s is not the copy this checkout last fetched; fetch and inspect it before the next push)", err, branch)
		}
		return err
	}
	_, err = git("update-ref", tracking, ref)
	return err
}

// pushClosePublication uses only the authority captured at finalization.
func pushClosePublication(git gitFn, p closeBranchPublication, lock func(func() error) error) error {
	ref := "refs/heads/" + p.branch
	if _, err := git("push", "-q", "--force-with-lease="+ref+":"+p.expect, p.target.PushURL, p.head+":"+ref); err != nil {
		return fmt.Errorf("was not pushed: %w (the captured remote lease no longer permits publication; fetch and inspect before pushing)", err)
	}
	update := func() error {
		tracking := gitx.RemoteTrackingRef(p.target.Remote, p.branch)
		seen, err := git("for-each-ref", "--format=%(objectname)", tracking)
		if err != nil {
			return err
		}
		if seen == p.head {
			return nil
		}
		if seen != p.expect {
			return fmt.Errorf("tracking metadata changed; newer value preserved")
		}
		expect := p.expect
		if expect == "" {
			expect = strings.Repeat("0", len(p.head))
		}
		_, err = git("update-ref", tracking, p.head, expect)
		return err
	}
	var err error
	if lock != nil {
		err = lock(update)
	} else {
		err = update()
	}
	if err != nil {
		return fmt.Errorf("push completed; tracking metadata was not updated: %w", err)
	}
	return nil
}

// deleteRemoteBranch removes branch from the publication remote only while it
// is still at head (the lease), then drops its remote-tracking ref. A branch
// already gone is not an error.
func deleteRemoteBranch(git gitFn, remote, branch, head string) error {
	ref := "refs/heads/" + branch
	cur, err := remoteRefTip(git, remote, ref)
	if err != nil {
		return err
	}
	if cur != "" {
		if _, err := git("push", "-q", "--force-with-lease="+ref+":"+head, remote, ":"+ref); err != nil {
			if strings.Contains(err.Error(), "stale info") {
				return fmt.Errorf("%w (the remote's %s moved past %s; inspect it before deleting)", err, branch, shortOID(head))
			}
			return err
		}
	}
	tracking := gitx.RemoteTrackingRef(remote, branch)
	if seen, err := git("for-each-ref", "--format=%(objectname)", tracking); err != nil || seen == "" {
		return err
	}
	_, err = git("update-ref", "-d", tracking)
	return err
}

// remoteRefTip is the commit a full ref names on the remote, "" when absent:
// the one read of a remote ref (BR-14).
func remoteRefTip(git gitFn, remote, ref string) (string, error) {
	out, err := git("ls-remote", remote, ref)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) == 2 && f[1] == ref {
			return f[0], nil
		}
	}
	return "", nil
}
