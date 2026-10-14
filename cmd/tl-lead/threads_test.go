package main

import "testing"

func TestUpsertThread(t *testing.T) {
	for _, tc := range []struct{ name, body, entry, want string }{
		{"replaces the same heading in place",
			"### Project: p\nold\n\n### Miscellaneous\n- x", "### Project: p\nnew",
			"### Project: p\nnew\n\n### Miscellaneous\n- x"},
		{"inserts before Miscellaneous",
			"### Project: p\ns\n\n### Miscellaneous\n- x", "### Train: t\ncars",
			"### Project: p\ns\n\n### Train: t\ncars\n\n### Miscellaneous\n- x"},
		{"appends when there is no Miscellaneous", "### Project: p\ns", "### Train: t",
			"### Project: p\ns\n\n### Train: t"},
		{"keeps a preamble", "intro line\n\n### Project: p", "### Train: t",
			"intro line\n\n### Project: p\n\n### Train: t"},
		{"an empty section", "", "### Miscellaneous\n- ask", "### Miscellaneous\n- ask"},
	} {
		got, err := upsertThread(tc.body, tc.entry, nil)
		if err != nil || got != tc.want {
			t.Errorf("%s: got %q, %v; want %q", tc.name, got, err, tc.want)
		}
	}
}
