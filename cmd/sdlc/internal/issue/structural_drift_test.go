package issue

import (
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// TestGatedSectionsSubsetOfModel enforces the invariant chain's left edge:
// every section the structural gate validates must exist in the cue creation
// template (gatedSections ⊆ scaffold.sections). A gate that requires a section
// `sdlc issue new` never writes would reject every fresh issue.
func TestGatedSectionsSubsetOfModel(t *testing.T) {
	model := map[string]bool{}
	var names []string
	for _, s := range vocab.Issue().Sections() {
		model[s.Name] = true
		names = append(names, s.Name)
	}
	for _, g := range gatedSections {
		if !model[g] {
			t.Errorf("structural gate targets %q, absent from issue.cue scaffold.sections %v — "+
				"reconcile structural.go (and PlanSectionBody if it's Plan) or issue.cue", g, names)
		}
	}
}
