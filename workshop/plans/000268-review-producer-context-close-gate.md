---
gate: boundary-review
issue: 268
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-28T21:02:55-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Important
          title: Exercise both sides of the legacy producer exception
          detail: 'construct/local/fix/SKILL.md:242–251 requires stopping on missing context while allowing confirmed uninterrupted legacy traffic. The recorded probes exercise neither legacy acceptance nor rejection. Cover every instance of this family: uninterrupted legacy, restored activation, retargeted activation, and unknown legacy status. Verify acceptance only for the first and clarify the exception''s precedence if the producer blocks it.'
          family: legacy-compatibility-coverage
          round: 1
      recipe: small-diff-review
      blocked: true
    - "n": 2
      timestamp: "2026-09-28T21:07:16-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'Verified /tmp/ariadne-268-legacy-probe.txt embeds the exact pinned skill and inspected /tmp/ariadne-268-legacy.txt: uninterrupted legacy activation is accepted; restored, retargeted, and unknown activations are refused with artifacts preserved. The response explicitly resolves exception precedence. The issue records these results and the verification-scope revision.'
          round: 2
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#268 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-28T21:02:55-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `legacy-compatibility-coverage` Exercise both sides of the legacy producer exception
  construct/local/fix/SKILL.md:242–251 requires stopping on missing context while allowing confirmed uninterrupted legacy traffic. The recorded probes exercise neither legacy acceptance nor rejection. Cover every instance of this family: uninterrupted legacy, restored activation, retargeted activation, and unknown legacy status. Verify acceptance only for the first and clarify the exception's precedence if the producer blocks it.

## Round 2 — 2026-09-28T21:07:16-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Verified /tmp/ariadne-268-legacy-probe.txt embeds the exact pinned skill and inspected /tmp/ariadne-268-legacy.txt: uninterrupted legacy activation is accepted; restored, retargeted, and unknown activations are refused with artifacts preserved. The response explicitly resolves exception precedence. The issue records these results and the verification-scope revision.

## Open findings

(none — every finding has been disposed)
