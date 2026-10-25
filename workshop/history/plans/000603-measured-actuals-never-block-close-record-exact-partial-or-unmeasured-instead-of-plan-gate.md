---
reviewer_runs:
    - version: 1
      intent:
        id: ebbca34d-f817-4bf1-a63d-5cac4ca2318d
        repository: github.com/xianxu/ariadne
        issue_id: "000603"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        reviewed_head: 332eecb1186b0b3de0520292530eb6897df4ca31
        slot: ariadne:2
        requested_at: "2026-10-24T23:10:40.754548Z"
      phase: finished
      pid: 48078
      started_at: "2026-10-24T23:10:40.919233Z"
      finished_at: "2026-10-24T23:11:09.91992Z"
      session_id: 2592f3cf-3d07-4851-b3fb-843d350f0b79
      outcome: returned
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
    - reviewer_runs:
        - version: 1
          intent:
            id: 143e7460-78aa-43f1-a357-e2163a2cb7dc
            repository: github.com/xianxu/ariadne
            issue_id: "000603"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: db2ca0ae03f0e1174cda6e59516a603125e66aa8
            slot: ariadne:2
            requested_at: "2026-10-24T23:08:45.875202Z"
          phase: finished
          pid: 80626
          started_at: "2026-10-24T23:08:46.019264Z"
          finished_at: "2026-10-24T23:09:11.358083Z"
          session_id: f8963ec4-81b2-4460-b8be-af31643b9a03
          outcome: returned
      "n": 2
      timestamp: "2026-10-24T16:09:11-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-1
          disposition: addressed
          note: New top-level unattributed member; extra-field preservation exists (measurement.go:74); kind enum and proof validator untouched.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Summary retains bounded Credited intervals consumed by loadAdoptedExclusions; Validate split so the proof call at actualbranch.go:313 checks shape only.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: 'Single meaning: lower bound; unbounded unknowns are upper bound with weight 0.'
          round: 2
        - id: PQ-4
          disposition: addressed
          note: All queue sites mapped to bounded/unbounded/hard-error outcomes; exit located at close.go:459.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: 838/851 confirmed as source-cap at branch proof, still live after 593.
          round: 2
      findings:
        - kind: defect
          id: PQ-6
          severity: Minor
          title: Tasks 1, 6 and 12 enumerate test cases in prose instead of one strategy line per risky function
          detail: 'Compress to strategy lines: credit coalescer gets a property test (output length at most MaxSummaryCredits, coalesced cover is a superset of the input cover); weighted calibration gets a property test (adding a weight-0 row leaves output unchanged).'
          family: test-case-enumeration-in-prose
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 76685b1e-8a9e-44e9-bf77-bd358dd4b475
            repository: github.com/xianxu/ariadne
            issue_id: "000603"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 332eecb1186b0b3de0520292530eb6897df4ca31
            slot: ariadne:2
            requested_at: "2026-10-24T23:10:23.920703Z"
          phase: finished
          pid: 40384
          started_at: "2026-10-24T23:10:24.116267Z"
          finished_at: "2026-10-24T23:10:37.518783Z"
          session_id: 08b9e5a4-00cf-4efe-83d6-83d4798bb77c
          outcome: returned
      "n": 3
      timestamp: "2026-10-24T16:10:37-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: PQ-6
          disposition: not-addressed
          note: Strategy lines were added for the coalescer and weighted calibration, but the prose case lists in Tasks 6 and 12 (and Task 1's coverage sentence) remain; delete them in favour of the strategy lines. Minor, carried to close review.
          round: 3
      blocked: false
content_hash: 50cedc7abb2d0cbae8f5dc9d38ff21cc8ff3fc2ec964f2c36a503039dee76645
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

## Round 2 — 2026-10-24T16:09:11-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — New top-level unattributed member; extra-field preservation exists (measurement.go:74); kind enum and proof validator untouched.
- PQ-2 — addressed — Summary retains bounded Credited intervals consumed by loadAdoptedExclusions; Validate split so the proof call at actualbranch.go:313 checks shape only.
- PQ-3 — addressed — Single meaning: lower bound; unbounded unknowns are upper bound with weight 0.
- PQ-4 — addressed — All queue sites mapped to bounded/unbounded/hard-error outcomes; exit located at close.go:459.
- PQ-5 — addressed — 838/851 confirmed as source-cap at branch proof, still live after 593.

### Raised

- **PQ-6** [Minor] `test-case-enumeration-in-prose` (kind: defect) Tasks 1, 6 and 12 enumerate test cases in prose instead of one strategy line per risky function
  Compress to strategy lines: credit coalescer gets a property test (output length at most MaxSummaryCredits, coalesced cover is a superset of the input cover); weighted calibration gets a property test (adding a weight-0 row leaves output unchanged).

## Round 3 — 2026-10-24T16:10:37-07:00 (claude) — passed

### Disposed

- PQ-6 — not-addressed — Strategy lines were added for the coalescer and weighted calibration, but the prose case lists in Tasks 6 and 12 (and Task 1's coverage sentence) remain; delete them in favour of the strategy lines. Minor, carried to close review.

## Open findings

- **PQ-6** [Minor] `test-case-enumeration-in-prose` (kind: defect) Tasks 1, 6 and 12 enumerate test cases in prose instead of one strategy line per risky function
