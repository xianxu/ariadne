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
    - "n": 2
      timestamp: "2026-10-11T23:59:55-07:00"
      dispatched: "2026-10-11T23:58:53-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: sdlc-binary.md:963-967 now says the fence can apply for plain sdlc calls, clone-only only for forms that stay sandboxed; matches sandbox.md rule.
          round: 2
        - id: BR-2
          disposition: addressed
          note: base_settings_test.go:58-74 requires "X *" for every cmd/ dir plus explicit pair list; passes at head; bare entry fails by inspection (revert not runnable in read-only mode).
          round: 2
        - id: BR-3
          disposition: addressed
          note: sandbox.md:23 now lists cd-other-dir, full path, $(...), pipe as live and says redirect is from the docs; Log records the pipe observation.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Issue Revisions entry records that bypass-permissions was inferred from the docs, not observed.
          round: 2
        - id: BR-5
          disposition: addressed
          note: $comment and sandbox.md:13 now name all ariadne cmd/ and pair binaries; sdlc/couch singled out only as the pre-allowed pair.
          round: 2
      findings:
        - id: BR-6
          severity: Minor
          title: sandbox.md:25 and the test header describe TestBaseSettingsExcludedVerbsMatchArguments as checking only pre-allowed verbs
          detail: 'The BR-2 fix added a second check (every cmd/ dir and pair binary must be "X *"), but the atlas line and the test''s header comment (base_settings_test.go:11-17) still cover only the first. 2nd finding in this family. Rule: when a fix round changes a test''s or binary''s behavior, grep its name across atlas/ and code comments and update every description in the same round. Measured: 1 atlas line plus 1 header comment.'
          family: consumer-docs-stale-after-behavior-change
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: eb685d1185db420c0d754dfa20d5a9042d16fa84
      blocked: false
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

## Round 2 — 2026-10-11T23:59:55-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — sdlc-binary.md:963-967 now says the fence can apply for plain sdlc calls, clone-only only for forms that stay sandboxed; matches sandbox.md rule.
- BR-2 — addressed — base_settings_test.go:58-74 requires "X *" for every cmd/ dir plus explicit pair list; passes at head; bare entry fails by inspection (revert not runnable in read-only mode).
- BR-3 — addressed — sandbox.md:23 now lists cd-other-dir, full path, $(...), pipe as live and says redirect is from the docs; Log records the pipe observation.
- BR-4 — addressed — Issue Revisions entry records that bypass-permissions was inferred from the docs, not observed.
- BR-5 — addressed — $comment and sandbox.md:13 now name all ariadne cmd/ and pair binaries; sdlc/couch singled out only as the pre-allowed pair.

### Raised

- **BR-6** [Minor] `consumer-docs-stale-after-behavior-change` sandbox.md:25 and the test header describe TestBaseSettingsExcludedVerbsMatchArguments as checking only pre-allowed verbs
  The BR-2 fix added a second check (every cmd/ dir and pair binary must be "X *"), but the atlas line and the test's header comment (base_settings_test.go:11-17) still cover only the first. 2nd finding in this family. Rule: when a fix round changes a test's or binary's behavior, grep its name across atlas/ and code comments and update every description in the same round. Measured: 1 atlas line plus 1 header comment.

## Open findings

- **BR-6** [Minor] `consumer-docs-stale-after-behavior-change` sandbox.md:25 and the test header describe TestBaseSettingsExcludedVerbsMatchArguments as checking only pre-allowed verbs
