---
id: 000392
status: open
created: 2026-10-15
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T21:39:59-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Weekly test-economy snapshot from local checks and landings

## Problem

#390's test-economy design needs a trustworthy weekly trend before changing evidence tiers or declaring package budgets. `sdlc test-stats` reports recent local attempts, but no durable weekly snapshot joins them to landed issues; older records expire after 30 days. The headline metric, full suites per landed issue, has no stored baseline.
