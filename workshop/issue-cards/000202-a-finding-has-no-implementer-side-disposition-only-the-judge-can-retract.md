---
id: 000202
status: open
created: 2026-08-21
updated: 2026-08-21
estimate_hours:
github_issue:
---

# A finding has no implementer-side disposition: only the judge can retract

## Problem

`construct/vocabulary/finding.cue` models advisory-ness for **severity** and not
for **authority**.

Severity is handled well. `categories.advisory: ["Minor"]` never blocks;
`hardBlocking: ["Critical"]` means an `Important` finding is demoted past the
round cap; `when."Important"` even reads "fix before the gate **if cheap**;
blocks until disposed". Those are real, and they work.

But the closing dispositions are:

```cue
dispositions: {
  closing: ["addressed", "withdrawn"]
  open:    ["not-addressed"]
}
whenDisposed: {
  "withdrawn": "the judge retracts it (mistaken, or overtaken by a design change)"
}
```

**Only the judge can withdraw.** There is no disposition meaning *the implementer
considered this and declined, with reasons*. A grep for rebuttal / dispute /
agent-note across `cmd/`, `pkg/` and `construct/` returns nothing: there is no
channel for the implementer to answer a finding at all.

AGENTS.md §3 matches: *"Fix Critical/Important before crossing the boundary"* —
imperative, with no "or record a reasoned disagreement".

The doctrine we want already exists in the base layer, in the wrong place.
`superpowers-receiving-code-review` has a whole **"When To Push Back"** section
("Push back with technical reasoning if wrong"). §3 explicitly says NOT to run
those skills at an SDLC boundary — so the pushback doctrine lives precisely where
the gate is not.

### The consequence: findings time out instead of being adjudicated

A **wrong** finding and a **correct-but-not-worth-it** finding leave through the
same door — the round cap — and the ledger does not distinguish them. So the next
round's judge inherits no counter-evidence and can re-derive the same mistake.

Measured in `tools#4`, which ran ten boundary rounds and 41 findings:

- **BR-40** claimed "the byte path is asserted by nothing". False: that
  measurement ran `-run TestPTY -tags conformance`; against the full suite the
  same mutation reddens `TestEditorLoopCtrlCExitsZero` with `exit = 9, want 0`.
  The implementer had no disposition for "disputed, here is the
  counter-measurement", so the rebuttal went into the plan's prose — the channel
  **BR-28 in that same issue had already measured as 0% addressed**. It exited by
  round cap, not by adjudication, and nothing records that it was wrong.
- **BR-33** was accepted with a stated engineering reason (the obvious fix
  collides with a separately-pinned invariant). No disposition for that either;
  it remains listed under "Open findings" at close.
