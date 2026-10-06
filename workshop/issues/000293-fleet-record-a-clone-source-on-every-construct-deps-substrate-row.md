---
id: 000293
status: open
deps: []
github_issue:
created: 2026-10-05
updated: 2026-10-05
estimate_hours:
card_mirror: '296dbf7971c5cc6bb83f086522b2b707aa7afa5f' # card fields mirrored from issue-cards; edit via sdlc
---

# Fleet: record a clone source on every construct/deps substrate row

## Problem

Most derivatives declare their substrate without a clone source, e.g.
`substrate ../ariadne`. That works in a primary checkout only because the
sibling happens to exist next to it (`~/workspace/ariadne`). In a Couch slot
environment (`~/workspace/worktree/<repo>-slotN/`) the sibling is absent and
weave has nowhere to clone it from, so setup fails with `missing substrate …
record its source in construct/deps`. Every new slot of such a repository fails
identically, because `construct/deps` reaches each slot through `main`.

The source column was introduced by #239 (typed rows, `weave link` records a
source). #243 (slots v2 dependency bindings) found pair and parley.nvim lacked
URLs and added them there only (pair#310, parley.nvim#274). The fleet rollout of
#241 covered the generated-file layout, not substrate sources. So the source
migration never reached the rest of the fleet.

Audit of `~/workspace/*/construct/deps` (2026-10-05):
- source declared: pair, parley.nvim (https); ducks, test-repo (ssh);
- no source: 42shots, astro, brain, brain-family, brain-private, kaggle, kbench,
  metis, metis.bak, nous, parli, robotics, tools, xianxu.dev, you-decide.

Observed on pair#367's smoke test: `tools:1`, `brain:1` and `parli:1` each had a
missing dependency for this reason. Found while planning pair#387 (the slot
reconciler), which reports the failure and hands it to the repository's `:0`
agent rather than working around it. The operator rejected a weave fallback
(infer the source from a sibling checkout's `origin`) as fragile: the source is
declared, not guessed.

## Spec

Captured for operator review; no implementation is authorized by this issue
creation.

- Record the source on every substrate row in every derivative, through each
  repository's own workflow (the brain repos on their capture rhythm). Pick one
  URL form for the fleet (https as pair and parley.nvim use, or ssh as ducks
  and test-repo use) and normalize the four existing ones to it.
- Keep it from regressing: a sourceless substrate row is reported by `weave
  verify-complete` (and/or `weave compile`) as an actionable warning or error
  naming the row and the fix, so the next derivative cannot be born without one.
- Sweep every repository under `~/workspace` rather than working from this list
  (#241's lesson: a list-driven fleet migration missed four repos).

## Done when

- Every `construct/deps` substrate row under `~/workspace` declares a source,
  verified by a sweep script whose output is recorded in the Log.
- weave flags a sourceless substrate row, with a test.
- A fresh Couch slot of a previously sourceless repository (e.g. `tools`)
  completes setup.

## Plan

- [ ]

## Log

### 2026-10-05

Filed from pair#387 planning at the operator's request ("I think this is the
migration that didn't finish"). Related open migration: #223 (untrack
weave-lowered substrate symlinks in derivatives).

## Revisions

- 2026-10-06 — Superseded in part by #296 (operator reversed the rejection of a
  weave fallback: migration belongs in the tool, not in a hand edit across every
  branch). Weave now recovers a sourceless row's source in a slot from the
  primary-side sibling's origin, records the source into the row when it runs
  in a primary checkout, and warns about rows it cannot (the "weave flags a
  sourceless row" half of this issue). The fleet-wide manual edit is dropped;
  what remains here is only normalizing the four existing URL forms, if still
  wanted. Candidate for `wontfix` or a narrowed retitle.
