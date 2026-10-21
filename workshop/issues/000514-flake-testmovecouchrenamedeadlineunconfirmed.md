---
id: 000514
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e1a6e40ff4e3f47d0de9c6ef98fdac7d36c2efc5' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T18:24:57-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/tl-lead.TestMoveCouchRenameDeadlineUnconfirmed]
flow: {kind: quick, provenance: inferred, spec: "2e844cfb", done: "a0bbe98b"}
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
- [ ] Run targeted normal/race and lint checks, recorded full suite, close review and landing.

## Log

### 2026-10-20

- Claimed and started on fresh main after #504 landed. The old test starts a 300 ms deadline before helper launch and assumes its WriteFile executes first; helper's subsequent 700 ms sleep adds a second scheduling assumption. Existing test-shrunk-deadline lint only checks assignments to capitalized timeout names and misses lowercase struct fields. Plan posted to TL under quick-flow self-approval.

- ariadne:3 independently confirmed this same family on unchanged #501 integration head 0ae08a92fc7f802677827d0c10d86036ab3503d0, failed attempt 3bb7f54f1a2d8e703a8d0cb4ceb62fad: rename effect absent after 300ms; known-refusal locked/stale and permission helper startup exceeded 1s. All eight SDLC shards passed. Its focused unchanged-head run of all three top-level tests passed every control in 0.527s. All three top-level tests are registered on #514; the saved failed log is .git/sdlc/test-stats/failed/3bb7f54f1a2d8e703a8d0cb4ceb62fad.log. Baseline remains queued; this is one common launch-budget family, not an unrelated flake (receipt 1cbf92ec-2f26-421f-88c8-2ff2e69a232d).
- Red evidence: new struct-literal lint cases fail the old guard; rename cancellation regression with the old runner returns an unreadable empty reply instead of an unconfirmed cancelled outcome. Parent-context propagation and observed-effect cancellation pass focused tests; ten -race TestMoveCouch runs pass (17.960s). Sibling startup deadlines removed. Lint sweep preserves two weave bound-only tests and an already-held in-process testgate lock with specific suppressions; the judge fake now cancels only once its second attempt begins instead of relying on 300ms.

- Final fixture sweep: controlled helper gate tests cancellation before and after the rename effect. Inherited-pipe child stays alive until explicit cleanup release, removing its fixed lifetime assumption. Cleanup context derives from the test deadline and is explicitly cancelled; t.Context itself cancels before cleanup and cannot gate cleanup readiness. Ten final race runs pass in 16.514s; the earlier 17.960s race pass preceded the additional pipe cleanup fixture. Judge deadline-on-cue, weave bound-only tests, testgate held-lock case and full lint checks pass. No production timeout value was raised.
