// Package estimate is the pure core of the estimate-shell (#117): it parses the
// `## Estimate` fenced block an issue carries, reconciles it against the issue's
// estimate_hours, and formats estimate↔actual calibration-ledger rows. It has no
// IO — the three thin seams (the change-code reconciliation guard, the
// estimate-quality judge, and the close-time ledger append) inject these pure
// functions. The shell is model-agnostic: it enforces that whatever model the
// `model:` provenance line names was actually applied (today estimate-logic-v3.1).
package estimate

// Item is one line of an estimate breakdown: a v2-lineage primitive slug and its
// design + impl hours. The hours are the agent's judgment (post spec-quality
// discount and model-version calibration per the model doc); this package only
// sums and reconciles them.
type Item struct {
	Slug   string
	Design float64
	Impl   float64
}

// Block is a parsed `## Estimate` fenced block.
type Block struct {
	Model        string  // provenance — the model version applied
	Familiarity  float64 // multiplies impl (v2 Step 5); default 1.0
	DesignBuffer float64 // lifts design (v2 Step 6); default 0.30
	Total        float64 // the asserted total, reconciled against Recomputed + estimate_hours
	Items        []Item
}

// Recomputed is the deterministic reconciliation total:
// Σdesign × (1+design-buffer) + Σimpl × familiarity (the v2-lineage shell).
func (b Block) Recomputed() float64 {
	var d, i float64
	for _, it := range b.Items {
		d += it.Design
		i += it.Impl
	}
	return d*(1+b.DesignBuffer) + i*b.Familiarity
}
