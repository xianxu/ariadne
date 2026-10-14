---
id: 000357
status: open
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
---

# test gate: the suite queue is polled, not FIFO, so a waiter can starve (51+ min) while later suites take the head

## Problem
