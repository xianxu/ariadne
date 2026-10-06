---
id: 000296
status: open
deps: []
github_issue:
created: 2026-10-06
updated: 2026-10-06
estimate_hours:
card_mirror: 'f50aca0d6ff8fc2d9995a30f84c15de7292bdc4c' # card fields mirrored from issue-cards; edit via sdlc
---

# weave: recover a sourceless substrate's source from the primary checkout's sibling, and migrate the row

## Problem

A `construct/deps` substrate row without a clone source (`substrate ../ariadne`)
works in a primary checkout only because the sibling already exists next to it.
In a Couch slot environment (`~/workspace/worktree/<repo>-slotN/`) the sibling is
absent, so `weave compile` stops with `missing substrate … record its source in
construct/deps` (`cmd/weave/internal/acquire/acquire.go`, the `row.Source == ""`
branch). Fifteen repositories under `~/workspace` have such rows (#293's audit).
Observed: `tools:1`, `brain:1`, `parli:1`. pair#387's slot reconciler can only
report it and hand it to the repository's `:0` agent.

#293 planned to fix this with a fleet-wide manual edit of every `construct/deps`,
and recorded the operator's rejection of a weave fallback as fragile. **The
operator reversed that decision on 2026-10-06:**

> even when I'm a single operator working on those repos, there are many branches
> and different stages of work going on, to properly migrate format is going to be
> pretty messy, and it's much better to design migration automation inside the
> tool itself, just like how I would typically approach migration.

A hand migration has to land on every branch of every repository. A migration
inside the tool runs wherever weave runs and needs no coordination.

## Spec

Captured for operator review; no implementation is authorized by this issue
creation.

- **Recover the source.** When a substrate destination is missing and its row has
  no source, weave resolves the declaring repository's primary checkout (the main
  worktree of its git common directory). It then takes the same relative path
  from there (`<primary>/../ariadne`). If that is a git checkout whose toplevel is
  itself and that has an `origin` URL, weave clones from that URL, as it would for
  a declared source.
- **Recover only when the evidence agrees.** The sibling must be the layer the row
  names. The minimal check is its repository name against the row's basename;
  whether the layer identity from `layergraph` is the stronger and still-cheap
  check is open. Weave refuses the recovery if the sibling is missing, not a
  checkout, has no `origin`, or does not match. It then keeps today's error,
  extended with what it tried ("no source declared; <path> is <reason>").
- **Say it loudly.** Every recovery prints one notice naming the row, the inferred
  URL and the sibling it came from. A recovered source is never silent.
- **Migrate the row (open question for the operator).** Write the recovered
  source into `construct/deps`, so the declaration converges to the explicit
  form. Where:
  - (a) Only when weave runs in the primary checkout, where an uncommitted edit
    is ordinary work.
  - (b) Also in a slot, where it dirties the host checkout on its resting branch.
    pair#387 treats a dirty host as work to keep, which argues against.
  - (c) Never write: the notice names the line to add, and the recovery stays
    the runtime path.
  The recommendation is (a), plus (c)'s notice in a slot.
- **Relation to #293.** This replaces #293's manual fleet edit. #293's other part,
  weave flagging a sourceless row so a new derivative isn't born without one,
  still holds as the warning half of this migration. Revise or close #293 to
  match when this is planned.
- **Downstream.** pair#387 needs no change: setup succeeds, so the reconciler's
  hand-off for this cause disappears on its own. Its remembered-failure memo keys
  on the inputs, so an upgraded weave must re-run setup. Checked 2026-10-06: the
  digest (`setupInputsDigest`) covers the host HEAD and every `construct/deps`, but
  not weave itself. Until pair adds weave's identity to it, an upgraded weave
  takes effect on `couch --reconcile` (which ignores the memo) or on any deps
  edit, not on a plain open.

## Done when

- A slot compile of a repository with a sourceless substrate row clones the
  substrate from the primary checkout's sibling's `origin`. The recovery notice
  names the row, URL and sibling. Covered by a test with real git and a
  stateful sibling.
- Each refusal (sibling missing, not a checkout, no `origin`, mismatched
  identity) keeps the missing-substrate error, extended with what was tried, and
  has a test.
- The chosen migration behaviour (a/b/c) is implemented and tested: the row ends
  up with an explicit source where the operator chose that it should.
- Live: `couch --reconcile tools:1` completes setup with no hand edit of tools'
  `construct/deps`.

## Plan

- [ ]

## Log

### 2026-10-06

Filed from pair#387's smoke test of `tools:1`, at the operator's request. The
fallback was first proposed during pair#387's planning and rejected (recorded in
#293); the operator reversed that here, preferring migration automation inside the
tool to a hand migration across every branch. The details are left local on `main`
(not moved) for the operator to refine.
