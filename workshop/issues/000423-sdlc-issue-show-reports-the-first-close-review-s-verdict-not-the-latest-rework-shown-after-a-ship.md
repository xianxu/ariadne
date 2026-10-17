---
id: 000423
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e8cc600d2fa8c37d930a7c46e57847372725e28d' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc issue show reports the first close review's verdict, not the latest (REWORK shown after a SHIP)

## Problem

## Spec

`sdlc issue show` reports an issue's review verdict from the first close review, not the latest one. Four slots reported it independently:
- #408;
- ariadne#418 (first FIX-THEN-SHIP, then SHIP);
- pair#617 (round 1 shown after a round 2 SHIP);
- ariadne#416 (first REWORK shown after a round 2 SHIP).

The completion binding, the ledger and the landed status are all correct. Only the readout picks the wrong review.

A reader of `issue show`, whether the TL, the operator or an agent deciding whether to land, sees a stale REWORK or FIX-THEN-SHIP on an issue that shipped.

## Done when

- `issue show`, text and JSON, reports the verdict of the review bound to the issue's completion, or the latest review when there is no completion, along with its round. A test covers an issue with a REWORK round 1 and a SHIP round 2.
- Earlier rounds stay visible as history, labelled as earlier.

## Plan

- [ ]

## Log

### 2026-10-17
