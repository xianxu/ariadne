package main

import (
	"fmt"
	"strings"
)

// stageFacts is everything checkStage decides on (#344): the slot staging its
// HEAD and the slot it stages into, observed as `sdlc move` observes them.
type stageFacts struct {
	From, To moveSide
	SameRepo bool
	Incoming []string // paths From.Head checks out
}

// checkStage holds every refusal rule of `sdlc stage`; nil means the target
// may be detached at From.Head. The source is never changed, so only its HEAD
// matters; the target must be clean and either resting or a leavable detach.
func checkStage(f stageFacts) error {
	from, to := f.From, f.To
	switch {
	case !f.SameRepo:
		return fmt.Errorf("%s and %s are different repositories", from.Root, to.Root)
	case from.Root == to.Root:
		return fmt.Errorf("%s is the current slot; stage into another one", to.Address)
	case from.Head == "":
		return fmt.Errorf("%s has no commit to stage", from.Root)
	}
	if err := checkTargetClean(to); err != nil {
		return err
	}
	if to.Branch != to.Resting && !(to.Branch == "" && to.Leavable) {
		return fmt.Errorf("%s is on %s, not its resting branch %s; %s", to.Address, orDetached(to.Branch), to.Resting, outOfTheWay(to))
	}
	if c := untrackedCollisions(to.Untracked, f.Incoming); len(c) > 0 {
		return fmt.Errorf("%s has untracked files that %s would overwrite: %s; move or remove them first", to.Address, shortOID(from.Head), strings.Join(c, ", "))
	}
	return nil
}

// checkUnstage holds the rules of `sdlc stage --clear`: a leavable, clean
// detached target returns to its resting branch. staged reports whether there
// is anything to clear (a target already resting has nothing).
func checkUnstage(to moveSide) (staged bool, err error) {
	switch {
	case to.Branch == to.Resting:
		return false, nil
	case to.Branch != "":
		return false, fmt.Errorf("%s is on %s, not a staged commit; nothing to clear", to.Address, to.Branch)
	case !to.Leavable:
		return false, fmt.Errorf("%s is on %s, not its resting branch %s; %s", to.Address, orDetached(to.Branch), to.Resting, outOfTheWay(to))
	}
	return true, checkTargetClean(to)
}

func checkTargetClean(to moveSide) error {
	if to.Operation != "" {
		return fmt.Errorf("%s has a Git operation in progress (%s); finish or abort it first", to.Address, to.Operation)
	}
	if len(to.Changes) > 0 {
		return fmt.Errorf("%s has uncommitted changes (%s); commit or remove them first", to.Address, statusLine(to.Changes[0]))
	}
	return nil
}
