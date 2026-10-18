---
id: 000436
status: done
created: 2026-10-17
updated: 2026-10-18
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
actual_hours: 0.02
tracker:
    version: 1
    completion:
        token: close-18cb6c1c2163
        repository: github.com/xianxu/ariadne
        reviewed_head: 00be65f9fe57f03a85b0b3c76df0827877f2bbd9
        evidence_commit: 43a86f4b3c3bebe5cf15e0dd368d70a2ae4202a9
        landed_commit: 72ae62473c2e732d746b7a2a6e0092f39a8f0c59
---

# tests: the remaining pid-file wait loops use testfix.AwaitPIDFile

## Problem

The #434 close review found hand-written pid-file wait loops outside cmd/sdlc that duplicate `testfix.AwaitPIDFile`: `internal/processgroup/process_unix_test.go` (2), `internal/gitx/lifetime_test.go` (2), `internal/gitx/rebasedpatch_context_test.go` (1), `internal/judge/cancellation_unix_test.go` (1). One helper keeps one readiness idiom (ARCH-DRY).
