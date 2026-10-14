---
id: 000370
status: done
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours: 4.82
card_mirror: 'bc940e39ebc1d19c880edb58c92bcf2c6159ec5c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T11:38:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 1.58
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

- [x] Scope `sdlc merge` with the existing manual repo-lock helper and prove a second same-repo command can acquire the lock during both a first and a repeated suite run.
- [x] Revalidate exact workspace/head and fresh main at each lock reacquisition around integration push and final landing; add deterministic head/main drift and interruption-retry tests.
- [x] Update workflow help and atlas, run targeted tests, `make -k tools`, and the recorded full suite.

Delivery follows the boundary review: close and ship the reviewed head.

Durable implementation steps: [merge-lock-scoping plan](../plans/000370-sdlc-merge-holds-the-repo-transaction-lock-for-its-whole-suite-wait-30-min-blocking-every-other-slot-s-stage-publish-and-claim-in-that-repo-plan.md).

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.5
item: api-integration         design=0.75 impl=0.60
item: api-integration         design=0.75 impl=0.60
item: cross-cutting-refactor  design=0.30 impl=0.20
item: atlas-docs              design=0.10 impl=0.08
item: milestone-review        design=0.10 impl=0.20
design-buffer: 0.15
total: 4.82
```

The two API-integration items cover the durable and legacy Git/GitHub landing paths. Each starts from a 1.5-hour point in v2's 1–3 hour design range, discounted by 0.5 to 0.75: the contract is settled, while exact legacy phase extraction and the concurrent fixture orchestration still need design work. Each uses 1.5 v2 implementation hours scaled to 0.60 by v3.1. Those implementation items include the path-specific real-lock proofs, repeated-suite and head/main drift, interruption recovery, and elapsed recorded-suite/CI/landing checks. The cross-cutting lock refactor uses 0.30 design and 0.50 × 0.40 = 0.20 implementation hours, including the pure transition and lock metadata unit tests; atlas/docs uses 0.10 and 0.20 × 0.40 = 0.08; one close review uses 0.10 and 0.50 × 0.40 = 0.20. Familiarity is 1.5 because concurrency/recovery behavior is novel but bounded by existing seams. The thorough plan earns the v3.1 15% design buffer: 2.00 × 1.15 + 1.68 × 1.5 = 4.82 ship-wall-clock hours.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* The calibration source is marked stale by `sdlc estimate-source`, so the per-primitive hours are provisional.

## Log

### 2026-10-14
- 2026-10-14: closed — Real-lock probes show other slots acquire during blocked/repeated suites and PR polling; pinned legacy HEAD and deleted-remote recovery, durable main/head drift, and fixture lifetime pass regression tests; make -k tools and recorded suite on bd47ebda passed (12 shards, 1353 cmd/sdlc tests, 54 packages, zero failures).; review verdict: SHIP

- Live evidence: ariadne:2's merge held `.git/sdlc.lock` through its first `make test` and a second full suite after main advanced; `sdlc move :5` for #240 waited until that merge finished. The lock interval exceeded 30 minutes and delayed this issue's own creation. ARCH-CONSTRAINTS: the suite queue/run is unbounded relative to a short repository transaction.
- Design: reuse `withRequiredRepoTransactionLock` (ARCH-DRY); keep `integrationState` as the only transition owner (ARCH-ORDER); use existing commit, lease, PR-head and archive evidence across unlocked phases (ARCH-FUNERAL/ARCH-PURPOSE).
- Plan-quality round 1 (REWORK): accepted findings PQ-1 (legacy `runCIGate` can run a suite), PQ-2 (main drift after replay needs a reset/retry transition), and PQ-3 (named-function test strategies); a Minor requested explicit non-goals. The plan revision addresses each class before code.
- Plan-quality round 2: CLEAN; PQ-1 through PQ-4 addressed, with no new findings. The estimate is derived after that review, per the gate's ordering.
- Change-code admitted the full flow. Estimate-quality returned INFO: clarify the design-range point, test coverage and elapsed shipping work; the derivation now says where each belongs without changing the reconciled total.
- Implementation: manual lock phases now bracket shared Git mutations, expected-head merge, and archive. Durable publish replay, GitHub registration and CI/suite waits are unlocked; legacy local checks are unlocked. A pre-push main move resets the unpublished integration and retries. Added real-lock probes during a blocked suite, repeated suites, and registration polling, plus exact-head/main drift tests. Targeted merge tests and `make -k tools` pass; recorded suite pending.
- Verification: pushed code head `230f61e2f43550c0405bb9c5b7bbf4dc7666f975`; `sdlc local-check run --name suite` passed and posted `sdlc-local/suite` on that head (12 shards, 1,349 cmd/sdlc tests, 54 other packages, zero failures; log SHA-256 prefix `e276d37509b6`). Focused lock/drift tests, `make -k tools`, and `git diff --check` passed. Review and landing remain.
- Close review round 1: REWORK, BR-1 Critical (`pinned-evidence-across-lock-gaps`) and BR-2 Important (`pre-gate-freshness`). Reproduced both with red tests: a concurrent commit/push after the legacy publish gate reached `PRMerge`; main moving during GitHub registration caused two suites, the first on stale main. The fixes pin and recheck the legacy HEAD and re-observe main before the first durable suite; both tests now pass. The recorded suite must rerun on the new code head.
- Recorded suite on `d91a5c14` found one compatibility failure: `TestRunMerge_DirtyAfterJudge_RefusesPreMerge` expected the existing “before the irreversible merge” recovery guidance. The earlier checkpoint refused safely but changed the message. Restored the guidance in the shared revalidation error; all `TestRunMerge_` tests and the registration-drift regression now pass. A new recorded suite is required for this follow-up code head.
- Verification after review fixes: `sdlc local-check run --name suite` passed on pushed head `ac688959a79255c7a17e256af44992c62b9b418d` (12 shards, 1,351 cmd/sdlc tests, 54 other packages, zero failures; log SHA-256 prefix `b2758c2ce53c`). The red-before/green-after BR-1 and BR-2 tests, merge regression tests, `make -k tools`, and `git diff --check` passed. Close review round 2 pending.
- Close review round 2: REWORK; BR-1/BR-2 addressed, new BR-3 Critical (`recovery-preconditions-respect-completed-effects`) and BR-4 Important (`concurrent-fixtures-own-worker-lifetime`). A deleted remote branch plus pruned tracking ref reproduced BR-3 before the fix; confirmed-landing recovery now uses local-head/main proof and new landing keeps strict remote-head checks. Blocked-suite tests now cancel and join workers during cleanup, with a cleanup-order regression. Full-suite verification pending on this new head.
- Verification after round-two fixes: pushed head `bd47ebda94789249fd3a45a69334f0712cb634d4`; recorded `sdlc-local/suite` passed (12 shards, 1,353 cmd/sdlc tests, 54 other packages, zero failures; log SHA-256 prefix `7fdaf1fbefc5`). Broad merge/landing targeted tests, `make -k tools`, and `git diff --check` passed. Close review round 3 pending.

## Revisions

- 2026-10-14 11:42 PDT — Plan-quality review exposed a second suite-running merge path and a pre-push state gap. Expanded the plan to cover legacy merge, added a pure reset/retry transition and function-level test strategies, and renamed the durable plan to the canonical issue slug. The Spec and Done-when remain unchanged.
- 2026-10-14 11:49 PDT — Clarified the admitted estimate after estimate-quality INFO: 1.5 is a chosen point within the v2 design range, not its midpoint; named the remaining design work and mapped tests plus ship-wall-clock verification to the integration items. Numeric estimate and acceptance contract unchanged.
- 2026-10-14 12:19 PDT — Split the completed verification step from its subsequent close-and-ship delivery, so the Plan checklist records implemented work before the close gate runs. The acceptance contract is unchanged.
