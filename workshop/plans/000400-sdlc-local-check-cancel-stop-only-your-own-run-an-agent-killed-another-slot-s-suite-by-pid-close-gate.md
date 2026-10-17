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

## Open findings

- **BR-1** [Important] `lease-outlives-signal-ownership` (kind: defect) Run lease stays held after stopSignals, so cancel during record/post/cleanup kills sdlc by default SIGTERM
- **BR-2** [Important] `artifact-removal-path-undelivered` (kind: defect) Stale leases are swept only by cancel, not by run as the design states
- **BR-3** [Minor] `plan-text-drift` (kind: defect) Spec/Design text still says local-checks/runs/ and process-group signalling with PGID
- **BR-4** [Minor] `cli-wiring-untested` (kind: defect) runLocalCheckCancel (top-level resolution plus runs dir) has no test; the slot-release Done-when is not asserted
