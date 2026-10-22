---
id: 000544
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'ad22d4bbac415e2cb9525355d53f88a9234657f7' # card fields mirrored from issue-cards; edit via sdlc
---

# issue migrate drops a legacy close's completion record; later issues' actual measurement then refuses (echo#1 → echo#3)

## Problem

## Spec

From ops-echo:0 (10-21 evening). echo#1 closed under the legacy workflow (merge 08124b3, archive 112fd43,
actual_hours 0.80) just before echo migrated to the tracker (`sdlc issue migrate --apply`, 6f0fe49). After the
migration, `sdlc issue show 1` reads `status done, completion: no close recorded`. echo#3's close now stops before
review: `sdlc actual --issue 3` refuses "cannot auto-compute (issue #1 adopted actual has no trustworthy saved
window: completion is missing)", from the adopted-actual exclusion (`cmd/sdlc/actualexclusion.go`, `adoptionRefuse`
with no saved window).

So the migration carries a legacy-closed issue's status and `actual_hours`, but not its close binding or saved
measurement window, and every later issue whose window overlaps it is blocked from measuring. Every legacy repo that
migrates with recently closed issues hits this.

Fix: `sdlc issue migrate` carries each legacy close's completion record (the close commit and saved measurement
window) onto the tracker. For already-migrated repos, `--reconcile` (or a narrow repair verb) backfills it from the
legacy close commit and archive, with the same evidence the legacy close used. Never invent a window: if none is
recoverable, mark the adopted actual as windowless-legacy, and let the exclusion treat a windowless issue whose
close predates the candidate's engagement as disjoint (skip), not refuse.

## Done when

- Migrating a legacy fixture with a closed issue keeps its completion record; `sdlc issue show` shows the close (test).
- In an already-migrated fixture, the repair restores it, and a later issue's `sdlc actual` measures (test).
- A legacy-closed issue with no recoverable window, closed before the candidate's claim, doesn't block measurement
  (test); an overlapping one still refuses.

## Plan

- [ ]

## Log

### 2026-10-21
