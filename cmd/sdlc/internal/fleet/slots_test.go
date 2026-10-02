package fleet

import (
	"strings"
	"testing"
)

// #289: the per-checkout verdict over probe outcome × working-tree state ×
// branch state × claim. Needs-recovery is reported whenever its facts were
// read; otherwise a failed probe is unknown; otherwise unlanded commits, an
// open issue or a claim hold work; otherwise ready.
func TestJudgeCheckout(t *testing.T) {
	n := func(v int) *int { return &v }
	ok := MeasuredFacts{Available: true, Head: "h", BaseAvailable: true, BaseRef: "origin/main", Ahead: n(0), Behind: n(0), DirtyCount: n(0)}
	with := func(f func(*TreeRow)) TreeRow {
		r := TreeRow{Branch: "main-slot1", Facts: ok, Issues: []IssueAssociation{}, Claims: []ClaimAssociation{}, ClaimsState: ClaimsPresent}
		f(&r)
		return r
	}
	openIssue := []IssueAssociation{{Ref: "r#000007", DeclaredStatus: "working", Provenance: IssueProvenanceBranchPrefix}}
	doneIssue := []IssueAssociation{{Ref: "r#000007", DeclaredStatus: "done", Provenance: IssueProvenanceBranchPrefix}}
	for _, tc := range []struct {
		name    string
		row     TreeRow
		verdict Verdict
		reasons string // comma-joined reason codes, in order
	}{
		{"clean on resting", with(func(*TreeRow) {}), VerdictReady, ""},
		{"merged feature branch", with(func(r *TreeRow) { r.Branch = "topic" }), VerdictReady, ""},
		{"closed-issue branch", with(func(r *TreeRow) { r.Branch = "000007-x"; r.Issues = doneIssue }), VerdictReady, ""},
		{"unlanded commits", with(func(r *TreeRow) { r.Branch = "topic"; r.Facts.Ahead = n(2) }), VerdictHoldsWork, "unlanded-commits"},
		{"unlanded on resting", with(func(r *TreeRow) { r.Facts.Ahead = n(1) }), VerdictHoldsWork, "unlanded-commits"},
		{"zero-commit open-issue branch", with(func(r *TreeRow) { r.Branch = "000007-x"; r.Issues = openIssue }), VerdictHoldsWork, "open-issue:r#000007"},
		{"claimed on resting", with(func(r *TreeRow) {
			r.Claims = []ClaimAssociation{{Ref: "r#000009", Status: "working"}}
		}), VerdictHoldsWork, "claimed:r#000009"},
		{"dirty", with(func(r *TreeRow) { r.Facts.DirtyCount = n(3) }), VerdictNeedsRecovery, "dirty"},
		{"operation", with(func(r *TreeRow) { r.Facts.Operation = "rebase-merge" }), VerdictNeedsRecovery, "operation:rebase-merge"},
		{"detached", with(func(r *TreeRow) { r.Branch, r.Detached = "", true }), VerdictNeedsRecovery, "detached"},
		{"dirty and unlanded", with(func(r *TreeRow) { r.Facts.DirtyCount = n(1); r.Branch = "topic"; r.Facts.Ahead = n(1) }), VerdictNeedsRecovery, "dirty"},
		{"facts unavailable", with(func(r *TreeRow) { r.Facts = MeasuredFacts{Error: "status failed"} }), VerdictUnknown, "probe:facts"},
		{"base unavailable", with(func(r *TreeRow) {
			r.Facts.BaseAvailable, r.Facts.BaseError, r.Facts.Ahead, r.Facts.Behind = false, "no base", nil, nil
		}), VerdictUnknown, "probe:base"},
		{"base unavailable but dirty", with(func(r *TreeRow) {
			r.Facts.BaseAvailable, r.Facts.BaseError, r.Facts.Ahead, r.Facts.Behind = false, "no base", nil, nil
			r.Facts.DirtyCount = n(1)
		}), VerdictNeedsRecovery, "dirty,probe:base"},
		{"operation probe failed", with(func(r *TreeRow) { r.Facts.OperationError = "denied" }), VerdictUnknown, "probe:operation"},
		{"issue branch, lookup failed", with(func(r *TreeRow) { r.Branch, r.IssuesError = "000007-x", "lookup issue 000007: tracker unreadable" }), VerdictUnknown, "probe:issue"},
		{"issue branch, lookup ok, no match", with(func(r *TreeRow) { r.Branch = "000007-x" }), VerdictReady, ""},
		{"claims unread, otherwise ready", with(func(r *TreeRow) { r.ClaimsState, r.ClaimsError = ClaimsUnknown, "x" }), VerdictUnknown, "probe:claims"},
		{"claims unread, already holding work", with(func(r *TreeRow) {
			r.ClaimsState, r.ClaimsError = ClaimsUnknown, "x"
			r.Branch, r.Facts.Ahead = "topic", n(1)
		}), VerdictHoldsWork, "unlanded-commits"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := JudgeCheckout(tc.row, "main-slot1")
			if got.Verdict != tc.verdict || strings.Join(got.Reasons, ",") != tc.reasons {
				t.Fatalf("got %s %v, want %s %s", got.Verdict, got.Reasons, tc.verdict, tc.reasons)
			}
			if strings.Contains(tc.reasons, "probe:") && len(got.Errors) == 0 {
				t.Fatalf("a failed probe must carry its error: %+v", got)
			}
		})
	}
}

// #289: the verdict order the slot fold uses.
func TestVerdictOrder(t *testing.T) {
	order := []Verdict{VerdictReady, VerdictHoldsWork, VerdictUnknown, VerdictMissing, VerdictNeedsRecovery}
	for i := 1; i < len(order); i++ {
		if order[i-1].rank() >= order[i].rank() {
			t.Fatalf("%s must rank below %s", order[i-1], order[i])
		}
	}
	if Worst(VerdictHoldsWork, VerdictMissing, VerdictReady) != VerdictMissing {
		t.Fatal("worst-of")
	}
}
