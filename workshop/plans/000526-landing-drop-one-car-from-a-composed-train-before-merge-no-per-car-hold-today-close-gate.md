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
---

# Gate ledger — ariadne#526 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T13:56:57-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `drop-preempts-completion` (kind: defect) Accepted drops must fence non-merge completion updates
  cmd/sdlc/landingtrain.go:176 checks pending drops before computing and publishing a completion, but UpdateTrain at cmd/sdlc/internal/landingline/traincontrol_io.go:178 commits membership changes without checking accepted drops under line.lock. A drop accepted between those operations can target a car removed by a bisect completion; the next ApplyTrainDrops then fails with “drop does not pin a unique riding car,” and the surviving conductor exits instead of automatically re-forming. A temporary Go-overlay regression reproduced this failure. Fence completion updates atomically against pending drops and route the pending result through applyTrainDrops using the pre-completion train; cover membership-changing and terminal completions. ARCH-ORDER, ARCH-PURPOSE.

## Open findings

- **BR-1** [Critical] `drop-preempts-completion` (kind: defect) Accepted drops must fence non-merge completion updates
