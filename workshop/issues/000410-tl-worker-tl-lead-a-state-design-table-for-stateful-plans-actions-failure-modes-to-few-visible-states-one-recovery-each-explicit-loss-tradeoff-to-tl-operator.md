---
id: 000410
status: open
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '6d8cb2e39c51e4ad04c19074e0a7b2b457a769e1' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-worker/tl-lead: a state design table for stateful plans (actions, failure modes to few visible states, one recovery each, explicit loss, tradeoff to TL/operator)

## Problem

Over-engineered state designs (pair#556/#584, #587) and under-engineered recovery have one root: plans for state that
users or automation can change don't state their action set, their failure modes, the recovery for each, or what state
can be lost. The operator (2026-10-16): no model is immune; adjust the prompt. A design that changes such state must
present the tradeoff explicitly, to the TL and often the operator, because it is almost always a tradeoff.

## Spec

tl-worker's planning rules (and tl-lead's plan review) gain a **state design table** for any plan that adds or changes
state a user or another system can act on:
1. **The actions:** what the user or automation can do (relaunch, resume, reboot, cancel, edit, retry…).
2. **The failure modes:** internal states, honestly including combinations, mapped onto a **small set of visible
   states**.
3. **Each visible state's one recovery action,** chosen from the action list.
4. **What can be lost** in each recovery (state, work, history), stated explicitly.
5. **The tradeoff summary** for the TL (and the operator when user-facing): what's covered, what isn't, and why.
The TL brings user-facing tradeoffs to the operator before code. Phrase-guarded in the help tests.

## Done when

- `tl-worker --help` and `tl-lead --help` carry the table and the escalation rule; help tests guard the phrases.

## Plan

- [ ]

## Log

### 2026-10-16
