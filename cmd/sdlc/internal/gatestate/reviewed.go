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
