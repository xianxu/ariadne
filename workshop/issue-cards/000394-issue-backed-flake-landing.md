---
id: 000394
status: done
created: 2026-10-15
updated: 2026-10-19
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
        token: close-610ea2d494ca
        repository: github.com/xianxu/ariadne
        reviewed_head: cf0284d281f887fbcea87863421bc1e8ace37c3c
        evidence_commit: e32aeb4eade726b4d89ccf9dae5688bf545a56e5
        landed_commit: 8461dcf13d6c71ca7022ef75e6ca61956dc11466
---

# Issue-backed flake landing and daily triage

## Problem

An intermittent suite failure currently either blocks a landing or invites an unexplained retry. Both hide the actual test-health cost. #390's approved rule requires one passing exact-head suite while letting an already identified, separately owned flake be diagnosed and rerun on the same head.
