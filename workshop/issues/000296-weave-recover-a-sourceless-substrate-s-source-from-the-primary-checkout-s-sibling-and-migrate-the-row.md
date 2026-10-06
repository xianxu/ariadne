---
id: 000296
status: working
deps: []
github_issue:
created: 2026-10-06
updated: 2026-10-06
estimate_hours:
card_mirror: 'a7904e910ebd87d62bfe25109259694c984f919d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-06T09:06:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "22b771d5", done: "3afd7eec"}
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

Design (operator chose migration (a) on 2026-10-06, see Revisions):

- **Where recovery applies.** Only inside a numbered environment, where
  `acquire.Policy` exists. The environment root mirrors the fleet directory:
  `<fleet>/worktree/<repo>-slotN/<name>` corresponds to `<fleet>/<name>`, and
  `ValidSlotDependency` already confines every substrate to a direct child of the
  environment root. So the primary-side sibling of a missing destination is
  `dir(Policy.PrimaryRoot)/base(dest)`; it holds for the host's rows and for rows
  of substrates cloned into the environment alike. `PrimaryRoot` is the
  Git-proved `env.Host.PrimaryRoot`, added to `Policy`. Outside an environment a
  missing sourceless row keeps today's error unchanged.
- **Evidence.** The sibling must exist, be its own git toplevel, have an
  `origin`, that origin must not be a local/file source (the policy refuses
  those), and the origin's repository name must equal `base(dest)`. Layer
  identity is then proved by the existing post-clone checks (`origin/main`,
  `construct/base.manifest`), so no separate `layergraph` identity probe is
  needed (Simplicity First). Each failure extends the missing-substrate error
  with `no source declared; <sibling> <reason>`.
- **Effect.** `Restore` treats the recovered URL exactly as a declared source
  (identity-conflict map, `Ensure`, dry-run `Missing`), and records a
  `Recovery{Owner, Path, URL, Sibling}` in `Result.Recovered`. The command layer
  prints one notice per recovery naming the row, URL, sibling and the line to
  add to `construct/deps` (option (c)'s notice, in a slot).
- **Migration (a).** After a successful non-dry-run restore, when weave runs in
  the primary checkout (no environment policy; `--absolute-git-dir` equals the
  common dir), each sourceless substrate row of the root's own `construct/deps`
  whose destination is a git toplevel with a non-file `origin` is upgraded in
  place with that origin, through the same row writer `weave link` uses
  (`declareSubstrate`, extracted from `recordLink`, ARCH-DRY). One notice per
  rewritten row. A row that can't be upgraded (no origin, local origin) gets a
  warning that slots can't restore it: the warning half of #293.
- ARCH-FUNERAL: creates nothing durable beyond the one-line edit to
  `construct/deps`, which the operator commits like any other edit.

Steps:

- [x] `Policy.PrimaryRoot` from environment discovery; `recoverSource` +
      `Result.Recovered` in `acquire.Restore`; refusal reasons in the error.
- [x] Tests (real git, stateful sibling): slot restore clones from the
      sibling's origin and reports the recovery; dry run reports it as missing
      without cloning; each refusal (missing, not a checkout, no origin, local
      origin, name mismatch) keeps the error with its reason.
- [x] Extract `declareSubstrate` from `recordLink`; `migrateSourceless` in the
      primary checkout; notices for recoveries and migrations shared by
      `compile` and `dependencies`.
- [x] Tests: primary compile rewrites the row (comment preserved) and is
      idempotent; slot / linked worktree / dry run never write; an unrecoverable
      row warns.
- [x] Atlas: weave acquisition note; revise #293 to drop the manual fleet edit.
- [x] Live: `couch --reconcile tools:1` completes setup with no hand edit.

## Log

### 2026-10-06

Filed from pair#387's smoke test of `tools:1`, at the operator's request. The
fallback was first proposed during pair#387's planning and rejected (recorded in
#293); the operator reversed that here, preferring migration automation inside the
tool to a hand migration across every branch. The details are left local on `main`
(not moved) for the operator to refine.

Moved details to main from :0 and claimed in :1. Operator chose migration (a)
(write the row in the primary checkout only; notice in a slot).

Implemented (21209aed). Recovery lives in `acquire.Restore` (`recoverSource`,
`Result.Recovered`), scoped by `Policy.PrimaryRoot`; migration + notices in
`cmd/weave/migrate.go` (`finishRestore`, shared by `compile` and
`dependencies`); `declareSubstrate` extracted from `recordLink` (ARCH-DRY).
Tests: `recover_test.go` (real git; clone, dry run, warm reuse, five refusals,
outside-environment), `migrate_test.go` (primary rewrite keeps comment +
idempotent, linked worktree / slot / dry run / failed restore never write with a
positive control, end-to-end `weave dependencies` in a numbered environment via
`url.insteadOf`). Mutation-checked: dropping the file/name/no-origin checks, the
primary-checkout check, the policy guard, or the `PrimaryRoot` wiring each fails
a test.

Live, `tools:1` (this branch's weave build, since the installed one predates it):
`weave dependencies --dry-run` printed `recovered git@github.com:xianxu/ariadne.git
from /Users/xianxu/workspace/ariadne`; the real run cloned
`tools-slot1/ariadne` on `main` with that origin, and `construct/deps` stayed
`substrate ../ariadne` (slot never writes). Then `couch --reconcile tools:1`:
`dep:ariadne present`, `setup present`, "tools:1 prepared". Compile regenerated
two base-layer files in tools:1 (`.gitignore` gains xx-couch, merge-check.yml),
ordinary drift from a newer ariadne, left in place. pair's setup memo doesn't
key on weave's identity (Spec), so other slots pick this up on
`couch --reconcile` or a deps edit once the weave rollout lands.

## Revisions

- 2026-10-06 — Spec open question resolved by the operator: migration (a), plus
  (c)'s notice in a slot. Layer-identity question resolved in the Plan: repository
  name match before the clone, existing `origin/main` + manifest checks after it.
  Recovery scoped to numbered environments (the only place a sibling can be
  missing while its primary-side counterpart exists).
- 2026-10-06 — Close review round 2 (FIX-THEN-SHIP): the error outside a
  numbered environment is no longer "unchanged" as the Plan said; it gains
  "(no source declared; no numbered environment to recover it from)", kept
  deliberately because it says why recovery didn't apply. The remote-origin
  probe is shared (`acquire.Client.RemoteOrigin`, `PrimaryCheckout`) and the
  migration warning now carries its reason; README documents the behaviour.

