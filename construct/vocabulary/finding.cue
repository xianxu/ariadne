// finding — the vocabulary of a gate finding: the severities a fresh-context judge may
// emit, which of them BLOCK a gate, and the dispositions a later round may assign to an
// earlier finding. Single source of truth for the `finding` noun (ariadne#187).
//
// The `agent-binary-handoff-schema` target names "the change-code plan/estimate judges" as
// the next boundary to schema, and this is that schema: once deterministic code branches
// on a severity, prose is the wrong source. sdlc reads the exported JSON; the plan-quality
// prompt renders the emitted set from it; the parser validates against it; the gate
// decision reads `categories.blocking`.
package finding

import "list"

// ── categories: the single concrete source of severity membership ──
// Stated once, here. #Severity is DERIVED via or(), so there is nothing to keep in sync.
// Only concrete data reaches the exported JSON — CUE definitions (#) do not export. The
// two categories PARTITION the severity set: every severity is in exactly one (a
// conformance test pins this).
//
// Names match code-review.md's long-standing buckets deliberately — the boundary review
// and the plan gate share ONE taxonomy (a drift test pins that too), so `Minor` here IS
// the "does not cost a round-trip" bucket.
categories: {
	blocking: ["Critical", "Important"] // undisposed ⇒ the gate refuses
	advisory: ["Minor"]                 // recorded for the close review; never blocks
}

// hardBlocking: the subset that still blocks PAST the round cap. Modeled rather than left
// as a `!= "Critical"` literal in the decision code, for the same reason `blocking` is.
// Must be a subset of categories.blocking (a conformance test pins this).
hardBlocking: ["Critical"]

kinds: ["defect", "unverifiable"]
#Kind: or(kinds)

#Severity: or(list.Concat([categories.blocking, categories.advisory]))

// ── dispositions: what a LATER round may say about an EARLIER finding ──
// PARTITIONED by the semantics the binary branches on, not a flat list. A flat list plus a
// prose gloss would put the closes-vs-leaves-open decision in a Go switch — the exact
// posture this file exists to reject. Concretely: adding `deferred` to a flat list would
// make the round validator accept it while the open-set computation had no case for it,
// wedging the finding open forever.
dispositions: {
	closing: ["addressed", "withdrawn", "disputed"] // the finding is settled; it stops blocking
	open: ["not-addressed"]                         // still open; it keeps blocking
}

#Disposition: or(list.Concat([dispositions.closing, dispositions.open]))

// ── authority: WHO may assign each disposition (ariadne#202) ──
// PARTITIONS the dispositions (a conformance test pins it). The judge disposes in its
// findings fence; the implementer disposes through `sdlc finding dispute`. Without this
// axis the only closing exit an implementer had was a fix, so a WRONG finding and a
// correct-but-not-worth-it one both left by the round cap, unadjudicated. A judge fence
// that emits an implementer disposition is a protocol error, exactly as an unknown id is.
authority: {
	judge: ["addressed", "withdrawn", "not-addressed"]
	implementer: ["disputed"]
}

// appealOnly: severities an IMPLEMENTER disposition cannot settle. A dispute of one is
// recorded and put to the next judge, but the finding keeps blocking until the judge
// withdraws it. The hard-blocking set by reference, not a copy: what blocks past the round
// cap is exactly what the implementer may not wave through themselves.
appealOnly: hardBlocking

// ── when: one-line semantics per severity (the documented-value source) ──
when: {
	"Critical":  "must fix before the gate is crossed"
	"Important": "fix before the gate if cheap; blocks until disposed"
	"Minor":     "note for the close review; never blocks a gate"
}

// ── whenDisposed: one-line semantics per disposition ──
whenDisposed: {
	"addressed":     "the plan changed to satisfy this finding"
	"not-addressed": "still open — the judge re-raises it this round"
	"withdrawn":     "the judge retracts it (mistaken, or overtaken by a design change)"
	"disputed":      "the implementer declines it with a rationale and a counter-measurement; it settles unless its severity is appeal-only, and the next judge answers it"
}

// ── discovery: where instances of this noun live. A gate ledger sidecar carries its
// findings in frontmatter, so a sidecar IS a set of finding instances. Repo-relative (the
// consumer joins to its repo root). Deliberately NOT `*-review.md` — that glob belongs to
// the verdict noun, and a gate ledger carries findings, not a boundary verdict. ──
discovery: {
	home: "workshop/plans" // the durable gate ledgers
	glob: "*-gate.md"      // change-code's `-plan-gate.md` + the boundary review's `-close-gate.md` (#194)
}

// ── #Finding: the structured handoff the judge emits + the binary validates ──
// Closed (fail-closed) — a finding is an atomic judgment, not an organically growing
// record. `id` is "new" on emission; the BINARY assigns the stable id, so the judge only
// ever has to REFER to identifiers it was handed.
#Finding: {
	id:       string // "new" when freshly raised; else a binary-assigned "<prefix>-<n>"
	severity: #Severity
	kind?: #Kind // absent means the legacy defect kind
	title:    string
	detail?:  string
	// family: a slug naming the underlying RULE a finding is an instance of, not the
	// symptom — `block-opener-rule`, not `bracket-depth-bug` (ariadne#194 M3). When a
	// family already has a disposed finding in a prior round, the gate escalates the
	// reviewer's recommendation from "fix this instance" to "state the rule that covers
	// all of them", which is the convergence mechanism the issue exists for.
	//
	// Deliberately FREE TEXT, not an enum: a family is DISCOVERED across rounds: the
	// point is to notice a recurrence nobody predicted, so a closed set would defeat it.
	// The binary normalizes on ingest (case, punctuation) and renders the in-play
	// vocabulary into the next round's prompt so the reviewer reuses a slug rather than
	// coining a synonym. Optional, so a transcript from before #194 M3 still parses.
	family?: string
}

// ── #Dispose: a later round's verdict on an earlier finding ──
#Dispose: {
	id:          string // a prior finding's binary-assigned id
	disposition: #Disposition
	note?:       string
}
