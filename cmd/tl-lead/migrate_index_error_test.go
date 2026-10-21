package main

import (
	"maps"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateAnswersRefusesUnreadableDiscussionIndex(t *testing.T) {
	for _, owner := range []string{"session", "outside"} {
		for name, index := range map[string]string{
			"null doc after valid": "discussions:\n  - doc: qa/op/a.md\n  - doc: null\n",
			"duplicate doc":        "discussions:\n  - doc: qa/op/a.md\n    doc: qa/op/b.md\n",
			"duplicate index":      "discussions:\n  - doc: qa/op/a.md\ndiscussions: []\n",
		} {
			t.Run(owner+"/"+name, func(t *testing.T) {
				root, session, f := migrateFixture(t)
				writeFile(t, filepath.Join(root, "qa/op/a.md"), "---\ntype: side-doc\n---\n\n# An answer\n")
				text := "---\noperator: op\n" + index + "---\n\n# Index owner\n"
				file := session
				if owner == "session" {
					if err := f.load(text); err != nil {
						t.Fatal(err)
					}
				} else {
					file = filepath.Join(root, "outside.md")
				}
				writeFile(t, file, text)
				snapshot, _, err := readMigration(deps{ss: f}, session, "")
				if err != nil {
					t.Fatal(err)
				}
				plan := planMigration(snapshot)
				rel, err := filepath.Rel(root, file)
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, conflict := range plan.Conflicts {
					if conflict.Path == filepath.ToSlash(rel) && strings.Contains(conflict.Why, "discussions:") {
						found = true
					}
				}
				if !found {
					t.Fatalf("missing named index conflict for %s: %+v", rel, plan.Conflicts)
				}
				before, registered := tree(t, root, f), maps.Clone(f.registered)
				var out strings.Builder
				steps := 0
				err = migrateAnswers(deps{ss: f}, &out, session, "", true, func() error { steps++; return nil })
				if exitCode(err) != 1 || !strings.Contains(out.String(), "conflict "+filepath.ToSlash(rel)+": discussions:") || !strings.Contains(out.String(), "nothing was written") {
					t.Fatalf("apply must refuse with named conflict: %v\n%s", err, out.String())
				}
				if !maps.Equal(before, tree(t, root, f)) || !maps.Equal(registered, f.registered) || f.applies != 0 || steps != 0 {
					t.Fatal("refused migration changed source files, registrations or applied edits")
				}
			})
		}
	}
}
