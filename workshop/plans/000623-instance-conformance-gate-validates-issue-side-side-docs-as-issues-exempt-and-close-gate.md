---
gate: boundary-review
issue: 623
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 81d71839-70dc-480a-9018-3c007e0b3b51
            repository: github.com/xianxu/ariadne
            issue_id: "000623"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/6b7beb50ca5b00b7
            reviewed_head: d2e00528b4fcf5a5f1bc14ea77c441a70247cd11
            slot: ariadne:1
            requested_at: "2026-10-25T19:52:14.527287Z"
          phase: finished
          pid: 75197
          started_at: "2026-10-25T19:52:14.674722Z"
          finished_at: "2026-10-25T19:58:02.729615Z"
          session_id: 01a1749e-ee0b-75d1-950f-a391228b1d3c
          outcome: returned
      actual_inputs:
        version: 4
        method: active-time-v3-branch-v1
        repository: github.com/xianxu/ariadne
        code_head: d2e00528b4fcf5a5f1bc14ea77c441a70247cd11
        tracker_head: 6bf4d45f75a4913271cd52373a8333704c4c722d
        branch_point: a141d5da4078701e806ca7716ca9591aff9fd0c9
        since: "2026-10-25T12:00:36-07:00"
        until: "2026-10-25T12:46:36-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-25T12:00:36-07:00"
              end: "2026-10-25T12:46:36-07:00"
              began: "2026-10-25T12:00:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 8e6e4fe998fb396c3b3c44b7591e4a0507ab8b2908f3da1a6b707c7fd3827d83
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 90733750
              windows:
                - start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T12:46:36-07:00"
            - harness: codex
              id: 01a17470-2f52-7c93-8055-adf1eca6d11e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: bd119fb91a0a7fa6882bdc644481a4422b8ba86e07b61c37ec4e5404fda39657
              started: "2026-10-25T19:01:11.805Z"
              spawned: true
              bytes: 706893
              windows:
                - start: "2026-10-25T19:01:11.805Z"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T19:01:11.805Z"
                  end: "2026-10-25T12:46:36-07:00"
            - harness: codex
              id: 01a17479-4bac-7480-bb83-5344817246c1
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 84f0494ffd37d41ec3ac40f959f1fb607a0e5b06d98c4d1f45ad913256b31a14
              started: "2026-10-25T19:11:08.824Z"
              spawned: true
              bytes: 409080
              windows:
                - start: "2026-10-25T19:11:08.824Z"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T19:11:08.824Z"
                  end: "2026-10-25T12:46:36-07:00"
            - harness: codex
              id: 01a1747d-a887-7140-86ab-40aba4e46c51
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: cb0973efa553d80997b3fbe416804ce7110bae778394a041030123fb848aabfe
              started: "2026-10-25T19:15:55.307Z"
              spawned: false
              bytes: 640710
              windows:
                - start: "2026-10-25T19:15:55.307Z"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T19:15:55.307Z"
                  end: "2026-10-25T12:46:36-07:00"
            - harness: codex
              id: 01a17481-4a0c-7800-8a3e-beaf8c21a285
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c6f66f7907c7fd0d6113676f09f895cf7ee1d86d37a47dc69bb96471c785eb58
              started: "2026-10-25T19:19:53.146Z"
              spawned: false
              bytes: 474411
              windows:
                - start: "2026-10-25T19:19:53.146Z"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T19:19:53.146Z"
                  end: "2026-10-25T12:46:36-07:00"
            - harness: codex
              id: 01a17481-a65a-7063-87d1-e529a7643001
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 164717b2fe4f62a64536bc90f8ca5a6fb512f1b00c78ce189c1a64464282d62c
              started: "2026-10-25T19:20:16.534Z"
              spawned: false
              bytes: 443020
              windows:
                - start: "2026-10-25T19:20:16.534Z"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T19:20:16.534Z"
                  end: "2026-10-25T12:46:36-07:00"
            - harness: codex
              id: 01a17483-4ea4-7913-9117-995a904d474a
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: f07c227f11cda09839054f2f2a60b73db0600398d6b92b809d68e9ef2de5686b
              started: "2026-10-25T19:22:04.947Z"
              spawned: true
              bytes: 389663
              windows:
                - start: "2026-10-25T19:22:04.947Z"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T19:22:04.947Z"
                  end: "2026-10-25T12:46:36-07:00"
            - harness: codex
              id: 01a17483-6f05-7592-800f-0ea86cd2423e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: dfe94b4182d1cc3c784ed83ac39d88b3e172279766fb6c2de8468342cdecc658
              started: "2026-10-25T19:22:13.342Z"
              spawned: true
              bytes: 433256
              windows:
                - start: "2026-10-25T19:22:13.342Z"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T19:22:13.342Z"
                  end: "2026-10-25T12:46:36-07:00"
            - harness: codex
              id: 01a1748e-d330-7ad1-8793-763991a75b0c
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 08a1cf6f4193dea17c661247d9f00323db6c236e053c8016d039737d6673236e
              started: "2026-10-25T19:34:39.798Z"
              spawned: true
              bytes: 760712
              windows:
                - start: "2026-10-25T19:34:39.798Z"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T19:34:39.798Z"
                  end: "2026-10-25T12:46:36-07:00"
            - harness: codex
              id: 01a1748f-3b11-77b0-862c-8ba6a3ef22be
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 6bd5758c001112ddaceec51b63501fe3dc9187389fd96e754320efaa5e0c165a
              started: "2026-10-25T19:35:06.344Z"
              spawned: true
              bytes: 806910
              windows:
                - start: "2026-10-25T19:35:06.344Z"
                  end: "2026-10-25T12:46:36-07:00"
              dispatch_windows:
                - start: "2026-10-25T19:35:06.344Z"
                  end: "2026-10-25T12:46:36-07:00"
        branches:
            - root: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              path: /Users/xianxu/workspace/ariadne/.git/worktrees/ariadne/logs/HEAD
              head: d2e00528b4fcf5a5f1bc14ea77c441a70247cd11
              branch: 000623-instance-conformance-gate-validates-issue-side-side-docs-as-issues-exempt-and
              digest: f2a33b8450acb7753a5aa64b85811e4a2c4c2eae8b9a6362e98d150748c60fb4
              bytes: 257103
              intervals:
                - source: claude/048d3d3d-9c88-4143-a8ca-5377f6b48e29
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/048d3d3d-9c88-4143-a8ca-5377f6b48e29
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/177ea708-a128-45a7-9521-84a27f9ebaca
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/177ea708-a128-45a7-9521-84a27f9ebaca
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/1e6274a6-91d3-460d-bb29-61edc45eb8c4
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/1e6274a6-91d3-460d-bb29-61edc45eb8c4
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/1f48235f-b720-4bcd-b5fe-e162d58f0bec
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/1f48235f-b720-4bcd-b5fe-e162d58f0bec
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/23e8ad58-6504-4864-b924-13cd5087f8f9
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/23e8ad58-6504-4864-b924-13cd5087f8f9
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/28f99a1f-4cf2-49f8-92e1-5988a081fc33
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/28f99a1f-4cf2-49f8-92e1-5988a081fc33
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/2d77d4c4-04f2-43e9-b90d-5de4222da317
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/2d77d4c4-04f2-43e9-b90d-5de4222da317
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/30512fb8-7532-440f-9b36-a0d00a565818
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/30512fb8-7532-440f-9b36-a0d00a565818
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/3706d265-d576-40bd-a767-59c9ab8b8d8e
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/3706d265-d576-40bd-a767-59c9ab8b8d8e
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/49065235-9ea1-4bb2-8435-6cf212e89ec6
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/49065235-9ea1-4bb2-8435-6cf212e89ec6
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/6120c1ee-b812-4574-bf69-1df94ba34f01
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/6120c1ee-b812-4574-bf69-1df94ba34f01
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/640fb250-6f1e-41b0-a0f9-a11aaea6af5b
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/640fb250-6f1e-41b0-a0f9-a11aaea6af5b
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/894d0fc1-b17e-417b-b917-bb4d91b960d0
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/894d0fc1-b17e-417b-b917-bb4d91b960d0
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/90a76a30-7c2d-4c41-861f-ec8116620a96
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/90a76a30-7c2d-4c41-861f-ec8116620a96
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/95fdaae8-b456-4412-9f20-918bacc4a224
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/95fdaae8-b456-4412-9f20-918bacc4a224
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/aa423ee5-ee3f-4bec-a755-2eb9d8f4fb8d
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/aa423ee5-ee3f-4bec-a755-2eb9d8f4fb8d
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/b8e0ea99-c622-47ba-a8ef-fd22b1094f1d
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/b8e0ea99-c622-47ba-a8ef-fd22b1094f1d
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/bfc16deb-0847-40c1-9257-4ff948ea8c68
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/bfc16deb-0847-40c1-9257-4ff948ea8c68
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/c0f573bb-4a7c-4721-92e7-46aa12aca9ea
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/c0f573bb-4a7c-4721-92e7-46aa12aca9ea
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/c261362e-1982-4392-b8c9-14570e962587
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/c261362e-1982-4392-b8c9-14570e962587
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/d516b8a1-78a9-495f-935c-8c727dbc2643
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/d516b8a1-78a9-495f-935c-8c727dbc2643
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/d72e8687-726e-4cbc-9e47-286bdbabe81b
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/d72e8687-726e-4cbc-9e47-286bdbabe81b
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/da0a391f-3f8d-4dd2-81f6-fedefb4f2714
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/da0a391f-3f8d-4dd2-81f6-fedefb4f2714
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/e14daa66-e63d-48a4-85d5-69be51028984
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/e14daa66-e63d-48a4-85d5-69be51028984
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/eb1812e1-d294-4b64-bf79-a4e5dcdb5b69
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/eb1812e1-d294-4b64-bf79-a4e5dcdb5b69
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/ec6cbacf-e569-468a-92f1-9eb07832a661
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/ec6cbacf-e569-468a-92f1-9eb07832a661
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/ec9d27b1-e4e6-40b9-9dcc-97ef822c0339
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/ec9d27b1-e4e6-40b9-9dcc-97ef822c0339
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/ee45c33a-8747-4636-9908-ebb6075ee508
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/ee45c33a-8747-4636-9908-ebb6075ee508
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/f0071752-4c7a-42a8-b72f-aca91c1d45c4
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/f0071752-4c7a-42a8-b72f-aca91c1d45c4
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/f8301fbd-d592-494f-b13c-81208ace8a83
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/f8301fbd-d592-494f-b13c-81208ace8a83
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: claude/f9489dc4-2060-413d-82df-54e22a6cc5a5
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: claude/f9489dc4-2060-413d-82df-54e22a6cc5a5
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a138fb-9249-7fe0-86f2-862b8adc68ef
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a138fb-9249-7fe0-86f2-862b8adc68ef
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a13b21-7327-7c60-9006-edb3c90098ae
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a13b21-7327-7c60-9006-edb3c90098ae
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a13bb6-8ece-7533-bf70-cb6022b40029
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a13bb6-8ece-7533-bf70-cb6022b40029
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a13bb9-c9d4-74e0-8386-4967333fab42
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a13bb9-c9d4-74e0-8386-4967333fab42
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a13bbc-0eea-7293-a1da-e6b517ba3bc5
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a13bbc-0eea-7293-a1da-e6b517ba3bc5
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a13bbc-c0df-7e53-9383-6398cd4255a9
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a13bbc-c0df-7e53-9383-6398cd4255a9
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a13db1-fa52-7b83-bef0-9e0cbb5fed94
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a13db1-fa52-7b83-bef0-9e0cbb5fed94
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a13db4-79e3-73e3-b1b7-22d2f5b5b8cb
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a13db4-79e3-73e3-b1b7-22d2f5b5b8cb
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a13db6-157f-7013-8b8e-ee975fe00bbd
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a13db6-157f-7013-8b8e-ee975fe00bbd
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a13db6-8751-73b1-ac45-2cfd823cd708
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a13db6-8751-73b1-ac45-2cfd823cd708
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a148dd-d9dc-7b41-8ff5-37fb3d72ae46
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a148dd-d9dc-7b41-8ff5-37fb3d72ae46
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a148df-e9ee-7061-9d8d-4bd9f106aedf
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a148df-e9ee-7061-9d8d-4bd9f106aedf
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a148e1-33c7-7f21-8981-aa85a762079f
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a148e1-33c7-7f21-8981-aa85a762079f
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a148e2-2b49-7230-8d38-9b246eb10ef1
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a148e2-2b49-7230-8d38-9b246eb10ef1
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a148e2-90ae-7eb0-8e51-7f557d118b67
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a148e2-90ae-7eb0-8e51-7f557d118b67
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a14d52-f902-7160-ae71-1d6bd5ab83ff
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a14d52-f902-7160-ae71-1d6bd5ab83ff
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a14d54-ead2-7b20-9c53-1b5e41ff0a87
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a14d54-ead2-7b20-9c53-1b5e41ff0a87
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a14d55-c2d0-7f43-a126-be76a08cf111
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a14d55-c2d0-7f43-a126-be76a08cf111
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a14d56-3dca-75a2-948b-5e7e712adae8
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a14d56-3dca-75a2-948b-5e7e712adae8
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a1532b-6586-7fa1-8acb-3efd179b40c5
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a1532b-6586-7fa1-8acb-3efd179b40c5
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a1532e-045a-7482-b76e-ae908931f88d
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a1532e-045a-7482-b76e-ae908931f88d
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a1532f-46e0-7510-8cc6-a1f8ae0b00ca
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a1532f-46e0-7510-8cc6-a1f8ae0b00ca
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a15331-cc8d-70f1-8ac0-7b2aa1094989
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a15331-cc8d-70f1-8ac0-7b2aa1094989
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a15332-3230-7401-85f1-610944224225
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a15332-3230-7401-85f1-610944224225
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 0
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source_index: 0
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a15f85-cb74-7033-af90-14126531a090
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a15f85-cb74-7033-af90-14126531a090
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a15f88-5db9-7a80-aa27-fa67fe645fe8
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a15f88-5db9-7a80-aa27-fa67fe645fe8
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a15f89-4991-7003-8c9b-63e011e604d2
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a15f89-4991-7003-8c9b-63e011e604d2
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a15f89-c603-70b3-ac34-0bbcd830b55c
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a15f89-c603-70b3-ac34-0bbcd830b55c
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a1619e-1b2a-7433-bd4f-759515fc5331
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a1619e-1b2a-7433-bd4f-759515fc5331
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a1619f-b959-7ec2-8197-3d094d728cb6
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a1619f-b959-7ec2-8197-3d094d728cb6
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a161a0-f963-7761-a38b-3d9423e96e16
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a161a0-f963-7761-a38b-3d9423e96e16
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a161a1-6ba6-7490-b5ac-e6c898495122
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a161a1-6ba6-7490-b5ac-e6c898495122
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a16272-8090-73f2-ba81-77e28978e17c
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a16272-8090-73f2-ba81-77e28978e17c
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a16274-a97e-73b3-b06a-933e39ee1bcc
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a16274-a97e-73b3-b06a-933e39ee1bcc
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a16276-7782-7033-8ceb-4e2afe871c23
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a16276-7782-7033-8ceb-4e2afe871c23
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a16278-05e4-76a0-a7d1-bba7766466c4
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a16278-05e4-76a0-a7d1-bba7766466c4
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a16278-88f7-7140-941b-a5b6f5ef1974
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a16278-88f7-7140-941b-a5b6f5ef1974
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a16335-672a-7933-b4ea-2925ff2e5217
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a16335-672a-7933-b4ea-2925ff2e5217
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a16339-875a-7873-833a-823b635425ec
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a16339-875a-7873-833a-823b635425ec
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a1633a-95d7-7431-bc8d-a693237878ce
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a1633a-95d7-7431-bc8d-a693237878ce
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a1633b-26e2-7b72-a1b9-a13cb5613d82
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a1633b-26e2-7b72-a1b9-a13cb5613d82
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a16394-faae-7e72-80ce-b26bdb3e5085
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a16394-faae-7e72-80ce-b26bdb3e5085
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a16398-2b41-7de1-8e6e-8e137e7d3bb9
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a16398-2b41-7de1-8e6e-8e137e7d3bb9
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a1639a-1f98-77b0-b97e-07c0a29bb4f3
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a1639a-1f98-77b0-b97e-07c0a29bb4f3
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a1639a-b758-7140-9b2a-174cba0e9519
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a1639a-b758-7140-9b2a-174cba0e9519
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a165b1-7420-76d0-9388-7e1db75d6bbb
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a165b1-7420-76d0-9388-7e1db75d6bbb
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a165b3-9b97-74f1-a0f9-38b1ea01b74b
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a165b3-9b97-74f1-a0f9-38b1ea01b74b
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a165b4-777c-76d0-9bd9-a3cee9f18d7c
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a165b4-777c-76d0-9bd9-a3cee9f18d7c
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source: codex/01a165b4-f501-7a82-b548-af05786c8e99
                  start: "2026-10-25T12:00:36-07:00"
                  end: "2026-10-25T19:00:50Z"
                  branch: main-slot1
                  kind: ownership-only
                - source: codex/01a165b4-f501-7a82-b548-af05786c8e99
                  start: "2026-10-25T19:00:50Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 1
                  start: "2026-10-25T19:01:11.805Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 2
                  start: "2026-10-25T19:11:08.824Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 3
                  start: "2026-10-25T19:15:55.307Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 4
                  start: "2026-10-25T19:19:53.146Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 5
                  start: "2026-10-25T19:20:16.534Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 6
                  start: "2026-10-25T19:22:04.947Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 7
                  start: "2026-10-25T19:22:13.342Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 8
                  start: "2026-10-25T19:34:39.798Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
                - source_index: 9
                  start: "2026-10-25T19:35:06.344Z"
                  end: "2026-10-25T12:46:36-07:00"
                  kind: branch
      "n": 1
      timestamp: "2026-10-25T12:58:05-07:00"
      dispatched: "2026-10-25T12:52:14-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Malformed typed side documents bypass conformance validation
          detail: 'cmd/sdlc/validategate.go:220-240 returns false when YAML parsing fails and the fallback cannot recognize the declaration. Closed frontmatter containing `"type": [invalid` or `type : [invalid` passes validateChangedInstances with zero validator calls; unterminated typed frontmatter also passes. A temporary overlay test reproduced all three against pinned production code. ARCH-SECURE/ARCH-PURPOSE: distinguish absent declarations from malformed typed input, preserve visible refusal, and add regression tests covering the declaration syntax class.'
          family: typed-input-fail-closed
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: README update appears missing for side-doc validation and deployment
          detail: construct/vocabulary/side-doc.cue:5 introduces the side-doc validation noun and its required type/parent contract, while atlas/workflow/sdlc-binary.md:393 documents paired vocabulary/sdlc deployment. README.md is unchanged in the pinned range, despite the plan's explicit README task. Add the validation invocation, typed-versus-untyped attachment behavior, and paired deployment guidance to the user entry point.
          family: user-surface-documentation
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: Required historical side-document repair lacks completion evidence
          detail: 'The issue''s Done when at workshop/issues/000623-instance-conformance-gate-validates-issue-side-side-docs-as-issues-exempt-and.md:42 explicitly requires archiving parley.nvim''s leftover file. Its latest evidence at lines 101-106 still records a requested owner-run repair, and the repair checklist remains unchecked. ARCH-PURPOSE: before whole-issue close, retain the owner''s repair commit, matching source/destination blob identities, active-path removal, and corrected archived linkage.'
          family: acceptance-deliverable-completion
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#623 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-25T12:58:05-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `typed-input-fail-closed` (kind: defect) Malformed typed side documents bypass conformance validation
  cmd/sdlc/validategate.go:220-240 returns false when YAML parsing fails and the fallback cannot recognize the declaration. Closed frontmatter containing `"type": [invalid` or `type : [invalid` passes validateChangedInstances with zero validator calls; unterminated typed frontmatter also passes. A temporary overlay test reproduced all three against pinned production code. ARCH-SECURE/ARCH-PURPOSE: distinguish absent declarations from malformed typed input, preserve visible refusal, and add regression tests covering the declaration syntax class.
- **BR-2** [Important] `user-surface-documentation` (kind: defect) README update appears missing for side-doc validation and deployment
  construct/vocabulary/side-doc.cue:5 introduces the side-doc validation noun and its required type/parent contract, while atlas/workflow/sdlc-binary.md:393 documents paired vocabulary/sdlc deployment. README.md is unchanged in the pinned range, despite the plan's explicit README task. Add the validation invocation, typed-versus-untyped attachment behavior, and paired deployment guidance to the user entry point.
- **BR-3** [Important] `acceptance-deliverable-completion` (kind: defect) Required historical side-document repair lacks completion evidence
  The issue's Done when at workshop/issues/000623-instance-conformance-gate-validates-issue-side-side-docs-as-issues-exempt-and.md:42 explicitly requires archiving parley.nvim's leftover file. Its latest evidence at lines 101-106 still records a requested owner-run repair, and the repair checklist remains unchecked. ARCH-PURPOSE: before whole-issue close, retain the owner's repair commit, matching source/destination blob identities, active-path removal, and corrected archived linkage.

## Open findings

- **BR-1** [Critical] `typed-input-fail-closed` (kind: defect) Malformed typed side documents bypass conformance validation
- **BR-2** [Important] `user-surface-documentation` (kind: defect) README update appears missing for side-doc validation and deployment
- **BR-3** [Important] `acceptance-deliverable-completion` (kind: defect) Required historical side-document repair lacks completion evidence
