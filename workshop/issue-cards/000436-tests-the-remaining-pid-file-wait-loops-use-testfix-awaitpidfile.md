---
id: 000436
status: open
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T23:09:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# tests: the remaining pid-file wait loops use testfix.AwaitPIDFile

## Problem

The #434 close review found hand-written pid-file wait loops outside cmd/sdlc that duplicate `testfix.AwaitPIDFile`: `internal/processgroup/process_unix_test.go` (2), `internal/gitx/lifetime_test.go` (2), `internal/gitx/rebasedpatch_context_test.go` (1), `internal/judge/cancellation_unix_test.go` (1). One helper keeps one readiness idiom (ARCH-DRY).
