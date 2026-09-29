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
    - "n": 3
      timestamp: "2026-09-28T19:18:32-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: writeArchive now emits weave (0755) and LICENSE (0644); TestArchiveCarriesBinaryAndLicense and release-weave.test.sh:60-61 fail without it; Log confirms the keg has LICENSE.
          round: 3
        - id: BR-2
          disposition: addressed
          note: README.md now ends with "## License / MIT — see LICENSE" and states that archives carry LICENSE (MIT).
          round: 3
        - id: BR-3
          disposition: addressed
          note: The template now puts license after version (packaging/homebrew/Formula/weave.rb:5-6), so future releases publish verbatim; the one-time tap swap is logged.
          round: 3
        - id: BR-4
          disposition: not-addressed
          note: 'Deferred to #265 in Log prose only; #265 on origin/main has no seed/merge-check retirement item. Record it in #265 so the deferral is durable.'
          round: 3
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

## Round 3 — 2026-09-28T19:18:32-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — writeArchive now emits weave (0755) and LICENSE (0644); TestArchiveCarriesBinaryAndLicense and release-weave.test.sh:60-61 fail without it; Log confirms the keg has LICENSE.
- BR-2 — addressed — README.md now ends with "## License / MIT — see LICENSE" and states that archives carry LICENSE (MIT).
- BR-3 — addressed — The template now puts license after version (packaging/homebrew/Formula/weave.rb:5-6), so future releases publish verbatim; the one-time tap swap is logged.
- BR-4 — not-addressed — Deferred to #265 in Log prose only; #265 on origin/main has no seed/merge-check retirement item. Record it in #265 so the deferral is durable.

## Open findings

- **BR-4** [Minor] `seed-once-residue-has-no-retirement` Consumers' seed-once merge-check.yml keep the dormant #250 tap fallback with no removal path
