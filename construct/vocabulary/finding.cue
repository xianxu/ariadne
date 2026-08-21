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

#Severity: or(list.Concat([categories.blocking, categories.advisory]))

// ── dispositions: what a LATER round may say about an EARLIER finding ──
// PARTITIONED by the semantics the binary branches on, not a flat list. A flat list plus a
// prose gloss would put the closes-vs-leaves-open decision in a Go switch — the exact
// posture this file exists to reject. Concretely: adding `deferred` to a flat list would
// make the round validator accept it while the open-set computation had no case for it,
// wedging the finding open forever.
dispositions: {
	closing: ["addressed", "withdrawn"] // the finding is settled; it stops blocking
	open: ["not-addressed"]             // still open; it keeps blocking
}

#Disposition: or(list.Concat([dispositions.closing, dispositions.open]))

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
	title:    string
	detail?:  string
}

// ── #Dispose: a later round's verdict on an earlier finding ──
#Dispose: {
	id:          string // a prior finding's binary-assigned id
	disposition: #Disposition
	note?:       string
}
