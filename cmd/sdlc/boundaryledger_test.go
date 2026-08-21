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
