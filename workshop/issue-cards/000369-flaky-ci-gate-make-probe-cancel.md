---
id: 000369
status: open
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
github_issue:
started: 2026-10-14T11:06:25-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# TestWaitCIGateBoundsRealMakeProbe/cancel is flaky on main: cancel sometimes returns after ~1.05s (the make child's sleep) instead of under 700ms; also TestSetupLeaseSurvivesRunnerParentDeath failed once under the full suite

## Problem
