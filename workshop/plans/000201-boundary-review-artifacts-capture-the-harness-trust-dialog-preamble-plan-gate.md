---
gate: plan-quality
issue: 201
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-08-25T12:30:25-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Important
          title: Name the changed functions, seam contract, purity boundary, and risky-function test strategies
          detail: The checklist never names Run, Dispatch, classifyRunResult, dispatchBoundaryReview, or writeReviewSidecar, nor the structured stdout/stderr value that will replace []byte. State which logic is pure versus process IO and give one compressed adversarial-input/mechanical-guard strategy for each risky tested function; “split,” “persist,” and “add coverage” are not executable as written. ARCH-PURE.
          family: executable-entity-contract
          round: 1
        - id: PQ-2
          severity: Important
          title: Define the external-agent fake and live conformance coverage
          detail: “Fake-process regression coverage” does not identify the production seam, persisted/configurable fake state, adapter dependency surface, or live conformance cadence. Specify a fake executable behind Run/Dispatch that independently models stdout, stderr, launch failure, and non-zero exit behavior, drive Claude/Codex/Gemini plus synchronous/heartbeat paths through it, and name how real-CLI stream behavior is periodically checked. ARCH-MOCK.
          family: external-cli-fake-contract
          round: 1
      blocked: true
---

# Gate ledger — ariadne#201 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-25T12:30:25-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Important] `executable-entity-contract` Name the changed functions, seam contract, purity boundary, and risky-function test strategies
  The checklist never names Run, Dispatch, classifyRunResult, dispatchBoundaryReview, or writeReviewSidecar, nor the structured stdout/stderr value that will replace []byte. State which logic is pure versus process IO and give one compressed adversarial-input/mechanical-guard strategy for each risky tested function; “split,” “persist,” and “add coverage” are not executable as written. ARCH-PURE.
- **PQ-2** [Important] `external-cli-fake-contract` Define the external-agent fake and live conformance coverage
  “Fake-process regression coverage” does not identify the production seam, persisted/configurable fake state, adapter dependency surface, or live conformance cadence. Specify a fake executable behind Run/Dispatch that independently models stdout, stderr, launch failure, and non-zero exit behavior, drive Claude/Codex/Gemini plus synchronous/heartbeat paths through it, and name how real-CLI stream behavior is periodically checked. ARCH-MOCK.

## Open findings

- **PQ-1** [Important] `executable-entity-contract` Name the changed functions, seam contract, purity boundary, and risky-function test strategies
- **PQ-2** [Important] `external-cli-fake-contract` Define the external-agent fake and live conformance coverage
