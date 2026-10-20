---
id: 000498
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '88eac5970b16a64f29cc832434a044f2aca4b516' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T14:48:54-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/train]
flow: {kind: quick, provenance: inferred, spec: "8876058c", done: "15e3cf4e"}
---

# flake: TestLandingTrainLandsEveryCarOnOneSuite/train

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/train: Same exact code head 68bfefbde7c363252ba67ab2ddcfd05ca1aaa48b: required recorded full suite attempt 11cab56f89656e3bab43c5a95e196394 completed all eight shards, 1854 cmd/sdlc tests and 75 other packages; only /train failed at landingtrain_test.go:414 finished() two-minute merge ceiling, variant 121.77s, parent 138.62s. Failed log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/11cab56f89656e3bab43c5a95e196394.log, SHA256 prefix 8930df16c1d9. In a clean detached checkout of that exact commit, go test ./cmd/sdlc -run ^TestLandingTrainLandsEveryCarOnOneSuite$ -count=1 -timeout=5m -v passed all variants: train 6.94s, null 7.12s, head 6.90s, parent 20.96s, package 21.500s. Git rev-parse before and after printed the same exact SHA. Focused log persisted at /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/evidence/000494/train-parent-same-head.log. This is an explicit same-head fail-to-pass classification; only /train is registered, not siblings. Prior #480 registered /head and landed the eight-shard envelope; this new observation shows the train wait can still intermittently exceed its ceiling at eight shards.

## Spec

The shared fixture completion wait derives its ceiling from half the time remaining before testing.T.Deadline(), after the test gate has accounted for queueing. This uses the actual supervising run budget and leaves time for a named failure plus all goroutine stacks before Go's alarm. With no outer deadline, reuse testgate.HangCeiling; an expired outer deadline fails immediately. Do not add a guessed larger constant or alter production lock windows.

Route finished(), wantSettled(), and the #494 real-merge fixture's competing context bound through this one policy (or remove the redundant bound in favor of explicit cancellation/cleanup). Preserve readiness/ordering assertions and bounded teardown. On timeout name the measured candidate: runDurableMerge final repository-lock cleanup, archive confirmation and checkout revalidation; dump goroutines without acquiring fixture/repository locks. Do not claim that this window is necessarily the blocker.

Fold #500's malformed-car fixture into this shared fix. Add a narrow pkg/golint mistake-class guard against fixed completion waits in this fixture family; intentional deadline semantics elsewhere remain allowed. No production state, persisted format or local-gate configuration changes. ARCH-DRY/PURPOSE: one policy for all consumers; ARCH-PURE: deterministic budget arithmetic with thin clock/diagnostic glue; ARCH-CONSTRAINTS/ORDER: an inner diagnostic deadline derived from the supervisor it must beat. Existing stateful Git/GitHub fixtures remain the conformance; no new external dependency or durable artifact family.

## Done when

- Every train fixture completion wait derives from the supervising budget, including the #498/#500 cases and sibling waits; no bare timeout bump.
- A stalled completion reports all goroutines and names the measured cleanup-lock candidate before the outer Go test alarm.
- Deterministic budget/diagnostic tests and a mistake-class lint regression fail without the correction; the affected real-merge controls and required full suite pass.
- Unregister both #498 and #500 only after verification, record #500 as folded, and land.

## Plan

- [ ] Test and implement the shared derived budget and timeout diagnostics; sweep completion consumers and add the narrow lint guard.
- [ ] Run deadline/diagnostic regressions, actual train cases and the recorded full suite; close with the binary-owned fresh review.
- [ ] Remove both flake registrations after proof, fold #500, land and report to the hub.

## Log

### 2026-10-20
- Dispatched after #494 landed as PR 347. #500 is folded by explicit hub direction (8345800b-3ef8-4886-9b48-cb04b823c141), reaffirmed at fleet resume. Claim/start-plan are on a fresh issue branch from main; no stacked development.
- #494's final integrated run reproduced #498 at 128.03s, then the exact-head targeted /train passed in 7.51s (all variants/package 28.053s). Both attempts remain in Git-common-dir/sdlc/test-stats; full attempt 5f041c1edb0d03cdaf60da7698eeb358 and rerun 66c0484ce3c3dce2cc269d1be9d88861. #500 records the same finished() ceiling in its malformed-car fixture (128.90s versus 6.11s targeted). Existing persistent logs are off the RAM disk.
- Prior #480 measurements found the runDurableMerge final cleanup transaction holds its lock for 2–5s on archive confirmation reads and checkout revalidation. Optimizing that window is a separate candidate, not this fix. parley.nvim#401's corrected pattern derives the inner deadline from the actual outer runner budget. Go's t.Deadline() is authoritative here; this repo has no separate current load multiplier. A constant bump leaves the same mismatch, and a new load estimator would duplicate the supervisor's authority.
- Quick plan posted to the hub, proceeding unless it objects. Production behavior and states stay unchanged; the only added failure output is a developer diagnostic on an already-failing test. The hub owns test-economy.

## Revisions

- 2026-10-20: Fold #500 per hub direction and clarify the original Done-when (“never ... a longer timeout”): derive the inner bound from the supervising run, rather than raising the fixed two-minute constant. Preserve the prior evidence and keep lock optimization outside scope.
