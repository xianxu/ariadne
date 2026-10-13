---
id: 000342
status: codecomplete
created: 2026-10-12
updated: 2026-10-13
estimate_hours:
github_issue:
started: 2026-10-13T15:47:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.28
tracker:
    version: 1
    completion:
        token: close-a1e3e3b8860f
        repository: github.com/xianxu/ariadne
        reviewed_head: 7fa5dce54b23874e8b2588dad7a9ff2cc2c30227
        evidence_commit: 652f9cc15de93053c71eab9870e2f035e2beb177
---

# agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR)

## Problem
