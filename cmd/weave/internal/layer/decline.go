package layer

import (
	"path"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
)

// Declination is one of the leaf's `decline` rows (#467) and the inherited
// rows it removed. An empty Removed is a stale opt-out, which the compile
// reports so it stays visible.
type Declination struct {
	Path    string
	Removed []intent.Intent
	// Extra is a second argument on the row: decline takes one path, so the
	// row declines nothing and the compile warns.
	Extra string
}

// Decline applies the leaf's `decline` rows: every ancestor row with a target
// at or under a declined path is dropped, and the leaf records what each row
// removed. The leaf's own rows are never declined, and an ancestor's decline
// rows are ignored, so a layer cannot opt its consumers out. Rows without a
// single target (prose, skill, ship), and an ancestor's internal rows (never
// selected for the leaf), are unaffected. No IO; it updates the given layers
// in place and runs once after the layer walk, so the plan, dry-run and the
// completeness check all see the same rows.
func Decline(layers []Layer) []Layer {
	if len(layers) == 0 {
		return layers
	}
	leaf := &layers[len(layers)-1]
	leaf.Declined = nil
	for _, in := range leaf.Intents {
		if in.Kind != intent.Decline {
			continue
		}
		d := Declination{Path: path.Clean(in.Source)}
		if in.Target != in.Source {
			d.Extra = in.Target
		}
		leaf.Declined = append(leaf.Declined, d)
	}
	if leaf.Declined == nil {
		return layers
	}
	for i := range layers[:len(layers)-1] {
		kept := layers[i].Intents[:0:0]
		for _, in := range layers[i].Intents {
			if d := declinedBy(leaf.Declined, in); d != nil {
				d.Removed = append(d.Removed, in)
				continue
			}
			kept = append(kept, in)
		}
		layers[i].Intents = kept
	}
	return layers
}

// declinedBy returns the declination whose path covers in's target, if in
// is a row that writes one target.
func declinedBy(ds []Declination, in intent.Intent) *Declination {
	if !intent.Selected(in.Visibility, false) {
		return nil
	}
	switch in.Kind {
	case intent.Symlink, intent.Seed, intent.SeedOnce, intent.Scaffold, intent.Touch, intent.Merge:
	default:
		return nil
	}
	target := path.Clean(in.Target)
	for i := range ds {
		if ds[i].Extra != "" {
			continue
		}
		if target == ds[i].Path || strings.HasPrefix(target, ds[i].Path+"/") {
			return &ds[i]
		}
	}
	return nil
}
