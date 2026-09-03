package main

import (
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// planWithFencedHeading is the shape that disarmed the close gates: an issue
// whose Plan quotes markdown — which issues here do constantly, because the
// deliverable often IS a markdown document — with real work after the fence.
const planWithFencedHeading = `---
id: 000999
status: working
---

# demo

## Plan

- [x] M1 — done
- [x] Add the scaffold. Example of what it emits:

` + "```markdown" + `
## Some heading the issue is quoting
` + "```" + `

- [ ] M2 — NOT done
- [ ] Wire the consumer

## Log

### 2026-09-02
`

// TestClosePlanGate_SeesItemsAfterAFencedHeading is #211's reason to exist.
//
// The old `^## ` terminator ended the Plan at the quoted heading, so the
// unchecked-items guard counted ZERO and `sdlc close` would pass an issue with
// two open items. The word-count gates fail safe (truncation only lowers a
// count below a threshold); these two count things whose ABSENCE means pass, so
// truncation flips them the dangerous way.
func TestClosePlanGate_SeesItemsAfterAFencedHeading(t *testing.T) {
	planBody, ok := issue.PlanSectionBody(planWithFencedHeading)
	if !ok {
		t.Fatal("## Plan not found")
	}
	unchecked := issue.PlanUncheckedRE.FindAllString(planBody, -1)
	if len(unchecked) != 2 {
		t.Errorf("close's plan-unchecked guard sees %d open item(s), want 2:\n%s", len(unchecked), planBody)
	}
	// And it must not have run past the Plan into the Log.
	if strings.Contains(planBody, "2026-09-02") {
		t.Error("Plan body leaked into ## Log — the section must still END at the next real heading")
	}
}

// TestMilestoneScan_SeesMilestonesAfterAFencedHeading covers the second gate on
// the same body: a milestone hidden behind a quoted heading is never asked for
// review evidence, so `sdlc close` finalizes without it.
func TestMilestoneScan_SeesMilestonesAfterAFencedHeading(t *testing.T) {
	planBody, ok := issue.PlanSectionBody(planWithFencedHeading)
	if !ok {
		t.Fatal("## Plan not found")
	}
	var got []string
	for _, m := range milestonePlanRE.FindAllStringSubmatch(planBody, -1) {
		got = append(got, m[1])
	}
	want := []string{"M1", "M2"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("findMilestonesMissingVerdict would see %v, want %v — M2's review evidence would never be demanded", got, want)
	}
}

// TestCountPlanItems_SeesItemsAfterAFencedHeading covers the consumer the first
// pass of this issue missed entirely (plan.go:30 → state.go), which is why the
// Spec enumerates call sites by grep rather than from memory.
func TestCountPlanItems_SeesItemsAfterAFencedHeading(t *testing.T) {
	total, ticked := issue.CountPlanItems(planWithFencedHeading)
	if total != 4 || ticked != 2 {
		t.Errorf("CountPlanItems = (%d total, %d ticked), want (4, 2) — `sdlc state` reports progress from this", total, ticked)
	}
}

// issueQuotingItsOwnLogFormat is the shape that made this a LIVE bug rather than
// a latent one: an issue whose Problem section quotes the `## Log` format the
// close verb writes. workshop/history/issues/000066-close-log-line-under-day-header.md
// is exactly this — first `## Log` at line 22 inside a fence, real one at 68.
const issueQuotingItsOwnLogFormat = `---
id: 000066
status: working
---

# close writes the log line under the day header

## Problem

Today ` + "`close`" + ` appends to the end of the section:

` + "```markdown" + `
## Log

### 2026-01-01
- an OLD entry quoted as an example
` + "```" + `

It should go under the matching day header instead.

## Plan

- [x] a

## Log

### 2026-09-02
- the real entry
`

// TestLogHasEntryToday_IgnoresAQuotedLogSection is the live instance from the
// M1 review. logHasEntryToday took the FIRST `^## Log`, which here is the quoted
// example — so the reopen guard read a fenced block instead of the real Log and
// answered from the wrong region of a real issue file.
func TestLogHasEntryToday_IgnoresAQuotedLogSection(t *testing.T) {
	if !logHasEntryToday(issueQuotingItsOwnLogFormat, "2026-09-02") {
		t.Error("today's real entry not found — the guard is reading the quoted ## Log")
	}
	if logHasEntryToday(issueQuotingItsOwnLogFormat, "2026-01-01") {
		t.Error("matched a date that exists ONLY inside the quoted example")
	}
}

// TestInsertLogLine_TargetsTheRealLogSection retires #66's last-match heuristic.
// Last-match was only accidentally right: it fails when a quoted `## Log` sits
// AFTER the real one, which the fence-aware lookup handles by construction.
func TestInsertLogLine_TargetsTheRealLogSection(t *testing.T) {
	got := insertLogLine(issueQuotingItsOwnLogFormat, "- 2026-09-02: closed")
	if !strings.Contains(got, "- the real entry\n- 2026-09-02: closed") &&
		!strings.Contains(got, "- 2026-09-02: closed\n- the real entry") {
		t.Errorf("line did not land in the real Log section:\n%s", got)
	}
	if strings.Contains(got, "an OLD entry quoted as an example\n- 2026-09-02: closed") {
		t.Error("line was filed into the QUOTED example block")
	}

	// The case last-match gets wrong and the scanner does not: a quoted ## Log
	// positioned after the real one.
	trailing := issueQuotingItsOwnLogFormat + "\n## Spec\n\n```markdown\n## Log\n\n### 2020-01-01\n```\n"
	got = insertLogLine(trailing, "- 2026-09-02: closed")
	if strings.Contains(got, "### 2020-01-01\n- 2026-09-02: closed") {
		t.Error("line filed into a TRAILING quoted ## Log — last-match's failure mode")
	}
}

// TestInsertLogLine_IgnoresAQuotedDayHeaderLater is BR-3 from the M2 review:
// anchoring the `## Log` HEADING fence-aware was only half the fix. The
// `### <date>` search still ran from that heading to EOF, so a quoted day header
// in a LATER section captured the insert — the same class of bug one level down.
//
// The real Log section deliberately has NO day header for today, so an unbounded
// search walks past it and finds the quoted one; a bounded search finds nothing
// and falls back to the top of the real section, which is correct.
func TestInsertLogLine_IgnoresAQuotedDayHeaderLater(t *testing.T) {
	body := `# t

## Log

### 2026-01-01
- an older entry

## Side quests

An example of the format:

` + "```markdown" + `
### 2026-09-02
- a QUOTED entry
` + "```" + `
`
	got := insertLogLine(body, "- 2026-09-02: closed")
	if strings.Contains(got, "### 2026-09-02\n- 2026-09-02: closed") {
		t.Errorf("close line filed under the QUOTED day header in a later section:\n%s", got)
	}
	// Asserted by ORDER rather than exact whitespace: insertLogLine's fallback
	// emits an extra blank line after `## Log` by design (documented at the
	// function), and pinning that here would couple this test to a quirk it is
	// not about.
	at := strings.Index(got, "- 2026-09-02: closed")
	logAt, olderAt := strings.Index(got, "## Log"), strings.Index(got, "### 2026-01-01")
	if at < 0 || at < logAt || at > olderAt {
		t.Errorf("close line did not land at the top of the real Log section:\n%s", got)
	}
}

// TestPlanGateContent_IgnoresAQuotedEstimateHeading is BR-6: the third swept
// site had no coverage.
//
// planGateContent strips the `## Estimate` section so the plan-quality gate's
// pass-through hash doesn't change when only the estimate does. Its line scan
// used to treat a `## ` inside a fence as a real heading, so an issue quoting an
// estimate block — which #211 and #208 both do — would start or stop stripping
// at an example and hash a different document than intended, re-dispatching the
// judge on an unchanged plan (or, worse, passing through a changed one).
func TestPlanGateContent_IgnoresAQuotedEstimateHeading(t *testing.T) {
	issueWith := func(estimate string) string {
		return `---
id: 000999
estimate_hours: 1.0
---

# t

## Spec

Quoting the block this issue adds:

` + "```markdown" + `
## Estimate

a quoted example
` + "```" + `

Real prose that must survive.

## Estimate

` + estimate + `

## Plan

- [x] a
`
	}
	got := planGateContent(issueWith("model: v3.1\ntotal: 1.0"))

	if !strings.Contains(got, "Real prose that must survive") {
		t.Errorf("stripping started at the QUOTED ## Estimate and ate the Spec:\n%s", got)
	}
	if strings.Contains(got, "model: v3.1") {
		t.Errorf("the real ## Estimate section was not stripped:\n%s", got)
	}
	if !strings.Contains(got, "- [x] a") {
		t.Errorf("stripping ran past the real ## Estimate into ## Plan:\n%s", got)
	}
	// The whole point: changing only the estimate must not change the hash input.
	if other := planGateContent(issueWith("model: v3.1\ntotal: 9.9")); other != got {
		t.Error("plan-gate content changed when only the estimate did — the pass-through hash would re-dispatch")
	}
}

// TestMilestoneTick_OnlyTicksTheRealPlan is BR-12: the write side of the class.
// The tick used to ReplaceAll over the WHOLE issue body, so a `- [ ] Mx` inside
// a quoted example — anywhere, including outside the Plan — was ticked too.
func TestMilestoneTick_OnlyTicksTheRealPlan(t *testing.T) {
	body := `# t

## Problem

The plan format looks like:

` + "```markdown" + `
- [ ] M1 — a quoted example
` + "```" + `

## Plan

- [ ] M1 — the real row
- [ ] M2 — later

## Log

- [ ] M1 — not a plan row at all
`
	pat := regexp.MustCompile(`(?m)^(- )\[[ .]\]( M1\b)`)
	start, end, ok := issue.SectionByteBounds(body, "Plan", issue.UnterminatedIsProse)
	if !ok {
		t.Fatal("Plan section not found")
	}
	lines := strings.Split(body[start:end], "\n")
	inside := issue.FenceSpans(lines, issue.UnterminatedIsProse)
	n := 0
	for i, line := range lines {
		if !inside[i] && pat.MatchString(line) {
			lines[i] = pat.ReplaceAllString(line, "${1}[x]${2}")
			n++
		}
	}
	got := body[:start] + strings.Join(lines, "\n") + body[end:]

	if n != 1 {
		t.Errorf("ticked %d row(s), want exactly 1 (the real Plan row)", n)
	}
	if !strings.Contains(got, "- [x] M1 — the real row") {
		t.Errorf("the real Plan row was not ticked:\n%s", got)
	}
	if !strings.Contains(got, "- [ ] M1 — a quoted example") {
		t.Errorf("a QUOTED example row was ticked:\n%s", got)
	}
	if !strings.Contains(got, "- [ ] M1 — not a plan row at all") {
		t.Errorf("a row outside ## Plan was ticked:\n%s", got)
	}
}
