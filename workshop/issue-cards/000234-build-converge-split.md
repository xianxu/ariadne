---
id: 000234
status: working
created: 2026-09-17
updated: 2026-10-10
estimate_hours:
github_issue:
started: 2026-10-10T17:41:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# Split actuals at the first boundary review: build time vs converge time

## Problem

Calibration compares an estimate against an actual that it cannot, by
construction, predict all of. The actual runs from claim to close, so it includes
every boundary-review round and the fixes each one forces. An estimate prices the
work up to "I believe it's done", and nobody budgets time to fix bugs they have
not written yet. Operator, 2026-09-17:

> don't worry about estimate. all we need is to add a calibration factor, e.g. at
> your current skill level, the amount of bugs you will write. no one budget time
> to fix bugs they write, so estimate necessarily not accounting for those closing
> rounds. … maybe we should be tracking time to boundary review and boundary
> review itself. and really the estimation quality is time to boundary review.

#231 is the example: 6.33h estimated. Its measured actual passed 9.6h during M2,
by which point M1 had taken three boundary rounds and M2 four. The overrun
is mostly convergence, not forecasting error, but today's est/actual ratio cannot
tell the two apart.
