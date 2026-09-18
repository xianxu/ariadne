package flow

import (
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
)

// The hard shell (#231): a quick-flow issue must stay inside this limit, and
// have no Mx rows, whoever declared it quick. It is the ONE statement of the
// threshold — the close-time check, start-plan's guidance, the help text and
// the constitution (which points at `sdlc change-code --help`) all read it from
// here. The shell measures size alone: how many files a change spreads across,
// and where they sit, do not count (operator, 2026-09-18) — below the limit,
// the tests and the one close review are the guard.
const (
	// MaxAddedLines is the most lines a quick diff may add to code files — code
	// files as churn.IsCodeFile classifies them (churn.CodeFileRule says which),
	// lines as git's insertions, the number the churn report and calibration
	// ledger use. Deleted lines do not count.
	MaxAddedLines = 100
)

// ShellSummary is the one sentence every surface prints for the shell.
func ShellSummary() string {
	return fmt.Sprintf("at most %d added lines in code files (%s), and no Mx milestones",
		MaxAddedLines, churn.CodeFileRule)
}
