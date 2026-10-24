---
id: 000594
status: working
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
github_issue:
started: 2026-10-24T08:46:27-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# Actual attribution excludes witnessed local staging intervals

## Problem

After #591, pair:6 cannot close pair#830 with numeric actuals because an explicitly witnessed local staging branch is treated as unknown/detached/ambiguous during concurrent claims.
