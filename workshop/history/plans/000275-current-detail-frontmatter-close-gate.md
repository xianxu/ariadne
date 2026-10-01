---
gate: boundary-review
issue: 275
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-30T16:49:57-07:00"
      agent: claude
      recipe: milestone-review
      blocked: false
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-09-30T17:06:11-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Plan's Core-concepts table says landingOwnedIssue carries baseline bytes and the planner stays pure; code injects a readCard IO closure into planLandingArchive
          detail: trackedArchiveBytes (landingarchive.go) calls owned.readCard (env.repo.ReadCardBlob) from inside planLandingArchive; there is no baseline field. Add a Revisions entry or pre-resolve an OID->blob map in selection/confirmation (ARCH-PURE).
          family: plan-table-drift
          round: 2
        - id: BR-2
          severity: Minor
          title: merge.md help text leaves a dangling "An" line after the reflow
          family: helptext-reflow
          round: 2
        - id: BR-3
          severity: Minor
          title: closeMirrorCommit staged-differs-from-HEAD index branch has no test
          family: untested-branch
          round: 2
        - id: BR-4
          severity: Minor
          title: Mirror commit is not retried after a crash between codecomplete publish and the mirror commit
          detail: Documented in the atlas; reconcile could call the idempotent commitCloseMirror when the completion is already finished on the source branch.
          family: crash-window-retry
          round: 2
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-09-30T17:21:00-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan table and Revisions now name the injected readCard seam; matches landingarchive.go:41-42 and :657.
          round: 3
        - id: BR-2
          disposition: addressed
          note: merge.md reflowed; no dangling line.
          round: 3
        - id: BR-3
          disposition: addressed
          note: TestReconcileRetriesAnInterruptedCloseMirror stages an edit and asserts the index keeps it (staged != oldBlob branch).
          round: 3
        - id: BR-4
          disposition: addressed
          note: retryCloseMirror runs in reconcile's defer; the test resets to evidenceRev and reconcile restores the mirror commit; a second run is a no-op.
          round: 3
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#275 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-30T16:49:57-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-09-30T17:06:11-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `plan-table-drift` Plan's Core-concepts table says landingOwnedIssue carries baseline bytes and the planner stays pure; code injects a readCard IO closure into planLandingArchive
  trackedArchiveBytes (landingarchive.go) calls owned.readCard (env.repo.ReadCardBlob) from inside planLandingArchive; there is no baseline field. Add a Revisions entry or pre-resolve an OID->blob map in selection/confirmation (ARCH-PURE).
- **BR-2** [Minor] `helptext-reflow` merge.md help text leaves a dangling "An" line after the reflow
- **BR-3** [Minor] `untested-branch` closeMirrorCommit staged-differs-from-HEAD index branch has no test
- **BR-4** [Minor] `crash-window-retry` Mirror commit is not retried after a crash between codecomplete publish and the mirror commit
  Documented in the atlas; reconcile could call the idempotent commitCloseMirror when the completion is already finished on the source branch.

## Round 3 — 2026-09-30T17:21:00-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Plan table and Revisions now name the injected readCard seam; matches landingarchive.go:41-42 and :657.
- BR-2 — addressed — merge.md reflowed; no dangling line.
- BR-3 — addressed — TestReconcileRetriesAnInterruptedCloseMirror stages an edit and asserts the index keeps it (staged != oldBlob branch).
- BR-4 — addressed — retryCloseMirror runs in reconcile's defer; the test resets to evidenceRev and reconcile restores the mirror commit; a second run is a no-op.

## Open findings

(none — every finding has been disposed)
