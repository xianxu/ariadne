---
id: 000436
status: codecomplete
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c4885ec9cfefa288d62d8048caf7b8bfce310c90' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T23:09:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "6e340b9c", done: "2f0e5866"}
actual_hours: 0.02
---

# tests: the remaining pid-file wait loops use testfix.AwaitPIDFile

## Problem

The #434 close review found hand-written pid-file wait loops outside cmd/sdlc that duplicate `testfix.AwaitPIDFile`: `internal/processgroup/process_unix_test.go` (2), `internal/gitx/lifetime_test.go` (2), `internal/gitx/rebasedpatch_context_test.go` (1), `internal/judge/cancellation_unix_test.go` (1). One helper keeps one readiness idiom (ARCH-DRY).

## Spec

## Done when

- Every single-file pid wait loop in cmd/sdlc/internal tests uses `testfix.AwaitPIDFile`; the touched tests pass.

## Plan

- [x] Convert processgroup (2), gitx lifetime (2), gitx rebasedpatch (1), judge `waitReviewChild` (1)


## Log


- 2026-10-18: closed — sdlc-local/suite green on 00be65f9 (exit 0, 10m7s). Six pid-file wait loops in internal/processgroup (2), gitx lifetime (2), gitx rebasedpatch (1) and judge waitReviewChild (1) now use testfix.AwaitPIDFile; processgroup, judge and the touched gitx tests pass.; review verdict: SHIP
### 2026-10-17

- Converted six loops; left the judge child's `exiting-parent` wait (it waits for any content, not a pid, inside the
  child process). processgroup, judge and the touched gitx tests pass.
- Close review: SHIP. Removed the stray empty Plan item. Kept: `waitReviewChild` now retries a non-numeric pid file until the ceiling instead of failing at once (tolerates a partial write; the reviewer judged it acceptable).
