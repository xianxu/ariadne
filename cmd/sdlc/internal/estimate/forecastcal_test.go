package estimate

import (
	"math"
	"testing"
)

func splitRow(issue string, est, build, conv float64) LedgerRow {
	return LedgerRow{Issue: issue, Estimate: est, BuildHours: build, ConvergeHours: conv, SplitSource: "boundary", FlowKind: "full"}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestCalibrateForecast(t *testing.T) {
	rows := []LedgerRow{
		splitRow("1", 2, 1, 0.5), // m = 0.75
		splitRow("2", 4, 2, 2),   // m = 1.0
		splitRow("3", 1, 1, 1),   // m = 2.0
		splitRow("3", 1, 9, 9),   // superseded: newest-per-issue keeps this one → m = 18
		{Issue: "4", Estimate: 3, BuildHours: 1, FlowKind: "quick", SplitSource: "boundary"}, // quick: excluded
		{Issue: "5", Estimate: 3, BuildHours: 1, SplitSource: "manual"},                      // no measured split
		splitRow("6", 0, 1, 1),  // no estimate: excluded
		splitRow("7", -2, 1, 1), // hand-edited negative estimate: excluded
		splitRow("8", 2, 1, -1), // negative converge: excluded
	}
	c, ok := CalibrateForecast(rows)
	if !ok || c.N != 3 {
		t.Fatalf("calibration = %+v ok=%v, want 3 rows", c, ok)
	}
	// Σest 7, Σbuild 12, Σconv 11.5 → point 23.5/7, k 11.5/12
	if !near(c.Point, 23.5/7) || !near(c.K, 11.5/12) {
		t.Fatalf("point %v k %v", c.Point, c.K)
	}
	if !(c.Low <= c.Point && c.Point <= c.High) || c.Collapsed {
		t.Fatalf("range %v–%v around %v (collapsed %v)", c.Low, c.High, c.Point, c.Collapsed)
	}
	if !near(c.Prior, 2) {
		t.Fatalf("prior = %v, want median estimate 2", c.Prior)
	}
}

func TestCalibrateForecastCollapsesBelowMinRows(t *testing.T) {
	c, ok := CalibrateForecast([]LedgerRow{splitRow("1", 2, 1, 1), splitRow("2", 2, 3, 1)})
	if !ok || !c.Collapsed || c.Low != c.Point || c.High != c.Point {
		t.Fatalf("calibration = %+v", c)
	}
	if _, ok := CalibrateForecast(nil); ok {
		t.Fatal("empty ledger calibrated")
	}
}
