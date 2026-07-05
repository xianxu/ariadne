package issue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Simple Title", "simple-title"},
		{"with-Existing-Dashes", "with-existing-dashes"},
		{"  leading whitespace", "leading-whitespace"},
		{"Trailing punctuation!", "trailing-punctuation"},
		{"--multiple--dashes--", "multiple-dashes"},
		{"Caps  AND   Spaces", "caps-and-spaces"},
		{"Symbols / & special?", "symbols-special"},
		{"Numbers 42 keep", "numbers-42-keep"},
		{"UPPER", "upper"},
		{"unicode café", "unicode-caf"}, // accent stripped, matches sed [^a-z0-9]
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := Slugify(c.in); got != c.want {
				t.Errorf("Slugify(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNextID(t *testing.T) {
	dir := t.TempDir()
	issues := filepath.Join(dir, "issues")
	history := filepath.Join(dir, "history")
	mustMkdir(t, issues, history)

	// Highest is 000031 in issues; history has lower numbers.
	writeFiles(t, history, "000005-old.md", "000010-older.md")
	writeFiles(t, issues, "000020-a.md", "000031-b.md", "not-an-issue.md")

	got, err := NextID(issues, history)
	if err != nil {
		t.Fatal(err)
	}
	if got != "000032" {
		t.Errorf("NextID = %q, want 000032", got)
	}
}

func TestNextID_HighestInHistory(t *testing.T) {
	dir := t.TempDir()
	issues := filepath.Join(dir, "issues")
	history := filepath.Join(dir, "history")
	mustMkdir(t, issues, history)
	writeFiles(t, history, "000099-done.md") // archived closed issue
	writeFiles(t, issues, "000050-active.md")

	got, err := NextID(issues, history)
	if err != nil {
		t.Fatal(err)
	}
	if got != "000100" {
		t.Errorf("NextID = %q, want 000100", got)
	}
}

func TestNextID_MissingDirs(t *testing.T) {
	dir := t.TempDir()
	got, err := NextID(filepath.Join(dir, "nope"), filepath.Join(dir, "nope2"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "000001" {
		t.Errorf("NextID = %q, want 000001 for empty dirs", got)
	}
}

// TestRender_ByteStable pins the exact blank-issue output so deriving the
// template from the cue model (#145) provably preserves bytes.
func TestRender_ByteStable(t *testing.T) {
	got := Render(ScaffoldSpec{ID: "000057", Title: "Some new thing", Today: "2026-05-31"})
	want := `---
id: 000057
status: open
deps: []
github_issue:
created: 2026-05-31
updated: 2026-05-31
estimate_hours:
---

# Some new thing

## Problem

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-05-31
`
	if got != want {
		t.Errorf("Render byte-drift:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

// TestRender_ByteStable_FromGitHub locks the --from-github path too, so the
// Problem-body special-case is provably byte-preserved by the refactor.
func TestRender_ByteStable_FromGitHub(t *testing.T) {
	got := Render(ScaffoldSpec{
		ID: "000057", Title: "Imported", Today: "2026-05-31",
		GithubIssue: "42", ProblemBody: "The GH issue body.\n\nMore detail.",
	})
	want := `---
id: 000057
status: open
deps: []
github_issue: 42
created: 2026-05-31
updated: 2026-05-31
estimate_hours:
---

# Imported

## Problem

The GH issue body.

More detail.

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-05-31
`
	if got != want {
		t.Errorf("Render byte-drift (from-github):\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

// TestRender_DrivenByModel asserts Render emits exactly the model's sections in
// the model's order — so a cue section add/rename/reorder propagates (#145).
func TestRender_DrivenByModel(t *testing.T) {
	out := Render(ScaffoldSpec{ID: "000057", Title: "x", Today: "2026-05-31"})
	last := -1
	for _, sec := range vocab.Issue().Sections() {
		idx := strings.Index(out, "## "+sec.Name+"\n")
		if idx < 0 {
			t.Errorf("Render output missing section %q", sec.Name)
			continue
		}
		if idx < last {
			t.Errorf("section %q out of model order", sec.Name)
		}
		last = idx
	}
}

// TestRender_Blank: a blank `issue new` yields every canonical field +
// section, with Problem/Spec present-but-empty (the regression the
// structural check is most sensitive to) and an empty github_issue.
func TestRender_Blank(t *testing.T) {
	out := Render(ScaffoldSpec{ID: "000057", Title: "Some new thing", Today: "2026-05-31"})

	// Round-trips through the frontmatter parser, with canonical fields.
	fm, body, err := Parse(out)
	if err != nil {
		t.Fatalf("rendered issue does not parse: %v", err)
	}
	for _, want := range []struct{ k, v string }{
		{"id", "000057"}, {"status", "open"}, {"deps", "[]"},
		{"github_issue", ""}, {"created", "2026-05-31"}, {"updated", "2026-05-31"},
	} {
		if got, _ := GetField(fm, want.k); got != want.v {
			t.Errorf("frontmatter %s = %q, want %q", want.k, got, want.v)
		}
	}
	if _, ok := GetField(fm, "estimate_hours"); !ok {
		t.Error("estimate_hours field missing")
	}
	for _, sec := range []string{"# Some new thing", "## Problem", "## Spec", "## Done when", "## Plan", "## Log"} {
		if !strings.Contains(body, sec) {
			t.Errorf("body missing %q\n---\n%s", sec, body)
		}
	}
	if !strings.Contains(body, "- [ ]") {
		t.Errorf("body missing empty Plan item:\n%s", body)
	}
	if strings.Contains(out, "actual_hours") || strings.Contains(body, "## Side quests") {
		t.Errorf("blank skeleton should not seed actual_hours / Side quests:\n%s", out)
	}
}

// TestRender_FromGitHub: --from-github sets github_issue and seeds the GH
// body under ## Problem (not as a loose post-title paragraph).
func TestRender_FromGitHub(t *testing.T) {
	out := Render(ScaffoldSpec{
		ID: "000057", Title: "Imported", Today: "2026-05-31",
		GithubIssue: "42", ProblemBody: "The GH issue body.\n\nMore detail.",
	})
	fm, body, err := Parse(out)
	if err != nil {
		t.Fatalf("does not parse: %v", err)
	}
	if got, _ := GetField(fm, "github_issue"); got != "42" {
		t.Errorf("github_issue = %q, want 42", got)
	}
	probIdx := strings.Index(body, "## Problem")
	specIdx := strings.Index(body, "## Spec")
	bodyIdx := strings.Index(body, "The GH issue body.")
	if bodyIdx < probIdx || bodyIdx > specIdx {
		t.Errorf("GH body should sit between ## Problem and ## Spec:\n%s", body)
	}
}

// TestRender_TargetAndDeps: optional fields render only when set.
func TestRender_TargetAndDeps(t *testing.T) {
	out := Render(ScaffoldSpec{ID: "000057", Title: "x", Today: "2026-05-31", Target: "my-target", Deps: []string{"repo#1", "repo#2"}})
	if !strings.Contains(out, "target: my-target") {
		t.Errorf("target line missing:\n%s", out)
	}
	if !strings.Contains(out, "deps: [repo#1, repo#2]") {
		t.Errorf("deps not rendered as list:\n%s", out)
	}
	bare := Render(ScaffoldSpec{ID: "000057", Title: "x", Today: "2026-05-31"})
	if strings.Contains(bare, "target:") {
		t.Errorf("no target line expected when unset:\n%s", bare)
	}
}

// TestScaffold_SpecialSectionsPresent guards the name-coupling: Render injects
// the --from-github body into "Problem" and the dated heading into "Log" by name.
// If the cue model renames either, this trips before a silent creation regression.
func TestScaffold_SpecialSectionsPresent(t *testing.T) {
	have := map[string]bool{}
	for _, sec := range vocab.Issue().Sections() {
		have[sec.Name] = true
	}
	for _, name := range []string{"Problem", "Log"} {
		if !have[name] {
			t.Errorf("scaffold model missing %q — Render special-cases it by name; update Render's switch if renamed", name)
		}
	}
}

func mustMkdir(t *testing.T, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func writeFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
