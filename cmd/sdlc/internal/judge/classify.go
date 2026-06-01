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
	VerdictUnknown       Verdict = "unknown"   // judge ran, no leading verdict found
)

var (
	// leadingMarkupRE strips leading markdown emphasis / heading / quote /
	// list markers from a line so an emphasized or headed verdict
	// (`**SHIP**`, `## SHIP`, `> REWORK`, `- FIX-THEN-SHIP`) still parses.
	leadingMarkupRE = regexp.MustCompile("^[ \t>#*_`-]+")

	// verdictTokenRE matches a verdict token that *opens* a line's content
	// and stands alone — followed only by emphasis-close / whitespace then
	// a confidence paren or end-of-line. The trailing guard rejects prose
	// like "SHIP-blocking" or "ship it after a tweak".
	verdictTokenRE = regexp.MustCompile("^(SHIP|FIX-THEN-SHIP|REWORK)[ \t*_`]*(\\(|$)")

	// structuralLineRE matches markdown lines that carry no verdict and no
	// prose — headings and horizontal rules. A reviewer may emit a title
	// or a "## Verdict" header before the verdict itself; we skip those.
	structuralLineRE = regexp.MustCompile("^(#{1,6}\\s|[-=*]{3,}\\s*$)")
)

// ParseVerdict extracts the verdict label from the agent's milestone-
// review output. The verdict must *lead* the report — but a reviewer
// commonly prefixes a markdown title and/or a `## Verdict` header, so we
// skip blank + structural (heading/rule) lines and tolerate emphasis,
// then require the first line of real content to be the verdict. The
// first *prose* line that isn't a verdict ends the search as Unknown —
// preserving precision (a stray "SHIP" buried later in the report is not
// mistaken for the verdict).
//
// Pure: no IO, deterministic on its input. Lives in the judge package
// alongside Classify so the prompt + parser sit next to each other.
func ParseVerdict(output string) Verdict {
	for _, line := range strings.Split(output, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		stripped := leadingMarkupRE.ReplaceAllString(t, "")
		if m := verdictTokenRE.FindStringSubmatch(stripped); m != nil {
			switch m[1] {
			case "SHIP":
				return VerdictShip
			case "FIX-THEN-SHIP":
				return VerdictFixThenShip
			case "REWORK":
				return VerdictRework
			}
		}
		// Not a verdict: skip a title / section header / rule and keep
		// looking; stop at the first real prose line.
		if structuralLineRE.MatchString(t) {
			continue
		}
		return VerdictUnknown
	}
	return VerdictUnknown
}
