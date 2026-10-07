---
id: 000297
status: open
deps: []
github_issue:
created: 2026-10-06
updated: 2026-10-06
estimate_hours:
card_mirror: 'b4535f4906f6d64ab608bc4f20eecde54d71dabc' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc: verification authority and an operator smoke-test state

## Problem

`codecomplete` blurs two different facts:
- the boundary review passed;
- the operator verified the work works.

They diverge in practice. In pair#387 the review passed, and the operator's live smoke
test still found a real gap: the remembered-failure digest ignored weave, fixed in
pair#387 itself.

SDLC records neither who may decide that a task is done nor whether an operator
smoke test is pending or happened. So a TL coordinating several slots cannot ask
"what is waiting for the operator to test?" The answer lives in chat transcripts.
The smoke-test steps a worker writes for the operator also live only in chat.

## Spec

Captured for operator review; no implementation is authorized by this issue
creation.

- **Verification authority, a card field: `verify: auto | operator`.**
  - It is set when work is assigned, by the TL or by the product-lead judgement
    (ariadne#298); the worker does not decide it at the end.
  - Unset means `operator`, the safe default (operator decision, 2026-10-06).
  - A worker may raise `auto` to `operator` when it finds risk mid-flight, and may
    never lower it.
  - It is changed through a verb or setter, as with the other card fields (#252),
    never by editing the mirror.
- **Smoke-test state is a field, not a new status.**
  - `verify` is orthogonal to the lifecycle. A `codecomplete` card with
    `verify: operator` and no recorded pass *is* "ready for smoke test".
  - Recording the operator's result (pass, or fail with a note) is a verb, and
    the record names its evidence. A fail returns the work to the worker.
- **The worker's handoff is a `## Smoke test` section in the issue body.**
  - It holds the exact steps and expected results the operator runs: the text
    workers currently write in chat.
  - It is required at close when `verify: operator`, and is the artifact the TL
    hands to the operator.
- **The merge gate.** `sdlc merge` refuses `verify: operator` work with no recorded
  pass. It has a precise `--no-<gate>` acknowledgment like the other gates, with
  the reason in the record.
- **The queue is one query.** "Assigned and awaiting smoke test" is a read-only
  listing across the tracker (#279-style observation), fast, from repo state alone.
- **State lives in the repository.** Messages between slots are hints, or later a
  way to resolve conflicting evidence, never the record (operator, 2026-10-06).
- **Out of scope.**
  - Where a smoke test runs: in place for most products; on the repo's `:0` bench
    for developer tools that build the apps.
  - A queue for that bench (deferred until it hurts).
  - Release and SRE stages (see the vision note
    `docs/vision/2026-10-06-01-pensive-before-and-after-the-sdlc.md`).

## Done when

- A `verify` card field defaults to operator when unset, is set at assignment, and
  can be raised but not lowered. Each rule has a test.
- Closing `verify: operator` work requires a `## Smoke test` section. Recording the
  operator's pass or fail is a verb whose effect is on the card, with a test for
  each outcome.
- `sdlc merge` refuses unverified `verify: operator` work and names the next action.
  `--no-<gate>` is honoured and recorded.
- One read-only query lists the work awaiting a smoke test, with each item's owner
  slot and smoke-test section, from repository state alone.

## Plan

- [ ]

## Log

### 2026-10-06

Filed from pair#362's brainstorm (cross-slot-work-scheduling). Operator decisions:
- close authority is set at assignment and defaults to operator;
- smoke-test state is a field, not a status;
- state lives in the repository, and messages are hints;
- the bench queue is deferred.
Consumers: pair#396 (TL-driven project execution) and ariadne#298 (product-lead skill,
which sets `verify`). Details left local for the operator to refine.
