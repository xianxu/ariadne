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
actual_hours: 0.26
tracker:
    version: 1
    completion:
        token: close-a842e69e61cc
        repository: github.com/xianxu/ariadne
        reviewed_head: e3ff505c5fb11818cf82b9d98fe20c055668b957
        evidence_commit: 4e2d53b96a3cad927ba868ecd70763f9650d23da
---

# agents: make temp dirs from $TMPDIR; never cd into a possibly-empty path (macOS mktemp ignores TMPDIR)

## Problem
