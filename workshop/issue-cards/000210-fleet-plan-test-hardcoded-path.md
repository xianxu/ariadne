---
id: 000210
status: open
created: 2026-09-02
updated: 2026-10-09
estimate_hours:
github_issue:
started: 2026-10-09T18:40:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# fleet plan test reads a hardcoded workshop/plans path that archiving breaks

## Problem

`TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`
(`cmd/sdlc/fleet_plan_test.go:14`) opens
`workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md` by literal path.
`dfeba9c` archived that plan to `workshop/history/`, which is the normal end of
a plan's life — so the test has been red ever since, for the intended behavior
of a different verb.

The suite is therefore red at every close boundary. That is corrosive in a way
the individual failure is not: a permanently-failing test trains every reader
(and every `--verified` claim) to explain a failure away instead of reading it,
and a real regression landing next to it is indistinguishable from the noise.
The #206 close review flagged it from outside for exactly that reason.

The bug is the lookup, not the archiving. `sdlc resolve` already answers "where
is issue N's plan, wherever it currently lives" — the archive-inclusive
resolution built for #144 — and this test predates using it.
