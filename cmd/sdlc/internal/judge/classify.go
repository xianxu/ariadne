package judge

import (
	"regexp"
	"strings"
)

// Outcome classifies an agent's output for a check. Mirrors
// scripts/lib.sh's three-way split: clean (no violations), info
// (informational reminder, e.g. REMINDER:), and failure (everything
// else — there's content that demands attention).
type Outcome int

const (
	Clean Outcome = iota
	Info
	Failure
)

func (o Outcome) String() string {
	switch o {
	case Clean:
		return "clean"
	case Info:
		return "info"
	case Failure:
		return "failure"
	}
	return "unknown"
}

// cleanRE matches the well-known "no findings" sentinels emitted by
// our prompt templates. Case-insensitive. Ported from
// scripts/lib.sh's is_clean_check_output grep.
var cleanRE = regexp.MustCompile(`(?i)no (DRY|PURE) violations found|all tests pass|no changes needed|in sync|no issue files changed`)

// infoRE matches reminder-style output (the lessons category) — not
// a failure, but worth surfacing.
var infoRE = regexp.MustCompile(`(?i)REMINDER:`)

// Classify returns the Outcome for a single agent's output. Empty
// output is treated as failure (the agent should have said *something*).
func Classify(output string) Outcome {
	s := strings.TrimSpace(output)
	if s == "" {
		return Failure
	}
	if cleanRE.MatchString(s) {
		return Clean
	}
	if infoRE.MatchString(s) {
		return Info
	}
	return Failure
}
