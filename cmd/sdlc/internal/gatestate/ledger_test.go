package gatestate

import (
	"reflect"
	"testing"
)

// TestAssignIDsAndApply pins the ID contract: the binary assigns stable sequential IDs, a
// later round refers to them, and IDs are never reused.
func TestAssignIDsAndApply(t *testing.T) {
	l := Ledger{Gate: "plan-quality", IDPrefix: "PQ"}

	// Round 1: three new findings get sequential stable IDs.
	r1 := AssignIDs(l, RoundReport{New: []Finding{
		{ID: "new", Severity: "Critical", Title: "seam in wrong layer"},
		{ID: "new", Severity: "Important", Title: "absorb layer swallows replies"},
		{ID: "new", Severity: "Minor", Title: "naming"},
	}}, 1, testTimestamp, testAgent)
	if got := ids(r1.New); !reflect.DeepEqual(got, []string{"PQ-1", "PQ-2", "PQ-3"}) {
		t.Fatalf("IDs = %v, want [PQ-1 PQ-2 PQ-3]", got)
	}
	for _, f := range r1.New {
		if f.Round != 1 {
			t.Errorf("finding %s has Round %d, want 1", f.ID, f.Round)
		}
	}
	l = Apply(l, r1)
	if n := len(OpenFindings(l)); n != 3 {
		t.Fatalf("open after round 1 = %d, want 3", n)
	}

	// Round 2: dispose two, raise one — the new ID CONTINUES the sequence. Reuse would
	// let a later round dispose the wrong finding.
	r2 := AssignIDs(l, RoundReport{
		Dispositions: dispose("PQ-1", "addressed", "PQ-2", "withdrawn"),
		New:          []Finding{{ID: "new", Severity: "Minor", Title: "comment density"}},
	}, 2, testTimestamp, testAgent)
	if r2.New[0].ID != "PQ-4" {
		t.Fatalf("new ID = %q, want PQ-4 (IDs never reuse)", r2.New[0].ID)
	}
	for _, d := range r2.Dispositions {
		if d.Round != 2 {
			t.Errorf("disposition %s has Round %d, want 2", d.ID, d.Round)
		}
	}
	l = Apply(l, r2)

	if got := ids(OpenFindings(l)); !reflect.DeepEqual(got, []string{"PQ-3", "PQ-4"}) {
		t.Fatalf("open after round 2 = %v, want [PQ-3 PQ-4]", got)
	}
}

// A disposition naming an unknown ID is a protocol error, not a silent no-op — otherwise a
// judge that hallucinates an ID quietly leaves the real finding blocking with no signal.
func TestApplyCheckedRejectsUnknownDispositionID(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Critical/seam")))
	r := round(2, dispose("PQ-99", "addressed"), nil)
	if _, err := ApplyChecked(l, r); err == nil {
		t.Error("disposing an unknown finding ID should error")
	}
}

// An unmodeled disposition must be refused at Apply even if it somehow survived parsing.
func TestApplyCheckedRejectsUnmodeledDisposition(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Critical/seam")))
	r := round(2, dispose("PQ-1", "deferred"), nil)
	if _, err := ApplyChecked(l, r); err == nil {
		t.Error("an unmodeled disposition should error")
	}
}

// A well-formed round applies cleanly.
func TestApplyCheckedAcceptsValidRound(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Critical/seam")))
	got, err := ApplyChecked(l, round(2, dispose("PQ-1", "addressed"), nil))
	if err != nil {
		t.Fatalf("valid round should apply: %v", err)
	}
	if len(got.Rounds) != 2 {
		t.Errorf("rounds = %d, want 2", len(got.Rounds))
	}
	if n := len(OpenFindings(got)); n != 0 {
		t.Errorf("open = %d, want 0", n)
	}
}

// `not-addressed` leaves the finding OPEN — the judge saying "still not fixed" must keep
// blocking, which is the difference between a gate with memory and one that forgets.
func TestOpenFindingsNotAddressedStaysOpen(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Important/absorb layer")),
		round(2, dispose("PQ-1", "not-addressed"), nil),
	)
	if got := ids(OpenFindings(l)); !reflect.DeepEqual(got, []string{"PQ-1"}) {
		t.Errorf("open = %v, want [PQ-1]", got)
	}
}

// A later disposition overrides an earlier one in both directions.
func TestOpenFindingsLastDispositionWins(t *testing.T) {
	reopened := ledgerWith(
		round(1, nil, findings("Critical/seam")),
		round(2, dispose("PQ-1", "addressed"), nil),
		round(3, dispose("PQ-1", "not-addressed"), nil),
	)
	if got := ids(OpenFindings(reopened)); !reflect.DeepEqual(got, []string{"PQ-1"}) {
		t.Errorf("re-opened: open = %v, want [PQ-1]", got)
	}
	settled := ledgerWith(
		round(1, nil, findings("Critical/seam")),
		round(2, dispose("PQ-1", "not-addressed"), nil),
		round(3, dispose("PQ-1", "addressed"), nil),
	)
	if n := len(OpenFindings(settled)); n != 0 {
		t.Errorf("settled: open = %d, want 0", n)
	}
}

// Apply must not alias the caller's backing array — two rounds applied to the same base
// ledger must not corrupt each other.
func TestApplyDoesNotAliasBaseLedger(t *testing.T) {
	base := ledgerWith(round(1, nil, findings("Critical/a")))
	a := Apply(base, round(2, nil, findings("Minor/b")))
	b := Apply(base, round(2, nil, findings("Minor/c")))
	if len(base.Rounds) != 1 {
		t.Errorf("base mutated: %d rounds", len(base.Rounds))
	}
	if a.Rounds[1].New[0].Title == b.Rounds[1].New[0].Title {
		t.Error("Apply aliased the base ledger's backing array")
	}
}

// TestDispositionCounts pins the close-time "finding disposition" tally.
func TestDispositionCounts(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/a", "Important/b", "Minor/c")),
		round(2, dispose("PQ-1", "addressed", "PQ-2", "withdrawn"), findings("Minor/d")),
	)
	addressed, withdrawn, open := DispositionCounts(l)
	if addressed != 1 || withdrawn != 1 || open != 2 {
		t.Errorf("counts = (%d,%d,%d), want (1,1,2)", addressed, withdrawn, open)
	}
}

// TestContentHashDiscriminates pins that the pass-through key actually distinguishes
// content — including a shift of text across the issue/plan boundary, which a naive
// concatenation would hash identically.
func TestContentHashDiscriminates(t *testing.T) {
	if ContentHash("issue", "plan") != ContentHash("issue", "plan") {
		t.Error("hash must be stable for identical input")
	}
	if ContentHash("issue", "plan") == ContentHash("issuep", "lan") {
		t.Error("hash must distinguish a boundary shift between issue and plan")
	}
	if ContentHash("issue", "plan") == ContentHash("issue", "plan ") {
		t.Error("hash must distinguish a one-character plan edit")
	}
}

// TestPassesUnchanged pins the short-circuit's three preconditions. Without it, B1's
// reorder would make every estimate-gate failure cost a fresh judge dispatch.
func TestPassesUnchanged(t *testing.T) {
	hash := ContentHash("issue", "plan")

	passing := ledgerWith(round(1, nil, findings("Minor/x")))
	passing.Rounds[0].Blocked = false
	passing.ContentHash = hash
	if !PassesUnchanged(passing, hash) {
		t.Error("unchanged content after a passing round must short-circuit")
	}
	if PassesUnchanged(passing, ContentHash("issue", "plan edited")) {
		t.Error("edited content must NOT short-circuit")
	}

	// Never cache away a refusal.
	blocking := ledgerWith(round(1, nil, findings("Critical/x")))
	blocking.Rounds[0].Blocked = true
	blocking.ContentHash = hash
	if PassesUnchanged(blocking, hash) {
		t.Error("a blocking round must never short-circuit")
	}

	// An empty ledger has nothing to pass through.
	if PassesUnchanged(Ledger{Gate: "plan-quality", IDPrefix: "PQ"}, hash) {
		t.Error("an empty ledger must not short-circuit")
	}
	// A ledger with rounds but no recorded hash must not short-circuit either.
	noHash := ledgerWith(round(1, nil, findings("Minor/x")))
	if PassesUnchanged(noHash, hash) {
		t.Error("a ledger with no ContentHash must not short-circuit")
	}
}
