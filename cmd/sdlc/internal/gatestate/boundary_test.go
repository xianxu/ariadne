package gatestate

import (
	"strings"
	"testing"
)

// #194 D1: one ledger per issue, but the round cap and open-findings set scope per
// BOUNDARY. Decide computes CapReached from len(l.Rounds) against DefaultRoundCap, so an
// unfiltered issue-wide boundary ledger would arrive at the whole-issue close already
// past the cap — silently demoting every Important finding on round one of the gate this
// work exists to strengthen. FilterBoundary is the caller-side pure transform that keeps
// Decide's and OpenFindings' signatures untouched.
func TestFilterBoundary_ScopesRoundsToOneBoundary(t *testing.T) {
	l := Ledger{Gate: "boundary-review", IssueNum: 194, IDPrefix: "BR", Rounds: []Round{
		{N: 1, Boundary: "M1", New: []Finding{{ID: "BR-1", Severity: "Important", Title: "m1"}}},
		{N: 2, Boundary: "M2", New: []Finding{{ID: "BR-2", Severity: "Important", Title: "m2"}}},
		{N: 3, Boundary: "M2", New: []Finding{{ID: "BR-3", Severity: "Minor", Title: "m2 again"}}},
		{N: 4, Boundary: "", New: []Finding{{ID: "BR-4", Severity: "Critical", Title: "close"}}},
	}}

	for _, tc := range []struct {
		boundary   string
		wantRounds int
		wantIDs    []string
	}{
		{"M1", 1, []string{"BR-1"}},
		{"M2", 2, []string{"BR-2", "BR-3"}},
		{"", 1, []string{"BR-4"}}, // the whole-issue close is its own boundary
	} {
		got := FilterBoundary(l, tc.boundary)
		if len(got.Rounds) != tc.wantRounds {
			t.Errorf("boundary %q: %d rounds, want %d", tc.boundary, len(got.Rounds), tc.wantRounds)
		}
		open := OpenFindings(got)
		if len(open) != len(tc.wantIDs) {
			t.Fatalf("boundary %q: %d open findings, want %d", tc.boundary, len(open), len(tc.wantIDs))
		}
		for i, want := range tc.wantIDs {
			if open[i].ID != want {
				t.Errorf("boundary %q: open[%d] = %s, want %s", tc.boundary, i, open[i].ID, want)
			}
		}
	}

	// The identity fields survive the transform — a filtered ledger is still THE ledger.
	if got := FilterBoundary(l, "M1"); got.IDPrefix != "BR" || got.IssueNum != 194 || got.Gate != "boundary-review" {
		t.Errorf("FilterBoundary dropped identity: %+v", got)
	}
}

// #194 D1: the round cap is what forced the scoping. Unfiltered, three milestones'
// rounds put the whole-issue close past DefaultRoundCap before it raises anything.
func TestFilterBoundary_KeepsEachBoundaryUnderTheRoundCap(t *testing.T) {
	l := Ledger{IDPrefix: "BR", Rounds: []Round{
		{N: 1, Boundary: "M1"}, {N: 2, Boundary: "M1"},
		{N: 3, Boundary: "M2"}, {N: 4, Boundary: "M3"},
		{N: 5, Boundary: "", New: []Finding{{ID: "BR-1", Severity: "Important", Title: "real"}}},
	}}
	if d := Decide(l, DefaultRoundCap); !d.CapReached {
		t.Fatal("precondition: the unfiltered ledger should already be past the cap")
	} else if len(d.Demoted) != 1 || len(d.OpenBlocking) != 0 {
		t.Fatalf("precondition: unfiltered demotes the finding, got demoted=%d blocking=%d", len(d.Demoted), len(d.OpenBlocking))
	}
	d := Decide(FilterBoundary(l, ""), DefaultRoundCap)
	if d.CapReached {
		t.Error("the whole-issue close is on its FIRST round; the cap must not be reached")
	}
	if len(d.OpenBlocking) != 1 || len(d.Demoted) != 0 {
		t.Errorf("an Important finding on round 1 must block, got blocking=%d demoted=%d", len(d.OpenBlocking), len(d.Demoted))
	}
}

// #194 D5: D1 and D2 do not compose on their own. A round SEEDED from the plan gate
// belongs to no single boundary — those findings were deferred to "the boundary review"
// generically (code-review.md's carry-forward), so scoping them to the boundary that
// happened to run first would make them invisible everywhere else. That would be a
// REGRESSION against the behavior being replaced.
func TestFilterBoundary_SeededFindingsAreVisibleAtEveryBoundary(t *testing.T) {
	l := Ledger{IDPrefix: "BR", Rounds: []Round{
		{N: 1, Boundary: BoundaryAll, New: []Finding{{ID: "BR-1", Severity: "Important", Title: "deferred from the plan gate"}}},
		{N: 2, Boundary: "M1", New: []Finding{{ID: "BR-2", Severity: "Minor", Title: "m1 only"}}},
	}}
	for _, boundary := range []string{"M1", "M2", ""} {
		open := OpenFindings(FilterBoundary(l, boundary))
		found := false
		for _, f := range open {
			if f.ID == "BR-1" {
				found = true
			}
		}
		if !found {
			t.Errorf("seeded finding BR-1 must be visible at boundary %q, got %+v", boundary, open)
		}
	}
	// ...and a boundary-scoped finding still is not.
	if open := OpenFindings(FilterBoundary(l, "M2")); len(open) != 1 {
		t.Errorf("M2 must see only the seeded finding, got %+v", open)
	}
}

// The boundary must survive the durable round-trip — it is persisted state the NEXT
// invocation scopes on, so a field that renders but does not parse back would silently
// collapse every round into the whole-issue boundary.
func TestBoundary_SurvivesSidecarRoundTrip(t *testing.T) {
	l := Ledger{Gate: "boundary-review", IssueNum: 194, IDPrefix: "BR", Rounds: []Round{
		{N: 1, Boundary: BoundaryAll, New: []Finding{{ID: "BR-1", Severity: "Important", Title: "seeded"}}},
		{N: 2, Boundary: "M1", New: []Finding{{ID: "BR-2", Severity: "Minor", Title: "m1"}}},
		{N: 3, Boundary: "", New: []Finding{{ID: "BR-3", Severity: "Critical", Title: "close"}}},
	}}
	got, err := ParseSidecar(Render(l, "ariadne"))
	if err != nil {
		t.Fatalf("ParseSidecar: %v", err)
	}
	if len(got.Rounds) != 3 {
		t.Fatalf("got %d rounds, want 3", len(got.Rounds))
	}
	for i, want := range []string{BoundaryAll, "M1", ""} {
		if got.Rounds[i].Boundary != want {
			t.Errorf("round %d boundary = %q, want %q", i+1, got.Rounds[i].Boundary, want)
		}
	}
}

// A plan-quality ledger has one boundary by construction, so its rounds must stay
// byte-identical to their pre-#194 form — omitempty, no stray `boundary: ""` key.
func TestBoundary_OmittedForSingleBoundaryGates(t *testing.T) {
	rendered := Render(Ledger{Gate: "plan-quality", IssueNum: 187, IDPrefix: "PQ", Rounds: []Round{
		{N: 1, New: []Finding{{ID: "PQ-1", Severity: "Important", Title: "x"}}},
	}}, "ariadne")
	if strings.Contains(rendered, "boundary:") {
		t.Errorf("a single-boundary gate must not render a boundary key:\n%s", rendered)
	}
}
