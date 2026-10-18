---
id: 000401
status: working
created: 2026-10-15
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T20:07:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead/tl-worker: hot-path plans state their expected cost per driver and the counted test that defends it

## Problem

Resource regressions (pair#579's 61× writes, pair#584's 15 s decode) passed plan review because nobody asked what an
operation should cost. Ops project resource-observability decided (10-15) that hot-path changes state their expected
cost and count it.
