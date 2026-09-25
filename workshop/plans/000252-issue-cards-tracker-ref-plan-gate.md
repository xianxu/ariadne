---
gate: plan-quality
issue: 252
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-25T14:36:12-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Specify cancellation propagation through the reused Git transaction API
          detail: 'ARCH-ORDER / ARCH-CONSTRAINTS: TrunkFile uses a background-context runner (cmd/sdlc/internal/gitx/trunkfile.go:60, :148), and UpdateMany accepts no context (updatemany.go:68). Explicitly plan command-context propagation through transaction reads, writes, publication and confirmation, bounded subprocess termination, and retention of uncertain outcomes; verify cancellation through the production command boundary.'
          family: cancellation-through-production-seam
          round: 1
        - id: PQ-2
          severity: Important
          title: Replace enumerated test cases with named functions and adversarial strategies
          detail: 'ARCH-PURE / ARCH-ORDER: Tasks 1–8 list individual cases and test-file targets instead of naming the production functions to unit-test. Compress those inventories into function names with one adversarial-input or event-sequence strategy and mechanical invariant per risky function, retaining integration acceptance requirements separately.'
          family: function-level-adversarial-test-contract
          round: 1
        - id: PQ-3
          severity: Minor
          title: State when external dependency conformance checks run
          detail: 'ARCH-MOCK: The plan names stateful doubles and implementation-time real-Git checks but no recurring conformance cadence. Identify the checks and their execution trigger for the Git and GitHub behavior relied upon by publication and landing recovery.'
          family: external-fake-conformance-cadence
          round: 1
      blocked: true
---

# Gate ledger — ariadne#252 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-25T14:36:12-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `cancellation-through-production-seam` Specify cancellation propagation through the reused Git transaction API
  ARCH-ORDER / ARCH-CONSTRAINTS: TrunkFile uses a background-context runner (cmd/sdlc/internal/gitx/trunkfile.go:60, :148), and UpdateMany accepts no context (updatemany.go:68). Explicitly plan command-context propagation through transaction reads, writes, publication and confirmation, bounded subprocess termination, and retention of uncertain outcomes; verify cancellation through the production command boundary.
- **PQ-2** [Important] `function-level-adversarial-test-contract` Replace enumerated test cases with named functions and adversarial strategies
  ARCH-PURE / ARCH-ORDER: Tasks 1–8 list individual cases and test-file targets instead of naming the production functions to unit-test. Compress those inventories into function names with one adversarial-input or event-sequence strategy and mechanical invariant per risky function, retaining integration acceptance requirements separately.
- **PQ-3** [Minor] `external-fake-conformance-cadence` State when external dependency conformance checks run
  ARCH-MOCK: The plan names stateful doubles and implementation-time real-Git checks but no recurring conformance cadence. Identify the checks and their execution trigger for the Git and GitHub behavior relied upon by publication and landing recovery.

## Open findings

- **PQ-1** [Important] `cancellation-through-production-seam` Specify cancellation propagation through the reused Git transaction API
- **PQ-2** [Important] `function-level-adversarial-test-contract` Replace enumerated test cases with named functions and adversarial strategies
- **PQ-3** [Minor] `external-fake-conformance-cadence` State when external dependency conformance checks run
