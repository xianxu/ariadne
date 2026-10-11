package gatestate

// LatestReviewed returns where review last stopped outside `exclude` (#304): the head
// of the newest round that carries a Reviewed commit and whose boundary is neither
// `exclude` nor BoundaryAll. It is the single answer to "what did the last finalized
// review read", which the window planner diffs against — replacing the commit-message
// grep for a hand-pasted Review-Verdict trailer (#197).
func LatestReviewed(l Ledger, exclude string) (sha, boundary string, ok bool) {
	for i := len(l.Rounds) - 1; i >= 0; i-- {
		r := l.Rounds[i]
		if r.Reviewed == "" || r.Boundary == exclude || r.Boundary == BoundaryAll {
			continue
		}
		return r.Reviewed, r.Boundary, true
	}
	return "", "", false
}

// LatestFinalizedAt returns the head the newest finalized round AT `boundary` read
// (#183): the anchor `sdlc close --fixed-then-ship` diffs from. LatestReviewed answers
// a different question (the last review OUTSIDE a boundary, for a milestone's
// interdiff); this one is the boundary's own last finalized review. A round with no
// Reviewed head (REWORK, halt, never ran) never finalized, so it never anchors.
func LatestFinalizedAt(l Ledger, boundary string) (sha string, ok bool) {
	for i := len(l.Rounds) - 1; i >= 0; i-- {
		if r := l.Rounds[i]; r.Boundary == boundary && r.Reviewed != "" {
			return r.Reviewed, true
		}
	}
	return "", false
}
