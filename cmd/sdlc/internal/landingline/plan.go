package landingline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// The machine plan (#475): one file per machine, beside the suite lock, that
// places each train-enabled repository on a wall-clock departure grid so heavy
// suites don't collide in the machine's suite slots. The steward writes it
// (`sdlc landing plan --write`); every admission reads it. Unknown members, at
// the top and in each repository's entry, are kept on decode and written back
// unchanged (ARCH-EVOLVE).

// PlanVersion is the plan file's format; a reader refuses another.
const PlanVersion = 1

// MachinePlan is the decoded plan file.
type MachinePlan struct {
	Version int
	Slots   int // the machine's full-suite slots the plan packed into
	Repos   map[string]PlanEntry
	extra   map[string]json.RawMessage
}

// PlanEntry is one repository's place on the grid and why it is there.
type PlanEntry struct {
	OffsetMinutes int
	Basis         string // human: the cost the offset was computed from
	extra         map[string]json.RawMessage
}

// Offset is the repository's grid offset for a train interval, and whether
// the plan places it at all. An offset outside the interval is reduced
// modulo it (an interval changed under the plan), never refused here.
func (p MachinePlan) Offset(repo string, interval time.Duration) (time.Duration, bool) {
	e, ok := p.Repos[repo]
	if !ok || interval <= 0 {
		return 0, false
	}
	return time.Duration(e.OffsetMinutes) * time.Minute % interval, true
}

// DecodePlan parses a plan file: a version it knows, a slot count, and
// offsets that are whole minutes in [0, 24h). Anything else names what is
// wrong; the caller falls back to today's clock, visibly.
func DecodePlan(data []byte) (MachinePlan, error) {
	var top map[string]json.RawMessage
	if err := strictObject(data, &top); err != nil {
		return MachinePlan{}, fmt.Errorf("plan: %w", err)
	}
	p := MachinePlan{Repos: map[string]PlanEntry{}, extra: map[string]json.RawMessage{}}
	for k, v := range top {
		var err error
		switch k {
		case "version":
			err = json.Unmarshal(v, &p.Version)
		case "slots":
			err = json.Unmarshal(v, &p.Slots)
		case "repos":
			var repos map[string]json.RawMessage
			if err = strictObject(v, &repos); err == nil {
				for name, raw := range repos {
					e, eerr := decodeEntry(raw)
					if eerr != nil {
						return MachinePlan{}, fmt.Errorf("plan: repos.%s: %w", name, eerr)
					}
					p.Repos[name] = e
				}
			}
		default:
			p.extra[k] = v
		}
		if err != nil {
			return MachinePlan{}, fmt.Errorf("plan: %s: %w", k, err)
		}
	}
	switch {
	case p.Version != PlanVersion:
		return MachinePlan{}, fmt.Errorf("plan: version %d, this sdlc reads %d", p.Version, PlanVersion)
	case p.Slots < 1:
		return MachinePlan{}, fmt.Errorf("plan: slots %d is not a positive slot count", p.Slots)
	}
	return p, nil
}

func decodeEntry(data []byte) (PlanEntry, error) {
	var m map[string]json.RawMessage
	if err := strictObject(data, &m); err != nil {
		return PlanEntry{}, err
	}
	e := PlanEntry{extra: map[string]json.RawMessage{}}
	seen := false
	for k, v := range m {
		var err error
		switch k {
		case "offset_minutes":
			seen = true
			err = json.Unmarshal(v, &e.OffsetMinutes)
		case "basis":
			err = json.Unmarshal(v, &e.Basis)
		default:
			e.extra[k] = v
		}
		if err != nil {
			return PlanEntry{}, fmt.Errorf("%s: %w", k, err)
		}
	}
	if !seen || e.OffsetMinutes < 0 || e.OffsetMinutes >= 24*60 {
		return PlanEntry{}, fmt.Errorf("offset_minutes %d is missing or outside 0..1439", e.OffsetMinutes)
	}
	return e, nil
}

// strictObject decodes one JSON object, refusing trailing data.
func strictObject(data []byte, into *map[string]json.RawMessage) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(into); err != nil {
		return fmt.Errorf("not a JSON object: %v", err)
	}
	if dec.More() {
		return fmt.Errorf("trailing data after the object")
	}
	if *into == nil {
		return fmt.Errorf("not a JSON object")
	}
	return nil
}

// MarshalJSON writes the known members and every unknown one it was read
// with; a known member is never overridden by an unknown of the same name.
func (p MachinePlan) MarshalJSON() ([]byte, error) {
	out := map[string]any{}
	for k, v := range p.extra {
		out[k] = v
	}
	repos := map[string]any{}
	for name, e := range p.Repos {
		m := map[string]any{}
		for k, v := range e.extra {
			m[k] = v
		}
		m["offset_minutes"], m["basis"] = e.OffsetMinutes, e.Basis
		repos[name] = m
	}
	out["version"], out["slots"], out["repos"] = p.Version, p.Slots, repos
	return json.MarshalIndent(out, "", "  ")
}

// WithOffsets is p with each repository's offset and basis replaced by plan's,
// keeping every entry's unknown members and every repository plan omits.
func (p MachinePlan) WithOffsets(slots int, placed map[string]PlanEntry) MachinePlan {
	n := MachinePlan{Version: PlanVersion, Slots: slots, Repos: map[string]PlanEntry{}, extra: p.extra}
	for name, e := range p.Repos {
		n.Repos[name] = e
	}
	for name, e := range placed {
		if old, ok := n.Repos[name]; ok {
			e.extra = old.extra
		}
		n.Repos[name] = e
	}
	return n
}

// RepoCost is what one repository's train suite costs the machine.
type RepoCost struct {
	Repo     string
	Weight   int           // full-suite slots it holds (1, or 2 for ariadne's sharded suite)
	Duration time.Duration // its measured p90 wall time
}

// PlanStep is the offsets' granularity, and the buffer each reservation
// carries for integration and start-up around the suite itself.
const PlanStep = 5 * time.Minute

// PackOffsets places each repository on a grid of interval so that, at every
// minute, the summed weight of the suites reserved there stays within slots.
// Heaviest (weight × duration) first, ties by name; each takes the earliest
// step-aligned offset that fits, else the one whose peak is lowest. A
// reservation is the duration rounded up to the step, plus one step.
func PackOffsets(costs []RepoCost, interval time.Duration, slots int) map[string]time.Duration {
	out := map[string]time.Duration{}
	minutes := int(interval / time.Minute)
	if minutes <= 0 {
		return out
	}
	order := append([]RepoCost(nil), costs...)
	sort.SliceStable(order, func(i, j int) bool {
		ci, cj := order[i].Weight*int(order[i].Duration), order[j].Weight*int(order[j].Duration)
		if ci != cj {
			return ci > cj
		}
		return order[i].Repo < order[j].Repo
	})
	load := make([]int, minutes)
	step := int(PlanStep / time.Minute)
	for _, c := range order {
		w := max(c.Weight, 1)
		span := int((c.Duration+PlanStep-1)/PlanStep)*step + step
		if span > minutes {
			span = minutes
		}
		best, bestPeak := 0, int(^uint(0)>>1)
		for o := 0; o < minutes; o += step {
			peak := 0
			for m := o; m < o+span; m++ {
				peak = max(peak, load[m%minutes]+w)
			}
			if peak <= slots {
				best, bestPeak = o, peak
				break
			}
			if peak < bestPeak {
				best, bestPeak = o, peak
			}
		}
		for m := best; m < best+span; m++ {
			load[m%minutes] += w
		}
		out[c.Repo] = time.Duration(best) * time.Minute
	}
	return out
}
