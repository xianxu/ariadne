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
