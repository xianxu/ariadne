package main

import (
	"fmt"

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
		return requireIssueBranch(id, cardPath, env.branch)
	}
	tip, err := env.git("for-each-ref", "--format=%(objectname)", "refs/heads/"+want)
	if err != nil {
		return err
	}
	if tip == "" {
		tip, err = remoteRefTip(env.git, env.target.Remote, "refs/heads/"+want)
		if err != nil {
			return err
		}
	}
	if tip != "" && env.head == tip {
		return nil
	}
	return fmt.Errorf("#%s: run from its issue branch %s; this checkout is on %s at %s, not the issue tip %s", issue.CLIRef(id), want, env.branch, shortOID(env.head), valueOr(shortOID(tip), "absent"))
}

func closeBranchRecoveryNote(env *trackerEnv, cardPath string) string {
	want := issue.BranchName(cardPath)
	if env.branch == want {
		return ""
	}
	return fmt.Sprintf("recovery branch %s has the same tip as %s; close leaves the renamed branch in place", env.branch, want)
}
