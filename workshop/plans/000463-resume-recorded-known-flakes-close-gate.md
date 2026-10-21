---
gate: boundary-review
issue: 463
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ffcd2dfa-1b1b-408e-899a-52af58613306
            repository: github.com/xianxu/ariadne
            issue_id: "000463"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/035c66f5d43c7223
            reviewed_head: 3817e5ce721cb875a1b7d154ff93329516268551
            slot: ariadne:3
            requested_at: "2026-10-21T18:05:30.678299Z"
          phase: finished
          pid: 49121
          started_at: "2026-10-21T18:05:30.777277Z"
          finished_at: "2026-10-21T18:08:33.95971Z"
          session_id: 01a15fa3-c627-7d42-aa18-a489ef653585
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 3817e5ce721cb875a1b7d154ff93329516268551
        tracker_head: 5b9f14406035b41e6bdaa3681f2d901e8b2dcc57
        branch_point: 6fd7dc750557b3f51bdefc5859ecd8f76027a9ab
        since: "2026-10-21T04:30:26-07:00"
        until: "2026-10-21T11:05:07-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T04:30:26-07:00"
              end: "2026-10-21T11:05:07-07:00"
              began: "2026-10-21T04:30:26-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6a2e2bbef93f77238828232c2bc7ef14fea9726407b6153fa21ccf3fdfa1834b
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 112292489
              windows:
                - start: "2026-10-21T04:30:26-07:00"
                  end: "2026-10-21T11:05:07-07:00"
            - harness: codex
              id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 7123903eccdf67807d202d957c193dbc868be47e816c35054a00ea94f963509a
              started: "2026-10-21T14:24:08.011Z"
              spawned: false
              bytes: 7654413
              windows:
                - start: "2026-10-21T14:24:08.011Z"
                  end: "2026-10-21T11:05:07-07:00"
            - harness: codex
              id: 01a15f86-6e56-7a21-9e02-00957cdf1673
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: c0df524de4599eb395f308258b3534e820cf0c4d15328ab393411b7cf51bed66
              started: "2026-10-21T17:33:28.177Z"
              spawned: true
              bytes: 741390
              windows:
                - start: "2026-10-21T17:33:28.177Z"
                  end: "2026-10-21T11:05:07-07:00"
            - harness: codex
              id: 01a15f8a-3fbc-7be2-9eac-270c02542bb1
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: e3f8fb4ee3f38671492de07cfedb5e468d55003edb65e36018febd064b6a8bbb
              started: "2026-10-21T17:37:38.685Z"
              spawned: false
              bytes: 645521
              windows:
                - start: "2026-10-21T17:37:38.685Z"
                  end: "2026-10-21T11:05:07-07:00"
            - harness: codex
              id: 01a15f8c-92d5-7c93-8304-6dbff1269850
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 3d4fc95794b1073430ce3218e925fd62df60bbb33614797fb650f55cc498ae7e
              started: "2026-10-21T17:40:10.946Z"
              spawned: false
              bytes: 493575
              windows:
                - start: "2026-10-21T17:40:10.946Z"
                  end: "2026-10-21T11:05:07-07:00"
            - harness: codex
              id: 01a15f8d-f1f4-7a81-b329-7833b350bc9a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6fa6ac8dcaded1a7f4b64977c4e685eae479761dac6e6b46357ab4d5c70b7d5c
              started: "2026-10-21T17:41:40.79Z"
              spawned: false
              bytes: 405882
              windows:
                - start: "2026-10-21T17:41:40.79Z"
                  end: "2026-10-21T11:05:07-07:00"
            - harness: codex
              id: 01a15f8e-6c73-7720-ba94-afd118432005
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 071e4919b294899562e1055438af6d30ea9c7824bad43ea65eb4fac32fecb7d1
              started: "2026-10-21T17:42:12.312Z"
              spawned: false
              bytes: 425870
              windows:
                - start: "2026-10-21T17:42:12.312Z"
                  end: "2026-10-21T11:05:07-07:00"
            - harness: codex
              id: 01a15f8f-9b72-7f33-be7b-b7ace1feffcc
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 8eab562605bec2536f9dd159c3c5ed028adcaca1f7531b2eb2741392e47d4981
              started: "2026-10-21T17:43:29.504Z"
              spawned: true
              bytes: 860438
              windows:
                - start: "2026-10-21T17:43:29.504Z"
                  end: "2026-10-21T11:05:07-07:00"
            - harness: codex
              id: 01a15f92-96cd-70c1-afe9-adf07468499d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 9aab2343084da6534fad0f8aa46ff107bc8640f12cb7457bde9cb76e87229597
              started: "2026-10-21T17:46:44.971Z"
              spawned: true
              bytes: 923639
              windows:
                - start: "2026-10-21T17:46:44.971Z"
                  end: "2026-10-21T11:05:07-07:00"
            - harness: codex
              id: 01a15f9d-1069-7d80-856e-485e528d6e64
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 15b192009beed780ac6928e17f09376cffbf6c63693638fcc008e0ceae6f0e4f
              started: "2026-10-21T17:58:11.455Z"
              spawned: true
              bytes: 564844
              windows:
                - start: "2026-10-21T17:58:11.455Z"
                  end: "2026-10-21T11:05:07-07:00"
        queue_waits:
            - attempt_id: bc29b3b2a562e6665944e1c21982e2ea
              slot: ariadne:3
              record_digest: a8f570f8066857b15e9f841ab6d8d012da34e3af27d1dddf318f7f1edac78136
              start: "2026-10-21T14:52:43.738627Z"
              end: "2026-10-21T15:05:28.843886Z"
              queue_seconds: 765.105259
              raw_seconds: 765.105259
              exact: true
            - attempt_id: 7b763ff7713502a1500f8ca24fce0e60
              slot: ariadne:3
              record_digest: a3360cca43eda55d15dcc097b7c4dfd1da6e310930e6aaf2ece7ff4586527331
              start: "2026-10-21T15:38:42.334429Z"
              end: "2026-10-21T15:54:24.407544Z"
              queue_seconds: 942.073115
              raw_seconds: 942.073115
              exact: true
        protected:
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:42:57.208Z"
              end: "2026-10-21T05:07:53-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T05:07:53-07:00"
              end: "2026-10-21T05:22:11-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:47:29.214Z"
              end: "2026-10-21T11:48:38.896Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:50:14.859Z"
              end: "2026-10-21T11:50:38.15Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:52:00.18Z"
              end: "2026-10-21T11:52:30.515Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:52:33.946Z"
              end: "2026-10-21T11:53:02.081Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T07:27:35-07:00"
              end: "2026-10-21T07:43:52-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T07:43:52-07:00"
              end: "2026-10-21T07:52:27-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T14:28:03.076Z"
              end: "2026-10-21T07:43:52-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T07:43:52-07:00"
              end: "2026-10-21T14:46:10.746Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T14:34:21.129Z"
              end: "2026-10-21T14:40:53.048Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T14:49:01.284Z"
              end: "2026-10-21T14:50:53.163Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T15:05:28.843886Z"
              end: "2026-10-21T15:19:38.003Z"
      "n": 1
      timestamp: "2026-10-21T11:08:34-07:00"
      dispatched: "2026-10-21T11:05:30-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: The publisher rejects every reconciliation receipt with an existing link
          detail: cmd/sdlc/localcheck_resume_store.go:239 calls s.proofError before reaching receipt publication. Every valid receipt requires an existing resume link, but teststats_evidence.go:177 rejects that linked parent, making the receipt branch unreachable. A temporary regression publishing a link followed by its valid receipt fails with “immutable execution fence requires validation.” Use publication-specific validation that permits the required link while preserving unsupported/conflicting-evidence checks, and add a regression through the real publisher. This blocks the promised terminal recovery path (ARCH-PURPOSE).
          family: proof-publication-reachability
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Crashed publication leaves .resume-* files permanently
          detail: cmd/sdlc/localcheck_resume_store.go:281 creates .resume-* files and relies exclusively on deferred removal. Process death bypasses that defer; teststats_store.go:588 recognizes only .attempt-* staging files and then skips non-JSON files. A temporary regression confirms that an abandoned 40-day-old .resume-* file survives production pruning. Add cleanup under the directory mutex, which excludes active publishers, and test interrupted publication cleanup. These files also escape the JSON record/byte capacity accounting (ARCH-FUNERAL).
          family: durable-staging-cleanup
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ed78b1e7-f8aa-4ec2-9329-f7ae0aa8e741
            repository: github.com/xianxu/ariadne
            issue_id: "000463"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/2d752b076d705416
            reviewed_head: 6caff439fe18afe6ac4ce3920f5e009137effa53
            slot: ariadne:3
            requested_at: "2026-10-21T18:13:55.953519Z"
          phase: finished
          pid: 92559
          started_at: "2026-10-21T18:13:56.048213Z"
          finished_at: "2026-10-21T18:17:21.535553Z"
          session_id: 01a15fab-7be6-7393-8ebd-cff2c80cf1cb
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 6caff439fe18afe6ac4ce3920f5e009137effa53
        tracker_head: 7eddb39383b32ec593a412a58d1ca4ffcbfa65e8
        branch_point: 6fd7dc750557b3f51bdefc5859ecd8f76027a9ab
        since: "2026-10-21T04:30:26-07:00"
        until: "2026-10-21T11:13:29-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T04:30:26-07:00"
              end: "2026-10-21T11:13:29-07:00"
              began: "2026-10-21T04:30:26-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6a2e2bbef93f77238828232c2bc7ef14fea9726407b6153fa21ccf3fdfa1834b
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 112292489
              windows:
                - start: "2026-10-21T04:30:26-07:00"
                  end: "2026-10-21T11:13:29-07:00"
            - harness: codex
              id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 744c59c4520f01ea449f66b1a63a5e8ae5743d91f711a853397ae4666c0b36b6
              started: "2026-10-21T14:24:08.011Z"
              spawned: false
              bytes: 8095775
              windows:
                - start: "2026-10-21T14:24:08.011Z"
                  end: "2026-10-21T11:13:29-07:00"
            - harness: codex
              id: 01a15f86-6e56-7a21-9e02-00957cdf1673
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: c0df524de4599eb395f308258b3534e820cf0c4d15328ab393411b7cf51bed66
              started: "2026-10-21T17:33:28.177Z"
              spawned: true
              bytes: 741390
              windows:
                - start: "2026-10-21T17:33:28.177Z"
                  end: "2026-10-21T11:13:29-07:00"
            - harness: codex
              id: 01a15f8a-3fbc-7be2-9eac-270c02542bb1
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: e3f8fb4ee3f38671492de07cfedb5e468d55003edb65e36018febd064b6a8bbb
              started: "2026-10-21T17:37:38.685Z"
              spawned: false
              bytes: 645521
              windows:
                - start: "2026-10-21T17:37:38.685Z"
                  end: "2026-10-21T11:13:29-07:00"
            - harness: codex
              id: 01a15f8c-92d5-7c93-8304-6dbff1269850
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 3d4fc95794b1073430ce3218e925fd62df60bbb33614797fb650f55cc498ae7e
              started: "2026-10-21T17:40:10.946Z"
              spawned: false
              bytes: 493575
              windows:
                - start: "2026-10-21T17:40:10.946Z"
                  end: "2026-10-21T11:13:29-07:00"
            - harness: codex
              id: 01a15f8d-f1f4-7a81-b329-7833b350bc9a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6fa6ac8dcaded1a7f4b64977c4e685eae479761dac6e6b46357ab4d5c70b7d5c
              started: "2026-10-21T17:41:40.79Z"
              spawned: false
              bytes: 405882
              windows:
                - start: "2026-10-21T17:41:40.79Z"
                  end: "2026-10-21T11:13:29-07:00"
            - harness: codex
              id: 01a15f8e-6c73-7720-ba94-afd118432005
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 071e4919b294899562e1055438af6d30ea9c7824bad43ea65eb4fac32fecb7d1
              started: "2026-10-21T17:42:12.312Z"
              spawned: false
              bytes: 425870
              windows:
                - start: "2026-10-21T17:42:12.312Z"
                  end: "2026-10-21T11:13:29-07:00"
            - harness: codex
              id: 01a15f8f-9b72-7f33-be7b-b7ace1feffcc
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 12d72e04ab906fbc85380bc75d65f59d5cf31b326e2302b1806381f6d493a5fb
              started: "2026-10-21T17:43:29.504Z"
              spawned: true
              bytes: 1012809
              windows:
                - start: "2026-10-21T17:43:29.504Z"
                  end: "2026-10-21T11:13:29-07:00"
            - harness: codex
              id: 01a15f92-96cd-70c1-afe9-adf07468499d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 9aab2343084da6534fad0f8aa46ff107bc8640f12cb7457bde9cb76e87229597
              started: "2026-10-21T17:46:44.971Z"
              spawned: true
              bytes: 923639
              windows:
                - start: "2026-10-21T17:46:44.971Z"
                  end: "2026-10-21T11:13:29-07:00"
            - harness: codex
              id: 01a15f9d-1069-7d80-856e-485e528d6e64
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 15b192009beed780ac6928e17f09376cffbf6c63693638fcc008e0ceae6f0e4f
              started: "2026-10-21T17:58:11.455Z"
              spawned: true
              bytes: 564844
              windows:
                - start: "2026-10-21T17:58:11.455Z"
                  end: "2026-10-21T11:13:29-07:00"
        queue_waits:
            - attempt_id: bc29b3b2a562e6665944e1c21982e2ea
              slot: ariadne:3
              record_digest: a8f570f8066857b15e9f841ab6d8d012da34e3af27d1dddf318f7f1edac78136
              start: "2026-10-21T14:52:43.738627Z"
              end: "2026-10-21T15:05:28.843886Z"
              queue_seconds: 765.105259
              raw_seconds: 765.105259
              exact: true
            - attempt_id: 7b763ff7713502a1500f8ca24fce0e60
              slot: ariadne:3
              record_digest: a3360cca43eda55d15dcc097b7c4dfd1da6e310930e6aaf2ece7ff4586527331
              start: "2026-10-21T15:38:42.334429Z"
              end: "2026-10-21T15:54:24.407544Z"
              queue_seconds: 942.073115
              raw_seconds: 942.073115
              exact: true
        protected:
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:42:57.208Z"
              end: "2026-10-21T05:07:53-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T05:07:53-07:00"
              end: "2026-10-21T05:22:11-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:47:29.214Z"
              end: "2026-10-21T11:48:38.896Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:50:14.859Z"
              end: "2026-10-21T11:50:38.15Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:52:00.18Z"
              end: "2026-10-21T11:52:30.515Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:52:33.946Z"
              end: "2026-10-21T11:53:02.081Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T07:27:35-07:00"
              end: "2026-10-21T07:43:52-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T07:43:52-07:00"
              end: "2026-10-21T07:52:27-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T14:28:03.076Z"
              end: "2026-10-21T07:43:52-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T07:43:52-07:00"
              end: "2026-10-21T14:46:10.746Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T14:34:21.129Z"
              end: "2026-10-21T14:40:53.048Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T14:49:01.284Z"
              end: "2026-10-21T14:50:53.163Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T15:05:28.843886Z"
              end: "2026-10-21T15:19:38.003Z"
        cuts:
            - boundary: M1
              dispatched: "2026-10-21T11:05:30-07:00"
              finalized: "2026-10-21T11:08:34-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-21T11:17:22-07:00"
      dispatched: "2026-10-21T11:13:55-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: localcheck_resume_store.go:249 now checks unsupported authority separately before validating the supported graph. TestResumeStorePublishesTerminalReceipt passes on head and fails with “immutable execution fence requires validation” when the old guard is restored.
          round: 2
        - id: BR-2
          disposition: addressed
          note: pruneResumeStaging runs under the directory mutex during publication and pruning. TestResumeStoreStagingLifetime and TestResumeStoreProcessStagingCrashCleanup pass on head; both fail with retained crash residue when cleanup is removed. The disposable checkout was removed.
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: 6caff439fe18afe6ac4ce3920f5e009137effa53
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a6f9d8aa-9cb7-4bd2-bd2c-9abd8f33c9bd
            repository: github.com/xianxu/ariadne
            issue_id: "000463"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/8e0428314e1b709d
            reviewed_head: 615fcd68453a0f37523fede1f6108f15ad6a4f59
            slot: ariadne:3
            requested_at: "2026-10-21T18:37:50.925332Z"
          phase: finished
          pid: 64878
          started_at: "2026-10-21T18:37:51.032273Z"
          finished_at: "2026-10-21T18:41:14.718736Z"
          session_id: 01a15fc1-6160-7231-9013-f4cb22100432
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 615fcd68453a0f37523fede1f6108f15ad6a4f59
        tracker_head: 7eddb39383b32ec593a412a58d1ca4ffcbfa65e8
        branch_point: 6fd7dc750557b3f51bdefc5859ecd8f76027a9ab
        since: "2026-10-21T04:30:26-07:00"
        until: "2026-10-21T11:37:12-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T04:30:26-07:00"
              end: "2026-10-21T11:37:12-07:00"
              began: "2026-10-21T04:30:26-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6a2e2bbef93f77238828232c2bc7ef14fea9726407b6153fa21ccf3fdfa1834b
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 112292489
              windows:
                - start: "2026-10-21T04:30:26-07:00"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 42ed00917c6cd5c76dbe1abbcf345011afd767c264dc9bb33d17aa64751e5333
              started: "2026-10-21T14:24:08.011Z"
              spawned: false
              bytes: 10066866
              windows:
                - start: "2026-10-21T14:24:08.011Z"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15ee2-6d4d-7563-bfbd-69a76a92de09
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 45f9c3170daa5b82a4c64c1595568c69b6db64129df184cb3bcfc06b3db965b7
              started: "2026-10-21T14:34:20.012Z"
              spawned: true
              bytes: 1293996
              windows:
                - start: "2026-10-21T14:34:20.012Z"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15f86-6e56-7a21-9e02-00957cdf1673
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: c0df524de4599eb395f308258b3534e820cf0c4d15328ab393411b7cf51bed66
              started: "2026-10-21T17:33:28.177Z"
              spawned: true
              bytes: 741390
              windows:
                - start: "2026-10-21T17:33:28.177Z"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15f8a-3fbc-7be2-9eac-270c02542bb1
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: e3f8fb4ee3f38671492de07cfedb5e468d55003edb65e36018febd064b6a8bbb
              started: "2026-10-21T17:37:38.685Z"
              spawned: false
              bytes: 645521
              windows:
                - start: "2026-10-21T17:37:38.685Z"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15f8c-92d5-7c93-8304-6dbff1269850
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 3d4fc95794b1073430ce3218e925fd62df60bbb33614797fb650f55cc498ae7e
              started: "2026-10-21T17:40:10.946Z"
              spawned: false
              bytes: 493575
              windows:
                - start: "2026-10-21T17:40:10.946Z"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15f8d-f1f4-7a81-b329-7833b350bc9a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6fa6ac8dcaded1a7f4b64977c4e685eae479761dac6e6b46357ab4d5c70b7d5c
              started: "2026-10-21T17:41:40.79Z"
              spawned: false
              bytes: 405882
              windows:
                - start: "2026-10-21T17:41:40.79Z"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15f8e-6c73-7720-ba94-afd118432005
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 071e4919b294899562e1055438af6d30ea9c7824bad43ea65eb4fac32fecb7d1
              started: "2026-10-21T17:42:12.312Z"
              spawned: false
              bytes: 425870
              windows:
                - start: "2026-10-21T17:42:12.312Z"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15f8f-9b72-7f33-be7b-b7ace1feffcc
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: db74c4405fc75460b4c014776f565a044767535d90f9a7f278a597e54b61185d
              started: "2026-10-21T17:43:29.504Z"
              spawned: true
              bytes: 1898357
              windows:
                - start: "2026-10-21T17:43:29.504Z"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15f92-96cd-70c1-afe9-adf07468499d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 159a36b5f0d3d2aed834e100dd132e8a6108674f4fb748bbf60e4c11046b6b90
              started: "2026-10-21T17:46:44.971Z"
              spawned: true
              bytes: 1805070
              windows:
                - start: "2026-10-21T17:46:44.971Z"
                  end: "2026-10-21T11:37:12-07:00"
            - harness: codex
              id: 01a15f9d-1069-7d80-856e-485e528d6e64
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 15b192009beed780ac6928e17f09376cffbf6c63693638fcc008e0ceae6f0e4f
              started: "2026-10-21T17:58:11.455Z"
              spawned: true
              bytes: 564844
              windows:
                - start: "2026-10-21T17:58:11.455Z"
                  end: "2026-10-21T11:37:12-07:00"
        queue_waits:
            - attempt_id: bc29b3b2a562e6665944e1c21982e2ea
              slot: ariadne:3
              record_digest: a8f570f8066857b15e9f841ab6d8d012da34e3af27d1dddf318f7f1edac78136
              start: "2026-10-21T14:52:43.738627Z"
              end: "2026-10-21T15:05:28.843886Z"
              queue_seconds: 765.105259
              raw_seconds: 765.105259
              exact: true
            - attempt_id: 7b763ff7713502a1500f8ca24fce0e60
              slot: ariadne:3
              record_digest: a3360cca43eda55d15dcc097b7c4dfd1da6e310930e6aaf2ece7ff4586527331
              start: "2026-10-21T15:38:42.334429Z"
              end: "2026-10-21T15:54:24.407544Z"
              queue_seconds: 942.073115
              raw_seconds: 942.073115
              exact: true
        protected:
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:42:57.208Z"
              end: "2026-10-21T05:07:53-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T05:07:53-07:00"
              end: "2026-10-21T05:22:11-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:47:29.214Z"
              end: "2026-10-21T11:48:38.896Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:50:14.859Z"
              end: "2026-10-21T11:50:38.15Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:52:00.18Z"
              end: "2026-10-21T11:52:30.515Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T11:52:33.946Z"
              end: "2026-10-21T11:53:02.081Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T07:27:35-07:00"
              end: "2026-10-21T07:43:52-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T07:43:52-07:00"
              end: "2026-10-21T07:52:27-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T14:28:03.076Z"
              end: "2026-10-21T07:43:52-07:00"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T07:43:52-07:00"
              end: "2026-10-21T14:46:10.746Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T14:34:21.129Z"
              end: "2026-10-21T14:40:53.048Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T14:49:01.284Z"
              end: "2026-10-21T14:50:53.163Z"
            - issue: "524"
              slot: ariadne:3
              start: "2026-10-21T15:05:28.843886Z"
              end: "2026-10-21T15:19:38.003Z"
        cuts:
            - boundary: M1
              dispatched: "2026-10-21T11:05:30-07:00"
              finalized: "2026-10-21T11:08:34-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-21T11:13:55-07:00"
              finalized: "2026-10-21T11:17:22-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-21T11:41:15-07:00"
      dispatched: "2026-10-21T11:37:50-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-3
          severity: Important
          title: Completed-child crashes permanently strand locked resume worktrees
          detail: 'cmd/sdlc/localcheck_resume.go:113 relies on deferred worktree retirement, but child publication at line 196 calls publishSuiteAttempt, which deletes the inflight owner binding (teststats_store.go:426). Killing the CLI during status publication therefore leaves a locked checkout without that binding. Repeated resume returns through line 89 without cleanup; reconcile rejects committed children at line 237; the stale-worktree sweeper skips locked inputs. A temporary production-CLI regression reproduced this: after killing at the post barrier, successful repeated resume and a sweep with the checkout aged 48 hours still retained it. ARCH-FUNERAL: this is the 2nd finding in family durable-staging-cleanup. Do NOT fix only this instance: state and enforce the rule that cleanup authority outlives every protected artifact, enumerate pre-link, pending, committed-child and reconciled windows, and provide crash-safe retirement for each without granting another execution or pass.'
          family: durable-staging-cleanup
          round: 3
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#463 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T11:08:34-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `proof-publication-reachability` (kind: defect) The publisher rejects every reconciliation receipt with an existing link
  cmd/sdlc/localcheck_resume_store.go:239 calls s.proofError before reaching receipt publication. Every valid receipt requires an existing resume link, but teststats_evidence.go:177 rejects that linked parent, making the receipt branch unreachable. A temporary regression publishing a link followed by its valid receipt fails with “immutable execution fence requires validation.” Use publication-specific validation that permits the required link while preserving unsupported/conflicting-evidence checks, and add a regression through the real publisher. This blocks the promised terminal recovery path (ARCH-PURPOSE).
- **BR-2** [Important] `durable-staging-cleanup` (kind: defect) Crashed publication leaves .resume-* files permanently
  cmd/sdlc/localcheck_resume_store.go:281 creates .resume-* files and relies exclusively on deferred removal. Process death bypasses that defer; teststats_store.go:588 recognizes only .attempt-* staging files and then skips non-JSON files. A temporary regression confirms that an abandoned 40-day-old .resume-* file survives production pruning. Add cleanup under the directory mutex, which excludes active publishers, and test interrupted publication cleanup. These files also escape the JSON record/byte capacity accounting (ARCH-FUNERAL).

## Round 2 — 2026-10-21T11:17:22-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — localcheck_resume_store.go:249 now checks unsupported authority separately before validating the supported graph. TestResumeStorePublishesTerminalReceipt passes on head and fails with “immutable execution fence requires validation” when the old guard is restored.
- BR-2 — addressed — pruneResumeStaging runs under the directory mutex during publication and pruning. TestResumeStoreStagingLifetime and TestResumeStoreProcessStagingCrashCleanup pass on head; both fail with retained crash residue when cleanup is removed. The disposable checkout was removed.

## Round 3 — 2026-10-21T11:41:15-07:00 (codex) — BLOCKED

### Raised

- **BR-3** [Important] `durable-staging-cleanup` (kind: defect) Completed-child crashes permanently strand locked resume worktrees
  cmd/sdlc/localcheck_resume.go:113 relies on deferred worktree retirement, but child publication at line 196 calls publishSuiteAttempt, which deletes the inflight owner binding (teststats_store.go:426). Killing the CLI during status publication therefore leaves a locked checkout without that binding. Repeated resume returns through line 89 without cleanup; reconcile rejects committed children at line 237; the stale-worktree sweeper skips locked inputs. A temporary production-CLI regression reproduced this: after killing at the post barrier, successful repeated resume and a sweep with the checkout aged 48 hours still retained it. ARCH-FUNERAL: this is the 2nd finding in family durable-staging-cleanup. Do NOT fix only this instance: state and enforce the rule that cleanup authority outlives every protected artifact, enumerate pre-link, pending, committed-child and reconciled windows, and provide crash-safe retirement for each without granting another execution or pass.

## Open findings

- **BR-3** [Important] `durable-staging-cleanup` (kind: defect) Completed-child crashes permanently strand locked resume worktrees
