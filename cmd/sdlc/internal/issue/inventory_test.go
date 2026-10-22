package issue

import "testing"

func TestIssueInventoryPath(t *testing.T) {
	roots := []string{"workshop/issues", "workshop/history", "workshop/history/issues"}
	for _, tc := range []struct {
		p, id string
		role  InventoryRole
	}{
		{"workshop/issues/000001-one.md", "000001", InventoryActive},
		{"workshop/history/000002-old.md", "000002", InventoryArchive},
		{"workshop/history/issues/000003-old.md", "000003", InventoryArchive},
		{"workshop/issues/side/000001-one.md", "", ""},
		{"workshop/history/plans/000001-one.md", "", ""},
		{"workshop/issues/../issues/000001-one.md", "", ""},
		{"./workshop/issues/000001-one.md", "", ""},
		{"/workshop/issues/000001-one.md", "", ""},
		{"workshop//issues/000001-one.md", "", ""},
		{"workshop/issues/000001-one.md/", "", ""},
		{"workshop/issues/00001-one.md", "", ""},
		{"workshop/issues/readme.md", "", ""},
		{"workshop/issues/000001-one.md\x00", "", ""},
		{"workshop/issues/000001-one\\x.md", "000001", InventoryActive},
	} {
		t.Run(tc.p, func(t *testing.T) {
			id, role, ok := IssueInventoryPath(tc.p, roots)
			if id != tc.id || role != tc.role || ok != (tc.id != "") {
				t.Fatalf("got %q %q %v", id, role, ok)
			}
		})
	}
	for _, roots := range [][]string{nil, {"./workshop/issues"}, {"/workshop/issues"}, {"workshop//issues"}, {"workshop/../workshop/issues"}} {
		if _, _, ok := IssueInventoryPath("workshop/issues/000001-one.md", roots); ok {
			t.Fatalf("invalid roots accepted: %v", roots)
		}
	}
}

func FuzzIssueInventoryPath(f *testing.F) {
	for _, p := range []string{"workshop/issues/000001-one.md", "workshop/history/000001-one.md", "workshop/history/issues/000001-one.md", "workshop/issues/../issues/000001-one.md", ""} {
		f.Add(p)
	}
	f.Fuzz(func(t *testing.T, p string) {
		roots := []string{"workshop/issues", "workshop/history", "workshop/history/issues"}
		id, role, ok := IssueInventoryPath(p, roots)
		if !ok {
			if id != "" || role != "" {
				t.Fatal("invalid path leaked classification")
			}
			return
		}
		want, _, parsed := ParseFilename(p)
		if !parsed || id != want || (role != InventoryActive && role != InventoryArchive) {
			t.Fatal("invalid classification")
		}
	})
}
