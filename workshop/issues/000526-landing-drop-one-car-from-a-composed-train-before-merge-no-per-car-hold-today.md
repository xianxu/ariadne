---
id: 000526
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e2a065976b0fde5df5630902fc545f76ee743b66' # card fields mirrored from issue-cards; edit via sdlc
---

# landing: drop one car from a composed train before merge (no per-car hold today)

## Problem

## Spec

There is no supported way to drop one car from a composed landing train. 10-21 07:15, pair train 234 (#768, #767,
#765): pair:2 found a late test-only defect in #768 and asked to hold its car; pair:9 (the conductor) found that cars
are pinned at build and merge rechecks only main and the train head, so pushing a fix or drafting the car's PR doesn't
reliably exclude its old commit. The only supported options were abandoning the whole train (requeueing the clean
cars) or closing the train PR mid-flight. The TL let the train land with #768 as is (the defect was test-portability
only), but a real "this car must not land" case would cost the whole train.

Add `sdlc landing drop --issue N` (the car's owner or :0): before the train's merge, remove that car, recompose the
train head from the remaining cars (reusing #521's re-formation path), and rerun only what the recomposition
invalidates. After merge has started, refuse with the exact state.

## Done when

- Dropping one car of three before merge lands the other two with a recomposed head; the dropped car returns to
  waiting with its reason.
- After merge start, drop refuses and names what happened.

## Plan

- [ ]

## Log

### 2026-10-21
