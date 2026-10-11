package estimate

import "testing"

// trustedRow is a split row (#234) whose build equals its actual, so these
// tests read the same as they did when drift compared est with actual.
func trustedRow(est, act float64) LedgerRow {
	return LedgerRow{Estimate: est, Actual: act, WindowTrusted: true, Model: "estimate-logic-v2", BuildHours: act, SplitSource: "dispatch"}
}

func trustedModelRow(issue, model string, est, act float64) LedgerRow {
	return LedgerRow{Issue: issue, Estimate: est, Actual: act, WindowTrusted: true, Model: model, BuildHours: act, SplitSource: "dispatch"}
}

// #234: drift compares the estimate with BUILD, the part an estimate predicts,
// and reads only rows that carry a split. A legacy row's actual counts every
// earlier boundary's verification tail (the pre-#234 rule) and must not mix in.
func TestDrift_EstimateAgainstBuild(t *testing.T) {
	split := func(est, build, converge float64) LedgerRow {
		return LedgerRow{Estimate: est, Actual: build + converge, BuildHours: build, ConvergeHours: converge,
			WindowTrusted: true, Model: "estimate-logic-v3.1", SplitSource: "dispatch"}
	}
	// est/actual is ~1 (no drift), but est/build is 3 (estimates run high
	// against what they predict): drift must warn.
	over := []LedgerRow{split(3, 1, 2), split(3, 1, 2), split(3, 1, 2)}
	if warn, _ := DriftVerdict(over, 3); !warn {
		t.Error("est 3 vs build 1 is 3x over; drift must read build, not actual")
	}
	// Legacy and typed-actual rows interleaved do not count toward n.
	legacy := LedgerRow{Estimate: 3, Actual: 1, WindowTrusted: true, Model: "estimate-logic-v3.1"}
	manual := LedgerRow{Estimate: 3, Actual: 1, WindowTrusted: true, Model: "estimate-logic-v3.1", SplitSource: "manual"}
	mixed := []LedgerRow{split(3, 1, 0), legacy, split(3, 1, 0), manual}
	if warn, _ := DriftVerdict(mixed, 3); warn {
		t.Error("only two split rows: fewer than n, no verdict")
	}
}

func TestDrift_AllOver(t *testing.T) {
	rows := []LedgerRow{trustedRow(5, 0.9), trustedRow(7, 0.35), trustedRow(3, 0.5)} // all >2× over
	warn, msg := DriftVerdict(rows, 3)
	if !warn {
		t.Fatal("expected drift warning")
	}
	if msg == "" {
		t.Error("expected a non-empty drift message")
	}
}

func TestDrift_Mixed_NoWarn(t *testing.T) {
	rows := []LedgerRow{trustedRow(5, 0.9), trustedRow(1, 1.0), trustedRow(3, 0.5)} // middle is on-target
	if warn, _ := DriftVerdict(rows, 3); warn {
		t.Error("mixed ratios should not warn")
	}
}

func TestDrift_ExcludesUntrusted(t *testing.T) {
	// Two trusted-over + one untrusted-over; n=3 → only 2 trusted → no verdict.
	rows := []LedgerRow{
		trustedRow(5, 0.9),
		trustedRow(7, 0.35),
		{Estimate: 9, Actual: 0.3, WindowTrusted: false},
	}
	if warn, _ := DriftVerdict(rows, 3); warn {
		t.Error("untrusted rows must be excluded; fewer than n trusted → no warn")
	}
}

func TestDrift_AllUnder(t *testing.T) {
	rows := []LedgerRow{trustedRow(0.5, 2), trustedRow(0.3, 3), trustedRow(0.4, 4)} // all >2× under
	warn, msg := DriftVerdict(rows, 3)
	if !warn || msg == "" {
		t.Errorf("expected under-estimate drift warning, got warn=%v msg=%q", warn, msg)
	}
}

func TestDrift_TooFew(t *testing.T) {
	if warn, _ := DriftVerdict([]LedgerRow{trustedRow(5, 0.9)}, 3); warn {
		t.Error("fewer than n trusted rows should not warn")
	}
}

func TestDrift_TrustedZeroActualExcluded(t *testing.T) {
	// A trusted row with actual==0 must NOT count (ratio 0 would read as <0.5
	// "under"); it drops the trusted count below n → no verdict.
	rows := []LedgerRow{
		{Estimate: 5, Actual: 0, WindowTrusted: true},
		trustedRow(5, 0.9),
		trustedRow(7, 0.35),
	}
	if warn, _ := DriftVerdict(rows, 3); warn {
		t.Error("trusted row with actual==0 must be excluded from the trusted count")
	}
}

func TestDrift_LatestModelDoesNotInheritPriorModelRows(t *testing.T) {
	rows := []LedgerRow{
		trustedModelRow("ariadne#1", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#2", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#3", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#4", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#5", "estimate-logic-v3.1", 5, 1),
	}
	if warn, _ := DriftVerdict(rows, 5); warn {
		t.Error("a new model revision should not inherit prior-model drift rows")
	}
}

func TestDrift_DedupesRepeatedIssueRows(t *testing.T) {
	rows := []LedgerRow{
		trustedModelRow("ariadne#1", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#2", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#3", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#4", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#1", "estimate-logic-v2", 1, 1),
	}
	if warn, _ := DriftVerdict(rows, 5); warn {
		t.Error("repeated rows for one issue should count only the latest row")
	}
}

func TestDrift_LatestUnknownModelDoesNotTrigger(t *testing.T) {
	rows := []LedgerRow{
		trustedModelRow("ariadne#1", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#2", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#3", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#4", "estimate-logic-v2", 5, 1),
		trustedModelRow("ariadne#5", "", 5, 1),
	}
	if warn, _ := DriftVerdict(rows, 5); warn {
		t.Error("latest row with no recognized model should not trigger a model drift warning")
	}
}

// TestDriftSampleExcludesQuickAndUpgraded: quick-flow rows carry no estimate,
// and an upgraded row's estimate never existed (it started quick) — neither is
// an estimate↔actual data point (#231). Three over-estimates plus two such rows
// is still a verdict over the three, not a dilution.
func TestDriftSampleExcludesQuickAndUpgraded(t *testing.T) {
	quick := trustedModelRow("r#10", "estimate-logic-v2", 0, 1)
	quick.FlowKind = "quick"
	upgraded := trustedModelRow("r#11", "estimate-logic-v2", 0, 3)
	upgraded.FlowKind, upgraded.FlowUpgraded = "full", true
	rows := []LedgerRow{
		trustedModelRow("r#1", "estimate-logic-v2", 5, 0.9),
		quick,
		trustedModelRow("r#2", "estimate-logic-v2", 7, 0.35),
		upgraded,
		trustedModelRow("r#3", "estimate-logic-v2", 3, 0.5),
	}
	if warn, _ := DriftVerdict(rows, 3); !warn {
		t.Error("quick/upgraded rows must not dilute the sample")
	}
	for _, r := range driftSample(rows, 5) {
		if r.FlowKind == "quick" || r.FlowUpgraded {
			t.Errorf("driftSample kept an excluded row: %+v", r)
		}
	}
}

// TestDriftQuickRowLastDoesNotDisable: a quick row has no estimate block, so no
// model — appended LAST it used to become the "latest trusted" row, whose
// unknown model switched the drift check off at the next close (#231 survey).
func TestDriftQuickRowLastDoesNotDisable(t *testing.T) {
	quick := LedgerRow{Issue: "r#9", Actual: 0.5, WindowTrusted: true, FlowKind: "quick"}
	rows := []LedgerRow{trustedRow(5, 0.9), trustedRow(7, 0.35), trustedRow(3, 0.5), quick}
	if warn, _ := DriftVerdict(rows, 3); !warn {
		t.Error("a trailing quick row disabled the drift check")
	}
}
