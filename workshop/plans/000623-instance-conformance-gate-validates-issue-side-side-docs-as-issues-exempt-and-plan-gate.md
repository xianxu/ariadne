---
reviewer_runs:
    - version: 1
      intent:
        id: 669faf68-7cfc-4d56-8b48-a6684b351cbf
        repository: github.com/xianxu/ariadne
        issue_id: "000623"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: codex
        root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        reviewed_head: a3bf0bbeb1baeb56375edc7ed80d0f08075e3066
        slot: ariadne:1
        requested_at: "2026-10-25T19:20:15.608731Z"
      phase: finished
      pid: 7362
      started_at: "2026-10-25T19:20:15.812567Z"
      finished_at: "2026-10-25T19:20:40.638136Z"
      session_id: 01a17481-a65a-7063-87d1-e529a7643001
      outcome: returned
gate: plan-quality
issue: 623
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: e5da8a7f-dde3-41c1-8b17-a70de0426be1
            repository: github.com/xianxu/ariadne
            issue_id: "000623"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: ca03b9cf0e0590c8ac0be7203c47520a5b2fe8ad
            slot: ariadne:1
            requested_at: "2026-10-25T19:15:54.121929Z"
          phase: finished
          pid: 22609
          started_at: "2026-10-25T19:15:54.242588Z"
          finished_at: "2026-10-25T19:16:45.771259Z"
          session_id: 01a1747d-a887-7140-86ab-40aba4e46c51
          outcome: returned
      "n": 1
      timestamp: "2026-10-25T12:16:57-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: PQ-1
          severity: Minor
          title: Align the recovery table with the approved missing-type compatibility rule
          detail: workshop/plans/000623-side-folder-rule-plan.md:67 says missing type produces a conformance error, while line 33 and the approved revision accept Markdown without a declared type as attachments. Update the table to distinguish accepted untyped attachments from invalid declared types and missing parents on typed documents.
          family: acceptance-contract-consistency
          round: 1
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 5d5685f6-3349-4472-b850-457a58f44632
            repository: github.com/xianxu/ariadne
            issue_id: "000623"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: a3bf0bbeb1baeb56375edc7ed80d0f08075e3066
            slot: ariadne:1
            requested_at: "2026-10-25T19:19:52.19435Z"
          phase: finished
          pid: 96834
          started_at: "2026-10-25T19:19:52.289223Z"
          finished_at: "2026-10-25T19:20:12.206537Z"
          session_id: 01a17481-4a0c-7800-8a3e-beaf8c21a285
          outcome: returned
      "n": 2
      timestamp: "2026-10-25T12:20:12-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Recovery-table lines 67–68 now distinguish accepted untyped attachments from invalid typed documents, matching the approved compatibility rule.
          round: 2
      blocked: false
content_hash: 7cb893bd8859ae084f5f27fa619b6c3f38ec0781185e156c5fe4d391ba974915
---

# Gate ledger — ariadne#623 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-25T12:16:57-07:00 (codex) — passed

### Raised

- **PQ-1** [Minor] `acceptance-contract-consistency` (kind: defect) Align the recovery table with the approved missing-type compatibility rule
  workshop/plans/000623-side-folder-rule-plan.md:67 says missing type produces a conformance error, while line 33 and the approved revision accept Markdown without a declared type as attachments. Update the table to distinguish accepted untyped attachments from invalid declared types and missing parents on typed documents.

## Round 2 — 2026-10-25T12:20:12-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Recovery-table lines 67–68 now distinguish accepted untyped attachments from invalid typed documents, matching the approved compatibility rule.

## Open findings

(none — every finding has been disposed)
