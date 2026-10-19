---
id: 000394
status: codecomplete
created: 2026-10-15
updated: 2026-10-18
estimate_hours: 1.6
github_issue:
started: 2026-10-18T15:44:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 3.86
tracker:
    version: 1
    completion:
        token: close-91d33d6baa9b
        repository: github.com/xianxu/ariadne
        reviewed_head: bda6cc0efc5613bf2994d9cdca538942bb2d42fd
        evidence_commit: 0c48667b757909b27614e8dd66c8173a13e20b8d
---

# Issue-backed flake landing and daily triage

## Problem

An intermittent suite failure currently either blocks a landing or invites an unexplained retry. Both hide the actual test-health cost. #390's approved rule requires one passing exact-head suite while letting an already identified, separately owned flake be diagnosed and rerun on the same head.
