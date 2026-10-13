---
id: 000235
status: codecomplete
created: 2026-09-18
updated: 2026-10-12
estimate_hours: 1.4
github_issue:
started: 2026-10-12T17:21:51-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.19
tracker:
    version: 1
    completion:
        token: close-c16edbf3ed2a
        repository: github.com/xianxu/ariadne
        reviewed_head: 1ffc38208e773a21271f8fdf4fd043c4eac152d4
        evidence_commit: 7d11abadac22ab9bf1c8e18593ca030351b84889
---

# Durable-plan lookup is exact-name: a plan under another slug is invisible

## Problem

sdlc finds an issue's durable plan by ONE exact name: the issue file's stem plus
`-plan.md` (`readOptionalPlanFile`, `cmd/sdlc/changecode.go`). A plan saved under
any other slug for the same issue id is invisible to it, silently.

Found in pair#283 (Log, 2026-09-18). The plan was saved as
`000283-default-cursor-style-plan.md`, while the issue file is
`000283-default-cursor-style-overrides-terminal.md`. Two consequences:

- **#231's flow inference said quick** although a durable plan existed. Renaming
  the file fixed it.
- **Plan-quality never saw the plan.** This predates #231: a misnamed plan is
  skipped by the plan gate too, and so the plan the agent wrote goes unreviewed.

The close-time review manifest resolves the plan the same way
(`reviewPlanPaths` / `optionalReviewPlanPath`, `cmd/sdlc/reviewwindow.go`), so
the boundary reviewer misses a misnamed plan as well.
