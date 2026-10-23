---
gate: plan-quality
issue: 561
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: aa86bfff-1727-46f4-8bb6-f6948ac2ea7f
            repository: github.com/xianxu/ariadne
            issue_id: "000561"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 1d8ff92a2cb0d124d0f043829442e96e6509bfa1
            slot: ariadne:2
            requested_at: "2026-10-23T07:28:18.14663Z"
          phase: finished
          pid: 76754
          started_at: "2026-10-23T07:28:18.285645Z"
          finished_at: "2026-10-23T07:32:05.516417Z"
          session_id: 01a167a9-1cfd-7c42-b326-59702214ac99
          outcome: returned
      "n": 1
      timestamp: "2026-10-23T00:32:05-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Replace prose case inventories with function-named adversarial strategies
          detail: The M1/M2 checklists enumerate test cases rather than supplying one adversarial-input strategy and mechanical guard per risky production function. Compress those inventories and name the functions tested, including provenance selection, proof decoding, measurement transitions and aggregation; executable tests should own the individual cases.
          family: function-named-test-strategy
          round: 1
        - id: PQ-2
          severity: Important
          title: Define and enforce the measurement lifecycle in the pure core
          detail: 'ARCH-ORDER and ARCH-PURE: the plan names conductor ownership and visible outcomes but no pure measurement transition function or state/event model. Name that owner, define legal transitions for recording, execution outcomes, resolution and interrupted finalization, prevent direct authoritative mutation, and specify sequence-based invariant testing.'
          family: explicit-state-transition-ownership
          round: 1
        - id: PQ-3
          severity: Important
          title: Specify strict nested proof decoding and unknown-field preservation
          detail: 'ARCH-EVOLVE and ARCH-SECURE: preserving Readiness.Extra does not preserve future members inside the new known BoardingProof object. Name its strict preserving codec and rewrite/merge behavior, including nested unknown members and duplicate-key rejection; require an older proof-aware reader to rewrite a newer proof without losing its unknown fields.'
          family: nested-format-preservation
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 591a2531-a9c7-4c48-9590-903d683bd02f
            repository: github.com/xianxu/ariadne
            issue_id: "000561"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 1d8ff92a2cb0d124d0f043829442e96e6509bfa1
            slot: ariadne:2
            requested_at: "2026-10-23T07:34:45.282053Z"
          phase: finished
          pid: 50778
          started_at: "2026-10-23T07:34:45.383441Z"
          finished_at: "2026-10-23T07:35:28.243099Z"
          session_id: 01a167af-04b8-7e72-81d0-27b0a05cb910
          outcome: returned
      "n": 2
      timestamp: "2026-10-23T00:35:28-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Function-named strategies cover provenance, decoding, transitions and aggregation with adversarial inputs and mechanical guards.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: AdvanceMeasurement owns encapsulated state, legal transitions, interrupted finalization and sequence-tested invariants.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Nested strict preserving codecs, binding-aware re-vouch merging and older-reader rewrite tests are explicitly required.
          round: 2
      blocked: false
content_hash: 388ef474ec21fcc1bdb8b5f52babc32f60c50dcb0af52aa7ccc74ec0824b4f8d
---

# Gate ledger — ariadne#561 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T00:32:05-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `function-named-test-strategy` Replace prose case inventories with function-named adversarial strategies
  The M1/M2 checklists enumerate test cases rather than supplying one adversarial-input strategy and mechanical guard per risky production function. Compress those inventories and name the functions tested, including provenance selection, proof decoding, measurement transitions and aggregation; executable tests should own the individual cases.
- **PQ-2** [Important] `explicit-state-transition-ownership` Define and enforce the measurement lifecycle in the pure core
  ARCH-ORDER and ARCH-PURE: the plan names conductor ownership and visible outcomes but no pure measurement transition function or state/event model. Name that owner, define legal transitions for recording, execution outcomes, resolution and interrupted finalization, prevent direct authoritative mutation, and specify sequence-based invariant testing.
- **PQ-3** [Important] `nested-format-preservation` Specify strict nested proof decoding and unknown-field preservation
  ARCH-EVOLVE and ARCH-SECURE: preserving Readiness.Extra does not preserve future members inside the new known BoardingProof object. Name its strict preserving codec and rewrite/merge behavior, including nested unknown members and duplicate-key rejection; require an older proof-aware reader to rewrite a newer proof without losing its unknown fields.

## Round 2 — 2026-10-23T00:35:28-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Function-named strategies cover provenance, decoding, transitions and aggregation with adversarial inputs and mechanical guards.
- PQ-2 — addressed — AdvanceMeasurement owns encapsulated state, legal transitions, interrupted finalization and sequence-tested invariants.
- PQ-3 — addressed — Nested strict preserving codecs, binding-aware re-vouch merging and older-reader rewrite tests are explicitly required.

## Open findings

(none — every finding has been disposed)
