package judge

import (
	"strings"
	"testing"
)

// TestArchitectureSectionsLossless: the registry splits into a preamble and one
// section per ARCH-* entry with nothing lost — re-joined, it is the registry
// byte for byte — so rendering all markers stays "the registry verbatim".
func TestArchitectureSectionsLossless(t *testing.T) {
	pre, secs, err := architectureSections(ArchitectureRegistry)
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString(pre)
	var markers []string
	for _, s := range secs {
		b.WriteString(s.text)
		markers = append(markers, s.marker)
	}
	if b.String() != ArchitectureRegistry {
		t.Error("preamble + sections do not re-join to the registry")
	}
	if strings.Join(markers, ",") != strings.Join(ArchitectureMarkers(), ",") {
		t.Errorf("section markers %v, want %v", markers, ArchitectureMarkers())
	}
}

// TestEveryPrincipleDeclaresQuickFlow: each entry must say whether the quick
// flow's small-diff review checks it, so a new principle cannot be added
// without that decision (#231).
func TestEveryPrincipleDeclaresQuickFlow(t *testing.T) {
	good := "# pre\n\n## ARCH-A — a\n\n- **quick-flow:** yes\n- **principle:** x\n\n## ARCH-B — b\n\n- **quick-flow:** no\n- **principle:** y\n"
	if _, _, err := architectureSections(good); err != nil {
		t.Fatalf("well-formed fixture: %v", err)
	}
	for name, bad := range map[string]string{
		"missing": "# pre\n\n## ARCH-A — a\n\n- **principle:** x\n",
		"maybe":   "# pre\n\n## ARCH-A — a\n\n- **quick-flow:** maybe\n- **principle:** x\n",
		"twice":   "# pre\n\n## ARCH-A — a\n\n- **quick-flow:** yes\n- **quick-flow:** no\n",
	} {
		if _, _, err := architectureSections(bad); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}

// TestQuickMarkers: the quick flow checks ARCH-DRY, ARCH-PURE and ARCH-PURPOSE —
// the operator's choice, recorded as a field on each registry entry, read here.
func TestQuickMarkers(t *testing.T) {
	if got := strings.Join(QuickMarkers(), ","); got != "ARCH-DRY,ARCH-PURE,ARCH-PURPOSE" {
		t.Errorf("QuickMarkers() = %s", got)
	}
}

// TestArchitectureBlockForSubset: the subset block carries the header, the
// preamble and only the selected entries; the full set is the old block exactly.
func TestArchitectureBlockForSubset(t *testing.T) {
	sub := ArchitectureBlockFor("at-review", QuickMarkers())
	if !strings.Contains(sub, "each of the 3 entries") {
		t.Errorf("subset header does not count 3 entries:\n%.200s", sub)
	}
	for _, m := range QuickMarkers() {
		if !strings.Contains(sub, "## "+m+" ") {
			t.Errorf("subset block lacks %s", m)
		}
	}
	if strings.Contains(sub, "## ARCH-MOCK") || strings.Contains(sub, "## ARCH-ORDER") {
		t.Error("subset block carries an unselected entry")
	}
	full := ArchitectureBlock("at-plan")
	want := "ARCHITECTURE PRINCIPLES — work through each of the 8 entries below explicitly, applying its `at-plan` lens; " +
		"cite the marker (e.g. ARCH-DRY) in any finding.\n\n" + ArchitectureRegistry
	if full != want {
		t.Error("the full block is no longer the header + the registry verbatim")
	}
}

// TestSmallDiffRecipeCarriesExactlyQuickMarkers: every ARCH-* marker in the
// rendered small-diff prompt is one the registry marks quick-flow: yes, and all
// of them are there — a registry edit flows into the recipe with no other change.
func TestSmallDiffRecipeCarriesExactlyQuickMarkers(t *testing.T) {
	got := markersIn(BuildPrompt(SmallDiffReview, goldenInput))
	if strings.Join(got, ",") != strings.Join(QuickMarkers(), ",") {
		t.Errorf("small-diff prompt carries %v, want exactly %v", got, QuickMarkers())
	}
}

// TestSmallDiffRecipeHasFocusAndContract: the recipe aims at the bug classes a
// small diff actually ships, and keeps the machine contract the boundary gate
// parses (verdict, findings, the pinned review window, prior rounds).
func TestSmallDiffRecipeHasFocusAndContract(t *testing.T) {
	p := BuildPrompt(SmallDiffReview, goldenInput)
	for _, want := range []string{
		"Enumerate the family", "in BOTH", "Doc and comment claims", "Boundaries",
		BoundaryReviewContract, "Prior rounds", goldenInput.ReviewWindow,
	} {
		if !strings.Contains(p, want) {
			t.Errorf("small-diff prompt lacks %q", want)
		}
	}
}

// TestMilestoneReviewRendersViaBoundaryTail: the shared tail is expanded into
// milestone-review too, so both recipes carry one copy of the prior-rounds and
// contract prose.
func TestMilestoneReviewRendersViaBoundaryTail(t *testing.T) {
	p := BuildPrompt(MilestoneReview, goldenInput)
	if strings.Contains(p, "{{") {
		t.Errorf("milestone-review prompt has an unexpanded token:\n%s", p[strings.Index(p, "{{"):][:60])
	}
	for _, want := range []string{"Prior rounds", BoundaryReviewContract, goldenInput.ReviewWindow} {
		if !strings.Contains(p, want) {
			t.Errorf("milestone-review prompt lacks %q", want)
		}
	}
}
