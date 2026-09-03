package main

import (
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
