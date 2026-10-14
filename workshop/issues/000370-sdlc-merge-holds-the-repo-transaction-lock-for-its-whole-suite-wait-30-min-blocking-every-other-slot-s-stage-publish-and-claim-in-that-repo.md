---
id: 000370
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: 'af877ffe20a60ba2c7e6ee5cc95637cac127612e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T11:38:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc merge holds the repo transaction lock for its whole suite wait (30+ min), blocking every other slot's stage, publish and claim in that repo

## Problem

Since ariadne#332 (sdlc merge integrates main and runs the full suite before landing), `sdlc merge` holds the repo's sdlc transaction lock for its whole run, including the suite queue wait and the suite itself, 30+ minutes on 2026-10-14, and a second run when main moved. Meanwhile every other slot in that repo is blocked from `sdlc stage` / `stage --clear`, `issue new` / `issue publish`, `claim`, `move` and other locked verbs: pair:6's stage clear waited on pair:3's merge (pid 3229); the TL's `issue new` for this very issue waited ~1h on ariadne:2's merge (pids 81291, 37511, which started a second suite); pair:6's P0 hotfix filing waited too. The operator made this a priority fix.

## Spec

Hold the transaction lock only around short git/tracker mutations: integrating main, recording state, the landing push, archive. Release it for replay review, GitHub registration polling, the CI wait and the suite (queue and run). On every reacquire, revalidate the exact branch HEAD and worktree and fetch main: HEAD drift refuses; main drift re-integrates and reruns the gates. Keep the integration state and the interrupted-merge recovery contract from #332 (crash between phases resumes or reports cleanly).

## Done when

- While a merge's suite is queued or running, another slot's `issue new`, `stage`, `claim` in the same repo proceed (test).
- HEAD drift during the suite refuses; main drift re-integrates and reruns; a crash between phases recovers (tests).

## Plan

- [ ] Scope `sdlc merge` with the existing manual repo-lock helper and prove a second same-repo command can acquire the lock during both a first and a repeated suite run.
- [ ] Revalidate exact workspace/head and fresh main at each lock reacquisition around integration push and final landing; add deterministic head/main drift and interruption-retry tests.
- [ ] Update workflow help and atlas, run targeted tests, `make -k tools`, and the recorded full suite; close and ship after the boundary review.

Durable implementation steps: [merge-lock-scoping plan](../plans/000370-sdlc-merge-holds-the-repo-transaction-lock-for-its-whole-suite-wait-30-min-blocking-every-other-slot-s-stage-publish-and-claim-in-that-repo-plan.md).

## Log

### 2026-10-14

- Live evidence: ariadne:2's merge held `.git/sdlc.lock` through its first `make test` and a second full suite after main advanced; `sdlc move :5` for #240 waited until that merge finished. The lock interval exceeded 30 minutes and delayed this issue's own creation. ARCH-CONSTRAINTS: the suite queue/run is unbounded relative to a short repository transaction.
- Design: reuse `withRequiredRepoTransactionLock` (ARCH-DRY); keep `integrationState` as the only transition owner (ARCH-ORDER); use existing commit, lease, PR-head and archive evidence across unlocked phases (ARCH-FUNERAL/ARCH-PURPOSE).
- Plan-quality round 1 (REWORK): accepted findings PQ-1 (legacy `runCIGate` can run a suite), PQ-2 (main drift after replay needs a reset/retry transition), and PQ-3 (named-function test strategies); a Minor requested explicit non-goals. The plan revision addresses each class before code.
- Plan-quality round 2: CLEAN; PQ-1 through PQ-4 addressed, with no new findings. The estimate is derived after that review, per the gate's ordering.

## Revisions

- 2026-10-14 11:42 PDT — Plan-quality review exposed a second suite-running merge path and a pre-push state gap. Expanded the plan to cover legacy merge, added a pure reset/retry transition and function-level test strategies, and renamed the durable plan to the canonical issue slug. The Spec and Done-when remain unchanged.
