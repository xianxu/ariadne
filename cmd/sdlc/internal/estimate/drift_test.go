package estimate

import "testing"

func trustedRow(est, act float64) LedgerRow {
	return LedgerRow{Estimate: est, Actual: act, WindowTrusted: true, Model: "estimate-logic-v2"}
}

func trustedModelRow(issue, model string, est, act float64) LedgerRow {
	return LedgerRow{Issue: issue, Estimate: est, Actual: act, WindowTrusted: true, Model: model}
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
