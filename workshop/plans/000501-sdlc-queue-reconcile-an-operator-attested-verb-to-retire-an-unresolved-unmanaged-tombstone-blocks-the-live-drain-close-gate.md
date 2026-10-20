---
gate: boundary-review
issue: 501
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 6b321ae9-ec4c-45dc-a499-529d5c459382
            repository: github.com/xianxu/ariadne
            issue_id: "000501"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/044530cff625a2b0
            reviewed_head: 31cc33fcfefcee50db5dc2118afef03e98e7e7b4
            slot: ariadne:3
            requested_at: "2026-10-20T18:18:30.441583Z"
          phase: finished
          pid: 49315
          started_at: "2026-10-20T18:18:30.594054Z"
          finished_at: "2026-10-20T18:19:48.152653Z"
          session_id: 01a15a89-50a4-7bf2-b941-6422bb5e2e79
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 31cc33fcfefcee50db5dc2118afef03e98e7e7b4
        tracker_head: 6321f298e978d09443ef3a4331061a475f0358a9
        branch_point: cb8ffe354ad44da23b6adae5fbc0211cda4adb84
        since: "2026-10-20T10:56:04-07:00"
        until: "2026-10-20T11:17:12-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T10:56:04-07:00"
              end: "2026-10-20T11:17:12-07:00"
              began: "2026-10-20T10:56:04-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 5c66d1fae11b7e44bc38a4b1b099070fde25f858f83150c0df993dbcd3403bd4
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 88430487
              windows:
                - start: "2026-10-20T10:56:04-07:00"
                  end: "2026-10-20T11:17:12-07:00"
      "n": 1
      timestamp: "2026-10-20T11:19:48-07:00"
      dispatched: "2026-10-20T11:18:30-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Important
          title: README omits the new queue reconcile command
          detail: README.md:596 documents queue pause/resume/status, but this range introduces reconcile and its required flags without updating README. Add the operator/:0 invocation, descendant-attestation requirement, and audit/retry behavior; atlas and embedded help updates do not satisfy the separate README gate.
          family: user-surface-documentation
          round: 1
        - id: BR-2
          severity: Important
          title: Done-when still requires the superseded live-machine acceptance
          detail: workshop/issues/000501-sdlc-queue-reconcile-an-operator-attested-verb-to-retire-an-unresolved-unmanaged-tombstone-blocks-the-live-drain.md:48 requires live complete=true after reconciling the original tombstone. Its approved revision at line 77 replaces that acceptance after reboot. Update the active Done-when to the approved private-conformance and read-only observation criteria, preserving the original requirement as history.
          family: acceptance-criteria-consistency
          round: 1
        - id: BR-3
          severity: Important
          title: Checked-off append and sync failure controls are missing
          detail: The operative plan's completed crash/retry task claims append/sync error coverage, but reconcile_test.go does not exercise errors from log.Write, log.Sync, or either directory sync in reconcile.go:235-258. The crash hook runs only after successful audit and root sync. Add controllable failure tests proving pre-unlink failures preserve the tombstone and post-unlink sync failure retries without duplicating audit evidence (ARCH-ORDER).
          family: durability-failure-coverage
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#501 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T11:19:48-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-surface-documentation` README omits the new queue reconcile command
  README.md:596 documents queue pause/resume/status, but this range introduces reconcile and its required flags without updating README. Add the operator/:0 invocation, descendant-attestation requirement, and audit/retry behavior; atlas and embedded help updates do not satisfy the separate README gate.
- **BR-2** [Important] `acceptance-criteria-consistency` Done-when still requires the superseded live-machine acceptance
  workshop/issues/000501-sdlc-queue-reconcile-an-operator-attested-verb-to-retire-an-unresolved-unmanaged-tombstone-blocks-the-live-drain.md:48 requires live complete=true after reconciling the original tombstone. Its approved revision at line 77 replaces that acceptance after reboot. Update the active Done-when to the approved private-conformance and read-only observation criteria, preserving the original requirement as history.
- **BR-3** [Important] `durability-failure-coverage` Checked-off append and sync failure controls are missing
  The operative plan's completed crash/retry task claims append/sync error coverage, but reconcile_test.go does not exercise errors from log.Write, log.Sync, or either directory sync in reconcile.go:235-258. The crash hook runs only after successful audit and root sync. Add controllable failure tests proving pre-unlink failures preserve the tombstone and post-unlink sync failure retries without duplicating audit evidence (ARCH-ORDER).

## Open findings

- **BR-1** [Important] `user-surface-documentation` README omits the new queue reconcile command
- **BR-2** [Important] `acceptance-criteria-consistency` Done-when still requires the superseded live-machine acceptance
- **BR-3** [Important] `durability-failure-coverage` Checked-off append and sync failure controls are missing
