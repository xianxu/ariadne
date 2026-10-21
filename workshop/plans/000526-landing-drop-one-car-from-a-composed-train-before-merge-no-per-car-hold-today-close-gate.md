---
gate: boundary-review
issue: 526
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 82b3bd5e-cfe2-4e36-9dcd-4aa6bb778cff
            repository: github.com/xianxu/ariadne
            issue_id: "000526"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/7df8449c136c8f51
            reviewed_head: 8fb25eb263f95e8c8169e4db3bc2560f4be1c1f3
            slot: ariadne:1
            requested_at: "2026-10-21T20:54:36.203878Z"
          phase: finished
          pid: 28366
          started_at: "2026-10-21T20:54:36.304738Z"
          finished_at: "2026-10-21T20:56:57.182586Z"
          session_id: 01a1603e-952a-7120-aafc-4ee8b5a583f5
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 8fb25eb263f95e8c8169e4db3bc2560f4be1c1f3
        tracker_head: 13432893d98a4059d567961eaad962202d499dd3
        branch_point: 510aab6ff3a55bfdf5472e45eedf5512f6d04420
        since: "2026-10-21T08:01:01-07:00"
        until: "2026-10-21T13:54:15-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T08:01:01-07:00"
              end: "2026-10-21T13:54:15-07:00"
              began: "2026-10-21T08:01:01-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 61d68e2b81ec47d64e243343efc6a555785f03f5494691e694282463a963ad42
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 46762411
              windows:
                - start: "2026-10-21T08:01:01-07:00"
                  end: "2026-10-21T13:54:15-07:00"
            - harness: codex
              id: 01a15f85-cb74-7033-af90-14126531a090
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 768a588ce1896dd636e5d2d18d80cb7513b5c5a35646278df19730d8dbf14ecb
              started: "2026-10-21T17:32:46.776Z"
              spawned: false
              bytes: 507257
              windows:
                - start: "2026-10-21T17:32:46.776Z"
                  end: "2026-10-21T13:54:15-07:00"
            - harness: codex
              id: 01a15f88-5db9-7a80-aa27-fa67fe645fe8
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 0bd423b53ffa623055c1f34b29fb129cc58b15b27213bb744bf3d7593fbc5f42
              started: "2026-10-21T17:35:35.162Z"
              spawned: false
              bytes: 460405
              windows:
                - start: "2026-10-21T17:35:35.162Z"
                  end: "2026-10-21T13:54:15-07:00"
            - harness: codex
              id: 01a15f89-4991-7003-8c9b-63e011e604d2
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d652a7acc55f498550339d5a93d8c536bdd2efd51207136cebe03f716094d0cc
              started: "2026-10-21T17:36:35.552Z"
              spawned: false
              bytes: 485972
              windows:
                - start: "2026-10-21T17:36:35.552Z"
                  end: "2026-10-21T13:54:15-07:00"
            - harness: codex
              id: 01a15f89-c603-70b3-ac34-0bbcd830b55c
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 07f0cb41b74dbcb1e97f70c4a7c21b8357b6f021d1d9db88b71e305364dfdd7e
              started: "2026-10-21T17:37:07.414Z"
              spawned: false
              bytes: 430693
              windows:
                - start: "2026-10-21T17:37:07.414Z"
                  end: "2026-10-21T13:54:15-07:00"
        queue_waits:
            - attempt_id: fa26fb3440d7a414de6e81dd807549dc
              slot: ariadne:1
              record_digest: 213f7bbaa3505f4cf4b6b8b2b24a64acf22d40d339eceda3a97a8a07c39e38ea
              start: "2026-10-21T18:26:51.909956Z"
              end: "2026-10-21T18:54:22.589164Z"
              queue_seconds: 1650.679208
              raw_seconds: 1650.679208
              exact: true
            - attempt_id: 96a1c200914cff8507ec4063a674e08f
              slot: ariadne:1
              record_digest: a00b73d0e467f5f0e5bd5a1daf0291ef5782ef196cb7b9893059ca6bb71dc8e5
              start: "2026-10-21T19:10:42.370565Z"
              end: "2026-10-21T19:44:56.817153Z"
              queue_seconds: 2054.446588
              raw_seconds: 2054.446588
              exact: true
            - attempt_id: a385d410534da10539607a2bdf2dc495
              slot: ariadne:1
              record_digest: 512830d9a59f80a23de4e4b1679058154972dccc466cf283286b07fb568297db
              start: "2026-10-21T20:03:31.967728Z"
              end: "2026-10-21T20:41:58.4643Z"
              queue_seconds: 2306.496572
              raw_seconds: 2306.496572
              exact: true
      "n": 1
      timestamp: "2026-10-21T13:56:57-07:00"
      dispatched: "2026-10-21T13:54:36-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Accepted drops must fence non-merge completion updates
          detail: cmd/sdlc/landingtrain.go:176 checks pending drops before computing and publishing a completion, but UpdateTrain at cmd/sdlc/internal/landingline/traincontrol_io.go:178 commits membership changes without checking accepted drops under line.lock. A drop accepted between those operations can target a car removed by a bisect completion; the next ApplyTrainDrops then fails with “drop does not pin a unique riding car,” and the surviving conductor exits instead of automatically re-forming. A temporary Go-overlay regression reproduced this failure. Fence completion updates atomically against pending drops and route the pending result through applyTrainDrops using the pre-completion train; cover membership-changing and terminal completions. ARCH-ORDER, ARCH-PURPOSE.
          family: drop-preempts-completion
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 68bcc8aa-4a0c-42c4-9036-5e99a3a5ff23
            repository: github.com/xianxu/ariadne
            issue_id: "000526"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/361c208e96fa5db1
            reviewed_head: e48ce11ae824a0ece2a49fa32cf5613040027edf
            slot: ariadne:1
            requested_at: "2026-10-21T21:11:25.066043Z"
          phase: finished
          pid: 24809
          started_at: "2026-10-21T21:11:25.217993Z"
          finished_at: "2026-10-21T21:14:15.266406Z"
          session_id: 01a1604d-faa9-7013-b17f-bf0e4c2cf76a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: e48ce11ae824a0ece2a49fa32cf5613040027edf
        tracker_head: 3f3670ff3cae584b4e04d06728e2baa53d58aeff
        branch_point: 510aab6ff3a55bfdf5472e45eedf5512f6d04420
        since: "2026-10-21T08:01:01-07:00"
        until: "2026-10-21T14:08:57-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T08:01:01-07:00"
              end: "2026-10-21T14:08:57-07:00"
              began: "2026-10-21T08:01:01-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: b9998641459c8d6df04cdcc9529b96284cd4e8a7e8f4c99e917a0f4000b422ba
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 47564946
              windows:
                - start: "2026-10-21T08:01:01-07:00"
                  end: "2026-10-21T14:08:57-07:00"
            - harness: codex
              id: 01a15f85-cb74-7033-af90-14126531a090
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 768a588ce1896dd636e5d2d18d80cb7513b5c5a35646278df19730d8dbf14ecb
              started: "2026-10-21T17:32:46.776Z"
              spawned: false
              bytes: 507257
              windows:
                - start: "2026-10-21T17:32:46.776Z"
                  end: "2026-10-21T14:08:57-07:00"
            - harness: codex
              id: 01a15f88-5db9-7a80-aa27-fa67fe645fe8
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 0bd423b53ffa623055c1f34b29fb129cc58b15b27213bb744bf3d7593fbc5f42
              started: "2026-10-21T17:35:35.162Z"
              spawned: false
              bytes: 460405
              windows:
                - start: "2026-10-21T17:35:35.162Z"
                  end: "2026-10-21T14:08:57-07:00"
            - harness: codex
              id: 01a15f89-4991-7003-8c9b-63e011e604d2
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d652a7acc55f498550339d5a93d8c536bdd2efd51207136cebe03f716094d0cc
              started: "2026-10-21T17:36:35.552Z"
              spawned: false
              bytes: 485972
              windows:
                - start: "2026-10-21T17:36:35.552Z"
                  end: "2026-10-21T14:08:57-07:00"
            - harness: codex
              id: 01a15f89-c603-70b3-ac34-0bbcd830b55c
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 07f0cb41b74dbcb1e97f70c4a7c21b8357b6f021d1d9db88b71e305364dfdd7e
              started: "2026-10-21T17:37:07.414Z"
              spawned: false
              bytes: 430693
              windows:
                - start: "2026-10-21T17:37:07.414Z"
                  end: "2026-10-21T14:08:57-07:00"
        queue_waits:
            - attempt_id: fa26fb3440d7a414de6e81dd807549dc
              slot: ariadne:1
              record_digest: 213f7bbaa3505f4cf4b6b8b2b24a64acf22d40d339eceda3a97a8a07c39e38ea
              start: "2026-10-21T18:26:51.909956Z"
              end: "2026-10-21T18:54:22.589164Z"
              queue_seconds: 1650.679208
              raw_seconds: 1650.679208
              exact: true
            - attempt_id: 96a1c200914cff8507ec4063a674e08f
              slot: ariadne:1
              record_digest: a00b73d0e467f5f0e5bd5a1daf0291ef5782ef196cb7b9893059ca6bb71dc8e5
              start: "2026-10-21T19:10:42.370565Z"
              end: "2026-10-21T19:44:56.817153Z"
              queue_seconds: 2054.446588
              raw_seconds: 2054.446588
              exact: true
            - attempt_id: a385d410534da10539607a2bdf2dc495
              slot: ariadne:1
              record_digest: 512830d9a59f80a23de4e4b1679058154972dccc466cf283286b07fb568297db
              start: "2026-10-21T20:03:31.967728Z"
              end: "2026-10-21T20:41:58.4643Z"
              queue_seconds: 2306.496572
              raw_seconds: 2306.496572
              exact: true
        cuts:
            - dispatched: "2026-10-21T13:54:36-07:00"
              finalized: "2026-10-21T13:56:57-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-21T14:14:16-07:00"
      dispatched: "2026-10-21T14:11:25-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: UpdateTrain now checks pending drops under line.lock, and commitTrainCompletion routes pending results through ApplyTrainDrops using the original train. TestTrainCompletionAndDropOrdering and TestLandingDropPreemptsComputedCompletion pass at head. Removing the storage guard fails all three drop-first completion cases; removing the conductor routing fails bisect, terminal, phase and abort cases.
          round: 2
      findings:
        - kind: defect
          id: BR-2
          severity: Important
          title: Drop acceptance erases unknown cleanup publication fields
          detail: cmd/sdlc/internal/landingline/traincontrol.go:190 replaces Cleanup.Publication with Publication, discarding unknown fields stored only in the cleanup snapshot. A temporary overlay regression accepted one drop, persisted future_cleanup_proof under cleanup.publication, then accepted another member's drop; ReadTrainControl returned an empty Extra map. ARCH-EVOLVE requires nested unknown fields to survive every rewrite. Retain the existing pending cleanup snapshot, or preserve its unknown fields when refreshing known values, and add a regression through RequestTrainDrop with metadata unique to cleanup.publication.
          family: preserve-unknown-fields-on-rewrite
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 1457ca26-38d1-4d0a-ae81-766a81d34d8e
            repository: github.com/xianxu/ariadne
            issue_id: "000526"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/73a7754a859655e6
            reviewed_head: 2eb193c29df4550c63266ec065030a071a234d4b
            slot: ariadne:1
            requested_at: "2026-10-21T21:18:52.460537Z"
          phase: finished
          pid: 33747
          started_at: "2026-10-21T21:18:52.5517Z"
          finished_at: "2026-10-21T21:21:01.929383Z"
          session_id: 01a16054-cd8f-7aa3-84cd-286b03b5e797
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 2eb193c29df4550c63266ec065030a071a234d4b
        tracker_head: 3f3670ff3cae584b4e04d06728e2baa53d58aeff
        branch_point: 510aab6ff3a55bfdf5472e45eedf5512f6d04420
        since: "2026-10-21T08:01:01-07:00"
        until: "2026-10-21T14:18:11-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T08:01:01-07:00"
              end: "2026-10-21T14:18:11-07:00"
              began: "2026-10-21T08:01:01-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 73a57f988d457351a615c770f596dcb3e42563ae01100fc2fc50f0f57c67330a
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 48448517
              windows:
                - start: "2026-10-21T08:01:01-07:00"
                  end: "2026-10-21T14:18:11-07:00"
            - harness: codex
              id: 01a15f85-cb74-7033-af90-14126531a090
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 768a588ce1896dd636e5d2d18d80cb7513b5c5a35646278df19730d8dbf14ecb
              started: "2026-10-21T17:32:46.776Z"
              spawned: false
              bytes: 507257
              windows:
                - start: "2026-10-21T17:32:46.776Z"
                  end: "2026-10-21T14:18:11-07:00"
            - harness: codex
              id: 01a15f88-5db9-7a80-aa27-fa67fe645fe8
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 0bd423b53ffa623055c1f34b29fb129cc58b15b27213bb744bf3d7593fbc5f42
              started: "2026-10-21T17:35:35.162Z"
              spawned: false
              bytes: 460405
              windows:
                - start: "2026-10-21T17:35:35.162Z"
                  end: "2026-10-21T14:18:11-07:00"
            - harness: codex
              id: 01a15f89-4991-7003-8c9b-63e011e604d2
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d652a7acc55f498550339d5a93d8c536bdd2efd51207136cebe03f716094d0cc
              started: "2026-10-21T17:36:35.552Z"
              spawned: false
              bytes: 485972
              windows:
                - start: "2026-10-21T17:36:35.552Z"
                  end: "2026-10-21T14:18:11-07:00"
            - harness: codex
              id: 01a15f89-c603-70b3-ac34-0bbcd830b55c
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 07f0cb41b74dbcb1e97f70c4a7c21b8357b6f021d1d9db88b71e305364dfdd7e
              started: "2026-10-21T17:37:07.414Z"
              spawned: false
              bytes: 430693
              windows:
                - start: "2026-10-21T17:37:07.414Z"
                  end: "2026-10-21T14:18:11-07:00"
        queue_waits:
            - attempt_id: fa26fb3440d7a414de6e81dd807549dc
              slot: ariadne:1
              record_digest: 213f7bbaa3505f4cf4b6b8b2b24a64acf22d40d339eceda3a97a8a07c39e38ea
              start: "2026-10-21T18:26:51.909956Z"
              end: "2026-10-21T18:54:22.589164Z"
              queue_seconds: 1650.679208
              raw_seconds: 1650.679208
              exact: true
            - attempt_id: 96a1c200914cff8507ec4063a674e08f
              slot: ariadne:1
              record_digest: a00b73d0e467f5f0e5bd5a1daf0291ef5782ef196cb7b9893059ca6bb71dc8e5
              start: "2026-10-21T19:10:42.370565Z"
              end: "2026-10-21T19:44:56.817153Z"
              queue_seconds: 2054.446588
              raw_seconds: 2054.446588
              exact: true
            - attempt_id: a385d410534da10539607a2bdf2dc495
              slot: ariadne:1
              record_digest: 512830d9a59f80a23de4e4b1679058154972dccc466cf283286b07fb568297db
              start: "2026-10-21T20:03:31.967728Z"
              end: "2026-10-21T20:41:58.4643Z"
              queue_seconds: 2306.496572
              raw_seconds: 2306.496572
              exact: true
        cuts:
            - dispatched: "2026-10-21T13:54:36-07:00"
              finalized: "2026-10-21T13:56:57-07:00"
              finalizing: false
            - dispatched: "2026-10-21T14:11:25-07:00"
              finalized: "2026-10-21T14:14:16-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-21T14:21:08-07:00"
      dispatched: "2026-10-21T14:18:52-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-2
          disposition: addressed
          note: traincontrol.go:193 preserves cleanup.publication-only unknown fields. TestTrainDropRequestsPreserveCleanupPublicationUnknownFields passes through RequestTrainDrop, ApplyTrainDrops, and FinishTrainControl; removing the preservation assignment in a temporary Go overlay makes it fail with lost metadata.
          round: 3
        - id: BR-1
          disposition: addressed
          note: Completion and abort routing retain the authoritative riders until the locked commit. Ordering and shell outcome tests pass; removing UpdateTrain's pending-drop guard in a temporary overlay makes all three drop-first completion regressions fail.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: 2eb193c29df4550c63266ec065030a071a234d4b
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 3229c300-5da7-4724-9e20-d605409890d8
            repository: github.com/xianxu/ariadne
            issue_id: "000526"
            recipe: fix-review
            round: 4
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/052b471e7ebeda4f
            reviewed_head: a887eb11140108b188dd37dd2abaebdb8144a18a
            slot: ariadne:1
            requested_at: "2026-10-21T23:33:10.948882Z"
          phase: finished
          pid: 61994
          started_at: "2026-10-21T23:33:11.118153Z"
          finished_at: "2026-10-21T23:34:15.14084Z"
          session_id: 01a160cf-c4ef-7322-9dd8-02013617efb4
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: a887eb11140108b188dd37dd2abaebdb8144a18a
        tracker_head: b89f70ab9a379ffde4f8fbcf02ac6ab4c1e50c06
        branch_point: 2dc46703996c42f4b0916515d4572cf439dd2a32
        since: "2026-10-21T08:01:01-07:00"
        until: "2026-10-21T16:32:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T08:01:01-07:00"
              end: "2026-10-21T16:32:10-07:00"
              began: "2026-10-21T08:01:01-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 3c84d594f2a999df9d5c72f6e444c30a62b0bc82efd63c57cba31aced0030b4d
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 50008035
              windows:
                - start: "2026-10-21T08:01:01-07:00"
                  end: "2026-10-21T16:32:10-07:00"
            - harness: codex
              id: 01a15f85-cb74-7033-af90-14126531a090
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 768a588ce1896dd636e5d2d18d80cb7513b5c5a35646278df19730d8dbf14ecb
              started: "2026-10-21T17:32:46.776Z"
              spawned: false
              bytes: 507257
              windows:
                - start: "2026-10-21T17:32:46.776Z"
                  end: "2026-10-21T16:32:10-07:00"
            - harness: codex
              id: 01a15f88-5db9-7a80-aa27-fa67fe645fe8
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 0bd423b53ffa623055c1f34b29fb129cc58b15b27213bb744bf3d7593fbc5f42
              started: "2026-10-21T17:35:35.162Z"
              spawned: false
              bytes: 460405
              windows:
                - start: "2026-10-21T17:35:35.162Z"
                  end: "2026-10-21T16:32:10-07:00"
            - harness: codex
              id: 01a15f89-4991-7003-8c9b-63e011e604d2
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d652a7acc55f498550339d5a93d8c536bdd2efd51207136cebe03f716094d0cc
              started: "2026-10-21T17:36:35.552Z"
              spawned: false
              bytes: 485972
              windows:
                - start: "2026-10-21T17:36:35.552Z"
                  end: "2026-10-21T16:32:10-07:00"
            - harness: codex
              id: 01a15f89-c603-70b3-ac34-0bbcd830b55c
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 07f0cb41b74dbcb1e97f70c4a7c21b8357b6f021d1d9db88b71e305364dfdd7e
              started: "2026-10-21T17:37:07.414Z"
              spawned: false
              bytes: 430693
              windows:
                - start: "2026-10-21T17:37:07.414Z"
                  end: "2026-10-21T16:32:10-07:00"
        queue_waits:
            - attempt_id: fa26fb3440d7a414de6e81dd807549dc
              slot: ariadne:1
              record_digest: 213f7bbaa3505f4cf4b6b8b2b24a64acf22d40d339eceda3a97a8a07c39e38ea
              start: "2026-10-21T18:26:51.909956Z"
              end: "2026-10-21T18:54:22.589164Z"
              queue_seconds: 1650.679208
              raw_seconds: 1650.679208
              exact: true
            - attempt_id: 96a1c200914cff8507ec4063a674e08f
              slot: ariadne:1
              record_digest: a00b73d0e467f5f0e5bd5a1daf0291ef5782ef196cb7b9893059ca6bb71dc8e5
              start: "2026-10-21T19:10:42.370565Z"
              end: "2026-10-21T19:44:56.817153Z"
              queue_seconds: 2054.446588
              raw_seconds: 2054.446588
              exact: true
            - attempt_id: a385d410534da10539607a2bdf2dc495
              slot: ariadne:1
              record_digest: 512830d9a59f80a23de4e4b1679058154972dccc466cf283286b07fb568297db
              start: "2026-10-21T20:03:31.967728Z"
              end: "2026-10-21T20:41:58.4643Z"
              queue_seconds: 2306.496572
              raw_seconds: 2306.496572
              exact: true
            - attempt_id: 6eef33affad1de835765e03ea278664e
              slot: ariadne:1
              record_digest: 42de6b9f9b677faa62f1958c73ed54fe385cd66b1fed066495137427c4c56137
              start: "2026-10-21T21:18:24.197852Z"
              end: "2026-10-21T22:34:11.710494Z"
              queue_seconds: 4547.512642
              raw_seconds: 4547.512642
              exact: true
        cuts:
            - dispatched: "2026-10-21T13:54:36-07:00"
              finalized: "2026-10-21T13:56:57-07:00"
              finalizing: false
            - dispatched: "2026-10-21T14:11:25-07:00"
              finalized: "2026-10-21T14:14:16-07:00"
              finalizing: false
            - dispatched: "2026-10-21T14:18:52-07:00"
              finalized: "2026-10-21T14:21:08-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-21T16:34:16-07:00"
      dispatched: "2026-10-21T16:33:10-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Both completion/drop lock orders remain covered at SplitAssessed and attribution entry; targeted conductor tests verify dropped and surviving outcome identities.
          round: 4
        - id: BR-2
          disposition: addressed
          note: The request-driven nested-metadata preservation regression remains intact and passes through acceptance, ApplyTrainDrops and FinishTrainControl.
          round: 4
      recipe: fix-review
      isolation: fenced
      reviewed: a887eb11140108b188dd37dd2abaebdb8144a18a
      blocked: false
---

# Gate ledger — ariadne#526 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T13:56:57-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `drop-preempts-completion` (kind: defect) Accepted drops must fence non-merge completion updates
  cmd/sdlc/landingtrain.go:176 checks pending drops before computing and publishing a completion, but UpdateTrain at cmd/sdlc/internal/landingline/traincontrol_io.go:178 commits membership changes without checking accepted drops under line.lock. A drop accepted between those operations can target a car removed by a bisect completion; the next ApplyTrainDrops then fails with “drop does not pin a unique riding car,” and the surviving conductor exits instead of automatically re-forming. A temporary Go-overlay regression reproduced this failure. Fence completion updates atomically against pending drops and route the pending result through applyTrainDrops using the pre-completion train; cover membership-changing and terminal completions. ARCH-ORDER, ARCH-PURPOSE.

## Round 2 — 2026-10-21T14:14:16-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — UpdateTrain now checks pending drops under line.lock, and commitTrainCompletion routes pending results through ApplyTrainDrops using the original train. TestTrainCompletionAndDropOrdering and TestLandingDropPreemptsComputedCompletion pass at head. Removing the storage guard fails all three drop-first completion cases; removing the conductor routing fails bisect, terminal, phase and abort cases.

### Raised

- **BR-2** [Important] `preserve-unknown-fields-on-rewrite` (kind: defect) Drop acceptance erases unknown cleanup publication fields
  cmd/sdlc/internal/landingline/traincontrol.go:190 replaces Cleanup.Publication with Publication, discarding unknown fields stored only in the cleanup snapshot. A temporary overlay regression accepted one drop, persisted future_cleanup_proof under cleanup.publication, then accepted another member's drop; ReadTrainControl returned an empty Extra map. ARCH-EVOLVE requires nested unknown fields to survive every rewrite. Retain the existing pending cleanup snapshot, or preserve its unknown fields when refreshing known values, and add a regression through RequestTrainDrop with metadata unique to cleanup.publication.

## Round 3 — 2026-10-21T14:21:08-07:00 (codex) — passed

### Disposed

- BR-2 — addressed — traincontrol.go:193 preserves cleanup.publication-only unknown fields. TestTrainDropRequestsPreserveCleanupPublicationUnknownFields passes through RequestTrainDrop, ApplyTrainDrops, and FinishTrainControl; removing the preservation assignment in a temporary Go overlay makes it fail with lost metadata.
- BR-1 — addressed — Completion and abort routing retain the authoritative riders until the locked commit. Ordering and shell outcome tests pass; removing UpdateTrain's pending-drop guard in a temporary overlay makes all three drop-first completion regressions fail.

## Round 4 — 2026-10-21T16:34:16-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Both completion/drop lock orders remain covered at SplitAssessed and attribution entry; targeted conductor tests verify dropped and surviving outcome identities.
- BR-2 — addressed — The request-driven nested-metadata preservation regression remains intact and passes through acceptance, ApplyTrainDrops and FinishTrainControl.

## Open findings

(none — every finding has been disposed)
