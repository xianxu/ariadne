// verdict — the vocabulary of a boundary-review verdict: the tokens a fresh-context
// reviewer may emit, their semantics, and the structured shape of the handoff back
// to the binary. Single source of truth for the `verdict` noun (ariadne#147). sdlc
// reads the exported JSON; the review prompt renders the emitted set from it; the
// parser validates against it. The `agent-binary-handoff-schema` target carries the
// *why*: every stochastic→deterministic handoff crosses a schema, never prose.
package verdict

import "list"

// ── categories: the single concrete source of verdict membership ──
// Stated once, here. #Emitted / #Token are DERIVED via or(), so there is nothing
// to keep in sync. Only `categories` (concrete data) reaches the exported JSON —
// CUE definitions (#) do not export. The three categories PARTITION the token set:
// every token is in exactly one (a conformance test pins this).
categories: {
	finalizing: ["SHIP", "FIX-THEN-SHIP"] // an acceptable verdict — the gate may finalize
	blocking:   ["REWORK"]                // do not cross the boundary; rework + re-run
	internal:   ["not-run", "unknown"]    // system-set, NEVER emitted by the reviewer
}

// #Emitted: the set a reviewer may emit in the structured block (finalizing + blocking).
#Emitted: or(list.Concat([categories.finalizing, categories.blocking]))

// #Token: every verdict value, including the system-internal ones the binary sets.
#Token: or(list.Concat([categories.finalizing, categories.blocking, categories.internal]))

// ── when: one-line semantics per token (the documented-value source) ──
when: {
	"SHIP":          "ready; ship it"
	"FIX-THEN-SHIP": "ship after addressing the findings (non-blocking at the gate)"
	"REWORK":        "blocking; needs rework before shipping — fix + re-run"
	"not-run":       "review skipped or errored (system-set; no judgment available)"
	"unknown":       "review ran but emitted no schema-valid verdict (system-set; investigate)"
}

// ── discovery: where instances of this noun live. The review sidecar (#136) carries
// the validated verdict in its frontmatter, so a sidecar IS a verdict instance.
// Repo-relative (the consumer joins to its repo root). ──
discovery: {
	home: "workshop/plans"  // the durable review sidecars + plans
	glob: "*-review.md"     // close-review / m<x>-review sidecars
}

// ── #Verdict: the structured handoff the reviewer emits + the binary validates ──
// A fenced ```verdict block (and, mirrored, the sidecar frontmatter). Closed
// (fail-closed) — a verdict is an atomic judgment, not an organically growing record.
#Verdict: {
	verdict:     #Emitted                     // the reviewer-emitted token
	confidence?: "high" | "medium" | "low"    // optional self-assessment
}
