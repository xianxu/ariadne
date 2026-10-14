---
id: 000369
status: done
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
github_issue:
started: 2026-10-14T11:06:25-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.94
tracker:
    version: 1
    completion:
        token: close-02f686c7c0b4
        repository: github.com/xianxu/ariadne
        reviewed_head: b97208e5484273f0cc54d44be9b8257d5b0cb584
        evidence_commit: f9653a63018fe180eada73dfd722496b028897a8
        landed_commit: 90416a249f606cb2d1eca1e6d8cdd7128c9c6937
---

# TestWaitCIGateBoundsRealMakeProbe/cancel is flaky on main: cancel sometimes returns after ~1.05s (the make child's sleep) instead of under 700ms; also TestSetupLeaseSurvivesRunnerParentDeath failed once under the full suite

## Problem
