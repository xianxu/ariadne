package main

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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

// stageHolder names who a stage belongs to (#421): the issue an issue branch
// carries, else the branch, else (detached) the slot.
func stageHolder(s moveSide) string {
	if id, _, ok := issue.ParseFilename(s.Branch + ".md"); ok {
		return "#" + issue.CLIRef(id)
	}
	if s.Branch != "" {
		return "branch:" + s.Branch
	}
	return "slot:" + s.Address
}

// normalizeHolder lets --abandoned name an issue as 421, #421 or 000421.
func normalizeHolder(h string) string {
	n := strings.TrimPrefix(h, "#")
	if n != "" && strings.Trim(n, "0123456789") == "" {
		return "#" + issue.CLIRef(n)
	}
	return h
}

// checkLease is the one rule of the stage lease (#421): with no holder, or the
// same holder, me may stage into (or move into, or clear) target; another
// holder refuses, naming it and both ways out.
func checkLease(held fleet.StageLease, me, target string) error {
	if held.Holder == "" || held.Holder == me {
		return nil
	}
	return fmt.Errorf("%s is staged for %s (%s, from %s at %s); only its holder changes it. When %s is done it runs `sdlc stage --clear %s`; if that stage is abandoned, free it with `sdlc stage --clear %s --abandoned %s`",
		target, held.Holder, orDetached(held.Branch), held.Slot, held.At.Format("2006-01-02 15:04 MST"), held.Holder, target, target, held.Holder)
}

// checkClearLease adds --abandoned to checkLease: naming the current holder is
// the acknowledgement, so a lease that changed hands is never freed by it.
func checkClearLease(held fleet.StageLease, me, abandoned, target string) error {
	if abandoned == "" {
		return checkLease(held, me, target)
	}
	if got := normalizeHolder(abandoned); held.Holder == "" || got != held.Holder {
		return fmt.Errorf("--abandoned %s: %s is held by %q, not %s; nothing was cleared", abandoned, target, held.Holder, got)
	}
	return nil
}
