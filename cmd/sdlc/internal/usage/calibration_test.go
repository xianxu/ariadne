package usage

import (
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
