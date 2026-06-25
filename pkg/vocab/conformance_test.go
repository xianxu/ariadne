package vocab

import "testing"

// TestIssueConformance is a fail-closed check, NOT a maintained list: every case
// is derived from the embedded model, so a status or transition the model can't
// place fails here automatically. (A consumer's own coverage is its responsibility;
// this guards the model's internal consistency at the binding.)
func TestIssueConformance(t *testing.T) {
	m := Issue()

	// Every status is in exactly one category and carries a non-empty `when`.
	for _, s := range m.AllStatuses() {
		n := 0
		for _, cat := range []string{"open", "active", "terminal"} {
			if m.inCategory(cat, s) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("status %q is in %d categories, want exactly 1", s, n)
		}
		if m.When[s] == "" {
			t.Errorf("status %q has no `when` semantics", s)
		}
	}

	// Every declared transition references known statuses and is accepted by
	// CanTransition (the API and the data agree).
	for _, tr := range m.Lifecycle {
		if !contains(m.AllStatuses(), tr.From) {
			t.Errorf("transition from unknown status %q", tr.From)
		}
		if !contains(m.AllStatuses(), tr.To) {
			t.Errorf("transition to unknown status %q", tr.To)
		}
		if !m.CanTransition(tr.From, tr.To) {
			t.Errorf("declared transition %s→%s not accepted by CanTransition", tr.From, tr.To)
		}
	}
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
