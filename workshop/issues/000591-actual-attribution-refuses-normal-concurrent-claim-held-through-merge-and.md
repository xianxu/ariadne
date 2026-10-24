---
id: 000591
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '9920f8e166d3a85e39757b232e5d81f415e6cb2b' # card fields mirrored from issue-cards; edit via sdlc
---

# Actual attribution refuses normal concurrent claim held through merge and resting train checkout

## Problem

Follow-up to ariadne#580, reported by ops:0 from pair:4 closing pair#856. Actual now refuses session 062e2184-2117-4c26-b929-ceec3b5a8030 in pair-slot4: concurrent claims have unknown, resting or ambiguous active branch during 2026-10-23T21:35:40-07:00 to 2026-10-24T04:36:48Z. The slot legitimately held pair#850 while merging (its branch, then resting checkout during the train) and pair#856. Earlier close rounds measured 2.72 h and 3.33 h. The hub authorized --no-actual for that close meanwhile.

## Spec

Diagnose the real ownership and per-worktree HEAD history across ordinary hold-while-merging/train transitions. Attribute supported activity without double counting; preserve explicit uncertainty when evidence cannot select an owner. Do not invent historical evidence or modify adopted snapshots. Keep #564's code on hold pending operator review.

## Done when

- A captured fixture of the reported merge/train resting transition reproduces the refusal before the fix and verifies the agreed attribution afterward.
- Concurrent-claim selection remains disjoint, and genuinely missing/ambiguous branch evidence and retained-snapshot corruption retain explicit safe behavior.
- Targeted regression/race checks, recorded pre-close and boundary review pass; semantics and limits are documented.

## Plan

- Investigate evidence and send the root cause and proposed attribution boundary to the hub before code.

## Log

2026-10-24: Dispatched by ops:0 as a #580 follow-up, prioritized before #564 implementation. Filing claims the issue in ariadne:3; diagnosis only so far.
