---
id: 000620
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'b6e2dd8d26fb34c2334b2ff36ae9147af585b9a9' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestKnownFlakeGate

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestKnownFlakeGate: failed in attempt 2a3963a2f77b656a3bec32d24a99ca80, passed on the same head in attempt 0c0df556fe88758c8a8562c1ee90f7ba

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-24
