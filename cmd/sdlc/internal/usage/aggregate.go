package usage

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// NormalizeResponses is PURE. Identity conflicts poison the entire response,
// including evidence outside the requested window. No conflicting final wins
// merely because its copy happened to be read last.
func NormalizeResponses(in []Response) ([]Response, Coverage, error) {
	var c Coverage
	byID := map[string]Response{}
	bad := map[string]bool{}
	for _, r := range in {
		if e := r.ValidateMeasurements(); e != nil {
			return nil, c, e
		}
		if r.ID == "" || r.Agent == "" || r.At.IsZero() {
			c.Add("malformed_records", "missing normalized response identity or time")
			continue
		}
		key := r.Agent + ":" + r.ID
		if bad[key] {
			continue
		}
		old, ok := byID[key]
		if !ok {
			if len(byID) >= MaxResponses {
				c.Add("bound_exceeded", "global response identity limit")
				return nil, c, fmt.Errorf("usage response identity limit exceeded")
			}
			byID[key] = r
			continue
		}
		if old.SourceID != r.SourceID || old.Model != r.Model || old.RequestID != r.RequestID || old.Origin != r.Origin {
			bad[key] = true
			delete(byID, key)
			c.Add("conflicts", "conflicting response attribution")
			continue
		}
		if old.equalMeasurements(r) {
			c.Add("duplicates", "")
			if old.Provisional && !r.Provisional {
				byID[key] = r
				c.Add("revisions", "")
			} else if old.Provisional == r.Provisional && r.At.After(old.At) {
				old.At = r.At
				byID[key] = old
			}
			continue
		}
		switch {
		case old.Provisional && old.monotoneMeasurements(r):
			byID[key] = r
			c.Add("revisions", "")
		case r.Provisional && r.monotoneMeasurements(old):
			c.Add("revisions", "")
		default:
			bad[key] = true
			delete(byID, key)
			c.Add("conflicts", "conflicting response terminal counters or revision")
		}
	}
	var out []Response
	for _, r := range byID {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].At.Equal(out[j].At) {
			return out[i].Agent+":"+out[i].ID < out[j].Agent+":"+out[j].ID
		}
		return out[i].At.Before(out[j].At)
	})
	return out, c, nil
}

func measuredDimensions(dst *Counters, v Counters) error {
	for _, x := range []struct {
		dst **int64
		src *int64
	}{{&dst.CacheRead, v.CacheRead}, {&dst.CacheWrite, v.CacheWrite}, {&dst.Reasoning, v.Reasoning}} {
		if x.src != nil {
			old := int64(0)
			if *x.dst != nil {
				old = **x.dst
			}
			n, e := checkedAdd(old, *x.src)
			if e != nil {
				return e
			}
			*x.dst = &n
		}
	}
	return nil
}

// measurements retains known subtotals while an unknown component poisons only
// that unit's total. This keeps native units independent in mixed reports.
type measurements struct {
	tokens                      *Counters
	credits                     *CreditAmount
	measured                    *Counters
	measuredCredits             *CreditAmount
	tokenUnknown, creditUnknown bool
}

func (m *measurements) add(r Response) error {
	if r.Tokens == nil {
		m.tokenUnknown = true
	} else {
		v := *r.Tokens
		if m.measured == nil {
			m.measured = &Counters{}
		}
		if m.tokens == nil {
			m.tokens = &v
		} else {
			s, e := m.tokens.Add(v)
			if e != nil {
				return e
			}
			m.tokens = &s
		}
		var e error
		m.measured.Input, e = checkedAdd(m.measured.Input, v.Input)
		if e != nil {
			return e
		}
		m.measured.Output, e = checkedAdd(m.measured.Output, v.Output)
		if e != nil {
			return e
		}
		if _, e = m.measured.Total(); e != nil {
			return e
		}
		if e = measuredDimensions(m.measured, v); e != nil {
			return e
		}
	}
	if r.Credits == nil {
		m.creditUnknown = true
	} else {
		v := *r.Credits
		if m.measuredCredits == nil {
			m.measuredCredits = &v
		} else {
			s, e := m.measuredCredits.Add(v)
			if e != nil {
				return e
			}
			m.measuredCredits = &s
		}
		m.credits = m.measuredCredits
	}
	return nil
}
func (m measurements) totals() (*Counters, *CreditAmount) {
	t, c := m.tokens, m.credits
	if m.tokenUnknown {
		t = nil
	}
	if m.creditUnknown {
		c = nil
	}
	return t, c
}
func (m measurements) rates(hours float64) (*float64, *float64) {
	t, c := m.totals()
	var tr, cr *float64
	if t != nil {
		n, _ := t.Total()
		v := float64(n) / hours
		tr = &v
	}
	if c != nil {
		v := c.Float64() / hours
		cr = &v
	}
	return tr, cr
}

type groupAccumulator struct {
	Group
	detail, burn measurements
}
type hourAccumulator struct {
	Hour
	detail measurements
}

// AggregateUsage normalizes globally before filtering and attributes each unique
// response once. Rates use elapsed wall hours independently for each native unit.
func AggregateUsage(in []Response, window, burnWindow Window, resolve func(Response) Attribution) (Report, error) {
	var report Report
	if window.Until.Before(window.Since) || !burnWindow.Until.After(burnWindow.Since) {
		return report, fmt.Errorf("detail window cannot be reversed and burn window must have positive duration")
	}
	rr, c, e := NormalizeResponses(in)
	report.Coverage = c
	if e != nil {
		return report, e
	}
	groups := map[string]*groupAccumulator{}
	hours := map[time.Time]*hourAccumulator{}
	var detailTotal, burnTotal measurements
	for _, r := range rr {
		detail, burn := window.contains(r.At), burnWindow.contains(r.At)
		if !detail && !burn {
			continue
		}
		a := Attribution{Include: true}
		if resolve != nil {
			a = resolve(r)
		}
		if a.Conflict {
			a = Attribution{Include: a.Include, Slot: UnknownSlot, Role: "unknown"}
			report.Coverage.Add("attribution_conflicts", "conflicting role binding")
		}
		if !a.Include {
			continue
		}
		if a.Slot == "" || a.Slot == UnknownSlot {
			a.Slot = UnknownSlot
			report.Coverage.Add("unknown_slot", "")
		}
		if a.Role == "" {
			a.Role = "unknown"
		}
		model := r.Model
		if model == "" {
			model = "unknown"
			report.Coverage.Add("unknown_model", "")
		}
		if r.Provisional {
			report.Coverage.Add("provisional", "")
		}
		report.Coverage.Add("accepted_responses", "")
		key := strings.Join([]string{a.Slot, r.Agent, model, a.Role, a.Round, a.Attempt}, "\x00")
		g := groups[key]
		if g == nil {
			g = &groupAccumulator{Group: Group{Slot: a.Slot, Agent: r.Agent, Model: model, Role: a.Role, Round: a.Round, Attempt: a.Attempt}}
			groups[key] = g
		}
		if burn {
			if e = g.burn.add(r); e != nil {
				return report, e
			}
			if e = burnTotal.add(r); e != nil {
				return report, e
			}
		}
		if !detail {
			continue
		}
		if e = detailTotal.add(r); e != nil {
			return report, e
		}
		if e = g.detail.add(r); e != nil {
			return report, e
		}
		g.Responses++
		at := r.At.UTC().Truncate(time.Hour)
		h := hours[at]
		if h == nil {
			start, end := at, at.Add(time.Hour)
			if start.Before(window.Since) {
				start = window.Since
			}
			if end.After(window.Until) {
				end = window.Until
			}
			h = &hourAccumulator{Hour: Hour{At: at, DurationHours: end.Sub(start).Hours()}}
			hours[at] = h
		}
		if e = h.detail.add(r); e != nil {
			return report, e
		}
		h.Responses++
	}
	report.Totals, report.Credits = detailTotal.totals()
	report.MeasuredDimensions = detailTotal.measured
	report.MeasuredCredits = detailTotal.measuredCredits
	duration := burnWindow.Until.Sub(burnWindow.Since).Hours()
	report.BurnTokensPerHour, report.BurnCreditsPerHour = burnTotal.rates(duration)
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		g := groups[k]
		g.Tokens, g.Credits = g.detail.totals()
		g.MeasuredCredits = g.detail.measuredCredits
		g.MeasuredDimensions = g.detail.measured
		g.BurnTokensPerHour, g.BurnCreditsPerHour = g.burn.rates(duration)
		report.Groups = append(report.Groups, g.Group)
	}
	for _, h := range hours {
		h.Tokens, h.Credits = h.detail.totals()
		h.MeasuredCredits = h.detail.measuredCredits
		h.MeasuredDimensions = h.detail.measured
		report.Hourly = append(report.Hourly, h.Hour)
	}
	sort.Slice(report.Hourly, func(i, j int) bool { return report.Hourly[i].At.Before(report.Hourly[j].At) })
	return report, nil
}
