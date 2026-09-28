---
id: '000195'
status: punt
created: 2026-08-20
updated: 2026-08-20
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
