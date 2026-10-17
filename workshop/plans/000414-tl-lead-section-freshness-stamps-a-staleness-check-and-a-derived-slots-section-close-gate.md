---
gate: boundary-review
issue: 414
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 56c4edc5-aa25-44aa-969d-98da2848a151
            repository: github.com/xianxu/ariadne
            issue_id: "000414"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/f0defb8211d44440
            reviewed_head: 78490f1455b440e25bf2c8f6848e9eb002e00826
            slot: ariadne:4
            requested_at: "2026-10-17T00:40:10.461601Z"
          phase: finished
          pid: 56019
          started_at: "2026-10-17T00:40:10.570033Z"
          finished_at: "2026-10-17T00:42:18.980192Z"
          session_id: b998f0b5-b665-4faf-ad0d-3f1ca29aaa6b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 78490f1455b440e25bf2c8f6848e9eb002e00826
        tracker_head: 9e64ed3994eb94b6d284fe5939c0ae7008439321
        branch_point: e1bb1ba557be3270d16129e8d9dd311cf1b22ccd
        since: "2026-10-16T16:53:29-07:00"
        until: "2026-10-16T17:03:05-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T16:53:29-07:00"
              end: "2026-10-16T17:03:05-07:00"
              began: "2026-10-16T16:53:29-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: ee6840fda242a9b9cc17306c801262657024714aaa4d8c952b89bab1888dad26
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 3961484
              windows:
                - start: "2026-10-16T16:53:29-07:00"
                  end: "2026-10-16T17:03:05-07:00"
      "n": 1
      timestamp: "2026-10-16T17:42:20-07:00"
      dispatched: "2026-10-16T17:40:10-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: Landing rule compares a minute-truncated stamp against a second-precision landing time
          detail: freshness.go:187 land.At.After(at), where at is parsed from a minute-resolution stamp; a landing earlier in the same minute as the write flags the just-written section stale. Compare land.At.Truncate(time.Minute).After(at).
          family: timestamp-precision-mismatch
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: Missing-section detection matches couch's error text with no conformance check pinning it
          detail: freshness.go:173 substring-matches "heading line, found 0"; add a missing-section read to TestSessionSyncConformance (TL_LEAD_LIVE) and an Actors() call for the couch --actors JSON shape.
          family: unpinned-external-contract
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Empty deps.landing and deps.now fall back to real git and the wall clock in tests
          detail: 'Most existing tests build deps{ss: f}, so every write''s printStale runs gitNewestLanding against the temp dir. Set the production values explicitly in main() and make an empty field a no-op.'
          family: nil-seam-defaults-to-production
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: Plan says check is one couch read and shows a claude·working label; code does 4 reads and renders claude / claude·idle
          detail: Add a Revisions entry so the plan matches the delivered behavior.
          family: plan-text-drift
          round: 1
      recipe: milestone-review
      isolation: fenced
      reviewed: 78490f1455b440e25bf2c8f6848e9eb002e00826
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 6c28ce3b-abe5-42c6-a3e7-593a675eb99b
            repository: github.com/xianxu/ariadne
            issue_id: "000414"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/cc9d3235aabb704b
            reviewed_head: 2059abecb29b36c46b1e01b733adba7b26abfb87
            slot: ariadne:4
            requested_at: "2026-10-17T00:43:41.072807Z"
          phase: finished
          pid: 77017
          started_at: "2026-10-17T00:43:41.404855Z"
          finished_at: "2026-10-17T00:44:40.920161Z"
          session_id: c9ea44d6-3002-4ccd-a435-c383448f6ddc
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 2059abecb29b36c46b1e01b733adba7b26abfb87
        tracker_head: b317b912078cb7a47a3ca8934862983b6f8e0e5b
        branch_point: e1bb1ba557be3270d16129e8d9dd311cf1b22ccd
        since: "2026-10-16T16:53:29-07:00"
        until: "2026-10-16T17:43:08-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T16:53:29-07:00"
              end: "2026-10-16T17:43:08-07:00"
              began: "2026-10-16T16:53:29-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 632e7d510ccdd7c4422d0652166f5d74d725d8ba3d73176a4ba7e2a1065828a7
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 4445122
              windows:
                - start: "2026-10-16T16:53:29-07:00"
                  end: "2026-10-16T17:43:08-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.042216666666666666
        cuts:
            - dispatched: "2026-10-16T17:40:10-07:00"
              finalized: "2026-10-16T17:42:20-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-16T17:44:42-07:00"
      dispatched: "2026-10-16T17:43:41-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: freshness.go:70 now compares land.At.Truncate(time.Minute) with the stamp time. Reverting it turns TestStatusCheckEdges red (verified in this checkout).
          round: 2
        - id: BR-2
          disposition: addressed
          note: missingSection is a named constant. TestSessionSyncConformance reads a missing section against it, and TestCouchActorsConformance pins the --actors shape. Both run only with TL_LEAD_LIVE=1, as round one asked.
          round: 2
        - id: BR-3
          disposition: addressed
          note: main.go now wires landing to gitNewestLanding, and an empty landing seam returns no landing, so tests no longer run git. An empty clock still falls back to time.Now; that is the usual Go pattern and reads no outside state.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The plan's Revisions entry (2026-10-16 close review) records the 4 session-sync reads, the agent and agent-idle labels, none for an empty slot, and that only SDLC siblings count as landing sources. This matches freshness.go and slots.go.
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: 2059abecb29b36c46b1e01b733adba7b26abfb87
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 914fcde3-b88a-4b75-aeea-2a3a1a64285c
            repository: github.com/xianxu/ariadne
            issue_id: "000414"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/f7c3c50102111df8
            reviewed_head: 87b44df497231d27fb4ec1a473f4efa0f0864aa3
            slot: ariadne:4
            requested_at: "2026-10-17T02:23:05.195304Z"
          phase: finished
          pid: 93006
          started_at: "2026-10-17T02:23:05.304299Z"
          finished_at: "2026-10-17T02:24:06.253653Z"
          session_id: 794e6073-dbda-4ab2-a93d-9009fb651564
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 87b44df497231d27fb4ec1a473f4efa0f0864aa3
        tracker_head: adc662fa47e13e1fd97aec5e2bfc5b19129c977e
        branch_point: 63b1a010fc175881b8311e8588676a70991a8ad3
        since: "2026-10-16T16:53:29-07:00"
        until: "2026-10-16T17:44:49-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T16:53:29-07:00"
              end: "2026-10-16T17:44:49-07:00"
              began: "2026-10-16T16:53:29-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 22b234de8053250e752b58c6b3ba953e196860943abfb5baa1d7788bb1a40dc1
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 6272872
              windows:
                - start: "2026-10-16T16:53:29-07:00"
                  end: "2026-10-16T17:44:49-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.042216666666666666
        cuts:
            - dispatched: "2026-10-16T17:40:10-07:00"
              finalized: "2026-10-16T17:42:20-07:00"
              finalizing: true
            - dispatched: "2026-10-16T17:43:41-07:00"
              finalized: "2026-10-16T17:44:42-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-16T19:24:08-07:00"
      dispatched: "2026-10-16T19:23:05-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Already disposed in round 2; the window has no code change that reverses the minute-truncation fix.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Already disposed in round 2; the missingSection constant and the conformance tests are unchanged in this window.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Already disposed in round 2; the seam wiring is unchanged in this window.
          round: 3
        - id: BR-4
          disposition: addressed
          note: Already disposed in round 2; the plan Revisions entry stands and the window does not edit it.
          round: 3
      recipe: fix-review
      isolation: checkout
      reviewed: 87b44df497231d27fb4ec1a473f4efa0f0864aa3
      blocked: false
---

# Gate ledger — ariadne#414 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T17:42:20-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `timestamp-precision-mismatch` (kind: defect) Landing rule compares a minute-truncated stamp against a second-precision landing time
  freshness.go:187 land.At.After(at), where at is parsed from a minute-resolution stamp; a landing earlier in the same minute as the write flags the just-written section stale. Compare land.At.Truncate(time.Minute).After(at).
- **BR-2** [Minor] `unpinned-external-contract` (kind: defect) Missing-section detection matches couch's error text with no conformance check pinning it
  freshness.go:173 substring-matches "heading line, found 0"; add a missing-section read to TestSessionSyncConformance (TL_LEAD_LIVE) and an Actors() call for the couch --actors JSON shape.
- **BR-3** [Minor] `nil-seam-defaults-to-production` (kind: defect) Empty deps.landing and deps.now fall back to real git and the wall clock in tests
  Most existing tests build deps{ss: f}, so every write's printStale runs gitNewestLanding against the temp dir. Set the production values explicitly in main() and make an empty field a no-op.
- **BR-4** [Minor] `plan-text-drift` (kind: defect) Plan says check is one couch read and shows a claude·working label; code does 4 reads and renders claude / claude·idle
  Add a Revisions entry so the plan matches the delivered behavior.

## Round 2 — 2026-10-16T17:44:42-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — freshness.go:70 now compares land.At.Truncate(time.Minute) with the stamp time. Reverting it turns TestStatusCheckEdges red (verified in this checkout).
- BR-2 — addressed — missingSection is a named constant. TestSessionSyncConformance reads a missing section against it, and TestCouchActorsConformance pins the --actors shape. Both run only with TL_LEAD_LIVE=1, as round one asked.
- BR-3 — addressed — main.go now wires landing to gitNewestLanding, and an empty landing seam returns no landing, so tests no longer run git. An empty clock still falls back to time.Now; that is the usual Go pattern and reads no outside state.
- BR-4 — addressed — The plan's Revisions entry (2026-10-16 close review) records the 4 session-sync reads, the agent and agent-idle labels, none for an empty slot, and that only SDLC siblings count as landing sources. This matches freshness.go and slots.go.

## Round 3 — 2026-10-16T19:24:08-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Already disposed in round 2; the window has no code change that reverses the minute-truncation fix.
- BR-2 — addressed — Already disposed in round 2; the missingSection constant and the conformance tests are unchanged in this window.
- BR-3 — addressed — Already disposed in round 2; the seam wiring is unchanged in this window.
- BR-4 — addressed — Already disposed in round 2; the plan Revisions entry stands and the window does not edit it.

## Open findings

(none — every finding has been disposed)
