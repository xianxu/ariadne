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
		if e := r.Tokens.Validate(); e != nil {
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
		if old.Tokens.equal(r.Tokens) {
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
		case old.Provisional && old.Tokens.monotone(r.Tokens):
			byID[key] = r
			c.Add("revisions", "")
		case r.Provisional && r.Tokens.monotone(old.Tokens):
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

type groupAccumulator struct {
	Group
	burn int64
	has  bool
}
type hourAccumulator struct {
	Hour
	has bool
}

func appendCounters(dst *Counters, has *bool, value Counters) error {
	if !*has {
		*dst = value
		*has = true
		return nil
	}
	r, e := dst.Add(value)
	*dst = r
	return e
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

// AggregateUsage is PURE. Attribution runs once, after global deduplication;
// its Include predicate owns issue scope and exclusive reviewer precedence.
// Rates use elapsed WALL hours in burnWindow, independent of active-time.
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
	hasTotal := false
	var burnTotal int64
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
		total, e := r.Tokens.Total()
		if e != nil {
			return report, e
		}
		if burn {
			if g.burn, e = checkedAdd(g.burn, total); e != nil {
				return report, e
			}
			if burnTotal, e = checkedAdd(burnTotal, total); e != nil {
				return report, e
			}
		}
		if !detail {
			continue
		}
		if e = appendCounters(&report.Totals, &hasTotal, r.Tokens); e != nil {
			return report, e
		}
		if e = measuredDimensions(&report.MeasuredDimensions, r.Tokens); e != nil {
			return report, e
		}
		if e = appendCounters(&g.Tokens, &g.has, r.Tokens); e != nil {
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
		if e = appendCounters(&h.Tokens, &h.has, r.Tokens); e != nil {
			return report, e
		}
		h.Responses++
	}
	duration := burnWindow.Until.Sub(burnWindow.Since).Hours()
	report.BurnTokensPerHour = float64(burnTotal) / duration
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		g := groups[k]
		g.BurnTokensPerHour = float64(g.burn) / duration
		report.Groups = append(report.Groups, g.Group)
	}
	for _, h := range hours {
		report.Hourly = append(report.Hourly, h.Hour)
	}
	sort.Slice(report.Hourly, func(i, j int) bool { return report.Hourly[i].At.Before(report.Hourly[j].At) })
	report.MeasuredDimensions.Input, report.MeasuredDimensions.Output = report.Totals.Input, report.Totals.Output
	if _, e = report.Totals.Total(); e != nil {
		return report, e
	}
	return report, nil
}
