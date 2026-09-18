package flow

import "fmt"

// The hard shell (#231): a quick-flow issue must stay inside these limits,
// whoever declared it quick. They are the ONE statement of the threshold — the
// close-time check, start-plan's guidance, the help text and the constitution
// (which points at `sdlc change-code --help`) all read them from here.
const (
	// MaxCodeFiles is the most code files a quick diff may change. Code files
	// exclude tests and docs (churn.IsCodeFile).
	MaxCodeFiles = 2
	// MaxChangedLines is the most lines a quick diff may add to code files —
	// git's insertions, the number the churn report and calibration ledger use.
	MaxChangedLines = 100
)

// ShellSummary is the one sentence every surface prints for the shell.
func ShellSummary() string {
	return fmt.Sprintf("at most %d code files and %d added lines (tests and docs excluded), "+
		"no declared shared surface touched, and no Mx milestones", MaxCodeFiles, MaxChangedLines)
}
