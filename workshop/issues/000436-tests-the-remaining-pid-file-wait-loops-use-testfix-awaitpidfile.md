---
id: 000436
status: open
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '44d0b3433b4f2aea83318cd744572f04505f1565' # card fields mirrored from issue-cards; edit via sdlc
---

# tests: the remaining pid-file wait loops use testfix.AwaitPIDFile

## Problem

The #434 close review found hand-written pid-file wait loops outside cmd/sdlc that duplicate `testfix.AwaitPIDFile`: `internal/processgroup/process_unix_test.go` (2), `internal/gitx/lifetime_test.go` (2), `internal/gitx/rebasedpatch_context_test.go` (1), `internal/judge/cancellation_unix_test.go` (1). One helper keeps one readiness idiom (ARCH-DRY).

## Spec

## Done when

- Every single-file pid wait loop in cmd/sdlc/internal tests uses `testfix.AwaitPIDFile`; the touched tests pass.

## Plan

- [ ]

## Log

### 2026-10-17
