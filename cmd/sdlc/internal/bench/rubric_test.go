package bench

import "testing"

func TestDefaultRubricGroups(t *testing.T) {
	r := DefaultRubric()
	groups := map[string]bool{}
	for _, d := range r.Subjective {
		groups[d.Group] = true
	}
	for _, c := range r.Objective {
		groups[c.Group] = true
	}
	if !groups["quality"] || !groups["workflow-fit"] {
		t.Fatalf("both groups must be present, got %v", groups)
	}
	if len(r.Objective) == 0 || len(r.Subjective) == 0 {
		t.Fatalf("rubric must have objective and subjective checks")
	}
}
