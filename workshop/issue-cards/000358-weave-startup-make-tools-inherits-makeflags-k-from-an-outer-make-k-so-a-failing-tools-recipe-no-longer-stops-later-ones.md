---
id: 000358
status: open
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
tracker:
    version: 1
---

# weave: startup make tools inherits MAKEFLAGS=k from an outer make -k, so a failing tools recipe no longer stops later ones

## Problem
