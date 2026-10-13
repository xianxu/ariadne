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
actual_hours: 0.32
tracker:
    version: 1
    completion:
        token: close-9a3984758384
        repository: github.com/xianxu/ariadne
        reviewed_head: ad470891aa1ce396f5c1289a8877066cd9b01759
        evidence_commit: 91ac4572a8d96481f7903d78c0e4ee4d8d6908c2
---

# agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR)

## Problem
