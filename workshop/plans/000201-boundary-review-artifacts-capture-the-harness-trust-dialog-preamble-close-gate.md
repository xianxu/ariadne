---
gate: boundary-review
issue: 201
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-08-25T12:56:00-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: classifyRunResult is declared PURE despite performing IO
          detail: The plan lists classifyRunResult under Pure entities, but the implementation writes to an injected io.Writer and reads launch-error environment context. Promote it to INTEGRATION or split pure classification from the IO shell, and append a plan revision.
          family: core-concept-classification
          round: 1
        - id: BR-2
          severity: Important
          title: The live agent stream check does not verify the semantic channel
          detail: The test requests STREAM_OK but accepts any non-empty stdout. Assert trimmed stdout equals STREAM_OK and stderr does not contain it so Claude, Codex, or Gemini stream drift cannot pass unnoticed.
          family: external-contract-conformance
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-08-25T13:11:37-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The Core concepts table and ARCH-PURE narrative now classify classifyRunResult as an INTEGRATION boundary, with a dated revision recording the correction.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The live conformance test now requires trimmed stdout to equal STREAM_OK exactly; the plan documents why stderr may legitimately echo the prompt sentinel.
          round: 2
      blocked: false
---

# Gate ledger — ariadne#201 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-25T12:56:00-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `core-concept-classification` classifyRunResult is declared PURE despite performing IO
  The plan lists classifyRunResult under Pure entities, but the implementation writes to an injected io.Writer and reads launch-error environment context. Promote it to INTEGRATION or split pure classification from the IO shell, and append a plan revision.
- **BR-2** [Important] `external-contract-conformance` The live agent stream check does not verify the semantic channel
  The test requests STREAM_OK but accepts any non-empty stdout. Assert trimmed stdout equals STREAM_OK and stderr does not contain it so Claude, Codex, or Gemini stream drift cannot pass unnoticed.

## Round 2 — 2026-08-25T13:11:37-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — The Core concepts table and ARCH-PURE narrative now classify classifyRunResult as an INTEGRATION boundary, with a dated revision recording the correction.
- BR-2 — addressed — The live conformance test now requires trimmed stdout to equal STREAM_OK exactly; the plan documents why stderr may legitimately echo the prompt sentinel.

## Open findings

(none — every finding has been disposed)
