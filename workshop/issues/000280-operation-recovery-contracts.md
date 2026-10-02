---
id: 000280
status: working
deps: [ariadne#277, ariadne#278, ariadne#279]
github_issue:
created: 2026-10-01
updated: 2026-10-02
estimate_hours:
card_mirror: '9717bec607118ffcf4d936794f28518aacce5efb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-02T10:21:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Publish tested operation recovery contracts

## Problem

A lossy multi-step request channel is useful if agents can verify effects and safely retry operations. Today those guarantees are not a discoverable, tested SDLC contract.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for operator review; no implementation is authorized by this issue creation.

Publish operation-level recovery contracts alongside the binaries/subcommands that own them. For claim, reclaim, planning/implementation entry, milestone/close and publication, enumerate observable effects, repetition behavior, scope/preconditions, lost-response handling and when guarantees cease. Do not label every command idempotent: distinguish read-only, duplicate-safe refusal, convergent retry and non-repeatable/uncertain effects.

Existing claim CAS and TestClaimRaceHasExactlyOneWinner are initial evidence. Repeated claim is safe against an already claimed issue, but later reopening/reclaiming changes the context. Arbitrary subsequent code edits are not made idempotent by the claim. Git transaction receipts and revisions should be reused where they already establish outcomes.

Put claim-before-work and owner-checked continuation in SDLC once, not in every scheduling prompt. Agent guidance should verify effects via read-only commands, treat a roughly 30-second interval as a revisit heuristic rather than proof of loss, and retry only under the operation contract. Unknown observations are not negative evidence.

## Done when

- Agent-discoverable help/docs define recovery behavior and observable evidence for the MVP workflow commands, linked to supporting implementations/tests.
- Race, duplicate, lost-ack and reopened/reclaimed-generation cases validate the declared guarantees.
- A scheduling example verifies claim, branch/activity and gate progress without restating defensive claim logic in every message.
- Unproven guarantees are explicitly unknown; documentation does not claim exactly-once execution of arbitrary agent instructions.

## Plan

Implementation plan to be designed after issue claim and start-plan; these are requirements, not an approved implementation plan.

## Log

### 2026-10-01

Captured from the performance → messaging guarantees → SDLC ownership/observability → recovery discussion. No implementation started.

### 2026-10-02 (implementation session)

- Operator authorized the work ("work on #280"). Claimed in ariadne:1 (the
  claimant is recorded) and ran start-plan; the branch sits at main, which
  includes #277, #278 and #279. Mapping where each MVP command's recovery
  behavior lives today (CAS, receipts, recovery reconcile, observe) before
  designing the contract surface.

