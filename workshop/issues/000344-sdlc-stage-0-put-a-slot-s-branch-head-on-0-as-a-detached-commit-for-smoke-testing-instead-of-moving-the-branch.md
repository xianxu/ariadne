---
id: 000344
status: open
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-12
estimate_hours:
card_mirror: '18fb6e2d4002deca4f11cd0c71d9a700ce81efbf' # card fields mirrored from issue-cards; edit via sdlc
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

## Plan

- [ ]

## Log

### 2026-10-12
