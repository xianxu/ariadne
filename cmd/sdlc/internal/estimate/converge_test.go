package estimate

import (
	"math"
	"testing"
)

// #234: k = Σconverge / Σbuild per (model, flow, recipe), over the newest row
// of each issue that carries a split. Legacy and typed-actual rows, which have
// no split, never count; quick rows (no estimate, no model) still do, because
// converge is calibrated as a factor, not estimated.
func TestConvergeFactors(t *testing.T) {
	row := func(issue, model, flow, recipe, source string, build, converge float64) LedgerRow {
		return LedgerRow{Issue: issue, Model: model, FlowKind: flow, ReviewRecipe: recipe, SplitSource: source,
			BuildHours: build, ConvergeHours: converge, Actual: build + converge, WindowTrusted: true}
	}
	rows := []LedgerRow{
		row("a#1", "estimate-logic-v3.1", "full", "milestone-review", "dispatch", 1.0, 0.2), // superseded by the re-close below
		row("a#1", "estimate-logic-v3.1", "full", "milestone-review", "dispatch", 2.0, 1.0),
		row("a#2", "estimate-logic-v3.1", "full", "milestone-review", "finalize", 2.0, 0.0),
		row("a#3", "", "quick", "small-diff-review", "no-review", 0.5, 0.0),
		row("a#4", "", "quick", "small-diff-review", "dispatch", 0.5, 0.5),
		{Issue: "a#5", Model: "estimate-logic-v3.1", FlowKind: "full", Actual: 3, WindowTrusted: true}, // legacy
		row("a#6", "estimate-logic-v3.1", "full", "milestone-review", "manual", 0, 0),                  // typed --actual
	}
	got := ConvergeFactors(rows)
	if len(got) != 2 {
		t.Fatalf("groups = %+v, want full/milestone-review and quick/small-diff-review", got)
	}
	full, quick := got[0], got[1]
	if full.Flow != "full" || full.N != 2 || full.Build != 4 || full.Converge != 1 || math.Abs(full.K()-0.25) > 1e-9 {
		t.Errorf("full group = %+v (k %v), want n 2, build 4, converge 1, k 0.25", full, full.K())
	}
	if quick.Model != "" || quick.N != 2 || math.Abs(quick.K()-0.5) > 1e-9 {
		t.Errorf("quick group = %+v (k %v), want n 2, k 0.5", quick, quick.K())
	}
	if ConvergeFactors(rows[5:]) != nil {
		t.Error("rows without a split form no group")
	}
}
