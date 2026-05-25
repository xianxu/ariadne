package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// TestMilestoneTickRegex mirrors the regex in runClose's milestone path:
//
//	(?m)^(- )\[[ .]\]( <milestone>\b)
//
// Verifies the boundary semantics — M4 must not match M4-extra or M40, [x]
// must not be re-ticked.
func TestMilestoneTickRegex(t *testing.T) {
	pat := regexp.MustCompile(`(?m)^(- )\[[ .]\]( M4\b)`)
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{
			"unchecked",
			"- [ ] M4 — port close-issue\n",
			"- [x] M4 — port close-issue\n",
		},
		{
			"in_progress",
			"- [.] M4 — port close-issue\n",
			"- [x] M4 — port close-issue\n",
		},
		{
			"already_done_unchanged",
			"- [x] M4 — port close-issue\n",
			"- [x] M4 — port close-issue\n",
		},
		{
			// NOTE: this matches because Python's \b (= Go RE2's \b) treats
			// the boundary between '4' (\w) and '-' (\W) as a word break.
			// close-issue.py has the same behavior — verified with python3
			// against the same input. Preserving Python parity over the
			// "M4-extra should not match" intuition.
			"M4_extra_DOES_match_via_word_boundary",
			"- [ ] M4-extra — different milestone\n",
			"- [x] M4-extra — different milestone\n",
		},
		{
			"M40_no_match",
			"- [ ] M40 — different milestone\n",
			"- [ ] M40 — different milestone\n",
		},
		{
			"M4b_no_match_for_M4",
			"- [ ] M4b — different milestone\n",
			"- [ ] M4b — different milestone\n",
		},
		{
			"mixed_plan",
			"- [ ] M3 — earlier\n- [ ] M4 — target\n- [ ] M5 — later\n",
			"- [ ] M3 — earlier\n- [x] M4 — target\n- [ ] M5 — later\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pat.ReplaceAllString(tt.in, "${1}[x]${2}")
			if got != tt.out {
				t.Errorf("got %q want %q", got, tt.out)
			}
		})
	}
}

// TestPlanUncheckedDetection mirrors the issue-close plan-scan: count items
// shaped like "- [ ] ..." or "- [.] ..." inside the ## Plan section.
func TestPlanUncheckedDetection(t *testing.T) {
	body := `## Spec

random spec stuff here.
- [ ] this is in Spec, must not be counted

## Plan

- [ ] M1 do thing
- [x] M2 done thing
- [.] M3 in progress
- [ ] M4 another

## Log

- [ ] this is in Log section, must not be counted
`
	planRE := regexp.MustCompile(`(?ms)^## Plan\s*\n(.*?)(?:^## |\z)`)
	m := planRE.FindStringSubmatchIndex(body)
	if m == nil {
		t.Fatal("Plan section not found")
	}
	planBody := body[m[2]:m[3]]
	uncheckedRE := regexp.MustCompile(`(?m)^- \[[ .]\] .*$`)
	unchecked := uncheckedRE.FindAllString(planBody, -1)
	if len(unchecked) != 3 {
		t.Errorf("unchecked count = %d, want 3; matches: %v", len(unchecked), unchecked)
	}
}

// TestPlanUncheckedDetection_PlanIsLastSection — when Plan is the trailing
// section (no following ##), the regex still captures the plan body via \z.
func TestPlanUncheckedDetection_PlanIsLastSection(t *testing.T) {
	body := "## Plan\n\n- [ ] M1 do thing\n- [.] M2 wip\n"
	planRE := regexp.MustCompile(`(?ms)^## Plan\s*\n(.*?)(?:^## |\z)`)
	m := planRE.FindStringSubmatchIndex(body)
	if m == nil {
		t.Fatal("Plan section not found")
	}
	planBody := body[m[2]:m[3]]
	uncheckedRE := regexp.MustCompile(`(?m)^- \[[ .]\] .*$`)
	unchecked := uncheckedRE.FindAllString(planBody, -1)
	if len(unchecked) != 2 {
		t.Errorf("unchecked count = %d, want 2; matches: %v", len(unchecked), unchecked)
	}
}

func TestInsertLogLine_ExistingSection(t *testing.T) {
	body := "# title\n\n## Plan\n\n- [x] M1 done\n\n## Log\n\n- 2026-05-01: started\n"
	got := insertLogLine(body, "- 2026-05-25: closed — tests pass")
	// Mirrors Python: re.sub greedy \s* consumes the trailing blank into
	// group 1, then the replacement adds another \n, producing three
	// newlines between '## Log' and the new entry. The existing entry
	// lands on the very next line (no blank).
	want := "# title\n\n## Plan\n\n- [x] M1 done\n\n## Log\n\n\n- 2026-05-25: closed — tests pass\n- 2026-05-01: started\n"
	if got != want {
		t.Errorf("insertLogLine mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestInsertLogLine_EmptyLogSection(t *testing.T) {
	body := "# title\n\n## Plan\n\n- [x] M1 done\n\n## Log\n"
	got := insertLogLine(body, "- 2026-05-25: closed — done")
	if !strings.Contains(got, "## Log\n\n- 2026-05-25: closed — done\n") {
		t.Errorf("expected log line inserted under header:\n%s", got)
	}
}

func TestInsertLogLine_NoLogSection(t *testing.T) {
	body := "# title\n\n## Plan\n\n- [x] M1 done\n"
	got := insertLogLine(body, "- 2026-05-25: closed — done")
	want := "# title\n\n## Plan\n\n- [x] M1 done\n\n## Log\n\n- 2026-05-25: closed — done\n"
	if got != want {
		t.Errorf("insertLogLine no-section mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// TestFrontmatterChainForIssueClose verifies the upsert chain runClose
// applies for an issue close: status → done, actual_hours → ACTUAL, updated → today.
func TestFrontmatterChainForIssueClose(t *testing.T) {
	doc := "---\nid: 000031\nstatus: working\nestimate_hours: 4\nactual_hours:\n---\n# title\n\nbody\n"
	fm, body, err := issue.Parse(doc)
	if err != nil {
		t.Fatal(err)
	}
	fm = issue.SetField(fm, "status", "done")
	fm = issue.SetField(fm, "actual_hours", "6.5")
	fm = issue.SetField(fm, "updated", "2026-05-25")
	out := issue.Compose(fm, body)
	wantFM := "id: 000031\nstatus: done\nestimate_hours: 4\nactual_hours: 6.5\nupdated: 2026-05-25"
	if !strings.Contains(out, wantFM) {
		t.Errorf("expected frontmatter ordered as:\n%s\ngot:\n%s", wantFM, out)
	}
}

// TestFrontmatterAppend_FieldAbsent verifies SetField appends when the
// field is missing. Mirrors the "actual_hours: (absent)" → "actual_hours: 4" path.
func TestFrontmatterAppend_FieldAbsent(t *testing.T) {
	fm := "id: 000031\nstatus: working\nestimate_hours: 4"
	got := issue.SetField(fm, "actual_hours", "6.5")
	want := "id: 000031\nstatus: working\nestimate_hours: 4\nactual_hours: 6.5"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
