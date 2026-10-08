// planningbranch.go — start-plan's early design branch (#252). Design commits
// belong on the issue's own branch from the first keystroke, so a resting
// branch never accumulates checkpoint commits that later diverge from main.
package main

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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
	name := issue.BranchName(detailPath)
	if env.branch == name {
		view, err := env.main.Snapshot()
		if err != nil {
			return 0, err
		}
		if err := refuseUnlandedBase(env, "HEAD", view.Ref(), id, name); err != nil {
			return 0, err
		}
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
	// #283: edits to this issue's own details (shaping under a claim) ride onto
	// the new branch; anything else would be swept into the issue's history.
	entries, err := env.statusEntries(env.root, "--untracked-files=no")
	if err != nil {
		return 0, err
	}
	if blocking := planningDirtyBlocking(entries, detailPath); len(blocking) > 0 {
		return 0, fmt.Errorf("%s has uncommitted tracked changes besides #%s's own details; commit or stash them before planning #%s (nothing was changed):\n  %s", env.resting, id, id, strings.Join(blocking, "\n  "))
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
		if err := refuseUnlandedBase(env, "refs/heads/"+name, pinned, id, name); err != nil {
			return 0, err
		}
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

// refuseUnlandedBase keeps one issue per branch, based on main (#272): a
// branch holding another unlanded issue branch's commits was started on (or
// absorbed) that issue's work, and planning on it would stack the two. Shared
// history alone cannot say whose work a commit is — a parent and the child
// built on it share the parent's commits — so ownership is read from the `#N`
// tag the constitution requires in commit subjects (gitx.SubjectOwnedBy); an untagged shared commit
// is left to that soft instruction. Commits main already has never count.
func refuseUnlandedBase(env *trackerEnv, tip, mainTip, id, name string) error {
	held, err := env.git("rev-list", tip, "^"+mainTip)
	if err != nil || held == "" {
		return err
	}
	onTip := map[string]bool{}
	for _, sha := range strings.Fields(held) {
		onTip[sha] = true
	}
	refs, err := env.git("for-each-ref", "--format=%(refname)", "refs/heads", "refs/remotes/"+env.target.Remote)
	if err != nil {
		return err
	}
	for _, ref := range strings.Fields(refs) {
		other := path.Base(ref)
		m := issueFamilyRE.FindStringSubmatch(other)
		if other == name || m == nil || m[1] == id {
			continue
		}
		unlanded, err := env.git("log", "--format=%H %s", ref, "^"+mainTip)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(unlanded, "\n") {
			sha, subject, _ := strings.Cut(line, " ")
			if onTip[sha] && gitx.SubjectOwnedBy(m[1], subject) {
				return fmt.Errorf("%s carries unlanded work of %s (%s %q): one issue per branch, based on main (#272).\n"+
					"      Land %s first, or restart #%s's design on a branch from main; fold the work into one issue if it is one change",
					name, other, shortSHA(sha), subject, other, id)
			}
		}
	}
	return nil
}

// planningDirtyBlocking returns the dirty tracked paths that block preparing an
// issue branch: everything except a plain modification of exactly detailPath
// (#283 D3). A rename or copy touching the details on either side blocks — the
// carried file must be the same file. Pure.
func planningDirtyBlocking(entries []gitx.StatusEntry, detailPath string) []string {
	var blocking []string
	for _, e := range entries {
		modified := strings.Trim(strings.ReplaceAll(e.XY, "M", ""), " ") == ""
		if e.Path == detailPath && e.Orig == "" && modified {
			continue
		}
		blocking = append(blocking, e.Path)
	}
	return blocking
}
