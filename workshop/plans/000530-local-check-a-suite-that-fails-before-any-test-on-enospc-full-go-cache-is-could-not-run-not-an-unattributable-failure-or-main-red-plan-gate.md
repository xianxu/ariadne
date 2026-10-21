---
gate: plan-quality
issue: 530
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 2ec53706-c73d-48f8-a4f9-ca54cddc7f89
            repository: github.com/xianxu/ariadne
            issue_id: "000530"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
            reviewed_head: eb4e9e55e962d59963edc145cae761cfe30926ba
            slot: ariadne:5
            requested_at: "2026-10-21T21:29:28.36678Z"
          phase: finished
          pid: 8269
          started_at: "2026-10-21T21:29:28.574985Z"
          finished_at: "2026-10-21T21:30:33.011004Z"
          session_id: 01a1605e-8315-77a1-bf79-667826f2aa0e
          outcome: returned
      "n": 1
      timestamp: "2026-10-21T14:30:33-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: Name risky functions and replace test-case inventories with strategy contracts
          detail: The pure-entity table names concepts/files, while Tasks 1–2 enumerate individual cases without naming the core functions under unit test and each function's adversarial-input or ordering strategy plus mechanical guard (ARCH-PURE). Name the diagnosis, eligibility/transition, decoding/preservation and evidence-projection functions; give each one concise strategy line, and move the case inventories into executable tests.
          family: risky-function-test-contract
          round: 1
      blocked: true
---

# Gate ledger — ariadne#530 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T14:30:33-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `risky-function-test-contract` (kind: defect) Name risky functions and replace test-case inventories with strategy contracts
  The pure-entity table names concepts/files, while Tasks 1–2 enumerate individual cases without naming the core functions under unit test and each function's adversarial-input or ordering strategy plus mechanical guard (ARCH-PURE). Name the diagnosis, eligibility/transition, decoding/preservation and evidence-projection functions; give each one concise strategy line, and move the case inventories into executable tests.

## Open findings

- **PQ-1** [Important] `risky-function-test-contract` (kind: defect) Name risky functions and replace test-case inventories with strategy contracts
