package estimate

import (
	"fmt"
	"math"
)

// Failure is one reconciliation violation, carrying a next-action message (the
// error IS the spec — it names exactly what to fix).
type Failure struct {
	Message string
}

// Check validates a parsed Block against the issue's frontmatter estimate_hours.
// It returns one Failure per violated rule; an empty slice means the estimate
// reconciles. Pure. The tolerance absorbs 1-decimal rounding:
// tol = max(0.05, 5% of total).
func Check(b Block, estimateHours float64) []Failure {
	var fs []Failure
	add := func(msg string) { fs = append(fs, Failure{Message: msg}) }

	switch {
	case b.Model == "":
		add("## Estimate: missing `model:` provenance line (name the model, e.g. estimate-logic-v2)")
	case !KnownModel(b.Model):
		add(fmt.Sprintf("## Estimate: unknown model %q — recognized: %v", b.Model, Models()))
	}

	if len(b.Items) == 0 {
		add("## Estimate: no `item:` lines — itemize the estimate by v2 primitive")
	}
	for _, it := range b.Items {
		if !KnownPrimitive(it.Slug) {
			add(fmt.Sprintf("## Estimate: unknown primitive %q — see `sdlc change-code --help` / helptext/estimate.md for the closed vocabulary", it.Slug))
		}
	}

	tol := math.Max(0.05, 0.05*b.Total)
	if recomputed := b.Recomputed(); math.Abs(recomputed-b.Total) > tol {
		add(fmt.Sprintf("## Estimate `total: %.2f` ≠ recomputed %.2f (Σdesign×%.2f + Σimpl×%.2f); fix the items or the total",
			b.Total, recomputed, 1+b.DesignBuffer, b.Familiarity))
	}
	if math.Abs(b.Total-estimateHours) > tol {
		add(fmt.Sprintf("frontmatter `estimate_hours: %.2f` ≠ ## Estimate `total: %.2f`; make them match", estimateHours, b.Total))
	}
	return fs
}
