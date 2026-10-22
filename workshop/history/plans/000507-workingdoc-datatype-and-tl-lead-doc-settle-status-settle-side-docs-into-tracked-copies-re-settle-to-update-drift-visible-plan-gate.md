---
reviewer_runs:
    - version: 1
      intent:
        id: b30a7ce2-e5d0-4dcd-a31b-ff81ce58473f
        repository: github.com/xianxu/ariadne
        issue_id: "000507"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: codex
        root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        reviewed_head: 1345f8477936c38b6f1cc8ea6bdc3677044e178e
        slot: ariadne:4
        requested_at: "2026-10-21T19:31:33.251149Z"
      phase: finished
      pid: 13105
      started_at: "2026-10-21T19:31:33.346794Z"
      finished_at: "2026-10-21T19:32:03.722861Z"
      session_id: 01a15ff2-8c6c-75f2-954f-6e999ca17b20
      outcome: returned
gate: plan-quality
issue: 507
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: e1d337d8-58fd-4db1-8b90-d83a43408337
            repository: github.com/xianxu/ariadne
            issue_id: "000507"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: bba0e23b00875dff04a5297f5737aca9fdc90175
            slot: ariadne:4
            requested_at: "2026-10-21T19:26:27.956784Z"
          phase: finished
          pid: 41811
          started_at: "2026-10-21T19:26:28.079255Z"
          finished_at: "2026-10-21T19:27:46.779248Z"
          session_id: 01a15fed-e412-7912-9c62-dafe05658042
          outcome: returned
      "n": 1
      timestamp: "2026-10-21T12:27:46-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Bind publication to the exact validated projection.
          detail: 'ARCH-SECURE / ARCH-ORDER: destination preparation does not bind the subsequent publisher''s file read to the planned bytes. Define an immutable-payload or expected-content contract for both publication modes, preserve intervening target edits, and test target mutation between preparation and publisher capture; post-publication verification is too late to prevent publishing an unvalidated revision.'
          family: publication-payload-authority
          round: 1
        - id: PQ-2
          severity: Important
          title: Name risky functions and replace prose case inventories with test strategies.
          detail: The pure-entity table does not name the functions to unit-test, and Tasks 1–3 enumerate cases rather than supplying the required compact strategy. Name the metadata/pointer parser, projection, marker scanner, settlement transition, and publication decision functions, with one adversarial input or ordering strategy and independent mechanical guard per risky function; leave individual cases to executable tests.
          family: function-test-strategy-contract
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: e00ecf8e-dec6-4ad8-a748-a2253345fcfd
            repository: github.com/xianxu/ariadne
            issue_id: "000507"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 1345f8477936c38b6f1cc8ea6bdc3677044e178e
            slot: ariadne:4
            requested_at: "2026-10-21T19:29:18.685383Z"
          phase: finished
          pid: 76753
          started_at: "2026-10-21T19:29:18.83888Z"
          finished_at: "2026-10-21T19:29:46.696897Z"
          session_id: 01a15ff0-7f50-7f21-ad10-ff631fbe06ef
          outcome: returned
      "n": 2
      timestamp: "2026-10-21T12:29:46-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Both publication modes use validated immutable payloads, preserve intervening target edits, and require controlled mutation tests.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: The superseding function-contract table names risky functions and gives adversarial strategies with independent mechanical guards.
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: d966dda3-9f70-4874-9921-3fe48d6f8cad
            repository: github.com/xianxu/ariadne
            issue_id: "000507"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
            reviewed_head: 1345f8477936c38b6f1cc8ea6bdc3677044e178e
            slot: ariadne:4
            requested_at: "2026-10-21T19:31:05.596468Z"
          phase: finished
          pid: 7061
          started_at: "2026-10-21T19:31:05.698778Z"
          finished_at: "2026-10-21T19:31:31.859959Z"
          session_id: 01a15ff2-206e-7f00-bcd0-f8fd39e9fdd5
          outcome: returned
      "n": 3
      timestamp: "2026-10-21T12:31:32-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Immutable payload guards and private-index publication bind confirmed copies to validated bytes.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: The superseding function table supplies adversarial strategies and independent mechanical guards.
          round: 3
      blocked: false
content_hash: 94087485e2e5471986997f4d4ddc158bedffab36ecfe13837d8c2fd2bd625836
---

# Gate ledger — ariadne#507 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T12:27:46-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `publication-payload-authority` Bind publication to the exact validated projection.
  ARCH-SECURE / ARCH-ORDER: destination preparation does not bind the subsequent publisher's file read to the planned bytes. Define an immutable-payload or expected-content contract for both publication modes, preserve intervening target edits, and test target mutation between preparation and publisher capture; post-publication verification is too late to prevent publishing an unvalidated revision.
- **PQ-2** [Important] `function-test-strategy-contract` Name risky functions and replace prose case inventories with test strategies.
  The pure-entity table does not name the functions to unit-test, and Tasks 1–3 enumerate cases rather than supplying the required compact strategy. Name the metadata/pointer parser, projection, marker scanner, settlement transition, and publication decision functions, with one adversarial input or ordering strategy and independent mechanical guard per risky function; leave individual cases to executable tests.

## Round 2 — 2026-10-21T12:29:46-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Both publication modes use validated immutable payloads, preserve intervening target edits, and require controlled mutation tests.
- PQ-2 — addressed — The superseding function-contract table names risky functions and gives adversarial strategies with independent mechanical guards.

## Round 3 — 2026-10-21T12:31:32-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Immutable payload guards and private-index publication bind confirmed copies to validated bytes.
- PQ-2 — addressed — The superseding function table supplies adversarial strategies and independent mechanical guards.

## Open findings

(none — every finding has been disposed)
