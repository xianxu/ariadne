---
gate: boundary-review
issue: 400
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 13c0eaf4-0fae-4dff-930f-7f171f1c9c44
            repository: github.com/xianxu/ariadne
            issue_id: "000400"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a29bae60eb39ed26
            reviewed_head: 7d17e9481d46572fc438a665404b73397230ee43
            slot: ariadne:2
            requested_at: "2026-10-17T22:40:28.740704Z"
          phase: finished
          pid: 50590
          started_at: "2026-10-17T22:40:28.869159Z"
          finished_at: "2026-10-17T22:47:38.293164Z"
          session_id: 6a9844fa-865f-4865-a495-5134587059cd
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7d17e9481d46572fc438a665404b73397230ee43
        tracker_head: 5b6dcd4f78c4b9c3d10bff111bdc0a66f78fd38e
        branch_point: 78403958cf4a69d55a1fafdd76cc95b303419fa6
        since: "2026-10-17T15:01:16-07:00"
        until: "2026-10-17T15:14:47-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T15:01:16-07:00"
              end: "2026-10-17T15:14:47-07:00"
              began: "2026-10-17T15:01:16-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9dd01cf1858bb0eaee7bbec2c82522534322168a821e7d703175c89d8d020baa
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 79253313
              windows:
                - start: "2026-10-17T15:01:16-07:00"
                  end: "2026-10-17T15:14:47-07:00"
      "n": 1
      timestamp: "2026-10-17T15:47:48-07:00"
      dispatched: "2026-10-17T15:40:28-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Run lease stays held after stopSignals, so cancel during record/post/cleanup kills sdlc by default SIGTERM
          detail: localcheck_run.go calls stopSignals() right after the check exits but releases the lease only in the last defer; publishSuiteAttempt, GH.PostStatus/PRComment and worktree removal run with default SIGTERM while the lease is held, so a cancel there kills the process (merge path included), leaks the worktree and may half-record, contrary to the help text and the code comment. Release the lease where signal ownership ends (or keep SIGTERM owned until release) and add a post-check cancel test.
          family: lease-outlives-signal-ownership
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Stale leases are swept only by cancel, not by run as the design states
          detail: 'Design: a crashed run''s lease is swept by the next cancel or run. Only cancelLocalChecks sweeps, so leases and .tmp leftovers from SIGKILLed runs persist until someone runs cancel. Factor the sweep out and call it from holdRunLease/runLocalCheck.'
          family: artifact-removal-path-undelivered
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Spec/Design text still says local-checks/runs/ and process-group signalling with PGID
          detail: The code uses local-check-runs/ and SIGTERMs the run pid without recording a PGID; add a Revisions entry.
          family: plan-text-drift
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: runLocalCheckCancel (top-level resolution plus runs dir) has no test; the slot-release Done-when is not asserted
          family: cli-wiring-untested
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: fa56c63f-3b2c-44ae-acf8-fcb5445e8ee8
            repository: github.com/xianxu/ariadne
            issue_id: "000400"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/7606c4fc25496e37
            reviewed_head: 5575e0caf791d4f856c756e6304ef1ee8935f58c
            slot: ariadne:2
            requested_at: "2026-10-17T23:36:02.338205Z"
          phase: finished
          pid: 60911
          started_at: "2026-10-17T23:36:02.543868Z"
          finished_at: "2026-10-17T23:43:09.158091Z"
          session_id: 84330501-9939-4150-af12-4a88b2748797
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 5575e0caf791d4f856c756e6304ef1ee8935f58c
        tracker_head: 5b6dcd4f78c4b9c3d10bff111bdc0a66f78fd38e
        branch_point: 78403958cf4a69d55a1fafdd76cc95b303419fa6
        since: "2026-10-17T15:01:16-07:00"
        until: "2026-10-17T15:59:27-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T15:01:16-07:00"
              end: "2026-10-17T15:59:27-07:00"
              began: "2026-10-17T15:01:16-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 0f74ad76f17c6b83afcb90eabe94125a0a2702bf4f05cf21bf39e9634bf8333b
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 80033999
              windows:
                - start: "2026-10-17T15:01:16-07:00"
                  end: "2026-10-17T15:59:27-07:00"
        queue_waits:
            - attempt_id: 45d14dba8f50875257459913725b41fc
              slot: ariadne:2
              record_digest: 2580a62a6756ab8e87a1f4184177dc7473b127b744922afe0db934987b083189
              start: "2026-10-17T22:14:54.62581Z"
              end: "2026-10-17T22:28:48.383154Z"
              queue_seconds: 833.757344
              raw_seconds: 833.757344
              exact: true
        cuts:
            - dispatched: "2026-10-17T15:40:28-07:00"
              finalized: "2026-10-17T15:47:48-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-17T16:43:09-07:00"
      dispatched: "2026-10-17T16:36:02-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: releaseLease() now precedes stopSignals() (localcheck_run.go:165); reverting it makes TestRunLocalCheckLeaseLifetime fail with "leases held while recording".
          round: 2
        - id: BR-2
          disposition: addressed
          note: runLocalCheck calls sweepStaleRunLeases before holdRunLease; reverting it leaves dead.json and TestRunLocalCheckLeaseLifetime fails.
          round: 2
        - id: BR-3
          disposition: addressed
          note: The Revisions entry (2026-10-17) records local-check-runs/ and pid-only SIGTERM with no PGID, matching localCheckRunsDir and cancelLocalChecks.
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestLocalCheckCancelOnlyOwnRun now cancels B through runLocalCheckCancel after t.Chdir(rootB); cancel returning implies release() ran, since the lease now ends after it.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Important
          title: README update appears missing for the new `sdlc local-check cancel` subcommand
          detail: README.md's local-check paragraph (around lines 464-473) documents only `local-check run`; the new user-typed `cancel` verb, the way to stop your own run instead of killing a PID from ps, is absent. Add one sentence there.
          family: user-surface-docs-gate
          round: 2
        - kind: defect
          id: BR-6
          severity: Minor
          title: Cancel passes the shared-directory lease's pid to syscall.Kill without bounding it
          detail: rec.PID is read from a directory every slot shares; pid 0 or -1 would signal a process group or every process. The start-time and held-lock checks make this hard to reach, but reject PID <= 1 when parsing the lease (ARCH-SECURE).
          family: untrusted-lease-fields-unvalidated
          round: 2
        - kind: defect
          id: BR-7
          severity: Minor
          title: A cancel that saw the lock held can SIGTERM the run just after it dropped signal ownership
          detail: 'This is the inverse of BR-1, not a repeat: between cancel''s flock probe and Kill, the run can call releaseLease and stopSignals, so the default action kills sdlc during recording or merge. The window is microseconds. Name it in the Spec''s residual note beside PID reuse; no ordering seam tests it (ARCH-ORDER).'
          family: lease-outlives-signal-ownership
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 753827d3-50e9-413f-b335-46424d66c350
            repository: github.com/xianxu/ariadne
            issue_id: "000400"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/6b373cd19d584dc1
            reviewed_head: c36adf9fb0c7378c3922916dfe8506985e6b0ace
            slot: ariadne:2
            requested_at: "2026-10-18T00:12:24.950019Z"
          phase: finished
          pid: 60421
          started_at: "2026-10-18T00:12:25.188093Z"
          finished_at: "2026-10-18T00:18:01.539417Z"
          session_id: 7bd0709c-7ea3-45e0-8717-8da84e2a88ed
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: c36adf9fb0c7378c3922916dfe8506985e6b0ace
        tracker_head: 0c5129f280e946ccd9ea8d0e5800507f818d99c5
        branch_point: 78403958cf4a69d55a1fafdd76cc95b303419fa6
        since: "2026-10-17T15:01:16-07:00"
        until: "2026-10-17T16:43:37-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T15:01:16-07:00"
              end: "2026-10-17T16:43:37-07:00"
              began: "2026-10-17T15:01:16-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5fbcb792e00d2fc0681f6f15b9adf55b53c8dc9eec2f42e24012482806bc11dc
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 80824683
              windows:
                - start: "2026-10-17T15:01:16-07:00"
                  end: "2026-10-17T16:43:37-07:00"
        queue_waits:
            - attempt_id: 45d14dba8f50875257459913725b41fc
              slot: ariadne:2
              record_digest: 2580a62a6756ab8e87a1f4184177dc7473b127b744922afe0db934987b083189
              start: "2026-10-17T22:14:54.62581Z"
              end: "2026-10-17T22:28:48.383154Z"
              queue_seconds: 833.757344
              raw_seconds: 833.757344
              exact: true
            - attempt_id: a72bc511f13918e950ab783f84a09d20
              slot: ariadne:2
              record_digest: ccade0c939c2cec53d0cf7a7443c8063f4a665a3683d85896af9ed0d4d91d730
              start: "2026-10-17T22:59:34.24628Z"
              end: "2026-10-17T23:24:36.56333Z"
              queue_seconds: 1502.31705
              raw_seconds: 1502.31705
              exact: true
        cuts:
            - dispatched: "2026-10-17T15:40:28-07:00"
              finalized: "2026-10-17T15:47:48-07:00"
              finalizing: false
            - dispatched: "2026-10-17T16:36:02-07:00"
              finalized: "2026-10-17T16:43:09-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-17T17:18:02-07:00"
      dispatched: "2026-10-17T17:12:24-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: README.md:472-473 now names `sdlc local-check cancel` beside local-check run and warns against PIDs from ps.
          round: 3
        - id: BR-6
          disposition: not-addressed
          note: The guard exists (localcheck_lease.go:162), but with it disabled TestCancelChecksProcessStartTime still passes, because pids -1/0/1 fail the start-time check and both messages say "not signalled". Stub processStartTime to return rec.PIDStarted in that loop and assert the guard's own message.
          round: 3
        - id: BR-7
          disposition: addressed
          note: The issue's Design residual bullet now names the race where a cancel lands just after the run released its lease, beside PID reuse; prose-only, as the finding asked.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: c36adf9fb0c7378c3922916dfe8506985e6b0ace
      blocked: false
---

# Gate ledger — ariadne#400 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T15:47:48-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `lease-outlives-signal-ownership` (kind: defect) Run lease stays held after stopSignals, so cancel during record/post/cleanup kills sdlc by default SIGTERM
  localcheck_run.go calls stopSignals() right after the check exits but releases the lease only in the last defer; publishSuiteAttempt, GH.PostStatus/PRComment and worktree removal run with default SIGTERM while the lease is held, so a cancel there kills the process (merge path included), leaks the worktree and may half-record, contrary to the help text and the code comment. Release the lease where signal ownership ends (or keep SIGTERM owned until release) and add a post-check cancel test.
- **BR-2** [Important] `artifact-removal-path-undelivered` (kind: defect) Stale leases are swept only by cancel, not by run as the design states
  Design: a crashed run's lease is swept by the next cancel or run. Only cancelLocalChecks sweeps, so leases and .tmp leftovers from SIGKILLed runs persist until someone runs cancel. Factor the sweep out and call it from holdRunLease/runLocalCheck.
- **BR-3** [Minor] `plan-text-drift` (kind: defect) Spec/Design text still says local-checks/runs/ and process-group signalling with PGID
  The code uses local-check-runs/ and SIGTERMs the run pid without recording a PGID; add a Revisions entry.
- **BR-4** [Minor] `cli-wiring-untested` (kind: defect) runLocalCheckCancel (top-level resolution plus runs dir) has no test; the slot-release Done-when is not asserted

## Round 2 — 2026-10-17T16:43:09-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — releaseLease() now precedes stopSignals() (localcheck_run.go:165); reverting it makes TestRunLocalCheckLeaseLifetime fail with "leases held while recording".
- BR-2 — addressed — runLocalCheck calls sweepStaleRunLeases before holdRunLease; reverting it leaves dead.json and TestRunLocalCheckLeaseLifetime fails.
- BR-3 — addressed — The Revisions entry (2026-10-17) records local-check-runs/ and pid-only SIGTERM with no PGID, matching localCheckRunsDir and cancelLocalChecks.
- BR-4 — addressed — TestLocalCheckCancelOnlyOwnRun now cancels B through runLocalCheckCancel after t.Chdir(rootB); cancel returning implies release() ran, since the lease now ends after it.

### Raised

- **BR-5** [Important] `user-surface-docs-gate` (kind: defect) README update appears missing for the new `sdlc local-check cancel` subcommand
  README.md's local-check paragraph (around lines 464-473) documents only `local-check run`; the new user-typed `cancel` verb, the way to stop your own run instead of killing a PID from ps, is absent. Add one sentence there.
- **BR-6** [Minor] `untrusted-lease-fields-unvalidated` (kind: defect) Cancel passes the shared-directory lease's pid to syscall.Kill without bounding it
  rec.PID is read from a directory every slot shares; pid 0 or -1 would signal a process group or every process. The start-time and held-lock checks make this hard to reach, but reject PID <= 1 when parsing the lease (ARCH-SECURE).
- **BR-7** [Minor] `lease-outlives-signal-ownership` (kind: defect) A cancel that saw the lock held can SIGTERM the run just after it dropped signal ownership
  This is the inverse of BR-1, not a repeat: between cancel's flock probe and Kill, the run can call releaseLease and stopSignals, so the default action kills sdlc during recording or merge. The window is microseconds. Name it in the Spec's residual note beside PID reuse; no ordering seam tests it (ARCH-ORDER).

## Round 3 — 2026-10-17T17:18:02-07:00 (claude) — passed

### Disposed

- BR-5 — addressed — README.md:472-473 now names `sdlc local-check cancel` beside local-check run and warns against PIDs from ps.
- BR-6 — not-addressed — The guard exists (localcheck_lease.go:162), but with it disabled TestCancelChecksProcessStartTime still passes, because pids -1/0/1 fail the start-time check and both messages say "not signalled". Stub processStartTime to return rec.PIDStarted in that loop and assert the guard's own message.
- BR-7 — addressed — The issue's Design residual bullet now names the race where a cancel lands just after the run released its lease, beside PID reuse; prose-only, as the finding asked.

## Open findings

- **BR-6** [Minor] `untrusted-lease-fields-unvalidated` (kind: defect) Cancel passes the shared-directory lease's pid to syscall.Kill without bounding it
