---
gate: boundary-review
issue: 290
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-02T16:10:00-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Deadline decision says timeouts are not cached and fall back to stale; code caches them and reports unknown
          detail: issues.go:116 keeps DeadlineExceeded entries, and the local fallback in Presence/LocalSnapshot shares the expired TrunkFile context, so a tracked repository with a fetched tracker reports unknown instead of stale. Either run the local fallback outside the deadline, or add a plan Revisions entry (the Log notes the deviation; the plan does not).
          family: plan-code-drift
          round: 1
        - id: BR-2
          severity: Important
          title: Core concepts table names warmRepoRecords in inventory.go; the code has warmRecords in issues.go
          detail: The table also says parseLsRemoteTip is unit-tested, but there is no direct test; it is only exercised through TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged. Revise the table or add the unit test.
          family: plan-code-drift
          round: 1
        - id: BR-3
          severity: Important
          title: 'Done-when items untested: concurrent output equals a sequential run; a hanging remote leaves other repositories unaffected'
          detail: TestHangingRemoteDegradesWithinTheDeadline uses a single repository, although the plan asked for a second one that reads present. No test compares a warmed CollectInventory against a sequential run; both checks were only done by hand per the Log.
          family: done-when-untested
          round: 1
        - id: BR-4
          severity: Important
          title: atlas/workflow/issue-tracker.md still says every snapshot fetches; the fetch skip now applies to every Snapshot caller
          detail: Lines 125 and 141 describe one tracker fetch per read. The plan listed this file for an update, but only sdlc-binary.md changed.
          family: atlas-stale-after-contract-change
          round: 1
        - id: BR-5
          severity: Minor
          title: probedTip survives an intervening fetch or refreshTip, so the "right after a probe" rule is not enforced
          detail: Clear probedTip inside fetch(). The skip also reaches write verbs (Initialized then Snapshot), contrary to the plan's "write paths untouched"; this looks safe because CAS still fetches, but the plan should say so.
          family: implicit-cross-call-state
          round: 1
        - id: BR-6
          severity: Minor
          title: snapshot_test.go puts the testfix import inside the standard-library import group
          family: import-grouping
          round: 1
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#290 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T16:10:00-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `plan-code-drift` Deadline decision says timeouts are not cached and fall back to stale; code caches them and reports unknown
  issues.go:116 keeps DeadlineExceeded entries, and the local fallback in Presence/LocalSnapshot shares the expired TrunkFile context, so a tracked repository with a fetched tracker reports unknown instead of stale. Either run the local fallback outside the deadline, or add a plan Revisions entry (the Log notes the deviation; the plan does not).
- **BR-2** [Important] `plan-code-drift` Core concepts table names warmRepoRecords in inventory.go; the code has warmRecords in issues.go
  The table also says parseLsRemoteTip is unit-tested, but there is no direct test; it is only exercised through TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged. Revise the table or add the unit test.
- **BR-3** [Important] `done-when-untested` Done-when items untested: concurrent output equals a sequential run; a hanging remote leaves other repositories unaffected
  TestHangingRemoteDegradesWithinTheDeadline uses a single repository, although the plan asked for a second one that reads present. No test compares a warmed CollectInventory against a sequential run; both checks were only done by hand per the Log.
- **BR-4** [Important] `atlas-stale-after-contract-change` atlas/workflow/issue-tracker.md still says every snapshot fetches; the fetch skip now applies to every Snapshot caller
  Lines 125 and 141 describe one tracker fetch per read. The plan listed this file for an update, but only sdlc-binary.md changed.
- **BR-5** [Minor] `implicit-cross-call-state` probedTip survives an intervening fetch or refreshTip, so the "right after a probe" rule is not enforced
  Clear probedTip inside fetch(). The skip also reaches write verbs (Initialized then Snapshot), contrary to the plan's "write paths untouched"; this looks safe because CAS still fetches, but the plan should say so.
- **BR-6** [Minor] `import-grouping` snapshot_test.go puts the testfix import inside the standard-library import group

## Open findings

- **BR-1** [Important] `plan-code-drift` Deadline decision says timeouts are not cached and fall back to stale; code caches them and reports unknown
- **BR-2** [Important] `plan-code-drift` Core concepts table names warmRepoRecords in inventory.go; the code has warmRecords in issues.go
- **BR-3** [Important] `done-when-untested` Done-when items untested: concurrent output equals a sequential run; a hanging remote leaves other repositories unaffected
- **BR-4** [Important] `atlas-stale-after-contract-change` atlas/workflow/issue-tracker.md still says every snapshot fetches; the fetch skip now applies to every Snapshot caller
- **BR-5** [Minor] `implicit-cross-call-state` probedTip survives an intervening fetch or refreshTip, so the "right after a probe" rule is not enforced
- **BR-6** [Minor] `import-grouping` snapshot_test.go puts the testfix import inside the standard-library import group
