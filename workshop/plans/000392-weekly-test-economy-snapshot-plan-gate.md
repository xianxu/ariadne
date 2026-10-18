---
reviewer_runs:
    - version: 1
      intent:
        id: a4f9a73d-8133-45da-88f7-d53fe67ee3c0
        repository: github.com/xianxu/ariadne
        issue_id: "000392"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: codex
        root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        reviewed_head: 867e586678a7b36ea131e928c71050b47a329d2d
        slot: ariadne:1
        requested_at: "2026-10-18T04:47:39.319346Z"
      phase: finished
      pid: 3377
      started_at: "2026-10-18T04:47:39.547965Z"
      finished_at: "2026-10-18T04:48:04.420176Z"
      session_id: 01a14d56-3dca-75a2-948b-5e7e712adae8
      outcome: returned
gate: plan-quality
issue: 392
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 3244a1aa-fb5a-4c5c-b159-80a3a74f062c
            repository: github.com/xianxu/ariadne
            issue_id: "000392"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 3799b2de2a1bf4421e816d1efa0f74b8fdc1524e
            slot: ariadne:1
            requested_at: "2026-10-18T04:44:04.86613Z"
          phase: finished
          pid: 83029
          started_at: "2026-10-18T04:44:05.21667Z"
          finished_at: "2026-10-18T04:45:01.033154Z"
          session_id: 01a14d52-f902-7160-ae71-1d6bd5ab83ff
          outcome: returned
      "n": 1
      timestamp: "2026-10-17T21:45:01-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Resolve the weekly cohort versus attempt-time conflict.
          detail: Plan lines 17 and 21 select attempts by weekly completion time while also promising all attempts from enqueue through landing for issues landed that week. An issue enqueued before the week can therefore lose earlier attempts from its numerator; an attempt for an issue landing next week has no stated disposition. Define the headline cohort, which evidence may come from outside the week, and how physical weekly counts reconcile with attributed counts (ARCH-PURPOSE, ARCH-ORDER).
          family: metric-population-consistency
          round: 1
        - id: PQ-2
          severity: Important
          title: Specify how landing records become trustworthy join inputs.
          detail: WeeklyEvidenceReader promises validated events, but landingline.Events only JSON-decodes records and accepts missing or semantically inconsistent fields (cmd/sdlc/internal/landingline/log.go:162–182). Name the validation function and identity invariants linking enqueue, landed event, completion binding and reachable main merge; define malformed, duplicate and conflicting evidence as unknown before attribution. Existing enqueue events lack PR identity (cmd/sdlc/landingqueue.go:120), so state how sequence and branch bind them to the landed event (ARCH-SECURE).
          family: evidence-boundary-validation
          round: 1
        - id: PQ-3
          severity: Important
          title: Replace prose case inventories with named functions and adversarial strategies.
          detail: Tasks 1–3 enumerate test cases instead of naming the production functions to unit-test and giving one adversarial-input/mechanical-guard strategy per risky function. Compress those inventories into strategies for window validation, aggregation, attribution, evidence decoding and publication; explicitly cover unknown-flake classification, which aggregateSuiteAttempts currently omits at cmd/sdlc/teststats.go:148–149.
          family: risk-based-test-contract
          round: 1
        - id: PQ-4
          severity: Minor
          title: Make the fleet-default rollout and project linkage verifiable.
          detail: The plan describes fleet discovery and reporting a link to the TL, but does not name the propagation or consumer-verification step that establishes the promised default after the pilot. Add that step and require confirmation of the ops project link, rather than ending verification at the handoff (ARCH-PURPOSE).
          family: default-rollout-completeness
          round: 1
        - id: PQ-5
          severity: Minor
          title: Define reader limits and the external-fetch verification seam.
          detail: “Bounded Git queries” supplies neither a timeout nor record/output limits and their exceeded behavior. State the batch operating envelope, reuse existing limits where appropriate, and identify how fetch success, offline failure and cancellation are exercised through the production seam plus a conformance check (ARCH-CONSTRAINTS, ARCH-MOCK).
          family: explicit-operating-envelope
          round: 1
        - id: PQ-6
          severity: Minor
          title: Complete the snapshot and temporary-file lifecycle contract.
          detail: Permanent retention needs an expected per-week storage cost and where growth is measured. Also name who removes temporary files left by process death; testing interrupted cleanup alone does not specify its production owner or bound (ARCH-FUNERAL).
          family: durable-artifact-lifecycle
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 18c77da9-db5a-4b9e-9d99-366949e7bd42
            repository: github.com/xianxu/ariadne
            issue_id: "000392"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 867e586678a7b36ea131e928c71050b47a329d2d
            slot: ariadne:1
            requested_at: "2026-10-18T04:46:12.160397Z"
          phase: finished
          pid: 43552
          started_at: "2026-10-18T04:46:12.51685Z"
          finished_at: "2026-10-18T04:46:40.830313Z"
          session_id: 01a14d54-ead2-7b20-9c53-1b5e41ff0a87
          outcome: returned
      "n": 2
      timestamp: "2026-10-17T21:46:41-07:00"
      agent: codex
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Plan line 21 separates the landed-in-week cohort from physical weekly attempts, includes cross-week evidence, defines reconciliation buckets and makes incomplete cohort ratios null.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Plan line 23 names validateLandingEvidence and defines ticket, issue, branch, PR, completion and Git ancestry invariants, with malformed or conflicting evidence becoming unknown before attribution.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Plan lines 59–66 name production functions and adversarial strategies, explicitly including the unknown failed-test class omitted by existing aggregation.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Plan line 68 requires consumer verification from both pilot checkouts after propagation and hub confirmation of the exact ops project link.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: Plan line 41 specifies record, byte and Git timeout limits, unknown coverage on exceeded limits or unavailable sources, production-seam fetch and cancellation fixtures, and live conformance.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: 'Plan lines 19 and 41 name permanent retention and backup ownership, expected storage below 50 KiB per repo-week, growth measurement in #402, and later-run cleanup of temporary files older than one day.'
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: ce636034-3582-4c15-a2d6-7a78dc7336df
            repository: github.com/xianxu/ariadne
            issue_id: "000392"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            reviewed_head: 867e586678a7b36ea131e928c71050b47a329d2d
            slot: ariadne:1
            requested_at: "2026-10-18T04:47:07.428604Z"
          phase: finished
          pid: 79129
          started_at: "2026-10-18T04:47:07.833041Z"
          finished_at: "2026-10-18T04:47:36.99879Z"
          session_id: 01a14d55-c2d0-7f43-a126-be76a08cf111
          outcome: returned
      "n": 3
      timestamp: "2026-10-17T21:47:37-07:00"
      agent: codex
      blocked: false
content_hash: dad7620cccef06a6ad688e6383cee72ea1de609a06903446fe32c23913d18bbd
---

# Gate ledger — ariadne#392 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T21:45:01-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `metric-population-consistency` Resolve the weekly cohort versus attempt-time conflict.
  Plan lines 17 and 21 select attempts by weekly completion time while also promising all attempts from enqueue through landing for issues landed that week. An issue enqueued before the week can therefore lose earlier attempts from its numerator; an attempt for an issue landing next week has no stated disposition. Define the headline cohort, which evidence may come from outside the week, and how physical weekly counts reconcile with attributed counts (ARCH-PURPOSE, ARCH-ORDER).
- **PQ-2** [Important] `evidence-boundary-validation` Specify how landing records become trustworthy join inputs.
  WeeklyEvidenceReader promises validated events, but landingline.Events only JSON-decodes records and accepts missing or semantically inconsistent fields (cmd/sdlc/internal/landingline/log.go:162–182). Name the validation function and identity invariants linking enqueue, landed event, completion binding and reachable main merge; define malformed, duplicate and conflicting evidence as unknown before attribution. Existing enqueue events lack PR identity (cmd/sdlc/landingqueue.go:120), so state how sequence and branch bind them to the landed event (ARCH-SECURE).
- **PQ-3** [Important] `risk-based-test-contract` Replace prose case inventories with named functions and adversarial strategies.
  Tasks 1–3 enumerate test cases instead of naming the production functions to unit-test and giving one adversarial-input/mechanical-guard strategy per risky function. Compress those inventories into strategies for window validation, aggregation, attribution, evidence decoding and publication; explicitly cover unknown-flake classification, which aggregateSuiteAttempts currently omits at cmd/sdlc/teststats.go:148–149.
- **PQ-4** [Minor] `default-rollout-completeness` Make the fleet-default rollout and project linkage verifiable.
  The plan describes fleet discovery and reporting a link to the TL, but does not name the propagation or consumer-verification step that establishes the promised default after the pilot. Add that step and require confirmation of the ops project link, rather than ending verification at the handoff (ARCH-PURPOSE).
- **PQ-5** [Minor] `explicit-operating-envelope` Define reader limits and the external-fetch verification seam.
  “Bounded Git queries” supplies neither a timeout nor record/output limits and their exceeded behavior. State the batch operating envelope, reuse existing limits where appropriate, and identify how fetch success, offline failure and cancellation are exercised through the production seam plus a conformance check (ARCH-CONSTRAINTS, ARCH-MOCK).
- **PQ-6** [Minor] `durable-artifact-lifecycle` Complete the snapshot and temporary-file lifecycle contract.
  Permanent retention needs an expected per-week storage cost and where growth is measured. Also name who removes temporary files left by process death; testing interrupted cleanup alone does not specify its production owner or bound (ARCH-FUNERAL).

## Round 2 — 2026-10-17T21:46:41-07:00 (codex) — passed

### Disposed

- PQ-1 — addressed — Plan line 21 separates the landed-in-week cohort from physical weekly attempts, includes cross-week evidence, defines reconciliation buckets and makes incomplete cohort ratios null.
- PQ-2 — addressed — Plan line 23 names validateLandingEvidence and defines ticket, issue, branch, PR, completion and Git ancestry invariants, with malformed or conflicting evidence becoming unknown before attribution.
- PQ-3 — addressed — Plan lines 59–66 name production functions and adversarial strategies, explicitly including the unknown failed-test class omitted by existing aggregation.
- PQ-4 — addressed — Plan line 68 requires consumer verification from both pilot checkouts after propagation and hub confirmation of the exact ops project link.
- PQ-5 — addressed — Plan line 41 specifies record, byte and Git timeout limits, unknown coverage on exceeded limits or unavailable sources, production-seam fetch and cancellation fixtures, and live conformance.
- PQ-6 — addressed — Plan lines 19 and 41 name permanent retention and backup ownership, expected storage below 50 KiB per repo-week, growth measurement in #402, and later-run cleanup of temporary files older than one day.

## Round 3 — 2026-10-17T21:47:37-07:00 (codex) — passed

## Open findings

(none — every finding has been disposed)
