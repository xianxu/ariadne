// Package usage normalizes reported local response usage. Counters are PURE:
// input includes caches, and reasoning is a subset of output, never extra spend.
package usage

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

type Counters struct {
	Input      int64  `json:"input"`
	Output     int64  `json:"output"`
	CacheRead  *int64 `json:"cache_read"`
	CacheWrite *int64 `json:"cache_write"`
	Reasoning  *int64 `json:"reasoning"`
}

func checkedAdd(a, b int64) (int64, error) {
	if a < 0 || b < 0 || a > math.MaxInt64-b {
		return 0, fmt.Errorf("negative or overflowing usage counters")
	}
	return a + b, nil
}
func (c Counters) Total() (int64, error) { return checkedAdd(c.Input, c.Output) }
func (c Counters) Validate() error {
	if _, e := c.Total(); e != nil {
		return e
	}
	for _, d := range []struct {
		v     *int64
		limit int64
	}{{c.CacheRead, c.Input}, {c.CacheWrite, c.Input}, {c.Reasoning, c.Output}} {
		if d.v != nil && (*d.v < 0 || *d.v > d.limit) {
			return fmt.Errorf("invalid usage subset")
		}
	}
	return nil
}
func samePtr(a, b *int64) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
func (c Counters) equal(d Counters) bool {
	return c.Input == d.Input && c.Output == d.Output && samePtr(c.CacheRead, d.CacheRead) && samePtr(c.CacheWrite, d.CacheWrite) && samePtr(c.Reasoning, d.Reasoning)
}
func (c Counters) monotone(d Counters) bool {
	if c.Input > d.Input || c.Output > d.Output {
		return false
	}
	for _, x := range [][2]*int64{{c.CacheRead, d.CacheRead}, {c.CacheWrite, d.CacheWrite}, {c.Reasoning, d.Reasoning}} {
		if (x[0] == nil) != (x[1] == nil) || x[0] != nil && *x[0] > *x[1] {
			return false
		}
	}
	return true
}
func sumPtr(a, b *int64) (*int64, error) {
	if a == nil || b == nil {
		return nil, nil
	}
	v, e := checkedAdd(*a, *b)
	return &v, e
}

// Add drops an unsupported dimension to null; its supported subtotal is retained
// separately by report aggregation in MeasuredDimensions.
func (c Counters) Add(d Counters) (Counters, error) {
	var r Counters
	var e error
	if r.Input, e = checkedAdd(c.Input, d.Input); e != nil {
		return r, e
	}
	if r.Output, e = checkedAdd(c.Output, d.Output); e != nil {
		return r, e
	}
	if r.CacheRead, e = sumPtr(c.CacheRead, d.CacheRead); e != nil {
		return r, e
	}
	if r.CacheWrite, e = sumPtr(c.CacheWrite, d.CacheWrite); e != nil {
		return r, e
	}
	r.Reasoning, e = sumPtr(c.Reasoning, d.Reasoning)
	return r, e
}

type Response struct {
	ID          string    `json:"id"`
	SourceID    string    `json:"source_id"`
	Agent       string    `json:"agent"`
	Model       string    `json:"model"`
	RequestID   string    `json:"request_id"`
	At          time.Time `json:"at"`
	Tokens      Counters  `json:"tokens"`
	Provisional bool      `json:"provisional"`
	Origin      string    `json:"origin"`
}
type Coverage struct {
	Counts   map[string]int64    `json:"counts"`
	Examples map[string][]string `json:"examples"`
}

func (c *Coverage) Add(kind, example string) {
	if c.Counts == nil {
		c.Counts = map[string]int64{}
	}
	c.Counts[kind]++
	if example != "" {
		if c.Examples == nil {
			c.Examples = map[string][]string{}
		}
		if len(c.Examples[kind]) < 3 {
			example = strings.ReplaceAll(example, "\n", " ")
			if len(example) > 160 {
				example = example[:160]
			}
			c.Examples[kind] = append(c.Examples[kind], example)
		}
	}
}
func (c *Coverage) Merge(d Coverage) {
	if c.Counts == nil {
		c.Counts = map[string]int64{}
	}
	for k, v := range d.Counts {
		c.Counts[k] += v
	}
	for k, ss := range d.Examples {
		for _, s := range ss {
			if c.Examples == nil {
				c.Examples = map[string][]string{}
			}
			if len(c.Examples[k]) < 3 {
				c.Examples[k] = append(c.Examples[k], s)
			}
		}
	}
}
func (c Coverage) Complete() bool {
	if !c.SpendComplete() {
		return false
	}
	for _, k := range []string{"unknown_model", "unknown_slot", "attribution_conflicts", "unknown_metadata", "invalid_lineage", "dispatch_scan_incomplete", "legacy_reviewer_binding_gaps", "unavailable_reviewer_bindings", "unfinished_reviewer_bindings", "reviewer_binding_bound", "excluded_unbound_responses"} {
		if c.Counts[k] > 0 {
			return false
		}
	}
	return true
}

// SpendComplete separates known spend from its attribution. Unknown checkout,
// model or role does not remove tokens from the account-wide denominator.
func (c Coverage) SpendComplete() bool {
	for _, k := range []string{"conflicts", "malformed_records", "unsupported_formats", "baseline_gaps", "reset_gaps", "provisional", "truncated_tails", "unavailable_sources", "bound_exceeded", "source_mutations", "uncovered_segments", "reconstructed_compaction", "unavailable_source", "candidate_limit", "unsupported_harness", "aggregate_usage_unreconciled"} {
		if c.Counts[k] > 0 {
			return false
		}
	}
	return true
}

type Window struct {
	Since time.Time `json:"since"`
	Until time.Time `json:"until"`
}

func (w Window) contains(t time.Time) bool { return !t.Before(w.Since) && t.Before(w.Until) }

type Attribution struct {
	Slot, Role, Round, Attempt string
	Include, Conflict          bool
}

const UnknownSlot = "unknown-slot"

type Group struct {
	Slot              string   `json:"slot"`
	Agent             string   `json:"agent"`
	Model             string   `json:"model"`
	Role              string   `json:"role"`
	Round             string   `json:"round"`
	Attempt           string   `json:"attempt"`
	Tokens            Counters `json:"tokens"`
	Responses         int64    `json:"responses"`
	BurnTokensPerHour float64  `json:"burn_tokens_per_hour"`
}
type Hour struct {
	At            time.Time `json:"at"`
	DurationHours float64   `json:"duration_hours"`
	Tokens        Counters  `json:"tokens"`
	Responses     int64     `json:"responses"`
}
type Report struct {
	Groups             []Group  `json:"groups"`
	Hourly             []Hour   `json:"hourly"`
	Totals             Counters `json:"totals"`
	MeasuredDimensions Counters `json:"measured_dimensions"`
	BurnTokensPerHour  float64  `json:"burn_tokens_per_hour"`
	Coverage           Coverage `json:"coverage"`
}

func (c Coverage) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Counts              map[string]int64    `json:"counts"`
		Examples            map[string][]string `json:"examples"`
		SpendComplete       bool                `json:"spend_complete"`
		AttributionComplete bool                `json:"attribution_complete"`
	}{c.Counts, c.Examples, c.SpendComplete(), c.Complete()})
}
