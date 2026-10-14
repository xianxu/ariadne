---
id: 000357
status: done
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
actual_hours: 0.20
tracker:
    version: 1
    completion:
        token: close-ce7dd6f93dfd
        repository: github.com/xianxu/ariadne
        reviewed_head: 603f60e74f8b2e3036374febe5eb650dd47ef1ed
        evidence_commit: 43826c76588d46593c06f29aae5a9374b72d4f3e
        landed_commit: 2f8041ba3656d6de5f9eaf128fde30685116e64a
---

# test gate: the suite queue is polled, not FIFO, so a waiter can starve (51+ min) while later suites take the head

## Problem
