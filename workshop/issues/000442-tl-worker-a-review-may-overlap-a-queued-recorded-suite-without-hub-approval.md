---
id: 000442
status: codecomplete
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd8158da14aebf3569654568c54a3d612678b6b1d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T08:30:28-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "e6b524c5", done: "6acb0d5f"}
actual_hours: 0.04
---

# tl-worker: a review may overlap a queued recorded suite without hub approval

## Problem

Slots ask the hub whether a milestone or close review may run while their recorded suite waits in the machine queue (pair:1 asked twice on 2026-10-18; queues ran 20+ minutes). The answer is always yes under two conditions, but the rule "run the full suite before a close" reads as forbidding it, so careful slots wait or ask.

## Spec

tl-worker help (the same planning or landing block as #432's draft-PR rule): a milestone or close review may overlap a queued recorded suite, provided (1) a recorded suite PASS on the final head is still required before landing, and (2) if the review changes code, the suite is requeued on the final head. No hub approval needed.

## Done when

- `tl-worker --help` carries the rule, phrase-guarded by its rules test.

## Plan

- [x] tl-worker help PERMISSIONS (beside #432's draft-PR rule): the overlap rule with both conditions; four `rules.tsv` rows (own scope), one per clause

## Log

### 2026-10-18
- 2026-10-18: closed — tl-worker --help carries the overlap rule in PERMISSIONS beside #432, each clause whole on one line and guarded by its own rules.tsv row (review-suite-overlap, -approval, -pass-first, -requeue); deleting any clause turns TestWorkerHelpCarriesEveryRule red (checked clause by clause). tl-worker and tl-lead tests pass. No atlas: a help-text rule, no new surface.; review verdict: SHIP
- Claimed by ariadne:2; quick flow. The rule sits after #432's draft-PR line in PERMISSIONS, phrase-guarded by `review-suite-overlap` and `review-suite-requeue` (own scope, so tl-lead must not restate it).
- Close review round 1 (FIX-THEN-SHIP): condition (1) had no guard, and a line wrap cut the approval phrase. Reflowed so each clause is whole on one line, with four own-scope rows (overlap, no approval, pass on the final head first, requeue). Checked: deleting any clause turns TestWorkerHelpCarriesEveryRule red.
