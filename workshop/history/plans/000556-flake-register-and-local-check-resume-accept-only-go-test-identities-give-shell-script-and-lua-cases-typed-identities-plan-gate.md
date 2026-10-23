---
reviewer_runs:
    - version: 1
      intent:
        id: 9dd3aae4-74ac-46cd-a20f-7e9c9a1fee99
        repository: github.com/xianxu/ariadne
        issue_id: "000556"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: codex
        root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        reviewed_head: d08194892307492a36e6a61e1805c38f6a884ad9
        slot: ariadne:3
        requested_at: "2026-10-22T20:36:37.795997Z"
      phase: finished
      pid: 75610
      started_at: "2026-10-22T20:36:38.259314Z"
      finished_at: "2026-10-22T20:37:18.63407Z"
      session_id: 01a16554-80c4-7241-8a61-38c69fe096d9
      outcome: returned
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
    - reviewer_runs:
        - version: 1
          intent:
            id: d59b6139-283f-4427-b37f-f4726e0c63f6
            repository: github.com/xianxu/ariadne
            issue_id: "000556"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: d08194892307492a36e6a61e1805c38f6a884ad9
            slot: ariadne:3
            requested_at: "2026-10-22T20:33:52.700487Z"
          phase: finished
          pid: 87585
          started_at: "2026-10-22T20:33:52.879803Z"
          finished_at: "2026-10-22T20:34:23.345927Z"
          session_id: 01a16551-f979-7fa3-ac0c-51d8dcae9374
          outcome: returned
      "n": 2
      timestamp: "2026-10-22T13:34:23-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Sole outer group ownership, inherited child membership, strict cleanup and synchronized production-path wrapper-death coverage resolve the ownership gap.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Function-named adversarial strategies and mechanical guards cover runner projection, evidence transitions, log attribution and saved-proof eligibility.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Producer/protocol and supported runtime updates require owner-run native conformance before landing; adopting peers repeat the native adapter fixture on every landing.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 9f9fc174-ce9d-4f48-9f70-06f01803d149
            repository: github.com/xianxu/ariadne
            issue_id: "000556"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
            reviewed_head: d08194892307492a36e6a61e1805c38f6a884ad9
            slot: ariadne:3
            requested_at: "2026-10-22T20:36:00.197891Z"
          phase: finished
          pid: 57697
          started_at: "2026-10-22T20:36:00.488627Z"
          finished_at: "2026-10-22T20:36:28.710443Z"
          session_id: 01a16553-ed37-7112-966f-bf4812180f38
          outcome: returned
      "n": 3
      timestamp: "2026-10-22T13:36:29-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Outer helper exclusively owns the process group; inherited children, strict cleanup and synchronized wrapper-death verification are explicit.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Risky functions have named adversarial strategies and mechanical guards.
          round: 3
        - id: PQ-3
          disposition: addressed
          note: Native conformance has explicit owners, change triggers and adoption landing cadence.
          round: 3
      blocked: false
content_hash: 454943b072b25c34f1a1d705e7d6c450dbfc510a96834e7b7ba1ba04cd742fca
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

## Round 2 — 2026-10-22T13:34:23-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Sole outer group ownership, inherited child membership, strict cleanup and synchronized production-path wrapper-death coverage resolve the ownership gap.
- PQ-2 — addressed — Function-named adversarial strategies and mechanical guards cover runner projection, evidence transitions, log attribution and saved-proof eligibility.
- PQ-3 — addressed — Producer/protocol and supported runtime updates require owner-run native conformance before landing; adopting peers repeat the native adapter fixture on every landing.

## Round 3 — 2026-10-22T13:36:29-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Outer helper exclusively owns the process group; inherited children, strict cleanup and synchronized wrapper-death verification are explicit.
- PQ-2 — addressed — Risky functions have named adversarial strategies and mechanical guards.
- PQ-3 — addressed — Native conformance has explicit owners, change triggers and adoption landing cadence.

## Open findings

(none — every finding has been disposed)
