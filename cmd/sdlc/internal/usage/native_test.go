package usage

import (
	"encoding/json"
	"fmt"
	"testing"
)

func nativeCredit(t *testing.T, s string) *CreditAmount {
	t.Helper()
	c, e := ParseCredit([]byte(s))
	if e != nil {
		t.Fatal(e)
	}
	return &c
}
func TestNativeMixedMeasurements(t *testing.T) {
	w := Window{stamp("2026-10-14T09:00:00Z"), stamp("2026-10-14T10:00:00Z")}
	rows := []Response{{ID: "q", Agent: "qoder", At: w.Since, Credits: nativeCredit(t, "0.123456789012345678")}, {ID: "c", Agent: "claude", At: w.Since, Tokens: &Counters{Input: 10, Output: 2}}}
	r, e := AggregateUsage(rows, w, w, nil)
	if e != nil {
		t.Fatal(e)
	}
	if r.Totals != nil || r.Credits != nil || r.BurnTokensPerHour != nil || r.BurnCreditsPerHour != nil {
		t.Fatalf("mixed totals fabricated: %+v", r)
	}
	if r.MeasuredDimensions == nil || r.MeasuredDimensions.Input != 10 || r.MeasuredCredits == nil || r.MeasuredCredits.String() != "0.123456789012345678" {
		t.Fatalf("lost measured subtotal: %+v", r)
	}
	if len(r.Hourly) != 1 || r.Hourly[0].MeasuredDimensions == nil || r.Hourly[0].MeasuredDimensions.Input != 10 {
		t.Fatalf("hourly subtotal lost: %+v", r.Hourly)
	}
	q, e := AggregateUsage(rows[:1], w, w, nil)
	if e != nil || q.Totals != nil || q.Credits == nil || q.Credits.String() != "0.123456789012345678" || q.BurnCreditsPerHour == nil {
		t.Fatalf("native credits: %+v %v", q, e)
	}
}
func TestNativeMeasurementRevisionsAndConflict(t *testing.T) {
	r := Response{ID: "q", Agent: "qoder", At: stamp("2026-10-14T09:00:00Z"), Credits: nativeCredit(t, "1"), Provisional: true}
	f := r
	f.Credits = nativeCredit(t, "2")
	f.Provisional = false
	for _, rows := range [][]Response{{r, f}, {f, r}} {
		got, c, e := NormalizeResponses(rows)
		if e != nil || len(got) != 1 || got[0].Credits.String() != "2" || c.Counts["conflicts"] != 0 {
			t.Fatalf("revision %+v %+v %v", got, c, e)
		}
	}
	bad := f
	bad.Credits = nativeCredit(t, "3")
	got, c, e := NormalizeResponses([]Response{f, bad})
	if e != nil || len(got) != 0 || c.Counts["conflicts"] != 1 {
		t.Fatalf("conflict %+v %+v %v", got, c, e)
	}
}
func TestCreditProjectionRequiresTwoReadings(t *testing.T) {
	now := stamp("2026-10-14T10:00:00Z")
	m := Meter{Version: 1, Observations: []Observation{{Agent: "qoder", Unit: "credits", At: now.Add(-3600000000000), Remaining: 100}}}
	c := nativeCredit(t, "5")
	p := ProjectCreditMeter(m, "qoder", "credits", now, *c, 2, true)
	if p.Status != "insufficient_calibration" {
		t.Fatal(p)
	}
	m.Observations = append(m.Observations, Observation{Agent: "qoder", Unit: "credits", At: now, Remaining: 90})
	p = ProjectCreditMeter(m, "qoder", "credits", now, *c, 2, true)
	if p.Status != "calibrated" || p.UnitsPerToken != nil || p.UnitsPerCredit == nil || *p.UnitsPerCredit != 2 || *p.HoursRemaining != 22.5 {
		t.Fatal(p)
	}
}

func TestNativeKnownZeroAndMeasurementValidation(t *testing.T) {
	w := Window{stamp("2026-10-14T09:00:00Z"), stamp("2026-10-14T10:00:00Z")}
	r := Response{ID: "zero", Agent: "qoder", At: w.Since, Credits: nativeCredit(t, "0")}
	if e := r.ValidateMeasurements(); e != nil {
		t.Fatal(e)
	}
	report, e := AggregateUsage([]Response{r}, w, w, nil)
	if e != nil || report.Totals != nil || report.Credits == nil || !report.Credits.IsZero() || report.BurnCreditsPerHour == nil || *report.BurnCreditsPerHour != 0 {
		t.Fatalf("%+v %v", report, e)
	}
	r.Credits = nil
	if e := r.ValidateMeasurements(); e == nil {
		t.Fatal("missing final measurement accepted")
	}
	r.Tokens = &Counters{Input: -1}
	if e := r.ValidateMeasurements(); e == nil {
		t.Fatal("negative tokens accepted")
	}
}
func FuzzNativeCreditPermutation(f *testing.F) {
	f.Add(uint16(123), uint16(456))
	f.Fuzz(func(t *testing.T, a, b uint16) {
		w := Window{stamp("2026-10-14T09:00:00Z"), stamp("2026-10-14T10:00:00Z")}
		x := Response{ID: "a", Agent: "qoder", At: w.Since, Credits: nativeCredit(t, fmt.Sprint(a))}
		y := x
		y.ID = "b"
		y.Credits = nativeCredit(t, fmt.Sprint(b))
		for _, rows := range [][]Response{{x, y, x}, {y, x, y}, {x, x, y, y}} {
			r, e := AggregateUsage(rows, w, w, nil)
			if e != nil || r.Credits == nil || r.Credits.String() != fmt.Sprint(uint64(a)+uint64(b)) || r.Totals != nil || r.BurnTokensPerHour != nil {
				t.Fatalf("%+v %v", r, e)
			}
		}
	})
}

func TestAggregateNativeTokenSumOverflow(t *testing.T) {
	w := Window{stamp("2026-10-14T09:00:00Z"), stamp("2026-10-14T10:00:00Z")}
	rows := []Response{{ID: "a", Agent: "codex", At: w.Since, Tokens: &Counters{Input: 9223372036854775807}}, {ID: "b", Agent: "claude", At: w.Since, Tokens: &Counters{Output: 1}}}
	if _, e := AggregateUsage(rows, w, w, nil); e == nil {
		t.Fatal("combined input/output overflow accepted")
	}
}

func TestNativeCreditOnlyTokenSubtotalsJSONNull(t *testing.T) {
	w := Window{stamp("2026-10-14T09:00:00Z"), stamp("2026-10-14T10:00:00Z")}
	row := Response{ID: "q", Agent: "qoder", At: w.Since, Credits: nativeCredit(t, "0")}
	for _, rows := range [][]Response{{row}, nil} {
		report, e := AggregateUsage(rows, w, w, nil)
		if e != nil {
			t.Fatal(e)
		}
		raw, e := json.Marshal(report)
		if e != nil {
			t.Fatal(e)
		}
		var value map[string]json.RawMessage
		if e = json.Unmarshal(raw, &value); e != nil {
			t.Fatal(e)
		}
		if string(value["measured_dimensions"]) != "null" {
			t.Fatalf("credit-only token subtotal fabricated: %s", raw)
		}
		for _, key := range []string{"groups", "hourly"} {
			var buckets []map[string]json.RawMessage
			if e = json.Unmarshal(value[key], &buckets); e != nil {
				t.Fatal(e)
			}
			for _, bucket := range buckets {
				if string(bucket["measured_dimensions"]) != "null" {
					t.Fatalf("%s fabricated token subtotal: %s", key, raw)
				}
			}
		}
	}
}
func TestNativeMeasuredTokenZeroSubtotalRemainsKnown(t *testing.T) {
	w := Window{stamp("2026-10-14T09:00:00Z"), stamp("2026-10-14T10:00:00Z")}
	report, e := AggregateUsage([]Response{{ID: "c", Agent: "claude", At: w.Since, Tokens: &Counters{}}}, w, w, nil)
	if e != nil {
		t.Fatal(e)
	}
	raw, e := json.Marshal(report)
	if e != nil {
		t.Fatal(e)
	}
	var value map[string]json.RawMessage
	if e = json.Unmarshal(raw, &value); e != nil {
		t.Fatal(e)
	}
	if string(value["measured_dimensions"]) == "null" {
		t.Fatalf("reported zero erased: %s", raw)
	}
}
