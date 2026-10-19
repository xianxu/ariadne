---
id: 000469
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '845d3ed0ed49c9bd54292fe355a63d1d973475e5' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T11:26:36-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "80696cc8", done: "4e5a89d6"}
---

# landingline: test reads the archived #449 plan; move the as-built list to the atlas (main is red)

## Problem

ariadne main is red since #449 landed: `cmd/sdlc/internal/landingline/trainstate_test.go` `TestPlanNamesEveryTrainEventAndEffect`
reads `../../../../workshop/plans/000449-landing-train-plan.md`, which the close archived to `workshop/history/`. Every
recorded suite now fails there (seen on #458's integrated suite, e07ab835), so every ariadne landing is blocked.

## Spec

Root cause: a test pinned to a process artifact whose lifecycle moves it (plans are archived on close). The "as built" list
of train events and effects is durable documentation, so it belongs in the atlas: move the list to the landing-train atlas
page (create it if absent, linked from atlas/index.md) and point the test there. Also add a guard so this class can't recur:
a lint or test that fails when any `_test.go` reads a path under `workshop/plans/` or `workshop/issues/`.

## Done when

- `go test ./cmd/sdlc/internal/landingline/` passes on main with the plan archived; the full suite is green.
- The guard refuses a test reading `workshop/plans/…` (a fixture proves it).

## Plan

- [ ]

## Log

### 2026-10-19
- 2026-10-19: closed — TestPlanNamesEveryTrainEventAndEffect reads atlas/workflow/sdlc-binary.md and passes; golint rule test-reads-workshop with table tests; BR-1 fixed: atlas names the rule and TestAtlasNamesEveryRule keeps the catalog list in step; whole-repo catalog scan clean; full suite exit 0 in 12m52s on 7879351a (the fix after it adds a test and docs); review verdict: FIX-THEN-SHIP

2026-10-19 (ariadne:2): fixed.
- `TestPlanNamesEveryTrainEventAndEffect` now reads `atlas/workflow/sdlc-binary.md`. The as-built events, effects and phases were moved there from the archived #449 plan.
- New `pkg/golint` rule `test-reads-workshop`: a call argument that climbs from a package straight into `workshop/plans` or `workshop/issues` is flagged in tests. Those files move to `workshop/history` when their issue lands. Temp-repo fixture paths are not matched.
- The whole-repo catalog scan is clean.
- Close review FIX-THEN-SHIP, every Minor fixed as a class:
  - Atlas lines naming archived plans at `workshop/plans/` (#449's and three older ones) now point at `workshop/history/plans/`. `TestAtlasCitesNoLivePlanPath` refuses a numbered live plan path in the atlas.
  - The as-built claim is narrowed to events and effects.
  - `test-reads-workshop` also catches a path split into Join segments (`"..", "workshop", "plans"`).
