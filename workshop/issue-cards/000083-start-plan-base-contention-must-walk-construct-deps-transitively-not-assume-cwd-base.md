---
id: '000083'
status: done
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 2.0
actual_hours: 2.0
---

# start-plan base-contention must walk construct/deps transitively, not assume cwd==base

## Problem

#82 M3 added a `start-plan` base-contention heads-up, gated on a predicate
`isBaseRepo(root)` = "`construct/` is a real directory, not a symlink". **That
premise is false**, so the feature is wrong in two ways:

1. **Wrong signal.** Every repo — base *and* derivative — has a *real*
   `construct/` directory carrying its own `base.manifest` / `deps` / `scripts`,
   with only specific *subpaths* symlinked (`construct/adapted -> ../../ariadne/
   construct/adapted`). Verified: `nous` and `pair` both have a real `construct/`.
   So `isBaseRepo` returns `true` in **every** derivative, and `start-plan` fires
   `base (nous): …` / `base (pair): …` in derivatives — mislabeling the derivative
   as the base and reporting *its own* git state as "base contention". The exact
   opposite of the intended "silent unless you're looking at the base". The
   shipped `TestIsBaseRepo` passed only because it encoded the same wrong model
   (real-dir vs a synthetic symlinked `construct/`).

2. **Wrong shape.** It's not "one base". The substrate graph is a **chain**
   declared in `construct/deps` (the sole substrate-graph carrier since #60):
   - `ariadne/construct/deps` — absent → ariadne is the **root** (no upstream).
   - `nous/construct/deps` → `substrate ../ariadne`.
   - `brain/construct/deps` → `substrate ../ariadne` (+ a `data` row).
   Because a repo reads **all** its transitive upstreams' working trees live (via
   the symlinks), the "moving ground" you build on is the whole **dependency
   path**, not a single base. Working in `brain`, contention in `ariadne` (and
   any intermediate substrate it depends on) is what matters — and `start-plan`
   in a derivative should surface *that*, which was the Spec's primary motivation
   ("base changes are discovered mid/late in a derivative session"). The cwd-only
   gather never reads upstream trees at all.

(Vendoring is gone — symlink-only model now — so there is no vendor-mode branch to
handle; `construct/deps` is the single source of the dependency edge.)
