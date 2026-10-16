---
id: 000410
status: open
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
github_issue:
---

# tl-worker/tl-lead: a state design table for stateful plans (actions, failure modes to few visible states, one recovery each, explicit loss, tradeoff to TL/operator)

## Problem

Over-engineered state designs (pair#556/#584, #587) and under-engineered recovery have one root: plans for state that
users or automation can change don't state their action set, their failure modes, the recovery for each, or what state
can be lost. The operator (2026-10-16): no model is immune; adjust the prompt. A design that changes such state must
present the tradeoff explicitly, to the TL and often the operator, because it is almost always a tradeoff.
