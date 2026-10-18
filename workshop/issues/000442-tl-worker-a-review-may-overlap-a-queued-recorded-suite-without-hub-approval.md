---
id: 000442
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1acced4e192a085747ee7a1d9ba9465e6c34de50' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T08:30:28-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "e6b524c5", done: "6acb0d5f"}
---

# tl-worker: a review may overlap a queued recorded suite without hub approval

## Problem

Slots ask the hub whether a milestone or close review may run while their recorded suite waits in the machine queue (pair:1 asked twice on 2026-10-18; queues ran 20+ minutes). The answer is always yes under two conditions, but the rule "run the full suite before a close" reads as forbidding it, so careful slots wait or ask.

## Spec

tl-worker help (the same planning or landing block as #432's draft-PR rule): a milestone or close review may overlap a queued recorded suite, provided (1) a recorded suite PASS on the final head is still required before landing, and (2) if the review changes code, the suite is requeued on the final head. No hub approval needed.

## Done when

- `tl-worker --help` carries the rule, phrase-guarded by its rules test.

## Plan

- [x] tl-worker help PERMISSIONS (beside #432's draft-PR rule): the overlap rule with both conditions; two `rules.tsv` rows (own scope)

## Log

### 2026-10-18
- Claimed by ariadne:2; quick flow. The rule sits after #432's draft-PR line in PERMISSIONS, phrase-guarded by `review-suite-overlap` and `review-suite-requeue` (own scope, so tl-lead must not restate it).
