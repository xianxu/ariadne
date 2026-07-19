package project

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

func baseline(hpw float64, ceiling int) estimate.ThroughputBaseline {
	return estimate.ThroughputBaseline{HoursPerWeek: hpw, Ceiling: ceiling, SpanStart: "2026-06-01", SpanEnd: "2026-06-28"}
}

func TestComputeForecast_Solo(t *testing.T) {
	// n=1, full share. 55h remaining ÷ 55h/wk = 1.0 wk = 7 days.
	f, err := ComputeForecast(baseline(55, 2),
		ProjectLoad{Name: "p", Status: "executing", RemainingHours: 55, RemainingSource: "board"},
		nil, "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	if f.N != 1 {
		t.Errorf("N = %d, want 1", f.N)
	}
	if f.SharePerWeek != 55 {
		t.Errorf("SharePerWeek = %.2f, want 55", f.SharePerWeek)
	}
	if f.ProjectedFinish != "2026-09-08" {
		t.Errorf("ProjectedFinish = %q, want 2026-09-08 (today+7)", f.ProjectedFinish)
	}
	if f.CeilingWarning != "" {
		t.Errorf("no ceiling warning expected at n=1: %q", f.CeilingWarning)
	}
}

func TestComputeForecast_TwoActiveOthers(t *testing.T) {
	// n=3 (this + 2 active). 36h ÷ (55/3=18.333h/wk) = 1.9636 wk → 13.745 days → ceil 14.
	others := []ProjectLoad{
		{Name: "a", Status: "executing", RemainingHours: 10, RemainingSource: "board"},
		{Name: "b", Status: "committed", RemainingHours: 20, RemainingSource: "board"},
	}
	f, err := ComputeForecast(baseline(55, 2),
		ProjectLoad{Name: "p", Status: "executing", RemainingHours: 36, RemainingSource: "board"},
		others, "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	if f.N != 3 {
		t.Errorf("N = %d, want 3", f.N)
	}
	share := 55.0 / 3.0
	if f.SharePerWeek < share-0.001 || f.SharePerWeek > share+0.001 {
		t.Errorf("SharePerWeek = %.4f, want %.4f", f.SharePerWeek, share)
	}
	if f.ProjectedFinish != "2026-09-15" {
		t.Errorf("ProjectedFinish = %q, want 2026-09-15 (today+14)", f.ProjectedFinish)
	}
	// n=3 > ceiling 2 → warning.
	if f.CeilingWarning == "" {
		t.Error("expected a ceiling warning at n=3 > ceiling 2")
	}
}

func TestComputeForecast_PausedExcludedButListed(t *testing.T) {
	others := []ProjectLoad{
		{Name: "metis-v2", Status: "paused", RemainingHours: 14, RemainingSource: "board"},
	}
	f, err := ComputeForecast(baseline(55, 2),
		ProjectLoad{Name: "p", Status: "executing", RemainingHours: 55, RemainingSource: "board"},
		others, "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	if f.N != 1 {
		t.Errorf("N = %d, want 1 (paused excluded from contention)", f.N)
	}
	if len(f.PausedRisks) != 1 || f.PausedRisks[0].Name != "metis-v2" {
		t.Errorf("paused project should be a named risk: %+v", f.PausedRisks)
	}
}

func TestComputeForecast_UnknownOtherWeightZero(t *testing.T) {
	// An 'unknown'-source other has no measurable load → excluded from n, noted.
	others := []ProjectLoad{
		{Name: "mystery", Status: "executing", RemainingSource: "unknown", Warning: "no board rows, no phase-a"},
	}
	f, err := ComputeForecast(baseline(40, 2),
		ProjectLoad{Name: "p", Status: "executing", RemainingHours: 40, RemainingSource: "board"},
		others, "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	if f.N != 1 {
		t.Errorf("N = %d, want 1 (unknown-load other excluded)", f.N)
	}
	joined := strings.Join(f.Notes, " ")
	if !strings.Contains(joined, "mystery") {
		t.Errorf("unknown-load other should be noted: %v", f.Notes)
	}
}

func TestComputeForecast_ZeroRemainingOtherDoesNotContend(t *testing.T) {
	// A committed/executing OTHER with 0 remaining (burned down, not yet closed)
	// consumes no throughput → must not bump N.
	others := []ProjectLoad{
		{Name: "done-soon", Status: "committed", RemainingHours: 0, RemainingSource: "board"},
	}
	f, err := ComputeForecast(baseline(40, 2),
		ProjectLoad{Name: "p", Status: "executing", RemainingHours: 40, RemainingSource: "board"},
		others, "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	if f.N != 1 {
		t.Errorf("a zero-remaining other must not contend: N = %d, want 1", f.N)
	}
}

func TestComputeForecast_ZeroRemainingErrors(t *testing.T) {
	_, err := ComputeForecast(baseline(55, 2),
		ProjectLoad{Name: "p", Status: "executing", RemainingHours: 0, RemainingSource: "board"},
		nil, "2026-09-01")
	if err == nil {
		t.Fatal("zero remaining should error (nothing to forecast)")
	}
}

func TestComputeForecast_ZeroBaselineErrors(t *testing.T) {
	_, err := ComputeForecast(baseline(0, 2),
		ProjectLoad{Name: "p", Status: "executing", RemainingHours: 40, RemainingSource: "board"},
		nil, "2026-09-01")
	if err == nil {
		t.Fatal("zero baseline h/wk should error (can't divide)")
	}
}

func TestRenderForecast_FullStatement(t *testing.T) {
	others := []ProjectLoad{
		{Name: "metis-v2", Status: "paused", RemainingHours: 14, RemainingSource: "board"},
		{Name: "a", Status: "executing", RemainingHours: 10, RemainingSource: "board"},
	}
	f, err := ComputeForecast(baseline(55, 2),
		ProjectLoad{Name: "p", Status: "executing", RemainingHours: 36, RemainingSource: "board"},
		others, "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	s := RenderForecast(f, "2026-09-01")
	for _, want := range []string{"36", "board", "lands", f.ProjectedFinish, "paused", "metis-v2"} {
		if !strings.Contains(s, want) {
			t.Errorf("statement missing %q:\n%s", want, s)
		}
	}
	// Deadline delta in days (projected is after 09-01 → "over").
	if !strings.Contains(s, "over") {
		t.Errorf("statement should show the deadline delta:\n%s", s)
	}
}

func TestRenderForecast_NoDeadline(t *testing.T) {
	f, err := ComputeForecast(baseline(55, 2),
		ProjectLoad{Name: "p", Status: "executing", RemainingHours: 55, RemainingSource: "board"},
		nil, "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	s := RenderForecast(f, "")
	if !strings.Contains(s, "no deadline") {
		t.Errorf("absent deadline should render a 'no deadline' clause:\n%s", s)
	}
}

func TestRenderForecast_PhaseAFallbackNoted(t *testing.T) {
	f, err := ComputeForecast(baseline(55, 2),
		ProjectLoad{Name: "p", Status: "committed", RemainingHours: 40, RemainingSource: "phase-a"},
		nil, "2026-09-01")
	if err != nil {
		t.Fatal(err)
	}
	s := RenderForecast(f, "2026-10-01")
	if !strings.Contains(s, "phase-a") {
		t.Errorf("phase-a source should be surfaced in the statement:\n%s", s)
	}
}
