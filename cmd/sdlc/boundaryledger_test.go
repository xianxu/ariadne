package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// A missing sidecar is the normal round-1 state — an empty ledger carrying the gate's
// identity, not an error.
func TestReadBoundaryGateLedger_MissingIsEmptyNotError(t *testing.T) {
	dir := t.TempDir()
	l, err := readBoundaryGateLedger(dir, "000194-x.md", 194)
	if err != nil {
		t.Fatalf("missing sidecar must not error: %v", err)
	}
	if len(l.Rounds) != 0 || l.IDPrefix != "BR" || l.Gate != "boundary-review" || l.IssueNum != 194 {
		t.Fatalf("want an empty BR ledger for #194, got %+v", l)
	}
}

// The behavior that makes the ledger trustworthy: a sidecar that EXISTS but does not
// parse is an ERROR. Silently resetting would erase every disposition and re-open
// findings the operator already addressed — the exact forgetting this gate prevents,
// and worse than the status quo because it would look like it worked.
func TestReadBoundaryGateLedger_CorruptIsErrorNotSilentReset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "000194-x-close-gate.md")
	if err := os.WriteFile(path, []byte("---\nthis: [is not: valid yaml\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := readBoundaryGateLedger(dir, "000194-x.md", 194)
	if err == nil {
		t.Fatal("a corrupt ledger must error, never silently reset to empty")
	}
	if !strings.Contains(err.Error(), "do NOT let the gate silently forget") {
		t.Errorf("the error must say why it refuses, got: %v", err)
	}
}

// The ledger is durable state the NEXT invocation reads, so the round-trip through disk
// must preserve ids, dispositions and boundaries.
func TestBoundaryGateLedger_RoundTripsThroughDisk(t *testing.T) {
	dir := t.TempDir()
	l := gatestate.Ledger{Gate: "boundary-review", IssueNum: 194, IDPrefix: "BR", Rounds: []gatestate.Round{
		{N: 1, Boundary: "M1", New: []gatestate.Finding{{ID: "BR-1", Severity: "Important", Title: "first", Round: 1}}},
		{N: 2, Boundary: "M1", Dispositions: []gatestate.Disposition{{ID: "BR-1", State: "addressed", Round: 2}}},
	}}
	if err := writeBoundaryGateLedger(dir, "000194-x.md", l, "ariadne"); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := readBoundaryGateLedger(dir, "000194-x.md", 194)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got.Rounds) != 2 {
		t.Fatalf("got %d rounds, want 2", len(got.Rounds))
	}
	if open := gatestate.OpenFindings(got); len(open) != 0 {
		t.Errorf("BR-1 was disposed addressed; it must not read as open: %+v", open)
	}
	if got.Rounds[0].Boundary != "M1" {
		t.Errorf("boundary lost in round-trip: %q", got.Rounds[0].Boundary)
	}
}

// The prose sidecar and the gate ledger are DIFFERENT artifacts for the same boundary and
// must not collide — verdict.cue's `*-review.md` glob asserts "carries a boundary
// verdict", which a findings ledger does not.
func TestBoundaryGateLedger_PathDoesNotCollideWithProseSidecar(t *testing.T) {
	ledger := boundaryGatePath("workshop/plans", "000194-x.md")
	for _, milestone := range []string{"", "M1"} {
		prose := sidecarPath("workshop/plans", "000194-x.md", milestone)
		if ledger == prose {
			t.Fatalf("ledger path collides with the prose sidecar: %s", ledger)
		}
	}
	if strings.HasSuffix(ledger, "-review.md") {
		t.Errorf("the ledger must stay out of verdict.cue's *-review.md family: %s", ledger)
	}
}

// #194 M2 end-to-end: a boundary review that emits the findings fence persists a ledger,
// and the NEXT review at the same boundary is shown those findings to dispose of.
func TestBoundaryReview_PersistsLedgerAndFeedsTheNextRound(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	p := boundaryReviewParams{
		Label: "#69 M1", Base: "abc1234", BaseLong: "abc1234", Head: "def5678",
		IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1", PlansDir: plansDir,
	}
	round := gatestate.RoundReport{New: []gatestate.Finding{
		{ID: "new", Severity: "Important", Title: "the oracle cannot see what it certifies"},
	}}
	var stderr strings.Builder
	d := persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &round}, "2026-08-20T18:00:00-07:00")

	if !d.Block || len(d.OpenBlocking) != 1 {
		t.Fatalf("an undisposed Important must block the boundary: %+v", d)
	}
	if d.OpenBlocking[0].ID != "BR-1" {
		t.Errorf("the binary assigns the stable id, got %q", d.OpenBlocking[0].ID)
	}
	// The next round at this boundary is shown it.
	prior := boundaryPriorFindings(&stderr, p)
	if !strings.Contains(prior, "BR-1") || !strings.Contains(prior, "oracle cannot see") {
		t.Errorf("the next round must be shown BR-1 to dispose of:\n%s", prior)
	}
	// A different boundary is not — the cap and the open set scope per boundary (D1).
	other := p
	other.Milestone = "M2"
	if got := boundaryPriorFindings(&stderr, other); strings.Contains(got, "BR-1") {
		t.Errorf("M1's finding must not block M2:\n%s", got)
	}
}

// Disposing the finding clears the boundary — the convergence mechanic.
func TestBoundaryReview_DisposingAFindingClearsTheGate(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	p := boundaryReviewParams{
		Label: "#69 M1", IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1", PlansDir: plansDir,
	}
	var stderr strings.Builder
	persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{
		New: []gatestate.Finding{{ID: "new", Severity: "Critical", Title: "boom"}},
	}}, "2026-08-20T18:00:00-07:00")

	d := persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{
		Dispositions: []gatestate.Disposition{{ID: "BR-1", State: "addressed", Note: "fixed"}},
	}}, "2026-08-20T18:30:00-07:00")
	if d.Block {
		t.Errorf("disposing the only blocking finding must clear the gate: %+v", d)
	}
	if d.Rounds != 2 {
		t.Errorf("both rounds must be recorded, got %d", d.Rounds)
	}
}

// #194 D2: the plan gate's still-open findings are SEEDED into this ledger on its first
// round, under this gate's id namespace — replacing code-review.md's instruction to read
// the plan-gate file directly, which would have put PQ-* and BR-* ids in one fence.
func TestBoundaryReview_SeedsDeferredPlanGateFindings(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	plansDir := t.TempDir()
	issueFile := filepath.Base(mustIssuePath(t, issuesDir, 69))

	planLedger := gatestate.Ledger{Gate: "plan-quality", IssueNum: 69, IDPrefix: "PQ", Rounds: []gatestate.Round{
		{N: 1, New: []gatestate.Finding{
			{ID: "PQ-1", Severity: "Minor", Title: "deferred to the boundary", Round: 1},
			{ID: "PQ-2", Severity: "Important", Title: "already handled", Round: 1},
		}},
		{N: 2, Dispositions: []gatestate.Disposition{{ID: "PQ-2", State: "addressed", Round: 2}}},
	}}
	if err := writePlanGateLedger(plansDir, issueFile, planLedger, "ariadne"); err != nil {
		t.Fatal(err)
	}

	p := boundaryReviewParams{IssuesDir: issuesDir, IssueNum: 69, Milestone: "M1", PlansDir: plansDir}
	var stderr strings.Builder
	persistBoundaryRound(&stderr, p, reviewResult{Agent: "claude", Round: &gatestate.RoundReport{}}, "2026-08-20T18:00:00-07:00")

	l, err := readBoundaryGateLedger(plansDir, issueFile, 69)
	if err != nil {
		t.Fatal(err)
	}
	open := gatestate.OpenFindings(l)
	if len(open) != 1 {
		t.Fatalf("only the still-open plan-gate finding seeds, got %+v", open)
	}
	if open[0].Severity != "Minor" || !strings.Contains(open[0].Title, "deferred to the boundary") {
		t.Errorf("seeded finding lost its identity: %+v", open[0])
	}
	if !strings.Contains(open[0].Detail, "PQ-1") {
		t.Errorf("the seeded finding must record its plan-gate origin: %q", open[0].Detail)
	}
	// D5: seeded findings are visible at EVERY boundary, since they were deferred to
	// "the boundary review" generically, not to whichever milestone ran first.
	for _, boundary := range []string{"M1", "M2", ""} {
		q := p
		q.Milestone = boundary
		if got := boundaryPriorFindings(&stderr, q); !strings.Contains(got, "deferred to the boundary") {
			t.Errorf("seeded finding must be visible at boundary %q:\n%s", boundary, got)
		}
	}
}

func mustIssuePath(t *testing.T, issuesDir string, n int) string {
	t.Helper()
	p, err := issueFilePath(issuesDir, n)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
