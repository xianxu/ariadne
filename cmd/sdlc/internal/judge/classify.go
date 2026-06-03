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

// verdictTokenLineRE matches a `VERDICT: <TOKEN>` line carrying any contract
// token. We scan the WHOLE output for it (not just the first non-empty line) —
// which is what fixes the preamble-before-verdict bug (#70): a judge that writes
// a title or "I've reviewed…" line before `VERDICT: CLEAN` is no longer mis-read.
//
// Precision guard: the token must be followed by emphasis-close/whitespace then
// a `(confidence …)` paren OR end-of-line — same trailing guard as the bare-token
// `verdictTokenRE`. This rejects a line that *quotes* the contract as prose
// (`VERDICT: BLOCK is the generic hard block`, `VERDICT: CLEAN means no issues`),
// which matters because judges review THIS parser and write such lines. The
// only residual (rare, low-risk) accept is a line that is *just* a standalone
// quoted token. Tolerant of leading markdown markup (stripped by the caller).
var verdictTokenLineRE = regexp.MustCompile("(?i)^VERDICT:[ \t]*(CLEAN|INFO|FAILURE|SHIP|FIX-THEN-SHIP|REWORK|BLOCK)[ \t*_`]*(\\(|$)")

// ParseVerdictToken scans output for the first `VERDICT:` line and returns its
// upper-cased token (e.g. "CLEAN", "SHIP"). ok=false when no VERDICT: line is
// present anywhere — the caller decides the fallback (legacy sentinels) or
// fail-closed. Pure; the single robust parse both Classify and ParseVerdict use.
func ParseVerdictToken(output string) (string, bool) {
	for _, line := range strings.Split(output, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		stripped := leadingMarkupRE.ReplaceAllString(t, "")
		if m := verdictTokenLineRE.FindStringSubmatch(stripped); m != nil {
			return strings.ToUpper(m[1]), true
		}
	}
	return "", false
}

// Classify returns the Outcome for a single agent's output. Empty output is
// treated as failure (the agent should have said *something*).
//
// Primary path: the structured `VERDICT: <TOKEN>` line (robust scan), mapped to
// an Outcome via the contract (contract.go). Legacy fallback: the Lessons
// `REMINDER:` line and the old DRY/PURE sentinels, kept so un-migrated outputs
// don't break before M2 folds them into the contract. No VERDICT line + no
// legacy match → Failure (fail closed; the agent broke the output contract).
func Classify(output string) Outcome {
	s := strings.TrimSpace(output)
	if s == "" {
		return Failure
	}
	if tok, ok := ParseVerdictToken(s); ok {
		return outcomeForToken(tok)
	}
	if infoRE.MatchString(s) { // Lessons REMINDER:
		return Info
	}
	if cleanRE.MatchString(s) { // legacy DRY/PURE "No X violations found"
		return Clean
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
	VerdictShip        Verdict = "SHIP"
	VerdictFixThenShip Verdict = "FIX-THEN-SHIP"
	VerdictRework      Verdict = "REWORK"
	VerdictNotRun      Verdict = "not-run" // judge skipped or errored
	VerdictUnknown     Verdict = "unknown" // judge ran, no leading verdict found
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

	// verdictConfidenceRE matches a verdict line carrying the confidence
	// parenthetical — e.g. "FIX-THEN-SHIP (confidence: high)". This is the
	// high-precision fallback for when the reviewer narrates investigation
	// prose *before* the verdict (so the leading scan stops early): prose
	// effectively never writes "<TOKEN> (confidence: …)", so accepting it
	// anywhere doesn't reintroduce the bare-token false positives that the
	// leading scan's precision guard exists to prevent.
	verdictConfidenceRE = regexp.MustCompile("(?m)^[ \t>#*_`-]*(SHIP|FIX-THEN-SHIP|REWORK)[ \t*_`]*\\([Cc]onfidence")
)

func verdictFor(token string) Verdict {
	switch token {
	case "SHIP":
		return VerdictShip
	case "FIX-THEN-SHIP":
		return VerdictFixThenShip
	case "REWORK":
		return VerdictRework
	}
	return VerdictUnknown
}

// ParseVerdict extracts the verdict label from the agent's milestone-
// review output. Two passes:
//
//  1. Leading scan (precise): the verdict should *lead* the report. Skip
//     blank + structural (title/heading/rule) lines, tolerate emphasis,
//     and require the first line of real content to be the verdict. Stop
//     at the first *prose* line — a stray "SHIP" buried later is not the
//     verdict.
//  2. Confidence fallback: if the reviewer narrated prose before the
//     verdict, accept a confidence-qualified verdict line anywhere (a
//     signal prose doesn't forge).
//
// Pure: no IO, deterministic on its input. Lives in the judge package
// alongside Classify so the prompt + parser sit next to each other.
func ParseVerdict(output string) Verdict {
	// Primary (#70): the structured `VERDICT: <TOKEN>` line, found robustly even
	// behind a preamble. Only a SHIP-family token is a milestone verdict; a
	// CLEAN/INFO/FAILURE on a VERDICT: line falls through to the legacy scan.
	if tok, ok := ParseVerdictToken(output); ok {
		if v := verdictFor(tok); v != VerdictUnknown {
			return v
		}
	}
	// Legacy: bare leading token (`SHIP`) + confidence fallback, for milestone
	// prompts not yet migrated to the `VERDICT:` prefix (M2 migrates them).
	for _, line := range strings.Split(output, "\n") {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		stripped := leadingMarkupRE.ReplaceAllString(t, "")
		if m := verdictTokenRE.FindStringSubmatch(stripped); m != nil {
			return verdictFor(m[1])
		}
		// Not a verdict: skip a title / section header / rule and keep
		// looking; a real prose line ends the leading scan.
		if structuralLineRE.MatchString(t) {
			continue
		}
		break
	}
	if m := verdictConfidenceRE.FindStringSubmatch(output); m != nil {
		return verdictFor(m[1])
	}
	return VerdictUnknown
}
