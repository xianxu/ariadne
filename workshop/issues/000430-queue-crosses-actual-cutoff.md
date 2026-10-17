---
id: 000430
status: codecomplete
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'aaf87d71804ab3e1adb74a9f1a10e477b5b2cc48' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T14:46:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "d8cd23ad", done: "5b60a1bb"}
actual_hours: 0.27
---

# Actual refuses a queue wait crossing the measurement cutoff

## Problem

pair#650's `sdlc actual --issue 650` refuses: suite attempt `2a875dd8cd12fd8a37285519857d2012` has zero matching pinned slot engagements. The attempt belongs to `pair:8` and queued from 12:24:30 to 12:48:27 PDT. The issue's latest commit at measurement is 12:26:08, so its pinned engagement is clipped at 12:26:08. `selectOwnedQueueAttempts` compares the *whole* queue wait with that clipped engagement, even though it only charges queue time inside the measurement window. Qoder's transcript exists at the pinned worktree path; the claimant is unchanged across the tracker history. The cutoff mismatch, not transcript discovery or relaunch identity, causes the refusal.

## Spec

An exact queue wait that crosses the measurement cutoff contributes only its intersection with the measured interval. Prove that selected intersection belongs to one pinned engagement for the attempt's slot; ownership after the cutoff is outside the sample and need not be inferred. A release, reclaim, or ownership gap *inside* the selected interval still refuses a numeric actual. Keep attempts from other slots excluded and legacy/unknown provenance conservative. ARCH-PURE: make this decision in the existing pure queue selector; keep record and tracker IO unchanged.

## Done when

- A regression derived from pinned tracker ownership and pair#650's timing admits the measured queue portion and no longer reports zero matching engagements.
- A release inside the measured queue portion still refuses, while a release after the measurement cutoff cannot change the already pinned portion.
- Targeted actual/queue tests, `make -k build`, and the recorded exact-head suite pass.

## Plan

- [x] Add a failing selector regression using a queue wait that starts before and acquires after the last issue commit; derive the engagement with the same cutoff used by actuals.
- [x] Select the clipped interval first, then prove sole slot ownership for that interval; retain the release/gap and foreign-slot guards (ARCH-DRY).
- [x] Run focused tests and build, then record a passing exact-head suite before close.

## Log

### 2026-10-17
- 2026-10-17: closed — pair#650-shaped exact queue cutoff test failed before the fix and passes after; release/gap inside the clipped wait still refuses, release after cutoff does not; focused selector and owned-actual resolver tests pass; make -k build and make sdlc-build pass; recorded suite on fa36207fca59a0363cc740d130b0acd0e3ad867d passes 12 shards, 67 packages, zero failures.; review verdict: SHIP

- Root-cause evidence: the saved attempt is exact and names `pair:8`; tracker commits at 12:13:30 and 12:14:35 PDT pin the same claimant. The code-head's last issue commit is 12:26:08, inside the queue wait. `DeriveEngagements(events, until)` truncates its owner window there, then the selector requires acquisition at 12:48:27 to precede that truncated end. Read-only inspection only; pair:8's checkout was not changed.
- TDD: the pair#650-shaped case failed with `0 matching pinned slot engagements` before the change, then passed after clipping. It covers a continuing claim, a release after the cutoff, a release inside the measured wait, and a release/reclaim gap inside it. The existing selector and legacy controls plus owned-actual resolver tests passed (60.6s). `make -k build` and `make sdlc-build` passed; the former correctly skipped the sentinel-owned `sdlc` binary, which the latter built.
- Read-only sweep of recorded exact attempts since October 16 across pair, ariadne, and parley.nvim found same-issue commits during a queue wait for pair#622, pair#628, pair#648, pair#650, ariadne#429, and parley.nvim#373. These are potential instances of the cutoff shape, not confirmed numeric-actual failures; a later measurement endpoint can remove the crossing. The bug is harness-independent.
- Recorded suite on pushed code head `fa36207fca59a0363cc740d130b0acd0e3ad867d` passed in 9m39s: 12 shards over 1,577 sdlc tests and 67 other packages, zero failures; saved log hash `a030590b5101`. The stage/live check follows SHIP, as granted by the hub.
