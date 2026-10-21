---
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

## Open findings

- **PQ-1** [Important] `publication-payload-authority` Bind publication to the exact validated projection.
- **PQ-2** [Important] `function-test-strategy-contract` Name risky functions and replace prose case inventories with test strategies.
