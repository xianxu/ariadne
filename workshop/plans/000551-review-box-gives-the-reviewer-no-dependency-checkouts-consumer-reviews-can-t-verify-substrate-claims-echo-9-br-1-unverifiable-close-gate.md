---
gate: boundary-review
issue: 551
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: be9481b7-d545-4949-9312-442f6cf56e9c
            repository: github.com/xianxu/ariadne
            issue_id: "000551"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/8b88f57d222b86c6
            reviewed_head: 0ae0016dbbe6634bfbf59132045e683bfa76862a
            slot: ariadne:1
            requested_at: "2026-10-22T13:05:07.029365Z"
          phase: finished
          pid: 91308
          started_at: "2026-10-22T13:05:07.208371Z"
          finished_at: "2026-10-22T13:10:10.698053Z"
          session_id: 01a163b7-1e4b-76a0-89a0-d52ba436dcc3
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 0ae0016dbbe6634bfbf59132045e683bfa76862a
        tracker_head: 535bcea8d3771e94a2cac6d275980be1a27109c9
        branch_point: cce05bc3f48f61cb63d010060ad2cb79e5ebd1c2
        since: "2026-10-22T05:22:03-07:00"
        until: "2026-10-22T05:52:06-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T05:22:03-07:00"
              end: "2026-10-22T05:52:06-07:00"
              began: "2026-10-22T05:22:03-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 951dd8309cb4ce6a95634f9707f04e114deb7ee532107020d50d28c8da9beb2d
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 77072203
              windows:
                - start: "2026-10-22T05:22:03-07:00"
                  end: "2026-10-22T05:52:06-07:00"
            - harness: codex
              id: 01a16390-3628-7323-8d67-22c2df895e92
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 916f86340855ddc9ea6609b2a056ed564b381e9529f9bc3e8faaaf0f52b5d8a6
              started: "2026-10-22T12:22:37.91Z"
              spawned: true
              bytes: 906552
              windows:
                - start: "2026-10-22T12:22:37.91Z"
                  end: "2026-10-22T05:52:06-07:00"
            - harness: codex
              id: 01a16394-faae-7e72-80ce-b26bdb3e5085
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 9b07a1eaefffa83ef7fa06fa1ef6c0a07ec6e9138e59918b8b816b0e955a310b
              started: "2026-10-22T12:27:50.737Z"
              spawned: false
              bytes: 665013
              windows:
                - start: "2026-10-22T12:27:50.737Z"
                  end: "2026-10-22T05:52:06-07:00"
            - harness: codex
              id: 01a16398-2b41-7de1-8e6e-8e137e7d3bb9
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 5837d7584f6c9541ac65f7588a89a3b772209ebaeee16e360a3235d0002c802a
              started: "2026-10-22T12:31:19.952Z"
              spawned: false
              bytes: 428201
              windows:
                - start: "2026-10-22T12:31:19.952Z"
                  end: "2026-10-22T05:52:06-07:00"
            - harness: codex
              id: 01a1639a-1f98-77b0-b97e-07c0a29bb4f3
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 8f3eb6c67955f78db4fec45778db206e7971efa4be02041429e01f73f2308a0f
              started: "2026-10-22T12:33:27.903Z"
              spawned: false
              bytes: 476521
              windows:
                - start: "2026-10-22T12:33:27.903Z"
                  end: "2026-10-22T05:52:06-07:00"
            - harness: codex
              id: 01a1639a-b758-7140-9b2a-174cba0e9519
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f8cea2db65988a428740000c106104c635065546950c1e25fd41cc9678cda9e4
              started: "2026-10-22T12:34:06.675Z"
              spawned: false
              bytes: 403592
              windows:
                - start: "2026-10-22T12:34:06.675Z"
                  end: "2026-10-22T05:52:06-07:00"
            - harness: codex
              id: 01a1639c-2465-7fa2-a2c9-b25296e4bb7a
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 12c376d00cca7e975f2733b6f73fc0fc14131ab80e0b9da4a588e1b1c1814bf7
              started: "2026-10-22T12:35:39.839Z"
              spawned: true
              bytes: 653339
              windows:
                - start: "2026-10-22T12:35:39.839Z"
                  end: "2026-10-22T05:52:06-07:00"
            - harness: codex
              id: 01a1639c-5953-7d23-bb18-fd9e20b7b025
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 7ba221fb55559b29d3e303a6f84f591a0df22284bbd2e66d8fd3c60678215029
              started: "2026-10-22T12:35:53.384Z"
              spawned: true
              bytes: 986401
              windows:
                - start: "2026-10-22T12:35:53.384Z"
                  end: "2026-10-22T05:52:06-07:00"
      "n": 1
      timestamp: "2026-10-22T06:10:11-07:00"
      dispatched: "2026-10-22T06:05:07-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README update appears missing for boundary-review dependency bindings
          detail: 'cmd/sdlc/milestoneclose.go:899 and cmd/sdlc/internal/reviewbox/confine.go:50 introduce disposable SDLC_DEP_* bindings for reviewers and preparation commands, but README.md is unchanged. Its boundary-preparation instructions at README.md:146 omit these bindings, while its existing dependency environment documentation at README.md:709 describes local checks. Add the boundary-review contract: last-fetched origin/main pins, no automatic fetch, setup failure recovery, isolation modes, and retry/cleanup lifetime, linking the updated atlas. This is the required README docs gate and an ARCH-PURPOSE completion gap.'
          family: user-surface-documentation
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 429bdc0f-ef80-458d-ace1-942344e84ca8
            repository: github.com/xianxu/ariadne
            issue_id: "000551"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/ab6625d403cb6bb6
            reviewed_head: 80e095147c447abf1a3c41799029a050eca5817f
            slot: ariadne:1
            requested_at: "2026-10-22T13:12:26.269452Z"
          phase: finished
          pid: 10699
          started_at: "2026-10-22T13:12:26.439918Z"
          finished_at: "2026-10-22T13:14:11.234551Z"
          session_id: 01a163bd-d22b-7150-91b7-3c87affaf86a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 80e095147c447abf1a3c41799029a050eca5817f
        tracker_head: 38bbc8b15ed646ac30364c1864a1a38586a65b6a
        branch_point: cce05bc3f48f61cb63d010060ad2cb79e5ebd1c2
        since: "2026-10-22T05:22:03-07:00"
        until: "2026-10-22T06:11:50-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T05:22:03-07:00"
              end: "2026-10-22T06:11:50-07:00"
              began: "2026-10-22T05:22:03-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 3d6227e30d86230436eef750d8243637e6ef9dd503b2382d302528bb77ff7405
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 77336312
              windows:
                - start: "2026-10-22T05:22:03-07:00"
                  end: "2026-10-22T06:11:50-07:00"
            - harness: codex
              id: 01a16390-3628-7323-8d67-22c2df895e92
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 916f86340855ddc9ea6609b2a056ed564b381e9529f9bc3e8faaaf0f52b5d8a6
              started: "2026-10-22T12:22:37.91Z"
              spawned: true
              bytes: 906552
              windows:
                - start: "2026-10-22T12:22:37.91Z"
                  end: "2026-10-22T06:11:50-07:00"
            - harness: codex
              id: 01a16394-faae-7e72-80ce-b26bdb3e5085
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 9b07a1eaefffa83ef7fa06fa1ef6c0a07ec6e9138e59918b8b816b0e955a310b
              started: "2026-10-22T12:27:50.737Z"
              spawned: false
              bytes: 665013
              windows:
                - start: "2026-10-22T12:27:50.737Z"
                  end: "2026-10-22T06:11:50-07:00"
            - harness: codex
              id: 01a16398-2b41-7de1-8e6e-8e137e7d3bb9
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 5837d7584f6c9541ac65f7588a89a3b772209ebaeee16e360a3235d0002c802a
              started: "2026-10-22T12:31:19.952Z"
              spawned: false
              bytes: 428201
              windows:
                - start: "2026-10-22T12:31:19.952Z"
                  end: "2026-10-22T06:11:50-07:00"
            - harness: codex
              id: 01a1639a-1f98-77b0-b97e-07c0a29bb4f3
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 8f3eb6c67955f78db4fec45778db206e7971efa4be02041429e01f73f2308a0f
              started: "2026-10-22T12:33:27.903Z"
              spawned: false
              bytes: 476521
              windows:
                - start: "2026-10-22T12:33:27.903Z"
                  end: "2026-10-22T06:11:50-07:00"
            - harness: codex
              id: 01a1639a-b758-7140-9b2a-174cba0e9519
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f8cea2db65988a428740000c106104c635065546950c1e25fd41cc9678cda9e4
              started: "2026-10-22T12:34:06.675Z"
              spawned: false
              bytes: 403592
              windows:
                - start: "2026-10-22T12:34:06.675Z"
                  end: "2026-10-22T06:11:50-07:00"
            - harness: codex
              id: 01a1639c-2465-7fa2-a2c9-b25296e4bb7a
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 12c376d00cca7e975f2733b6f73fc0fc14131ab80e0b9da4a588e1b1c1814bf7
              started: "2026-10-22T12:35:39.839Z"
              spawned: true
              bytes: 653339
              windows:
                - start: "2026-10-22T12:35:39.839Z"
                  end: "2026-10-22T06:11:50-07:00"
            - harness: codex
              id: 01a1639c-5953-7d23-bb18-fd9e20b7b025
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 7ba221fb55559b29d3e303a6f84f591a0df22284bbd2e66d8fd3c60678215029
              started: "2026-10-22T12:35:53.384Z"
              spawned: true
              bytes: 986401
              windows:
                - start: "2026-10-22T12:35:53.384Z"
                  end: "2026-10-22T06:11:50-07:00"
        cuts:
            - dispatched: "2026-10-22T06:05:07-07:00"
              finalized: "2026-10-22T06:10:11-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-22T06:14:12-07:00"
      dispatched: "2026-10-22T06:12:26-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:146–160 now documents disposable bindings, last-fetched origin/main pins, no automatic fetch, repair/rerun recovery, isolation modes, and retry/cleanup lifetime, linking atlas/workflow/sdlc-binary.md. The documented behavior agrees with reviewdependencies.go, checkout.go, and confinement wiring; this correction changes prose only.
          round: 2
      findings:
        - kind: unverifiable
          id: BR-2
          severity: Important
          title: Kernel enforcement of dependency write protection could not be verified
          detail: 'At cmd/sdlc/reviewdependencies_live_test.go:22, WF_LIVE_REVIEW_DEPENDENCIES=1 go test ./cmd/sdlc -run ''^TestReviewDependencyLiveFence$'' -count=1 failed with exit status 71: sandbox-exec: sandbox_apply: Operation not permitted. ARCH-SECURE requires evidence that dependency clones and source Git directories reject writes while the primary review checkout remains writable. Provide captured passing output for this exact pinned head from a host permitting the OS fence, or rerun review in such an environment; the issue''s prose report alone does not establish that measurement.'
          family: os-fence-conformance
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 76bd4bc4-3b03-4e80-a6f2-09d0ad6f88d1
            repository: github.com/xianxu/ariadne
            issue_id: "000551"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/577d555fe9b0d28f
            reviewed_head: 09357c3bb13f8a2ea6999213102adac2ec8478c3
            slot: ariadne:1
            requested_at: "2026-10-22T13:18:30.335014Z"
          phase: finished
          pid: 73047
          started_at: "2026-10-22T13:18:30.430788Z"
          finished_at: "2026-10-22T13:20:00.730979Z"
          session_id: 01a163c3-5f85-72e1-8138-aeb9620f2db9
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 09357c3bb13f8a2ea6999213102adac2ec8478c3
        tracker_head: 38bbc8b15ed646ac30364c1864a1a38586a65b6a
        branch_point: cce05bc3f48f61cb63d010060ad2cb79e5ebd1c2
        since: "2026-10-22T05:22:03-07:00"
        until: "2026-10-22T06:17:21-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T05:22:03-07:00"
              end: "2026-10-22T06:17:21-07:00"
              began: "2026-10-22T05:22:03-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: df0aa61b108526331c0f2917f61898c598c74cf4c9b76c38aa985e5c5f4c8e3b
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 77614250
              windows:
                - start: "2026-10-22T05:22:03-07:00"
                  end: "2026-10-22T06:17:21-07:00"
            - harness: codex
              id: 01a16390-3628-7323-8d67-22c2df895e92
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 916f86340855ddc9ea6609b2a056ed564b381e9529f9bc3e8faaaf0f52b5d8a6
              started: "2026-10-22T12:22:37.91Z"
              spawned: true
              bytes: 906552
              windows:
                - start: "2026-10-22T12:22:37.91Z"
                  end: "2026-10-22T06:17:21-07:00"
            - harness: codex
              id: 01a16394-faae-7e72-80ce-b26bdb3e5085
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 9b07a1eaefffa83ef7fa06fa1ef6c0a07ec6e9138e59918b8b816b0e955a310b
              started: "2026-10-22T12:27:50.737Z"
              spawned: false
              bytes: 665013
              windows:
                - start: "2026-10-22T12:27:50.737Z"
                  end: "2026-10-22T06:17:21-07:00"
            - harness: codex
              id: 01a16398-2b41-7de1-8e6e-8e137e7d3bb9
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 5837d7584f6c9541ac65f7588a89a3b772209ebaeee16e360a3235d0002c802a
              started: "2026-10-22T12:31:19.952Z"
              spawned: false
              bytes: 428201
              windows:
                - start: "2026-10-22T12:31:19.952Z"
                  end: "2026-10-22T06:17:21-07:00"
            - harness: codex
              id: 01a1639a-1f98-77b0-b97e-07c0a29bb4f3
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 8f3eb6c67955f78db4fec45778db206e7971efa4be02041429e01f73f2308a0f
              started: "2026-10-22T12:33:27.903Z"
              spawned: false
              bytes: 476521
              windows:
                - start: "2026-10-22T12:33:27.903Z"
                  end: "2026-10-22T06:17:21-07:00"
            - harness: codex
              id: 01a1639a-b758-7140-9b2a-174cba0e9519
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f8cea2db65988a428740000c106104c635065546950c1e25fd41cc9678cda9e4
              started: "2026-10-22T12:34:06.675Z"
              spawned: false
              bytes: 403592
              windows:
                - start: "2026-10-22T12:34:06.675Z"
                  end: "2026-10-22T06:17:21-07:00"
            - harness: codex
              id: 01a1639c-2465-7fa2-a2c9-b25296e4bb7a
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 12c376d00cca7e975f2733b6f73fc0fc14131ab80e0b9da4a588e1b1c1814bf7
              started: "2026-10-22T12:35:39.839Z"
              spawned: true
              bytes: 653339
              windows:
                - start: "2026-10-22T12:35:39.839Z"
                  end: "2026-10-22T06:17:21-07:00"
            - harness: codex
              id: 01a1639c-5953-7d23-bb18-fd9e20b7b025
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 7ba221fb55559b29d3e303a6f84f591a0df22284bbd2e66d8fd3c60678215029
              started: "2026-10-22T12:35:53.384Z"
              spawned: true
              bytes: 986401
              windows:
                - start: "2026-10-22T12:35:53.384Z"
                  end: "2026-10-22T06:17:21-07:00"
        queue_waits:
            - attempt_id: 4638dd482de33746116a1788f13ba510
              slot: ariadne:1
              record_digest: 901439c548ca236a6e2afb309bb15b8fbc1cd5204e407812aecf18bf6358786a
              start: "2026-10-22T12:52:57.664702Z"
              end: "2026-10-22T13:03:19.045086Z"
              queue_seconds: 621.380384
              raw_seconds: 621.380384
              exact: true
        cuts:
            - dispatched: "2026-10-22T06:05:07-07:00"
              finalized: "2026-10-22T06:10:11-07:00"
              finalizing: false
            - dispatched: "2026-10-22T06:12:26-07:00"
              finalized: "2026-10-22T06:14:12-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-22T06:20:01-07:00"
      dispatched: "2026-10-22T06:18:30-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md now documents dependency bindings, pins, modes, recovery and lifecycle, consistent with reviewdependencies.go, reviewisolation.go and the atlas.
          round: 3
        - id: BR-2
          disposition: withdrawn
          note: 'Re-running the exact live-fence command failed here with sandbox_apply: Operation not permitted. The committed host evidence records exit 0 and PASS; all six recorded source blob IDs match the pinned head, and cmd/ plus pkg/ are unchanged since the tested commit. The fixture exercises production isolation and asserts dependency clone/source/common-dir write denial while retaining primary-clone write access.'
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: 09357c3bb13f8a2ea6999213102adac2ec8478c3
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 373c21f3-51a9-4b71-b0db-5907bb64fac3
            repository: github.com/xianxu/ariadne
            issue_id: "000551"
            recipe: fix-review
            round: 4
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/2b5e00e85b8ca3cd
            reviewed_head: 1151135bfd024f4b709f1af158a58ed52f7a2ce9
            slot: ariadne:1
            requested_at: "2026-10-22T13:34:21.738757Z"
          phase: finished
          pid: 52595
          started_at: "2026-10-22T13:34:21.943766Z"
          finished_at: "2026-10-22T13:35:45.251725Z"
          session_id: 01a163d1-e4e0-7e22-ac5b-6e15fd4c4444
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 1151135bfd024f4b709f1af158a58ed52f7a2ce9
        tracker_head: 546c8d1a5b8094ad7e4d5605c2f5edf4d11e3314
        branch_point: cce05bc3f48f61cb63d010060ad2cb79e5ebd1c2
        since: "2026-10-22T05:22:03-07:00"
        until: "2026-10-22T06:32:33-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T05:22:03-07:00"
              end: "2026-10-22T06:32:33-07:00"
              began: "2026-10-22T05:22:03-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: da0a0e8d0d999d89fac35e41b3b11ea9b7f85ac579baec9d042b790f152e07f7
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 78336402
              windows:
                - start: "2026-10-22T05:22:03-07:00"
                  end: "2026-10-22T06:32:33-07:00"
            - harness: codex
              id: 01a16390-3628-7323-8d67-22c2df895e92
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 916f86340855ddc9ea6609b2a056ed564b381e9529f9bc3e8faaaf0f52b5d8a6
              started: "2026-10-22T12:22:37.91Z"
              spawned: true
              bytes: 906552
              windows:
                - start: "2026-10-22T12:22:37.91Z"
                  end: "2026-10-22T06:32:33-07:00"
            - harness: codex
              id: 01a16394-faae-7e72-80ce-b26bdb3e5085
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 9b07a1eaefffa83ef7fa06fa1ef6c0a07ec6e9138e59918b8b816b0e955a310b
              started: "2026-10-22T12:27:50.737Z"
              spawned: false
              bytes: 665013
              windows:
                - start: "2026-10-22T12:27:50.737Z"
                  end: "2026-10-22T06:32:33-07:00"
            - harness: codex
              id: 01a16398-2b41-7de1-8e6e-8e137e7d3bb9
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 5837d7584f6c9541ac65f7588a89a3b772209ebaeee16e360a3235d0002c802a
              started: "2026-10-22T12:31:19.952Z"
              spawned: false
              bytes: 428201
              windows:
                - start: "2026-10-22T12:31:19.952Z"
                  end: "2026-10-22T06:32:33-07:00"
            - harness: codex
              id: 01a1639a-1f98-77b0-b97e-07c0a29bb4f3
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 8f3eb6c67955f78db4fec45778db206e7971efa4be02041429e01f73f2308a0f
              started: "2026-10-22T12:33:27.903Z"
              spawned: false
              bytes: 476521
              windows:
                - start: "2026-10-22T12:33:27.903Z"
                  end: "2026-10-22T06:32:33-07:00"
            - harness: codex
              id: 01a1639a-b758-7140-9b2a-174cba0e9519
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: f8cea2db65988a428740000c106104c635065546950c1e25fd41cc9678cda9e4
              started: "2026-10-22T12:34:06.675Z"
              spawned: false
              bytes: 403592
              windows:
                - start: "2026-10-22T12:34:06.675Z"
                  end: "2026-10-22T06:32:33-07:00"
            - harness: codex
              id: 01a1639c-2465-7fa2-a2c9-b25296e4bb7a
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 5569a95c3509dc490e2023b309f8576cd73d905767447bd97b51733535a62d8f
              started: "2026-10-22T12:35:39.839Z"
              spawned: true
              bytes: 759884
              windows:
                - start: "2026-10-22T12:35:39.839Z"
                  end: "2026-10-22T06:32:33-07:00"
            - harness: codex
              id: 01a1639c-5953-7d23-bb18-fd9e20b7b025
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 1b80ffff8de6bb107bae9cb399b0c5155d84ac16225547fa5743260f85ce3f1e
              started: "2026-10-22T12:35:53.384Z"
              spawned: true
              bytes: 1175269
              windows:
                - start: "2026-10-22T12:35:53.384Z"
                  end: "2026-10-22T06:32:33-07:00"
        queue_waits:
            - attempt_id: 4638dd482de33746116a1788f13ba510
              slot: ariadne:1
              record_digest: 901439c548ca236a6e2afb309bb15b8fbc1cd5204e407812aecf18bf6358786a
              start: "2026-10-22T12:52:57.664702Z"
              end: "2026-10-22T13:03:19.045086Z"
              queue_seconds: 621.380384
              raw_seconds: 621.380384
              exact: true
        cuts:
            - dispatched: "2026-10-22T06:05:07-07:00"
              finalized: "2026-10-22T06:10:11-07:00"
              finalizing: false
            - dispatched: "2026-10-22T06:12:26-07:00"
              finalized: "2026-10-22T06:14:12-07:00"
              finalizing: false
            - dispatched: "2026-10-22T06:18:30-07:00"
              finalized: "2026-10-22T06:20:01-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-22T06:35:46-07:00"
      dispatched: "2026-10-22T06:34:21-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-3
          severity: Critical
          title: scope changed after close — run a full `sdlc close`
          detail: 'cmd/sdlc/reviewisolation.go:65-66 adds dependency-manifest output to close diagnostics, and workshop/plans/000551-review-dependency-checkouts-plan.md:129-133 records new post-close runtime work. Neither disposed finding requested this added output: BR-1 concerned README documentation and BR-2 concerned host fence evidence. The narrow review cannot admit this expanded behavior; run a full sdlc close over the resulting implementation.'
          family: post-close-scope-control
          round: 4
      recipe: fix-review
      isolation: fenced
      blocked: true
disputes:
    - id: BR-2
      rationale: 'The nested reviewer sandbox cannot apply a second OS fence, but the host can. No implementation change is required: committed host evidence includes the exact passing command output and tested HEAD/source blob identities, which remain unchanged in this evidence-only revision.'
      evidence: 'env WF_LIVE_REVIEW_DEPENDENCIES=1 go test ./cmd/sdlc -run ^TestReviewDependencyLiveFence$ -count=1 -v: exit 0, PASS, test 0.91s / command 1.391s on 80e095147c447abf1a3c41799029a050eca5817f. Raw output and immutable relevant blob IDs: workshop/plans/000551-host-fence-evidence.md. The fixture proves dependency clone/source/common-dir writes and ref updates are denied while the primary clone stays writable.'
      after: 2
      timestamp: "2026-10-22T13:17:19Z"
      by: Xian Xu on Xian’s MacBook Pro at ariadne:1 (/Users/xianxu/workspace/worktree/ariadne-slot1/ariadne)
---

# Gate ledger — ariadne#551 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T06:10:11-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-surface-documentation` (kind: defect) README update appears missing for boundary-review dependency bindings
  cmd/sdlc/milestoneclose.go:899 and cmd/sdlc/internal/reviewbox/confine.go:50 introduce disposable SDLC_DEP_* bindings for reviewers and preparation commands, but README.md is unchanged. Its boundary-preparation instructions at README.md:146 omit these bindings, while its existing dependency environment documentation at README.md:709 describes local checks. Add the boundary-review contract: last-fetched origin/main pins, no automatic fetch, setup failure recovery, isolation modes, and retry/cleanup lifetime, linking the updated atlas. This is the required README docs gate and an ARCH-PURPOSE completion gap.

## Round 2 — 2026-10-22T06:14:12-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — README.md:146–160 now documents disposable bindings, last-fetched origin/main pins, no automatic fetch, repair/rerun recovery, isolation modes, and retry/cleanup lifetime, linking atlas/workflow/sdlc-binary.md. The documented behavior agrees with reviewdependencies.go, checkout.go, and confinement wiring; this correction changes prose only.

### Raised

- **BR-2** [Important] `os-fence-conformance` (kind: unverifiable) Kernel enforcement of dependency write protection could not be verified
  At cmd/sdlc/reviewdependencies_live_test.go:22, WF_LIVE_REVIEW_DEPENDENCIES=1 go test ./cmd/sdlc -run '^TestReviewDependencyLiveFence$' -count=1 failed with exit status 71: sandbox-exec: sandbox_apply: Operation not permitted. ARCH-SECURE requires evidence that dependency clones and source Git directories reject writes while the primary review checkout remains writable. Provide captured passing output for this exact pinned head from a host permitting the OS fence, or rerun review in such an environment; the issue's prose report alone does not establish that measurement.

## Round 3 — 2026-10-22T06:20:01-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — README.md now documents dependency bindings, pins, modes, recovery and lifecycle, consistent with reviewdependencies.go, reviewisolation.go and the atlas.
- BR-2 — withdrawn — Re-running the exact live-fence command failed here with sandbox_apply: Operation not permitted. The committed host evidence records exit 0 and PASS; all six recorded source blob IDs match the pinned head, and cmd/ plus pkg/ are unchanged since the tested commit. The fixture exercises production isolation and asserts dependency clone/source/common-dir write denial while retaining primary-clone write access.

## Round 4 — 2026-10-22T06:35:46-07:00 (codex) — BLOCKED

### Raised

- **BR-3** [Critical] `post-close-scope-control` (kind: defect) scope changed after close — run a full `sdlc close`
  cmd/sdlc/reviewisolation.go:65-66 adds dependency-manifest output to close diagnostics, and workshop/plans/000551-review-dependency-checkouts-plan.md:129-133 records new post-close runtime work. Neither disposed finding requested this added output: BR-1 concerned README documentation and BR-2 concerned host fence evidence. The narrow review cannot admit this expanded behavior; run a full sdlc close over the resulting implementation.

## Disputes

Findings the implementer declined (`sdlc finding dispute`), each answered by the
next review round.

- **BR-2** — after round 2 (2026-10-22T13:17:19Z, Xian Xu on Xian’s MacBook Pro at ariadne:1 (/Users/xianxu/workspace/worktree/ariadne-slot1/ariadne))
  rationale: The nested reviewer sandbox cannot apply a second OS fence, but the host can. No implementation change is required: committed host evidence includes the exact passing command output and tested HEAD/source blob identities, which remain unchanged in this evidence-only revision.
  evidence: env WF_LIVE_REVIEW_DEPENDENCIES=1 go test ./cmd/sdlc -run ^TestReviewDependencyLiveFence$ -count=1 -v: exit 0, PASS, test 0.91s / command 1.391s on 80e095147c447abf1a3c41799029a050eca5817f. Raw output and immutable relevant blob IDs: workshop/plans/000551-host-fence-evidence.md. The fixture proves dependency clone/source/common-dir writes and ref updates are denied while the primary clone stays writable.

## Open findings

- **BR-3** [Critical] `post-close-scope-control` (kind: defect) scope changed after close — run a full `sdlc close`
