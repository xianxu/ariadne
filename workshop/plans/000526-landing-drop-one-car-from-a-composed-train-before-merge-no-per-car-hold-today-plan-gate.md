---
gate: plan-quality
issue: 526
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: a39e03c9-0caa-4cb2-91f0-532d7a31a2bd
            repository: github.com/xianxu/ariadne
            issue_id: "000526"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 8fc7499edcb717fbf3e81045a4977104bbc0424f
            slot: ariadne:1
            requested_at: "2026-10-21T17:32:45.978824Z"
          phase: finished
          pid: 82352
          started_at: "2026-10-21T17:32:46.071189Z"
          finished_at: "2026-10-21T17:33:35.08641Z"
          session_id: 01a15f85-cb74-7033-af90-14126531a090
          outcome: returned
      "n": 1
      timestamp: "2026-10-21T10:33:35-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Specify a durable cleanup barrier across survivor admission paths.
          detail: 'ARCH-ORDER / ARCH-FUNERAL: the plan requires confirmed old-train cleanup before survivors land, but existing closePR and reclaimTrains only log failures, and the one-car path bypasses reclamation (cmd/sdlc/landingtrain.go:647, :715, :819). Name the failure-aware cleanup result, persisted old-train identity/pending obligation, confirmation evidence, and admission enforcement covering both serial and train successors after conductor death; control-record retirement must preserve that obligation until confirmation.'
          family: cleanup-confirmation-before-successor
          round: 1
        - id: PQ-2
          severity: Important
          title: Compress test inventories into named risky functions and strategies.
          detail: Plan lines 88–108 enumerate test cases, contrary to this gate's test-plan contract, while new pure decision and codec functions lack names. Name the functions to unit-test and give each risky function one strategy line identifying its adversarial input or event-sequence class and mechanical guard; retain integration acceptance invariants without the case inventories.
          family: function-level-test-strategy
          round: 1
      blocked: true
---

# Gate ledger — ariadne#526 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T10:33:35-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `cleanup-confirmation-before-successor` Specify a durable cleanup barrier across survivor admission paths.
  ARCH-ORDER / ARCH-FUNERAL: the plan requires confirmed old-train cleanup before survivors land, but existing closePR and reclaimTrains only log failures, and the one-car path bypasses reclamation (cmd/sdlc/landingtrain.go:647, :715, :819). Name the failure-aware cleanup result, persisted old-train identity/pending obligation, confirmation evidence, and admission enforcement covering both serial and train successors after conductor death; control-record retirement must preserve that obligation until confirmation.
- **PQ-2** [Important] `function-level-test-strategy` Compress test inventories into named risky functions and strategies.
  Plan lines 88–108 enumerate test cases, contrary to this gate's test-plan contract, while new pure decision and codec functions lack names. Name the functions to unit-test and give each risky function one strategy line identifying its adversarial input or event-sequence class and mechanical guard; retain integration acceptance invariants without the case inventories.

## Open findings

- **PQ-1** [Important] `cleanup-confirmation-before-successor` Specify a durable cleanup barrier across survivor admission paths.
- **PQ-2** [Important] `function-level-test-strategy` Compress test inventories into named risky functions and strategies.
