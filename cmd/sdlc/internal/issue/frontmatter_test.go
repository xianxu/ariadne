package issue

import (
	"strings"
	"testing"
)

func TestParse_RoundTrip(t *testing.T) {
	doc := "---\nid: 000031\nstatus: working\n---\n# title\n\nbody here\n"
	fm, body, err := Parse(doc)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if fm != "id: 000031\nstatus: working" {
		t.Errorf("fm mismatch: %q", fm)
	}
	if body != "# title\n\nbody here\n" {
		t.Errorf("body mismatch: %q", body)
	}
	if got := Compose(fm, body); got != doc {
		t.Errorf("Compose round-trip mismatch:\n  want %q\n  got  %q", doc, got)
	}
}

func TestParse_NoFrontmatter(t *testing.T) {
	if _, _, err := Parse("# no frontmatter here\n"); err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestParse_EmptyFrontmatter(t *testing.T) {
	doc := "---\n\n---\nbody\n"
	fm, body, err := Parse(doc)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if fm != "" {
		t.Errorf("expected empty fm, got %q", fm)
	}
	if body != "body\n" {
		t.Errorf("body mismatch: %q", body)
	}
}

func TestGetField(t *testing.T) {
	fm := "id: 000031\nstatus: working\nestimate_hours: 4\nactual_hours:\n"
	tests := []struct {
		name      string
		wantValue string
		wantOK    bool
	}{
		{"id", "000031", true},
		{"status", "working", true},
		{"estimate_hours", "4", true},
		{"actual_hours", "", true},
		{"missing", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetField(fm, tt.name)
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.wantValue {
				t.Errorf("value = %q, want %q", got, tt.wantValue)
			}
		})
	}
}

func TestSetField_ReplacePreservesOrder(t *testing.T) {
	fm := "id: 000031\nstatus: working\nestimate_hours: 4\nactual_hours:"
	got := SetField(fm, "actual_hours", "6.5")
	want := "id: 000031\nstatus: working\nestimate_hours: 4\nactual_hours: 6.5"
	if got != want {
		t.Errorf("SetField replace mismatch:\n  want %q\n  got  %q", want, got)
	}
}

func TestSetField_AppendsWhenAbsent(t *testing.T) {
	fm := "id: 000031\nstatus: working"
	got := SetField(fm, "actual_hours", "6.5")
	want := "id: 000031\nstatus: working\nactual_hours: 6.5"
	if got != want {
		t.Errorf("SetField append mismatch:\n  want %q\n  got  %q", want, got)
	}
}

func TestSetField_AppendsTrimsTrailingNewlines(t *testing.T) {
	// The close-issue.py path is: new_fm.rstrip() + "\n<field>: <value>"
	fm := "id: 000031\nstatus: working\n\n"
	got := SetField(fm, "updated", "2026-05-25")
	want := "id: 000031\nstatus: working\nupdated: 2026-05-25"
	if got != want {
		t.Errorf("SetField append-with-trailing-ws mismatch:\n  want %q\n  got  %q", want, got)
	}
}

func TestSetField_StatusFlip(t *testing.T) {
	fm := "id: 000031\nstatus: working\nestimate_hours: 4"
	got := SetField(fm, "status", "done")
	if !strings.Contains(got, "status: done") {
		t.Errorf("expected status: done in %q", got)
	}
	if strings.Contains(got, "status: working") {
		t.Errorf("status: working still present in %q", got)
	}
}

func TestSetField_UpsertChain(t *testing.T) {
	// Simulate close-issue.py's issue-close chain: status, actual_hours, updated.
	fm := "id: 000031\nstatus: working\nestimate_hours: 4\nactual_hours:"
	fm = SetField(fm, "status", "done")
	fm = SetField(fm, "actual_hours", "6.5")
	fm = SetField(fm, "updated", "2026-05-25")
	want := "id: 000031\nstatus: done\nestimate_hours: 4\nactual_hours: 6.5\nupdated: 2026-05-25"
	if fm != want {
		t.Errorf("upsert chain mismatch:\n  want %q\n  got  %q", want, fm)
	}
}
