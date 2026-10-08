// boundarypush.go — the issue branch on the publication remote (#286). The
// owner pushes it at every sdlc boundary (start-plan, milestone-close, close,
// unclaim) so started work survives a lost machine, and merge and abandon
// delete it when the work ends. The claim plus no stacking (#272) make the
// owner its only writer, which is what makes a leased force-push safe.
package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// boundaryPush pushes the checkout's issue branch after a boundary verb's own
// effect has landed. It never fails the verb (#286 D2): the next boundary or
// `sdlc pr` pushes again. A resting branch, main or a detached HEAD has
// nothing to push.
func boundaryPush(env *trackerEnv, stderr io.Writer, verb string) {
	branch := env.branch
	if branch == "" || env.onRest() || branch == "main" {
		return
	}
	if err := pushIssueBranch(env, branch); err != nil {
		cwarn(stderr, fmt.Sprintf("%s: %s was not pushed: %v — the %s itself is done; the next boundary or `sdlc pr` pushes it", verb, branch, err, verb))
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
// It also records the upstream, so a plain `git push` works afterwards.
func leasedBranchPush(git gitFn, remote, branch string) error {
	tracking := gitx.RemoteTrackingRef(remote, branch)
	expect, err := git("for-each-ref", "--format=%(objectname)", tracking)
	if err != nil {
		return err
	}
	ref := "refs/heads/" + branch
	if _, err := git("push", "-q", "-u", "--force-with-lease="+ref+":"+expect, remote, ref+":"+ref); err != nil {
		if strings.Contains(err.Error(), "stale info") {
			return fmt.Errorf("%w (the remote's %s is not the copy this checkout last fetched; fetch and inspect it before the next push)", err, branch)
		}
		return err
	}
	_, err = git("update-ref", tracking, ref)
	return err
}

// deleteRemoteBranch removes branch from the publication remote only while it
// is still at head (the lease), then drops its remote-tracking ref. A branch
// already gone is not an error.
func deleteRemoteBranch(git gitFn, remote, branch, head string) error {
	ref := "refs/heads/" + branch
	out, err := git("ls-remote", "--heads", remote, ref)
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) != "" {
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
