---
id: 000381
status: open
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
github_issue:
severity: low
started: 2026-10-14T21:16:15-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:0
    worktree: /Users/xianxu/workspace/ariadne
    repository: github.com/xianxu/ariadne
tracker:
    version: 1
---

# test-shard.test.py writes its temp dir under /tmp, so make test fails in a sandboxed slot

## Problem
