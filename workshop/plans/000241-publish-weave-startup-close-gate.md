---
gate: boundary-review
issue: 241
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-28T19:04:39-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Release archives contain only the weave binary; the MIT notice is not bundled with distributed copies
          detail: release-weave.test.sh:60 asserts a single tar entry. Consider shipping LICENSE in each archive (and doc.install in the formula) so redistributed binaries carry the notice.
          family: license-notice-travels-with-artifact
          round: 1
        - id: BR-2
          severity: Minor
          title: README does not state the project license
          detail: 'Fold a one-line License: MIT into the M2 README sweep alongside the install instructions.'
          family: readme-reflects-public-surface
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#241 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-28T19:04:39-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `license-notice-travels-with-artifact` Release archives contain only the weave binary; the MIT notice is not bundled with distributed copies
  release-weave.test.sh:60 asserts a single tar entry. Consider shipping LICENSE in each archive (and doc.install in the formula) so redistributed binaries carry the notice.
- **BR-2** [Minor] `readme-reflects-public-surface` README does not state the project license
  Fold a one-line License: MIT into the M2 README sweep alongside the install instructions.

## Open findings

- **BR-1** [Minor] `license-notice-travels-with-artifact` Release archives contain only the weave binary; the MIT notice is not bundled with distributed copies
- **BR-2** [Minor] `readme-reflects-public-surface` README does not state the project license
