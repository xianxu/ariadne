package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// TestCloseVerb: the mode→verb mapping is single-sourced (#146). A milestone tag
// selects `sdlc milestone-close`; empty selects the whole-issue `sdlc close`.
func TestCloseVerb(t *testing.T) {
	if got := closeVerb(""); got != "sdlc close" {
		t.Errorf(`closeVerb("") = %q, want "sdlc close"`, got)
	}
	if got := closeVerb("M2"); got != "sdlc milestone-close" {
		t.Errorf(`closeVerb("M2") = %q, want "sdlc milestone-close"`, got)
	}
}

// TestRerunCmd: a close gate refusal re-run hint must pick the verb by mode — a
// milestone points at `sdlc milestone-close`, NEVER the removed `close --milestone`
// bypass (#146). The whole-issue form stays `sdlc close`.
func TestRerunCmd(t *testing.T) {
	ms := rerunCmd("31", "M4", " --actual <hours>")
	if !strings.Contains(ms, "sdlc milestone-close --issue 31 --milestone M4") {
		t.Errorf("milestone re-run should use milestone-close; got: %s", ms)
	}
	if strings.Contains(ms, "sdlc close --issue 31 --milestone") {
		t.Errorf("milestone re-run must NOT suggest the removed close --milestone path: %s", ms)
	}
	issueForm := rerunCmd("31", "", " --actual 2.5")
	if !strings.HasPrefix(issueForm, "sdlc close --issue 31 --actual 2.5 --verified") {
		t.Errorf("whole-issue re-run shape wrong: %s", issueForm)
	}
	if strings.Contains(issueForm, "--milestone") {
		t.Errorf("whole-issue re-run must not carry --milestone: %s", issueForm)
	}
}

// ── #67: per-gate --no-<gate> bypass flags ───────────────────────────────────

// TestCloseFlags_Skip pins the skip() contract: --force waives EVERY gate, while
// each --no-<gate> waives ONLY its own. A typo'd field would let one flag leak
// into another gate (or none), which this catches.
func TestCloseFlags_Skip(t *testing.T) {
	gates := []string{"actual", "verified", "reclose", "atlas", "verdict", "plan", "project"}

	// --force ⇒ all gates skipped.
	force := &closeFlags{Force: true}
	for _, g := range gates {
		if !force.skip(g) {
			t.Errorf("--force should skip gate %q", g)
		}
	}

	// Each --no-<gate> ⇒ exactly its own gate, nothing else.
	cases := []struct {
		gate string
		set  func(*closeFlags)
	}{
		{"actual", func(f *closeFlags) { f.NoActual = true }},
		{"verified", func(f *closeFlags) { f.NoVerified = true }},
		{"reclose", func(f *closeFlags) { f.NoReclose = true }},
		{"atlas", func(f *closeFlags) { f.NoAtlas = true }},
		{"verdict", func(f *closeFlags) { f.NoVerdict = true }},
		{"plan", func(f *closeFlags) { f.NoPlanCheck = true }},
		{"project", func(f *closeFlags) { f.NoProject = true }},
	}
	for _, c := range cases {
		f := &closeFlags{}
		c.set(f)
		for _, g := range gates {
			got := f.skip(g)
			want := g == c.gate
			if got != want {
				t.Errorf("with --no-%s set: skip(%q) = %v, want %v (per-gate isolation broken)", c.gate, g, got, want)
			}
		}
	}

	// A clean flag set skips nothing.
	none := &closeFlags{}
	for _, g := range gates {
		if none.skip(g) {
			t.Errorf("no bypass flag set: skip(%q) should be false", g)
		}
	}
	// Unknown gate name is never skipped (default arm).
	if (&closeFlags{Force: false}).skip("bogus") {
		t.Error("unknown gate should not be skipped without --force")
	}
}

// TestCloseCmd_Registered asserts the per-gate flags (and --force) are wired
// onto the command, mirroring TestMergeCmd_Registered.
func TestCloseCmd_Registered(t *testing.T) {
	cmd := NewCloseCmd()
	for _, flag := range []string{
		"force", "no-actual", "no-verified", "no-reclose-guard",
		"no-atlas", "no-verdict", "no-plan-check", "no-project",
	} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("close command missing flag: --%s", flag)
		}
	}
}

// TestClose_MilestoneRefusesWithRedirect: the close command still parses the
// (now hidden, deprecated) --milestone flag, but refuses it with a redirect to
// milestone-close rather than silently doing a no-review milestone close (#146).
func TestClose_MilestoneRefusesWithRedirect(t *testing.T) {
	cmd := NewCloseCmd()
	f := cmd.Flags().Lookup("milestone")
	if f == nil {
		t.Fatal("--milestone should still parse (hidden), to give a friendly refusal not `unknown flag`")
	}
	if !f.Hidden {
		t.Error("--milestone should be hidden from `close --help` (#146)")
	}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--issue", "31", "--milestone", "M4", "--actual", "1", "--verified", "x"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("close --milestone should refuse")
	}
	if !strings.Contains(err.Error(), "milestone-close") {
		t.Errorf("refusal should redirect to milestone-close; got: %v", err)
	}
}

// TestMilestoneCloseCmd_RegistersBypasses asserts the per-gate flags are also
// exposed on milestone-close (it forwards them into runClose).
func TestMilestoneCloseCmd_RegistersBypasses(t *testing.T) {
	cmd := NewMilestoneCloseCmd()
	for _, flag := range []string{
		"no-actual", "no-verified", "no-reclose-guard",
		"no-atlas", "no-verdict", "no-plan-check", "no-project",
	} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("milestone-close command missing flag: --%s", flag)
		}
	}
}

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

// #66: a dated log line whose date matches an existing `### <date>` day header
// is filed directly under that header (top of the day's group), not orphaned
// above it at the top of the ## Log section.
func TestInsertLogLine_UnderMatchingDayHeader(t *testing.T) {
	body := "# title\n\n## Log\n\n### 2026-05-25\n- Implemented the thing\n"
	got := insertLogLine(body, "- 2026-05-25: closed — tests pass")
	want := "# title\n\n## Log\n\n### 2026-05-25\n- 2026-05-25: closed — tests pass\n- Implemented the thing\n"
	if got != want {
		t.Errorf("insertLogLine day-header mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

// #66: when the day header's date does NOT match the log line's date, fall back
// to top-of-section (don't misfile a 06-02 line under a 05-25 header).
func TestInsertLogLine_DayHeaderDateMismatch_FallsBack(t *testing.T) {
	body := "# title\n\n## Log\n\n### 2026-05-25\n- old work\n"
	got := insertLogLine(body, "- 2026-06-02: closed — done")
	// No `### 2026-06-02` header → original top-of-section behavior; the line
	// must NOT land under the 05-25 header.
	if strings.Contains(got, "### 2026-05-25\n- 2026-06-02: closed") {
		t.Errorf("06-02 line misfiled under 05-25 header:\n%s", got)
	}
	if !strings.Contains(got, "## Log\n\n\n- 2026-06-02: closed — done\n### 2026-05-25\n") {
		t.Errorf("expected fallback top-of-section insert:\n%q", got)
	}
}

// #73: day headers routinely carry a suffix (`### <date> — session summary`);
// the line must still file directly under such a header, not orphan at the top.
func TestInsertLogLine_UnderSuffixedDayHeader(t *testing.T) {
	body := "# t\n\n## Log\n\n### 2026-05-25 — closeout\n- Revisited\n"
	got := insertLogLine(body, "- 2026-05-25: closed — done")
	want := "# t\n\n## Log\n\n### 2026-05-25 — closeout\n- 2026-05-25: closed — done\n- Revisited\n"
	if got != want {
		t.Errorf("line not filed under suffixed day header:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

// #66 (found by dogfooding): a meta-issue can quote `## Log` and `### <date>`
// inside an earlier section (e.g. a fenced code block in ## Problem). The line
// must land in the REAL (last) `## Log` section, not the quoted one.
func TestInsertLogLine_IgnoresEarlierLogHeaderInProse(t *testing.T) {
	body := "# t\n\n## Problem\n\n```\n## Log\n\n### 2026-05-25\n- example\n```\n\n## Log\n\n### 2026-05-25\n- real work\n"
	got := insertLogLine(body, "- 2026-05-25: closed — done")
	// The example block (between the fences) must be untouched.
	want := "# t\n\n## Problem\n\n```\n## Log\n\n### 2026-05-25\n- example\n```\n\n## Log\n\n### 2026-05-25\n- 2026-05-25: closed — done\n- real work\n"
	if got != want {
		t.Errorf("line filed into the quoted Log, not the real one:\n--- got ---\n%q\n--- want ---\n%q", got, want)
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

// TestFrontmatterChainForIssueClose verifies the upsert chain runClose applies for
// an issue close: status → codecomplete (#160), actual_hours → ACTUAL, updated → today.
func TestFrontmatterChainForIssueClose(t *testing.T) {
	doc := "---\nid: 000031\nstatus: working\nestimate_hours: 4\nactual_hours:\n---\n# title\n\nbody\n"
	fm, body, err := issue.Parse(doc)
	if err != nil {
		t.Fatal(err)
	}
	fm = issue.SetField(fm, "status", "codecomplete")
	fm = issue.SetField(fm, "actual_hours", "6.5")
	fm = issue.SetField(fm, "updated", "2026-05-25")
	out := issue.Compose(fm, body)
	wantFM := "id: 000031\nstatus: codecomplete\nestimate_hours: 4\nactual_hours: 6.5\nupdated: 2026-05-25"
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

func TestRunClose_NoActualWritesNotApplicableSentinel(t *testing.T) {
	repoRoot, err := gitx.RepoTopLevel()
	if err != nil {
		t.Fatal(err)
	}
	issuesDir := closeRepo(t, 135)
	f := &closeFlags{
		Issue:     135,
		NoActual:  true,
		Verified:  "administrative close; no measured actual applies",
		NoAtlas:   true,
		IssuesDir: issuesDir,
		BrainDir:  "../nonexistent-brain",
	}
	if err := runClose(io.Discard, io.Discard, f); err != nil {
		t.Fatalf("runClose: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(issuesDir, "000135-x.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "actual_hours: "+issue.ActualNotApplicableSentinel) {
		t.Fatalf("missing actual sentinel:\n%s", text)
	}
	if strings.Contains(text, "actual_hours:\n") {
		t.Fatalf("actual_hours remained blank:\n%s", text)
	}
	if _, err := exec.LookPath("cue"); err != nil {
		t.Skip("cue not on PATH")
	}
	fm, _, err := issue.Parse(text)
	if err != nil {
		t.Fatal(err)
	}
	dataPath := filepath.Join(t.TempDir(), "issue.yaml")
	if err := os.WriteFile(dataPath, []byte(fm+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	schemaPath := filepath.Join(repoRoot, "construct", "vocabulary", "issue.cue")
	cmd := exec.Command("cue", "vet", "-d", "#Issue", dataPath, schemaPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("closed issue frontmatter does not conform to #Issue: %v\n%s", err, out)
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

// TestPartitionMissingVerdicts pins the trailing-vs-midstream split (#175):
// a missing milestone BEFORE the last verdict-carrying one is a genuine
// skipped-review violation (midstream); missing milestones after it (or all,
// when none carries a verdict — the single-pass case) are trailing and
// covered by the imminent issue-close review.
func TestPartitionMissingVerdicts(t *testing.T) {
	cases := []struct {
		name               string
		ordered, missing   []string
		wantMid, wantTrail []string
	}{
		{"single-pass: all missing → all trailing",
			[]string{"M1", "M2", "M3"}, []string{"M1", "M2", "M3"},
			nil, []string{"M1", "M2", "M3"}},
		{"midstream: miss before a verdict-carrying row",
			[]string{"M1", "M2", "M3"}, []string{"M2"},
			[]string{"M2"}, nil}, // M3 has a verdict → M2's boundary was crossed unreviewed
		{"mixed: M1 missing before M2's verdict, M3 trailing",
			[]string{"M1", "M2", "M3"}, []string{"M1", "M3"},
			[]string{"M1"}, []string{"M3"}},
		{"none missing",
			[]string{"M1", "M2"}, nil, nil, nil},
		{"only last missing → trailing",
			[]string{"M1", "M2"}, []string{"M2"}, nil, []string{"M2"}},
		// The reopened-issue shape: prior milestones all reviewed, one new
		// trailing Mx added by the reopen — second most likely real-world hit.
		{"reopened issue: new trailing row after all-reviewed history",
			[]string{"M1", "M2", "M3"}, []string{"M3"}, nil, []string{"M3"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mid, trail := partitionMissingVerdicts(tc.ordered, tc.missing)
			if strings.Join(mid, ",") != strings.Join(tc.wantMid, ",") {
				t.Errorf("midstream = %v, want %v", mid, tc.wantMid)
			}
			if strings.Join(trail, ",") != strings.Join(tc.wantTrail, ",") {
				t.Errorf("trailing = %v, want %v", trail, tc.wantTrail)
			}
		})
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
	_, missing, err := findMilestonesMissingVerdict(body, "31", "workshop/issues/000031-x.md")
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
	_, missing, err := findMilestonesMissingVerdict(body, "31", "workshop/issues/000031-x.md")
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
	ordered, missing, err := findMilestonesMissingVerdict(planBody, "31", issuePath)
	if err != nil {
		t.Fatalf("findMilestonesMissingVerdict: %v", err)
	}
	if wantOrdered := []string{"M1", "M2", "M3"}; strings.Join(ordered, ",") != strings.Join(wantOrdered, ",") {
		t.Errorf("ordered = %v, want %v", ordered, wantOrdered)
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
	_, missing, err := findMilestonesMissingVerdict(planBody, "31", issuePath)
	if err != nil {
		t.Fatalf("findMilestonesMissingVerdict: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing milestones, got %v", missing)
	}
}

// TestFindMilestonesMissingVerdict_SpaceBeforeColonSubject is the regression
// for the #56 close bug: a milestone-close subject can read
// `#31 M1 close: …` — the milestone followed by more subject words before the
// colon — not only `#31 M1: …`. The matcher must detect the trailer in both.
func TestFindMilestonesMissingVerdict_SpaceBeforeColonSubject(t *testing.T) {
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
	if err := os.WriteFile(issuePath, []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", issuePath)
	// Milestone followed by a space + more words before the colon.
	runGit("commit", "-q",
		"-m", "#31 M1 close: review SHIP — tidy",
		"-m", "Body.\n\nReview-Verdict: SHIP\nReview-Window: abc1234..HEAD")

	planBody := "## Plan\n\n- [x] **M1 — first**\n\n## Log\n"
	_, missing, err := findMilestonesMissingVerdict(planBody, "31", issuePath)
	if err != nil {
		t.Fatalf("findMilestonesMissingVerdict: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("`#31 M1 close:` subject should satisfy the verdict check; got missing %v", missing)
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
		// #175: the refusal must cite §3's don't-over-split rule and name
		// the sanctioned fold-to-plain-checkboxes recovery.
		"AGENTS.md",
		"§3",
		"plain checkboxes",
		// Pinned by internal/processmanual/gatesig.go (friction attribution).
		"Or pass --no-verdict (or --force); record",
	}
	for _, w := range want {
		if !strings.Contains(msg, w) {
			t.Errorf("formatMissingVerdicts missing %q in:\n%s", w, msg)
		}
	}
}

// TestFormatTrailingVerdictAccepted_ContractElements pins the loud
// acceptance line (#175): names the accepted milestones, says what covers
// them, and hints §3 for next time.
func TestFormatTrailingVerdictAccepted_ContractElements(t *testing.T) {
	msg := formatTrailingVerdictAccepted([]string{"M1", "M2"})
	for _, w := range []string{
		"M1, M2",                      // names the accepted milestones
		"issue-close boundary review", // what covers them
		"#175",                        // provenance
		"plain checkboxes",            // §3 hint for next time
	} {
		if !strings.Contains(msg, w) {
			t.Errorf("formatTrailingVerdictAccepted missing %q in:\n%s", w, msg)
		}
	}
	// The acceptance line is a fresh cinfo output on the close path — it must
	// not match any gatesig ack/refusal classifier pattern (#177 precedent).
	assertNoGatesigCollision(t, msg)
}

// TestFormatTrailingNeedsJudge_ContractElements pins the refusal fired when
// trailing unclosed milestones exist but --no-judge skips the very review
// that would cover them (#175). Shares the gatesig-pinned closing line so
// friction measurement keys on one no-verdict signature.
func TestFormatTrailingNeedsJudge_ContractElements(t *testing.T) {
	msg := formatTrailingNeedsJudge("31", []string{"M1"})
	for _, w := range []string{
		"M1",
		"--no-judge", // names the premise-killer
		"sdlc judge milestone-review --issue 31 --milestone M1",
		"Or pass --no-verdict (or --force); record", // same gatesig signature
	} {
		if !strings.Contains(msg, w) {
			t.Errorf("formatTrailingNeedsJudge missing %q in:\n%s", w, msg)
		}
	}
}

// TestFormatFixThenShipProtocol_ContractElements pins the post-FIX-THEN-SHIP
// next-action block (#174): fix NOW (pre-commit), bundle into ONE commit,
// do NOT re-run close, why (publish-gate anchor), and the post-commit escape
// hatch (re-run the boundary verb). Verb-parameterized so a milestone close
// names `sdlc milestone-close` (same closeVerb threading as the REWORK arm).
func TestFormatFixThenShipProtocol_ContractElements(t *testing.T) {
	msg := formatFixThenShipProtocol("sdlc close")
	for _, w := range []string{
		"FIX-THEN-SHIP",
		"before committing",   // fix NOW, pre-commit
		"ONE commit",          // bundle fixes + close mutations
		"Do NOT re-run",       // the anti-loop instruction
		"publish gate",        // the why (anchor semantics)
		"re-run `sdlc close`", // the post-commit escape hatch
	} {
		if !strings.Contains(msg, w) {
			t.Errorf("formatFixThenShipProtocol missing %q in:\n%s", w, msg)
		}
	}
	// Milestone variant: verb threaded into the anti-loop line, and the
	// escape hatch speaks next-boundary coverage, NOT issue-close anchor
	// semantics (which don't apply — no codecomplete anchor at a milestone).
	ms := formatFixThenShipProtocol("sdlc milestone-close")
	if !strings.Contains(ms, "re-run `sdlc milestone-close`") {
		t.Errorf("milestone verb not threaded into the anti-loop line:\n%s", ms)
	}
	if !strings.Contains(ms, "NEXT boundary review") {
		t.Errorf("milestone escape hatch should point at the next boundary review:\n%s", ms)
	}
	if strings.Contains(ms, "anchor advances") {
		t.Errorf("milestone variant must not claim anchor semantics:\n%s", ms)
	}
	assertNoGatesigCollision(t, msg)
	assertNoGatesigCollision(t, ms)
}
