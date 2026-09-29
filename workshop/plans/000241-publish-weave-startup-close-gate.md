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
    - "n": 2
      timestamp: "2026-09-28T19:16:56-07:00"
      agent: claude
      findings:
        - id: BR-3
          severity: Minor
          title: Tap formula differs from the tag's build-of-record artifact by a hand stanza swap
          detail: The template at weave-v0.1.0 (4de31b2d) has license before version, so brew audit --strict needed a manual edit in the tap (1d3ecf3). The fix in 9cf6ef7d means future releases match README step 4 exactly, and the Log records this one exception.
          family: release-artifact-is-published-verbatim
          round: 2
        - id: BR-4
          severity: Minor
          title: 'Consumers'' seed-once merge-check.yml keep the dormant #250 tap fallback with no removal path'
          detail: This is harmless now that the tap resolves, but seed-once copies never receive the upstream deletion (ARCH-FUNERAL). Consider a one-off sweep when a consumer is next touched.
          family: seed-once-residue-has-no-retirement
          round: 2
      boundary: M2
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

## Round 2 — 2026-09-28T19:16:56-07:00 (claude) — passed

### Raised

- **BR-3** [Minor] `release-artifact-is-published-verbatim` Tap formula differs from the tag's build-of-record artifact by a hand stanza swap
  The template at weave-v0.1.0 (4de31b2d) has license before version, so brew audit --strict needed a manual edit in the tap (1d3ecf3). The fix in 9cf6ef7d means future releases match README step 4 exactly, and the Log records this one exception.
- **BR-4** [Minor] `seed-once-residue-has-no-retirement` Consumers' seed-once merge-check.yml keep the dormant #250 tap fallback with no removal path
  This is harmless now that the tap resolves, but seed-once copies never receive the upstream deletion (ARCH-FUNERAL). Consider a one-off sweep when a consumer is next touched.

## Open findings

- **BR-1** [Minor] `license-notice-travels-with-artifact` Release archives contain only the weave binary; the MIT notice is not bundled with distributed copies
- **BR-2** [Minor] `readme-reflects-public-surface` README does not state the project license
- **BR-3** [Minor] `release-artifact-is-published-verbatim` Tap formula differs from the tag's build-of-record artifact by a hand stanza swap
- **BR-4** [Minor] `seed-once-residue-has-no-retirement` Consumers' seed-once merge-check.yml keep the dormant #250 tap fallback with no removal path
