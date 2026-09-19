package issue

import (
	"strings"
	"testing"
)

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

// TestMilestonesInPlanOrder pins the ONE milestone enumeration every reader uses
// (#231): close's verdict gate, the worktree sizing hint, and change-code's flow
// inference. The em-dash form is the dominant real-world shape, so a
// colon-only parser that missed it was not a style difference but a wrong
// answer.
func TestMilestonesInPlanOrder(t *testing.T) {
	cases := []struct {
		name, plan string
		want       []string
	}{
		{"em-dash", "- [ ] M1 — first\n- [x] M2 — second\n", []string{"M1", "M2"}},
		{"colon", "- [ ] M1: first\n- [ ] M12c: later\n", []string{"M1", "M12c"}},
		{"bold", "- [x] **M1 — first**\n- [.] **M4b — wip**\n", []string{"M1", "M4b"}},
		{"duplicate after a revision", "- [x] M1 — a\n- [ ] M2 — b\n- [ ] M1 — a again\n", []string{"M1", "M2"}},
		{"plain items are not milestones", "- [ ] write the test\n- [ ] Mx is prose\n", nil},
		{"a word starting with M is not a tag", "- [ ] Measure it\n", nil},
	}
	for _, c := range cases {
		got := MilestonesInPlanOrder(c.plan)
		if strings.Join(got, ",") != strings.Join(c.want, ",") {
			t.Errorf("%s: MilestonesInPlanOrder = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestMilestonesInPlanOrderSkipsFencedRows drives the enumeration through the
// fence-filtered body, which is how every production caller obtains it.
func TestMilestonesInPlanOrderSkipsFencedRows(t *testing.T) {
	body := "## Plan\n\n- [ ] M1 — real\n\n```markdown\n- [ ] M9 — quoted example\n```\n\n## Log\n"
	plan, ok := PlanItemsBody(body)
	if !ok {
		t.Fatal("## Plan not found")
	}
	if got := MilestonesInPlanOrder(plan); strings.Join(got, ",") != "M1" {
		t.Errorf("got %v, want [M1] — a milestone quoted in a fence is not a milestone", got)
	}
}
