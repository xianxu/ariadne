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
// our prompt templates. Case-insensitive. Legacy grep path — kept as a
// fallback for outputs that don't carry a `VERDICT:` line (older
// prompts, agents that ignored the instruction).
var cleanRE = regexp.MustCompile(`(?i)no (DRY|PURE) violations found|all tests pass|no changes needed|in sync|no issue files changed`)

// infoRE matches reminder-style output (the lessons category) — not
// a failure, but worth surfacing.
var infoRE = regexp.MustCompile(`(?i)REMINDER:`)

// verdictLineRE matches the structured `VERDICT:` line that Plan +
// Specs prompts instruct subagents to emit as line 1 (Lessons skips
// the agent entirely). Same shape as MilestoneReview's first-line
// verdict (`SHIP | FIX-THEN-SHIP | REWORK`) but with the Outcome
// trio as labels so the prompt → classifier mapping is 1:1.
//
// Tolerant on leading whitespace and the optional `(confidence: …)`
// parenthetical — the prompt asks for it but we don't punish drift.
var verdictLineRE = regexp.MustCompile(`^\s*VERDICT:\s*(CLEAN|INFO|FAILURE)\b`)

// parseVerdictLine looks for the structured verdict on the first
// non-empty line of the output. Returns (outcome, true) on hit, or
// (Failure, false) when no verdict line is present — letting the
// caller fall back to the legacy grep path.
func parseVerdictLine(s string) (Outcome, bool) {
	for _, line := range strings.Split(s, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		m := verdictLineRE.FindStringSubmatch(t)
		if m == nil {
			return Failure, false
		}
		switch m[1] {
		case "CLEAN":
			return Clean, true
		case "INFO":
			return Info, true
		case "FAILURE":
			return Failure, true
		}
		return Failure, false
	}
	return Failure, false
}

// Classify returns the Outcome for a single agent's output. Empty
// output is treated as failure (the agent should have said *something*).
//
// Two-tier strategy: prefer the structured `VERDICT:` line emitted by
// the migrated prompts; fall back to the legacy sentinel grep so older
// prompt outputs and stray agent prose don't break overnight.
func Classify(output string) Outcome {
	s := strings.TrimSpace(output)
	if s == "" {
		return Failure
	}
	if o, ok := parseVerdictLine(s); ok {
		return o
	}
	if cleanRE.MatchString(s) {
		return Clean
	}
	if infoRE.MatchString(s) {
		return Info
	}
	return Failure
}

// Verdict names the discrete outcomes the milestone-review prompt is
// instructed to emit on its first line. The string values are the
// canonical labels — used both in git-trailer values (Review-Verdict:)
// and in the human-mirror log line — so any addition here must also
// land in the prompt template (prompts.go MilestoneReview branch) and
// in the verifier helper (close.go's milestone-verdict guard).
type Verdict string

const (
	VerdictShip          Verdict = "SHIP"
	VerdictFixThenShip   Verdict = "FIX-THEN-SHIP"
	VerdictRework        Verdict = "REWORK"
	VerdictNotRun        Verdict = "not-run"   // judge skipped or errored
	VerdictUnknown       Verdict = "unknown"   // judge ran, first line unparseable
)

// verdictRE matches the milestone-review prompt's first-line verdict
// shape:
//
//	SHIP | FIX-THEN-SHIP | REWORK   (confidence: high | medium | low)
//
// Tolerant on whitespace around the pipes and on the confidence
// parenthetical (the prompt asks for it but we don't punish drift).
// Anchored at start-of-line of the first non-empty line; the prompt
// instructs the agent to emit this as line 1 of the response.
var verdictRE = regexp.MustCompile(`(?m)^\s*(SHIP|FIX-THEN-SHIP|REWORK)\b`)

// ParseVerdict extracts the verdict label from the agent's milestone-
// review output. Returns one of VerdictShip / VerdictFixThenShip /
// VerdictRework if the first non-empty line opens with one of those
// tokens, else VerdictUnknown.
//
// Pure: no IO, deterministic on its input. Lives in the judge package
// alongside Classify so the prompt + parser sit next to each other.
func ParseVerdict(output string) Verdict {
	// Walk to the first non-empty line. The prompt promises line 1,
	// but reviewers sometimes preface with a blank line or banner —
	// don't be brittle about it.
	for _, line := range strings.Split(output, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		m := verdictRE.FindStringSubmatch(t)
		if m == nil {
			return VerdictUnknown
		}
		switch m[1] {
		case "SHIP":
			return VerdictShip
		case "FIX-THEN-SHIP":
			return VerdictFixThenShip
		case "REWORK":
			return VerdictRework
		}
		return VerdictUnknown
	}
	return VerdictUnknown
}
