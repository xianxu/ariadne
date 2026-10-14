package main

import "testing"

func TestPlanBar(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"milestones counted when present",
			"## Plan\n\n- [x] M1 — a\n  - [ ] sub\n- [ ] M2 — b\n- [ ] M3 — c\n\n## Log\n", "▉▕▕ M 1/3"},
		{"checkboxes otherwise", "## Plan\n\n- [x] one\n- [ ] two\n- [ ] three\n- [ ] four\n", "▉▕▕▕ tasks 1/4"},
		{"no plan section", "## Spec\n\ntext\n", "no plan yet"},
		{"empty plan", "## Plan\n\n## Log\n", "no plan yet"},
		{"fenced checkboxes ignored", "## Plan\n\n- [x] real\n```\n- [ ] quoted\n```\n", "▉ tasks 1/1"},
	} {
		if got := planBar(tc.body); got != tc.want {
			t.Errorf("%s: planBar = %q, want %q", tc.name, got, tc.want)
		}
	}
}
