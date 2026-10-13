---
id: 000305
status: working
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-12
estimate_hours:
card_mirror: 'bbc031aba9c39bcf40a3b44aa9f46f04698ff7c3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T18:48:46-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Reopening a done issue does not un-archive its details

## Problem

Filed from #286. A merged (`done`) issue's details are archived to `workshop/history/issues/` on main. Reopening it (`set-status working`) leaves them there, and no issue branch is restored. Work resumed on a fresh branch either re-creates `workshop/issues/X` beside the archived copy, or merges main and has git follow the archive rename, which puts the live details back in `history/`.

#286 solved this for abandoned issues (`restoreAbandoned`: merge main, then move the archived details and plans back out in one commit). A `done` issue has no kept branch, but the un-archive half applies as is.

## Done when

- Reopening a `done` issue whose details are archived on main moves them (and its plan artifacts) back to `workshop/issues/` and `workshop/plans/` on the reopened branch, so its next landing un-archives them cleanly through the transfer guard.

## Spec

Today a `done` reopen cannot work at all in a tracked repository: the reopen
guard reads the Log from `workshop/issues/X`, which is archived (so it refuses
unless `--force`), and `start-plan` then refuses because the details are not on
main's live path.

`sdlc issue set-status working` on a `done` card whose details are archived on
main (history copy present, live path absent) restores them before the card
write, mirroring #286's abandon reopen (ARCH-DRY: one shared un-archive step):

1. Refuse a foreign owner; run from the resting branch or the issue's branch
   (a rerun).
2. On rest, fast-forward to main (as `claim` does), so the archived copy is in
   the checkout.
3. The reopen's Log entry goes into the archived copy in place
   (`workshop/history/issues/X`); the guard reads it there, before anything
   changes (`--force` skips it as before). That file is the only dirt allowed.
4. Create the issue branch at main (or reuse it on a rerun), `git mv` the
   details and the issue's plan artifacts back (the Log edit rides along), and
   commit once: `#N: reopen: un-archive`. The same name is safe: `sdlc pr` opens
   a new PR, and #148 refuses only when no open PR exists.
5. Push the branch, then the card write (done → working, claimant), then the
   mirror commit and push (`finishReopen` without an archive ref).

Each step is recognised on rerun (branch exists, HEAD already un-archived).
Creates nothing durable beyond the issue branch, which its landing removes
(ARCH-FUNERAL). A legacy repository and a done issue not archived on main are
unchanged.

## Plan

- [x] Extract the un-archive step from `restoreAbandoned` into a shared helper.
- [x] `restoreDone` + wire it into `reopenAbandoned` (renamed for both cases).
- [x] Tests: a done issue reopened from rest lands its details live on a pushed
      branch that passes the transfer guard, plans included; the guard refuses
      without the Log entry and changes nothing; a rerun after a failed push
      un-archives once.
- [x] Help (`set-status.md`) and recovery catalog.

## Log



- 2026-10-13: closed — make test green at 225c7da3 (origin/main incl. #346 merged; 1112 cmd/sdlc tests, 0 failed); round 3: BR-7 TestCheckUnarchived (postcondition fails on the landed branch, holds after reopen), BR-6 class TestReopenDoneDirtyTreeRefusesBeforeTheFastForward + TestReopenDoneGuardRefusalNamesTheFastForward; earlier rounds and done-when TestReopenUnarchivesADoneIssue still green; review verdict: SHIP
- 2026-10-13: flow upgraded quick → full — 327 added lines in code files (limit 100); an earlier round of this close already ran the full review
### 2026-10-08
- 2026-10-12: implemented (`cmd/sdlc/reopendone.go`, shared `unarchiveIssue` in abandon.go); `go test -run TestReopen` green. Paused on the TL's quota hold before the help/catalog row, `change-code` and close. Next: help + catalog, then `sdlc change-code --issue 305`, the full suite, then `sdlc close`.
- 2026-10-13: resumed after the hold; merged origin/main (clean); help + catalog row (`TestReopenUnarchivesADoneIssue`, `TestReopenDoneRerunAfterAFailedPush` as proofs); catalog/help tests green.
- 2026-10-13: close round 1 FIX-THEN-SHIP. Fixed all four: BR-1 (stale tracking ref → `leaseOnRemoteBranch` refreshes or drops it before the push), BR-2 (a lost finish is recognised by `unfinishedDoneReopen` from the un-archive commit on top), BR-3 (an unlanded local branch is checked before switching), BR-4 (refusal and --force tests). BR-1/BR-2 tests mutation-checked (fail without the fix). Lesson added.
- 2026-10-13: close round 2 REWORK. BR-5 (Critical): a reopen from the landed branch itself pushed the old tip un-restored. Fixed by the class: every branch decision is made before the first change (keyed on whether the branch holds main's archive commit), the landed branch refuses toward the resting branch, and the un-archived postcondition is checked before the push. BR-6: the guard refusal names the fast-forward. Tests: `TestReopenDoneRefusals/the landed branch itself` (mutation-checked), `TestReopenDoneGuardRefusalNamesTheFastForward`.
- 2026-10-13: close round 3 FIX-THEN-SHIP. BR-7: the postcondition (`checkUnarchived`) now requires HEAD to descend from main's archive commit, tested directly on the landed branch (`TestCheckUnarchived`). BR-6 by the class: the dirty-tree check moved above the fast-forward (`dirtyBesides` exempts both copies), and every later refusal names the move through one helper (`TestReopenDoneDirtyTreeRefusesBeforeTheFastForward`).
