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

// TestGetField_EmptyMiddleField is the regression for the `\s*`-spans-newline
// bug: an empty field followed by another line must return "", not the next
// line's value. (The case above only had an empty field at EOF, which hid it.)
func TestGetField_EmptyMiddleField(t *testing.T) {
	fm := "id: 000056\nstatus: open\ngithub_issue:\ncreated: 2026-05-31\n"
	if got, ok := GetField(fm, "github_issue"); !ok || got != "" {
		t.Errorf("GetField(github_issue) = %q, %v; want \"\", true", got, ok)
	}
	if got, _ := GetField(fm, "created"); got != "2026-05-31" {
		t.Errorf("GetField(created) = %q; want 2026-05-31", got)
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

// TestSetFieldRoundTripsFlowMap: the #231 flow record is a one-line YAML map
// precisely so these line-based helpers can carry it. Setting it twice must
// replace the line in place (no duplicate, no orphaned child lines), and the
// quoted hashes and braces must come back byte-identical — `$` and `\` are the
// only ReplaceAllString hazards, and the record never contains them.
func TestSetFieldRoundTripsFlowMap(t *testing.T) {
	const v1 = `{kind: quick, provenance: inferred, spec: "12345678", done: "5e6f7a8b"}`
	const v2 = `{kind: full, provenance: inferred}`
	fm := "id: 000001\nstatus: working\nestimate_hours:"
	fm = SetField(fm, "flow", v1)
	if got, ok := GetField(fm, "flow"); !ok || got != v1 {
		t.Fatalf("after first set: GetField = %q, %v; want %q", got, ok, v1)
	}
	fm = SetField(fm, "flow", v2)
	if got, _ := GetField(fm, "flow"); got != v2 {
		t.Errorf("after second set: GetField = %q, want %q", got, v2)
	}
	if n := strings.Count(fm, "\nflow:"); n != 1 {
		t.Errorf("flow: appears %d times after two sets, want 1:\n%s", n, fm)
	}
	if got, _ := GetField(fm, "estimate_hours"); got != "" {
		t.Errorf("neighbouring empty field disturbed: estimate_hours = %q", got)
	}
}

// TestSetFieldReplacesBlockValue: frontmatter is hand-editable, so a field may
// arrive in block form (`flow:` followed by indented `kind:`/`provenance:`
// lines, or a `related:` block list). Replacing only the key line used to leave
// the indented children orphaned under the new value — invalid YAML written by
// the very rewrite meant to repair it (#231 M1 review). The whole block goes.
func TestSetFieldReplacesBlockValue(t *testing.T) {
	fm := "id: 000001\nflow:\n  kind: quick\n  provenance: inferred\nstatus: working\nrelated:\n  - a.md\n  - b.md"
	fm = SetField(fm, "flow", "{kind: full, provenance: inferred}")
	fm = SetField(fm, "related", "[a.md]")
	want := "id: 000001\nflow: {kind: full, provenance: inferred}\nstatus: working\nrelated: [a.md]"
	if fm != want {
		t.Errorf("SetField left the block behind:\n got %q\nwant %q", fm, want)
	}
}
