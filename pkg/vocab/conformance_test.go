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

// TestVerdictConformance: every verdict token is in exactly one category with
// non-empty `when`, the Is*/Emitted predicates agree with the categories, and the
// system-internal tokens are never emittable. Fail-closed, derived from the model (#147).
func TestVerdictConformance(t *testing.T) {
	m := Verdict()
	cats := []string{"finalizing", "blocking", "internal"}

	var all []string
	for _, cat := range cats {
		all = append(all, m.Categories[cat]...)
	}
	for _, tok := range all {
		n := 0
		for _, cat := range cats {
			if m.inCategory(cat, tok) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("token %q is in %d categories, want exactly 1", tok, n)
		}
		if m.When[tok] == "" {
			t.Errorf("token %q has no `when` semantics", tok)
		}
		wantEmitted := m.inCategory("finalizing", tok) || m.inCategory("blocking", tok)
		if m.IsEmitted(tok) != wantEmitted {
			t.Errorf("IsEmitted(%q)=%v, want %v", tok, m.IsEmitted(tok), wantEmitted)
		}
		if m.IsFinalizing(tok) != m.inCategory("finalizing", tok) {
			t.Errorf("IsFinalizing(%q) disagrees with its category", tok)
		}
		if m.IsBlocking(tok) != m.inCategory("blocking", tok) {
			t.Errorf("IsBlocking(%q) disagrees with its category", tok)
		}
		if contains(m.Emitted(), tok) != wantEmitted {
			t.Errorf("Emitted() membership of %q=%v, want %v", tok, contains(m.Emitted(), tok), wantEmitted)
		}
	}
	for _, tok := range m.Categories["internal"] {
		if m.IsEmitted(tok) {
			t.Errorf("system-internal token %q must not be reviewer-emittable", tok)
		}
	}
}

// TestProjectConformance mirrors TestIssueConformance for the project noun
// (ariadne#180): fail-closed, derived from the embedded model.
func TestProjectConformance(t *testing.T) {
	m := Project()

	// Every status is in exactly one category and carries a non-empty `when`.
	for _, s := range m.AllStatuses() {
		n := 0
		for _, cat := range projectCategoryOrder {
			if inCat(m.Categories, cat, s) {
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

	// Every declared transition references known statuses, agrees with
	// CanTransition, and is returned by TransitionFor.
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
		if got := m.TransitionFor(tr.From, tr.To); got == nil || got.Event != tr.Event {
			t.Errorf("TransitionFor(%s, %s) = %+v, want event %q", tr.From, tr.To, got, tr.Event)
		}
	}

	// The deliberate absence: no paused→done edge (close requires executing).
	if m.CanTransition("paused", "done") {
		t.Errorf("model declares paused→done; the design requires resume-first")
	}
	if m.TransitionFor("paused", "done") != nil {
		t.Errorf("TransitionFor(paused, done) must be nil")
	}
}
