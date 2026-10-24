---
id: 000594
status: done
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
actual_hours: 0.80
tracker:
    version: 1
    completion:
        token: close-439afe2ebf23
        repository: github.com/xianxu/ariadne
        reviewed_head: 0e1eeeb84505767cb8aab13fee3756943c5886ec
        evidence_commit: 42a67a547432734c090abf9f238f19e7b0e6b269
        landed_commit: e18d8b11ba19c418804716e9eacd318a87bf1750
---

# Actual attribution excludes witnessed local staging intervals

## Problem

After #591, pair:6 cannot close pair#830 with numeric actuals because an explicitly witnessed local staging branch is treated as unknown/detached/ambiguous during concurrent claims.
