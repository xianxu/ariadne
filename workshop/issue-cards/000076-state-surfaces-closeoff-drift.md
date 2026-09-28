---
id: '000076'
status: done
created: 2026-06-03
updated: 2026-06-04
estimate_hours: 2
actual_hours: 0.24
---

# sdlc state surfaces done-but-open close-off drift

## Problem

`sdlc` gates the **entry** to each workflow stage (claim, change-code, close,
merge all refuse without evidence) but never surfaces a **stale exit**: an issue
whose work has shipped yet was never formally closed. The work is done; the
bookkeeping isn't — exactly the close-off hygiene drift `sdlc` exists to kill.

Concrete specimen (#51, found 2026-06-03): the in-place-branch workflow shipped
2026-05-31 (`ed4b550`) and became the live default everyone used for a week, but
the issue sat `status: open`. The only thing holding it open was a *deferred-
validation* checkbox ("live dogfood — best done as the first post-bootstrap
branch-based task"). That validation happened — repeatedly, automatically (every
`change-code → pr → merge` since) — but **nothing triggered closing the loop**.
A one-off ad-hoc scan on 2026-06-03 found this isn't unique: several open/working
issues have merged work + near-complete plans.

`sdlc state` already exists for "structural drift detection" (warn-only) and
already has `detectDrift` + `DriftFinding` (e.g. it flags "working with N plan
items, none ticked"). This is the natural home for the close-off check — the
inverse of "started but no progress": **finished but never closed**.
