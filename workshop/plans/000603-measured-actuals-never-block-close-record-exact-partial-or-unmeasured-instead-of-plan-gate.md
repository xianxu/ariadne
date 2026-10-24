---
gate: plan-quality
issue: 603
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 0a540598-f360-49b3-8d41-c8d11d694a9f
            repository: github.com/xianxu/ariadne
            issue_id: "000603"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 5f6d775f9495619076224aaf0aa308edb18d1aa9
            slot: ariadne:2
            requested_at: "2026-10-24T23:04:15.345807Z"
          phase: finished
          pid: 40525
          started_at: "2026-10-24T23:04:15.4432Z"
          finished_at: "2026-10-24T23:05:33.436172Z"
          session_id: 38102b04-f394-4829-b3ce-d232f5afa140
          outcome: returned
      "n": 1
      timestamp: "2026-10-24T16:05:33-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: New branch-note kinds are rejected by the strict gatestate kind enum and mishandled by the branch-proof validator
          detail: measurement_branch.go:97 rejects unknown kinds; actualbranch.go:321-329 treats any non-excluded kind as a selected witness; older peer binaries replaying the snapshot would refuse. The plan must state the enum extension, older-reader behaviour (version boundary or safe-ignore argument), proof-validator handling, and where source-unattributed notes persist when the source has no branch capture.
          family: persisted-enum-extension-unplanned
          round: 1
        - kind: defect
          id: PQ-2
          severity: Important
          title: MeasurementSummary drops the windows exclusion needs, so every later overlapping close goes unmeasured
          detail: actualexclusion.go:369-372 errors with no trustworthy saved window; oversized windows are the ones most likely to overlap peers. Retain merged per-source windows in the summary so exclusion still works, and handle the second m.Validate() inside validateOwnedBranchProof (actualbranch.go:313).
          family: degradation-cascades-to-peers
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: partial hours are a lower bound in D3 but an upper bound in D4 and the queue-with-bounds case
          family: confidence-semantics-undefined
          round: 1
        - kind: defect
          id: PQ-4
          severity: Minor
          title: Queue degradation spans four distinct queueEvidenceError sites; explainActual does not itself exit (caller at close.go:459)
          detail: 'Sites: actualinputs.go:437, 472, 483 and actualqueue_filter.go:55. Name which become partial and which stay unmeasured.'
          family: unbacked-existing-behavior-claim
          round: 1
        - kind: defect
          id: PQ-5
          severity: Minor
          title: The pair 838 fixture may test a shape already fixed by 593; confirm what it exercises
          family: replay-fixture-scope
          round: 1
      blocked: true
---

# Gate ledger — ariadne#603 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T16:05:33-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `persisted-enum-extension-unplanned` (kind: defect) New branch-note kinds are rejected by the strict gatestate kind enum and mishandled by the branch-proof validator
  measurement_branch.go:97 rejects unknown kinds; actualbranch.go:321-329 treats any non-excluded kind as a selected witness; older peer binaries replaying the snapshot would refuse. The plan must state the enum extension, older-reader behaviour (version boundary or safe-ignore argument), proof-validator handling, and where source-unattributed notes persist when the source has no branch capture.
- **PQ-2** [Important] `degradation-cascades-to-peers` (kind: defect) MeasurementSummary drops the windows exclusion needs, so every later overlapping close goes unmeasured
  actualexclusion.go:369-372 errors with no trustworthy saved window; oversized windows are the ones most likely to overlap peers. Retain merged per-source windows in the summary so exclusion still works, and handle the second m.Validate() inside validateOwnedBranchProof (actualbranch.go:313).
- **PQ-3** [Minor] `confidence-semantics-undefined` (kind: defect) partial hours are a lower bound in D3 but an upper bound in D4 and the queue-with-bounds case
- **PQ-4** [Minor] `unbacked-existing-behavior-claim` (kind: defect) Queue degradation spans four distinct queueEvidenceError sites; explainActual does not itself exit (caller at close.go:459)
  Sites: actualinputs.go:437, 472, 483 and actualqueue_filter.go:55. Name which become partial and which stay unmeasured.
- **PQ-5** [Minor] `replay-fixture-scope` (kind: defect) The pair 838 fixture may test a shape already fixed by 593; confirm what it exercises

## Open findings

- **PQ-1** [Important] `persisted-enum-extension-unplanned` (kind: defect) New branch-note kinds are rejected by the strict gatestate kind enum and mishandled by the branch-proof validator
- **PQ-2** [Important] `degradation-cascades-to-peers` (kind: defect) MeasurementSummary drops the windows exclusion needs, so every later overlapping close goes unmeasured
- **PQ-3** [Minor] `confidence-semantics-undefined` (kind: defect) partial hours are a lower bound in D3 but an upper bound in D4 and the queue-with-bounds case
- **PQ-4** [Minor] `unbacked-existing-behavior-claim` (kind: defect) Queue degradation spans four distinct queueEvidenceError sites; explainActual does not itself exit (caller at close.go:459)
- **PQ-5** [Minor] `replay-fixture-scope` (kind: defect) The pair 838 fixture may test a shape already fixed by 593; confirm what it exercises
