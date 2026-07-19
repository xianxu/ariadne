package project

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// forecast.go — the pure calendar forecast core (#182). Divides a project's
// remaining issue-hours by its measured share of throughput to project a
// finish date. No IO: value inputs, deterministic output. The one place
// calendar math lives — a future roadmap rollup is "call this per project and
// sum," with no changes here. INFORMS, never blocks: the caller records the
// statement, it never refuses on the answer.

const forecastISO = "2006-01-02"

// ProjectLoad is one project's contention weight: its status and remaining
// hours. RemainingSource records where the hours came from (board = summed
// unfinished issue estimates; phase-a = the PRD-level fallback; unknown =
// neither resolved, so the project weighs 0 and carries a Warning).
type ProjectLoad struct {
	Name            string
	Repo            string
	Status          string
	RemainingHours  float64
	RemainingSource string // board | phase-a | unknown
	Warning         string
}

// isActiveContention reports whether a load competes for throughput NOW: a
// committed or executing project with a measurable (non-unknown) load.
func (l ProjectLoad) isActiveContention() bool {
	if l.RemainingSource == "unknown" {
		return false
	}
	return l.Status == "committed" || l.Status == "executing"
}

// Forecast is the computed projection plus its full arithmetic trail, so every
// surface can render an auditable one-line statement.
type Forecast struct {
	ProjectedFinish string  // ISO date
	N               int     // active projects sharing throughput (incl. this one)
	SharePerWeek    float64 // baseline h/wk ÷ N
	Remaining       float64
	RemainingSource string
	HoursPerWeek    float64
	CeilingWarning  string
	PausedRisks     []ProjectLoad
	Notes           []string
}

// ComputeForecast projects this project's finish date. n = this project + every
// OTHER committed/executing project with a measurable load; share = baseline
// h/wk ÷ n; finish = today + ceil(remaining/share × 7) days. The attention
// ceiling is a WARNING threshold only (n > ceiling → note), never arithmetic —
// parallelism is already priced into the measured throughput. Paused others
// weigh 0 and become named risk lines. Errors when there's nothing to divide
// (zero remaining, zero/negative baseline) so the caller falls back rather than
// printing a bogus date.
func ComputeForecast(b estimate.ThroughputBaseline, this ProjectLoad, others []ProjectLoad, today string) (Forecast, error) {
	if b.HoursPerWeek <= 0 {
		return Forecast{}, fmt.Errorf("throughput baseline is %.2f h/wk — bless a baseline before forecasting", b.HoursPerWeek)
	}
	if this.RemainingHours <= 0 {
		return Forecast{}, fmt.Errorf("project %q has no remaining hours to forecast", this.Name)
	}
	todayT, err := time.Parse(forecastISO, today)
	if err != nil {
		return Forecast{}, fmt.Errorf("bad today %q: %w", today, err)
	}

	f := Forecast{
		N:               1,
		Remaining:       this.RemainingHours,
		RemainingSource: this.RemainingSource,
		HoursPerWeek:    b.HoursPerWeek,
	}
	for _, o := range others {
		switch {
		case o.Status == "paused":
			f.PausedRisks = append(f.PausedRisks, o)
		case o.isActiveContention():
			f.N++
		case o.RemainingSource == "unknown" && (o.Status == "committed" || o.Status == "executing"):
			f.Notes = append(f.Notes, fmt.Sprintf("%s has no measurable load (weighs 0)", o.Name))
		}
	}

	f.SharePerWeek = b.HoursPerWeek / float64(f.N)
	weeks := this.RemainingHours / f.SharePerWeek
	days := int(math.Ceil(weeks * 7.0))
	f.ProjectedFinish = todayT.AddDate(0, 0, days).Format(forecastISO)

	if f.N > b.Ceiling {
		f.CeilingWarning = fmt.Sprintf("%d active projects exceed your attention ceiling of %d — forecast degrades", f.N, b.Ceiling)
	}
	if this.RemainingSource == "phase-a" {
		f.Notes = append(f.Notes, "remaining is the Phase-A estimate (breakdown not yet resolved)")
	}
	return f, nil
}

// RenderForecast is the identical one-paragraph statement every consumer prints.
// deadline (ISO, may be empty) drives the over/under clause.
func RenderForecast(f Forecast, deadline string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "forecast: %.1fh remaining (%s) ÷ %.1fh/wk ÷ %d active → ~%.1fh/wk share → lands ~%s",
		f.Remaining, f.RemainingSource, f.HoursPerWeek, f.N, f.SharePerWeek, f.ProjectedFinish)

	if deadline == "" {
		b.WriteString(" (no deadline set)")
	} else if delta, ok := dayDelta(deadline, f.ProjectedFinish); ok {
		switch {
		case delta > 0:
			fmt.Fprintf(&b, " (deadline %s: %d days over)", deadline, delta)
		case delta < 0:
			fmt.Fprintf(&b, " (deadline %s: %d days under)", deadline, -delta)
		default:
			fmt.Fprintf(&b, " (deadline %s: on time)", deadline)
		}
	} else {
		fmt.Fprintf(&b, " (deadline %s: unparsable)", deadline)
	}

	for _, p := range f.PausedRisks {
		fmt.Fprintf(&b, ". paused: %s (%.0fh) — resuming invalidates this", p.Name, p.RemainingHours)
	}
	if f.CeilingWarning != "" {
		fmt.Fprintf(&b, ". [%s]", f.CeilingWarning)
	}
	for _, n := range f.Notes {
		fmt.Fprintf(&b, ". %s", n)
	}
	return b.String()
}

// dayDelta returns projected − deadline in whole days (positive = over/late).
func dayDelta(deadline, projected string) (int, bool) {
	d, err1 := time.Parse(forecastISO, deadline)
	p, err2 := time.Parse(forecastISO, projected)
	if err1 != nil || err2 != nil {
		return 0, false
	}
	return int(p.Sub(d).Hours() / 24), true
}
