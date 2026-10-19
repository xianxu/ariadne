package vocab

import "testing"

func TestCardOwnershipDiscovery(t *testing.T) {
	d := Issue().Discovery()
	if d.Home != "workshop/issues" || d.Cards != "workshop/issue-cards" || d.Tracker != "issue-tracker" {
		t.Fatalf("discovery: %+v", d)
	}
	want := map[string]bool{"id": true, "status": true, "started": true, "created": true, "updated": true, "estimate_hours": true, "actual_hours": true, "github_issue": true, "claimant": true, "title": true, "area": true, "severity": true, "flake_tests": true}
	for _, field := range Issue().CardFields() {
		if !want[field.Name] || field.Kind == "" || field.Setter == "" {
			t.Fatalf("unexpected/duplicate/incomplete field: %+v", field)
		}
		delete(want, field.Name)
	}
	if len(want) != 0 {
		t.Fatalf("missing fields: %v", want)
	}
}

// #348: the classification fields name their value sets, and those sets exist.
func TestCardEnums(t *testing.T) {
	m := Issue()
	if got := m.Enums["severity"]; len(got) != 4 || got[0] != "critical" || got[3] != "low" {
		t.Fatalf("severity enum: %v", got)
	}
	if m.RepoEnums.DeclarationPath != ".sdlc/vocabulary.json" {
		t.Fatalf("declaration path: %q", m.RepoEnums.DeclarationPath)
	}
	area := m.RepoEnums.Fields["area"]
	if area.Syntax != "path" || area.Segment == "" || area.MaxDepth != 4 || !area.Multi || area.MaxItems != 8 {
		t.Fatalf("area spec: %+v", area)
	}
	for _, f := range m.CardFields() {
		switch f.Kind {
		case "enum":
			if _, ok := m.Enums[f.Enum]; !ok {
				t.Fatalf("%s names undeclared enum %q", f.Name, f.Enum)
			}
		case "repo-enum":
			if _, ok := m.RepoEnums.Fields[f.Enum]; !ok {
				t.Fatalf("%s names undeclared repo enum %q", f.Name, f.Enum)
			}
		}
	}
}
