---
gate: boundary-review
issue: 592
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 18c2d474-635e-4e18-82b6-b8b6ffc211eb
            repository: github.com/xianxu/ariadne
            issue_id: "000592"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/73f1a73dc6e3f297
            reviewed_head: 468ee2e4bcbd8df4c4c11b28f7bb781b9ac45a94
            slot: ariadne:5
            requested_at: "2026-10-24T13:17:37.172065Z"
          phase: finished
          pid: 74607
          started_at: "2026-10-24T13:17:37.306114Z"
          finished_at: "2026-10-24T13:19:14.296472Z"
          session_id: 8892a979-c8af-4f63-854d-0ff754996d41
          outcome: returned
      actual_inputs:
        version: 3
        method: active-time-v3-branch-v1
        repository: github.com/xianxu/ariadne
        code_head: 468ee2e4bcbd8df4c4c11b28f7bb781b9ac45a94
        tracker_head: bdccff146bfa3bd4df8124c914ac16ffced97eef
        branch_point: b3a1850018e1664c2dc38b21e7075a77fa748d8d
        since: "2026-10-24T04:45:20-07:00"
        until: "2026-10-24T05:13:49-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-24T04:45:20-07:00"
              end: "2026-10-24T05:13:49-07:00"
              began: "2026-10-24T04:45:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 5ba07377647149fb4c1307587de68a9b27f296903a792a3502eee49ce7f851ea
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 17485608
              windows:
                - start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T05:13:49-07:00"
              dispatch_windows:
                - start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T05:13:49-07:00"
        branches:
            - root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              path: /Users/xianxu/workspace/ariadne/.git/worktrees/ariadne4/logs/HEAD
              head: 468ee2e4bcbd8df4c4c11b28f7bb781b9ac45a94
              branch: 000592-flake-testqueuerealhelperpause
              digest: 544877c8d13a78d1aaae1f93201a90d1a65f805cb5059c5fb722d5897567aef9
              bytes: 128537
              intervals:
                - source: claude/03bd9c39-c2ac-4044-9095-3ae230ea12e1
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: claude/03bd9c39-c2ac-4044-9095-3ae230ea12e1
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: claude/2b1c35f1-9bee-4064-83d0-2a67ebaaf586
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: claude/2b1c35f1-9bee-4064-83d0-2a67ebaaf586
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136db-6cd1-7a31-87f3-271411613545
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136db-6cd1-7a31-87f3-271411613545
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136e3-0510-7ad2-b6e3-fc22d2393185
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136e3-0510-7ad2-b6e3-fc22d2393185
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136e5-856a-70f1-af77-586688329bf5
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136e5-856a-70f1-af77-586688329bf5
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136e7-5937-7813-92b3-81b395dd6373
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136e7-5937-7813-92b3-81b395dd6373
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136e8-0981-76e3-a40c-a39877b8de78
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136e8-0981-76e3-a40c-a39877b8de78
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e41-814f-7363-9ace-652b9e0c79db
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e41-814f-7363-9ace-652b9e0c79db
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e42-77b7-7362-8d9e-081b86faacd8
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e42-77b7-7362-8d9e-081b86faacd8
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e45-0e99-7c22-8c44-efccc86d13fd
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e45-0e99-7c22-8c44-efccc86d13fd
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e46-56b2-7e11-8b51-a1afe9e0b5ae
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e46-56b2-7e11-8b51-a1afe9e0b5ae
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e46-cfd5-7dc3-b742-855d83a3b01c
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e46-cfd5-7dc3-b742-855d83a3b01c
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13ef7-195b-7873-93c2-eab974272198
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13ef7-195b-7873-93c2-eab974272198
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13ef8-6498-7093-9ea5-e3860d6eb2dc
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13ef8-6498-7093-9ea5-e3860d6eb2dc
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13ef8-d531-7bb3-8af6-4b9b02a2a7b1
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13ef8-d531-7bb3-8af6-4b9b02a2a7b1
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13efa-087b-7371-ba41-bd3cf7ae4281
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13efa-087b-7371-ba41-bd3cf7ae4281
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13efa-81b5-7583-a528-ca503a8de772
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13efa-81b5-7583-a528-ca503a8de772
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a146c6-7afd-7f20-953b-a22feb894466
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a146c6-7afd-7f20-953b-a22feb894466
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a146c7-9d2c-7ad2-8f7a-84a895b8235d
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a146c7-9d2c-7ad2-8f7a-84a895b8235d
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a146c8-cac0-7591-8d7e-d5e067bc1582
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a146c8-cac0-7591-8d7e-d5e067bc1582
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a146c9-2fd7-76f1-80fa-e78d01c5de6a
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a146c9-2fd7-76f1-80fa-e78d01c5de6a
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a149da-2cf6-78a0-bb82-2002cb7aa24f
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a149da-2cf6-78a0-bb82-2002cb7aa24f
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a149dd-61f3-7f70-949d-c3f57e4db6d8
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a149dd-61f3-7f70-949d-c3f57e4db6d8
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a149de-3582-74b2-8e00-383b44721aef
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a149de-3582-74b2-8e00-383b44721aef
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a149de-941b-7e10-816b-37abe3202d36
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a149de-941b-7e10-816b-37abe3202d36
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a15a09-0f59-7b93-ad53-4f28144d42d9
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a15a09-0f59-7b93-ad53-4f28144d42d9
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a1605e-8315-77a1-bf79-667826f2aa0e
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a1605e-8315-77a1-bf79-667826f2aa0e
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16060-ab56-7fd0-aa39-0a388d2dcde3
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16060-ab56-7fd0-aa39-0a388d2dcde3
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16062-3f0a-7fb3-ba73-0663ba2ed7ae
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16062-3f0a-7fb3-ba73-0663ba2ed7ae
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16062-a981-70c3-9e8e-eb45dbf8ac5d
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16062-a981-70c3-9e8e-eb45dbf8ac5d
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a161f6-e0be-7a71-95c6-b39b5659fc6d
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a161f6-e0be-7a71-95c6-b39b5659fc6d
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a161f8-89af-7061-9d69-1b6b31e0af42
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a161f8-89af-7061-9d69-1b6b31e0af42
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a161f9-65c5-74a2-b7de-507a69d2553c
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a161f9-65c5-74a2-b7de-507a69d2553c
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a161f9-c28e-76d2-8a85-5ed3bff58c1f
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a161f9-c28e-76d2-8a85-5ed3bff58c1f
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16201-e130-75f1-91de-47f8e68f03d4
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16201-e130-75f1-91de-47f8e68f03d4
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16202-3630-72a2-b708-c8663e97194c
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16202-3630-72a2-b708-c8663e97194c
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16569-2ac8-7262-8166-f38064a3678e
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16569-2ac8-7262-8166-f38064a3678e
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a1656b-d8d4-7f62-b4e0-d964ce2b3bff
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a1656b-d8d4-7f62-b4e0-d964ce2b3bff
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a1656d-6400-7a00-8f94-942f722b558c
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a1656d-6400-7a00-8f94-942f722b558c
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a1656e-33d9-7091-8b81-5f329965ee2e
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a1656e-33d9-7091-8b81-5f329965ee2e
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T05:13:49-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
      "n": 1
      timestamp: "2026-10-24T06:19:14-07:00"
      dispatched: "2026-10-24T06:17:37-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: 'ARCH-DRY: committed/committed_read and the shim PYTHONPATH setup are duplicated across the two test scripts'
          detail: 'Instances: test-lock-exit.test.py:43-64 and test-lock-pause.test.py:97-118 are word-for-word copies apart from HELPER.try_lock vs lock.try_lock. The shim PYTHONPATH construction appears at test-lock-exit.test.py:110-112 and test-lock-pause.test.py:256-257. Consider one shared fixture module under scripts/test/fixtures if more suites adopt locked reads.'
          family: duplicated-test-fixture-helper
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 468ee2e4bcbd8df4c4c11b28f7bb781b9ac45a94
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 02ac161e-4e88-448b-848d-437be980c141
            repository: github.com/xianxu/ariadne
            issue_id: "000592"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/2210dddbfe1a19d7
            reviewed_head: 6eb2f26b668f5a19d8dbcf0b15c87ec2d36adfba
            slot: ariadne:5
            requested_at: "2026-10-24T15:34:23.110993Z"
          phase: finished
          pid: 9839
          started_at: "2026-10-24T15:34:23.231717Z"
          finished_at: "2026-10-24T15:35:13.012775Z"
          session_id: 4b50d2ed-2439-44f8-82c6-45bc05458031
          outcome: returned
      actual_inputs:
        version: 3
        method: active-time-v3-branch-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 6eb2f26b668f5a19d8dbcf0b15c87ec2d36adfba
        tracker_head: 8794c0e4eccfdc6d99d2088ce64d85f93992b394
        branch_point: b3a1850018e1664c2dc38b21e7075a77fa748d8d
        since: "2026-10-24T04:45:20-07:00"
        until: "2026-10-24T06:22:23-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-24T04:45:20-07:00"
              end: "2026-10-24T06:22:23-07:00"
              began: "2026-10-24T04:45:20-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 4107f901968e55b78607ffc17aeb0d95c938e902498c2ca768fec2737492c640
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 17900698
              windows:
                - start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T06:22:23-07:00"
              dispatch_windows:
                - start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T06:22:23-07:00"
        branches:
            - root: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              path: /Users/xianxu/workspace/ariadne/.git/worktrees/ariadne4/logs/HEAD
              head: 6eb2f26b668f5a19d8dbcf0b15c87ec2d36adfba
              branch: 000592-flake-testqueuerealhelperpause
              digest: 1b2d04f42fde042d7389e8060ed46d020f15313d3664bab6fdb6b75e4e2e0da2
              bytes: 129024
              intervals:
                - source: claude/03bd9c39-c2ac-4044-9095-3ae230ea12e1
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: claude/03bd9c39-c2ac-4044-9095-3ae230ea12e1
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: claude/2b1c35f1-9bee-4064-83d0-2a67ebaaf586
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: claude/2b1c35f1-9bee-4064-83d0-2a67ebaaf586
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136db-6cd1-7a31-87f3-271411613545
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136db-6cd1-7a31-87f3-271411613545
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136e3-0510-7ad2-b6e3-fc22d2393185
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136e3-0510-7ad2-b6e3-fc22d2393185
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136e5-856a-70f1-af77-586688329bf5
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136e5-856a-70f1-af77-586688329bf5
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136e7-5937-7813-92b3-81b395dd6373
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136e7-5937-7813-92b3-81b395dd6373
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a136e8-0981-76e3-a40c-a39877b8de78
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a136e8-0981-76e3-a40c-a39877b8de78
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e41-814f-7363-9ace-652b9e0c79db
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e41-814f-7363-9ace-652b9e0c79db
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e42-77b7-7362-8d9e-081b86faacd8
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e42-77b7-7362-8d9e-081b86faacd8
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e45-0e99-7c22-8c44-efccc86d13fd
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e45-0e99-7c22-8c44-efccc86d13fd
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e46-56b2-7e11-8b51-a1afe9e0b5ae
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e46-56b2-7e11-8b51-a1afe9e0b5ae
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13e46-cfd5-7dc3-b742-855d83a3b01c
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13e46-cfd5-7dc3-b742-855d83a3b01c
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13ef7-195b-7873-93c2-eab974272198
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13ef7-195b-7873-93c2-eab974272198
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13ef8-6498-7093-9ea5-e3860d6eb2dc
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13ef8-6498-7093-9ea5-e3860d6eb2dc
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13ef8-d531-7bb3-8af6-4b9b02a2a7b1
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13ef8-d531-7bb3-8af6-4b9b02a2a7b1
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13efa-087b-7371-ba41-bd3cf7ae4281
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13efa-087b-7371-ba41-bd3cf7ae4281
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a13efa-81b5-7583-a528-ca503a8de772
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a13efa-81b5-7583-a528-ca503a8de772
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a146c6-7afd-7f20-953b-a22feb894466
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a146c6-7afd-7f20-953b-a22feb894466
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a146c7-9d2c-7ad2-8f7a-84a895b8235d
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a146c7-9d2c-7ad2-8f7a-84a895b8235d
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a146c8-cac0-7591-8d7e-d5e067bc1582
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a146c8-cac0-7591-8d7e-d5e067bc1582
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a146c9-2fd7-76f1-80fa-e78d01c5de6a
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a146c9-2fd7-76f1-80fa-e78d01c5de6a
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a149da-2cf6-78a0-bb82-2002cb7aa24f
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a149da-2cf6-78a0-bb82-2002cb7aa24f
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a149dd-61f3-7f70-949d-c3f57e4db6d8
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a149dd-61f3-7f70-949d-c3f57e4db6d8
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a149de-3582-74b2-8e00-383b44721aef
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a149de-3582-74b2-8e00-383b44721aef
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a149de-941b-7e10-816b-37abe3202d36
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a149de-941b-7e10-816b-37abe3202d36
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a15a09-0f59-7b93-ad53-4f28144d42d9
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a15a09-0f59-7b93-ad53-4f28144d42d9
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a1605e-8315-77a1-bf79-667826f2aa0e
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a1605e-8315-77a1-bf79-667826f2aa0e
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16060-ab56-7fd0-aa39-0a388d2dcde3
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16060-ab56-7fd0-aa39-0a388d2dcde3
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16062-3f0a-7fb3-ba73-0663ba2ed7ae
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16062-3f0a-7fb3-ba73-0663ba2ed7ae
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16062-a981-70c3-9e8e-eb45dbf8ac5d
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16062-a981-70c3-9e8e-eb45dbf8ac5d
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a161f6-e0be-7a71-95c6-b39b5659fc6d
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a161f6-e0be-7a71-95c6-b39b5659fc6d
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a161f8-89af-7061-9d69-1b6b31e0af42
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a161f8-89af-7061-9d69-1b6b31e0af42
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a161f9-65c5-74a2-b7de-507a69d2553c
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a161f9-65c5-74a2-b7de-507a69d2553c
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a161f9-c28e-76d2-8a85-5ed3bff58c1f
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a161f9-c28e-76d2-8a85-5ed3bff58c1f
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16201-e130-75f1-91de-47f8e68f03d4
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16201-e130-75f1-91de-47f8e68f03d4
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16202-3630-72a2-b708-c8663e97194c
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16202-3630-72a2-b708-c8663e97194c
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a16569-2ac8-7262-8166-f38064a3678e
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a16569-2ac8-7262-8166-f38064a3678e
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a1656b-d8d4-7f62-b4e0-d964ce2b3bff
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a1656b-d8d4-7f62-b4e0-d964ce2b3bff
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a1656d-6400-7a00-8f94-942f722b558c
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a1656d-6400-7a00-8f94-942f722b558c
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
                - source: codex/01a1656e-33d9-7091-8b81-5f329965ee2e
                  start: "2026-10-24T04:45:20-07:00"
                  end: "2026-10-24T11:45:36Z"
                  branch: main-slot5
                  kind: ownership-only
                - source: codex/01a1656e-33d9-7091-8b81-5f329965ee2e
                  start: "2026-10-24T11:45:36Z"
                  end: "2026-10-24T06:22:23-07:00"
                  branch: 000592-flake-testqueuerealhelperpause
                  kind: branch
        queue_waits:
            - attempt_id: dbc52a9d719e0d1ba2598126794860dc
              slot: ariadne:5
              record_digest: 47bb0e8533c5769ee0332af7ac9913f61632d66e8ce79f9ade268fd2862fc2d6
              start: "2026-10-24T12:13:57.775432Z"
              end: "2026-10-24T12:51:43.752968Z"
              queue_seconds: 2265.977536
              raw_seconds: 2265.977536
              exact: true
        cuts:
            - dispatched: "2026-10-24T06:17:37-07:00"
              finalized: "2026-10-24T06:19:14-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-24T08:35:13-07:00"
      dispatched: "2026-10-24T08:34:23-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: committed, committed_read and slow_writes moved to scripts/test/fixtures/queue_reads.py; both scripts delegate to it (exit:46-50,97; pause:100-104,241); same flock LOCK_NB semantics as try_lock; exit 11 OK and pause 19 OK at head.
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: 6eb2f26b668f5a19d8dbcf0b15c87ec2d36adfba
      blocked: false
---

# Gate ledger — ariadne#592 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T06:19:14-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `duplicated-test-fixture-helper` (kind: defect) ARCH-DRY: committed/committed_read and the shim PYTHONPATH setup are duplicated across the two test scripts
  Instances: test-lock-exit.test.py:43-64 and test-lock-pause.test.py:97-118 are word-for-word copies apart from HELPER.try_lock vs lock.try_lock. The shim PYTHONPATH construction appears at test-lock-exit.test.py:110-112 and test-lock-pause.test.py:256-257. Consider one shared fixture module under scripts/test/fixtures if more suites adopt locked reads.

## Round 2 — 2026-10-24T08:35:13-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — committed, committed_read and slow_writes moved to scripts/test/fixtures/queue_reads.py; both scripts delegate to it (exit:46-50,97; pause:100-104,241); same flock LOCK_NB semantics as try_lock; exit 11 OK and pause 19 OK at head.

## Open findings

(none — every finding has been disposed)
