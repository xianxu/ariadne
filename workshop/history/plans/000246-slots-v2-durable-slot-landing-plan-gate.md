---
gate: plan-quality
issue: 246
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-23T16:46:54-07:00"
      agent: codex
      blocked: true
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-09-23T16:50:12-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Minor
          title: Replace the undefined ARCH-STATE marker with ARCH-ORDER
          detail: The plan cites `ARCH-STATE` in the operating-envelope paragraph (plan:66), but the architecture registry defines this principle as ARCH-ORDER. Rename the marker so the state/event design is machine-recognizable and consistently reviewed.
          family: architecture-marker-integrity
          round: 2
      blocked: false
content_hash: 6bcf4bc9f9f488d3ff0d6071cc6823e0aff334206febd3fbb4a35c0bb99fa862
---

# Gate ledger — ariadne#246 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-23T16:46:54-07:00 (codex) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-09-23T16:50:12-07:00 (codex) — passed

### Raised

- **PQ-1** [Minor] `architecture-marker-integrity` Replace the undefined ARCH-STATE marker with ARCH-ORDER
  The plan cites `ARCH-STATE` in the operating-envelope paragraph (plan:66), but the architecture registry defines this principle as ARCH-ORDER. Rename the marker so the state/event design is machine-recognizable and consistently reviewed.

## Open findings

- **PQ-1** [Minor] `architecture-marker-integrity` Replace the undefined ARCH-STATE marker with ARCH-ORDER
