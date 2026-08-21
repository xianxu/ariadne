package judge

import "testing"

// #194 M2: the boundary review was the only LLM gate with no memory. Its prompt must
// now carry the prior rounds' findings, the way plan-quality's has since #187 — that
// asymmetry is what made the reviewer renumber C1/C2/I1 fresh every round.
func TestMilestoneReviewPrompt_CarriesPriorFindings(t *testing.T) {
	in := goldenInput
	in.PriorFindings = "### BR-2 (Important, round 1)\nthe reviewer said this last time"
	got := BuildPrompt(MilestoneReview, in)

	if !contains(got, "BR-2") {
		t.Error("milestone-review prompt must embed the prior findings block")
	}
	if !contains(got, "the reviewer said this last time") {
		t.Error("prior findings body missing from the rendered prompt")
	}
	// The output fence must be there too, or the reviewer has no way to dispose of them.
	if !contains(got, "```findings") {
		t.Error("milestone-review prompt must instruct the findings fence so rounds can dispose")
	}
	if !contains(got, "dispose:") {
		t.Error("the fence instruction must include the dispose key")
	}
}

// Round one has no prior rounds; the placeholder must degrade to a stated absence
// rather than an empty hole the reviewer might read as an omission.
func TestMilestoneReviewPrompt_FirstRoundStatesNoPriorRounds(t *testing.T) {
	in := goldenInput
	in.PriorFindings = ""
	if got := BuildPrompt(MilestoneReview, in); !contains(got, "(no prior rounds)") {
		t.Error("a first-round boundary review must say there are no prior rounds")
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}
