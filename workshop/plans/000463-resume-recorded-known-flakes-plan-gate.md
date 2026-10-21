---
gate: plan-quality
issue: 463
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 81d8d6f5-858b-49d0-a0d0-f2da391bfaf4
            repository: github.com/xianxu/ariadne
            issue_id: "000463"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: 32a1e4184c4b5b22a435a81961d5a7a659af7aa7
            slot: ariadne:3
            requested_at: "2026-10-21T17:37:37.610895Z"
          phase: finished
          pid: 31615
          started_at: "2026-10-21T17:37:37.82379Z"
          finished_at: "2026-10-21T17:39:00.998136Z"
          session_id: 01a15f8a-3fbc-7be2-9eac-270c02542bb1
          outcome: returned
      "n": 1
      timestamp: "2026-10-21T10:39:01-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Define interrupted bundle retirement so pruning cannot restore resume eligibility
          detail: 'ARCH-ORDER/ARCH-FUNERAL: M1 promises bundle pruning under the parent lock, but does not specify the state after partial deletion or process death. Removing the link before its parent leaves an apparently unresumed failure; define a retirement protocol whose intermediate states cannot authorize execution or success, and test interrupted retirement through the production pruner with deterministic failure barriers.'
          family: durable-retirement-monotonicity
          round: 1
        - id: PQ-2
          severity: Important
          title: Name risky functions and compress the prose test inventories into strategies
          detail: 'ARCH-PURE/ARCH-ORDER: The current M1/M2 checklist enumerates malformed inputs and execution scenarios, while ResumeProof and ResumeTransition are entity names rather than named functions to unit-test. Name the proposed decoding, proof resolution, transition and retirement functions, give each one an adversarial input or ordering strategy plus an independent mechanical guard, and move the individual case inventories into executable tests.'
          family: named-risk-function-test-contract
          round: 1
      blocked: true
---

# Gate ledger — ariadne#463 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T10:39:01-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `durable-retirement-monotonicity` Define interrupted bundle retirement so pruning cannot restore resume eligibility
  ARCH-ORDER/ARCH-FUNERAL: M1 promises bundle pruning under the parent lock, but does not specify the state after partial deletion or process death. Removing the link before its parent leaves an apparently unresumed failure; define a retirement protocol whose intermediate states cannot authorize execution or success, and test interrupted retirement through the production pruner with deterministic failure barriers.
- **PQ-2** [Important] `named-risk-function-test-contract` Name risky functions and compress the prose test inventories into strategies
  ARCH-PURE/ARCH-ORDER: The current M1/M2 checklist enumerates malformed inputs and execution scenarios, while ResumeProof and ResumeTransition are entity names rather than named functions to unit-test. Name the proposed decoding, proof resolution, transition and retirement functions, give each one an adversarial input or ordering strategy plus an independent mechanical guard, and move the individual case inventories into executable tests.

## Open findings

- **PQ-1** [Important] `durable-retirement-monotonicity` Define interrupted bundle retirement so pruning cannot restore resume eligibility
- **PQ-2** [Important] `named-risk-function-test-contract` Name risky functions and compress the prose test inventories into strategies
