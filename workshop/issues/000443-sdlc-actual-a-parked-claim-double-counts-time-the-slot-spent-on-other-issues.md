---
id: 000443
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '112a88c64ce5127589fe38b48fcf0f45633dabe0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T10:57:47-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc actual: a parked claim double-counts time the slot spent on other issues

## Problem

A slot that holds a claim on one issue while it works others gets the same time counted twice. pair:6 held pair#636 from 2026-10-17 10:14 to 2026-10-18 08:41, parked behind pair#647, and meanwhile worked pair#649 and pair#656, which each adopted their own measured actuals. `sdlc actual --issue 636` then reported 9.28 h over 11 admitted sessions, mostly time already counted toward #649 and #656. The slot had to close with --no-actual.

## Spec

Active time a slot spends is attributed once. When measuring an issue, exclude intervals that are already attributed to another issue's adopted actual from the same slot, or, more simply, attribute each interval to the issue whose branch or commits it touched. Pick the rule that needs the least new evidence and say why. A parked claim (owner kept, no work) should accrue nothing while its slot works something else.

## Done when

- Measuring an issue whose claim overlaps another adopted issue from the same slot does not count the overlap twice; a test reproduces the #636 shape.
- The refusal or report names the excluded intervals and why.

## Plan

- [ ] Use another issue's adopted actual and pinned measurement inputs as exclusion authority, never a live claim or an unscoped main commit. For each overlapping same-slot adopted issue, replay its pinned inputs, take the source-scoped activity runs credited to that issue, and subtract their time interval from the candidate before attribution. An ambiguous/missing pin refuses a numeric actual rather than silently counting overlap. The exclusion is conservative when a credited run was split fractionally among issues; report its issue, slot and interval.
- [ ] Add a red #636-shaped integration fixture: a parked issue and two intervening adopted issues in the same slot. Assert preview, close adoption and correction share the rule; no credited interval is counted again. Exercise differing slots, missing historical evidence and replay after a later transcript grows.
- [ ] Implement the pure interval subtraction in `internal/activetime`, the pinned adopted-issue evidence loader in the `actualinputs` IO seam, and concise interval/refusal diagnostics. Preserve current branch commit scoping (#270).
- [ ] Run targeted tests, `make -k`, the recorded full suite on the pushed head, then close review and land on SHIP.

## Log

### 2026-10-18

- Planning choice: branch-only attribution cannot guarantee exclusivity because #270 intentionally hides unrelated main commits from a branch's segment boundaries. Reusing completion-pinned actual inputs needs no new durable ledger (ARCH-DRY, ARCH-FUNERAL). A replay gap remains unknown, not zero (ARCH-PURE). This can conservatively undercount activity where two issues genuinely overlap in the same slot; the report names every excluded interval.
