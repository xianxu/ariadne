package bench

// Dimension is a subjective grading axis decided by blind head-to-head judging.
type Dimension struct {
	Key    string  `json:"key"`
	Group  string  `json:"group"` // "quality" | "workflow-fit"
	Weight float64 `json:"weight"`
	Prompt string  `json:"prompt"` // what the judge evaluates
}

// ObjectiveCheck is a measured grading axis (deterministic, agent-neutral).
type ObjectiveCheck struct {
	Key    string  `json:"key"`
	Group  string  `json:"group"`
	Weight float64 `json:"weight"`
}

// Rubric carries both grading groups; it is the single source of the
// quality/workflow-fit split that Scorecard scoring and Leaderboard aggregation
// both read.
type Rubric struct {
	Objective  []ObjectiveCheck `json:"objective"`
	Subjective []Dimension      `json:"subjective"`
}

// DefaultRubric is the standard grading set used by `sdlc bench freeze` unless a
// task overrides it.
func DefaultRubric() Rubric {
	return Rubric{
		Objective: []ObjectiveCheck{
			{Key: "build", Group: "quality", Weight: 1},
			{Key: "existing-tests", Group: "quality", Weight: 1},
			{Key: "new-tests", Group: "quality", Weight: 1},
			{Key: "completed", Group: "quality", Weight: 1},
			{Key: "artifact-log", Group: "workflow-fit", Weight: 1},
			{Key: "artifact-plan-ticked", Group: "workflow-fit", Weight: 1},
			{Key: "artifact-atlas", Group: "workflow-fit", Weight: 1},
			{Key: "gates-run", Group: "workflow-fit", Weight: 1},
		},
		Subjective: []Dimension{
			{Key: "elegance", Group: "quality", Weight: 1,
				Prompt: "Which solution is more elegant — DRY, pure core, root-cause not patch?"},
			{Key: "design-reasoning", Group: "quality", Weight: 1,
				Prompt: "Which reasons better about design/UI subtleties (read its spec/plan/diff)?"},
			{Key: "doc-quality", Group: "quality", Weight: 1,
				Prompt: "Which produced clearer, better-judged docs/spec/plan?"},
			{Key: "gate-judgment", Group: "workflow-fit", Weight: 1,
				Prompt: "Which made better decisions at the SDLC gates?"},
		},
	}
}
