---
id: 000382
status: open
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
github_issue:
started: 2026-10-14T21:42:00-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# test-shard.test.py creates its temp dir under /tmp, which agent sandboxes deny; use TMPDIR

## Problem
