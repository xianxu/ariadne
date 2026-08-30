---
gate: boundary-review
issue: 205
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-08-29T18:19:56-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: The registry omits the Spec's explicit input-scale constraint
          detail: cmd/sdlc/internal/judge/architecture.md:100 reduces workload/input scale and growth to workload and growth, and cmd/sdlc/internal/judge/judge_test.go:138 pins that omission. Add explicit input scale and a regression assertion before regenerating derived outputs (ARCH-PURPOSE).
          family: operating-envelope-semantic-completeness
          round: 1
        - id: BR-2
          severity: Critical
          title: The Core concepts table names conceptual or unmodified entities as modified code
          detail: workshop/plans/000205-make-operating-constraints-explicit-plan.md:19-41 includes nongreppable aliases and claims modifications at production paths absent from the diff. Append a plan revision using actual symbols/files and statuses consistent with the committed range.
          family: core-concept-inventory-accuracy
          round: 1
        - id: BR-3
          severity: Important
          title: CLI tests permit marker-only delivery with the principle body missing
          detail: cmd/sdlc/archprinciples_test.go:12-23 and cmd/sdlc/startplan_test.go:21-42 stayed green in a scratch mutation after ArchitectureBlock was replaced by marker-only text. Assert full registry or complete entry delivery at both seams to satisfy Done-when (ARCH-PURPOSE).
          family: derived-consumer-semantic-coverage
          round: 1
      blocked: true
---

# Gate ledger — ariadne#205 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-29T18:19:56-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `operating-envelope-semantic-completeness` The registry omits the Spec's explicit input-scale constraint
  cmd/sdlc/internal/judge/architecture.md:100 reduces workload/input scale and growth to workload and growth, and cmd/sdlc/internal/judge/judge_test.go:138 pins that omission. Add explicit input scale and a regression assertion before regenerating derived outputs (ARCH-PURPOSE).
- **BR-2** [Critical] `core-concept-inventory-accuracy` The Core concepts table names conceptual or unmodified entities as modified code
  workshop/plans/000205-make-operating-constraints-explicit-plan.md:19-41 includes nongreppable aliases and claims modifications at production paths absent from the diff. Append a plan revision using actual symbols/files and statuses consistent with the committed range.
- **BR-3** [Important] `derived-consumer-semantic-coverage` CLI tests permit marker-only delivery with the principle body missing
  cmd/sdlc/archprinciples_test.go:12-23 and cmd/sdlc/startplan_test.go:21-42 stayed green in a scratch mutation after ArchitectureBlock was replaced by marker-only text. Assert full registry or complete entry delivery at both seams to satisfy Done-when (ARCH-PURPOSE).

## Open findings

- **BR-1** [Critical] `operating-envelope-semantic-completeness` The registry omits the Spec's explicit input-scale constraint
- **BR-2** [Critical] `core-concept-inventory-accuracy` The Core concepts table names conceptual or unmodified entities as modified code
- **BR-3** [Important] `derived-consumer-semantic-coverage` CLI tests permit marker-only delivery with the principle body missing
