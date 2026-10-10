---
gate: boundary-review
issue: 254
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-09T21:18:44-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: New `<start ISO> → HEAD` window label has no test, and the fixtures still use SHA-style literals
          detail: actual.go:113 changed the label format, but actual_test.go:81 and close_adopt_test.go:38,55,80,119 still build "abc12345 → HEAD"-shaped fixtures, and no test asserts that computeActual emits the ISO start.
          family: label-format-unpinned
          round: 1
        - id: BR-2
          severity: Minor
          title: firstSHA in computeActual now serves only as the empty-window check
          detail: The label no longer uses it. A rename (or a comment) would stop readers assuming it still feeds the label.
          family: vestigial-variable
          round: 1
      recipe: small-diff-review
      blocked: false
---

# Gate ledger — ariadne#254 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-09T21:18:44-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `label-format-unpinned` New `<start ISO> → HEAD` window label has no test, and the fixtures still use SHA-style literals
  actual.go:113 changed the label format, but actual_test.go:81 and close_adopt_test.go:38,55,80,119 still build "abc12345 → HEAD"-shaped fixtures, and no test asserts that computeActual emits the ISO start.
- **BR-2** [Minor] `vestigial-variable` firstSHA in computeActual now serves only as the empty-window check
  The label no longer uses it. A rename (or a comment) would stop readers assuming it still feeds the label.

## Open findings

- **BR-1** [Minor] `label-format-unpinned` New `<start ISO> → HEAD` window label has no test, and the fixtures still use SHA-style literals
- **BR-2** [Minor] `vestigial-variable` firstSHA in computeActual now serves only as the empty-window check
