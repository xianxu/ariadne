---
id: 000494
status: codecomplete
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T06:34:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.41
tracker:
    version: 1
    completion:
        token: close-f8411da90b82
        repository: github.com/xianxu/ariadne
        reviewed_head: 56a0a9c84276d7bb914673cec39fc6983424f55d
        evidence_commit: 9ba91bf9d2df046d3424ddc4f388990ada9aca03
---

# Train car's own merge warns 'active issue missing without confirmed archive' after the train already archived it

## Problem

In pair's first multi-car train (176, 10-20 05:48), car pair#717's own `sdlc merge` (PR 483) ended with "active issue
missing without confirmed archive", although the train had already archived the issue (it's under history on main
and its card is done). Harmless, but it reads as an error at the end of a successful landing, and a worker can't tell.
