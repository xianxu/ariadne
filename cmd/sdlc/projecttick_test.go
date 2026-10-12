package main

import (
	"errors"
	"strings"
	"testing"
)

const tickProject = "---\nstatus: executing\n---\n\n# p\n\n## tasks\n\n" +
	"- [ ] [ariadne#42 M1] first\n- [ ] [ariadne#42 M2] second\n- [ ] [ariadne#7] other\n\n## details\n\n" +
	"<a id=\"ariadne-42-m1\"></a>\n### ariadne#42 M1 — first\n\nProse.\n"

func tickInput(milestone string) projectEditInput {
	return projectEditInput{Milestone: milestone, RepoName: "ariadne", IssueStr: "42", Actual: "1.5", Today: "2026-10-12", Est: "2", Label: "ariadne/p.md"}
}

func TestProjectCloseEditIssueTicksEveryRemainingRow(t *testing.T) {
	got, applied, warns, err := projectCloseEdit(tickProject, tickInput(""))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "- [x] [ariadne#42 M1]") || !strings.Contains(got, "- [x] [ariadne#42 M2]") || !strings.Contains(got, "- [ ] [ariadne#7]") {
		t.Fatalf("issue close ticks only its own rows:\n%s", got)
	}
	if len(applied) != 1 || len(warns) != 1 || !strings.Contains(warns[0], "multiple ariadne#42 task rows") {
		t.Fatalf("applied %v warns %v", applied, warns)
	}
}

func TestProjectCloseEditMilestoneTicksAndUpsertsItsBlock(t *testing.T) {
	got, applied, _, err := projectCloseEdit(tickProject, tickInput("M1"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "- [x] [ariadne#42 M1]") || strings.Contains(got, "- [x] [ariadne#42 M2]") {
		t.Fatalf("milestone close ticks its row alone:\n%s", got)
	}
	if !strings.Contains(got, "1.5h") || !strings.Contains(got, "2026-10-12") || len(applied) != 2 {
		t.Fatalf("detail block not upserted (applied %v):\n%s", applied, got)
	}
}

func TestProjectCloseEditMissingBlockRefusesUnlessWaived(t *testing.T) {
	got, _, _, err := projectCloseEdit(tickProject, tickInput("M2"))
	var missing errMissingDetailBlock
	if !errors.As(err, &missing) || !strings.Contains(err.Error(), `<a id="ariadne-42-m2">`) || !strings.Contains(err.Error(), "`--no-project` skips the project update entirely") {
		t.Fatalf("missing block: %v", err)
	}
	if got != tickProject {
		t.Fatal("a refused edit must return the text unchanged")
	}
	in := tickInput("M2")
	in.WaiveDetailBlock = true
	got, _, warns, err := projectCloseEdit(tickProject, in)
	if err != nil || !strings.Contains(got, "- [x] [ariadne#42 M2]") || len(warns) != 1 || !strings.Contains(warns[0], "--force") {
		t.Fatalf("waived: %v %v\n%s", err, warns, got)
	}
}
