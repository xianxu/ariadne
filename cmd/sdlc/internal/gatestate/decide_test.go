package gatestate

import (
	"strings"
	"testing"
)

// TestDecideConvergesWhenPriorBlockersAddressed is the issue's HEADLINE Done-when: "a
// second change-code on a plan whose Critical/Important findings were addressed passes
// without new blocking findings at a lower severity." The new Minor must not cost a round.
func TestDecideConvergesWhenPriorBlockersAddressed(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/seam", "Important/absorb-layer")),
		round(2, dispose("PQ-1", "addressed", "PQ-2", "addressed"), findings("Minor/naming")),
	)
	d := Decide(l, 3)
	if d.Block {
		t.Fatalf("gate blocked after all blockers addressed: %s", d.Reason)
	}
	if d.OpenMinor != 1 {
		t.Errorf("OpenMinor = %d, want 1 (recorded, not blocking)", d.OpenMinor)
	}
	if !strings.Contains(d.Reason, "close review") {
		t.Errorf("passing reason should say where advisory findings go, got %q", d.Reason)
	}
}

func TestDecideBlocksOnOpenCritical(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Critical/seam")))
	d := Decide(l, 3)
	if !d.Block {
		t.Fatal("an open Critical must block")
	}
	if !strings.Contains(d.Reason, "PQ-1") || !strings.Contains(d.Reason, "seam") {
		t.Errorf("reason must name the blocking finding, got %q", d.Reason)
	}
}

func TestDecideBlocksOnOpenImportant(t *testing.T) {
	if d := Decide(ledgerWith(round(1, nil, findings("Important/absorb"))), 3); !d.Block {
		t.Fatal("an open Important must block")
	}
}

func TestDecideBlocksOnNotAddressed(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Important/absorb-layer")),
		round(2, dispose("PQ-1", "not-addressed"), nil),
	)
	if d := Decide(l, 3); !d.Block {
		t.Fatal("`not-addressed` must leave the finding blocking")
	}
}

func TestDecideWithdrawnDoesNotBlock(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/mistaken")),
		round(2, dispose("PQ-1", "withdrawn"), nil),
	)
	if d := Decide(l, 3); d.Block {
		t.Fatalf("a withdrawn finding must not block: %s", d.Reason)
	}
}

// Minor never blocks, no matter how many accumulate — this is what stops the observed
// "descend to the next-deepest layer every round" loop from costing round-trips.
func TestDecideMinorNeverBlocks(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Minor/a", "Minor/b")),
		round(2, nil, findings("Minor/c", "Minor/d")),
	)
	d := Decide(l, 3)
	if d.Block {
		t.Fatalf("Minor findings must never block: %s", d.Reason)
	}
	if d.OpenMinor != 4 {
		t.Errorf("OpenMinor = %d, want 4", d.OpenMinor)
	}
}

// Past the cap, Important is recorded but no longer costs a round-trip; Critical still does.
func TestDecideRoundCapDemotesImportantButNotCritical(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Minor/a")), round(2, nil, findings("Minor/b")),
		round(3, nil, findings("Minor/c")), round(4, nil, findings("Important/late")),
	)
	d := Decide(l, 3)
	if d.Block {
		t.Fatalf("past the cap an open Important must not block: %s", d.Reason)
	}
	if !d.CapReached || len(d.Demoted) != 1 {
		t.Errorf("want CapReached with 1 demoted finding, got %v / %d", d.CapReached, len(d.Demoted))
	}
	if !strings.Contains(d.Reason, "round cap") {
		t.Errorf("reason must disclose the demotion, got %q", d.Reason)
	}

	l2 := ledgerWith(
		round(1, nil, findings("Minor/a")), round(2, nil, findings("Minor/b")),
		round(3, nil, findings("Minor/c")), round(4, nil, findings("Important/late")),
		round(5, nil, findings("Critical/real")),
	)
	if d := Decide(l2, 3); !d.Block {
		t.Fatal("an open Critical must block even past the round cap")
	}
}

// At exactly the cap the demotion has NOT yet kicked in — an off-by-one here would let a
// real Important through a round early.
func TestDecideCapIsExclusive(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Minor/a")), round(2, nil, findings("Minor/b")),
		round(3, nil, findings("Important/still-blocking")),
	)
	d := Decide(l, 3)
	if d.CapReached {
		t.Error("cap must not be reached at exactly roundCap rounds")
	}
	if !d.Block {
		t.Fatal("an Important raised at round 3 with cap 3 must still block")
	}
}

// An empty ledger (no round dispatched yet) must not block. Decide is normally consulted
// only after a round is applied, but a defensive read of a fresh ledger must be safe.
func TestDecideEmptyLedgerPasses(t *testing.T) {
	if d := Decide(Ledger{Gate: "plan-quality", IDPrefix: "PQ"}, 3); d.Block {
		t.Fatal("an empty ledger must not block")
	}
}

// A zero/negative cap falls back to the default rather than demoting everything at round 1.
func TestDecideZeroCapUsesDefault(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Important/x")))
	if d := Decide(l, 0); !d.Block {
		t.Fatal("cap 0 must fall back to DefaultRoundCap, not demote immediately")
	}
}
