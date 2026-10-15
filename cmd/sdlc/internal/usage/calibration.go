package usage

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Observations are account meter readings, not inferred token quotas.
type Observation struct {
	Agent     string    `json:"agent"`
	Unit      string    `json:"unit"`
	Remaining float64   `json:"remaining"`
	At        time.Time `json:"at"`
}
type Meter struct {
	Version      int           `json:"version"`
	Observations []Observation `json:"observations"`
}
type Projection struct {
	Agent          string   `json:"agent"`
	Unit           string   `json:"unit"`
	Status         string   `json:"status"`
	Remaining      *float64 `json:"remaining"`
	UnitsPerCredit *float64 `json:"units_per_credit"`
	UnitsPerToken  *float64 `json:"units_per_token"`
	UnitsPerHour   *float64 `json:"units_per_hour"`
	HoursRemaining *float64 `json:"hours_remaining"`
}

func ValidateObservation(o Observation) error {
	supported := false
	for _, agent := range Agents() {
		if o.Agent == agent {
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf("unsupported meter agent %q", o.Agent)
	}
	if o.Unit != "credits" && o.Unit != "percent" {
		return fmt.Errorf("meter unit must be credits or percent")
	}
	if o.At.IsZero() || math.IsNaN(o.Remaining) || math.IsInf(o.Remaining, 0) || o.Remaining < 0 || o.Unit == "percent" && o.Remaining > 100 {
		return fmt.Errorf("invalid meter observation")
	}
	return nil
}
func MeterPoints(m Meter, agent, unit string) []Observation {
	var out []Observation
	for _, o := range m.Observations {
		if o.Agent == agent && o.Unit == unit {
			out = append(out, o)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}
func ValidateMeter(m Meter) error {
	if m.Version != 0 && m.Version != 1 {
		return fmt.Errorf("unsupported meter version %d", m.Version)
	}
	counts := map[string]int{}
	last := map[string]Observation{}
	for _, o := range m.Observations {
		if e := ValidateObservation(o); e != nil {
			return e
		}
		key := o.Agent + "/" + o.Unit
		counts[key]++
		if counts[key] > 2 {
			return fmt.Errorf("too many meter points for %s", key)
		}
		if p, ok := last[key]; ok && (!o.At.After(p.At) || o.Remaining > p.Remaining) {
			return fmt.Errorf("invalid meter ordering for %s", key)
		}
		last[key] = o
	}
	return nil
}

// ObserveMeter owns retention and reset semantics. It has no IO.
func ObserveMeter(m Meter, o Observation) (Meter, error) {
	if e := ValidateMeter(m); e != nil {
		return Meter{}, e
	}
	if e := ValidateObservation(o); e != nil {
		return Meter{}, e
	}
	points := MeterPoints(m, o.Agent, o.Unit)
	if len(points) > 0 {
		last := points[len(points)-1]
		if o.At.Before(last.At) {
			return Meter{}, fmt.Errorf("stale meter observation")
		}
		if o.At.Equal(last.At) {
			if o.Remaining == last.Remaining {
				return m, nil
			}
			return Meter{}, fmt.Errorf("conflicting meter observation at same timestamp")
		}
		if o.Remaining > last.Remaining {
			points = nil
		}
	}
	points = append(points, o)
	if len(points) > 2 {
		points = points[len(points)-2:]
	}
	out := Meter{Version: 1}
	for _, p := range m.Observations {
		if p.Agent != o.Agent || p.Unit != o.Unit {
			out.Observations = append(out.Observations, p)
		}
	}
	out.Observations = append(out.Observations, points...)
	sort.Slice(out.Observations, func(i, j int) bool {
		a, b := out.Observations[i], out.Observations[j]
		if a.Agent != b.Agent {
			return a.Agent < b.Agent
		}
		if a.Unit != b.Unit {
			return a.Unit < b.Unit
		}
		return a.At.Before(b.At)
	})
	return out, nil
}

// ProjectMeter consumes measured tokens between the two readings, across all
// local slots. Attribution gaps do not change the account denominator.
func ProjectMeter(m Meter, agent, unit string, now time.Time, tokens int64, burn float64, complete bool) Projection {
	return projectMeter(m, agent, unit, now, float64(tokens), burn, complete, false)
}

func ProjectCreditMeter(m Meter, agent, unit string, now time.Time, credits CreditAmount, burn float64, complete bool) Projection {
	return projectMeter(m, agent, unit, now, credits.Float64(), burn, complete, true)
}
func projectMeter(m Meter, agent, unit string, now time.Time, measured, burn float64, complete, credit bool) Projection {
	p := Projection{Agent: agent, Unit: unit, Status: "insufficient_calibration"}
	points := MeterPoints(m, agent, unit)
	if len(points) == 0 {
		return p
	}
	latest := points[len(points)-1]
	p.Remaining = &latest.Remaining
	if len(points) < 2 {
		return p
	}
	if e := ValidateMeter(m); e != nil {
		p.Status = "invalid_calibration"
		return p
	}
	if now.Before(latest.At) || now.Sub(latest.At) > 7*24*time.Hour || now.Sub(points[0].At) > 7*24*time.Hour {
		p.Status = "stale_calibration"
		return p
	}
	if !complete {
		p.Status = "incomplete_evidence"
		return p
	}
	if measured <= 0 {
		p.Status = "no_measured_tokens"
		if credit {
			p.Status = "no_measured_credits"
		}
		return p
	}
	delta := points[0].Remaining - latest.Remaining
	if delta <= 0 {
		p.Status = "nondecreasing_meter"
		return p
	}
	if burn <= 0 || math.IsNaN(burn) || math.IsInf(burn, 0) {
		p.Status = "no_recent_burn"
		return p
	}
	rate := delta / measured
	hour := rate * burn
	hours := latest.Remaining / hour
	if math.IsNaN(rate) || math.IsInf(rate, 0) || math.IsNaN(hour) || math.IsInf(hour, 0) || math.IsNaN(hours) || math.IsInf(hours, 0) {
		p.Status = "invalid_projection"
		return p
	}
	p.Status = "calibrated"
	if credit {
		p.UnitsPerCredit = &rate
	} else {
		p.UnitsPerToken = &rate
	}
	p.UnitsPerHour = &hour
	p.HoursRemaining = &hours
	return p
}
