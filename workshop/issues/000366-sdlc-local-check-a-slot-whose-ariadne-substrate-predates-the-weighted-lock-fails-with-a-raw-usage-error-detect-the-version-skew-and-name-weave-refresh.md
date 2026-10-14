---
id: 000366
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '40c5f2c15a02338c3398906fd81a8c9dcc5187e6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T09:00:28-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc local-check: a slot whose ariadne substrate predates the weighted lock fails with a raw usage error; detect the version skew and name 'weave refresh'

## Problem

After ariadne#332 landed (2026-10-14), the shared `sdlc` (`~/workspace/ariadne/bin/sdlc`) passes `--slots N` to the slot's `scripts/test-lock.py`, which comes from the slot's own ariadne substrate (`worktree/<repo>-slotN/ariadne`). Slots whose substrate predates ariadne#346 (pair:2, :4, :5, parley.nvim:2, :3 on 10-14) fail `sdlc local-check run --name suite` with a raw argparse usage error. pair:5 hit it re-closing pair#547. The fix is `weave refresh`, but nothing says so.

## Spec

Before invoking a substrate helper, `sdlc` checks the helper's capability (e.g. a version or capability line, or `--help` for `--slots`) and, on mismatch, refuses with one line naming the slot's substrate path and its commit, and the action `weave refresh` in the checkout. Same for other helpers sdlc calls with newer flags. No automatic refresh (it changes the slot's substrate under a running agent).

## Done when

- With a stale substrate fixture, `sdlc local-check run --name suite` refuses before queueing with the 'run weave refresh' message (test).
- A current substrate runs unchanged.

## Plan

- [ ] Preflight required helper capabilities and refuse before acquisition with supplying checkout/HEAD and an executable repair.
- [ ] Add substrate-only refresh on clean issue branches, preserving consumer Git state and existing dependency safeguards.
- [ ] Verify regressions, update atlas, run full acceptance, then enter close review.

## Log

### 2026-10-14

## Revisions

- 2026-10-14 08:20 — add (TL): the named fix must work on an **issue branch**. `weave refresh` advances the
  repository's own branch too, so it refuses on an issue branch that's ahead of main (pair:5 on pair#547), and
  `--rebase` would rewrite the issue branch. Provide a substrate-only refresh (e.g. `weave refresh --substrate`):
  fast-forward the slot's private substrate checkouts and recompile, leaving the repository's branch alone; the
  refusal message names it.

- 2026-10-14 — Refreshed main-slot5 before claiming: f5b513d43608477ba059418ed9fd84e6068fc705 → 407e48a39057503fab7cdecaf6c69ce338ce7631; compile succeeded. Claimed #366 and entered planning. Read-only consumer inventory found acquireSuiteSlot as the single direct SDLC substrate-script caller; explicit and automatic local-check share it.
- 2026-10-14 — Proposed durable plan: workshop/plans/000366-substrate-helper-refresh-plan.md. Preserve the consumer as an observed read-only subject while private dependencies advance. Capability probes use existing argparse help rather than requiring a new helper protocol. Source-owned helper edge is explicitly submitted for hub review: a no-dependency source branch cannot adopt newer helper source through substrate-only refresh without moving that branch. Implementation awaits plan approval.
