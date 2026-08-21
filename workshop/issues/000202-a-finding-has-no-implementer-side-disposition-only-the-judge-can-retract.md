---
id: 000202
status: open
deps: []
github_issue:
created: 2026-08-21
updated: 2026-08-21
estimate_hours:
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

## Spec

Add an implementer-side closing disposition — `disputed` is the working name —
with these properties:

- **It is `closing`, not `open`.** It stops the finding blocking. `finding.cue`'s
  own comment already explains why dispositions are partitioned rather than flat:
  so a new one cannot be accepted by the round validator while the open-set
  computation has no case for it. This change is the shape that comment
  anticipates (it uses `deferred` as its example).
- **It requires evidence, not an opinion.** A rationale is mandatory, and where
  the finding shipped with a measurement, so must the disposition — the rule
  `tools#4` landed in lessons.md is *a finding is closed only when you have
  re-run the measurement that produced it*, and a dispute is the case where
  re-running it produced a different answer. A `disputed` with no counter-measure
  should be a protocol error, exactly as a disposition naming an unknown id
  already is.
- **The next judge must see it.** Carry the rationale into the following round's
  prompt so the judge either re-raises **against the evidence** or withdraws. The
  current failure mode is that a rebuttal is invisible to the next round.
- **It is loud.** Per the escape-hatch principle: every structure needs a bypass,
  and using it must be announced, never silent. It belongs in the churn ledger as
  its own column beside `addressed` / `withdrawn` (`churnreport.go` already treats
  a closing disposition with no column as a defect), and on the operator gate
  line.

## Design risks

- **Abuse.** An agent that disputes everything escapes every gate. This is the
  main reason to require a counter-measurement and to make the count visible per
  boundary: a round that disputes 6 of 8 findings should look alarming on the
  gate line. Consider whether `Critical` should be disputable at all, or whether
  a dispute on a `hardBlocking` finding needs operator acknowledgement.
- **It must not become the lazy exit.** `tools#4` is also the evidence *for* the
  gate: rounds 7-10 caught a 9.6 MB blob, a guard that read the index instead of
  history, and a guard that reported green on a partial scan — all self-inflicted
  by the implementer, all correctly refused. A dispute channel must not make that
  refusal cheaper to dodge. The asymmetry to encode: disputing costs evidence,
  fixing costs a diff.
- **Judge learning.** Worth deciding whether a sustained dispute should feed back
  into the judge prompt at all, or stay per-issue.

## Related

- **#195** — finding-family escalation. Same underlying problem (findings that
  recur without converging) from the other side: #195 is about the judge
  escalating, this is about the implementer answering.
- **#201** — review artifacts capture harness stderr. Also surfaced by `tools#4`.
- `construct/vocabulary/finding.cue` is the single source; `cmd/sdlc/boundaryledger.go`,
  `cmd/sdlc/churnreport.go` and the judge prompt are the consumers that must derive
  from it rather than restate it.
