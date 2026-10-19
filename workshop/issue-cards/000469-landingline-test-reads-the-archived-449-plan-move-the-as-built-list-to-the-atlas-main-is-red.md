---
id: 000469
status: working
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T11:26:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# landingline: test reads the archived #449 plan; move the as-built list to the atlas (main is red)

## Problem

ariadne main is red since #449 landed: `cmd/sdlc/internal/landingline/trainstate_test.go` `TestPlanNamesEveryTrainEventAndEffect`
reads `../../../../workshop/plans/000449-landing-train-plan.md`, which the close archived to `workshop/history/`. Every
recorded suite now fails there (seen on #458's integrated suite, e07ab835), so every ariadne landing is blocked.
