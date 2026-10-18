---
id: 000442
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '625048eed86c961752635d82432e45e19875954d' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-worker: a review may overlap a queued recorded suite without hub approval

## Problem

Slots ask the hub whether a milestone or close review may run while their recorded suite waits in the machine queue (pair:1 asked twice on 2026-10-18; queues ran 20+ minutes). The answer is always yes under two conditions, but the rule "run the full suite before a close" reads as forbidding it, so careful slots wait or ask.

## Spec

tl-worker help (the same planning or landing block as #432's draft-PR rule): a milestone or close review may overlap a queued recorded suite, provided (1) a recorded suite PASS on the final head is still required before landing, and (2) if the review changes code, the suite is requeued on the final head. No hub approval needed.

## Done when

- `tl-worker --help` carries the rule, phrase-guarded by its rules test.

## Plan

- [ ]

## Log

### 2026-10-18
