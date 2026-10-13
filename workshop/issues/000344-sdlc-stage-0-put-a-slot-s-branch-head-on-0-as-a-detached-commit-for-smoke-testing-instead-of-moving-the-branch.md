---
id: 000344
status: codecomplete
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-13
estimate_hours: 1.32
card_mirror: 'd29e4e7cc326591ede59d7cb5564c8a80d9caadd' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T14:01:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 1.00
---

# sdlc stage :0: put a slot's branch HEAD on :0 as a detached commit for smoke testing, instead of moving the branch

## Problem

To let the operator smoke-test a slot's work, the TL has the slot run `sdlc move :0`: the issue branch moves to the `:0` checkout (where the operator's Couch builds from), and the slot loses its branch. That costs ceremony every time: the slot must move the branch back before close (close measures hours from the slot's own transcripts; 2026-10-12 pair#491 had to move back), it can't keep working meanwhile, and each worker rediscovers the procedure (the operator noticed a worker 'discovering how to stage at :0').

## Spec

The operator (2026-10-12): `sdlc stage :0` (run from the slot holding the branch) puts the branch's current HEAD on `:0` as a **detached commit**, and optionally builds there (`--build`, the repo's `make build`). The slot keeps its branch and keeps working; it closes and merges from its own slot as usual.
- Refuses when `:0` has uncommitted changes; replaces any earlier staged commit (prints what it replaced).
- `sdlc stage --clear :0` (or any later normal use) returns `:0` to its resting branch; landing does not require it.
- Prints the staged commit and, with `--build`, the built binaries' vcs.revision, so the TL can verify in one step.
- Retire `sdlc move :0` from the smoke-test procedure in docs and the TL skill (keep `move` for genuinely moving a branch).

## Done when

- A slot stages its HEAD on `:0` with one command, keeps its branch, and later closes from its own slot; `:0` dirty refuses; a second stage replaces the first; tests cover each.
- Docs and the xx-tl skill's smoke procedure use `sdlc stage`.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec               design=0.3  impl=0.05
item: smaller-go-module        design=0.05 impl=0.12
item: smaller-go-module        design=0.05 impl=0.2
item: smaller-go-module        design=0.05 impl=0.12
item: smaller-go-module        design=0.05 impl=0.1
item: atlas-docs               design=0.05 impl=0.08
design-buffer: 0.1
total: 1.32
```

Items, in order:
- issue-spec: exploration, the durable plan, two plan-quality rounds.
- smaller-go-module ×4:
  1. the pure rules (`checkStage`/`checkUnstage`, `Leavable`, `checkMove`);
  2. `stage`/`--clear` observe-switch-verify-record, and move's record clear;
  3. `--build` and the binaries' revisions;
  4. the fleet `staged` hold.
- atlas-docs: stage/move help, workspace-branching, sdlc-binary.

The rate is the same as #308's (familiar `move` code, real-git fixtures that
already exist).

## Plan

Durable plan: `workshop/plans/000344-sdlc-stage-0-put-a-slot-s-branch-head-on-0-as-a-detached-commit-for-smoke-testing-instead-of-moving-the-branch-plan.md`.

- [x] Pure rules: `checkStage`, `checkUnstage`, `moveSide.Leavable`, and move accepting a staged target.
- [x] `sdlc stage [:N]` and `--clear`: observe, switch detached, verify, record; move clears the record.
- [x] `--build`: `make build` in the target, `bin/*` `vcs.revision` against the staged commit.
- [x] Docs: the stage/move help and the atlas. The xx-tl change lives in ops, so the hub gets it in the landing report.

## Log

### 2026-10-12

### 2026-10-13
- 2026-10-13: closed — make test green at 911b8afa: 12 shards, 1110 cmd/sdlc tests plus 46 packages, 0 failed. Round 1 fixes: TestStageAcceptsADetachARefKeeps covers stage, --clear and move on a ref-contained detach with no record (mutation-checked: all three fail without the for-each-ref rule). The minors: a repeat-stage test, fleet render staged=true, an unreadable build info is reported, and the plan boxes and xx-tl wording are accurate. Earlier evidence stands: the stage, refusal, move, fleet-hold and build tests, and the record-trust and TreeRow.Staged mutation checks.; review verdict: SHIP

- Dispatched by the TL (ops:0); claimed in ariadne:4. The diff will leave the
  quick-flow shell (~200 code lines), so the flow is pinned to full.
- Plan-quality round 1 raised PQ-1: a staged `:0` read as a detached
  needs-recovery in the fleet view. It now gets the hold `staged`, from an exact
  match against the record. PQ-2: the record can go stale, and is trusted only
  on an exact HEAD match. The effect order was stated, and the test strategy is
  one table.
- Mutation-checked:
  - without the record trust, re-staging after an amend refuses
    (`TestStageKeepsTheBranchAndReplaces`);
  - without `TreeRow.Staged`, the fleet hold test fails
    (`TestMoveReplacesAStagedSlot`).
- The xx-tl skill lives in ops (`~/workspace/ops/skills-draft/xx-tl`), which
  workers must not edit. Its manual detached procedure (fetch, then
  `checkout --detach FETCH_HEAD`, then `checkout main`) is to become
  `sdlc stage [--build]` / `sdlc stage --clear`; the hub is told at landing.
- Close review round 1 (FIX-THEN-SHIP). BR-1: the ref-containment half of
  `Leavable` had no test; `TestStageAcceptsADetachARefKeeps` now covers stage,
  --clear and move. All four Minors are fixed:
  - fleet render `staged=true`;
  - a repeat-stage test;
  - an unreadable build info is reported, and only "not Go" is skipped;
  - the plan boxes are ticked, and the xx-tl wording is accurate.
