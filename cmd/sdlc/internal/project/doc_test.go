package project

import (
	"strings"
	"testing"
)

const projectDocFixture = `---
type: project
name: demo
status: executing
deadline: 2026-09-01
---
# demo

## PRD
Ship the project noun.

## Estimate

## Breakdown

- [ ] provider interface skeleton [charon#13 M1]
- [x] finished work [ariadne#180 M1]
- [.] active work [ariadne#180 M2]
- [-] blocked work [nous#4]
- [~] cancelled work
- [ ] plain-text task

## Log
### 2026-07-16
Started.
`

func TestParseDoc(t *testing.T) {
	d, err := ParseDoc(projectDocFixture)
	if err != nil {
		t.Fatal(err)
	}
	if got := d.FM("status"); got != "executing" {
		t.Fatalf("FM(status) = %q, want executing", got)
	}
	if got := len(d.Tasks); got != 6 {
		t.Fatalf("len(Tasks) = %d, want 6", got)
	}
	want := Task{
		LineIdx: 9,
		State:   ' ',
		Title:   "provider interface skeleton",
		RefText: "charon#13 M1",
	}
	if got := d.Tasks[0]; got != want {
		t.Fatalf("Tasks[0] = %+v, want %+v", got, want)
	}
	if got := d.Tasks[4]; got.State != '~' || got.Title != "cancelled work" || got.RefText != "" {
		t.Fatalf("Tasks[4] = %+v, want cancelled plain task", got)
	}
	if got := d.SectionBody("PRD"); got != "Ship the project noun.\n" {
		t.Fatalf("SectionBody(PRD) = %q", got)
	}
	if got := d.SectionBody("Missing"); got != "" {
		t.Fatalf("SectionBody(Missing) = %q, want empty", got)
	}
}

func TestDocRenderRoundTrip(t *testing.T) {
	d, err := ParseDoc(projectDocFixture)
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Render(); got != projectDocFixture {
		t.Fatalf("Render changed an unmodified document:\n--- got ---\n%s\n--- want ---\n%s", got, projectDocFixture)
	}
}

func TestDocSetTaskState(t *testing.T) {
	d, err := ParseDoc(projectDocFixture)
	if err != nil {
		t.Fatal(err)
	}
	d.SetTaskState(0, 'x')

	want := strings.Replace(
		projectDocFixture,
		"- [ ] provider interface skeleton [charon#13 M1]",
		"- [x] provider interface skeleton [charon#13 M1]",
		1,
	)
	if got := d.Render(); got != want {
		t.Fatalf("SetTaskState changed more than the selected row:\n%s", got)
	}
	if got := d.Tasks[0].State; got != 'x' {
		t.Fatalf("Tasks[0].State = %q, want x", got)
	}
}

func TestDocSetFM(t *testing.T) {
	d, err := ParseDoc(projectDocFixture)
	if err != nil {
		t.Fatal(err)
	}
	d.SetFM("status", "paused")
	d.SetFM("updated", "2026-07-17")

	if got := d.FM("status"); got != "paused" {
		t.Fatalf("FM(status) = %q, want paused", got)
	}
	if got := d.FM("updated"); got != "2026-07-17" {
		t.Fatalf("FM(updated) = %q", got)
	}
	if got := d.Render(); !strings.Contains(got, "status: paused\ndeadline: 2026-09-01\nupdated: 2026-07-17\n---") {
		t.Fatalf("SetFM did not preserve field order/append semantics:\n%s", got)
	}
}

func TestDocAppendToSection(t *testing.T) {
	d, err := ParseDoc(projectDocFixture)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AppendToSection("PRD", "Second requirement.\n"); err != nil {
		t.Fatal(err)
	}

	want := "## PRD\nShip the project noun.\n\nSecond requirement.\n\n## Estimate"
	if got := d.Render(); !strings.Contains(got, want) {
		t.Fatalf("AppendToSection placed block incorrectly:\n%s", got)
	}
	if got := d.SectionBody("PRD"); got != "Ship the project noun.\n\nSecond requirement.\n" {
		t.Fatalf("SectionBody(PRD) after append = %q", got)
	}
	if got := d.Tasks[0].Title; got != "provider interface skeleton" {
		t.Fatalf("task index was not rebuilt after append: %+v", d.Tasks[0])
	}
}
