package judge

import (
	"strings"
	"testing"
)

func TestArchitectureEvolveContract(t *testing.T) {
	entry, err := architectureEntry(ArchitectureRegistry, "ARCH-EVOLVE")
	if err != nil {
		t.Fatal(err)
	}
	for clause, wants := range map[string][]string{
		"principle": {"evolve additively", "tolerate and preserve unknown fields", "capture them on decode", "write them back unchanged", "nested members", "known-field types", "required fields", "duplicate keys", "never changes the meaning", "ignoring a field would be unsafe"},
		"at-plan":   {"each persisted format", "why ignoring", "unknown fields", "read-modify-write"},
		"at-review": {"unknown fields", "unchanged", "required fields", "duplicate keys", "version", "unsafe"},
	} {
		text, err := architectureClause(entry, clause)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range wants {
			if !strings.Contains(text, want) {
				t.Errorf("%s lacks %q", clause, want)
			}
		}
	}
	if !strings.Contains(entry, "- **quick-flow:** yes") {
		t.Fatal("small record rewrites must check compatibility too")
	}
	for _, category := range []Category{PlanQuality, MilestoneReview, SmallDiffReview} {
		if !strings.Contains(BuildPrompt(category, goldenInput), entry) {
			t.Errorf("%s lost the complete evolution principle", category)
		}
	}
}
