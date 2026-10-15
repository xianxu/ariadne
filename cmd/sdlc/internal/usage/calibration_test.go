package usage

import (
	"encoding/json"
	"math"
	"testing"
	"time"
)

func TestMeterTransitionsAndProjection(t *testing.T) {
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	first := Observation{Agent: "codex", Unit: "credits", Remaining: 100, At: now.Add(-time.Hour)}
	m, err := ObserveMeter(Meter{}, first)
	if err != nil {
		t.Fatal(err)
	}
	if p := ProjectMeter(m, "codex", "credits", now, 1000, 200, true); p.Status != "insufficient_calibration" || p.HoursRemaining != nil {
		t.Fatalf("one point: %+v", p)
	}
	second := first
	second.At = now
	second.Remaining = 90
	m, err = ObserveMeter(m, second)
	if err != nil {
		t.Fatal(err)
	}
	p := ProjectMeter(m, "codex", "credits", now, 1000, 200, true)
	if p.HoursRemaining == nil || *p.HoursRemaining != 45 {
		t.Fatalf("projection %+v", p)
	}
	same, err := ObserveMeter(m, second)
	if err != nil || len(same.Observations) != 2 {
		t.Fatalf("duplicate %+v %v", same, err)
	}
	conflict := second
	conflict.Remaining = 89
	if _, err = ObserveMeter(m, conflict); err == nil {
		t.Fatal("same-time conflict accepted")
	}
	stale := first
	stale.At = now.Add(-2 * time.Hour)
	if _, err = ObserveMeter(m, stale); err == nil {
		t.Fatal("stale observation accepted")
	}
	reset := second
	reset.At = now.Add(time.Hour)
	reset.Remaining = 110
	m, err = ObserveMeter(m, reset)
	if err != nil || len(m.Observations) != 1 {
		t.Fatalf("reset %+v %v", m, err)
	}
}
func TestMeterRefusesUnsupportedEvidence(t *testing.T) {
	now := time.Now().UTC()
	m := Meter{Version: 1, Observations: []Observation{{Agent: "claude", Unit: "percent", Remaining: 80, At: now.Add(-time.Hour)}, {Agent: "claude", Unit: "percent", Remaining: 70, At: now}}}
	for _, tc := range []struct {
		name     string
		tokens   int64
		burn     float64
		complete bool
		at       time.Time
		status   string
	}{{"gap", 100, 10, false, now, "incomplete_evidence"}, {"zero", 0, 10, true, now, "no_measured_tokens"}, {"idle", 100, 0, true, now, "no_recent_burn"}, {"stale", 100, 10, true, now.Add(8 * 24 * time.Hour), "stale_calibration"}} {
		t.Run(tc.name, func(t *testing.T) {
			p := ProjectMeter(m, "claude", "percent", tc.at, tc.tokens, tc.burn, tc.complete)
			if p.Status != tc.status || p.HoursRemaining != nil {
				t.Fatalf("%+v", p)
			}
		})
	}
	for _, o := range []Observation{{Agent: "gemini", Unit: "credits", At: now}, {Agent: "codex", Unit: "dollars", At: now}, {Agent: "claude", Unit: "percent", Remaining: 101, At: now}} {
		if _, err := ObserveMeter(Meter{}, o); err == nil {
			t.Fatalf("invalid accepted %+v", o)
		}
	}
}

func TestMeterExtremeFiniteInputsNeverPublishNonFiniteProjection(t *testing.T) {
	now := stamp("2026-10-14T10:00:00Z")
	for _, agent := range []string{"codex", "qoder"} {
		m := Meter{Version: 1, Observations: []Observation{{Agent: agent, Unit: "credits", Remaining: math.MaxFloat64, At: now.Add(-time.Hour)}, {Agent: agent, Unit: "credits", Remaining: 1, At: now}}}
		var projections []Projection
		if agent == "codex" {
			projections = append(projections, ProjectMeter(m, agent, "credits", now, 1, 2, true))
		} else {
			for _, s := range []string{"1", "0.000000000000000000000000000000000001"} {
				c := nativeCredit(t, s)
				projections = append(projections, ProjectCreditMeter(m, agent, "credits", now, *c, 2, true))
			}
		}
		for _, p := range projections {
			if p.Status != "invalid_projection" || p.UnitsPerToken != nil || p.UnitsPerCredit != nil || p.UnitsPerHour != nil || p.HoursRemaining != nil {
				t.Fatalf("nonfinite projection accepted: %+v", p)
			}
			if _, e := json.Marshal(p); e != nil {
				t.Fatal(e)
			}
		}
	}
}

func FuzzMeterProjectionFiniteJSON(f *testing.F) {
	f.Add(math.MaxFloat64, 2.0, "0.000000000000000000000000000000000001")
	f.Add(100.0, 10.0, "1")
	f.Fuzz(func(t *testing.T, remaining, burn float64, raw string) {
		if math.IsNaN(remaining) || math.IsInf(remaining, 0) || remaining <= 1 || math.IsNaN(burn) || math.IsInf(burn, 0) || burn <= 0 {
			return
		}
		credits, e := ParseCredit([]byte(raw))
		if e != nil {
			return
		}
		now := stamp("2026-10-14T10:00:00Z")
		for _, agent := range []string{"codex", "qoder"} {
			m := Meter{Version: 1, Observations: []Observation{{Agent: agent, Unit: "credits", Remaining: remaining, At: now.Add(-time.Hour)}, {Agent: agent, Unit: "credits", Remaining: 1, At: now}}}
			var p Projection
			if agent == "qoder" {
				p = ProjectCreditMeter(m, agent, "credits", now, credits, burn, true)
			} else {
				p = ProjectMeter(m, agent, "credits", now, 1, burn, true)
			}
			if _, e = json.Marshal(p); e != nil {
				t.Fatalf("projection not finite: %+v %v", p, e)
			}
		}
	})
}
