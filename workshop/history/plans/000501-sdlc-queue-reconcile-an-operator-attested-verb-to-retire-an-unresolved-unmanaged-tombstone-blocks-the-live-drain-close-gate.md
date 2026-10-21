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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ff52f87a-1f6b-45af-8c29-ead031b83408
            repository: github.com/xianxu/ariadne
            issue_id: "000501"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/9ecd27334cbf5ccb
            reviewed_head: fe03d734f03caabf9bd30cf238d24e9dcc8f8008
            slot: ariadne:3
            requested_at: "2026-10-20T18:26:09.264032Z"
          phase: finished
          pid: 94922
          started_at: "2026-10-20T18:26:09.361557Z"
          finished_at: "2026-10-20T18:27:50.029833Z"
          session_id: 01a15a90-5063-7b11-be1a-62e724797d81
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: fe03d734f03caabf9bd30cf238d24e9dcc8f8008
        tracker_head: 6321f298e978d09443ef3a4331061a475f0358a9
        branch_point: cb8ffe354ad44da23b6adae5fbc0211cda4adb84
        since: "2026-10-20T10:56:04-07:00"
        until: "2026-10-20T11:25:09-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T10:56:04-07:00"
              end: "2026-10-20T11:25:09-07:00"
              began: "2026-10-20T10:56:04-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6764c7ef4291c5444d32969639ddd2dd1d007ca31a005da44714612a4291d118
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 88874354
              windows:
                - start: "2026-10-20T10:56:04-07:00"
                  end: "2026-10-20T11:25:09-07:00"
        cuts:
            - dispatched: "2026-10-20T11:18:30-07:00"
              finalized: "2026-10-20T11:19:48-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T11:27:51-07:00"
      dispatched: "2026-10-20T11:26:09-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:605–612 now documents the command, required flags, operator/:0 authority, descendant attestation, permanent audit and retry behavior, consistent with queue_reconcile.go and reconcile.go.
          round: 2
        - id: BR-2
          disposition: addressed
          note: The issue's active Done-when at lines 48–49 now specifies private executable recovery and read-only post-reboot observation. Revisions preserves the original live complete=true criterion as history, consistent with the approved acceptance change.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestReconcileIOFailures covers nine append and log/root/post-unlink sync failures, checking tombstone preservation, torn-log refusal and audit-preserving retry. All pass. Reverting the production fault-seam wiring through a temporary Go overlay makes all nine fail; the overlay was removed without modifying repository files.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: fe03d734f03caabf9bd30cf238d24e9dcc8f8008
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ce8b2607-1759-40c8-aa90-4b714c61d103
            repository: github.com/xianxu/ariadne
            issue_id: "000501"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/f166f79965ff8072
            reviewed_head: 904e2a9ffb758a43ee764c12b8dad823a8c351b6
            slot: ariadne:3
            requested_at: "2026-10-20T21:35:25.496379Z"
          phase: finished
          pid: 15015
          started_at: "2026-10-20T21:35:25.606211Z"
          finished_at: "2026-10-20T21:36:22.323527Z"
          session_id: 01a15b3d-98f5-7a51-b8d7-e3baab00f67d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 904e2a9ffb758a43ee764c12b8dad823a8c351b6
        tracker_head: 35d7fbc90bf6f04d50c40761276142b897a334c9
        branch_point: 24dc067b4af6884677ba3dcaf74e266b7ac16021
        since: "2026-10-20T10:56:04-07:00"
        until: "2026-10-20T13:43:15-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T10:56:04-07:00"
              end: "2026-10-20T13:43:15-07:00"
              began: "2026-10-20T10:56:04-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: f51b30db7b9406eeef55f1070851678dd52730878b39bb9b592c38ff3dc8d0f9
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 91839433
              windows:
                - start: "2026-10-20T10:56:04-07:00"
                  end: "2026-10-20T13:43:15-07:00"
        queue_waits:
            - attempt_id: ebe98520937f7852044ae44a3e40d04a
              slot: ariadne:3
              record_digest: 2f53a9f106640b6a470d0e1ec15f518642eff2ef7e2995dcb1f934de0f5d1d4a
              start: "2026-10-20T18:25:46.825424Z"
              end: "2026-10-20T19:11:39.665103Z"
              queue_seconds: 2752.839679
              raw_seconds: 2752.839679
              exact: true
        cuts:
            - dispatched: "2026-10-20T11:18:30-07:00"
              finalized: "2026-10-20T11:19:48-07:00"
              finalizing: false
            - dispatched: "2026-10-20T11:26:09-07:00"
              finalized: "2026-10-20T11:27:51-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-20T14:36:22-07:00"
      dispatched: "2026-10-20T14:35:25-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:605–612 retains the command, authority, attestation, audit and retry documentation.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Active Done-when retains approved private recovery and read-only live acceptance; the superseded criterion remains in Revisions. This window adds no acceptance change.
          round: 3
        - id: BR-3
          disposition: addressed
          note: The previously verified IO failure controls and production wiring are unchanged; all nine controls pass within the queuegate package run.
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: 904e2a9ffb758a43ee764c12b8dad823a8c351b6
      blocked: false
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

## Round 2 — 2026-10-20T11:27:51-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — README.md:605–612 now documents the command, required flags, operator/:0 authority, descendant attestation, permanent audit and retry behavior, consistent with queue_reconcile.go and reconcile.go.
- BR-2 — addressed — The issue's active Done-when at lines 48–49 now specifies private executable recovery and read-only post-reboot observation. Revisions preserves the original live complete=true criterion as history, consistent with the approved acceptance change.
- BR-3 — addressed — TestReconcileIOFailures covers nine append and log/root/post-unlink sync failures, checking tombstone preservation, torn-log refusal and audit-preserving retry. All pass. Reverting the production fault-seam wiring through a temporary Go overlay makes all nine fail; the overlay was removed without modifying repository files.

## Round 3 — 2026-10-20T14:36:22-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — README.md:605–612 retains the command, authority, attestation, audit and retry documentation.
- BR-2 — addressed — Active Done-when retains approved private recovery and read-only live acceptance; the superseded criterion remains in Revisions. This window adds no acceptance change.
- BR-3 — addressed — The previously verified IO failure controls and production wiring are unchanged; all nine controls pass within the queuegate package run.

## Open findings

(none — every finding has been disposed)
