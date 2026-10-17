---
reviewer_runs:
    - version: 1
      intent:
        id: 5af826c5-5763-49d8-b2eb-6c507eec8ff0
        repository: github.com/xianxu/ariadne
        issue_id: "000420"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: codex
        root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        reviewed_head: 79cf8b68d53005d3ab567d5adbe5d2ec72cfa7d1
        slot: ariadne:1
        requested_at: "2026-10-17T08:02:49.772936Z"
      phase: finished
      pid: 56876
      started_at: "2026-10-17T08:02:49.903232Z"
      finished_at: "2026-10-17T08:03:12.960763Z"
      session_id: 01a148e2-90ae-7eb0-8e51-7f557d118b67
      outcome: returned
gate: plan-quality
issue: 420
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: c51e40bc-3906-4282-b414-e2f41fe1cad6
            repository: github.com/xianxu/ariadne
            issue_id: "000420"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 851e83e18e613572394888570b48ed7992255921
            slot: ariadne:1
            requested_at: "2026-10-17T07:57:40.850408Z"
          phase: finished
          pid: 89725
          started_at: "2026-10-17T07:57:40.952012Z"
          finished_at: "2026-10-17T07:58:38.1136Z"
          session_id: 01a148dd-d9dc-7b41-8ff5-37fb3d72ae46
          outcome: returned
      "n": 1
      timestamp: "2026-10-17T00:58:38-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Define work-span classification so queued execution cannot restore its own wait
          detail: The plan calls stage/build/edit/test command spans work, without excluding the enclosing suite invocation's own queue interval. A synchronous suite command spans acquisition and execution (cmd/sdlc/localcheck_run.go:96–125), so restoring its entire call/result span can undo queue subtraction. Specify how wrappers, yielded execution, later polls and incomplete calls establish work identity and endpoints, and ensure only independently evidenced work restores queued time (ARCH-PURPOSE, ARCH-ORDER).
          family: work-evidence-excludes-waiting
          round: 1
        - id: PQ-2
          severity: Important
          title: Separate historical replay from an explicitly requested queue-aware correction
          detail: 'Task 1 preserves old snapshot semantics, but Task 3 expects ordinary sdlc actual on completed #416 to incorporate current attempts. Completed measurements return retained inputs at cmd/sdlc/actualinputs.go:175–188, and correct-actual uses the ordinary measurement path at cmd/sdlc/actualcorrection.go:33–34. Define a generation-bound fresh correction measurement path that pins new queue evidence while preserving ordinary historical replay, or explicitly choose exclusion for #416; do not leave these incompatible behaviors to the implementer (ARCH-PURPOSE).'
          family: historical-replay-versus-remeasurement
          round: 1
        - id: PQ-3
          severity: Important
          title: Specify the producer and compatibility contract for exact queue endpoints
          detail: The plan promises new records with queue-start/acquisition timestamps, but its tasks omit the recorder and attempt-schema changes. SuiteAttempt currently has neither endpoint (cmd/sdlc/teststats.go:11–32), and decodeSuiteAttempt rejects unknown fields and unsupported versions (cmd/sdlc/teststats_store.go:86–88,124–125). Name the writer/schema changes, endpoint validation and duration consistency rules, and old/new-format compatibility policy so the promised evidence is actually produced and safely consumed (ARCH-SECURE).
          family: persisted-schema-producer-contract
          round: 1
        - id: PQ-4
          severity: Important
          title: Replace prose test inventories with named functions and adversarial strategies
          detail: Tasks 1 and 3 enumerate test cases, while the new risky selection, span-extraction and interval-subtraction functions are unnamed; QueueAttribution is presented as an entity rather than a callable contract. Compress these inventories into named unit-tested functions with one adversarial-input class and mechanical guard per function, including sequence testing for partial/yielded tool executions and invariant testing for interval unions. Retain the required end-to-end regression as the acceptance proof.
          family: risky-function-test-strategy
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 19e40789-fdeb-4a4d-be43-cc5ba9044e6e
            repository: github.com/xianxu/ariadne
            issue_id: "000420"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: a0eb988eab709022660627a1e09419ef9d029970
            slot: ariadne:1
            requested_at: "2026-10-17T07:59:55.936491Z"
          phase: finished
          pid: 26300
          started_at: "2026-10-17T07:59:56.077283Z"
          finished_at: "2026-10-17T08:00:26.677709Z"
          session_id: 01a148df-e9ee-7061-9d8d-4bd9f106aedf
          outcome: returned
      "n": 2
      timestamp: "2026-10-17T01:00:26-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Structural wait classification excludes suite/merge wrappers and polls; yielded calls require final results before independently evidenced work can restore time.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: A generation-bound actualCorrectionCandidate selects and pins fresh queue evidence while ordinary completed reads retain historical replay semantics.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: The plan names the endpoint recorder, optional v1 fields, validation and duration consistency rules, and legacy-record interpretation.
          round: 2
        - id: PQ-4
          disposition: not-addressed
          note: 'Named helpers and interval properties improve the plan, but Tasks 1–3 retain prose case inventories and omit explicit unit-test strategies for bindQueueSource and MeasurementQueueWait.Validate. Apply one rule across every risky function: name its unit-test target, adversarial-input class and mechanical guard, then remove redundant inventories while retaining the end-to-end acceptance regression (ARCH-PURE, ARCH-ORDER).'
          round: 2
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: ebd07cc4-2a14-4437-ac38-13d7ccf6c2d9
            repository: github.com/xianxu/ariadne
            issue_id: "000420"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: def091cbd0eeb9fae7857880e8695b2270d65b0c
            slot: ariadne:1
            requested_at: "2026-10-17T08:01:20.536575Z"
          phase: finished
          pid: 46288
          started_at: "2026-10-17T08:01:20.624658Z"
          finished_at: "2026-10-17T08:01:46.547167Z"
          session_id: 01a148e1-33c7-7f21-8981-aa85a762079f
          outcome: returned
      "n": 3
      timestamp: "2026-10-17T01:01:46-07:00"
      agent: codex
      dispose:
        - id: PQ-4
          disposition: addressed
          note: Named callable targets now have adversarial strategies and mechanical guards, including execution-sequence and interval-invariant tests; end-to-end acceptance remains.
          round: 3
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 3dfe7216-111b-4c55-821a-2cd7b3d78a0a
            repository: github.com/xianxu/ariadne
            issue_id: "000420"
            recipe: plan-quality
            round: 4
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 79cf8b68d53005d3ab567d5adbe5d2ec72cfa7d1
            slot: ariadne:1
            requested_at: "2026-10-17T08:02:23.78515Z"
          phase: finished
          pid: 54652
          started_at: "2026-10-17T08:02:23.953646Z"
          finished_at: "2026-10-17T08:02:48.551349Z"
          session_id: 01a148e2-2b49-7230-8d38-9b246eb10ef1
          outcome: returned
      "n": 4
      timestamp: "2026-10-17T01:02:48-07:00"
      agent: codex
      blocked: false
content_hash: 119150f53728f1188575c9a574891c720758b9b842882500b3bb89205d310388
---

# Gate ledger — ariadne#420 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T00:58:38-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `work-evidence-excludes-waiting` Define work-span classification so queued execution cannot restore its own wait
  The plan calls stage/build/edit/test command spans work, without excluding the enclosing suite invocation's own queue interval. A synchronous suite command spans acquisition and execution (cmd/sdlc/localcheck_run.go:96–125), so restoring its entire call/result span can undo queue subtraction. Specify how wrappers, yielded execution, later polls and incomplete calls establish work identity and endpoints, and ensure only independently evidenced work restores queued time (ARCH-PURPOSE, ARCH-ORDER).
- **PQ-2** [Important] `historical-replay-versus-remeasurement` Separate historical replay from an explicitly requested queue-aware correction
  Task 1 preserves old snapshot semantics, but Task 3 expects ordinary sdlc actual on completed #416 to incorporate current attempts. Completed measurements return retained inputs at cmd/sdlc/actualinputs.go:175–188, and correct-actual uses the ordinary measurement path at cmd/sdlc/actualcorrection.go:33–34. Define a generation-bound fresh correction measurement path that pins new queue evidence while preserving ordinary historical replay, or explicitly choose exclusion for #416; do not leave these incompatible behaviors to the implementer (ARCH-PURPOSE).
- **PQ-3** [Important] `persisted-schema-producer-contract` Specify the producer and compatibility contract for exact queue endpoints
  The plan promises new records with queue-start/acquisition timestamps, but its tasks omit the recorder and attempt-schema changes. SuiteAttempt currently has neither endpoint (cmd/sdlc/teststats.go:11–32), and decodeSuiteAttempt rejects unknown fields and unsupported versions (cmd/sdlc/teststats_store.go:86–88,124–125). Name the writer/schema changes, endpoint validation and duration consistency rules, and old/new-format compatibility policy so the promised evidence is actually produced and safely consumed (ARCH-SECURE).
- **PQ-4** [Important] `risky-function-test-strategy` Replace prose test inventories with named functions and adversarial strategies
  Tasks 1 and 3 enumerate test cases, while the new risky selection, span-extraction and interval-subtraction functions are unnamed; QueueAttribution is presented as an entity rather than a callable contract. Compress these inventories into named unit-tested functions with one adversarial-input class and mechanical guard per function, including sequence testing for partial/yielded tool executions and invariant testing for interval unions. Retain the required end-to-end regression as the acceptance proof.

## Round 2 — 2026-10-17T01:00:26-07:00 (codex) — BLOCKED

### Disposed

- PQ-1 — addressed — Structural wait classification excludes suite/merge wrappers and polls; yielded calls require final results before independently evidenced work can restore time.
- PQ-2 — addressed — A generation-bound actualCorrectionCandidate selects and pins fresh queue evidence while ordinary completed reads retain historical replay semantics.
- PQ-3 — addressed — The plan names the endpoint recorder, optional v1 fields, validation and duration consistency rules, and legacy-record interpretation.
- PQ-4 — not-addressed — Named helpers and interval properties improve the plan, but Tasks 1–3 retain prose case inventories and omit explicit unit-test strategies for bindQueueSource and MeasurementQueueWait.Validate. Apply one rule across every risky function: name its unit-test target, adversarial-input class and mechanical guard, then remove redundant inventories while retaining the end-to-end acceptance regression (ARCH-PURE, ARCH-ORDER).

## Round 3 — 2026-10-17T01:01:46-07:00 (codex) — passed

### Disposed

- PQ-4 — addressed — Named callable targets now have adversarial strategies and mechanical guards, including execution-sequence and interval-invariant tests; end-to-end acceptance remains.

## Round 4 — 2026-10-17T01:02:48-07:00 (codex) — passed

## Open findings

(none — every finding has been disposed)
