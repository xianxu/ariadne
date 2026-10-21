---
id: 000517
status: codecomplete
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T23:25:39-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.43
tracker:
    version: 1
    completion:
        token: close-d3f4c3926cd9
        repository: github.com/xianxu/ariadne
        reviewed_head: 0cc02ff1769bf20089b7c3996ba965269c566744
        evidence_commit: b2c1d93657de0baea92ba9e6403a7e58c8619fd4
---

# Prepare generated inputs before affected graph discovery

## Problem

Pair's fresh affected graph checkout has no generated runtime bundle. Native go list -test ./... fails at cmd/internal/runtimebundle/embed.go:8:41 with pattern assets/runtime/files: no matching files found. Running the existing runtimebundle generator in that same checkout makes the graph load (408 entries measured on 2026-10-20). Every affected run currently falls back to a full suite, including docs-only changes. Pair's command runs after graph discovery, too late to establish the prerequisite.
