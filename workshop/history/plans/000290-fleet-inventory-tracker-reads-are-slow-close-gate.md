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
    - "n": 2
      timestamp: "2026-10-02T16:19:09-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan Revisions now states that timeouts are cached and report unknown; this was one of the finding's two accepted options.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: The name is fixed in Revisions, but the plan still calls parseLsRemoteTip "pure, unit-tested" and no direct test exists. A table test (wrong ref, malformed OID, multi-line output) is cheap.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestWarmedInventoryEqualsSequential and TestHangingRemoteDegradesOnlyItsRepository (healthy repo reads present) pass under -race.
          round: 2
        - id: BR-4
          disposition: addressed
          note: issue-tracker.md lines 125, 142 and 260 now describe ls-remote plus a fetch only when the tracker moved.
          round: 2
        - id: BR-5
          disposition: not-addressed
          note: The plan wording is fixed and fetch() clears probedTip (trunkfile.go:201), but no test runs the RemoteExists, refreshTip, Snapshot sequence, so reverting the clear stays green.
          round: 2
        - id: BR-6
          disposition: addressed
          note: The testfix import is now in its own group (snapshot_test.go:10).
          round: 2
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-02T16:20:59-07:00"
      agent: claude
      dispose:
        - id: BR-2
          disposition: addressed
          note: Core concepts table rewritten to as-built names and files (warmRecords in issues.go, trackedRoots in inventory.go); unit-test claim dropped.
          round: 3
        - id: BR-5
          disposition: addressed
          note: fetch() clears probedTip (trunkfile.go:201); the probe-Read-Snapshot block of TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged expects 4 fetches and gets 3 without the clear; Revisions records the skip reaching write verbs.
          round: 3
      findings:
        - id: BR-7
          severity: Minor
          title: Hanging-remote test bound (5s vs 500ms deadline) cannot detect the walk waiting a second time
          detail: 'This is the 2nd finding in family done-when-untested. Rule: each Done-when clause needs an assertion that fails when that clause is violated, not just a loose wall-clock bound. Here, count cache loads for the stuck repository (expect 1) so a warm-up/walk key mismatch or an uncached timeout fails the test.'
          family: done-when-untested
          round: 3
        - id: BR-8
          severity: Minor
          title: parseLsRemoteTip (declared pure) has no colocated unit test
          detail: It is only exercised through the real-git snapshot test; add a table test for a different ref name, a malformed object ID and empty output.
          family: pure-entity-untested
          round: 3
      recipe: milestone-review
      blocked: false
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

## Round 2 — 2026-10-02T16:19:09-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Plan Revisions now states that timeouts are cached and report unknown; this was one of the finding's two accepted options.
- BR-2 — not-addressed — The name is fixed in Revisions, but the plan still calls parseLsRemoteTip "pure, unit-tested" and no direct test exists. A table test (wrong ref, malformed OID, multi-line output) is cheap.
- BR-3 — addressed — TestWarmedInventoryEqualsSequential and TestHangingRemoteDegradesOnlyItsRepository (healthy repo reads present) pass under -race.
- BR-4 — addressed — issue-tracker.md lines 125, 142 and 260 now describe ls-remote plus a fetch only when the tracker moved.
- BR-5 — not-addressed — The plan wording is fixed and fetch() clears probedTip (trunkfile.go:201), but no test runs the RemoteExists, refreshTip, Snapshot sequence, so reverting the clear stays green.
- BR-6 — addressed — The testfix import is now in its own group (snapshot_test.go:10).

## Round 3 — 2026-10-02T16:20:59-07:00 (claude) — passed

### Disposed

- BR-2 — addressed — Core concepts table rewritten to as-built names and files (warmRecords in issues.go, trackedRoots in inventory.go); unit-test claim dropped.
- BR-5 — addressed — fetch() clears probedTip (trunkfile.go:201); the probe-Read-Snapshot block of TestSnapshotSkipsTheFetchWhenTheTipIsUnchanged expects 4 fetches and gets 3 without the clear; Revisions records the skip reaching write verbs.

### Raised

- **BR-7** [Minor] `done-when-untested` Hanging-remote test bound (5s vs 500ms deadline) cannot detect the walk waiting a second time
  This is the 2nd finding in family done-when-untested. Rule: each Done-when clause needs an assertion that fails when that clause is violated, not just a loose wall-clock bound. Here, count cache loads for the stuck repository (expect 1) so a warm-up/walk key mismatch or an uncached timeout fails the test.
- **BR-8** [Minor] `pure-entity-untested` parseLsRemoteTip (declared pure) has no colocated unit test
  It is only exercised through the real-git snapshot test; add a table test for a different ref name, a malformed object ID and empty output.

## Open findings

- **BR-7** [Minor] `done-when-untested` Hanging-remote test bound (5s vs 500ms deadline) cannot detect the walk waiting a second time
- **BR-8** [Minor] `pure-entity-untested` parseLsRemoteTip (declared pure) has no colocated unit test
