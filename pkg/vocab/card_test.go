package vocab

import "testing"

func TestCardOwnershipDiscovery(t *testing.T) {
	d := Issue().Discovery()
	if d.Home != "workshop/issues" || d.Cards != "workshop/issue-cards" || d.Tracker != "issue-tracker" {
		t.Fatalf("discovery: %+v", d)
	}
	want := map[string]bool{"id": true, "status": true, "started": true, "created": true, "updated": true, "estimate_hours": true, "actual_hours": true, "github_issue": true, "claimant": true, "title": true}
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
