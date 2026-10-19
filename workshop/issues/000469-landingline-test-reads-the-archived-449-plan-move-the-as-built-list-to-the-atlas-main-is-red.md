---
id: 000469
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c129e7b95feb377a387dee9bc678a866a5eeaa58' # card fields mirrored from issue-cards; edit via sdlc
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
