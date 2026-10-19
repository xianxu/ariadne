---
id: 000469
status: codecomplete
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
actual_hours: 0.29
tracker:
    version: 1
    completion:
        token: close-27422eb522c4
        repository: github.com/xianxu/ariadne
        reviewed_head: 7b2bff5878d86def1d2aa555929b5136b90e1ceb
        evidence_commit: 85f371104f76a2d23ab3c09f0ae2eeafa7a24919
---

# landingline: test reads the archived #449 plan; move the as-built list to the atlas (main is red)

## Problem

ariadne main is red since #449 landed: `cmd/sdlc/internal/landingline/trainstate_test.go` `TestPlanNamesEveryTrainEventAndEffect`
reads `../../../../workshop/plans/000449-landing-train-plan.md`, which the close archived to `workshop/history/`. Every
recorded suite now fails there (seen on #458's integrated suite, e07ab835), so every ariadne landing is blocked.
