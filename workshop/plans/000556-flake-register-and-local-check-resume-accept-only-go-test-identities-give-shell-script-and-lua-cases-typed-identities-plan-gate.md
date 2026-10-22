---
gate: plan-quality
issue: 556
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: d17cd1a7-b5e1-482b-84b8-06de09a96bde
            repository: github.com/xianxu/ariadne
            issue_id: "000556"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: a3f81c32ae131786fb4aedc9eb3c080dc0a6330f
            slot: ariadne:3
            requested_at: "2026-10-22T20:28:14.439856Z"
          phase: finished
          pid: 95190
          started_at: "2026-10-22T20:28:14.577793Z"
          finished_at: "2026-10-22T20:30:06.256108Z"
          session_id: 01a1654c-cfd0-7572-af52-de00856fd1e6
          outcome: returned
      "n": 1
      timestamp: "2026-10-22T13:30:06-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Resolve subprocess ownership across wrapper death before reusing the managed helper
          detail: 'ARCH-ORDER: The plan promises outer cleanup after wrapper death, but runManagedCommandBatch creates a new process group for each command; nested reuse separates the unit child from the outer owner''s group. Specify group membership, cancellation and observation ownership, and how the outer owner retains authority if the wrapper dies; require a synchronized production-path death test proving descendants stop before resources retire.'
          family: nested-process-ownership
          round: 1
        - id: PQ-2
          severity: Important
          title: Replace test-case inventories with function-named adversarial strategies
          detail: 'ARCH-PURE: Task lines 133, 140 and 148 enumerate test cases, while several pure-entity rows name types or responsibilities rather than the functions to test. Compress these into one strategy line per risky function naming its adversarial input class and mechanical guard, including runner parsing/projection, evidence transitions, typed log projection and saved-proof eligibility; leave concrete case inventories in executable tests.'
          family: function-level-test-strategy
          round: 1
        - id: PQ-3
          severity: Minor
          title: Name when native runner conformance is repeated
          detail: 'ARCH-MOCK: The plan requires real shell and Neovim/Plenary conformance during implementation but does not identify an ongoing cadence or rerun trigger. Name the owner and repeat policy so changes in the external runner are checked against the persisted fake model.'
          family: external-conformance-cadence
          round: 1
      blocked: true
---

# Gate ledger — ariadne#556 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T13:30:06-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `nested-process-ownership` Resolve subprocess ownership across wrapper death before reusing the managed helper
  ARCH-ORDER: The plan promises outer cleanup after wrapper death, but runManagedCommandBatch creates a new process group for each command; nested reuse separates the unit child from the outer owner's group. Specify group membership, cancellation and observation ownership, and how the outer owner retains authority if the wrapper dies; require a synchronized production-path death test proving descendants stop before resources retire.
- **PQ-2** [Important] `function-level-test-strategy` Replace test-case inventories with function-named adversarial strategies
  ARCH-PURE: Task lines 133, 140 and 148 enumerate test cases, while several pure-entity rows name types or responsibilities rather than the functions to test. Compress these into one strategy line per risky function naming its adversarial input class and mechanical guard, including runner parsing/projection, evidence transitions, typed log projection and saved-proof eligibility; leave concrete case inventories in executable tests.
- **PQ-3** [Minor] `external-conformance-cadence` Name when native runner conformance is repeated
  ARCH-MOCK: The plan requires real shell and Neovim/Plenary conformance during implementation but does not identify an ongoing cadence or rerun trigger. Name the owner and repeat policy so changes in the external runner are checked against the persisted fake model.

## Open findings

- **PQ-1** [Important] `nested-process-ownership` Resolve subprocess ownership across wrapper death before reusing the managed helper
- **PQ-2** [Important] `function-level-test-strategy` Replace test-case inventories with function-named adversarial strategies
- **PQ-3** [Minor] `external-conformance-cadence` Name when native runner conformance is repeated
