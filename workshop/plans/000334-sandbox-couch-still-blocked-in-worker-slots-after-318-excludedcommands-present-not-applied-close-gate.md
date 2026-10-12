---
gate: boundary-review
issue: 334
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T23:53:22-07:00"
      dispatched: "2026-10-11T23:52:14-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: atlas/workflow/sdlc-binary.md:963 still says slot agents run sdlc inside the Bash sandbox (fence refused, clone-only)
          detail: With "sdlc *" excluded, a plain leading sdlc call is unsandboxed, so the reviewbox fence probe can now succeed. Reword it to cover only sdlc calls that stay sandboxed (#334).
          family: consumer-docs-stale-after-behavior-change
          round: 1
        - id: BR-2
          severity: Minor
          title: TestBaseSettingsExcludedVerbsMatchArguments hard-codes only sdlc/couch; weave/datatype/doc-review/vocabulary/pair* can revert to bare names unnoticed
          detail: Derive the ariadne list from cmd/* (and keep the pair list explicit) so every exempted workbench binary must appear as "X *".
          family: guard-test-covers-subset
          round: 1
        - id: BR-3
          severity: Minor
          title: sandbox.md says "verified live" for redirect and chained forms, but the Log only tested cd-other-dir, full path and $(...)
          family: doc-overclaims-verification
          round: 1
        - id: BR-4
          severity: Minor
          title: Done-when asks for bypass-permissions mode, but that mode was argued from the docs, not checked live
          family: done-when-partially-inferred
          round: 1
        - id: BR-5
          severity: Minor
          title: settings.ariadne.json $comment, and the atlas line on permissions.allow, still describe only sdlc/couch
          family: stale-inline-comment
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
---

# Gate ledger — ariadne#334 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T23:53:22-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `consumer-docs-stale-after-behavior-change` atlas/workflow/sdlc-binary.md:963 still says slot agents run sdlc inside the Bash sandbox (fence refused, clone-only)
  With "sdlc *" excluded, a plain leading sdlc call is unsandboxed, so the reviewbox fence probe can now succeed. Reword it to cover only sdlc calls that stay sandboxed (#334).
- **BR-2** [Minor] `guard-test-covers-subset` TestBaseSettingsExcludedVerbsMatchArguments hard-codes only sdlc/couch; weave/datatype/doc-review/vocabulary/pair* can revert to bare names unnoticed
  Derive the ariadne list from cmd/* (and keep the pair list explicit) so every exempted workbench binary must appear as "X *".
- **BR-3** [Minor] `doc-overclaims-verification` sandbox.md says "verified live" for redirect and chained forms, but the Log only tested cd-other-dir, full path and $(...)
- **BR-4** [Minor] `done-when-partially-inferred` Done-when asks for bypass-permissions mode, but that mode was argued from the docs, not checked live
- **BR-5** [Minor] `stale-inline-comment` settings.ariadne.json $comment, and the atlas line on permissions.allow, still describe only sdlc/couch

## Open findings

- **BR-1** [Important] `consumer-docs-stale-after-behavior-change` atlas/workflow/sdlc-binary.md:963 still says slot agents run sdlc inside the Bash sandbox (fence refused, clone-only)
- **BR-2** [Minor] `guard-test-covers-subset` TestBaseSettingsExcludedVerbsMatchArguments hard-codes only sdlc/couch; weave/datatype/doc-review/vocabulary/pair* can revert to bare names unnoticed
- **BR-3** [Minor] `doc-overclaims-verification` sandbox.md says "verified live" for redirect and chained forms, but the Log only tested cd-other-dir, full path and $(...)
- **BR-4** [Minor] `done-when-partially-inferred` Done-when asks for bypass-permissions mode, but that mode was argued from the docs, not checked live
- **BR-5** [Minor] `stale-inline-comment` settings.ariadne.json $comment, and the atlas line on permissions.allow, still describe only sdlc/couch
