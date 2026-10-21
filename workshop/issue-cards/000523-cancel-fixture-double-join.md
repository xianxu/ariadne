---
id: 000523
status: done
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
github_issue:
started: 2026-10-21T06:19:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.34
tracker:
    version: 1
    completion:
        token: close-ae3bb0f3413a
        repository: github.com/xianxu/ariadne
        reviewed_head: 90f1be0734c552f5dd6d6fa806d3a516bfeeb8c2
        evidence_commit: 5fd48a62f279eb0f91696ff7555941d739aa883d
        landed_commit: 2be41ca0ec221ae5bb100e25c06ac3b858ce92ab
---

# Native cancellation fixture waits four minutes after successful joins

## Problem

TestLocalCheckCancelOnlyOwnRun spends about four minutes waiting in successful cleanup after both cancellation assertions already joined their native children. This inflates the machine-wide suite wall time without exercising more behavior.
