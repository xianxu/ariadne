package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// moveSide is one slot as `sdlc move` observed it (#260).
type moveSide struct {
	Root, Address, Branch, Resting, Head string
	RestHead                             string             // commit Resting points at
	Changes                              []gitx.StatusEntry // tracked, staged, submodule
	Untracked                            []string
	Operation                            string
}

// moveFacts is everything checkMove decides on. It is compared with
// reflect.DeepEqual to prove nothing changed between preflight and switch.
type moveFacts struct {
	From, To    moveSide
	SameRepo    bool
	Incoming    []string // paths From.Branch checks out
	Parked      []string // To.Resting commits absent from From.Branch (reported only)
	Unpublished []string // To.Resting commits absent from its upstream (reported only)
}

// checkMove holds every refusal rule of `sdlc move`; nil means the two
// switches may run. Parked commits are not a refusal: the resting ref is
// untouched, so they are only absent from the test.
func checkMove(f moveFacts) error {
	from, to := f.From, f.To
	switch {
	case !f.SameRepo:
		return fmt.Errorf("%s and %s are different repositories", from.Root, to.Root)
	case from.Root == to.Root:
		return fmt.Errorf("%s is already the current slot", to.Address)
	case from.Branch == "":
		return fmt.Errorf("%s is detached; switch to the issue branch to move", from.Root)
	case from.Branch == from.Resting:
		return fmt.Errorf("%s is on its resting branch %s; there is no issue branch to move", from.Address, from.Resting)
	case to.Branch != to.Resting:
		return fmt.Errorf("%s is on %s, not its resting branch %s; move that branch out first", to.Address, orDetached(to.Branch), to.Resting)
	}
	for _, s := range []moveSide{from, to} {
		if s.Operation != "" {
			return fmt.Errorf("%s has a Git operation in progress (%s); finish or abort it first", s.Address, s.Operation)
		}
		if len(s.Changes) > 0 {
			return fmt.Errorf("%s has uncommitted changes (%s); commit or remove them first", s.Address, statusLine(s.Changes[0]))
		}
	}
	if len(from.Untracked) > 0 {
		return fmt.Errorf("%s has untracked files (%s); commit or remove them first", from.Address, from.Untracked[0])
	}
	if c := untrackedCollisions(to.Untracked, f.Incoming); len(c) > 0 {
		return fmt.Errorf("%s has untracked files that %s would overwrite: %s; move or remove them first", to.Address, from.Branch, strings.Join(c, ", "))
	}
	return nil
}

func orDetached(branch string) string {
	if branch == "" {
		return "a detached HEAD"
	}
	return branch
}

// untrackedCollisions returns the untracked paths that checking out incoming
// would overwrite: the same path, or one being a directory prefix of the other.
func untrackedCollisions(untracked, incoming []string) []string {
	var out []string
	for _, u := range untracked {
		for _, p := range incoming {
			if u == p || strings.HasPrefix(p, u+"/") || strings.HasPrefix(u, p+"/") {
				out = append(out, u)
				break
			}
		}
	}
	sort.Strings(out)
	return out
}
