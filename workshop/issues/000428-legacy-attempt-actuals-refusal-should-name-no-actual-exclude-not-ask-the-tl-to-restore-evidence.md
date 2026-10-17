---
id: 000428
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '5e0e23dd3726724f0ff4421959a405664f4028bc' # card fields mirrored from issue-cards; edit via sdlc
---

# Legacy-attempt actuals refusal should name --no-actual (exclude), not 'ask the TL to restore evidence'

## Problem

## Spec

Since ariadne#420, `sdlc close`, `milestone-close` and `actual` refuse a numeric actual when a suite attempt in the window is legacy, meaning it has no exact `queued_at` and `acquired_at`. The refusal tells the worker to "ask the TL/hub to restore named measurement evidence". For a legacy attempt that evidence never existed and can't be restored. The approved disposition is to exclude the sample and close with `--no-actual` and a reason. On 2026-10-17 at least three slots stopped and asked the hub (parley.nvim:1 on #373 M1, pair:5 on #622 M2, ariadne:5 on #427), costing a round trip each.

Wanted: when every blocking attempt is legacy, the refusal says so and names the exact command, for example `sdlc milestone-close --issue N --milestone M --no-actual 'legacy suite attempt <id> lacks exact queue endpoints; excluded from calibration per ariadne#420'`. The "restore evidence" wording stays only for cases where evidence could exist but is missing or malformed. Keep the refusal; don't auto-apply `--no-actual`.

## Done when

- A test: a window whose only blocker is a legacy attempt gets a refusal naming `--no-actual` with the attempt ID and the reason text. A malformed new-format record still gets the restore wording.
- `close`, `milestone-close` and `actual --issue` give the same wording.

## Plan

- [ ]

## Log

### 2026-10-17
