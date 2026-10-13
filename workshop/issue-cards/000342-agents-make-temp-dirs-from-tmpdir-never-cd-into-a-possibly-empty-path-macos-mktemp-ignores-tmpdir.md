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
actual_hours: 0.37
tracker:
    version: 1
    completion:
        token: close-dd2fd98b4b52
        repository: github.com/xianxu/ariadne
        reviewed_head: 7b167cedfd0e447270a97af75cf90db5f372ce71
        evidence_commit: 32c0d2d95192889181b80c3302dcbb116318e183
---

# agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR)

## Problem
