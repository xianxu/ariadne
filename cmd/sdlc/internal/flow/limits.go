package flow

import (
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
)

// The hard shell (#231): a quick-flow issue must stay inside these limits, and
// have no Mx rows, whoever declared it quick. They are the ONE statement of the
// thresholds — the change-code inference, the close-time check, start-plan's
// guidance, the help text and the constitution (which points at `sdlc
// change-code --help`) all read them from here. The shell measures size alone:
// how many files a change spreads across, and where they sit, do not count
// (operator, 2026-09-18) — below the limits, the tests and the one close review
// are the guard.
const (
	// MaxAddedLines is the most lines a quick diff may add to code files — code
	// files as churn.IsCodeFile classifies them (churn.CodeFileRule says which),
	// lines as git's insertions, the number the churn report and calibration
	// ledger use. Deleted lines do not count.
	MaxAddedLines = 100
	// MaxDesignLines is the most lines the design may run — DesignRule says
	// which text. A plan runs longer than the code it describes (it cites code
	// and justifies it), so this limit is its own number rather than a multiple
	// of MaxAddedLines. It replaced "a durable plan exists" as the design-side
	// signal (operator, 2026-09-18): a short plan no longer forces the full flow.
	MaxDesignLines = 500
	// MaxAddedLinesAfterReview is how far a quick issue's final diff may grow
	// once its small-diff close review has run. Close measures the head the
	// review saw; fixes made after the verdict ride into the close commit
	// unmeasured, so the publish check measures the final diff and, past this,
	// sends the issue back to close — which upgrades it and runs the full
	// review. Twice the line limit (operator, 2026-09-18): fixes answering the
	// review's own findings are accepted up to there, which keeps the re-review
	// rare.
	MaxAddedLinesAfterReview = 2 * MaxAddedLines
)

// DesignRule is the one statement of what the design limit counts.
const DesignRule = "## Spec, ## Plan and the durable plan, in lines"

// AfterReviewSummary is the one sentence every surface prints for the limit on
// fixes made after a quick issue's close review.
func AfterReviewSummary() string {
	return fmt.Sprintf("a quick issue's final diff, fixes made after its close review included, may add at most "+
		"%d lines in code files (twice the shell's %d) before the publish check sends it back to "+
		"`sdlc close` for the full review", MaxAddedLinesAfterReview, MaxAddedLines)
}

// ShellSummary is the one sentence every surface prints for the shell.
func ShellSummary() string {
	return fmt.Sprintf("at most %d added lines in code files (%s), a design of at most %d lines (%s), "+
		"and no Mx milestones", MaxAddedLines, churn.CodeFileRule, MaxDesignLines, DesignRule)
}
