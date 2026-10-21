---
id: 000514
status: done
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '5e21c4a6de1776e7c4f81a731f2382b3b0290831' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T18:24:57-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchRenameDeadlineUnconfirmed, github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchKnownRefusals, github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchPermissionFailsClearly]
flow: {kind: quick, provenance: inferred, spec: "2e844cfb", done: "a0bbe98b"}
actual_hours: 2.27
---

# flake: TestMoveCouchRenameDeadlineUnconfirmed

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchRenameDeadlineUnconfirmed: 2026-10-20 ariadne:2 integration of origin/main 94bf14d9262f4ced7db0ebdb43b44583d7d8e2d5, HEAD 4cf622a73b25b158202e54a200f94c8cc591f4fa, exact unchanged source/index tree c7acf7c0b00664949af14a68bd6d27b446cd8328. TL_LEAD_RESUME_CONFORMANCE=1 go test -race ./cmd/tl-lead -count=1 failed only TestMoveCouchRenameDeadlineUnconfirmed at status_move_timeout_test.go:88: rename never published, effect missing after 300ms. Then go test -race ./cmd/tl-lead -run ^TestMoveCouchRenameDeadlineUnconfirmed$ -count=10 -v passed all ten with no source changes. The test starts a real race-instrumented helper and assumes it reaches WriteFile before the 300ms subprocess deadline; startup under load can consume that budget. Fix should synchronize helper readiness/effect before exercising cancellation rather than assume process-start timing.

## Spec

The rename interruption fixture must publish its simulated rename before cancellation can fire. Reuse a parent context in the subprocess boundary (Background when absent); production retains the migration-only five-second timeout. The test derives its safety bound from the test deadline, observes the effect as readiness, then cancels. Keep the independent elapsed-time deadline assertion. Replace fixed helper stalls whose successful startup matters with explicit blocking/readiness, and sweep the sibling helper modes. Extend the existing shrunk-deadline lint to duration fields in test struct literals: the missed `timeout: 300 * time.Millisecond` syntax (ARCH-PURPOSE, ARCH-DRY). Creates no new durable state or external effects; helper processes are cancelled/reaped and temporary fixtures belong to t.TempDir.

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [x] Add a deterministic delayed-start regression and reproduce the premature cancellation; extend lint fixture cases.
- [x] Gate cancellation on helper readiness using a parent context; sweep sibling startup assumptions.
- [x] Run targeted normal/race and lint checks and launch the recorded full suite; retain the landing hold until green. Close owns its review; merge follows SHIP and suite pass.

## Log

### 2026-10-21
- 2026-10-21: closed — Integrated main and resolved the sole documentation conflict by retaining both lint catalog entries. Both prior failing landing-train tests and all train/null/head controls pass on the integrated head (35.012s); full tl-lead passes (15.700s), golint passes (0.665s). No production or test changes beyond main integration. Fresh recorded suite launched; landing requires its green result. (fixed-then-ship) (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

- Resumed after Couch restart: the normal-order landing runner stopped at a documentation conflict; no live runner remained. The recorded suite on 211ff252 finished all non-SDLC packages green but failed TestLandingTrainWaiterAcceptsConductorArchive and TestLandingTrainLandsEveryCarOnOneSuite/head (saved log 5c257f02f30dc9dd6bc85babbd9edaaa.log, sha256 prefix 011e1f6c7f1a). TL identified intervening main changes #511/#512/#521 and requested targeted verification after integration, with no flake registration absent same-head evidence.
- Integrated main e3a4c617f55f29aa489e4718aecddcd4a8679254 in dbf7090ccbb0e52ad05fe5f04ad66cabcc948d55. The sole conflict was atlas/workflow/sdlc-binary.md's catalog list: retained main's landing-completion rule and this issue's struct-literal deadline coverage. Both landing-train tests and all train/null/head controls PASS on the integrated head (35.012s); full tl-lead PASS (15.700s), golint PASS (0.665s). Per TL's classification this is base drift; no flake registered. Launching a fresh recorded suite and narrow conflict re-close; landing remains gated on green verification.

### 2026-10-20
- 2026-10-20: closed — Integrated full suite completed all eight SDLC shards green; only new #485 actors_bound_test collided with the expanded lint guard. Initial actor snapshot is now in-process and counted; real stalled optional recheck retains 500ms bound with a specific bound-only suppression. Focused actor/lint PASS (0.694s/0.702s); race actor/Couch fixture family count=3 PASS (7.792s). Fresh recorded suite launched on 211ff2523; re-entry into normal landing order after this review, with green final suite still required. (fixed-then-ship) (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP
- 2026-10-20: closed — Readiness-gated rename cancellation covers both sides of the effect; ten final race fixture runs PASS (16.514s), full tl-lead PASS (12.093s), focused race tl-lead/judge PASS (2.726s/1.189s), full lint PASS (0.469s), other affected package checks pass. Fixed 300ms/1s startup budgets removed; real elapsed-bound test retained. Recorded full suite launched on 39a3dbe40, pending; queued review overlap allowed by tl-worker, landing requires its pass. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

- Claimed and started on fresh main after #504 landed. The old test starts a 300 ms deadline before helper launch and assumes its WriteFile executes first; helper's subsequent 700 ms sleep adds a second scheduling assumption. Existing test-shrunk-deadline lint only checks assignments to capitalized timeout names and misses lowercase struct fields. Plan posted to TL under quick-flow self-approval.

- ariadne:3 independently confirmed this same family on unchanged #501 integration head 0ae08a92fc7f802677827d0c10d86036ab3503d0, failed attempt 3bb7f54f1a2d8e703a8d0cb4ceb62fad: rename effect absent after 300ms; known-refusal locked/stale and permission helper startup exceeded 1s. All eight SDLC shards passed. Its focused unchanged-head run of all three top-level tests passed every control in 0.527s. All three top-level tests are registered on #514; the saved failed log is .git/sdlc/test-stats/failed/3bb7f54f1a2d8e703a8d0cb4ceb62fad.log. Baseline remains queued; this is one common launch-budget family, not an unrelated flake (receipt 1cbf92ec-2f26-421f-88c8-2ff2e69a232d).
- Red evidence: new struct-literal lint cases fail the old guard; rename cancellation regression with the old runner returns an unreadable empty reply instead of an unconfirmed cancelled outcome. Parent-context propagation and observed-effect cancellation pass focused tests; ten -race TestMoveCouch runs pass (17.960s). Sibling startup deadlines removed. Lint sweep preserves two weave bound-only tests and an already-held in-process testgate lock with specific suppressions; the judge fake now cancels only once its second attempt begins instead of relying on 300ms.

- Final fixture sweep: controlled helper gate tests cancellation before and after the rename effect. Inherited-pipe child stays alive until explicit cleanup release, removing its fixed lifetime assumption. Cleanup context derives from the test deadline and is explicitly cancelled; t.Context itself cancels before cleanup and cannot gate cleanup readiness. Ten final race runs pass in 16.514s; the earlier 17.960s race pass preceded the additional pipe cleanup fixture. Judge deadline-on-cue, weave bound-only tests, testgate held-lock case and full lint checks pass. No production timeout value was raised.

- Final focused verification: full tl-lead package PASS (12.093s), focused race tl-lead PASS (2.726s) and judge PASS (1.189s), lint PASS (0.469s); final fixture-family ten-race run PASS (16.514s). PR #357 opened and the recorded full suite launched on 39a3dbe40. No full-suite pass claimed yet. Close review may overlap its queue under tl-worker; landing requires green suite and SHIP.

## Revisions

### 2026-10-20 — separate pre-close verification from landing

The final Plan row previously bundled suite, close review and landing into one pre-close checkbox. It now records focused checks plus launching the recorded suite; suite pass and SHIP remain mandatory before landing. No behavior or acceptance scope changes.

- Acceptance checkpoint: fresh close round 1 returned SHIP with no findings (evidence commit f6a608aaad03cd5d2f45ac4a8be3cd35e480f71d; measured actual 0.52h). Recorded full suite on 39a3dbe40e48a07fb48aa7610d359f8fe363e15d PASS: all eight shards, 1,874 cmd/sdlc tests and 75 other packages, zero failures, wall 740.6s; log sha256 prefix 8764bf0febdd. Entering the SDLC landing line; no manual integration or gate bypass.

- Landing integration on 4a9c0f1d finished all eight SDLC shards green but failed the catalog on newly landed #485 actors_bound_test.go:24; saved failed log 78be3153120592375aeddd88c2c9f76d.log, sha256 prefix 3fe9ee87ec38. The actor test combined a healthy initial subprocess with a 500ms stalled recheck budget. Replaced only the initial observation with an in-process stateful snapshot; the optional recheck still runs a real stalled subprocess and asserts its bound. It now legitimately uses the documented bound-only lint exception; no timeout value increased. Cancelled only this checkout's queued automatic main-baseline diagnosis after the full integration run had completed; merge exited normally on its failed suite. TL revoked the unused grant and requested normal-order re-entry. Focused actor/lint checks pass (0.694s/0.702s); narrow re-close precedes re-entry.
