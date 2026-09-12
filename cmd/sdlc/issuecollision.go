// issuecollision.go — the pure decision behind #207's publish path: given an id,
// the path we mean to write, and the trunk's id space, decide whether to publish,
// re-allocate, or refuse.
//
// It is separate from the git calls because the decision is where the design
// lives and the git is incidental (ARCH-PURE) — but more concretely, because an
// earlier draft of #207 collapsed these three outcomes into one ("if the id is
// taken, re-allocate") and would have renumbered every existing issue on every
// `sdlc issue sync`. Three outcomes need three names.
package main

import (
	"path/filepath"
	"sort"
)

type collisionVerdict int

const (
	// verdictPublish — the id is ours or free.
	verdictPublish collisionVerdict = iota
	// verdictReallocate — a different slug holds this id and NOTHING references
	// ours yet, so stepping aside is cheap.
	verdictReallocate
	// verdictRefuse — a different slug holds this id and ours is already
	// referenced. Renumbering here is worse than the collision.
	verdictRefuse
)

func (v collisionVerdict) String() string {
	switch v {
	case verdictPublish:
		return "publish"
	case verdictReallocate:
		return "re-allocate"
	case verdictRefuse:
		return "refuse"
	}
	return "unknown"
}

// decideCollision returns the verdict and, when it is not publish, the trunk
// paths that conflict — so the caller can name them rather than reporting a
// generic contention error.
//
// firstPublication is the whole hinge. Renumbering is safe ONLY before anything
// references the id; by claim time it has leaked into the branch name, and after
// that into commit subjects agents grep, `deps:` in sibling issues, and review
// sidecar filenames (ariadne#188). So `issue new` re-allocates and `issue
// sync`/`claim` refuse, and the caller says which it is.
func decideCollision(id int, myPath string, space map[int][]string, pending map[string]bool, firstPublication bool) (collisionVerdict, []string) {
	var foreign []string
	mineOnTrunk := false
	for _, p := range space[id] {
		// This publish REMOVES it. A slug rename is one atomic commit carrying
		// Delete(old) and Write(new); reading the raw trunk made that refuse
		// forever, since the old name is always still there when we look
		// (#207 BR-19).
		if pending[p] {
			continue
		}
		// Same file NAME in a different directory is the same issue moving —
		// archiving to history, or being restored from it — not a collision.
		if filepath.Base(p) == filepath.Base(myPath) {
			mineOnTrunk = true
			continue
		}
		foreign = append(foreign, p)
	}
	sort.Strings(foreign)

	switch {
	case len(foreign) == 0:
		// Free, or holding only our own prior publication. Republishing our own
		// body is the common case and must never be read as a collision.
		return verdictPublish, nil
	case mineOnTrunk:
		// Ours AND a foreign path: the trunk already carries a duplicate. Never
		// re-allocate here — our id is published and therefore referenced.
		return verdictRefuse, foreign
	case firstPublication:
		return verdictReallocate, foreign
	default:
		return verdictRefuse, foreign
	}
}

// nextFreeID walks up from `from` past every id the trunk holds.
//
// It starts from the id we wanted rather than from the space's maximum, because
// the goal is the next free slot, and a gap below the maximum is a legitimate
// one to take.
func nextFreeID(from int, space map[int][]string) int {
	for id := from; ; id++ {
		if len(space[id]) == 0 {
			return id
		}
	}
}
