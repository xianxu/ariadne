---
id: 000332
status: working
deps: []
github_issue:
created: 2026-10-11
updated: 2026-10-13
estimate_hours: 2.35
card_mirror: '89f3d390ea3abfa5a267585072506bda46e83a45' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T22:09:00-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# sdlc merge integrates a moved main itself, then re-runs the suite before landing

## Problem

A branch that passed its close review can sit unmerged for hours, because main keeps moving under it. ariadne#330 passed its close review at 00:53 on 10-11 and was still open at 13:30: each `sdlc merge --yes` (run by the operator, since the auto-mode classifier refuses it for workers until ariadne#318) found main had moved and stopped with conflicts or a stale base, and the worker re-merged main by hand, waited for the operator again, and lost the race again. The operator's question (10-11): shouldn't `sdlc merge` just do that?

## Spec

`sdlc merge` integrates origin/main itself when main moved since the branch's last integration:
- **Clean merge** (git resolves it): commit the merge on the issue branch, re-run the repo's full suite (the same command `close` names as its suite), and land if it passes. No new review: the reviewed diff is unchanged, and #269's freshness rule already defines what main content a review must have seen; reuse it to decide if a re-review is owed (main code that touches the branch's files since the review → re-review; otherwise not).
- **Conflicts:** stop, leave the branch as it was (merge aborted), and print the conflicting paths and the next action for the agent (resolve, then close's `--fixed-then-ship` path, which already demands `--full-suite`).
- **Suite failure after a clean merge:** stop and name the failing tests; with a baseline run of main's suite, distinguish failures the merge introduced from ones already on main (and say so).
- Retry loop: if main moves again during the suite run, integrate again (bounded, e.g. 3 times), then stop with a clear message.
- This is also where the suite becomes enforced at landing, not agent-typed evidence: in repos without an in-binary test gate, `close` records `--verified` text, and failures have reached pair main that way (pair#450).

## Done when

- A branch whose main moved without conflicts lands with one `sdlc merge --yes`: merge commit, suite run, land (test with a fake remote).
- A conflicting branch stops with the paths and the next action, and the branch is unchanged (test).
- A suite failure introduced by the merge is reported as such, and one already on main is labeled pre-existing (test).

## Plan

Durable plan: `workshop/plans/000332-sdlc-merge-integrates-a-moved-main-itself-then-re-runs-the-suite-before-landing-plan.md`.

- [x] M1 — the suite as a recorded, enforced check (`sdlc-local/suite`: declared or default `make test` with the Makefile's weight; required on every landing; a docs-only-distance ancestor record counts; lock weight fix; `sdlc local-check run --name suite` documented as the close suite)
- [x] M2 — integration loop in the durable path (merge main, conflict abort, publish-replay pre-check, lease push, bounded CI wait, re-integrate up to 3, resume an interrupted integration); legacy path refuses a moved main
- [x] M3 — failure labelling: a baseline run on main classifies introduced vs pre-existing

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module      design=0.05 impl=0.2
item: smaller-go-module      design=0.05 impl=0.2
item: atlas-docs             design=0.05 impl=0.06
item: milestone-review       design=0.0 impl=0.16
item: greenfield-go-module   design=0.25 impl=0.3
item: smaller-go-module      design=0.05 impl=0.16
item: smaller-go-module      design=0.05 impl=0.12
item: milestone-review       design=0.0 impl=0.16
item: smaller-go-module      design=0.05 impl=0.2
item: milestone-review       design=0.0 impl=0.16
design-buffer: 0.15
total: 2.35
```

- The items, in order:
  - M1: suite check (weight, default suite, effective declaration, main-side probe); ancestor-record rule with GitHub statuses; docs; M1 review.
  - M2: the integration state machine and its shell (merge, abort, replay pre-check, lease push, PR-head sync, resume); CI wait; legacy refusal plus recovery contract; M2 review.
  - M3: parser, classifier and baseline; the close review.
- Design is at ×0.2, because the plan and R1 settle the decisions. Impl is at 40% of the v2 table midpoints.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log



- 2026-10-14: closed M3 — BR-15: real test-lock helper and nested make reproduced baseline self-wait and repeated-check self-wait before correction (3.186s); preserving per-lock-scope owner made baseline, reintegration retry, and multiple named checks pass with local-run/landing regressions (8.271s). Full make -k test on f9e520df passed 12/12 cmd/sdlc shards (1,289 tests) and 51 other packages in 430.6s; gofmt and git diff --check clean. M3 classification and unrecorded baseline tests remain green; no gates bypassed.; review verdict: SHIP
- 2026-10-14: closed M2 — R8: CI Makefile probe and replay/tracker/archive/quick-flow Git commands now inherit cancellation and remaining CI budget; real slow-command regressions failed before correction and passed after. Broad landing/publish/archive/CI cmd/sdlc regressions 202.192s, full gitx 41.630s, final cancellation/integration/quick-flow cmd/sdlc 25.906s, full tracker 100.489s; gofmt and git diff --check clean. BR-14 previously accepted; no gates bypassed.; review verdict: SHIP
### 2026-10-11

### 2026-10-13
- 2026-10-13: closed M1 — R4: probe cleanup mutation fails both cases; four probe/runner descendant cases and read-only live CommitStatuses conformance against ariadne PR 229 pass in 8.298s. R3 broader local-check and CI-gate regressions passed in 28.346s; git diff --check clean.; review verdict: SHIP

- Mapped the landing path (ariadne:1). `runDurableMerge` has no branch-contains-main check: GitHub merges a stale head server-side, so the landed tree is never tested. The re-review rule after integrating main is already the publish gate's (#304: the reviewed patch is replayed onto the new base; a code conflict owes `--fixed-then-ship`), not #269 freshness (stale/fresh only). sdlc never runs the suite; #345's `runLocalCheck` is the engine (exact sha, detached worktree, machine lock, recorded status). A bug was found there: `acquireSuiteSlot` takes weight 1, so `make test` goes reentrant and ariadne's weight-2 sharded suite would hold one slot.
- Draft design sent to the hub:
  - integrate in `runDurableMerge`; on conflict abort and name the paths;
  - pre-check the publish collision on the merge commit before running the suite;
  - push with a lease;
  - the suite as a REQUIRED recorded local check on the exact head (enforced at every landing), with a `weight` field;
  - on failure, a baseline run on main labels failures introduced or pre-existing;
  - bounded CI wait;
  - re-integrate when main moves again, at most 3 times;
  - an interrupted push resumes via an `Integrated-Main:` trailer.
- Hub approved (2026-10-13): (a) declaration, else default `make test` (Makefile weight); (b) every landing needs `sdlc-local/suite` on its head, with `sdlc local-check run --name suite` as THE close suite; (c) the legacy path refuses a moved main; (d) a 20-minute CI wait. Keep the baseline on failure. Close's own commits change the head, so a record on an ancestor with a docs-only distance counts (it reuses the publish gate's docs/code line).
- (superseded) Open questions to the hub: the suite source (declaration and/or a `make test` default), enforcement on every landing, the legacy path, the CI wait bound.

- Codex resumed at `5a4e3cdf09b31f8591625586130295767131e7c9`; M1 round 2 returned REWORK. R3 records the complete correction of optional-suite enforcement, failed-parent lookup, descendant-held probe pipes and README guidance. New tests reproduced the bypass/error/lifetime failures before correction. Initial post-fix tests printed PASS but concurrent documentation edits tripped the hermeticity guard; verification must run with all writers quiet.
- R3 verification: the combined declaration, CI-gate, ancestor, make-probe, local-check command/runner, weight and status-parser regressions passed (`go test ./cmd/sdlc -run '^(TestDeclaredSuite|TestRunCIGate|TestLocalRecordsOnAncestors|TestEffectiveLocalChecks|TestMakeProbe|TestProbeMakeTestFacts|TestDefaultSuiteFactsComeFromMain|TestParseMakeTestFacts|TestParseLocalChecks|TestAcquireSuiteSlot|TestLocalCheck|TestSelectLocalChecks|TestRunLocalCheck|TestParseCommitStatuses|TestDecideCIGate)' -count=1`, 28.346s). The same subprocess lifetime fix covers the local-check runner; both descendant tests failed before it and passed in this run.
- M1 round 3 accepted R3 implementation but returned REWORK for a probe-cleanup test false negative and missing live status conformance. R4 tightens the exit bound to one second (removing group cleanup now fails both cases in 4.867s) and adds `TestLiveCommitStatusesConformance`, read-only against four independently pinned status records on ariadne PR #229. Live conformance plus all four probe/runner process cases passed in 8.298s with `SDLC_LIVE_GH=1 SDLC_LIVE_COMMIT_STATUSES_GH="$(command -v gh)" go test ./cmd/sdlc -run '^(TestLiveCommitStatusesConformance|TestMakeProbeBoundsDescendantPipes|TestRunLocalCheckTerminatesDescendants)$' -count=1 -v`.
- M2 implemented: bounded main integration, abort/rollback before push, pinned lease, PR-head sync, validated interruption recovery, pending-only CI wait with cached immutable facts and suite runtime excluded, final main re-fetch, and legacy pre/post-check refusal. New integration and legacy tests failed before implementation; CI waiting tests failed before its implementation. Broad `go test ./cmd/sdlc -run '^(TestIntegrationPRSync|TestNextIntegrationStep|TestLanding|TestWaitCIGate|TestRunMerge_|TestRecoveryContractsAreProven)' -count=1 -v` ran 321.920s: every selected test passed except the old instance-window fixture, whose static fake never registered the automatic push. Updated that fixture to follow pushes and assert the integrated-main window; its targeted rerun passed in 3.749s. R5 records delivered entities and the non-atomic server main-pin limit; README/help/atlas/recovery updated.
- M2 boundary attempt at `77a9668453d70378d18a8c13874810e973553d4c` failed operationally: gpt-6-astra returned HTTP 429/model capacity, then HTTP 503/model capacity on SDLC's automatic retry. No verdict or findings were produced. SDLC recorded `unknown`, did not finalize M2, and instructed consultation before rerunning. Reported to hub ops:0 (receipt `c7dfc9f5-9a2c-42af-a518-9d23eeaaaba3`), requesting retry authorization or an available reviewer/model. M3 remains unstarted; full suite and final close remain pending. No gates bypassed.

- M2 authorized retry completed on `99b0ee0916d97cfeb3182c570f6b88b9494cfb52`: REWORK, three Important findings (BR-12 inflight deadline, BR-13 integration Git lifetimes, BR-14 state ownership). Reported verdict to hub (receipt `6f5e354f-90d9-4c40-ab7d-75d5c121f670`). R6 records the class enumeration and enforced corrections. Four CI budget regressions failed before correction. Removing retry counting and post-exit cleanup failed state and descendant tests (cmd/sdlc 3.027s, gitx 4.841s). Restored focused tests passed (cmd/sdlc 6.139s, gitx 3.146s), including production cancelled-push recovery. Full gitx package passed 43.069s. Broader landing/legacy/CI/state/process/recovery regressions passed cmd/sdlc 339.763s and gitx 3.339s with all writers stopped. M2 review retry pending; M3 unstarted.

- M2 boundary round 7 on `5fb37ac030d00755217c474413e0d60a5f273993`: REWORK. BR-14 disposed; BR-12 and BR-13 remain. Reviewer reproduced a Makefile facts probe exceeding a 40 ms CI budget by 1.07s, and a cancelled publish replay continuing 1.06s through background Git calls; archive ownership reads also bypass the bounded runner. R7 revises the call-graph enumeration and requires context propagation with real slow-command regressions. Reported verdict to hub (receipt `0258db8d-ccb4-4349-a73d-75d5c121f670`).

- R8 implementation: CI facts probe inherits remaining budget; publish replay and archive/tracker ownership reads use contextual Git calls; quick-flow measurement uses contextual base and numstat. Real-process RED evidence: Makefile budget/cancel ~1.6s, publish/ownership/quick-numstat ignored cancellation; tracker workspace/fetched-tracker modes ~0.68/0.70s. Focused new SDLC regressions passed 1.818s; full gitx package passed 41.630s; broad landing/publish/archive/CI SDLC regressions passed 202.192s before the final tracker-reader correction. Reverify after that correction, then retry M2 review.

- R8 verification after final tracker reader correction: new real cancellation/integration/quick-flow regressions passed in cmd/sdlc 25.906s; full tracker package passed 100.489s. Earlier R8 broad landing/publish/archive/CI run passed cmd/sdlc 202.192s, and full gitx package passed 41.630s. Git diff --check clean. M2 review retry pending.

- M2 boundary round 8 SHIP at `7592b7bb325547dc8c4080d0f8206767d0493d84` with no open findings; reported to the hub. M3 RED: `TestLandingSuiteFailureLabels` failed both introduced/pre-existing and unparsed-main cases against the existing generic suite refusal. Parser tests failed to build without the new pure functions. M3 GREEN: focused landing, parser and local-check regressions passed in 7.675s; parser fuzz ran 1,919 executions; broader landing/CI/local-check regressions passed in 178.311s. The baseline uses the same check selected from main, keeps a local log, and cannot post a status/comment. Full suite and boundary close remain pending.

- First full `make -k test` after M3 ran 481.2s and failed four classes in cmd/sdlc: `TestVerbContextsReachTrackerReads` found M2's context-free compatibility wrappers; two stateful landing fixtures held a stale PR head after the integration push and timed out; the legacy archive-asset fixture constructed a moved-main state, now correctly refused before its assertion. The correction removes unused production wrappers, centralizes the bounded nil-context fallback, preserves recovery context values, makes the production suite declaration read contextual, lets the two fakes observe pushes, and integrates main manually in the legacy asset fixture before probing archive materialization. All four originally failing tests passed together in 40.605s; full gitx and tracker packages passed in 41.458s and 23.637s. The full suite rerun remains pending.

- Full `make -k test` rerun on `860d52a767efaa0624db028fdd563213717e36a3` passed: 12/12 cmd/sdlc shards, 1,287 cmd/sdlc tests, and 51 other packages; 392.7s wall, zero failures. M3 milestone review and issue close remain pending.

- M3 boundary round 9 REWORK on `6868d4d5693919e9bbd323a02d71f3f73b6ce0be`: one Critical BR-15, lock reentry loses the actual holder identity on the second suite run. The reviewer reproduced an isolated helper self-wait; reported to the hub. New real-lock baseline and repeated-check tests failed before correction (3.186s), then passed with the existing local-run and landing-label tests (8.271s). R11 records the class sweep: baseline, reintegration retries, and multiple named checks all preserve the owner. Broader verification and the boundary retry remain pending.

- Full `make -k test` on `f9e520df05c5cd6aa6f6725a8dfcd8ddadb5899b` passed after BR-15: 12/12 cmd/sdlc shards, 1,289 cmd/sdlc tests, 51 other packages, zero failures, 430.6s wall. M3 boundary retry pending.
