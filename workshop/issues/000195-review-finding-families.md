---
id: 000195
status: open
deps: []
github_issue:
created: 2026-08-20
updated: 2026-08-20
estimate_hours:
---

# boundary reviews: carry findings across rounds, tag families, escalate on repeat

## Problem

A boundary review that returns REWORK is re-run after the fixes. Each run starts
from nothing: findings are numbered `C1`, `C2`, `I1`… fresh every round, and the
sidecar appends the new round below the old one. Nothing connects round N's
findings to round N-1's.

**The plan gate already solved this and the boundary review did not inherit it.**
`gatestate.Ledger` (`cmd/sdlc/planreview.go:44,66`, `cmd/sdlc/changecode.go:589`)
gives plan-quality findings binary-assigned stable ids and requires each re-run to
dispose of every prior one (`addressed` / `not-addressed` / `withdrawn`). The
boundary review has no equivalent. That asymmetry costs two distinct things.

### 1. No convergence signal

`tools#1` ran four M1 rounds plus two close rounds. Every round found real,
live-verified defects, so no round was arguable — but there was no way to tell
whether round 5 would find more, or whether the findings were shrinking toward a
fixed point. `WF_PLAN_ROUND_CAP` bounds the plan gate; nothing bounds this.

### 2. The same finding FAMILY recurs, and nobody says so

The sharper cost. Across `tools#1`, one underlying rule was missing — *"a
part-of-speech word only opens a block when structurally placed"* — and it
surfaced three times in three shapes:

| round | shape | what got fixed |
|---|---|---|
| M1 r1 | `(banked as adjective)` — POS before a closing paren | the trailing-boundary set |
| M1 r2 | `[with adjective or noun modifier]` — POS inside a bracket | bracket-depth tracking |
| M1 r2 | `a noun phrase functioning as` — POS in plain prose | **the rule** (`opensBlock`) |

Each round I fixed the *instance*. Only on the third did I write the rule. A
human reviewer would have said "you are patching cases" on the second one — and
the same pattern repeated in the close review: `]` as a block opener was a fourth
instance of that same family, found one round later still.

Same story on a second family: an oracle that cannot see the thing it certifies.
Round 3 found the raw-notation check asking `isPronunciation` to grade its own
output; round 4 found `strayStress` was honest but too narrow; the close review
found the no-data-loss invariant was one-directional. Three rounds, one family:
*the measurement cannot fail in the direction that matters*.

**Estimate: family escalation would have collapsed at least two of `tools#1`'s
four M1 rounds.**

## Spec

Two changes, one mechanical and one to the reviewer's prompt.

### A. Give boundary reviews the ledger the plan gate already has

Reuse `gatestate.Ledger` rather than writing a second one (ARCH-DRY — it already
models exactly this: stable ids, per-round dispositions, carry-forward). A
boundary re-run then:

- reads the prior rounds' findings,
- must dispose of each before raising new ones,
- and cannot silently renumber, so "is this new?" is answerable mechanically.

### B. Tag a `family:` on each finding, and escalate on repeat

A short slug naming the *underlying rule*, not the symptom — `block-opener-rule`,
`oracle-blind-direction`, `doc-claim-exceeds-measurement`. Assigned by the
reviewer, carried in the ledger.

When a finding's family already has ≥1 disposed finding in a prior round, the
reviewer is instructed to say so explicitly and change its recommendation:

> **This is the 3rd finding in family `block-opener-rule`.** Rounds 1 and 2 fixed
> instances. Do not fix this instance — state the rule that covers all three, and
> fix that. If the rule cannot be stated, say why, and record the family in
> `Limits` with its measured prevalence.

That instruction is the substance of this issue. The ledger exists to make it
mechanical rather than dependent on the reviewer noticing across a document it
was never shown.

### C. Convergence line in the verdict

One line per re-run, cheap and human-readable:

> round 4 — 2 new findings, 0 repeat families, 6 disposed. **Converging.**
> round 2 — 3 new findings, 2 repeat families. **Not converging: fix rules.**

This is the signal that was missing. It also gives an honest basis for a round
cap: capping on *count* is arbitrary, capping when families stop repeating is not.

## Done when

- [ ] Boundary-review findings have stable ids and survive across rounds, using
      `gatestate.Ledger` rather than a second implementation.
- [ ] A re-run must dispose of every prior finding before raising new ones.
- [ ] Each finding carries a `family:` slug.
- [ ] A repeat family changes the reviewer's recommendation from "fix this" to
      "state the rule", and the verdict names the family and the repeat count.
- [ ] The verdict carries a one-line convergence summary.
- [ ] Verified against a real multi-round history — `tools#1`'s
      `000001-define-m1-review.md` has four rounds and two clear families, so it
      is a ready-made fixture: a correct implementation flags
      `block-opener-rule` at round 2, not round 3.

## Plan

- [ ] Design via `sdlc start-plan` before implementing.

## Log

### 2026-08-20

Filed alongside `#194` after a session running 8 boundary reviews across two
issues in `tools`. Complementary problems: `#194` is about the review blocking
the tree, this is about the review having no memory.

Worth stating plainly — **the reviews were right every time**. Every round found
defects that reproduced live. The failure is not verdict quality; it is that a
stateless reviewer cannot see that it is describing the same missing rule for the
third time, so the author keeps being told to fix instances.

The plan gate's `#187` stateful ledger is the precedent, and the argument for it
applies unchanged one stage later.
