---
gate: plan-quality
issue: 527
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 67ee5d78-573c-4d67-8761-57f46e064775
            repository: github.com/xianxu/ariadne
            issue_id: "000527"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 7c936dbbd67c4d49f2f00927995597d2071ec494
            slot: ariadne:2
            requested_at: "2026-10-21T15:47:26.595819Z"
          phase: finished
          pid: 55718
          started_at: "2026-10-21T15:47:26.693931Z"
          finished_at: "2026-10-21T15:48:40.407012Z"
          session_id: 01a15f25-5e5d-7422-bcc1-86bd149b2924
          outcome: returned
      "n": 1
      timestamp: "2026-10-21T08:48:40-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: PQ-1
          severity: Important
          title: Replace the test-case inventory with named-function strategies
          detail: workshop/plans/000527-landing-split-attribution-plan.md:85–88 enumerates cases, but the plan does not name the new classification and impact-proof functions that will be unit-tested. Compress this inventory into one adversarial-input and mechanical-guard strategy per risky function, including Transition, the new pure classifier, and the impact-proof helper; retain the motivating production regression and verification commands. This makes the ARCH-PURE test boundary executable without duplicating the forthcoming test code.
          family: function-level-test-strategy
          round: 1
        - id: PQ-2
          severity: Important
          title: Specify and enforce the attribution assessment deadline
          detail: The plan at lines 73 and 105 promises bounded assessment but leaves the graph context to be confirmed. Existing graph preparation has its own timeout, while affected.LoadGoGraph subsequently runs go list using only its supplied context (cmd/sdlc/localcheck_affected.go:139–142; cmd/sdlc/internal/affected/golist.go:119–145), so preparation's deadline does not bound the whole assessment. For ARCH-CONSTRAINTS, state a concrete overall budget and its basis, name the context owner covering graph/diff/registry work, and define expiry as an unverified non-blaming stop with bounded cleanup; require a controllably stalled dependency test.
          family: bounded-attribution-io
          round: 1
      blocked: true
---

# Gate ledger — ariadne#527 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T08:48:40-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `function-level-test-strategy` Replace the test-case inventory with named-function strategies
  workshop/plans/000527-landing-split-attribution-plan.md:85–88 enumerates cases, but the plan does not name the new classification and impact-proof functions that will be unit-tested. Compress this inventory into one adversarial-input and mechanical-guard strategy per risky function, including Transition, the new pure classifier, and the impact-proof helper; retain the motivating production regression and verification commands. This makes the ARCH-PURE test boundary executable without duplicating the forthcoming test code.
- **PQ-2** [Important] `bounded-attribution-io` Specify and enforce the attribution assessment deadline
  The plan at lines 73 and 105 promises bounded assessment but leaves the graph context to be confirmed. Existing graph preparation has its own timeout, while affected.LoadGoGraph subsequently runs go list using only its supplied context (cmd/sdlc/localcheck_affected.go:139–142; cmd/sdlc/internal/affected/golist.go:119–145), so preparation's deadline does not bound the whole assessment. For ARCH-CONSTRAINTS, state a concrete overall budget and its basis, name the context owner covering graph/diff/registry work, and define expiry as an unverified non-blaming stop with bounded cleanup; require a controllably stalled dependency test.

## Open findings

- **PQ-1** [Important] `function-level-test-strategy` Replace the test-case inventory with named-function strategies
- **PQ-2** [Important] `bounded-attribution-io` Specify and enforce the attribution assessment deadline
