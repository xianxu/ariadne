// planningbranch.go — start-plan's early design branch (#252). Design commits
// belong on the issue's own branch from the first keystroke, so a resting
// branch never accumulates checkpoint commits that later diverge from main.
package main

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

// planningBranchResult says what preparation did, for the caller's report.
type planningBranchResult int

const (
	planningAlreadyOnBranch planningBranchResult = iota
	planningCreatedBranch
	planningSwitchedBranch
)

// preparePlanningBranch puts the checkout on the issue branch (named by the
// details file stem). From rest it branches at freshly pinned main after proving
// the details are there, never moving the resting ref or dropping local work.
// Any other branch refuses: it is someone's unrelated work.
func preparePlanningBranch(env *trackerEnv, id, detailPath string) (planningBranchResult, error) {
	name := strings.TrimSuffix(path.Base(detailPath), ".md")
	if env.branch == name {
		return planningAlreadyOnBranch, nil
	}
	if !env.onRest() {
		current := env.branch
		if current == "" {
			current = "a detached HEAD"
		}
		return 0, fmt.Errorf("this checkout is on %s, not its resting branch %s or #%s's branch %s.\n"+
			"      Plan #%s from another checkout (a free slot), or finish/park the work here first", current, env.resting, id, name, id)
	}
	if dirty, err := env.git("status", "--porcelain", "--untracked-files=no"); err != nil {
		return 0, err
	} else if dirty != "" {
		return 0, fmt.Errorf("%s has uncommitted tracked changes; commit or stash them before planning #%s (nothing was changed):\n%s", env.resting, id, dirty)
	}
	view, err := env.main.Snapshot()
	if err != nil {
		return 0, err
	}
	pinned := view.Ref()
	if contained, err := env.gitTest("merge-base", "--is-ancestor", "HEAD", pinned); err != nil {
		return 0, err
	} else if !contained {
		return 0, fmt.Errorf("%s has commits that are not on main (ahead or diverged); reconcile them — push, move them to a branch, or drop them deliberately — before planning #%s. Nothing was changed", env.resting, id)
	}
	present, err := view.Exists(detailPath)
	if err != nil {
		return 0, err
	}
	if !present {
		return 0, fmt.Errorf("#%s's details (%s) are not on main; its creation is incomplete, so it cannot be planned yet", id, detailPath)
	}
	exists, err := env.gitTest("rev-parse", "--verify", "-q", "refs/heads/"+name)
	if err != nil {
		return 0, err
	}
	if exists {
		if _, err := env.git("switch", "-q", name); err != nil {
			return 0, fmt.Errorf("switch to existing %s: %w", name, err)
		}
		env.branch = name
		return planningSwitchedBranch, nil
	}
	if _, err := env.git("switch", "-q", "--no-track", "-c", name, pinned); err != nil {
		return 0, fmt.Errorf("create %s at main: %w", name, err)
	}
	head, err := env.git("rev-parse", "HEAD")
	if err != nil || head != pinned {
		return 0, errors.New("planning branch is not at the pinned main commit")
	}
	env.branch, env.head = name, head
	return planningCreatedBranch, nil
}
