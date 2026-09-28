---
id: 000255
status: open
deps: [000252]
github_issue:
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
card_mirror: 'babc5f2ddf57a41726035e5b0a0d85a3000f209c' # card fields mirrored from issue-cards; edit via sdlc
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
- ~~The legacy writers and legacy mode are deleted, with their tests and
  docs.~~ Moved to #258 (see Revisions, 2026-09-28).

## Plan

- [x] Phase B: land #252; unfreeze `ariadne:0`; `weave compile` every `:0`; parley.nvim:0 pulls the cutover
- [x] parley.nvim (cut over in #252's A4 canary; finish its `:0` catch-up)
- [x] kaggle
- [x] xianxu.dev
- [x] kbench
- [x] metis
- [x] you-decide
- [x] nous
- [x] 42shots
- [x] ariadne (clear its dry-run blockers first)
- [x] tools (clear its blockers first)
- [x] pair (clear its blockers first)
- [x] ducks (new project, created 2026-09-27)
- [x] astro, parli: cut over (operator: issues will improve these repos too)
- [x] Phase D: delete the legacy writers and legacy mode → spun off to #258 (deferred, `punt`)

## Log


- 2026-09-28: closed — Fleet cutover complete: 15 work repositories report already migrated (per-repo digests and roots in the Log table); exclusions recorded (brains by charter; metis.bak and test-repo: no remote, no issues). Tool bugs found were fixed with regression tests: #256, #257. #241 layout adopted fleet-wide as step 0. Checklist rows all resolved; Phase D spun off to #258 (punt, revisit by 2026-11-30). --no-actual: sdlc actual under-attributes (#254).; review verdict: SHIP
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
- Every cutover now starts with step 0: the repository on #241's weave
  layout. Most of the fleet never adopted #239 (symlinked Makefiles,
  pre-inventory generated output), so `weave compile` fails there. Adoption
  procedure and findings in #241's Log; metis, kaggle, kbench adopted
  (committed locally, unpushed, as each carries earlier unpushed operator
  commits).
- **Migration gap (to fix before nous cuts over):** `PlanTrackerMigration`
  skips a branch whose only edit to an issue is deleting it ("a branch's own
  archive move"). nous's local `000048-shim-oauth-…` branch closed #48 and
  archived it as done without landing (≈1.2k lines, 2026-06-11), while main
  says `working`; the dry run reports 0 refusals. After cutover the card
  would stay `working` and landing the branch would modify/delete-conflict
  with the converted details. Fix: refuse "a branch archives an issue main
  still has active; land or drop it before cutover", with a test. For nous:
  land #48 in legacy mode (merge main in, PR) before its cutover.
- nous branch cleanup: deleted merged `nous-14` (+ worktree) and `branch-32`;
  kept #48's branch (operator: keep, land later).
- nous #48 landed (nous PR #10, 2026-09-27; `done`, archived): nous's
  archived-on-branch blocker is gone. The migration-tool gap it exposed still
  needs its refusal + test before any further cutover.
- The archived-on-branch migration gap is fixed by #256: the dry run now
  refuses a branch that archives an issue main still has active (renames and
  deletions main made too stay silent). Dry runs with it: pair, kaggle, kbench
  0 refusals; ariadne 2 pre-existing (`origin/feature` edits archived #148,
  #159 — drop or land that branch before ariadne's cutover).
- **ariadne cut over (2026-09-27), moved ahead of kaggle** because every
  legacy-mode ariadne change re-created the copy-publication conflict #252
  removes (#256's PR hit it). Step 0: `weave compile` already clean. Blockers:
  deleted `origin/feature` (a July test-fixture push editing archived #148,
  #159) and six merged remote branches. Dry run: 0 refusals, 251 cards, 50
  details converted, 35 inferences (N/A actuals, Problem from preamble,
  `## Problem` inserted in #15, #23, #114, #123, #130, #131), 2 duplicate IDs
  (000040, 000096). Applied `f810548c9edd8a73`: tracker root aee6be5b,
  migration commit 0c567478. `:0` and slots 1–3 fast-forwarded. Verified: issue
  list (`:0` and a slot), already migrated, recovery empty. First card write:
  claimed #255 → `issue-tracker` 38d41d8b, main untouched.
- **Batch 1 cut over (2026-09-27):** kaggle (root e9029056), kbench
  (0d475da7), metis (e81d3097), xianxu.dev (222a1b75), ducks (2e02bdc1),
  you-decide (d8ba181f). All dry runs 0 refusals; each `:0` pulled; verified
  (issue list, already migrated, recovery empty). kbench's pre-cutover branch
  `000028-turn-level-conductor` needs `--reconcile` before reuse.
- you-decide's apply stopped between bootstrap and the main commit: its
  pre-push publish gate ran lint-ids, which judged the checkout's (absent)
  marker instead of the pushed commit's. Finished by re-running the same
  `--apply` with hooks skipped once (operator-authorized); fixed properly in
  #257 (lint-ids judges the marker at `--head`; landed PR #137 — ariadne's
  first tracker-mode issue, and the PR merged with no issue-file conflict).
- Remaining: pair, tools, nous (batch 2); 42shots, astro, parli (no issues:
  cut over or exclude).
- **Batch 2 and the issue-less repos cut over (2026-09-28):** pair (root
  c039dc4b; 332 cards, 95 details; slots 1–2 fast-forwarded; #292's branch
  reconciles on next use), nous (bccecf82), tools (ea8ec461), and — operator:
  "issues are there to improve those repos too" — 42shots (1160c082), astro
  (f18ab102), parli (74a1985d; slot 1 fast-forwarded) with empty trackers.
  tools first published #82's local design (filed 2026-09-27 23:09, legacy
  divergence) with `issue publish --commit`, so its migration converted the
  full body. All verified: issue list, already migrated, recovery empty.
- **Every non-brain ariadne-layer repository in `~/workspace` is on the issue
  tracker** (sweep: 42shots, ariadne, astro, ducks, kaggle, kbench, metis,
  nous, pair, parley.nvim, parli, sdlc-smoke, tools, xianxu.dev, you-decide).
  Brains (brain, brain-family, brain-private, robotics) are excluded by
  charter. Done-when bullets 1–2 hold; Phase D (delete the legacy writers and
  legacy mode) remains.
- Phase D deferred by the operator (2026-09-28): sdlc changes rarely, so keep
  legacy mode and the legacy writers for a while as a fallback in case the
  cutover missed something; delete them later. The `sdlc-smoke` sandbox repo
  (#252 smoke test A2) is being deleted from GitHub by the operator.
- robotics (2026-09-28): operator: a work repository not yet started, not a
  brain. Dropped `.brain/` and the empty `nous brain new` scaffolding (1c113dc;
  nous's brain watcher drops a directory without `.brain/config.md` on its next
  scan — one stray autosave was folded into the adoption commit), adopted the
  #241 layout (fcaf0ec), created private `xianxu/robotics` as origin, and cut
  over with an empty tracker (root 3656af77). The `sdlc-smoke` sandbox is
  deleted (operator). Brains remaining legacy by charter: brain, brain-family,
  brain-private.
- **Per-repository results** (dry run → apply; all 0 refusals, verified
  already migrated with recovery empty):

  | Repository | Cards (inferred), details converted, duplicate IDs | Digest | Tracker root |
  |---|---|---|---|
  | parley.nvim | 290 (A4 canary, #252) | — | 41de715a |
  | ariadne | 251 (35), 50, 2 | f810548c9edd8a73 | aee6be5b |
  | kaggle | 10 (0), 5, 0 | 49c1cf6c94054ad5 | e9029056 |
  | kbench | 28 (0), 5, 0 | d31157b817c449dd | 0d475da7 |
  | metis | 69 (0), 15, 0 | e3036e98d2889a3a | e81d3097 |
  | xianxu.dev | 4 (0), 2, 0 | 69553139df1b397a | 222a1b75 |
  | you-decide | 14 (4), 10, 0 | 48f7a22721f27216 | d8ba181f |
  | ducks | 11 (0), 11, 0 | dc1b0ca2024a9bb9 | 2e02bdc1 |
  | pair | 332 (38), 95, 1 | d4815982eadc2ec3 | c039dc4b |
  | nous | 53 (15), 16, 0 | 892406db4b8ff766 | bccecf82 |
  | tools | 82 (1), 23, 0 | aee2ef2c0cb7e0cb | ea8ec461 |
  | 42shots | 0 | 95699770925b8bac | 1160c082 |
  | astro | 0 | 933eaf04121344b9 | f18ab102 |
  | parli | 0 | 6daea0c5988c8506 | 74a1985d |
  | robotics | 0 | ada41ec698ac9acb | 3656af77 |

- **Excluded** (close review BR-1): brains (brain, brain-family,
  brain-private) by charter; `metis.bak` and `test-repo`, single-commit
  scaffolds with no remote and no issues — a tracker needs a publication
  remote. The 2026-09-28 sweep line above listed `sdlc-smoke` as cut over; it
  was the #252 A2 sandbox and has since been deleted, so it is not part of the
  fleet.
- Close review BR-2: the cutover checklist now records every C-row's result
  (digest, root), C11 superseded by A4, C12 decided, C13 robotics, the
  exclusions, and Phase D pointing at #258; the live tracker-mode
  `milestone-close` was not exercised (marked skipped).

## Revisions

### 2026-09-28 — Phase D spun off to #258

Reason: every work repository is on the tracker, and the operator chose to
keep legacy mode for a while as a fallback (sdlc changes rarely). Holding #255
open on one deliberately deferred row served no purpose.

- Done when: "the legacy writers and legacy mode are deleted" moves to #258,
  filed `punt` with the removal list from this Spec. #255 closes on the fleet
  cutover (Done when 1–2).
- Plan: the Phase D row points at #258.
