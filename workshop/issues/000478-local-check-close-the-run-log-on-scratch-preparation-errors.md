---
id: 000478
status: open
deps: [ariadne#477]
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
severity: low
collaboration-mode: collaboration
card_mirror: 'd19634ca4163790d89555cb0ed8fcc00a1027d19' # card fields mirrored from issue-cards; edit via sdlc
---

# local-check: close the run log on scratch preparation errors

## Problem

The third close review of ariadne#477 recorded Minor BR-3: runUnderSlot opens logf before prepareScratch, and its setup-error return bypasses the explicit Close later in the function. CLI exit eventually closes the descriptor, but the function should release its own resource on every exit.

## Spec

- Scope ownership of the run log immediately after successful creation so setup errors release it without waiting for GC or process exit; retain normal write/close error reporting.
- Audit this function's early exits for the same resource-handle lifetime class.
- Add focused regression evidence for setup failure and normal completion, using private fixtures rather than the live sandbox.
- Evaluate a narrow Go lint guard for this mechanical resource-ownership class, following the worker contract.

## Done when

- A regression that fails before the fix proves the run log is closed when scratch preparation fails.
- Normal completion, timeout and cancellation controls still pass.

## Plan

- [ ]

## Log

Filed unclaimed from #477's SHIP review advisory (resource-handles-close-on-every-exit). No implementation is included in #477's final reviewed code.
