package issue

import (
	"strings"
	"testing"
)

const setterCard = "---\nid: 000001\nstatus: open\nestimate_hours:\n---\n\n# Old title\n\n## Problem\nReport.\n"

func TestSetCardFieldChangesOnlyAnOwnedField(t *testing.T) {
	out, err := SetCardField([]byte(setterCard), "estimate_hours", "3")
	if err != nil || string(out) != strings.Replace(setterCard, "estimate_hours:\n", "estimate_hours: 3\n", 1) {
		t.Fatalf("%q %v", out, err)
	}
	for _, c := range []struct{ name, value string }{
		{"deps", "[a]"}, {"title", "x"}, {"estimate_hours", "1\nstatus: done"}, {"estimate_hours", "-2"}, {"status", "bogus"},
	} {
		if _, err := SetCardField([]byte(setterCard), c.name, c.value); err == nil {
			t.Errorf("accepted %s=%q", c.name, c.value)
		}
	}
}

func TestSetCardTitleReplacesOnlyTheH1(t *testing.T) {
	out, err := SetCardTitle([]byte(setterCard), "  New title ")
	if err != nil || string(out) != strings.Replace(setterCard, "# Old title", "# New title", 1) {
		t.Fatalf("%q %v", out, err)
	}
	for _, bad := range []string{"", "  ", "two\nlines"} {
		if _, err := SetCardTitle([]byte(setterCard), bad); err == nil {
			t.Errorf("accepted title %q", bad)
		}
	}
}
