---
id: 000340
status: working
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: '21805f03d1959df16967a420619371bd757bfd2c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T15:55:00-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# main is red: TestVerbContextsReachTrackerReads refuses #302's AssertCancelEndsReceiveSide (context.Background)

## Problem

Reported by ariadne:2 (10-12 ~14:0x): after #302 (PR #202, 47a4e463) landed, `TestVerbContextsReachTrackerReads` fails on a clean origin/main worktree: `internal/testfix/blockingorigin.go:59` (`AssertCancelEndsReceiveSide`) uses `context.Background`, which the context guard refuses. `make test` is red on main for every ariadne slot (alongside #338's failure).

## Spec

Make `AssertCancelEndsReceiveSide` take the caller's context (the test's), or otherwise satisfy the context guard's rule, rather than weakening the guard. Check #302's other new helpers for the same pattern.

## Done when

- `TestVerbContextsReachTrackerReads` passes on main; #302's pin still fails when its fix is reverted; the cause in the Log.

## Plan

- [x] `AssertCancelEndsReceiveSide` derives its cancel context from `t.Context()` (the caller's) instead of `context.Background()`; the guard is left as it is.
- [x] Sweep #302's other helpers (`internal/testfix` non-test files) for `context.Background`/`TODO`: none remain.
- [x] Verify: the guard passes; both #302 pins pass, and fail with the group kill reverted; `make test`.

## Log

### 2026-10-12
- Cause: #302's close-review fix-up (47a4e463) moved the cancel assertion out of the two `_test.go` files into `internal/testfix/blockingorigin.go`. That is a non-test file, which is exactly what `TestVerbContextsReachTrackerReads` scans (every `internal/*` package minus `_test.go`). The `context.Background()` it brought along was legal in a test file and refused in the helper. #302's `make test` runs all predate that commit, and its re-close cited "prior evidence stands", so nothing ran the guard on the moved code.
- Fix: `context.WithCancel(t.Context())`. The test's context is the caller's context the guard asks for, and it is cancelled at test end anyway. The guard is unchanged.
- Evidence: `TestVerbContextsReachTrackerReads` passes. `TestTrackerGitCancelEndsTheReceiveSide` (cmd/sdlc) and `TestGitCancelEndsTheReceiveSide` (gitx) pass, and with `processgroup.Terminate` temporarily signalling only the direct child (no group kill) both fail with "the receive side outlived the cancelled push: pid … still running after 2m0s"; restored afterwards. The `gitx` and `testfix` packages pass.
