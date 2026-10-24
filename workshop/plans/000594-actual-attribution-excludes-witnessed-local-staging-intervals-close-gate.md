---
gate: boundary-review
issue: 594
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 03b6fa5f-cf43-4fbf-96f5-87bcc85df82c
            repository: github.com/xianxu/ariadne
            issue_id: "000594"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/07bf2fc457c06052
            reviewed_head: 6c77a6d37dbd2451ea82f0f0494bf60f0b6f9fa7
            slot: ariadne:3
            requested_at: "2026-10-24T16:27:51.193187Z"
          phase: finished
          pid: 54651
          started_at: "2026-10-24T16:27:51.363338Z"
          finished_at: "2026-10-24T16:32:55.873585Z"
          session_id: 01a16ebd-73ca-70f0-9da3-fdbf9769a9fa
          outcome: returned
      actual_inputs:
        version: 3
        method: active-time-v3-branch-v1
        repository: github.com/xianxu/ariadne
        code_head: 6c77a6d37dbd2451ea82f0f0494bf60f0b6f9fa7
        tracker_head: 0d7e51814ce961e4453c70427535fc497804cd1f
        branch_point: 6caf95c6fcfdb977b62c347bc36ff3eec47c36b4
        since: "2026-10-24T08:46:27-07:00"
        until: "2026-10-24T09:22:45-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-24T08:46:27-07:00"
              end: "2026-10-24T09:22:45-07:00"
              began: "2026-10-24T08:46:27-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 7e8e860d60834fccb41a9ee63a896658f2b5c9b72af93b96671286fe55be8ed8
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 163529145
              windows:
                - start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
              dispatch_windows:
                - start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T09:22:45-07:00"
        branches:
            - root: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              path: /Users/xianxu/workspace/ariadne/.git/worktrees/ariadne2/logs/HEAD
              head: 6c77a6d37dbd2451ea82f0f0494bf60f0b6f9fa7
              branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
              digest: 2d68c580a8cc27ab352e82dd27f98860fde3bb6355032c34aec16321a4aeab9f
              bytes: 225621
              intervals:
                - source: claude/07001e07-aa9e-45ac-a366-a34947285d4b
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/07001e07-aa9e-45ac-a366-a34947285d4b
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: claude/0bee769e-759a-4645-add0-43e76a1823b3
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/0bee769e-759a-4645-add0-43e76a1823b3
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: claude/32511e55-cf5d-4bab-92b4-4f93fbd03d84
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/32511e55-cf5d-4bab-92b4-4f93fbd03d84
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: claude/8b775857-6abb-458a-93e4-863438c3504b
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/8b775857-6abb-458a-93e4-863438c3504b
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: claude/8ba32d66-cfe7-4e87-a0ce-7f1d82a68003
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/8ba32d66-cfe7-4e87-a0ce-7f1d82a68003
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: claude/95da5486-d3a1-4dcc-b424-606bbdd06194
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/95da5486-d3a1-4dcc-b424-606bbdd06194
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: claude/b8b4369a-5a88-4014-bd34-762ad8ca0a0b
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/b8b4369a-5a88-4014-bd34-762ad8ca0a0b
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: claude/b966b671-7eec-484a-93fc-930a54bdbc62
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/b966b671-7eec-484a-93fc-930a54bdbc62
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: claude/e2899d01-dec7-4f66-8693-ab58809cfdc0
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/e2899d01-dec7-4f66-8693-ab58809cfdc0
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: claude/e2f2ea1b-4750-419b-b6d9-dd31aef63b7d
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: claude/e2f2ea1b-4750-419b-b6d9-dd31aef63b7d
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a138fd-1ac0-7a83-826e-60623c226386
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a138fd-1ac0-7a83-826e-60623c226386
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a13938-a1e9-7e83-a898-adb79bfffa4f
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a13938-a1e9-7e83-a898-adb79bfffa4f
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1393e-9b07-7b22-aefe-7ed65c88f61a
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1393e-9b07-7b22-aefe-7ed65c88f61a
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a13940-6122-7a82-960b-8e3ce864a398
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a13940-6122-7a82-960b-8e3ce864a398
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a13941-20a2-7c90-bb43-4eacd46e1c2a
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a13941-20a2-7c90-bb43-4eacd46e1c2a
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a139a3-0d78-77e3-843f-4c495f197b78
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a139a3-0d78-77e3-843f-4c495f197b78
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a139a8-b4bd-7d53-b015-e33dd4773766
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a139a8-b4bd-7d53-b015-e33dd4773766
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a139aa-cfe5-7772-9017-95afa9631039
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a139aa-cfe5-7772-9017-95afa9631039
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a139ab-84b9-7631-8210-3a5ad2c2761a
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a139ab-84b9-7631-8210-3a5ad2c2761a
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a13b1c-bf44-71e2-81bd-5d24934b826e
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a13b1c-bf44-71e2-81bd-5d24934b826e
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a13b23-a09c-7d03-9c26-d94c5e01b8a5
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a13b23-a09c-7d03-9c26-d94c5e01b8a5
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a13b26-2c15-7842-8c20-78ad9adc63ca
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a13b26-2c15-7842-8c20-78ad9adc63ca
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a13b26-c5e7-7711-8e4a-fc316e11fce7
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a13b26-c5e7-7711-8e4a-fc316e11fce7
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a14074-981e-74f2-af8c-a6ddf477fc6d
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a14074-981e-74f2-af8c-a6ddf477fc6d
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a14079-9246-7ea3-91fb-c457e3886da9
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a14079-9246-7ea3-91fb-c457e3886da9
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1407a-af6f-7152-833b-a43cafac9231
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1407a-af6f-7152-833b-a43cafac9231
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1407b-1424-7c31-964f-c6839b6c5951
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1407b-1424-7c31-964f-c6839b6c5951
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a14367-9218-7872-89c3-f9a28c524834
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a14367-9218-7872-89c3-f9a28c524834
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a14610-8410-7663-9460-18fdcdccd3b7
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a14610-8410-7663-9460-18fdcdccd3b7
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a14612-6f7e-73e3-9e62-6e216ffbc42b
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a14612-6f7e-73e3-9e62-6e216ffbc42b
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a14613-017f-7ac0-8533-29575dd19183
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a14613-017f-7ac0-8533-29575dd19183
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a14630-09d3-7033-b75e-5c5236168d01
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a14630-09d3-7033-b75e-5c5236168d01
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a14630-adfa-7cb1-870c-a59856084acb
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a14630-adfa-7cb1-870c-a59856084acb
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a151b8-57ff-7551-be32-48930d0927c7
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a151b8-57ff-7551-be32-48930d0927c7
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a151b9-c33a-7220-be50-1faab4d207c0
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a151b9-c33a-7220-be50-1faab4d207c0
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a151ba-d5e9-78c1-aab9-a87098ddc125
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a151ba-d5e9-78c1-aab9-a87098ddc125
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a151bb-4195-78e2-97a7-b858be9075ec
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a151bb-4195-78e2-97a7-b858be9075ec
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15339-4997-7732-a34b-d1b10f63c233
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15339-4997-7732-a34b-d1b10f63c233
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1533c-911d-7a81-a6c2-c638e867866f
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1533c-911d-7a81-a6c2-c638e867866f
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15340-7c8a-71a2-9902-70db36be1854
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15340-7c8a-71a2-9902-70db36be1854
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15340-f6cb-7f03-932e-bf9d83fd583f
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15340-f6cb-7f03-932e-bf9d83fd583f
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1586d-3992-7242-ab2c-4427a72c88aa
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1586d-3992-7242-ab2c-4427a72c88aa
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15871-6263-7d92-8c75-64419f58cf1d
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15871-6263-7d92-8c75-64419f58cf1d
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15872-cb08-7721-aad1-06cfa3bdf4f5
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15872-cb08-7721-aad1-06cfa3bdf4f5
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15873-4a93-7760-99c7-2c400abf6c0a
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15873-4a93-7760-99c7-2c400abf6c0a
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15d03-86e1-7a71-a2ce-ef4bafdda938
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15d03-86e1-7a71-a2ce-ef4bafdda938
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15d05-4be6-7122-9be6-0cc770381dfe
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15d05-4be6-7122-9be6-0cc770381dfe
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15d06-5a0b-7be3-a48a-c58866001c81
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15d06-5a0b-7be3-a48a-c58866001c81
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15d06-d164-7dd0-9b43-b0e9ee056cd1
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15d06-d164-7dd0-9b43-b0e9ee056cd1
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15e49-a810-7030-882b-a3730b969269
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15e49-a810-7030-882b-a3730b969269
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15e4c-2f6f-7b32-8342-ddb6680a0213
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15e4c-2f6f-7b32-8342-ddb6680a0213
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15e4d-ca88-7cd2-99df-2aa22f58516a
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15e4d-ca88-7cd2-99df-2aa22f58516a
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15e4e-4de2-7cb2-84aa-011086f3ced2
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15e4e-4de2-7cb2-84aa-011086f3ced2
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15ed8-b1d1-7783-855a-322febf4f2fe
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15ed8-b1d1-7783-855a-322febf4f2fe
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15f8a-3fbc-7be2-9eac-270c02542bb1
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15f8a-3fbc-7be2-9eac-270c02542bb1
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15f8c-92d5-7c93-8304-6dbff1269850
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15f8c-92d5-7c93-8304-6dbff1269850
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15f8d-f1f4-7a81-b329-7833b350bc9a
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15f8d-f1f4-7a81-b329-7833b350bc9a
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15f8e-6c73-7720-ba94-afd118432005
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15f8e-6c73-7720-ba94-afd118432005
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15fe3-3ce9-7380-81ac-0733e2fc7fe3
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15fe3-3ce9-7380-81ac-0733e2fc7fe3
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a15fe3-b725-77f2-a403-1438ca779990
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a15fe3-b725-77f2-a403-1438ca779990
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a16228-99a6-7fa1-b81c-ce96606a3570
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a16228-99a6-7fa1-b81c-ce96606a3570
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1622b-9e56-77e3-9f46-1cdfdd4d1f74
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1622b-9e56-77e3-9f46-1cdfdd4d1f74
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1622d-d3a2-7533-ae5e-39ecf8100f7a
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1622d-d3a2-7533-ae5e-39ecf8100f7a
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1622e-518e-78d3-8696-6dc1300c5de2
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1622e-518e-78d3-8696-6dc1300c5de2
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a16317-99c2-7321-b65e-efad3a7a3868
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a16317-99c2-7321-b65e-efad3a7a3868
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a16319-c9db-7753-afde-d93fa736a5d9
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a16319-c9db-7753-afde-d93fa736a5d9
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1631b-7f81-7e30-8522-8d7348ace077
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1631b-7f81-7e30-8522-8d7348ace077
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1631c-0110-7b91-853c-5a66df171f44
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1631c-0110-7b91-853c-5a66df171f44
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a1654c-cfd0-7572-af52-de00856fd1e6
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a1654c-cfd0-7572-af52-de00856fd1e6
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a16551-f979-7fa3-ac0c-51d8dcae9374
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a16551-f979-7fa3-ac0c-51d8dcae9374
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a16553-ed37-7112-966f-bf4812180f38
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a16553-ed37-7112-966f-bf4812180f38
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
                - source: codex/01a16554-80c4-7241-8a61-38c69fe096d9
                  start: "2026-10-24T08:46:27-07:00"
                  end: "2026-10-24T15:46:38Z"
                  branch: main-slot3
                  kind: resting-excluded
                - source: codex/01a16554-80c4-7241-8a61-38c69fe096d9
                  start: "2026-10-24T15:46:38Z"
                  end: "2026-10-24T09:22:45-07:00"
                  branch: 000594-actual-attribution-excludes-witnessed-local-staging-intervals
                  kind: branch
      "n": 1
      timestamp: "2026-10-24T09:32:57-07:00"
      dispatched: "2026-10-24T09:27:51-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README update appears missing for witnessed staging exclusions
          detail: README.md:1206–1216 describes the actual/active-time attribution contract but omits the newly supported stage/* exclusion and its visible, unassigned branch-window minutes. README.md is unchanged in this range, while cmd/sdlc/helptext/actual.md:54 and active-time.md:85 document the changed behavior. This family's affected README consumers are actual/close and active-time, both covered by this paragraph. Update it with the concurrent-claim exclusion, unchanged sole-reservation fallback, and strict unknown/detached refusal.
          family: user-facing-contract-doc-parity
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: Core concepts tables omit PURE/INTEGRATION classifications
          detail: The plan's tables at lines 13 and 21 omit the required Kind column. Enumerated entries are IsStagingBranch, SelectBranchActivity, MeasurementBranchInterval, measurementBranchWarning/validateOwnedBranchProof, and Existing HEAD capture/decoder. Add explicit classifications; separate capture from decoding, since capture performs IO while decoding, warning rendering, and proof validation are pure.
          family: core-concept-classification
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#594 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T09:32:57-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-facing-contract-doc-parity` (kind: defect) README update appears missing for witnessed staging exclusions
  README.md:1206–1216 describes the actual/active-time attribution contract but omits the newly supported stage/* exclusion and its visible, unassigned branch-window minutes. README.md is unchanged in this range, while cmd/sdlc/helptext/actual.md:54 and active-time.md:85 document the changed behavior. This family's affected README consumers are actual/close and active-time, both covered by this paragraph. Update it with the concurrent-claim exclusion, unchanged sole-reservation fallback, and strict unknown/detached refusal.
- **BR-2** [Minor] `core-concept-classification` (kind: defect) Core concepts tables omit PURE/INTEGRATION classifications
  The plan's tables at lines 13 and 21 omit the required Kind column. Enumerated entries are IsStagingBranch, SelectBranchActivity, MeasurementBranchInterval, measurementBranchWarning/validateOwnedBranchProof, and Existing HEAD capture/decoder. Add explicit classifications; separate capture from decoding, since capture performs IO while decoding, warning rendering, and proof validation are pure.

## Open findings

- **BR-1** [Important] `user-facing-contract-doc-parity` (kind: defect) README update appears missing for witnessed staging exclusions
- **BR-2** [Minor] `core-concept-classification` (kind: defect) Core concepts tables omit PURE/INTEGRATION classifications
