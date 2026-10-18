---
id: 000443
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7e923b81546513748d421d2304d614f9d4ef94da' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-18
