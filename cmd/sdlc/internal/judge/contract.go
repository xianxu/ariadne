package judge

import "strings"

// The judge output contract (#70). Every agent-emitting judge leads its result
// with a `VERDICT: <TOKEN> (confidence: …)` line; the classifier reads ONLY that
// token — prose, findings, and severity tags below it are advisory. This file is
// the single source of truth for the token taxonomy that both the classifier
// (classify.go) and the prompts (prompts.go, via ContractBlock — wired in M2)
// derive from, so the two can't drift.
//
// Token → (Classify Outcome, milestone Verdict, blocking) — the contract table:
//
//	TOKEN          Outcome   Verdict          blocking  meaning
//	────────────── ───────── ──────────────── ───────── ─────────────────────────
//	CLEAN          Clean     —                no        no issues; ship
//	INFO           Info      —                no        non-blocking notes only
//	SHIP           Info      VerdictShip      no        milestone: ship
//	FIX-THEN-SHIP  Info      VerdictFixThenShip no      milestone: ship after fixes
//	FAILURE        Failure   —                YES       must fix before shipping
//	REWORK         Failure   VerdictRework    YES       milestone: rework
//	BLOCK          Failure   —                YES       generic hard block
//
// The gate is binary (blocking vs not); the token carries the human nuance.
// Lessons is the one exception — it runs no agent and emits a fixed `REMINDER:`
// line (→ Info), documented in the schema doc.

// ContractPreamble is the shared output-format instruction every agent-emitting
// judge embeds verbatim — the machine-read part of the contract. Category token
// meanings (what CLEAN/SHIP/etc. mean for THAT check) live in each prompt; this
// is only the FORMAT the parser depends on, so prompt and parser can't drift.
// The human-readable mirror is construct/judge-output-contract.md (kept in sync
// by a drift test). #70.
const ContractPreamble = `OUTPUT CONTRACT (machine-read — do not deviate). Your response's FIRST line
MUST be exactly:

    VERDICT: <TOKEN> (confidence: high | medium | low)

The parser reads ONLY this <TOKEN>. Findings, notes, and severity tags below it
are advisory — a non-blocking verdict WITH notes still PASSES the gate. Do not
put a title, heading, or any preamble above the VERDICT line; it must lead.`

// BoundaryReviewContract is the output contract for the boundary review (#147).
// Unlike the shared ContractPreamble (used by the pre-merge tri-state judges, which
// requires the VERDICT: line to lead), this yields first-line precedence to the
// fenced ```verdict block — the authoritative structured handoff the binary reads
// via ParseVerdictBlock. The bare VERDICT: line is accepted only as a fallback, so
// the two "lead" instructions no longer conflict in the milestone-review prompt.
const BoundaryReviewContract = "OUTPUT CONTRACT (machine-read — do not deviate). LEAD your response with the\n" +
	"fenced ```verdict block shown above — that is the authoritative handoff the binary\n" +
	"reads (its `verdict:` value is one of the listed tokens). Everything after the block\n" +
	"is advisory: a non-blocking verdict WITH findings still PASSES the gate. A bare\n" +
	"`VERDICT: <TOKEN>` line is accepted only as a FALLBACK when the block is absent."

// blockingTokens are the verdict tokens that fail a gate.
var blockingTokens = map[string]bool{
	"FAILURE": true,
	"REWORK":  true,
	"BLOCK":   true,
}

// ContractTokens is the canonical token set, in contract order (non-blocking
// first, then blocking). The schema doc + a drift test (M2) assert against this.
var ContractTokens = []string{
	"CLEAN", "INFO", "SHIP", "FIX-THEN-SHIP", // non-blocking
	"FAILURE", "REWORK", "BLOCK", // blocking
}

// Blocking reports whether a verdict token fails the gate. Only called with a
// token already parsed from a VERDICT: line; the "no VERDICT line at all" case is
// handled fail-closed by Classify, not here.
func Blocking(token string) bool {
	return blockingTokens[strings.ToUpper(token)]
}

// outcomeForToken maps a parsed verdict token to the tri-state Outcome the
// preflight/milestone consumers gate on: CLEAN→Clean, any blocking token→Failure,
// every other (non-blocking) token→Info.
func outcomeForToken(token string) Outcome {
	switch strings.ToUpper(token) {
	case "CLEAN":
		return Clean
	default:
		if Blocking(token) {
			return Failure
		}
		return Info
	}
}
