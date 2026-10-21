---
id: 000498
status: codecomplete
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'a4a74cd5a662efdedb9412723c8b9dce42df042c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T14:48:54-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flake_tests:
flow: {kind: quick, provenance: inferred, spec: "8876058c", done: "15e3cf4e"}
actual_hours: 1.42
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

- [x] Test and implement the shared derived budget and timeout diagnostics; sweep completion consumers and add the narrow lint guard.
- [x] Run deadline/diagnostic regressions, actual train cases and the recorded full suite; close with the binary-owned fresh review.
- [ ] Remove both flake registrations after proof, fold #500, land and report to the hub.

## Log

### 2026-10-20
- 2026-10-20: closed — Resolved concurrent #504 catalog/list insertions by retaining all entries and both matcher implementations; pkg/golint passed1.117s; fresh integrated full suite passed with zero failures/reruns. Since tested implementation only issue bookkeeping changed. Separate production readiness-clock bug filed516 with controlled-clock red/green, no production change to it here. --no-project: hub owns test-economy. --no-plan-check: remaining checklist item is landing/report after this close; verification, removal and fold are done. (fixed-then-ship; full suite: Resolved-integration suite on 73f469513b02c34d39cadf9dd613734b89ba2c51 passed all8 shards,1877 cmd/sdlc tests and75 other packages, zero failures/reruns,764.2s. Attempt a1a8983f0805bb5c183a550dedd34044, sdlc-local/suite success, log SHA256 prefix1afa234db846, persistent copy common test-stats/evidence/000498/integrated-suite.log.); review verdict: SHIP
- 2026-10-20: closed — BR-1 resolved: recorded suite on exact code e51983dee5a21f532fa8d1ae83749d55b59043f3 passed all 8 shards, 1877 cmd/sdlc tests and 75 other packages, zero failures/reruns, 878.6s. Attempt 706d1c6db69cb8835e2befae20926eb5; persistent suite.log SHA256 prefix bb46b4829d2f, sdlc-local/suite success. Required real controls and actual-helper diagnostics passed; reviewer independently repeated them. Since tested code only docs/review bookkeeping changed. Both flake registrations explicitly removed after proof and #500 triaged duplicate/folded; no prolonged lock hold claimed. --no-plan-check acknowledges remaining checklist wording is this close plus subsequent landing/report, with verification/removal already complete. --no-project because hub owns test-economy.; review verdict: SHIP
- Dispatched after #494 landed as PR 347. #500 is folded by explicit hub direction (8345800b-3ef8-4886-9b48-cb04b823c141), reaffirmed at fleet resume. Claim/start-plan are on a fresh issue branch from main; no stacked development.
- #494's final integrated run reproduced #498 at 128.03s, then the exact-head targeted /train passed in 7.51s (all variants/package 28.053s). Both attempts remain in Git-common-dir/sdlc/test-stats; full attempt 5f041c1edb0d03cdaf60da7698eeb358 and rerun 66c0484ce3c3dce2cc269d1be9d88861. #500 records the same finished() ceiling in its malformed-car fixture (128.90s versus 6.11s targeted). Existing persistent logs are off the RAM disk.
- Prior #480 measurements found the runDurableMerge final cleanup transaction holds its lock for 2–5s on archive confirmation reads and checkout revalidation. Optimizing that window is a separate candidate, not this fix. parley.nvim#401's corrected pattern derives the inner deadline from the actual outer runner budget. Go's t.Deadline() is authoritative here; this repo has no separate current load multiplier. A constant bump leaves the same mismatch, and a new load estimator would duplicate the supervisor's authority.
- Quick plan posted to the hub, proceeding unless it objects. Production behavior and states stay unchanged; the only added failure output is a developer diagnostic on an already-failing test. The hub owns test-economy.


- Hub approved the concrete plan (receipt 5f720053-8ab6-49f2-8b2c-68b962262d45). Guard retained: repeated dumps establishing final cleanup lock holds of tens of seconds require a separate product issue with evidence, rather than absorbing that slowdown into a larger fixture budget.
- Implemented one pure remaining-deadline policy and thin shared completion observer. finished() and wantSettled() retain error identity; the archive regression uses explicit cancellation and its existing cancel/release/join teardown. Diagnostics collect the full goroutine profile without fixture/repository locks and identify the measured cleanup window as a candidate. Added the incident-linked narrow lint for these completion helpers (including the new shared helper) and the competing archive context bound; intentional readiness/context-expiration tests remain allowed.
- TDD: missing policy initially failed compilation; the new lint test initially reported no such catalog rule. Against the old completion helpers, the real diagnostic subprocesses reached Go's outer eight-second alarm instead of emitting the named failure. After correction, helper/budget/error/diagnostic tests passed (8.409s), and the full pkg/golint package passed (0.548s). Initial red/green output is retained in the agent tool transcript; no nonexistent raw file is cited.
- A deliberately broad targeted selection, go test ./cmd/sdlc -run '^(TestLanding|TestQueueMerge)' -count=1 -timeout=10m, exhausted its package supervisor (600.253s). Near the end, TestLandingTrainReintegratesADocsOnlyMain and TestLandingCarMergedWithoutAWordFinishes emitted the new named diagnostic with derived inner budgets 7.310s and 1.124s; Go's alarm then fired during TestLandingTrainLaggingListingIsItsOwnMerge setup. Stacks include final cleanup checkout revalidation/loadTopology/forkExec, but snapshots do not establish a lock-hold duration or repeated tens-of-seconds holds. Preserve this failed run, not a passing-retry claim: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/evidence/000498/targeted-landing-queue.log. The host wall clock advanced roughly two hours during the run, so wall elapsed alone is not a product measurement.
- The focused required controls plus helper tests passed in a fresh supervising run (81.065s): archive conformance 19.13s; all-car train/null/head 17.56/15.36/12.02s; malformed car 8.38s; diagnostic subprocesses 4.04s each, named failures before their actual eight-second Go alarm. Command: go test ./cmd/sdlc -run '^TestLanding(TrainLandsEveryCarOnOneSuite|TrainHoldsAMalformedCarNamingTheFile|TrainWaiterAcceptsConductorArchive|Completion.*)$' -count=1 -timeout=10m -v. Persistent raw output: /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/evidence/000498/targeted-required-controls.log. This is targeted coverage; the required recorded sharded full suite remains next.

- Recorded suite queued on pushed e51983dee5a21f532fa8d1ae83749d55b59043f3 (attempt 706d1c6db69cb8835e2befae20926eb5). Per worker contract, close review overlapped the queue with explicit pending evidence and --no-plan-check; no checklist was falsely completed. Round 1 returned REWORK solely for BR-1, missing recorded full-suite proof. It found no implementation defect and independently passed pkg/golint plus the focused controls (71.636s). Close is not finalized; supply the suite result, resolve failures, then re-close and unregister only after proof. Hub owns the project (--no-project).

- Required recorded full suite PASS on the exact implementation commit e51983dee5a21f532fa8d1ae83749d55b59043f3: all eight shards, 1,877 cmd/sdlc tests and 75 other packages; zero failures or reruns, wall 878.6s. Queued 1,886s before execution. Attempt 706d1c6db69cb8835e2befae20926eb5, persistent raw log /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/e51983dee5a21f532fa8d1ae83749d55b59043f3/suite.log, SHA256 prefix bb46b4829d2f; sdlc-local/suite posted success. Only documentation/review bookkeeping has changed since that code commit. This supplies BR-1's missing exact-code full-suite proof; no test retry was used.
- After that pass, explicitly unregistered #498's /train and #500's malformed-car test via sdlc flake unregister. #500's unstarted duplicate card was triaged wontfix (folded into #498); its details record the common fix and proof on this same branch. No independent #500 code branch or review is implied. Final review/close and landing remain next; hub owns project records.

- Landing runner reached its turn and refused integration conflicts with #504 in the pkg/golint Catalog and atlas catalog list. Resolved by retaining all three entries (our completion rule and both session-move rules); no matcher behavior changed. Full pkg/golint passed (1.117s), staged diff whitespace check clean. Integrated-main parent is 03e116cb767b0dcf9c9f312e3c282ac2127b9fee. A fresh integrated suite and narrow conflict re-close are still required before retrying landing.
- Hub/ariadne:3 reported a newly observed archive-fixture failure on main baseline 94bf14d9262f4ced7db0ebdb43b44583d7d8e2d5 (attempt 6537ba8d2b38cce637a077e8f11be119). The entire suite took 727.5s, but the actual fixture failed in 10.86s with a train-alone/no-other-car-ready event, not a two-minute completion timeout. Investigating readiness ordering separately before changing this issue's approved completion-budget scope; raw baseline evidence is already persistent in the common test-stats store.

- New archive failure is a separate production readiness-clock ordering bug, filed as #516 with hub approval (4299f256-1c1f-4790-b39b-9b6d25545b20). Deterministic production-path callback refreshes a waiting vouch at t0+1ms during the conductor's own readiness check: early captured now yields alone/no-other-ready (1.24s); a TEMP detached-checkout resample after readiness read yields a landed real two-car train (3.37s). No correction or fixture masking is on #498. Evidence lives in common test-stats/evidence/000498/readiness-diagnosis (red.log, green.log, clock-resample.patch, readiness_diagnosis_test.go).
- The original archive fixture then passed unchanged on exact failed baseline 94bf14d9262f4ced7db0ebdb43b44583d7d8e2d5: 18.11s (package18.597s), Git HEAD before/after identical, detached checkout clean. baseline-original.log and baseline-result.json are in the same persistent evidence directory. Registered only that fixture on #516 with the observed exact-head fail-to-pass evidence. #498/#500 completion flakes remain retired. #516 is unclaimed until this issue lands, then assigned here by hub direction; no stacked implementation. #516's details already hold the clock reproduction and persistent evidence pointers; append this registry observation to its Log on pickup.
- Fresh resolved-integration suite queued on 73f469513b02c34d39cadf9dd613734b89ba2c51, attempt a1a8983f0805bb5c183a550dedd34044. Subsequent changes are only this bookkeeping. A passing integrated suite and narrow --fixed-then-ship conflict review remain required before retrying landing.

- Fresh integrated suite PASS on 73f469513b02c34d39cadf9dd613734b89ba2c51: all eight shards, 1,877 cmd/sdlc tests and 75 other packages, zero failures or reruns, wall764.2s. Attempt a1a8983f0805bb5c183a550dedd34044; log /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/73f469513b02c34d39cadf9dd613734b89ba2c51/suite.log, SHA256 prefix1afa234db846; sdlc-local/suite success. Waited1,749s for both machine slots. Raw copy retained as common test-stats/evidence/000498/integrated-suite.log; original branch proof and baseline failure are also copied into this evidence area. The only pending checklist action is landing/report; removal and fold are already complete. Narrow conflict re-close follows this exact-code proof.

## Revisions

- 2026-10-20: Fold #500 per hub direction and clarify the original Done-when (“never ... a longer timeout”): derive the inner bound from the supervising run, rather than raising the fixed two-minute constant. Preserve the prior evidence and keep lock optimization outside scope.
