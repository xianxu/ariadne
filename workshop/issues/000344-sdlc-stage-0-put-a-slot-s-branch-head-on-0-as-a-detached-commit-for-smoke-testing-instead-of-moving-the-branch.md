---
id: 000344
status: working
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-13
estimate_hours: 1.32
card_mirror: '51c8337684f3b8095279c4e27ae7a169bf672beb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T14:01:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
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


Durable plan: `workshop/plans/000344-sdlc-stage-0-put-a-slot-s-branch-head-on-0-as-a-detached-commit-for-smoke-testing-instead-of-moving-the-branch-plan.md`.

- [ ] Pure rules: `checkStage`, `checkUnstage`, `moveSide.Leavable`, and move accepting a staged target.
- [ ] `sdlc stage [:N]` and `--clear`: observe, switch detached, verify, record; move clears the record.
- [ ] `--build`: `make build` in the target, `bin/*` `vcs.revision` against the staged commit.
- [ ] Docs: stage/move help, the atlas, and the xx-tl change sent to the hub (it lives in ops).

## Log

### 2026-10-12

### 2026-10-13

- Dispatched by the TL (ops:0); claimed in ariadne:4. The diff will leave the
  quick-flow shell (~200 code lines), so the flow is pinned to full.
