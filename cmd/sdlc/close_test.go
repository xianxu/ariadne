package main

import (
	"os"
	"os/exec"
	"path/filepath"
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

// TestMilestonePlanRE_Enumerates verifies the plan-section milestone
// regex picks up the tags whether or not the milestone label is
// emphasized with `**`. Drives findMilestonesMissingVerdict via the
// shared milestonePlanRE.
func TestMilestonePlanRE_Enumerates(t *testing.T) {
	body := `## Plan

- [x] **M1 — first**
- [x] **M2 — second**
- [ ] M3 — unticked, no emphasis
- [.] **M4b — wip with suffix**
- [ ] **M10 — multi-digit**

## Log

- 2026-05-26: closed
`
	planM := issue.PlanSectionRE.FindStringSubmatchIndex(body)
	if planM == nil {
		t.Fatal("plan section not found")
	}
	planBody := body[planM[2]:planM[3]]
	matches := milestonePlanRE.FindAllStringSubmatch(planBody, -1)
	got := make([]string, 0, len(matches))
	for _, m := range matches {
		got = append(got, m[1])
	}
	want := []string{"M1", "M2", "M3", "M4b", "M10"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("milestones = %v, want %v", got, want)
	}
}

// TestFindMilestonesMissingVerdict_AllPresent confirms the helper
// returns no missing milestones when every plan entry has a matching
// close commit carrying the Review-Verdict trailer. Driven via a fake
// git runner — milestoneHasVerdictCommit's exec.Command is too far
// inside gitx to stub here, so we cover the plan-parse / iteration
// shape with a body that has zero milestones (vacuous truth case).
func TestFindMilestonesMissingVerdict_NoMilestones(t *testing.T) {
	body := `## Plan

Just prose, no milestone bullets.

- [ ] some non-milestone task

## Log
`
	missing, err := findMilestonesMissingVerdict(body, "31", "workshop/issues/000031-x.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing milestones, got %v", missing)
	}
}

// TestFindMilestonesMissingVerdict_NoPlanSection — vacuous-truth case
// for issues that never grew a Plan section.
func TestFindMilestonesMissingVerdict_NoPlanSection(t *testing.T) {
	body := "# title\n\nNo plan here.\n"
	missing, err := findMilestonesMissingVerdict(body, "31", "workshop/issues/000031-x.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing milestones, got %v", missing)
	}
}

// TestFindMilestonesMissingVerdict_Integration drives the full check
// path against a real temp git repo so the `git log --grep ...
// --all-match` semantics aren't faked. We create three milestone-close
// commits — M1 with trailer, M2 without trailer, M3 missing entirely —
// and assert the verifier reports {M2, M3} as the missing set.
func TestFindMilestonesMissingVerdict_Integration(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Minimal repo with a deterministic identity (tests run in any env).
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}
	runGit("init", "-q", "-b", "main")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test")
	runGit("config", "commit.gpgsign", "false")

	issuesDir := "workshop/issues"
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	issuePath := filepath.Join(issuesDir, "000031-x.md")

	writeAndCommit := func(content, subject, body string) {
		if err := os.WriteFile(issuePath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		runGit("add", issuePath)
		args := []string{"commit", "-q", "-m", subject}
		if body != "" {
			args = append(args, "-m", body)
		}
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}

	// M1: close commit WITH trailer — should be detected as present.
	writeAndCommit("v1", "#31 M1: close — tick milestone",
		"Body explaining M1 work.\n\nReview-Verdict: SHIP\nReview-Window: abc1234..HEAD")

	// M2: close commit WITHOUT trailer — should be detected as missing.
	writeAndCommit("v2", "#31 M2: close — tick milestone", "Body explaining M2 work, no trailer here.")

	// M3 has no commit at all — should also be detected as missing.

	planBody := `## Plan

- [x] **M1 — first**
- [x] **M2 — second**
- [x] **M3 — third**

## Log
`
	missing, err := findMilestonesMissingVerdict(planBody, "31", issuePath)
	if err != nil {
		t.Fatalf("findMilestonesMissingVerdict: %v", err)
	}
	want := []string{"M2", "M3"}
	if strings.Join(missing, ",") != strings.Join(want, ",") {
		t.Errorf("missing = %v, want %v", missing, want)
	}
}

// TestFindMilestonesMissingVerdict_AllPresent — same temp-repo posture,
// but every milestone gets a close commit with the trailer. Expectation:
// empty missing slice.
func TestFindMilestonesMissingVerdict_AllPresent(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}
	runGit("init", "-q", "-b", "main")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test")
	runGit("config", "commit.gpgsign", "false")

	issuesDir := "workshop/issues"
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	issuePath := filepath.Join(issuesDir, "000031-x.md")

	commit := func(content, milestone, verdict string) {
		if err := os.WriteFile(issuePath, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		runGit("add", issuePath)
		subject := "#31 " + milestone + ": close — tick milestone"
		body := "Body.\n\nReview-Verdict: " + verdict + "\nReview-Window: abc1234..HEAD"
		cmd := exec.Command("git", "commit", "-q", "-m", subject, "-m", body)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit: %v — %s", err, out)
		}
	}
	commit("v1", "M1", "SHIP")
	commit("v2", "M2", "FIX-THEN-SHIP")

	planBody := `## Plan

- [x] **M1 — first**
- [x] **M2 — second**

## Log
`
	missing, err := findMilestonesMissingVerdict(planBody, "31", issuePath)
	if err != nil {
		t.Fatalf("findMilestonesMissingVerdict: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing milestones, got %v", missing)
	}
}

// TestFormatMissingVerdicts_ContractElements verifies the next-action
// error message names the missing milestones, suggests the rerun
// command for each, shows the trailer shape, and documents --force.
// This is a CLI contract (operator pastes the suggested commands), so
// drift here breaks downstream agent workflows.
func TestFormatMissingVerdicts_ContractElements(t *testing.T) {
	msg := formatMissingVerdicts("31", []string{"M2", "M4"})
	want := []string{
		"milestones M2, M4 lack Review-Verdict trailer",
		"sdlc judge milestone-review --issue 31 --milestone M2",
		"sdlc judge milestone-review --issue 31 --milestone M4",
		"Review-Verdict: SHIP",
		"Review-Window:",
		"--force",
	}
	for _, w := range want {
		if !strings.Contains(msg, w) {
			t.Errorf("formatMissingVerdicts missing %q in:\n%s", w, msg)
		}
	}
}
