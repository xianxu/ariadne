---
id: 000516
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '167689d3a8bb3cbc4ec40a4ccfc2ff10a0f2abbe' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T23:05:19-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainWaiterAcceptsConductorArchive]
---

# Train boarding rejects readiness refreshed during conductor checks

## Problem

The conductor can discard a valid refreshed waiter as future/stale. maybeConductTrain samples landingNow before its own potentially slow readiness checks, then reads waiting-car readiness and passes the earlier timestamp to Board. A waiter refreshed during those checks has At greater than that sample, so Readiness.Current rejects it. Refreshing readiness should not make a ready car disappear.

Observed on main baseline 94bf14d9262f4ced7db0ebdb43b44583d7d8e2d5: TestLandingTrainWaiterAcceptsConductorArchive failed in 10.86s with train Cars:[2], Result:alone, Detail:no other car is ready; the complete suite took 727.5s (attempt 6537ba8d2b38cce637a077e8f11be119). This is separate from #498/#500 completion ceilings. The historical log lacks vouch timestamps, so the exact historical interleaving is not established; the deterministic production-path reproduction independently proves the matching race.

## Spec

Board against a clock sample consistent with the readiness snapshot, after the conductor's own gate checks and readiness reads. Keep genuine stale/future-vouch refusals and train scheduling semantics intact. Preserve the archive fixture's real two-car train, outcome delivery and cleanup ordering; do not pause refreshes or add sleeps to hide the product bug. No persisted-format change or completion-ceiling increase.

## Done when

- A controlled-clock production-path test synchronously refreshing a waiting car during the conductor's own readiness callback boards both cars and completes the real fake GitHub train; old ordering fails deterministically without sleeps/load.
- Existing genuinely stale/future readiness controls remain valid; real multi-car and archive conformance pass.
- Recorded suite and fresh review pass; any associated flaky fixture registration is removed only after proof, and the fix lands.

## Plan

- [x] After #498 lands, claim/start from main and review the smallest snapshot/clock ordering correction.
- [ ] Promote the controlled-clock refresh reproduction with stale/future controls, resample after readiness reads immediately before Board, and verify real archive/multicar controls.
- [ ] Run the recorded suite, remove the associated flake registration after proof, and close with the binary-owned fresh review.

## Log

- 2026-10-20: Discovered while completing #498, separately approved for filing and later pickup by ops:0 (4299f256-1c1f-4790-b39b-9b6d25545b20). #498 must land first; no production correction is on its branch.
- Controlled-clock red/green on detached 73f469513b02c34d39cadf9dd613734b89ba2c51: readiness callback advances the clock by 1ms and refreshes a ready waiting car. Current code returns trained=false, no error, train alone/no-other-ready (test1.24s). A TEMP-only resample immediately after reading readiness makes the same test return trained=true with real fake train PR100 merged and Cars:[2 3], Result:landed (test3.37s). This is a production-path test, not just a Board arithmetic test. No fixed sleep, load generator or current-branch production edit was used.
- Persistent evidence: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/evidence/000498/readiness-diagnosis/{red.log,green.log,clock-resample.patch,readiness_diagnosis_test.go}. The baseline failure log is /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/6537ba8d2b38cce637a077e8f11be119.log. Evidence is off the RAM disk.

- Pickup after #498 landed as PR #355: fresh issue branch from main. Quick plan posted to hub, proceeding unless it objects; live-conformance decision requested. Keep the early clock for pre-board serial exits, resample immediately after the readiness snapshot for boarding/departure. ARCH-ORDER/PURE: fix ordering at the existing injected clock seam, retain pure Board/Current semantics. ARCH-DRY/PURPOSE: one production Board call site; no fixture workaround. ARCH-MOCK: reuse stateful Git/GitHub train setup. ARCH-CONSTRAINTS: one additional O(1) clock read per conduction; no new loops or capacity. ARCH-ERROR/FUNERAL/EVOLVE: no new error, resource, durable family or format. A lexical lint cannot prove relative clock/snapshot ordering across callbacks; the controlled interleaving regression enforces this semantic race.
- Registry evidence: original archive test passed unchanged on the exact failed baseline 94bf14d9262f4ced7db0ebdb43b44583d7d8e2d5 in18.11s (package18.597s), clean detached checkout with identical Git HEAD before/after. baseline-original.log and baseline-result.json are in the persistent diagnosis evidence directory. Only TestLandingTrainWaiterAcceptsConductorArchive is registered here, not the #498/#500 completion flakes.

## Revisions

- 2026-10-20: After #498 landed, refine pickup checklist into implementation and verified close steps; acceptance and scope unchanged.
