package issue

import (
	"strings"
	"testing"
)

// TestCheckStructural walks the four deterministic gates that
// `sdlc change-code` enforces before letting implementation begin:
// Spec present, Plan present, Done-when present, Estimate present.
// Each is refusable with --force; the test pins the failure shape
// (name + message) so callers can render and act on them.
func TestCheckStructural(t *testing.T) {
	good := joinDoc(
		`---
id: 000099
status: working
estimate_hours: 1.5
related: [cmd/sdlc/changecode.go]
---`,
		"# Title",
		"",
		"## Problem",
		strings.Repeat("word ", 30),
		"",
		"## Done when",
		"- something works",
		"",
		"## Spec",
		strings.Repeat("word ", 60), // >50 words
		"",
		"## Plan",
		"- [ ] do the thing",
		"",
		"## Log",
		"",
	)

	tests := []struct {
		name        string
		text        string
		wantFailure []string // names of expected failures
	}{
		{"all gates pass", good, nil},
		{
			"missing Spec section",
			strings.Replace(good, "## Spec", "## Other", 1),
			[]string{"spec-present"},
		},
		{
			"empty Spec under 50 words",
			strings.Replace(good, strings.Repeat("word ", 60), "tiny", 1),
			[]string{"spec-present"},
		},
		{
			"empty Plan checklist",
			strings.Replace(good, "- [ ] do the thing", "- [ ] ", 1),
			[]string{"plan-present"},
		},
		{
			"no Plan checkbox items at all",
			strings.Replace(good, "- [ ] do the thing", "just prose, no checklist", 1),
			[]string{"plan-present"},
		},
		{
			"missing Done-when AND no related frontmatter",
			strings.Replace(
				strings.Replace(good, "## Done when\n- something works\n", "", 1),
				"related: [cmd/sdlc/changecode.go]\n", "", 1,
			),
			[]string{"done-when-present"},
		},
		{
			"missing Done-when but related populated — passes",
			strings.Replace(good, "## Done when\n- something works\n", "", 1),
			nil,
		},
		{
			"missing estimate_hours",
			strings.Replace(good, "estimate_hours: 1.5\n", "", 1),
			[]string{"estimate-present"},
		},
		{
			"zero estimate_hours fails",
			strings.Replace(good, "estimate_hours: 1.5", "estimate_hours: 0", 1),
			[]string{"estimate-present"},
		},
		{
			"non-numeric estimate_hours fails",
			strings.Replace(good, "estimate_hours: 1.5", "estimate_hours: TBD", 1),
			[]string{"estimate-present"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckStructural(tt.text)
			names := failureNames(got)
			if !sameStrings(names, tt.wantFailure) {
				t.Errorf("CheckStructural failures = %v, want %v", names, tt.wantFailure)
			}
		})
	}
}

func TestCheckStructural_MissingFrontmatterIsHardFail(t *testing.T) {
	// No frontmatter at all should produce a single "frontmatter-present"
	// failure (and not panic). Without frontmatter the other gates can't
	// even run, so we short-circuit at the top.
	got := CheckStructural("# Just a title\n\nNo frontmatter here.\n")
	names := failureNames(got)
	if len(names) != 1 || names[0] != "frontmatter-present" {
		t.Errorf("missing frontmatter: got %v, want [frontmatter-present]", names)
	}
}

func joinDoc(parts ...string) string {
	return strings.Join(parts, "\n") + "\n"
}

func failureNames(fs []StructuralFailure) []string {
	if len(fs) == 0 {
		return nil
	}
	names := make([]string, len(fs))
	for i, f := range fs {
		names[i] = f.Name
	}
	return names
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
