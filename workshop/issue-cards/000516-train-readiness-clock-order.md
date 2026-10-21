---
id: 000516
status: open
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainWaiterAcceptsConductorArchive]
started: 2026-10-20T23:05:19-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Train boarding rejects readiness refreshed during conductor checks

## Problem

The conductor can discard a valid refreshed waiter as future/stale. maybeConductTrain samples landingNow before its own potentially slow readiness checks, then reads waiting-car readiness and passes the earlier timestamp to Board. A waiter refreshed during those checks has At greater than that sample, so Readiness.Current rejects it. Refreshing readiness should not make a ready car disappear.

Observed on main baseline 94bf14d9262f4ced7db0ebdb43b44583d7d8e2d5: TestLandingTrainWaiterAcceptsConductorArchive failed in 10.86s with train Cars:[2], Result:alone, Detail:no other car is ready; the complete suite took 727.5s (attempt 6537ba8d2b38cce637a077e8f11be119). This is separate from #498/#500 completion ceilings. The historical log lacks vouch timestamps, so the exact historical interleaving is not established; the deterministic production-path reproduction independently proves the matching race.
