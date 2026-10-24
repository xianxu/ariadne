---
id: 000583
status: done
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
github_issue:
started: 2026-10-23T20:24:51-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.15
tracker:
    version: 1
    completion:
        token: close-39ad67555f36
        repository: github.com/xianxu/ariadne
        reviewed_head: d4a79fa2665c5d708e3807f554b27fa0202c32b7
        evidence_commit: 59c37fbf604d8e32ec5d03c87daf0f76757e240d
        landed_commit: 0a7fc836289b6dbca7d1bcb3faf56990a5df8404
---

# Pre-close and affected checks reuse an existing same-head full-suite pass instead of queueing a new full run

## Problem
