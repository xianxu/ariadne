package project

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
)

// trainforecast.go — the pure forecast core (#329). A project's remaining work
// is its trains' unlanded cars; each car's expected focused hours are its
// estimate × the calibrated multiplier. Trains run in parallel, each serially,
// and a ⇒ edge holds a car until its prerequisite lands, so the time to done is
// the DAG's critical path. The output is a range of focused hours, never a
// date: a date is set against the forecast, never by it. No IO.

// CarEstimate is what the caller resolved for one car from its issue card.
type CarEstimate struct {
	Estimate    float64 // estimate_hours; 0 when Unestimated
	Terminal    bool    // landed (or otherwise ended): counts 0
	Unestimated bool    // no estimate: the calibration's class prior stands in
	LookupErr   string  // the card couldn't be read: the prior stands in, noted
}

// TrainForecast is the computed range plus what the statement cites.
type TrainForecast struct {
	Remaining, RemainingLow, RemainingHigh float64 // Σ over unlanded cars
	Path, PathLow, PathHigh                float64 // critical path
	LastTrain                              string  // the train the critical path ends in
	ActiveTrains, Trains                   int
	Cal                                    estimate.ForecastCalibration
	Notes                                  []string
}

// ForecastTrains schedules p's cars. A car missing from cars is treated as a
// lookup failure (prior, noted). An edge whose prerequisite is in no train is
// outside this project's plan and is ignored. A cycle is an error naming a car
// on it.
func ForecastTrains(p TrainPlan, cars map[IssueRef]CarEstimate, cal estimate.ForecastCalibration) (TrainForecast, error) {
	f := TrainForecast{Trains: len(p.Trains), Cal: cal}
	if len(p.Trains) == 0 {
		return f, fmt.Errorf("no trains to forecast (## Trains is empty)")
	}

	// Nodes in first-seen order (deterministic); preds = previous car in its
	// train(s) plus in-plan ⇒ prerequisites.
	var order []IssueRef
	trainOf := map[IssueRef]string{}
	preds := map[IssueRef]map[IssueRef]bool{}
	addPred := func(to, from IssueRef) {
		if preds[to] == nil {
			preds[to] = map[IssueRef]bool{}
		}
		preds[to][from] = true
	}
	for _, t := range p.Trains {
		for i, c := range t.Cars {
			if _, seen := trainOf[c]; !seen {
				trainOf[c] = t.Name
				order = append(order, c)
			}
			if i > 0 && t.Cars[i-1] != c {
				addPred(c, t.Cars[i-1])
			}
		}
	}
	for _, e := range p.Edges {
		_, fromIn := trainOf[e.From]
		_, toIn := trainOf[e.To]
		if fromIn && toIn && e.From != e.To {
			addPred(e.To, e.From)
		}
	}

	// Per-car hours at the three multipliers.
	var unestimated, missing []string
	hours := map[IssueRef][3]float64{}
	mult := [3]float64{cal.Low, cal.Point, cal.High}
	for _, c := range order {
		ce, ok := cars[c]
		if !ok {
			ce = CarEstimate{LookupErr: "not resolved"}
		}
		if ce.Terminal && ce.LookupErr == "" {
			continue // landed: 0 hours
		}
		est := ce.Estimate
		switch {
		case ce.LookupErr != "":
			est = cal.Prior
			missing = append(missing, c.String()+" ("+ce.LookupErr+")")
		case ce.Unestimated || est <= 0:
			est = cal.Prior
			unestimated = append(unestimated, c.String())
		}
		var h [3]float64
		for i := range mult {
			h[i] = est * mult[i]
		}
		hours[c] = h
	}

	// Kahn's algorithm over the DAG; finish(c) = max(finish(preds)) + hours(c).
	indeg := map[IssueRef]int{}
	succ := map[IssueRef][]IssueRef{}
	for _, c := range order {
		for pr := range preds[c] {
			indeg[c]++
			succ[pr] = append(succ[pr], c)
		}
	}
	for k := range succ {
		sort.Slice(succ[k], func(i, j int) bool { return succ[k][i].String() < succ[k][j].String() })
	}
	var queue []IssueRef
	for _, c := range order {
		if indeg[c] == 0 {
			queue = append(queue, c)
		}
	}
	finish := map[IssueRef][3]float64{}
	done := 0
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		done++
		var start [3]float64
		for pr := range preds[c] {
			for i := range start {
				if finish[pr][i] > start[i] {
					start[i] = finish[pr][i]
				}
			}
		}
		h := hours[c]
		var fin [3]float64
		for i := range fin {
			fin[i] = start[i] + h[i]
		}
		finish[c] = fin
		for _, s := range succ[c] {
			if indeg[s]--; indeg[s] == 0 {
				queue = append(queue, s)
			}
		}
	}
	if done < len(order) {
		return f, fmt.Errorf("the trains and ⇒ edges form a cycle through %s", carOnCycle(order, preds, indeg))
	}

	var end IssueRef
	for _, c := range order {
		h := hours[c]
		f.RemainingLow += h[0]
		f.Remaining += h[1]
		f.RemainingHigh += h[2]
		fin := finish[c]
		if fin[1] > f.Path {
			f.Path, end = fin[1], c
		}
		if fin[0] > f.PathLow {
			f.PathLow = fin[0]
		}
		if fin[2] > f.PathHigh {
			f.PathHigh = fin[2]
		}
	}
	if f.Path > 0 {
		f.LastTrain = trainOf[end]
	}
	for _, t := range p.Trains {
		for _, c := range t.Cars {
			if hours[c][1] > 0 {
				f.ActiveTrains++
				break
			}
		}
	}
	if cal.Collapsed {
		f.Notes = append(f.Notes, fmt.Sprintf("only %d calibration rows (range needs %d of weight at least %.1f); the range is the point", cal.N, estimate.MinRangeRows, estimate.MinRangeWeight))
	}
	if cal.Partial > 0 {
		f.Notes = append(f.Notes, fmt.Sprintf("%d of %d calibration rows are partial actuals, weighted by their attributed share (#603)", cal.Partial, cal.N))
	}
	if len(unestimated) > 0 {
		f.Notes = append(f.Notes, fmt.Sprintf("unestimated, at the %.1fh prior: %s", cal.Prior, strings.Join(unestimated, ", ")))
	}
	if len(missing) > 0 {
		f.Notes = append(f.Notes, fmt.Sprintf("card not read, at the %.1fh prior: %s", cal.Prior, strings.Join(missing, ", ")))
	}
	return f, nil
}

// carOnCycle names a car that lies ON a cycle, not merely downstream of one.
// After Kahn's algorithm stalls, every stalled car has a stalled predecessor
// (otherwise it would have been released), so walking stalled predecessors from
// any stalled car must revisit a car; the first revisited car is on the cycle.
// The smallest predecessor by name is taken at each step, so the answer is
// deterministic.
func carOnCycle(order []IssueRef, preds map[IssueRef]map[IssueRef]bool, indeg map[IssueRef]int) IssueRef {
	var c IssueRef
	for _, o := range order {
		if indeg[o] > 0 {
			c = o
			break
		}
	}
	seen := map[IssueRef]bool{}
	for !seen[c] {
		seen[c] = true
		var next IssueRef
		found := false
		for p := range preds[c] {
			if indeg[p] > 0 && (!found || p.String() < next.String()) {
				next, found = p, true
			}
		}
		if !found {
			return c // unreachable for a stalled car; keeps the walk total
		}
		c = next
	}
	return c
}

// RenderTrainForecast is the one-paragraph statement every surface prints.
func RenderTrainForecast(f TrainForecast) string {
	var b strings.Builder
	if f.Remaining == 0 {
		fmt.Fprintf(&b, "forecast: every car of %d train(s) has landed; 0h remaining", f.Trains)
	} else {
		fmt.Fprintf(&b, "forecast: critical path %.1f–%.1fh (point %.1fh, ends in train %q); %.1fh remaining across %d of %d train(s)",
			f.PathLow, f.PathHigh, f.Path, f.LastTrain, f.Remaining, f.ActiveTrains, f.Trains)
	}
	fmt.Fprintf(&b, "; calibration n=%d, ×%.2f (%.2f–%.2f), k=%.2f", f.Cal.N, f.Cal.Point, f.Cal.Low, f.Cal.High, f.Cal.K)
	for _, n := range f.Notes {
		fmt.Fprintf(&b, ". %s", n)
	}
	return b.String()
}
