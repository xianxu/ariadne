---
gate: boundary-review
issue: 559
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f50cf827-dee8-4196-bedb-d0eede06b58b
            repository: github.com/xianxu/ariadne
            issue_id: "000559"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/46c59f09c10a22d1
            reviewed_head: 43ebab480fae890e82ce4523aca1699dd86de858
            slot: ariadne:5
            requested_at: "2026-10-23T15:29:33.805137Z"
          phase: finished
          pid: 11490
          started_at: "2026-10-23T15:29:33.909432Z"
          finished_at: "2026-10-23T15:33:07.086281Z"
          session_id: 6227478e-b7cb-4b6b-b305-10a1afe5c288
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 43ebab480fae890e82ce4523aca1699dd86de858
        tracker_head: 46658e66bfd451c34db4e53e7f2153c1f129010f
        branch_point: a561b0aba633d3103cbf9343c08117b8c0a1fab3
        since: "2026-10-23T06:42:13-07:00"
        until: "2026-10-23T08:29:11-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T06:42:13-07:00"
              end: "2026-10-23T08:29:11-07:00"
              began: "2026-10-23T06:42:13-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: d14f05f8a77b9faf1d27fa2a6f9811163827f59df33dbde1c569d643069bb81f
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 6629917
              windows:
                - start: "2026-10-23T06:42:13-07:00"
                  end: "2026-10-23T08:29:11-07:00"
        queue_waits:
            - attempt_id: 7fd768d2363d5a753323403277cf6b8b
              slot: ariadne:5
              record_digest: c166128bad2fd09d411fdfc1205b9541a04f83eefbf45e5ec2d6365ce66ef930
              start: "2026-10-23T14:30:32.940516Z"
              end: "2026-10-23T15:11:49.518838Z"
              queue_seconds: 2476.578322
              raw_seconds: 2476.578322
              exact: true
      "n": 1
      timestamp: "2026-10-23T08:33:07-07:00"
      dispatched: "2026-10-23T08:29:33-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: A light-lane overrun in a typed-runner suite errors out without retrying or recording, so the repo loops in the lane
          detail: 'bindTypedCheck sets StrictGroupCleanup. On the light deadline run.TimedOut is true, so the deferred finishResumeScratch (localcheck_resume.go:362) returns a non-nil runnerCleanupUnconfirmed and clears retirable on the shared runnerLifetime. runSuiteLanes then returns that error with no normal retry, no attempt and no lane_overrun_seconds, so laneEligibility stays true and every later run overruns again. Fix: exclude flakeContract/StrictGroupCleanup in lightLaneFor (localcheck_run.go:83), or confirm and retry the overrun, plus a regression test.'
          family: lane-overrun-strict-cleanup-path
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: The plan's overrun recording, capability probe and eligibility name no longer match the code; no Revisions entry
          detail: The plan says a separate attempt is recorded with outcome lane_overrun, the helper advertises lane-v1, and the function is suiteLaneEligibility. The code puts lane_overrun_seconds on the retry's attempt, probes for --lane in acquire --help, and names the function laneEligibility. The atlas matches the code; the plan needs a dated Revisions entry.
          family: plan-code-drift
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: The overrun length is dropped when the normal retry errors, and the light phase's queue wait and log are not kept
          detail: runSuiteLanes returns the error with the overrun only in memory (suitelane.go:631). A paused or interrupted retry leaves the repo eligible. os.Create truncates the light run's log on retry.
          family: lane-overrun-evidence-loss
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: 'Plan-promised tests missing: dead light waiter pruning, and the run loop exercised against the real helper'
          detail: The shell test checks ticket cleanup only after a normal release. TestRunLocalCheckLight* uses a stateless argv-logging fake helper rather than the real helper the plan names.
          family: plan-test-commitments
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 466aeaa1-5dc4-43e3-b9c9-70c83812d9ff
            repository: github.com/xianxu/ariadne
            issue_id: "000559"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/85eec7f191200bb0
            reviewed_head: c021d115ae1ce19f3afa780de8c71b9659cdfe16
            slot: ariadne:5
            requested_at: "2026-10-23T15:37:52.939971Z"
          phase: finished
          pid: 6299
          started_at: "2026-10-23T15:37:53.044573Z"
          finished_at: "2026-10-23T15:51:05.257443Z"
          session_id: e8d1a633-db38-48d2-bc9e-e6178e039fa5
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: c021d115ae1ce19f3afa780de8c71b9659cdfe16
        tracker_head: 46658e66bfd451c34db4e53e7f2153c1f129010f
        branch_point: a561b0aba633d3103cbf9343c08117b8c0a1fab3
        since: "2026-10-23T06:42:13-07:00"
        until: "2026-10-23T08:37:30-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T06:42:13-07:00"
              end: "2026-10-23T08:37:30-07:00"
              began: "2026-10-23T06:42:13-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: cd6ad6a4251a86389cb001ca213aae9706b1630350170b631d4a4246270b9112
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 7045091
              windows:
                - start: "2026-10-23T06:42:13-07:00"
                  end: "2026-10-23T08:37:30-07:00"
        queue_waits:
            - attempt_id: 7fd768d2363d5a753323403277cf6b8b
              slot: ariadne:5
              record_digest: c166128bad2fd09d411fdfc1205b9541a04f83eefbf45e5ec2d6365ce66ef930
              start: "2026-10-23T14:30:32.940516Z"
              end: "2026-10-23T15:11:49.518838Z"
              queue_seconds: 2476.578322
              raw_seconds: 2476.578322
              exact: true
        cuts:
            - dispatched: "2026-10-23T08:29:33-07:00"
              finalized: "2026-10-23T08:33:07-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T08:51:06-07:00"
      dispatched: "2026-10-23T08:37:52-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: lightLaneFor excludes flakeContract suites (localcheck_run.go:167); TestLightLaneForSkipsTypedRunnerSuites goes red with the guard removed and green with it.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Plan Revisions 2026-10-23 round-1 entry records lane_overrun_seconds, the --lane probe and laneEligibility; the atlas matches the code.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: The overrun is now named in the retry error, but the kept .light log is deleted at publish and its test fails; see the new Critical finding.
          round: 2
        - id: BR-4
          disposition: addressed
          note: The dead light waiter pruning shell case and the real-helper runLocalCheck subtest were added; both pass.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: The kept light-run log is deleted on publish, so TestRunLocalCheckLightOverrunRetriesInNormalLine fails at head
          detail: 'This is the 2nd finding in family lane-overrun-evidence-loss. localcheck_run.go:309 renames the light log to <writer.LogPath>.light inside the inflight attempt directory. publishSuiteAttempt (teststats_store.go:420-432) copies only the main log to local-checks/<sha>/suite.log and removes that directory, so the .light file is lost. The test fails every time at suitelane_run_test.go:118 ("the light run''s log was not kept: []"), and the plan''s Revisions entry claims the log is kept. Rule to fix rather than this instance: every piece of light-phase evidence (overrun length, light log, light queue wait) must travel through the attempt publish path (the SuiteAttempt record or a file publishSuiteAttempt places next to the published log), never only the inflight directory or memory. Prevalence so far: 3 sites (log truncated on retry, now deleted on publish; overrun length only in the error text; light queue wait dropped).'
          family: lane-overrun-evidence-loss
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 72c144d3-0970-4369-a5bb-fce00c285ac6
            repository: github.com/xianxu/ariadne
            issue_id: "000559"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/df7d7a9c630e4adc
            reviewed_head: 547c6fa0e5bb7b2ede22caa970feb3dfde405be2
            slot: ariadne:5
            requested_at: "2026-10-23T16:00:44.821418Z"
          phase: finished
          pid: 11795
          started_at: "2026-10-23T16:00:44.984053Z"
          finished_at: "2026-10-23T16:04:23.434052Z"
          session_id: b54a5d1c-9b71-4761-b877-3ae28da37076
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 547c6fa0e5bb7b2ede22caa970feb3dfde405be2
        tracker_head: 580d7306fcb2a8b844eea332191c567aa390acf2
        branch_point: a561b0aba633d3103cbf9343c08117b8c0a1fab3
        since: "2026-10-23T06:42:13-07:00"
        until: "2026-10-23T08:52:56-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T06:42:13-07:00"
              end: "2026-10-23T08:52:56-07:00"
              began: "2026-10-23T06:42:13-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: a2461c9d8e211edcf9fddd3b3b4066d40fe5fc0c8c46a727e2cef5995cbfd38b
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 7277761
              windows:
                - start: "2026-10-23T06:42:13-07:00"
                  end: "2026-10-23T08:52:56-07:00"
        queue_waits:
            - attempt_id: 7fd768d2363d5a753323403277cf6b8b
              slot: ariadne:5
              record_digest: c166128bad2fd09d411fdfc1205b9541a04f83eefbf45e5ec2d6365ce66ef930
              start: "2026-10-23T14:30:32.940516Z"
              end: "2026-10-23T15:11:49.518838Z"
              queue_seconds: 2476.578322
              raw_seconds: 2476.578322
              exact: true
        cuts:
            - dispatched: "2026-10-23T08:29:33-07:00"
              finalized: "2026-10-23T08:33:07-07:00"
              finalizing: false
            - dispatched: "2026-10-23T08:37:52-07:00"
              finalized: "2026-10-23T08:51:06-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-23T09:04:24-07:00"
      dispatched: "2026-10-23T09:00:44-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-5
          disposition: addressed
          note: The .light rename is gone. The partial log is discarded on purpose, and the plan Revisions and atlas say so. TestRunLocalCheckLightOverrunRetriesInNormalLine passes at head. Reverting suitelane.go:146-149 turns it and TestRunSuiteLanesQueueTimeCoversTheLightPhase red.
          round: 3
        - id: BR-3
          disposition: not-addressed
          note: 'The light queue wait (now in the queue time) and the truncated log (now discarded on purpose) are resolved. One part remains: when the normal retry errors, the overrun is still only in the error text (suitelane.go:131-135). The cost is bounded at one wasted 240s light run. Minor, non-blocking.'
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: 547c6fa0e5bb7b2ede22caa970feb3dfde405be2
      blocked: false
---

# Gate ledger — ariadne#559 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T08:33:07-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `lane-overrun-strict-cleanup-path` (kind: defect) A light-lane overrun in a typed-runner suite errors out without retrying or recording, so the repo loops in the lane
  bindTypedCheck sets StrictGroupCleanup. On the light deadline run.TimedOut is true, so the deferred finishResumeScratch (localcheck_resume.go:362) returns a non-nil runnerCleanupUnconfirmed and clears retirable on the shared runnerLifetime. runSuiteLanes then returns that error with no normal retry, no attempt and no lane_overrun_seconds, so laneEligibility stays true and every later run overruns again. Fix: exclude flakeContract/StrictGroupCleanup in lightLaneFor (localcheck_run.go:83), or confirm and retry the overrun, plus a regression test.
- **BR-2** [Important] `plan-code-drift` (kind: defect) The plan's overrun recording, capability probe and eligibility name no longer match the code; no Revisions entry
  The plan says a separate attempt is recorded with outcome lane_overrun, the helper advertises lane-v1, and the function is suiteLaneEligibility. The code puts lane_overrun_seconds on the retry's attempt, probes for --lane in acquire --help, and names the function laneEligibility. The atlas matches the code; the plan needs a dated Revisions entry.
- **BR-3** [Minor] `lane-overrun-evidence-loss` (kind: defect) The overrun length is dropped when the normal retry errors, and the light phase's queue wait and log are not kept
  runSuiteLanes returns the error with the overrun only in memory (suitelane.go:631). A paused or interrupted retry leaves the repo eligible. os.Create truncates the light run's log on retry.
- **BR-4** [Minor] `plan-test-commitments` (kind: defect) Plan-promised tests missing: dead light waiter pruning, and the run loop exercised against the real helper
  The shell test checks ticket cleanup only after a normal release. TestRunLocalCheckLight* uses a stateless argv-logging fake helper rather than the real helper the plan names.

## Round 2 — 2026-10-23T08:51:06-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — lightLaneFor excludes flakeContract suites (localcheck_run.go:167); TestLightLaneForSkipsTypedRunnerSuites goes red with the guard removed and green with it.
- BR-2 — addressed — Plan Revisions 2026-10-23 round-1 entry records lane_overrun_seconds, the --lane probe and laneEligibility; the atlas matches the code.
- BR-3 — not-addressed — The overrun is now named in the retry error, but the kept .light log is deleted at publish and its test fails; see the new Critical finding.
- BR-4 — addressed — The dead light waiter pruning shell case and the real-helper runLocalCheck subtest were added; both pass.

### Raised

- **BR-5** [Critical] `lane-overrun-evidence-loss` (kind: defect) The kept light-run log is deleted on publish, so TestRunLocalCheckLightOverrunRetriesInNormalLine fails at head
  This is the 2nd finding in family lane-overrun-evidence-loss. localcheck_run.go:309 renames the light log to <writer.LogPath>.light inside the inflight attempt directory. publishSuiteAttempt (teststats_store.go:420-432) copies only the main log to local-checks/<sha>/suite.log and removes that directory, so the .light file is lost. The test fails every time at suitelane_run_test.go:118 ("the light run's log was not kept: []"), and the plan's Revisions entry claims the log is kept. Rule to fix rather than this instance: every piece of light-phase evidence (overrun length, light log, light queue wait) must travel through the attempt publish path (the SuiteAttempt record or a file publishSuiteAttempt places next to the published log), never only the inflight directory or memory. Prevalence so far: 3 sites (log truncated on retry, now deleted on publish; overrun length only in the error text; light queue wait dropped).

## Round 3 — 2026-10-23T09:04:24-07:00 (claude) — passed

### Disposed

- BR-5 — addressed — The .light rename is gone. The partial log is discarded on purpose, and the plan Revisions and atlas say so. TestRunLocalCheckLightOverrunRetriesInNormalLine passes at head. Reverting suitelane.go:146-149 turns it and TestRunSuiteLanesQueueTimeCoversTheLightPhase red.
- BR-3 — not-addressed — The light queue wait (now in the queue time) and the truncated log (now discarded on purpose) are resolved. One part remains: when the normal retry errors, the overrun is still only in the error text (suitelane.go:131-135). The cost is bounded at one wasted 240s light run. Minor, non-blocking.

## Open findings

- **BR-3** [Minor] `lane-overrun-evidence-loss` (kind: defect) The overrun length is dropped when the normal retry errors, and the light phase's queue wait and log are not kept
