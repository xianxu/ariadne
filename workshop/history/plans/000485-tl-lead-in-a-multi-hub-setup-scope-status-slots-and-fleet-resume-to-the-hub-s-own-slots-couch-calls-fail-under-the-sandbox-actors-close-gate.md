---
gate: boundary-review
issue: 485
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: dbb0e422-cd8e-4d0a-9599-53970afb6e4d
            repository: github.com/xianxu/ariadne
            issue_id: "000485"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/842400f973c425ef
            reviewed_head: 5bba914916306320fbf216b85f44d6d2d7789970
            slot: ariadne:2
            requested_at: "2026-10-20T20:59:53.024685Z"
          phase: finished
          pid: 30645
          started_at: "2026-10-20T20:59:53.116926Z"
          finished_at: "2026-10-20T21:01:56.136081Z"
          session_id: 01a15b1d-0e93-7ac2-bdcf-91cc56375674
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 5bba914916306320fbf216b85f44d6d2d7789970
        tracker_head: f622feef7956f9400e2444a4b30881b03a5ece50
        branch_point: 24dc067b4af6884677ba3dcaf74e266b7ac16021
        since: "2026-10-20T11:08:14-07:00"
        until: "2026-10-20T13:57:37-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T11:08:14-07:00"
              end: "2026-10-20T13:57:37-07:00"
              began: "2026-10-20T11:08:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 64dfd271f821e69ee64459c888d7f23e61ea1d007f38ada57e52683cb3b63b49
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 7126654
              windows:
                - start: "2026-10-20T11:08:14-07:00"
                  end: "2026-10-20T13:57:37-07:00"
            - harness: codex
              id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: b20d592925805f75a89d42e8d31e1e1e42b08182f35b1814085367d87b8a2dec
              started: "2026-10-20T18:21:54.178Z"
              spawned: false
              bytes: 4203995
              windows:
                - start: "2026-10-20T18:21:54.178Z"
                  end: "2026-10-20T13:57:37-07:00"
            - harness: codex
              id: 01a15b0f-1fb0-7c61-8192-ad8d7feab469
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f6143bd731812dbc31b21a9248391f25910c3ce19d4cced28dd1a07f7cc39b24
              started: "2026-10-20T20:44:40.563Z"
              spawned: false
              bytes: 710729
              windows:
                - start: "2026-10-20T20:44:40.563Z"
                  end: "2026-10-20T13:57:37-07:00"
            - harness: codex
              id: 01a15b11-3e3c-7b13-9638-200eca31f132
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3c5b6458933de9d0999aeb456c8a04edc004eb22d5275d8f6d3164ca261d6e71
              started: "2026-10-20T20:46:59.473Z"
              spawned: false
              bytes: 527719
              windows:
                - start: "2026-10-20T20:46:59.473Z"
                  end: "2026-10-20T13:57:37-07:00"
            - harness: codex
              id: 01a15b12-b635-70e3-ba3e-bda0960cd786
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 603f3e663280b80917eb29a2a8825ea99102d046f4180b5e42f4ef3fbd4fe049
              started: "2026-10-20T20:48:35.847Z"
              spawned: false
              bytes: 473883
              windows:
                - start: "2026-10-20T20:48:35.847Z"
                  end: "2026-10-20T13:57:37-07:00"
            - harness: codex
              id: 01a15b13-34ce-76e0-9c60-7c28afe426bd
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 35d96d5f6e1b94e4aebaac5eebecb5733f05b9d921992a4b769da44612cbc692
              started: "2026-10-20T20:49:08.201Z"
              spawned: false
              bytes: 410752
              windows:
                - start: "2026-10-20T20:49:08.201Z"
                  end: "2026-10-20T13:57:37-07:00"
            - harness: codex
              id: 01a15b14-a455-73d3-b6eb-747fa5c51f2e
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              parent_id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              digest: 72185cab0393d8eaa2a7687a8a68c779d640195643b16efcbedca5fff17ee0b9
              started: "2026-10-20T20:50:42.072Z"
              spawned: true
              bytes: 660546
              windows:
                - start: "2026-10-20T20:50:42.072Z"
                  end: "2026-10-20T13:57:37-07:00"
        queue_waits:
            - attempt_id: 81f3e3a5eb5fca51613b7f244d34f648
              slot: ariadne:2
              record_digest: dc35c572ea957b5aa408af5805405251f9de0acc4e40870856e0cff460b04a3d
              start: "2026-10-20T20:04:33.83285Z"
              end: "2026-10-20T20:26:00.637372Z"
              queue_seconds: 1286.804522
              raw_seconds: 1286.804522
              exact: true
      "n": 1
      timestamp: "2026-10-20T14:01:56-07:00"
      dispatched: "2026-10-20T13:59:53-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Bound actor observation so an uncertain resume can finish recovery
          detail: cmd/tl-lead/fleet.go:268 synchronously calls readActors before printing recovery guidance, naming untouched slots, or returning nonzero. Production execActors.Actors uses exec.Command(...).Run without a timeout (actors.go:27–30), so a hung Couch probe after an uncertain resume blocks that required terminal result indefinitely. Under ARCH-CONSTRAINTS/ARCH-ORDER, bound and terminate the probe, print the stopped/untouched result before optional observation, and add a regression with an initially successful actors read followed by a stalled recheck.
          family: recovery-probes-must-be-bounded
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Resolve the catalog failures introduced by modeled-delay tests
          detail: TestTrackedGoFilesPassTheCatalog fails on cmd/tl-lead/resume_result_test.go:93,98,108 with test-fixed-sleep. These sleeps advance synctest virtual time rather than guess wall-clock ordering; preserve that coverage and add the repository-supported, reasoned lint:allow annotations on the flagged lines, then rerun the catalog check.
          family: tracked-source-checks-must-pass
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: README update appears missing for both new --all flags
          detail: cmd/tl-lead/status.go:384 and fleet.go:307 introduce user-facing --all flags, but README.md is unchanged in the pinned range. Add concise usage documentation covering hub-scoped defaults and visibility-only expansion for fleet resume, with a link to the updated atlas details, as required by this boundary's README gate.
          family: user-surface-requires-readme-documentation
          round: 1
        - kind: defect
          id: BR-4
          severity: Important
          title: Add conformance coverage for Couch terminal resume receipts
          detail: cmd/tl-lead/resume_result.go:29–35 now depends on receipt fields, terminal status values, and their agreement with process exit. Tests supply hand-authored receipts; the existing TestResumeCandidatesLive checks only candidates, and conformance_test.go covers other Couch surfaces. Under ARCH-MOCK, add an opt-in check against a supported Couch binary in an isolated test configuration and document its cadence, so protocol drift cannot leave all fake-backed tests green.
          family: external-protocol-requires-conformance
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 97d7a882-902f-443c-9f7e-d3ea91ccc2a8
            repository: github.com/xianxu/ariadne
            issue_id: "000485"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/b9dedea227c4a119
            reviewed_head: e2d3af977ec4ae3d9e2758dc6f01cb59596a5b65
            slot: ariadne:2
            requested_at: "2026-10-20T21:07:29.696605Z"
          phase: finished
          pid: 13186
          started_at: "2026-10-20T21:07:29.816371Z"
          finished_at: "2026-10-20T21:09:24.021323Z"
          session_id: 01a15b24-06a8-7b01-bfd1-46bf8a7ef92c
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: e2d3af977ec4ae3d9e2758dc6f01cb59596a5b65
        tracker_head: f622feef7956f9400e2444a4b30881b03a5ece50
        branch_point: 24dc067b4af6884677ba3dcaf74e266b7ac16021
        since: "2026-10-20T11:08:14-07:00"
        until: "2026-10-20T14:06:48-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T11:08:14-07:00"
              end: "2026-10-20T14:06:48-07:00"
              began: "2026-10-20T11:08:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 64dfd271f821e69ee64459c888d7f23e61ea1d007f38ada57e52683cb3b63b49
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 7126654
              windows:
                - start: "2026-10-20T11:08:14-07:00"
                  end: "2026-10-20T14:06:48-07:00"
            - harness: codex
              id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c555be022376b1fe0f987bee93019232b5f32f10987644150fee464a303d95c9
              started: "2026-10-20T18:21:54.178Z"
              spawned: false
              bytes: 4650289
              windows:
                - start: "2026-10-20T18:21:54.178Z"
                  end: "2026-10-20T14:06:48-07:00"
            - harness: codex
              id: 01a15b0f-1fb0-7c61-8192-ad8d7feab469
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f6143bd731812dbc31b21a9248391f25910c3ce19d4cced28dd1a07f7cc39b24
              started: "2026-10-20T20:44:40.563Z"
              spawned: false
              bytes: 710729
              windows:
                - start: "2026-10-20T20:44:40.563Z"
                  end: "2026-10-20T14:06:48-07:00"
            - harness: codex
              id: 01a15b11-3e3c-7b13-9638-200eca31f132
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3c5b6458933de9d0999aeb456c8a04edc004eb22d5275d8f6d3164ca261d6e71
              started: "2026-10-20T20:46:59.473Z"
              spawned: false
              bytes: 527719
              windows:
                - start: "2026-10-20T20:46:59.473Z"
                  end: "2026-10-20T14:06:48-07:00"
            - harness: codex
              id: 01a15b12-b635-70e3-ba3e-bda0960cd786
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 603f3e663280b80917eb29a2a8825ea99102d046f4180b5e42f4ef3fbd4fe049
              started: "2026-10-20T20:48:35.847Z"
              spawned: false
              bytes: 473883
              windows:
                - start: "2026-10-20T20:48:35.847Z"
                  end: "2026-10-20T14:06:48-07:00"
            - harness: codex
              id: 01a15b13-34ce-76e0-9c60-7c28afe426bd
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 35d96d5f6e1b94e4aebaac5eebecb5733f05b9d921992a4b769da44612cbc692
              started: "2026-10-20T20:49:08.201Z"
              spawned: false
              bytes: 410752
              windows:
                - start: "2026-10-20T20:49:08.201Z"
                  end: "2026-10-20T14:06:48-07:00"
            - harness: codex
              id: 01a15b14-a455-73d3-b6eb-747fa5c51f2e
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              parent_id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              digest: 5e9dcb922d8ded48c83b6ca3e2bee06930b68cfccbc252a12e44ee9279909c92
              started: "2026-10-20T20:50:42.072Z"
              spawned: true
              bytes: 1237235
              windows:
                - start: "2026-10-20T20:50:42.072Z"
                  end: "2026-10-20T14:06:48-07:00"
        queue_waits:
            - attempt_id: 81f3e3a5eb5fca51613b7f244d34f648
              slot: ariadne:2
              record_digest: dc35c572ea957b5aa408af5805405251f9de0acc4e40870856e0cff460b04a3d
              start: "2026-10-20T20:04:33.83285Z"
              end: "2026-10-20T20:26:00.637372Z"
              queue_seconds: 1286.804522
              raw_seconds: 1286.804522
              exact: true
        cuts:
            - dispatched: "2026-10-20T13:59:53-07:00"
              finalized: "2026-10-20T14:01:56-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T14:09:25-07:00"
      dispatched: "2026-10-20T14:07:29-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: actors.go:37–40 bounds observation and pipe cleanup; fleet.go prints stopped/untouched results and guidance before rechecking. TestFleetResumeBoundedRecheck passes and failed after 30.15 seconds when the timeout was temporarily disabled.
          round: 2
        - id: BR-2
          disposition: addressed
          note: resume_result_test.go:93,99,110 contain reasoned synctest lint allowances. Removing them reproduced all three catalog failures; restoring them made the catalog pass.
          round: 2
        - id: BR-3
          disposition: addressed
          note: README.md:286–291 now documents hub-scoped Slots, both --all flags, visibility-only expansion, and serial stop behavior, linking atlas/workflow/tl-skills.md. These descriptions agree with the command wiring and planner.
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestCouchResumeReceiptConformance passed against the installed real Couch binary for queued/running polling and succeeded/refused/failed terminal receipts, including exit mapping and classification. Scratch configuration isolates it from live slots; atlas documents its execution cadence.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: e2d3af977ec4ae3d9e2758dc6f01cb59596a5b65
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 8c03baeb-45b7-4fcf-acf8-ebb4f6509e50
            repository: github.com/xianxu/ariadne
            issue_id: "000485"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/114570be163e2e0a
            reviewed_head: 928bc8cc72440d319fb3271090fca90943f3c03c
            slot: ariadne:2
            requested_at: "2026-10-21T01:25:49.630784Z"
          phase: finished
          pid: 80323
          started_at: "2026-10-21T01:25:49.747457Z"
          finished_at: "2026-10-21T01:26:53.233552Z"
          session_id: 01a15c10-8967-79e0-bfb8-14926f435dbc
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 928bc8cc72440d319fb3271090fca90943f3c03c
        tracker_head: 195b81c46ba251e8257b1f276cc6df936c901577
        branch_point: 94bf14d9262f4ced7db0ebdb43b44583d7d8e2d5
        since: "2026-10-20T11:08:14-07:00"
        until: "2026-10-20T18:24:55-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T11:08:14-07:00"
              end: "2026-10-20T18:24:55-07:00"
              began: "2026-10-20T11:08:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 64dfd271f821e69ee64459c888d7f23e61ea1d007f38ada57e52683cb3b63b49
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 7126654
              windows:
                - start: "2026-10-20T11:08:14-07:00"
                  end: "2026-10-20T18:24:55-07:00"
            - harness: codex
              id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 2d26f3ef5c02acec70d53c7f97ec0096c45622d02c2e9646f0ecb57614555f88
              started: "2026-10-20T18:21:54.178Z"
              spawned: false
              bytes: 7204547
              windows:
                - start: "2026-10-20T18:21:54.178Z"
                  end: "2026-10-20T18:24:55-07:00"
            - harness: codex
              id: 01a15b0f-1fb0-7c61-8192-ad8d7feab469
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f6143bd731812dbc31b21a9248391f25910c3ce19d4cced28dd1a07f7cc39b24
              started: "2026-10-20T20:44:40.563Z"
              spawned: false
              bytes: 710729
              windows:
                - start: "2026-10-20T20:44:40.563Z"
                  end: "2026-10-20T18:24:55-07:00"
            - harness: codex
              id: 01a15b11-3e3c-7b13-9638-200eca31f132
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 3c5b6458933de9d0999aeb456c8a04edc004eb22d5275d8f6d3164ca261d6e71
              started: "2026-10-20T20:46:59.473Z"
              spawned: false
              bytes: 527719
              windows:
                - start: "2026-10-20T20:46:59.473Z"
                  end: "2026-10-20T18:24:55-07:00"
            - harness: codex
              id: 01a15b12-b635-70e3-ba3e-bda0960cd786
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 603f3e663280b80917eb29a2a8825ea99102d046f4180b5e42f4ef3fbd4fe049
              started: "2026-10-20T20:48:35.847Z"
              spawned: false
              bytes: 473883
              windows:
                - start: "2026-10-20T20:48:35.847Z"
                  end: "2026-10-20T18:24:55-07:00"
            - harness: codex
              id: 01a15b13-34ce-76e0-9c60-7c28afe426bd
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 35d96d5f6e1b94e4aebaac5eebecb5733f05b9d921992a4b769da44612cbc692
              started: "2026-10-20T20:49:08.201Z"
              spawned: false
              bytes: 410752
              windows:
                - start: "2026-10-20T20:49:08.201Z"
                  end: "2026-10-20T18:24:55-07:00"
            - harness: codex
              id: 01a15b14-a455-73d3-b6eb-747fa5c51f2e
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              parent_id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              digest: 5e9dcb922d8ded48c83b6ca3e2bee06930b68cfccbc252a12e44ee9279909c92
              started: "2026-10-20T20:50:42.072Z"
              spawned: true
              bytes: 1237235
              windows:
                - start: "2026-10-20T20:50:42.072Z"
                  end: "2026-10-20T18:24:55-07:00"
        queue_waits:
            - attempt_id: 81f3e3a5eb5fca51613b7f244d34f648
              slot: ariadne:2
              record_digest: dc35c572ea957b5aa408af5805405251f9de0acc4e40870856e0cff460b04a3d
              start: "2026-10-20T20:04:33.83285Z"
              end: "2026-10-20T20:26:00.637372Z"
              queue_seconds: 1286.804522
              raw_seconds: 1286.804522
              exact: true
        cuts:
            - dispatched: "2026-10-20T13:59:53-07:00"
              finalized: "2026-10-20T14:01:56-07:00"
              finalizing: false
            - dispatched: "2026-10-20T14:07:29-07:00"
              finalized: "2026-10-20T14:09:25-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-20T18:26:53-07:00"
      dispatched: "2026-10-20T18:25:49-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      recipe: fix-review
      isolation: fenced
      reviewed: 928bc8cc72440d319fb3271090fca90943f3c03c
      blocked: false
---

# Gate ledger — ariadne#485 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T14:01:56-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `recovery-probes-must-be-bounded` (kind: defect) Bound actor observation so an uncertain resume can finish recovery
  cmd/tl-lead/fleet.go:268 synchronously calls readActors before printing recovery guidance, naming untouched slots, or returning nonzero. Production execActors.Actors uses exec.Command(...).Run without a timeout (actors.go:27–30), so a hung Couch probe after an uncertain resume blocks that required terminal result indefinitely. Under ARCH-CONSTRAINTS/ARCH-ORDER, bound and terminate the probe, print the stopped/untouched result before optional observation, and add a regression with an initially successful actors read followed by a stalled recheck.
- **BR-2** [Important] `tracked-source-checks-must-pass` (kind: defect) Resolve the catalog failures introduced by modeled-delay tests
  TestTrackedGoFilesPassTheCatalog fails on cmd/tl-lead/resume_result_test.go:93,98,108 with test-fixed-sleep. These sleeps advance synctest virtual time rather than guess wall-clock ordering; preserve that coverage and add the repository-supported, reasoned lint:allow annotations on the flagged lines, then rerun the catalog check.
- **BR-3** [Important] `user-surface-requires-readme-documentation` (kind: defect) README update appears missing for both new --all flags
  cmd/tl-lead/status.go:384 and fleet.go:307 introduce user-facing --all flags, but README.md is unchanged in the pinned range. Add concise usage documentation covering hub-scoped defaults and visibility-only expansion for fleet resume, with a link to the updated atlas details, as required by this boundary's README gate.
- **BR-4** [Important] `external-protocol-requires-conformance` (kind: defect) Add conformance coverage for Couch terminal resume receipts
  cmd/tl-lead/resume_result.go:29–35 now depends on receipt fields, terminal status values, and their agreement with process exit. Tests supply hand-authored receipts; the existing TestResumeCandidatesLive checks only candidates, and conformance_test.go covers other Couch surfaces. Under ARCH-MOCK, add an opt-in check against a supported Couch binary in an isolated test configuration and document its cadence, so protocol drift cannot leave all fake-backed tests green.

## Round 2 — 2026-10-20T14:09:25-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — actors.go:37–40 bounds observation and pipe cleanup; fleet.go prints stopped/untouched results and guidance before rechecking. TestFleetResumeBoundedRecheck passes and failed after 30.15 seconds when the timeout was temporarily disabled.
- BR-2 — addressed — resume_result_test.go:93,99,110 contain reasoned synctest lint allowances. Removing them reproduced all three catalog failures; restoring them made the catalog pass.
- BR-3 — addressed — README.md:286–291 now documents hub-scoped Slots, both --all flags, visibility-only expansion, and serial stop behavior, linking atlas/workflow/tl-skills.md. These descriptions agree with the command wiring and planner.
- BR-4 — addressed — TestCouchResumeReceiptConformance passed against the installed real Couch binary for queued/running polling and succeeded/refused/failed terminal receipts, including exit mapping and classification. Scratch configuration isolates it from live slots; atlas documents its execution cadence.

## Round 3 — 2026-10-20T18:26:53-07:00 (codex) — passed

## Open findings

(none — every finding has been disposed)
