---
id: 000592
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '9389040f42f529655ce0e366492d1c4665821572' # card fields mirrored from issue-cards; edit via sdlc
---

# flake: TestQueueRealHelperPause

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestQueueRealHelperPause: failed in attempt c5fff316216ec03a781b4796f770a3e5, passed on the same head in attempt f3122a02c4fdbbaba2e5dd1f03b2b89f

## Spec

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [ ]

## Log

### 2026-10-24
