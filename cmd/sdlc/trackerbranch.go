package main

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// requireIssueBranch binds a branch-local effect to the card named by --issue.
// Ownership of the card alone does not prove that HEAD is on its branch.
func requireIssueBranch(id, cardPath, current string) error {
	want := issue.BranchName(cardPath)
	if current != want {
		return fmt.Errorf("#%s: run from its issue branch %s; this checkout is on %s", issue.CLIRef(id), want, valueOr(current, "a detached HEAD"))
	}
	return nil
}

// requireCloseIssueBranch also admits a renamed recovery branch at exactly the
// canonical issue tip. Close adds its own evidence commit after this check;
// ancestry alone would admit another issue's later work and is not a proof.
func requireCloseIssueBranch(env *trackerEnv, id, cardPath string) error {
	want := issue.BranchName(cardPath)
	if env.branch == want {
		return nil
	}
	if env.branch == "" || env.onRest() {
		return withLandedRecovery(env, id, want, requireIssueBranch(id, cardPath, env.branch))
	}
	tip, err := issueTip(env, want)
	if err != nil {
		return err
	}
	if tip != "" && env.head == tip {
		return nil
	}
	// #483: once the issue branch has landed, close runs from a follow-up
	// branch of this issue cut from main, carrying no other issue's work.
	if strings.HasPrefix(env.branch, id+"-") {
		if _, landed := issueBranchLanded(env, want); landed {
			return refuseUnlandedBase(env, "HEAD", gitx.RemoteTrackingRef(env.target.Remote, "main"), id, env.branch)
		}
	}
	return withLandedRecovery(env, id, want, fmt.Errorf("#%s: run from its issue branch %s; this checkout is on %s at %s, not the issue tip %s", issue.CLIRef(id), want, env.branch, shortOID(env.head), valueOr(shortOID(tip), "absent")))
}

// withLandedRecovery adds the landed issue's next action to a close refusal.
func withLandedRecovery(env *trackerEnv, id, want string, err error) error {
	if err == nil {
		return nil
	}
	if pr, landed := issueBranchLanded(env, want); landed {
		return fmt.Errorf("%w\n      %s", err, landedRecovery(id, want, env.target.Remote, pr))
	}
	return err
}

// issueTip is the issue branch's tip: the local ref, else the remote's, ""
// when both are gone.
func issueTip(env *trackerEnv, want string) (string, error) {
	tip, err := env.git("for-each-ref", "--format=%(objectname)", "refs/heads/"+want)
	if err != nil || tip != "" {
		return tip, err
	}
	return remoteRefTip(env.git, env.target.Remote, "refs/heads/"+want)
}

// closeBranchRecoveryNote describes the branch requireCloseIssueBranch
// admitted: a renamed branch at the issue tip, or else (#483) a follow-up
// branch of a landed issue, the only other branch it admits.
func closeBranchRecoveryNote(env *trackerEnv, cardPath string) string {
	want := issue.BranchName(cardPath)
	if env.branch == want {
		return ""
	}
	if tip, err := issueTip(env, want); err == nil && tip != env.head {
		return fmt.Sprintf("%s already landed; close runs from follow-up branch %s, whose PR lands the close evidence (#483)", want, env.branch)
	}
	return fmt.Sprintf("recovery branch %s has the same tip as %s; close leaves the renamed branch in place", env.branch, want)
}
