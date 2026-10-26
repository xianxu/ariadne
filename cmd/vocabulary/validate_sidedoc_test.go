package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTitleCaseNouns(t *testing.T) {
	for _, tc := range []struct{ noun, definition string }{
		{"", ""},
		{"issue", "Issue"},
		{"pensive", "Pensive"},
		{"side-doc", "SideDoc"},
		{"fleet-policy", "FleetPolicy"},
	} {
		t.Run(tc.noun, func(t *testing.T) {
			if got := titleCase(tc.noun); got != tc.definition {
				t.Errorf("titleCase(%q) = %q, want %q", tc.noun, got, tc.definition)
			}
		})
	}
}

func TestValidateInstance_SideDoc(t *testing.T) {
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue not on PATH")
	}
	schema := schemaT(t, "side-doc")
	for _, tc := range []struct {
		name, fm, wantField string
	}{
		{"minimal without marker or lifecycle", "type: side-doc\nparent: absent-plan.md\n", ""},
		{"nested parent and additive metadata", "type: side-doc\nparent: plan.side/other.md\nmarker: evidence\ncustom: {nested: [one, two]}\n", ""},
		{"missing type", "parent: plan.md\n", "type"},
		{"wrong type", "type: issue\nparent: plan.md\n", "type"},
		{"null type", "type: null\nparent: plan.md\n", "type"},
		{"missing parent", "type: side-doc\n", "parent"},
		{"empty parent", "type: side-doc\nparent: \"\"\n", "parent"},
		{"null parent", "type: side-doc\nparent: null\n", "parent"},
		{"numeric parent", "type: side-doc\nparent: 623\n", "parent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			md := filepath.Join(t.TempDir(), "evidence.md")
			if err := os.WriteFile(md, []byte("---\n"+tc.fm+"---\n# Evidence\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			diags, err := validateInstanceFile(osCue{}, md, schema, "side-doc")
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantField == "" {
				if len(diags) != 0 {
					t.Fatalf("valid side-doc rejected: %v", diags)
				}
				return
			}
			for _, d := range diags {
				if d.Field == tc.wantField || strings.Contains(d.Message, tc.wantField) {
					return
				}
			}
			t.Fatalf("want rejection naming %q, got %v", tc.wantField, diags)
		})
	}
}
