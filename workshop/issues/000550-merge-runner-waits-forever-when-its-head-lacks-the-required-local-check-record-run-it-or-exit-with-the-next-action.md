---
id: 000550
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3985b5eb51ec47f48c50ea3bb5304d5bd94f95e4' # card fields mirrored from issue-cards; edit via sdlc
---

# merge runner waits forever when its head lacks the required local-check record; run it or exit with the next action

## Problem

## Spec

pair#779 (10-21 19:02 → 10-22 05:00): its retained `sdlc merge --yes` runner sat first in pair's landing line for ten
hours, repeatedly logging "train due; not ready: PR535 needs recorded local checks on its head: affected or its
full-suite fallback", while three later changes landed past it. Nobody acted: the slot believed admission would run
the check, and the slot was idle (its turn had ended), so the repeated message was never read. Train readiness
excludes a car with no recorded local check **before** admission, so the runner can never make progress by waiting.

Fix: when the only thing keeping a retained merge runner from readiness is a missing local-check record on the exact
pushed head, the runner starts that check itself (the same `sdlc local-check run` under the normal queue), records it,
and proceeds; or, if policy forbids that, it exits with a clear next action rather than waiting forever. Either way
it doesn't loop silently.

## Done when

- A merge runner whose head lacks the required local-check record runs (or clearly demands) it instead of waiting
  indefinitely (test with a fake queue/check runner).
- A head that already has the record is unaffected (test).

## Plan

- [ ]

## Log

### 2026-10-22
