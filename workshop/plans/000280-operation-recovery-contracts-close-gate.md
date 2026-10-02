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

## Open findings

- **BR-1** [Important] `contract-scope-unstated` Recovery contracts describe tracker repositories only but render into every repository's help; claim, change-code and set-status are wrong in legacy repositories
- **BR-2** [Important] `plan-table-drift` The plan's Core concepts table names render.go, Page() and JSON(), which do not exist (Section is in contracts.go; JSON was dropped in PQ-2)
- **BR-3** [Minor] `stale-doc-comment` The recovery package doc still describes the dropped {{RECOVERY}} placeholder
- **BR-4** [Minor] `dangling-help-pointer` Every M1 help section points at sdlc help recovery, which does not exist until M2
- **BR-5** [Minor] `caller-guard-gaps` TestCardPublishCallers does not require reclaimEffect and skips other var-declared function literals
- **BR-6** [Minor] `plan-proof-omitted` The plan's lost-acknowledgement proof TestUpdateMany_LostAcknowledgmentIsUncertainWithoutReplay is cited by no catalog entry
