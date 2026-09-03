package issue

import (
	"regexp"
	"strings"
)

// Plan-section parsing — shared between cmd/sdlc/close (which refuses an
// issue close if Plan has unticked items) and cmd/sdlc/state (which
// counts ticked vs total for the inspection surface).
//
// Moved here from cmd/sdlc package-level so both callers reference one
// source of truth — per M2 review I5, the cross-file coupling via package
// vars was an implicit dependency the next refactor could break.

// PlanItemRE matches one `- [s] ...` plan item where `s` is the state
// char (space, `x`, or `.`). The captured state char lets callers count
// total vs ticked in a single pass.
var PlanItemRE = regexp.MustCompile(`(?m)^- \[([ x.])\] `)

// PlanUncheckedRE matches one `- [ ] ...` or `- [.] ...` plan item line
// (i.e., NOT a ticked `[x]`). Used by close to refuse an issue close
// when the Plan still has open items.
var PlanUncheckedRE = regexp.MustCompile(`(?m)^- \[[ .]\] .*$`)

// CountPlanItems counts total and ticked plan items inside the `## Plan`
// section of an issue body. Returns (0, 0) if no Plan section exists.
func CountPlanItems(body string) (total, ticked int) {
	section, ok := PlanItemsBody(body)
	if !ok {
		return 0, 0
	}
	for _, mm := range PlanItemRE.FindAllStringSubmatch(section, -1) {
		total++
		if mm[1] == "x" {
			ticked++
		}
	}
	return total, ticked
}

// TickMilestone marks every `- [ ] <Mx>` / `- [.] <Mx>` row in the REAL Plan
// section as done, returning the new body and how many rows it changed.
//
// PURE, and extracted for that reason (#211 M2 review). The logic lived inline
// in close.go's IO shell, so its test could only reproduce it line for line —
// asserting a copy of the code rather than the code (ARCH-PURE: "if a test needs
// to restate the logic to reach it, the logic is in the wrong place"). Third
// instance of that shape in this issue, so the fix is the extraction, not
// another test.
//
// Two filters, both load-bearing. Before #211 this was a ReplaceAll over the
// WHOLE document, so a `- [ ] M1` inside a quoted example — or in ## Log, or in
// ## Problem — was ticked along with the real row:
//
//	SectionByteBounds  confines the rewrite to the real ## Plan
//	FenceSpans         skips rows inside a fenced example within it
//
// n == 0 does not distinguish "no Plan section" from "no matching row"; callers
// that report a cause to the operator should use HasSection-style checks rather
// than inferring one.
func TickMilestone(body, milestone string) (string, int) {
	start, end, ok := SectionByteBounds(body, "Plan", UnterminatedIsProse)
	if !ok {
		return body, 0
	}
	pat := regexp.MustCompile(`(?m)^(- )\[[ .]\]( ` + regexp.QuoteMeta(milestone) + `\b)`)
	lines := strings.Split(body[start:end], "\n")
	inside := FenceSpans(lines, UnterminatedIsProse)
	n := 0
	for i, line := range lines {
		if inside[i] || !pat.MatchString(line) {
			continue
		}
		lines[i] = pat.ReplaceAllString(line, "${1}[x]${2}")
		n++
	}
	if n == 0 {
		return body, 0
	}
	return body[:start] + strings.Join(lines, "\n") + body[end:], n
}
