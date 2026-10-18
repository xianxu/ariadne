---
id: 000432
status: working
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T18:01:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-worker: an early draft PR for CI overlap needs no hub approval

## Problem

Slots ask the hub for permission before opening an early draft PR so GitHub CI overlaps a queued local suite (pair:3 asked twice on 2026-10-17). The hub always says yes, because the suite, close review and merge gates still run before landing. Each ask costs a hub round trip while the suite queue is long.
