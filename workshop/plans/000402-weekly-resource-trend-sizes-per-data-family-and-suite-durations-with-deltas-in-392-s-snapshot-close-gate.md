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

## Open findings

- **BR-1** [Critical] `overlap-invalidates-all-claimants` Later claimants of an already-blocked file remain falsely complete
- **BR-2** [Critical] `immutable-publication-retry` Advancing the observation clock breaks same-week publication retries
- **BR-3** [Critical] `persisted-measurement-validation` Invalid persisted measurements are accepted as complete comparison evidence
- **BR-4** [Critical] `core-concepts-match-implementation` Five Core concepts entries do not exist at their declared locations
- **BR-5** [Important] `durable-artifact-lifecycle` Larger weekly artifacts have no explicit retention or permanent-growth contract
