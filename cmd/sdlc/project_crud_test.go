package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunProjectNewCreatesSelfValidatingScaffold(t *testing.T) {
	dir := t.TempDir()
	origToday, origValidate := projectTodayFn, validateFrontmatterFn
	projectTodayFn = func() string { return "2026-07-16" }
	var validatedNoun, validatedPath string
	validateFrontmatterFn = func(noun, path string) (string, bool, error) {
		validatedNoun, validatedPath = noun, path
		return "", true, nil
	}
	t.Cleanup(func() { projectTodayFn, validateFrontmatterFn = origToday, origValidate })

	f := &projectNewFlags{Slug: "demo", Goal: "Make projects computable.", DoneWhen: "The board is trustworthy.", ProjectsDir: dir}
	var out, errOut bytes.Buffer
	if err := runProjectNew(&out, &errOut, f); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "demo.md")
	if strings.TrimSpace(out.String()) != want {
		t.Fatalf("stdout = %q, want %q", out.String(), want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Fatal(err)
	}
	if err := runProjectValidate(&out, &errOut, &projectValidateFlags{}, []string{want}); err != nil {
		t.Fatal(err)
	}
	if validatedNoun != "project" || validatedPath != want {
		t.Fatalf("validated %q:%q", validatedNoun, validatedPath)
	}
	if err := runProjectNew(&out, &errOut, f); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second create error = %v", err)
	}
}

func TestProjectScaffoldConformsToProjectVocabularyProcess(t *testing.T) {
	dir := t.TempDir()
	f := &projectNewFlags{Slug: "demo", Goal: "Make projects computable.", DoneWhen: "The board is trustworthy.", ProjectsDir: dir}
	if err := runProjectNew(&bytes.Buffer{}, &bytes.Buffer{}, f); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "run", "../vocabulary", "validate-instance", "--type", "project", filepath.Join(dir, "demo.md"))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated scaffold is nonconforming: %v\n%s", err, out)
	}
}

func TestProjectScaffoldPreservesYAMLHostileScalars(t *testing.T) {
	dir := t.TempDir()
	goal := "Ship: safely #1 \"yes\"\nwithout corruption"
	doneWhen := "2026-08-01 is true: yes"
	f := &projectNewFlags{Slug: "demo", Goal: goal, DoneWhen: doneWhen, ProjectsDir: dir}
	if err := runProjectNew(&bytes.Buffer{}, &bytes.Buffer{}, f); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "demo.md")
	cmd := exec.Command("go", "run", "../vocabulary", "validate-instance", "--type", "project", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("hostile scaffold is nonconforming: %v\n%s", err, out)
	}
	d, err := readProject(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := d.FM("goal"); got != goal {
		t.Fatalf("goal round-trip = %q, want %q", got, goal)
	}
	if got := d.FM("done_when"); got != doneWhen {
		t.Fatalf("done_when round-trip = %q, want %q", got, doneWhen)
	}
}

func TestProjectSlugConsumersRejectTraversal(t *testing.T) {
	dir := t.TempDir()
	slug := "../outside"
	var out, errOut bytes.Buffer
	if err := runProjectNew(&out, &errOut, &projectNewFlags{Slug: slug, Goal: "g", DoneWhen: "d", ProjectsDir: dir}); err == nil || !strings.Contains(err.Error(), "invalid project slug") {
		t.Errorf("new traversal = %v", err)
	}
	if err := runProjectShow(&out, &errOut, &projectShowFlags{Slug: slug, ProjectsDir: dir}); err == nil || !strings.Contains(err.Error(), "invalid project slug") {
		t.Errorf("show traversal = %v", err)
	}
	orig := validateFrontmatterFn
	calls := 0
	validateFrontmatterFn = func(string, string) (string, bool, error) { calls++; return "", true, nil }
	t.Cleanup(func() { validateFrontmatterFn = orig })
	if err := runProjectValidate(&out, &errOut, &projectValidateFlags{Slug: slug, ProjectsDir: dir}, nil); err == nil || !strings.Contains(err.Error(), "invalid project slug") {
		t.Errorf("validate traversal = %v", err)
	}
	if calls != 0 {
		t.Fatalf("validator called %d times for traversal", calls)
	}
	if err := runProjectSetStatus(&out, &errOut, &projectSetStatusFlags{Slug: slug, To: "defined", ProjectsDir: dir}); err == nil || !strings.Contains(err.Error(), "invalid project slug") {
		t.Errorf("set-status traversal = %v", err)
	}
}

func TestProjectNewRequiresGoalAndDoneWhenFlags(t *testing.T) {
	cmd := newProjectNewCmd()
	for _, name := range []string{"goal", "done-when"} {
		flag := cmd.Flags().Lookup(name)
		if flag == nil || flag.Annotations[cobra.BashCompOneRequiredFlag] == nil {
			t.Errorf("--%s is not required", name)
		}
	}
}

func TestRunProjectListAndShow(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("alpha", "---\ntype: project\nname: alpha\nstatus: executing\ndeadline: 2026-09-01\n---\n## Breakdown\n- [x] done [ariadne#1]\n- [ ] todo [ariadne#2]\n")
	write("beta", "---\ntype: project\nname: beta\nstatus: ideation\n---\n## Breakdown\n")
	var out, errOut bytes.Buffer
	if err := runProjectList(&out, &errOut, &projectListFlags{ProjectsDir: dir}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"alpha  executing  2026-09-01", "beta  ideation  -"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("list missing %q:\n%s", want, out.String())
		}
	}
	out.Reset()
	if err := runProjectShow(&out, &errOut, &projectShowFlags{ProjectsDir: dir, Slug: "alpha"}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"alpha.md", "status: executing", "tasks: 1/2 done"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("show missing %q:\n%s", want, out.String())
		}
	}
}
