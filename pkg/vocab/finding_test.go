package vocab

import (
	"strings"
	"testing"
)

// TestFindingConformance is a fail-closed check, NOT a maintained list: every case is
// derived from the embedded model, so a severity or disposition the model can't place
// fails here automatically. Mirrors TestIssueConformance's posture.
//
// The disposition half is the load-bearing one: adding a disposition to finding.cue
// without placing it in the closing/open partition fails HERE, before it can reach
// gatestate.OpenFindings as an unhandled case that silently wedges a finding open forever.
func TestFindingConformance(t *testing.T) {
	m := Finding()

	// Every severity is in exactly one category and carries non-empty `when` semantics.
	for _, s := range m.Severities() {
		n := 0
		for _, cat := range []string{"blocking", "advisory"} {
			if contains(m.Categories[cat], s) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("severity %q is in %d categories, want exactly 1", s, n)
		}
		if m.When[s] == "" {
			t.Errorf("severity %q has no `when` semantics", s)
		}
		// Blocks() must agree with the category data it derives from.
		if want := contains(m.Categories["blocking"], s); m.Blocks(s) != want {
			t.Errorf("Blocks(%q) = %v, disagrees with categories.blocking", s, m.Blocks(s))
		}
	}

	// Every disposition is in exactly one category, carries semantics, and Closes()
	// agrees with the partition.
	for _, d := range m.AllDispositions() {
		n := 0
		for _, cat := range []string{"closing", "open"} {
			if contains(m.Dispositions[cat], d) {
				n++
			}
		}
		if n != 1 {
			t.Errorf("disposition %q is in %d categories, want exactly 1", d, n)
		}
		if m.WhenDisposed[d] == "" {
			t.Errorf("disposition %q has no semantics", d)
		}
		if !m.IsDisposition(d) {
			t.Errorf("IsDisposition(%q) = false for a modeled disposition", d)
		}
		if want := contains(m.Dispositions["closing"], d); m.Closes(d) != want {
			t.Errorf("Closes(%q) = %v, disagrees with dispositions.closing", d, m.Closes(d))
		}
	}

	// hardBlocking must be a SUBSET of blocking: a severity that blocks only AFTER the
	// round cap — but not before it — would be incoherent.
	for _, s := range m.HardBlocking {
		if !m.Blocks(s) {
			t.Errorf("hardBlocking %q is not in categories.blocking", s)
		}
		if !m.BlocksPastCap(s) {
			t.Errorf("BlocksPastCap(%q) = false for a hardBlocking severity", s)
		}
	}

	// The model must reject what it doesn't declare — the parser's fail-closed contract
	// rests on these.
	if m.IsSeverity("Catastrophic") {
		t.Error("IsSeverity accepted an unmodeled severity")
	}
	if m.IsDisposition("deferred") {
		t.Error("IsDisposition accepted an unmodeled disposition")
	}
}

// TestFindingRenderBlockInstruction pins that the prompt's handoff instruction is rendered
// FROM the model — every severity, every disposition, and the fence itself. This is what
// keeps the prompt's accepted set from drifting out of finding.cue (#187).
func TestFindingRenderBlockInstruction(t *testing.T) {
	m := Finding()
	out := m.RenderBlockInstruction()

	if !strings.Contains(out, "```findings") {
		t.Error("instruction must show the fenced ```findings block")
	}
	for _, s := range m.Severities() {
		if !strings.Contains(out, s) {
			t.Errorf("instruction omits severity %q", s)
		}
		if !strings.Contains(out, m.When[s]) {
			t.Errorf("instruction omits the gloss for severity %q", s)
		}
	}
	for _, d := range m.AllDispositions() {
		if !strings.Contains(out, d) {
			t.Errorf("instruction omits disposition %q", d)
		}
		if !strings.Contains(out, m.WhenDisposed[d]) {
			t.Errorf("instruction omits the gloss for disposition %q", d)
		}
	}
	// The judge must be told the binary owns id assignment — otherwise it invents ids
	// and the cross-round references stop resolving.
	if !strings.Contains(out, "id: new") {
		t.Error("instruction must tell the judge to emit `id: new` for a new finding")
	}
}

// TestFindingSeverityOrder pins blocking-before-advisory ordering, which the prompt and
// the ledger rendering both rely on for a stable, severity-descending presentation.
func TestFindingSeverityOrder(t *testing.T) {
	got := Finding().Severities()
	want := []string{"Critical", "Important", "Minor"}
	if len(got) != len(want) {
		t.Fatalf("Severities() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Severities()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// #194 close review BR-29: the model-drift guard must pin the `family` key. The fence is
// the ONLY place the judge is told to emit it; drop the line and every family count stays
// zero while FamilyCounts, the escalation and the convergence signal all look correct.
func TestRenderBlockInstruction_NamesTheFamilyKey(t *testing.T) {
	got := Finding().RenderBlockInstruction()
	for _, want := range []string{"family: <slug>", "underlying RULE", "REUSE the matching slug"} {
		if !strings.Contains(got, want) {
			t.Errorf("the emitted fence instruction must carry %q, or families are never populated", want)
		}
	}
}
