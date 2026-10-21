---
id: 000516
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '15b08a5e35121bf8836c56f9d72a77a7405b9bec' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ] After #498 lands, claim/start from main and review the smallest snapshot/clock ordering correction.
- [ ] Promote the deterministic reproduction into a permanent regression, apply the correction, and verify adjacent readiness/scheduling controls and the recorded suite.

## Log

- 2026-10-20: Discovered while completing #498, separately approved for filing and later pickup by ops:0 (4299f256-1c1f-4790-b39b-9b6d25545b20). #498 must land first; no production correction is on its branch.
- Controlled-clock red/green on detached 73f469513b02c34d39cadf9dd613734b89ba2c51: readiness callback advances the clock by 1ms and refreshes a ready waiting car. Current code returns trained=false, no error, train alone/no-other-ready (test1.24s). A TEMP-only resample immediately after reading readiness makes the same test return trained=true with real fake train PR100 merged and Cars:[2 3], Result:landed (test3.37s). This is a production-path test, not just a Board arithmetic test. No fixed sleep, load generator or current-branch production edit was used.
- Persistent evidence: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/evidence/000498/readiness-diagnosis/{red.log,green.log,clock-resample.patch,readiness_diagnosis_test.go}. The baseline failure log is /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/6537ba8d2b38cce637a077e8f11be119.log. Evidence is off the RAM disk.
