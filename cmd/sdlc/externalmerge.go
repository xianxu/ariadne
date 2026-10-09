// externalmerge.go — merges done outside sdlc (#287). An issue branch merged
// into main by the GitHub button (not `sdlc merge`) leaves its card short of
// done. sdlc notices from git alone — the pushed branch tip (#286 keeps it
// current) or the close's evidence commit is on main — reports it read-only
// in `sdlc state`, and finishes the bookkeeping in reconcile and the next
// merge. Squash and rebase merges leave neither on main and are not seen.
package main

import "github.com/xianxu/ariadne/cmd/sdlc/internal/issue"

type mergeVerdict int

const (
	mergeNone        mergeVerdict = iota
	mergeSettle                   // the close is on main: done, archive, delete the branch
	mergeCloseNeeded              // merged without a close: the owner closes it
	mergeClaimNeeded              // merged, unowned: claim, then close
)

// mergeFacts are what one started issue is judged on.
type mergeFacts struct {
	EvidenceOnMain bool // the completion binding's evidence commit is on main
	BranchMerged   bool // the pushed issue branch's tip is on main
	Owner          issue.Ownership
}

// externalMergeVerdict classifies an issue against main. Landed evidence
// binds the close, so it settles whoever owns the card.
func externalMergeVerdict(f mergeFacts) mergeVerdict {
	switch {
	case f.EvidenceOnMain:
		return mergeSettle
	case !f.BranchMerged:
		return mergeNone
	case f.Owner == issue.OwnershipUnknown:
		return mergeClaimNeeded
	}
	return mergeCloseNeeded
}
