---
gate: boundary-review
issue: 280
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-02T11:08:23-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Recovery contracts describe tracker repositories only but render into every repository's help; claim, change-code and set-status are wrong in legacy repositories
          detail: claim says "Never pushes main" and change-code says "Never pushes", but runLegacyClaim (legacymode.go) and syncLegacyIssue (changecode.go) publish to main. Either scope each Section to issue-tracker repositories (#252) and mark legacy as unknown, or add legacy wording to those contracts.
          family: contract-scope-unstated
          round: 1
        - id: BR-2
          severity: Important
          title: The plan's Core concepts table names render.go, Page() and JSON(), which do not exist (Section is in contracts.go; JSON was dropped in PQ-2)
          detail: Add a Revisions entry so the table matches the code, and mark Page() and Example as M2.
          family: plan-table-drift
          round: 1
        - id: BR-3
          severity: Minor
          title: The recovery package doc still describes the dropped {{RECOVERY}} placeholder
          family: stale-doc-comment
          round: 1
        - id: BR-4
          severity: Minor
          title: Every M1 help section points at sdlc help recovery, which does not exist until M2
          family: dangling-help-pointer
          round: 1
        - id: BR-5
          severity: Minor
          title: TestCardPublishCallers does not require reclaimEffect and skips other var-declared function literals
          family: caller-guard-gaps
          round: 1
        - id: BR-6
          severity: Minor
          title: The plan's lost-acknowledgement proof TestUpdateMany_LostAcknowledgmentIsUncertainWithoutReplay is cited by no catalog entry
          family: plan-proof-omitted
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-02T11:10:43-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: recovery.Scope rendered by Section into every help section (contracts.go:135); TestSectionMarksUnprovenClaims asserts Scope/legacy/unknown; verified in sdlc claim --help.
          round: 2
        - id: BR-2
          disposition: addressed
          note: 'Plan Revisions entry supersedes the table rows: no render.go, JSON dropped (PQ-2), Page/Example are M2.'
          round: 2
        - id: BR-3
          disposition: addressed
          note: Package doc now names attachRecoveryContracts; no placeholder reference remains.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Revisions records that sdlc help recovery lands on this branch in M2 before anything ships; acceptable at an intra-branch milestone.
          round: 2
        - id: BR-5
          disposition: addressed
          note: reclaimEffect added to allowed (so required by the not-seen loop); guard inspects every ValueSpec initializer.
          round: 2
        - id: BR-6
          disposition: addressed
          note: catalog.go:25 cites the test; it exists in internal/gitx and passes; it shares the updateMany core with UpdateCardWithTrailers.
          round: 2
      findings:
        - id: BR-7
          severity: Minor
          title: atlas/workflow/recovery-contracts.md presents the contracts as the rulebook without their issue-tracker-only scope
          detail: 'Repeat of the family. Rule: every surface describing the contracts states the scope; generated surfaces (help sections, M2 Page) render recovery.Scope, and the hand-written atlas page carries one scope line. Fix the class in M2 by rendering Scope in Page() and adding the line to the atlas.'
          family: contract-scope-unstated
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#280 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T11:08:23-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `contract-scope-unstated` Recovery contracts describe tracker repositories only but render into every repository's help; claim, change-code and set-status are wrong in legacy repositories
  claim says "Never pushes main" and change-code says "Never pushes", but runLegacyClaim (legacymode.go) and syncLegacyIssue (changecode.go) publish to main. Either scope each Section to issue-tracker repositories (#252) and mark legacy as unknown, or add legacy wording to those contracts.
- **BR-2** [Important] `plan-table-drift` The plan's Core concepts table names render.go, Page() and JSON(), which do not exist (Section is in contracts.go; JSON was dropped in PQ-2)
  Add a Revisions entry so the table matches the code, and mark Page() and Example as M2.
- **BR-3** [Minor] `stale-doc-comment` The recovery package doc still describes the dropped {{RECOVERY}} placeholder
- **BR-4** [Minor] `dangling-help-pointer` Every M1 help section points at sdlc help recovery, which does not exist until M2
- **BR-5** [Minor] `caller-guard-gaps` TestCardPublishCallers does not require reclaimEffect and skips other var-declared function literals
- **BR-6** [Minor] `plan-proof-omitted` The plan's lost-acknowledgement proof TestUpdateMany_LostAcknowledgmentIsUncertainWithoutReplay is cited by no catalog entry

## Round 2 — 2026-10-02T11:10:43-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — recovery.Scope rendered by Section into every help section (contracts.go:135); TestSectionMarksUnprovenClaims asserts Scope/legacy/unknown; verified in sdlc claim --help.
- BR-2 — addressed — Plan Revisions entry supersedes the table rows: no render.go, JSON dropped (PQ-2), Page/Example are M2.
- BR-3 — addressed — Package doc now names attachRecoveryContracts; no placeholder reference remains.
- BR-4 — addressed — Revisions records that sdlc help recovery lands on this branch in M2 before anything ships; acceptable at an intra-branch milestone.
- BR-5 — addressed — reclaimEffect added to allowed (so required by the not-seen loop); guard inspects every ValueSpec initializer.
- BR-6 — addressed — catalog.go:25 cites the test; it exists in internal/gitx and passes; it shares the updateMany core with UpdateCardWithTrailers.

### Raised

- **BR-7** [Minor] `contract-scope-unstated` atlas/workflow/recovery-contracts.md presents the contracts as the rulebook without their issue-tracker-only scope
  Repeat of the family. Rule: every surface describing the contracts states the scope; generated surfaces (help sections, M2 Page) render recovery.Scope, and the hand-written atlas page carries one scope line. Fix the class in M2 by rendering Scope in Page() and adding the line to the atlas.

## Open findings

- **BR-7** [Minor] `contract-scope-unstated` atlas/workflow/recovery-contracts.md presents the contracts as the rulebook without their issue-tracker-only scope
