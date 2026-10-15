package issue

import (
	"strings"
	"testing"
)

func TestParseDetails(t *testing.T) {
	got, err := ParseDetails("## Problem\n\nIt breaks.\n\n## Spec\n\nFix it:\n\n```md\n## Not a heading\n# nor a title\n```\n\n## Done when\n\n- it works\n")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"Problem":   "It breaks.",
		"Spec":      "Fix it:\n\n```md\n## Not a heading\n# nor a title\n```",
		"Done when": "- it works",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d sections %v, want %d", len(got), got, len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
	// An explicitly empty section is kept as empty.
	if got, err := ParseDetails("## Plan\n"); err != nil || got["Plan"] != "" || len(got) != 1 {
		t.Errorf("empty Plan: %v %v", got, err)
	}
}

func TestParseDetailsRefuses(t *testing.T) {
	for _, c := range []struct{ name, text, want string }{
		{"empty", " \n\n", "empty"},
		{"frontmatter", "---\nid: 1\n---\n## Spec\nx\n", "frontmatter"},
		{"title", "# Title\n## Spec\nx\n", "`# ` title"},
		{"preamble", "intro\n## Spec\nx\n", "before the first"},
		{"fenced preamble", "```\nx\n```\n## Spec\nx\n", "before the first"},
		{"unknown", "## Spec\nx\n## Design\ny\n", "not an issue section"},
		{"duplicate", "## Spec\nx\n## Spec\ny\n", "appears twice"},
		{"unclosed fence", "## Spec\n```\n## Done when\n", "never closes"},
	} {
		_, err := ParseDetails(c.text)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: err %v, want it to mention %q", c.name, err, c.want)
		}
	}
}

// Supplied sections replace the template's; the rest (Log's dated heading)
// stay as the template has them.
func TestRenderWithSections(t *testing.T) {
	out := Render(ScaffoldSpec{ID: "000009", Title: "T", Today: "2026-10-15",
		Sections: map[string]string{"Problem": "P body", "Done when": "- done"}})
	for _, want := range []string{"## Problem\n\nP body\n\n", "## Done when\n\n- done\n\n", "## Log\n\n### 2026-10-15\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("render lacks %q:\n%s", want, out)
		}
	}
	if blank := Render(ScaffoldSpec{ID: "000009", Title: "T", Today: "2026-10-15"}); blank == out {
		t.Error("sections had no effect")
	}
}
