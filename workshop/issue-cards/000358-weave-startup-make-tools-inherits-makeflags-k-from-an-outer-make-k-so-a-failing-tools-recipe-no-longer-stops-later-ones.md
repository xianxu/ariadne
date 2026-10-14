---
id: 000358
status: done
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
github_issue:
started: 2026-10-13T18:04:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.09
tracker:
    version: 1
    completion:
        token: close-7c8366df91b9
        repository: github.com/xianxu/ariadne
        reviewed_head: 2bc43b0cacc6bb967a0dd121341d8843ebb81d9d
        evidence_commit: 8737806f2da78fd455c369ba9c010982ac4bbe05
        landed_commit: 6417f54979fad5218eb59da032e208eda0dbfe34
---

# weave: startup make tools inherits MAKEFLAGS=k from an outer make -k, so a failing tools recipe no longer stops later ones

## Problem
