---
gate: plan-quality
issue: 304
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-09T20:20:59-07:00"
      agent: claude
      blocked: true
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-10-09T20:23:06-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Minor
          title: D8 settle-path scan bound cites a "card closed date" that the card schema lacks
          detail: issue.Completion and the vocabulary carry no close date; `updated` moves on later setter edits and could start the --since scan after the evidence commit, so settle never matches and the card stays codecomplete. Name the concrete source (e.g. the committer date of the tracker commit that published codecomplete) or bound the scan by commit count.
          family: unbacked-existing-field-claim
          round: 2
        - id: PQ-2
          severity: Minor
          title: Task 6 must order commitMilestoneEvidence and pinReviewed before milestonePush in finalizeBoundaryReview
          detail: close.go calls milestonePush right after applyClose on the milestone path; if the new evidence commit is inserted after it, the push goes out without the evidence commit.
          family: effect-ordering-unstated
          round: 2
        - id: PQ-3
          severity: Minor
          title: Tasks 1, 2, 3, 6, 9 and 10 list test cases in prose; compress each to one strategy line plus its mutation guard
          family: prose-test-enumeration
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-09T20:24:01-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: D8/Task 10 now bound the scan by the committer date of the tracker commit that introduced the token, with a test that moves `updated`.
          round: 3
        - id: PQ-2
          disposition: addressed
          note: Task 6 orders the evidence commit and pin after applyClose and before milestonePush, with a test that the pushed tip equals the evidence commit.
          round: 3
        - id: PQ-3
          disposition: not-addressed
          note: Deliberately kept as the TL-approved test contract (Revisions); stays a Minor for the close review.
          round: 3
      blocked: false
content_hash: 56bdbf214f79a66cf7b808d9021756c6beb305b01f4bc14abf073bb12fe57218
---

# Gate ledger — ariadne#304 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-09T20:20:59-07:00 (claude) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-10-09T20:23:06-07:00 (claude) — passed

### Raised

- **PQ-1** [Minor] `unbacked-existing-field-claim` D8 settle-path scan bound cites a "card closed date" that the card schema lacks
  issue.Completion and the vocabulary carry no close date; `updated` moves on later setter edits and could start the --since scan after the evidence commit, so settle never matches and the card stays codecomplete. Name the concrete source (e.g. the committer date of the tracker commit that published codecomplete) or bound the scan by commit count.
- **PQ-2** [Minor] `effect-ordering-unstated` Task 6 must order commitMilestoneEvidence and pinReviewed before milestonePush in finalizeBoundaryReview
  close.go calls milestonePush right after applyClose on the milestone path; if the new evidence commit is inserted after it, the push goes out without the evidence commit.
- **PQ-3** [Minor] `prose-test-enumeration` Tasks 1, 2, 3, 6, 9 and 10 list test cases in prose; compress each to one strategy line plus its mutation guard

## Round 3 — 2026-10-09T20:24:01-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — D8/Task 10 now bound the scan by the committer date of the tracker commit that introduced the token, with a test that moves `updated`.
- PQ-2 — addressed — Task 6 orders the evidence commit and pin after applyClose and before milestonePush, with a test that the pushed tip equals the evidence commit.
- PQ-3 — not-addressed — Deliberately kept as the TL-approved test contract (Revisions); stays a Minor for the close review.

## Open findings

- **PQ-3** [Minor] `prose-test-enumeration` Tasks 1, 2, 3, 6, 9 and 10 list test cases in prose; compress each to one strategy line plus its mutation guard
