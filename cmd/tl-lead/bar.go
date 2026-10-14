package main

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/pkg/markdoc"
)

// planBar renders an issue's progress from its body: its Mx milestones when the
// Plan has any, else its checkboxes. The counts come from the same scanners the
// close gate uses, so the bar can't disagree with a close.
func planBar(body string) string {
	items, ok := markdoc.PlanItemsBody(body)
	if !ok {
		return "no plan yet"
	}
	if ms := markdoc.MilestonesInPlanOrder(items); len(ms) > 0 {
		ticked := markdoc.TickedMilestones(items)
		done := 0
		for _, m := range ms {
			if ticked[m] {
				done++
			}
		}
		return bar(done, len(ms), "M")
	}
	total, done := markdoc.CountPlanItems(body)
	if total == 0 {
		return "no plan yet"
	}
	return bar(done, total, "tasks")
}

func bar(done, total int, unit string) string {
	return fmt.Sprintf("%s%s %s %d/%d", strings.Repeat("▉", done), strings.Repeat("▕", total-done), unit, done, total)
}
