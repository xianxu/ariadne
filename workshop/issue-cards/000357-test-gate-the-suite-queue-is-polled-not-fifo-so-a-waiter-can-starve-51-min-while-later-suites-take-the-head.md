---
id: 000357
status: codecomplete
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
github_issue:
started: 2026-10-13T19:46:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.18
tracker:
    version: 1
    completion:
        token: close-2994e48bafdf
        repository: github.com/xianxu/ariadne
        reviewed_head: 83634985882f1566e69c81d746218cacb9ce1d50
        evidence_commit: 677c2df09571c9b2072e73b56967e13e44274701
---

# test gate: the suite queue is polled, not FIFO, so a waiter can starve (51+ min) while later suites take the head

## Problem
