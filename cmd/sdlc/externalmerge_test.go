package main

import (
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// #287: every combination of the facts. Landed evidence settles, whoever
// owns the card; a merged branch without it needs a close from the owner, or
// a claim first; anything else is not an outside merge.
func TestExternalMergeVerdict(t *testing.T) {
	for _, evidence := range []bool{false, true} {
		for _, merged := range []bool{false, true} {
			for _, own := range []issue.Ownership{issue.OwnershipMine, issue.OwnershipForeign, issue.OwnershipUnknown} {
				f := mergeFacts{EvidenceOnMain: evidence, BranchMerged: merged, Owner: own}
				want := mergeNone
				switch {
				case evidence:
					want = mergeSettle
				case merged && own == issue.OwnershipUnknown:
					want = mergeClaimNeeded
				case merged:
					want = mergeCloseNeeded
				}
				if got := externalMergeVerdict(f); got != want {
					t.Errorf("%+v: got %v, want %v", f, got, want)
				}
			}
		}
	}
}
