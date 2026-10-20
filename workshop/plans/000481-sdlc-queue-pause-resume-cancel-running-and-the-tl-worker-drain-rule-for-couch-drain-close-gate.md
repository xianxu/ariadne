---
gate: boundary-review
issue: 481
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 5b887303-5339-4b9b-83e7-6ca77bfa298f
            repository: github.com/xianxu/ariadne
            issue_id: "000481"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/c96aaf0e0e803e05
            reviewed_head: 7bedb19c29d5c52404bd2922a9d8df1e9e454102
            slot: ariadne:3
            requested_at: "2026-10-20T08:55:51.943938Z"
          phase: finished
          pid: 22460
          started_at: "2026-10-20T08:55:52.085778Z"
          finished_at: "2026-10-20T08:58:18.363307Z"
          session_id: 01a15886-335e-7d91-a2b3-e17641373b01
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7bedb19c29d5c52404bd2922a9d8df1e9e454102
        tracker_head: 97ccd9410a99b8cadb917699d8779fb664a986de
        branch_point: 8a44224ff0f867837283368f62c0344bcd322b25
        since: "2026-10-20T01:19:56-07:00"
        until: "2026-10-20T01:55:25-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T01:19:56-07:00"
              end: "2026-10-20T01:55:25-07:00"
              began: "2026-10-20T01:19:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c4d5253780abccdf1ce30648f0b45486020f245af078260cd194a77f930957a6
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 70180593
              windows:
                - start: "2026-10-20T01:19:56-07:00"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a15865-a812-7ae1-b293-fd0d91059cdb
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 6e279d4bf2f8f0edc2ff5186fecf9a261f1445104597c5d55d57a401d8475e40
              started: "2026-10-20T08:20:19.645Z"
              spawned: true
              bytes: 1570280
              windows:
                - start: "2026-10-20T08:20:19.645Z"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a1586d-3992-7242-ab2c-4427a72c88aa
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c5a8d5f93adc68a80491a25a1d6b595b3594c243959b777603d21fd6c28ba5ec
              started: "2026-10-20T08:28:36.266Z"
              spawned: false
              bytes: 710285
              windows:
                - start: "2026-10-20T08:28:36.266Z"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a15871-6263-7d92-8c75-64419f58cf1d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 404c0af4ca27bfe70cd74e03ba6f942cc7bbf8678821ab9511639517491ced4a
              started: "2026-10-20T08:33:08.701Z"
              spawned: false
              bytes: 515584
              windows:
                - start: "2026-10-20T08:33:08.701Z"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a15872-cb08-7721-aad1-06cfa3bdf4f5
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: ca135eb5ab2f41048986c73ba8d8265f86e5859e72499657056a1a6fe19621e4
              started: "2026-10-20T08:34:40.935Z"
              spawned: false
              bytes: 469904
              windows:
                - start: "2026-10-20T08:34:40.935Z"
                  end: "2026-10-20T01:55:25-07:00"
            - harness: codex
              id: 01a15873-4a93-7760-99c7-2c400abf6c0a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: fbe94f17549253511c8ace0b7a7bdba4c7b8816404600763a10222d53182f8eb
              started: "2026-10-20T08:35:13.533Z"
              spawned: false
              bytes: 424682
              windows:
                - start: "2026-10-20T08:35:13.533Z"
                  end: "2026-10-20T01:55:25-07:00"
      "n": 1
      timestamp: "2026-10-20T01:58:18-07:00"
      dispatched: "2026-10-20T01:55:51-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: Durable cancellation intent is not reconciled after partial publication or owner death
          detail: 'cmd/sdlc/internal/queuegate/store.go:244 writes cancel/<id>.json before rewriting the lease, but lease.go:156-168 checks only the lease phase. If the pause process dies between those writes, CancellationRequested returns false and Observe never consumes the persisted intent. Separately, Observe at store.go:333-343 converts a dead interrupting lease into unresolved evidence without retaining the promised interruption record or removing its cancellation intent; prior-boot reconciliation later removes only the unresolved record. This loses interruption diagnostics and leaves permanent cancellation files that can eventually exceed operationFiles'' reader limit. ARCH-ORDER/ARCH-FUNERAL: reconcile exact-ID intents across partial publication and stale-lease recovery, preserve interruption evidence alongside descendant uncertainty, and clean intents after their last reader. Add deterministic crash-window and repeated recovery/boot tests.'
          family: durable-cancellation-reconciliation
          round: 1
        - id: BR-2
          severity: Critical
          title: The Core concepts table names entities and transition behavior absent from the implementation
          detail: 'workshop/plans/000481-queue-drain-plan.md:31-35 lists PauseSet, OperationPhase and Interruption in model.go and says Transition returns requested interruption IDs with owner state protected from direct mutation. The implementation instead exposes Control, string phases and Operation-based interruption records; Transition returns only (Control, error). LegacyObserver at plan line 46 is also implemented as observeLegacyQueue rather than the named entity. Under the required Core concepts cross-check, these are plan/code contradictions. ARCH-PURPOSE/ARCH-ORDER: append a Revisions entry reconciling the table and structural-enforcement claims with the actual API, or implement the promised model.'
          family: core-concept-traceability
          round: 1
        - id: BR-3
          severity: Important
          title: Queue command deadlines stop applying once the control mutex is acquired
          detail: 'cmd/sdlc/internal/queuegate/store.go:67-72 checks context only while acquiring the mutex, then invokes a callback without deadline checks. requestInterruptions at lines 209-258 can perform durable intent writes and lease rewrites for 4096 operations while holding that mutex; Observe likewise scans and reconciles records without checking context inside its transaction. Consequently queuecmd.go:140''s three-second timeout does not enforce the plan''s under-four-second target or the peer''s five-second ceiling. ARCH-CONSTRAINTS: propagate cancellation through long transaction loops, retain recoverable partial progress, and measure representative upper-bound workloads.'
          family: transaction-deadline-enforcement
          round: 1
        - id: BR-4
          severity: Minor
          title: Strict JSON validation is duplicated across the new boundary readers
          detail: 'cmd/sdlc/queue_legacy.go:285-350 repeats the recursive duplicate-key, depth, UTF-8 and trailing-data validation in internal/queuegate/codec.go:17-85. ARCH-DRY: extract a shared strict JSON inspection helper while retaining each reader''s schema-specific decoding.'
          family: shared-json-boundary-validation
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#481 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T01:58:18-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `durable-cancellation-reconciliation` Durable cancellation intent is not reconciled after partial publication or owner death
  cmd/sdlc/internal/queuegate/store.go:244 writes cancel/<id>.json before rewriting the lease, but lease.go:156-168 checks only the lease phase. If the pause process dies between those writes, CancellationRequested returns false and Observe never consumes the persisted intent. Separately, Observe at store.go:333-343 converts a dead interrupting lease into unresolved evidence without retaining the promised interruption record or removing its cancellation intent; prior-boot reconciliation later removes only the unresolved record. This loses interruption diagnostics and leaves permanent cancellation files that can eventually exceed operationFiles' reader limit. ARCH-ORDER/ARCH-FUNERAL: reconcile exact-ID intents across partial publication and stale-lease recovery, preserve interruption evidence alongside descendant uncertainty, and clean intents after their last reader. Add deterministic crash-window and repeated recovery/boot tests.
- **BR-2** [Critical] `core-concept-traceability` The Core concepts table names entities and transition behavior absent from the implementation
  workshop/plans/000481-queue-drain-plan.md:31-35 lists PauseSet, OperationPhase and Interruption in model.go and says Transition returns requested interruption IDs with owner state protected from direct mutation. The implementation instead exposes Control, string phases and Operation-based interruption records; Transition returns only (Control, error). LegacyObserver at plan line 46 is also implemented as observeLegacyQueue rather than the named entity. Under the required Core concepts cross-check, these are plan/code contradictions. ARCH-PURPOSE/ARCH-ORDER: append a Revisions entry reconciling the table and structural-enforcement claims with the actual API, or implement the promised model.
- **BR-3** [Important] `transaction-deadline-enforcement` Queue command deadlines stop applying once the control mutex is acquired
  cmd/sdlc/internal/queuegate/store.go:67-72 checks context only while acquiring the mutex, then invokes a callback without deadline checks. requestInterruptions at lines 209-258 can perform durable intent writes and lease rewrites for 4096 operations while holding that mutex; Observe likewise scans and reconciles records without checking context inside its transaction. Consequently queuecmd.go:140's three-second timeout does not enforce the plan's under-four-second target or the peer's five-second ceiling. ARCH-CONSTRAINTS: propagate cancellation through long transaction loops, retain recoverable partial progress, and measure representative upper-bound workloads.
- **BR-4** [Minor] `shared-json-boundary-validation` Strict JSON validation is duplicated across the new boundary readers
  cmd/sdlc/queue_legacy.go:285-350 repeats the recursive duplicate-key, depth, UTF-8 and trailing-data validation in internal/queuegate/codec.go:17-85. ARCH-DRY: extract a shared strict JSON inspection helper while retaining each reader's schema-specific decoding.

## Open findings

- **BR-1** [Critical] `durable-cancellation-reconciliation` Durable cancellation intent is not reconciled after partial publication or owner death
- **BR-2** [Critical] `core-concept-traceability` The Core concepts table names entities and transition behavior absent from the implementation
- **BR-3** [Important] `transaction-deadline-enforcement` Queue command deadlines stop applying once the control mutex is acquired
- **BR-4** [Minor] `shared-json-boundary-validation` Strict JSON validation is duplicated across the new boundary readers
