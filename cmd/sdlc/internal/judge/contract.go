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
