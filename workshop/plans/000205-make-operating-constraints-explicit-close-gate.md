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
    - "n": 2
      timestamp: "2026-08-29T18:34:59-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Removing explicit input scale from the registry makes TestArchitectureRegistry_ConstraintsContract fail.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: The revised inventory is accurate, but no regression fails without it as required by the claimed-fix contract.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Marker-only replacements at both CLI delivery calls make their seam tests fail.
          round: 2
      findings:
        - id: BR-4
          severity: Important
          title: The semantic contract accepts case-preserving negation of every required predicate
          detail: 'cmd/sdlc/internal/judge/judge_test.go:132-163 validates 18 phrases by substring, while lines 230-239 mutate only one phrase and accidentally lowercase it. This is the 2nd finding in family operating-envelope-semantic-completeness: define the affirmative-semantics rule and sweep all 18 predicates rather than fixing only operator confirmation (ARCH-PURPOSE).'
          family: operating-envelope-semantic-completeness
          round: 2
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

## Round 2 — 2026-08-29T18:34:59-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Removing explicit input scale from the registry makes TestArchitectureRegistry_ConstraintsContract fail.
- BR-2 — not-addressed — The revised inventory is accurate, but no regression fails without it as required by the claimed-fix contract.
- BR-3 — addressed — Marker-only replacements at both CLI delivery calls make their seam tests fail.

### Raised

- **BR-4** [Important] `operating-envelope-semantic-completeness` The semantic contract accepts case-preserving negation of every required predicate
  cmd/sdlc/internal/judge/judge_test.go:132-163 validates 18 phrases by substring, while lines 230-239 mutate only one phrase and accidentally lowercase it. This is the 2nd finding in family operating-envelope-semantic-completeness: define the affirmative-semantics rule and sweep all 18 predicates rather than fixing only operator confirmation (ARCH-PURPOSE).

## Open findings

- **BR-2** [Critical] `core-concept-inventory-accuracy` The Core concepts table names conceptual or unmodified entities as modified code
- **BR-4** [Important] `operating-envelope-semantic-completeness` The semantic contract accepts case-preserving negation of every required predicate
