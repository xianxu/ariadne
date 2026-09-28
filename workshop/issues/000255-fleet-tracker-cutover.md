---
id: 000255
status: open
deps: [000252]
github_issue:
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
---

# Fleet cutover to the issue tracker; delete the legacy writers

## Problem

#252 built the issue tracker (cards on `issue-tracker`, details with
`card_mirror`), the one-time migration (`sdlc issue migrate`), the cutover
guard, and a legacy mode that runs the pre-#252 workflow wherever a repository
has not cut over. It proved them through smoke tests (A1, A2), a legacy-mode
soak and a canary cutover of parley.nvim (A3, A4).

What #252 does not do is move the rest of the fleet. Its "existing issue files
are migrated" criterion moves here. The legacy writers, which the binary keeps
only for repositories not yet cut over, are deleted once every repository has
moved.

## Spec

- Follow `workshop/plans/000255-fleet-tracker-cutover-checklist.md` (the
  cutover checklist carried over from #252). It covers Phase B (ship #252,
  unfreeze `ariadne:0`, recompile each repository's `:0`), Phase C (one
  repository at a time: clear blockers, freeze it, dry run, `--apply`,
  read-only verify, first card write, unfreeze), and Phase D (delete the
  legacy writers).
- Per repository, record in the Log: the dry-run counts (cards, inferences,
  duplicates, refusals and how each was cleared), the apply digest, the
  verification, and any bug found — fixed with a regression test, in the
  style of #252's A1–A4 findings.
- Phase D removes: legacy mode (`cmd/sdlc/legacymode.go` and its dispatch),
  `issue publish`, the `issue sync --push` path, the Makefile/Python shell
  fallbacks that allocate IDs or write status, the legacy-equivalence harness,
  and the "legacy repositories" text in help, README, atlas and
  `AGENTS.base.md`.

## Done when

- Every ariadne-layer repository in `~/workspace` has cut over (the dry run
  says *already migrated*), or is recorded here as deliberately excluded
  (brains are excluded by charter; decide astro and parli).
- Existing issue files are migrated in each of them (moved from #252's
  Done when).
- The legacy writers and legacy mode are deleted, with their tests and docs.

## Plan

- [x] Phase B: land #252; unfreeze `ariadne:0`; `weave compile` every `:0`; parley.nvim:0 pulls the cutover
- [ ] parley.nvim (cut over in #252's A4 canary; finish its `:0` catch-up)
- [ ] kaggle
- [ ] xianxu.dev
- [ ] kbench
- [ ] metis
- [ ] you-decide
- [ ] nous
- [ ] 42shots
- [ ] ariadne (clear its dry-run blockers first)
- [ ] tools (clear its blockers first)
- [ ] pair (clear its blockers first)
- [ ] ducks (new project, created 2026-09-27; dry-run not yet taken)
- [ ] astro, parli: decide
- [ ] Phase D: delete the legacy writers and legacy mode

## Log

### 2026-09-27

- Filed from #252 so #252 can close on the tooling. The per-repository
  blockers from the 2026-09-26/27 read-only dry runs are in the checklist;
  re-run each dry run before its cutover.
- Added ducks (`../ducks`), a new ariadne-layer project, to the cutover list.
- Phase B done: #252 landed (PR #135, ec7f997b; archived in e23d2b74) after
  issue review round 13 SHIP. `ariadne:0` unfrozen and rebuilt; parley.nvim:0
  pulled (its 2 local #289 syncs were already on main), reads cards, dry run
  says already migrated; `parley.nvim-slot1/ariadne` back on main.
- `weave compile` across the fleet is only partial, and every failure
  reproduces with the pre-#252 `weave` (see checklist Phase B): metis's seeded
  Makefile (blocks kaggle, kbench), you-decide's tools, parli never compiled.
  These become Phase C step-1 blockers for their repositories.
- Cleared ariadne's C8 blockers 1a/1b/1d: deleted the stale local and remote
  branches; #119 abandoned (wontfix). 1c resolved by #252 landing.
