---
id: 000340
status: open
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: '7ae5c55ff243e990993c161ea8622bf50b23656c' # card fields mirrored from issue-cards; edit via sdlc
---

# main is red: TestVerbContextsReachTrackerReads refuses #302's AssertCancelEndsReceiveSide (context.Background)

## Problem

Reported by ariadne:2 (10-12 ~14:0x): after #302 (PR #202, 47a4e463) landed, `TestVerbContextsReachTrackerReads` fails on a clean origin/main worktree: `internal/testfix/blockingorigin.go:59` (`AssertCancelEndsReceiveSide`) uses `context.Background`, which the context guard refuses. `make test` is red on main for every ariadne slot (alongside #338's failure).

## Spec

Make `AssertCancelEndsReceiveSide` take the caller's context (the test's), or otherwise satisfy the context guard's rule, rather than weakening the guard. Check #302's other new helpers for the same pattern.

## Done when

- `TestVerbContextsReachTrackerReads` passes on main; #302's pin still fails when its fix is reverted; the cause in the Log.

## Plan

- [ ]

## Log

### 2026-10-12
