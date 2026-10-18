---
gate: boundary-review
issue: 402
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 8557d2d9-08f3-4b05-9078-85445ca09e1b
            repository: github.com/xianxu/ariadne
            issue_id: "000402"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a2497d12bfb46a88
            reviewed_head: d8e59a282ee3fe622c792b7e341c48d2fc2a0f4f
            slot: ariadne:1
            requested_at: "2026-10-18T10:07:26.57511Z"
          phase: finished
          pid: 81507
          started_at: "2026-10-18T10:07:26.682458Z"
          finished_at: "2026-10-18T10:11:04.331967Z"
          session_id: 01a14e7b-02cf-7940-a0a6-9a698db10237
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: d8e59a282ee3fe622c792b7e341c48d2fc2a0f4f
        tracker_head: 829b7947f22644849f7dc8bbd03c61a827b19b4a
        branch_point: c1e63fe96e857efb9ef65c7135361fc2d7db21db
        since: "2026-10-17T21:36:12-07:00"
        until: "2026-10-18T03:06:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T21:36:12-07:00"
              end: "2026-10-17T21:39:45-07:00"
              began: "2026-10-17T21:36:12-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            - start: "2026-10-18T02:10:12-07:00"
              end: "2026-10-18T03:06:41-07:00"
              began: "2026-10-18T02:10:12-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 44cb914a25d2798ed93cafbab3375059bf2737d35454f59e358e6ec98eb86e95
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 94818390
              windows:
                - start: "2026-10-17T21:36:12-07:00"
                  end: "2026-10-17T21:39:45-07:00"
                - start: "2026-10-18T02:10:12-07:00"
                  end: "2026-10-18T03:06:41-07:00"
        queue_waits:
            - attempt_id: 4e02d8879bb372c94db5eb2b9dbc8a22
              slot: ariadne:1
              record_digest: b65a25ebb76bdd426621ac53b80ae800d79de22cc41657188b3e6d3678a6436b
              start: "2026-10-18T09:37:18.15096Z"
              end: "2026-10-18T09:57:38.487823Z"
              queue_seconds: 1220.336863
              raw_seconds: 1220.336863
              exact: true
      "n": 1
      timestamp: "2026-10-18T03:11:04-07:00"
      dispatched: "2026-10-18T03:07:26-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Later claimants of an already-blocked file remain falsely complete
          detail: cmd/sdlc/weeklyresource_collect.go:161–162 returns immediately for a blocked file without marking its newly matched family incomplete. A temporary regression fixture with two overlapping families under the first root and a third family under a later root reports the third family as Complete:true, Bytes:0, Files:0 despite its matching file. Mark every matching claimant unknown regardless of traversal order, and test permutations and additional claimants. This violates the declared overlap contract and ARCH-PURPOSE.
          family: overlap-invalidates-all-claimants
          round: 1
        - id: BR-2
          severity: Critical
          title: Advancing the observation clock breaks same-week publication retries
          detail: cmd/sdlc/weeklystats_cmd.go:140–150 recollects resources and inserts a fresh observed_at before calling the byte-comparing immutable writer. A regression probe with unchanged source fixtures and a one-second clock advance fails with “already exists with different bytes.” Preserve and return an existing validated publication on retry, including concurrent-publication recovery, without rewriting its observation time. The committed rerun test freezes the clock and misses this ARCH-ORDER contract violation.
          family: immutable-publication-retry
          round: 1
        - id: BR-3
          severity: Critical
          title: Invalid persisted measurements are accepted as complete comparison evidence
          detail: cmd/sdlc/weeklyresource_previous.go:60–63 checks section presence but does not validate measurement semantics. A regression probe supplying a complete family with bytes:-100 is accepted as complete; compareResourceBytes then permits it to produce a numeric delta. Validate persisted rows before comparison, including nonnegative measurements, required values and unique identities, and return named unknown coverage for invalid evidence. Add malformed-but-valid-JSON fixtures across the measurement types (ARCH-SECURE).
          family: persisted-measurement-validation
          round: 1
        - id: BR-4
          severity: Critical
          title: Five Core concepts entries do not exist at their declared locations
          detail: workshop/plans/000402-weekly-resource-trend-plan.md:28–32 names ResourceDelta, RenderResourceSection, ResourceCollector, PreviousResourceReader and WeeklySnapshotAttachment, but repository searches find none of these code entities. The implementation instead uses helpers including compareResourceBytes, renderWeeklyResourceSection, collectResourceDetails, readPreviousResourceSection and runWeeklySnapshot. Append a Revisions entry mapping the actual entities, paths and PURE/INTEGRATION classifications, and reconcile the table and descriptions. The required Core concepts cross-check classifies this contradiction as Critical (ARCH-PURPOSE).
          family: core-concepts-match-implementation
          round: 1
        - id: BR-5
          severity: Important
          title: Larger weekly artifacts have no explicit retention or permanent-growth contract
          detail: workshop/plans/000402-weekly-resource-trend-plan.md:40 delegates retention to the brain charter without identifying a removal routine, bound, or explicit permanent-retention cost. The new resources payload adds approximately 13.4 KB of formatted JSON to the inspected weekly artifact, while cmd/sdlc/weeklystats_write.go retains dated publications indefinitely. Name the applicable cleanup policy or explicitly document permanent retention, expected growth and its measurement location (ARCH-FUNERAL).
          family: durable-artifact-lifecycle
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 169a0263-61e7-4f0b-8fa6-adef344430a5
            repository: github.com/xianxu/ariadne
            issue_id: "000402"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/4cbcc6c7a059796a
            reviewed_head: 9ffb0ecdcc9faa408f39ee0edb7819de0b8594f6
            slot: ariadne:1
            requested_at: "2026-10-18T11:00:31.757132Z"
          phase: finished
          pid: 32612
          started_at: "2026-10-18T11:00:31.845585Z"
          finished_at: "2026-10-18T11:03:28.370063Z"
          session_id: 01a14eab-9ccb-7710-98e8-4767140526de
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 9ffb0ecdcc9faa408f39ee0edb7819de0b8594f6
        tracker_head: 829b7947f22644849f7dc8bbd03c61a827b19b4a
        branch_point: c1e63fe96e857efb9ef65c7135361fc2d7db21db
        since: "2026-10-17T21:36:12-07:00"
        until: "2026-10-18T03:24:44-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T21:36:12-07:00"
              end: "2026-10-17T21:39:45-07:00"
              began: "2026-10-17T21:36:12-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            - start: "2026-10-18T02:10:12-07:00"
              end: "2026-10-18T03:24:44-07:00"
              began: "2026-10-18T02:10:12-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 29c52c16b0c8871c8d606f4f03b27fccf25abf654dc73f929bbe70f3096be82f
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 96258357
              windows:
                - start: "2026-10-17T21:36:12-07:00"
                  end: "2026-10-17T21:39:45-07:00"
                - start: "2026-10-18T02:10:12-07:00"
                  end: "2026-10-18T03:24:44-07:00"
        queue_waits:
            - attempt_id: 4e02d8879bb372c94db5eb2b9dbc8a22
              slot: ariadne:1
              record_digest: b65a25ebb76bdd426621ac53b80ae800d79de22cc41657188b3e6d3678a6436b
              start: "2026-10-18T09:37:18.15096Z"
              end: "2026-10-18T09:57:38.487823Z"
              queue_seconds: 1220.336863
              raw_seconds: 1220.336863
              exact: true
        cuts:
            - dispatched: "2026-10-18T03:07:26-07:00"
              finalized: "2026-10-18T03:11:04-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T04:03:28-07:00"
      dispatched: "2026-10-18T04:00:31-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: weeklyresource_collect.go marks later blocked-file claimants incomplete. The committed regression passes at head and fails with the pre-fix collector restored; existing fixtures also cover earlier assignments and multiple claimants.
          round: 2
        - id: BR-2
          disposition: addressed
          note: weeklystats_cmd.go returns validated existing publications before collection and recovers a concurrent winner. Clock-advance and concurrent-publication tests pass at head and both fail with the pre-fix command restored.
          round: 2
        - id: BR-3
          disposition: addressed
          note: Both persisted readers call validatePersistedWeeklyResources. Malformed-measurement fixtures pass at head and fail with the pre-fix previous-week reader restored, including negative values, duplicate identities and invalid quantiles.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The pinned plan diff replaces the five nonexistent names with actual functions. Every Core concepts row now resolves to its declared file; the PURE functions have direct tests without IO. The first-boundary-review Revisions entry records the reconciliation.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Plan line 40 explicitly states permanent retention, no cleanup job, estimated incremental and total annual growth, and the measurement directory. weeklystats_write.go retains dated publications; the inspected pilot is 41,887 bytes, consistent with the documented total.
          round: 2
      findings:
        - kind: defect
          id: BR-6
          severity: Critical
          title: Device trends remain unknown permanently after a recoverable comparison failure
          detail: 'cmd/sdlc/weeklyresource_device.go:70–78 changes a valid current observation to unknown after an epoch change; the next week rejects that stored observation because its state is unknown, and repeats indefinitely. weeklyresource_machine.go:113–124 has the same failure for wear after a missing sample, while lines 100–107 propagate it after invalid prior evidence or an observation-time gap. Two disposable regression probes fail at head: boot2 values 10 then 30 never yield delta 20, and readable wear values 8 then 9 after a missing sample never yield delta 1. Preserve measurement validity separately from comparison status across both gauge types, and test recovery sequences through persisted weekly sections (ARCH-ORDER, ARCH-PURPOSE).'
          family: observation-comparison-state-separation
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 696e6e5e-722d-4487-acc5-715172e0a7a8
            repository: github.com/xianxu/ariadne
            issue_id: "000402"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/bf9d2c394e51393e
            reviewed_head: a8f715692083efa7fcedc984606430e858a65677
            slot: ariadne:1
            requested_at: "2026-10-18T11:32:08.364288Z"
          phase: finished
          pid: 85468
          started_at: "2026-10-18T11:32:08.456669Z"
          finished_at: "2026-10-18T11:33:55.055272Z"
          session_id: 01a14ec8-8d76-7e43-adb5-294ac7e4cb04
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: a8f715692083efa7fcedc984606430e858a65677
        tracker_head: 829b7947f22644849f7dc8bbd03c61a827b19b4a
        branch_point: c1e63fe96e857efb9ef65c7135361fc2d7db21db
        since: "2026-10-17T21:36:12-07:00"
        until: "2026-10-18T04:10:42-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T21:36:12-07:00"
              end: "2026-10-17T21:39:45-07:00"
              began: "2026-10-17T21:36:12-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            - start: "2026-10-18T02:10:12-07:00"
              end: "2026-10-18T04:10:42-07:00"
              began: "2026-10-18T02:10:12-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: fe7c8ccb7a4601217e2e0c3e4644048cc7901976989646f94ad3f5150dcc62bd
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 97634822
              windows:
                - start: "2026-10-17T21:36:12-07:00"
                  end: "2026-10-17T21:39:45-07:00"
                - start: "2026-10-18T02:10:12-07:00"
                  end: "2026-10-18T04:10:42-07:00"
        queue_waits:
            - attempt_id: 4e02d8879bb372c94db5eb2b9dbc8a22
              slot: ariadne:1
              record_digest: b65a25ebb76bdd426621ac53b80ae800d79de22cc41657188b3e6d3678a6436b
              start: "2026-10-18T09:37:18.15096Z"
              end: "2026-10-18T09:57:38.487823Z"
              queue_seconds: 1220.336863
              raw_seconds: 1220.336863
              exact: true
            - attempt_id: bcd876f87a239feb60716ada43bb379d
              slot: ariadne:1
              record_digest: 7b00147a4c5ed7d8634117118bb8aade9742f995d6304c3a8d0ded4fb2c6f898
              start: "2026-10-18T10:25:05.549855Z"
              end: "2026-10-18T10:51:54.501875Z"
              queue_seconds: 1608.95202
              raw_seconds: 1608.95202
              exact: true
        cuts:
            - dispatched: "2026-10-18T03:07:26-07:00"
              finalized: "2026-10-18T03:11:04-07:00"
              finalizing: false
            - dispatched: "2026-10-18T04:00:31-07:00"
              finalized: "2026-10-18T04:03:28-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-18T04:33:55-07:00"
      dispatched: "2026-10-18T04:32:08-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Blocked files invalidate later claimants; the three-claimant fixture passes in both root orders.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Exact-week validation precedes collection; advancing-clock retry and concurrent-publication fixtures pass.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Previous-week and existing-week readers share persisted-resource validation; missing-field and semantic-corruption fixtures pass.
          round: 3
        - id: BR-4
          disposition: addressed
          note: The plan's Core concepts table at lines 27–33 names existing entities at their declared paths; PURE functions contain no IO.
          round: 3
        - id: BR-5
          disposition: addressed
          note: Plan line 41 explicitly specifies permanent brain retention and annual growth costs; the inspected 41,939-byte pilot supports the stated total-size estimate.
          round: 3
        - id: BR-6
          disposition: addressed
          note: Device comparisons consume prior.Observed independently of comparison state. JSON-round-trip recovery tests pass at head and fail when the former prior-state guards are restored, reproducing lost write and wear deltas.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: a8f715692083efa7fcedc984606430e858a65677
      blocked: false
---

# Gate ledger — ariadne#402 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T03:11:04-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `overlap-invalidates-all-claimants` Later claimants of an already-blocked file remain falsely complete
  cmd/sdlc/weeklyresource_collect.go:161–162 returns immediately for a blocked file without marking its newly matched family incomplete. A temporary regression fixture with two overlapping families under the first root and a third family under a later root reports the third family as Complete:true, Bytes:0, Files:0 despite its matching file. Mark every matching claimant unknown regardless of traversal order, and test permutations and additional claimants. This violates the declared overlap contract and ARCH-PURPOSE.
- **BR-2** [Critical] `immutable-publication-retry` Advancing the observation clock breaks same-week publication retries
  cmd/sdlc/weeklystats_cmd.go:140–150 recollects resources and inserts a fresh observed_at before calling the byte-comparing immutable writer. A regression probe with unchanged source fixtures and a one-second clock advance fails with “already exists with different bytes.” Preserve and return an existing validated publication on retry, including concurrent-publication recovery, without rewriting its observation time. The committed rerun test freezes the clock and misses this ARCH-ORDER contract violation.
- **BR-3** [Critical] `persisted-measurement-validation` Invalid persisted measurements are accepted as complete comparison evidence
  cmd/sdlc/weeklyresource_previous.go:60–63 checks section presence but does not validate measurement semantics. A regression probe supplying a complete family with bytes:-100 is accepted as complete; compareResourceBytes then permits it to produce a numeric delta. Validate persisted rows before comparison, including nonnegative measurements, required values and unique identities, and return named unknown coverage for invalid evidence. Add malformed-but-valid-JSON fixtures across the measurement types (ARCH-SECURE).
- **BR-4** [Critical] `core-concepts-match-implementation` Five Core concepts entries do not exist at their declared locations
  workshop/plans/000402-weekly-resource-trend-plan.md:28–32 names ResourceDelta, RenderResourceSection, ResourceCollector, PreviousResourceReader and WeeklySnapshotAttachment, but repository searches find none of these code entities. The implementation instead uses helpers including compareResourceBytes, renderWeeklyResourceSection, collectResourceDetails, readPreviousResourceSection and runWeeklySnapshot. Append a Revisions entry mapping the actual entities, paths and PURE/INTEGRATION classifications, and reconcile the table and descriptions. The required Core concepts cross-check classifies this contradiction as Critical (ARCH-PURPOSE).
- **BR-5** [Important] `durable-artifact-lifecycle` Larger weekly artifacts have no explicit retention or permanent-growth contract
  workshop/plans/000402-weekly-resource-trend-plan.md:40 delegates retention to the brain charter without identifying a removal routine, bound, or explicit permanent-retention cost. The new resources payload adds approximately 13.4 KB of formatted JSON to the inspected weekly artifact, while cmd/sdlc/weeklystats_write.go retains dated publications indefinitely. Name the applicable cleanup policy or explicitly document permanent retention, expected growth and its measurement location (ARCH-FUNERAL).

## Round 2 — 2026-10-18T04:03:28-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — weeklyresource_collect.go marks later blocked-file claimants incomplete. The committed regression passes at head and fails with the pre-fix collector restored; existing fixtures also cover earlier assignments and multiple claimants.
- BR-2 — addressed — weeklystats_cmd.go returns validated existing publications before collection and recovers a concurrent winner. Clock-advance and concurrent-publication tests pass at head and both fail with the pre-fix command restored.
- BR-3 — addressed — Both persisted readers call validatePersistedWeeklyResources. Malformed-measurement fixtures pass at head and fail with the pre-fix previous-week reader restored, including negative values, duplicate identities and invalid quantiles.
- BR-4 — addressed — The pinned plan diff replaces the five nonexistent names with actual functions. Every Core concepts row now resolves to its declared file; the PURE functions have direct tests without IO. The first-boundary-review Revisions entry records the reconciliation.
- BR-5 — addressed — Plan line 40 explicitly states permanent retention, no cleanup job, estimated incremental and total annual growth, and the measurement directory. weeklystats_write.go retains dated publications; the inspected pilot is 41,887 bytes, consistent with the documented total.

### Raised

- **BR-6** [Critical] `observation-comparison-state-separation` (kind: defect) Device trends remain unknown permanently after a recoverable comparison failure
  cmd/sdlc/weeklyresource_device.go:70–78 changes a valid current observation to unknown after an epoch change; the next week rejects that stored observation because its state is unknown, and repeats indefinitely. weeklyresource_machine.go:113–124 has the same failure for wear after a missing sample, while lines 100–107 propagate it after invalid prior evidence or an observation-time gap. Two disposable regression probes fail at head: boot2 values 10 then 30 never yield delta 20, and readable wear values 8 then 9 after a missing sample never yield delta 1. Preserve measurement validity separately from comparison status across both gauge types, and test recovery sequences through persisted weekly sections (ARCH-ORDER, ARCH-PURPOSE).

## Round 3 — 2026-10-18T04:33:55-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Blocked files invalidate later claimants; the three-claimant fixture passes in both root orders.
- BR-2 — addressed — Exact-week validation precedes collection; advancing-clock retry and concurrent-publication fixtures pass.
- BR-3 — addressed — Previous-week and existing-week readers share persisted-resource validation; missing-field and semantic-corruption fixtures pass.
- BR-4 — addressed — The plan's Core concepts table at lines 27–33 names existing entities at their declared paths; PURE functions contain no IO.
- BR-5 — addressed — Plan line 41 explicitly specifies permanent brain retention and annual growth costs; the inspected 41,939-byte pilot supports the stated total-size estimate.
- BR-6 — addressed — Device comparisons consume prior.Observed independently of comparison state. JSON-round-trip recovery tests pass at head and fail when the former prior-state guards are restored, reproducing lost write and wear deltas.

## Open findings

(none — every finding has been disposed)
