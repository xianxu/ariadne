package issue

import "testing"

func TestCountPlanItems_MixedStates(t *testing.T) {
	body := `## Problem
something

## Plan

- [ ] M1 — open
- [x] M2 — done
- [.] M3 — in progress
- [ ] M4 — later

## Log
- [ ] this is outside the Plan section, must not count
`
	total, ticked := CountPlanItems(body)
	if total != 4 || ticked != 1 {
		t.Errorf("got (total=%d, ticked=%d) want (4, 1)", total, ticked)
	}
}

func TestCountPlanItems_NoPlanSection(t *testing.T) {
	if total, ticked := CountPlanItems("## Problem\nnothing else\n"); total != 0 || ticked != 0 {
		t.Errorf("got (%d, %d) want (0, 0)", total, ticked)
	}
}

func TestCountPlanItems_PlanAtEnd(t *testing.T) {
	body := "## Problem\np\n\n## Plan\n\n- [ ] only item\n"
	total, ticked := CountPlanItems(body)
	if total != 1 || ticked != 0 {
		t.Errorf("got (%d, %d) want (1, 0)", total, ticked)
	}
}

func TestPlanUncheckedRE_DoesNotMatchTicked(t *testing.T) {
	if PlanUncheckedRE.MatchString("- [x] done") {
		t.Error("PlanUncheckedRE should not match - [x] done")
	}
	if !PlanUncheckedRE.MatchString("- [ ] open") {
		t.Error("PlanUncheckedRE should match - [ ] open")
	}
	if !PlanUncheckedRE.MatchString("- [.] in-progress") {
		t.Error("PlanUncheckedRE should match - [.] in-progress")
	}
}
