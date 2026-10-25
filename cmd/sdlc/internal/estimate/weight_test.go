package estimate

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func randomCalibrationRow(r *rand.Rand, i int) LedgerRow {
	build := 0.2 + r.Float64()*4
	row := LedgerRow{Issue: fmt.Sprintf("ariadne#%d", 1000+i), Estimate: 0.5 + r.Float64()*5, Actual: build, Model: "estimate-logic-v3.1",
		WindowTrusted: true, FlowKind: "full", SplitSource: "dispatch", BuildHours: build, ConvergeHours: r.Float64() * 2, ReviewRecipe: "milestone-review"}
	row.Actual += row.ConvergeHours
	return row
}

// #603 property: a row of weight 0 changes no calibration output, and a ledger
// of exact rows calibrates exactly as rows without a confidence column did.
func TestCalibrationWeightProperty(t *testing.T) {
	r := rand.New(rand.NewSource(603))
	for trial := 0; trial < 200; trial++ {
		var rows []LedgerRow
		for i := 0; i < 1+r.Intn(12); i++ {
			rows = append(rows, randomCalibrationRow(r, i))
		}
		exact := append([]LedgerRow(nil), rows...)
		for i := range exact {
			exact[i].Confidence, exact[i].AttributedShare = "exact", 1
		}
		f1, ok1 := CalibrateForecast(rows)
		if f2, ok2 := CalibrateForecast(exact); ok1 != ok2 || f1 != f2 || !reflect.DeepEqual(ConvergeFactors(rows), ConvergeFactors(exact)) {
			t.Fatalf("trial %d: exact rows calibrate differently: %+v vs %+v", trial, f1, f2)
		}
		zero := randomCalibrationRow(r, 999)
		zero.Confidence, zero.AttributedShare = "partial", 0
		unmeasured := randomCalibrationRow(r, 998)
		unmeasured.Confidence = "unmeasured"
		with := append(append([]LedgerRow(nil), rows...), zero, unmeasured)
		f2, ok2 := CalibrateForecast(with)
		if ok1 != ok2 || f1 != f2 || !reflect.DeepEqual(ConvergeFactors(rows), ConvergeFactors(with)) {
			t.Fatalf("trial %d: a weight-0 row changed calibration: %+v vs %+v", trial, f1, f2)
		}
		w1, m1 := DriftVerdict(rows, 3)
		w2, m2 := DriftVerdict(with, 3)
		if w1 != w2 || m1 != m2 {
			t.Fatalf("trial %d: a weight-0 row changed drift", trial)
		}
	}
}

// A partial row counts by its share in the sums and stays out of the range.
func TestCalibrationPartialRowIsWeighted(t *testing.T) {
	full := LedgerRow{Issue: "ariadne#1", Estimate: 2, Actual: 2, WindowTrusted: true, SplitSource: "dispatch", BuildHours: 2, FlowKind: "full"}
	half := LedgerRow{Issue: "ariadne#2", Estimate: 2, Actual: 4, WindowTrusted: true, SplitSource: "dispatch", BuildHours: 4, FlowKind: "full", Confidence: "partial", AttributedShare: 0.25}
	c, ok := CalibrateForecast([]LedgerRow{full, half})
	// Σw·build / Σw·est = (2 + 0.25·4) / (2 + 0.25·2) = 3 / 2.5
	if !ok || c.N != 2 || c.Partial != 1 || c.Point != 3/2.5 || !c.Collapsed {
		t.Fatalf("weighted forecast = %+v", c)
	}
	f := ConvergeFactors([]LedgerRow{full, half})
	if len(f) != 1 || f[0].N != 2 || f[0].Partial != 1 || f[0].Build != 3 {
		t.Fatalf("weighted converge = %+v", f)
	}
}
