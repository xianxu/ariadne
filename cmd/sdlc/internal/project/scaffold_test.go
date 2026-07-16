package project

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestRenderScaffoldDerivesProjectModel(t *testing.T) {
	got := RenderScaffold(ScaffoldSpec{
		Name: "demo", Goal: "Make projects computable.",
		DoneWhen: "The derived board is trustworthy.", Today: "2026-07-16",
	})
	for _, want := range []string{
		"type: project",
		"status: " + vocab.Project().InitialStatus(),
		"created: 2026-07-16", "updated: 2026-07-16",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("scaffold missing %q:\n%s", want, got)
		}
	}
	for _, section := range vocab.Project().Sections() {
		if !strings.Contains(got, "## "+section.Name) {
			t.Errorf("scaffold missing model section %q", section.Name)
		}
		if section.Seed != "" && !strings.Contains(got, section.Seed) {
			t.Errorf("scaffold missing seed %q", section.Seed)
		}
	}
	d, err := ParseDoc(got)
	if err != nil {
		t.Fatal(err)
	}
	if d.FM("name") != "demo" {
		t.Fatalf("parsed name = %q", d.FM("name"))
	}
	if d.FM("goal") != "Make projects computable." || d.FM("done_when") != "The derived board is trustworthy." {
		t.Fatalf("quoted fields did not round-trip: goal=%q done_when=%q", d.FM("goal"), d.FM("done_when"))
	}
}

func TestSummarizeAndRender(t *testing.T) {
	d, err := ParseDoc(`---
type: project
name: alpha
status: executing
deadline: 2026-09-01
---
## Breakdown
- [x] done [ariadne#1]
- [ ] todo [ariadne#2]
`)
	if err != nil {
		t.Fatal(err)
	}
	s := Summarize("workshop/projects/alpha.md", d)
	if s.Name != "alpha" || s.Done != 1 || s.Total != 2 {
		t.Fatalf("summary = %+v", s)
	}
	if got := RenderListRow(s); got != "alpha  executing  2026-09-01\n" {
		t.Fatalf("list row = %q", got)
	}
	show := RenderShow(s)
	for _, want := range []string{"workshop/projects/alpha.md", "status: executing", "tasks: 1/2 done"} {
		if !strings.Contains(show, want) {
			t.Errorf("show missing %q:\n%s", want, show)
		}
	}
}

func TestResolvePathRejectsTraversal(t *testing.T) {
	if got, err := ResolvePath("workshop/projects", "demo"); err != nil || got != "workshop/projects/demo.md" {
		t.Fatalf("ResolvePath valid = %q, %v", got, err)
	}
	for _, slug := range []string{"../history/projects/demo", "nested/demo", ".", "demo.md", "Demo Space"} {
		if _, err := ResolvePath("workshop/projects", slug); err == nil {
			t.Errorf("ResolvePath accepted %q", slug)
		}
	}
}
