---
gate: boundary-review
issue: 385
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f19d47bd-d5af-41fd-8399-853500dce3c4
            repository: github.com/xianxu/ariadne
            issue_id: "000385"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/1f99f510016bcc98
            reviewed_head: d4c9d29701d73e39a2718e22398b484d4c694460
            slot: ariadne:3
            requested_at: "2026-10-15T18:42:18.468641Z"
          phase: finished
          pid: 38476
          started_at: "2026-10-15T18:42:18.70588Z"
          finished_at: "2026-10-15T18:45:11.374353Z"
          session_id: 01a140df-4f76-7342-8669-a30a2b5886ed
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d4c9d29701d73e39a2718e22398b484d4c694460
        tracker_head: 21ffc918fcf1b9a62d2bb492ea5324405730ae1b
        branch_point: 9d03c5735e15a06f70e552dfcbcdd21f4d52e343
        since: "2026-10-15T09:33:04-07:00"
        until: "2026-10-15T11:41:48-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T09:33:04-07:00"
              end: "2026-10-15T11:41:48-07:00"
              began: "2026-10-15T09:33:04-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a138fd-1ac0-7a83-826e-60623c226386
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c5684de9a8f25398e32a10b0274bfab9f3f1eac2099a79b2cd1351bb3a7f1f36
              started: "2026-10-14T05:58:30.184Z"
              spawned: false
              bytes: 41949541
              windows:
                - start: "2026-10-15T09:33:04-07:00"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a14074-981e-74f2-af8c-a6ddf477fc6d
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a9886dc5ea07325421724dbd964a263d52c38a01666ce05f66b4eb483be6e957
              started: "2026-10-15T16:45:47.228Z"
              spawned: false
              bytes: 758658
              windows:
                - start: "2026-10-15T16:45:47.228Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a14079-9246-7ea3-91fb-c457e3886da9
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 1f1689dceec0617335228ae3ffa7158942cc8f392d39eed6bd1e8e0323cd5c72
              started: "2026-10-15T16:51:14.597Z"
              spawned: false
              bytes: 453960
              windows:
                - start: "2026-10-15T16:51:14.597Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407a-af6f-7152-833b-a43cafac9231
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 467ba6fd9787954503097fe5d6f76e5f3eacf111156a527937edb4a48121c222
              started: "2026-10-15T16:52:26.427Z"
              spawned: false
              bytes: 462233
              windows:
                - start: "2026-10-15T16:52:26.427Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407b-1424-7c31-964f-c6839b6c5951
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 9bfc53e280ec295615d2f254082bc30c1f99379efc2f80040203fcef658d93a0
              started: "2026-10-15T16:52:54.44Z"
              spawned: false
              bytes: 420615
              windows:
                - start: "2026-10-15T16:52:54.44Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407b-ee59-71d1-98d5-f8a27ea6615a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 9efef5aa5f9e50eef2d508357bf71aee275445db8b2a49fb239a8fb0e8da70a8
              started: "2026-10-15T16:53:46.292Z"
              spawned: true
              bytes: 611271
              windows:
                - start: "2026-10-15T16:53:46.292Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407d-b788-7921-97ed-2aa273e67570
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 23ed2409e504187554c6d6aa4eba3868f59c74027970bb50e97cb1bd40c15e76
              started: "2026-10-15T16:55:43.336Z"
              spawned: true
              bytes: 806710
              windows:
                - start: "2026-10-15T16:55:43.336Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a1407e-0e37-7f82-9b51-1e2425d58944
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 55afe6e1788546348bdf2b561bae29401bfb9ab0da7a8b1288c2e96b41876b19
              started: "2026-10-15T16:56:05.521Z"
              spawned: true
              bytes: 1268169
              windows:
                - start: "2026-10-15T16:56:05.521Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a14080-72b2-7e43-bc09-63ffe1e09102
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: e1a877e295cf76c410c5c74c7a817e582583faf1e507b98fe977b821fe7d9247
              started: "2026-10-15T16:58:42.319Z"
              spawned: true
              bytes: 1553084
              windows:
                - start: "2026-10-15T16:58:42.319Z"
                  end: "2026-10-15T11:41:48-07:00"
            - harness: codex
              id: 01a140bb-446b-78a2-bad0-8a87ab45e0d6
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: bea2c41d92607fceb2ea2d50a0cf2458e402a022f020adbe379546da89cb68ea
              started: "2026-10-15T18:02:57.072Z"
              spawned: true
              bytes: 811720
              windows:
                - start: "2026-10-15T18:02:57.072Z"
                  end: "2026-10-15T11:41:48-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.03448333333333333
      "n": 1
      timestamp: "2026-10-15T11:45:12-07:00"
      dispatched: "2026-10-15T11:42:18-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Important
          title: Qoder nonzero token counters lack the promised format diagnostic
          detail: cmd/sdlc/internal/usage/qoder.go:117 ignores the parsed token fields. A temporary overlay test with input_tokens=123 and output_tokens=456 returned null tokens but only sources/records coverage and spend_complete=true. The plan at lines 137–140 explicitly requires a visible diagnostic for unrecognized counter semantics. Validate all four placeholder fields, retain null tokens, and add regression coverage for nonzero and invalid values (ARCH-SECURE, ARCH-PURPOSE).
          family: unsupported-format-visibility
          round: 1
        - id: BR-2
          severity: Important
          title: README omits Qoder usage and links to a removed atlas anchor
          detail: README.md:130–145 still describes Claude/Codex tokens only, omitting Qoder credit reporting, calibration support, and schema-version-2 nullable measurements. Its local-token-usage-365 link no longer matches the renamed atlas heading. Update this passage and link in the same boundary, as required by the README docs gate (ARCH-PURPOSE).
          family: user-surface-documentation
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: c6262c8f-db3d-4050-afb1-07d45375f688
            repository: github.com/xianxu/ariadne
            issue_id: "000385"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a9751bcd78f9e3a6
            reviewed_head: 0ffe2ebd220d7aca981c81bbad63697e45dc9b2f
            slot: ariadne:3
            requested_at: "2026-10-15T19:23:40.725396Z"
          phase: finished
          pid: 42027
          started_at: "2026-10-15T19:23:40.827993Z"
          finished_at: "2026-10-15T19:25:58.415464Z"
          session_id: 01a14105-2e87-7582-99cf-85f33824b443
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 0ffe2ebd220d7aca981c81bbad63697e45dc9b2f
        tracker_head: c8dae7145ab83fa68c863c4626b237609adc7934
        branch_point: ea2cee461c058080a25b03e689e9142c3ea21a63
        since: "2026-10-15T09:33:04-07:00"
        until: "2026-10-15T12:23:26-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T09:33:04-07:00"
              end: "2026-10-15T12:23:26-07:00"
              began: "2026-10-15T09:33:04-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a138fd-1ac0-7a83-826e-60623c226386
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: aeb23d80e3b52e12e0207614e0577870ff803ba07898417cfd9bac47d95cbd98
              started: "2026-10-14T05:58:30.184Z"
              spawned: false
              bytes: 42852198
              windows:
                - start: "2026-10-15T09:33:04-07:00"
                  end: "2026-10-15T12:23:26-07:00"
            - harness: codex
              id: 01a14074-981e-74f2-af8c-a6ddf477fc6d
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a9886dc5ea07325421724dbd964a263d52c38a01666ce05f66b4eb483be6e957
              started: "2026-10-15T16:45:47.228Z"
              spawned: false
              bytes: 758658
              windows:
                - start: "2026-10-15T16:45:47.228Z"
                  end: "2026-10-15T12:23:26-07:00"
            - harness: codex
              id: 01a14079-9246-7ea3-91fb-c457e3886da9
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 1f1689dceec0617335228ae3ffa7158942cc8f392d39eed6bd1e8e0323cd5c72
              started: "2026-10-15T16:51:14.597Z"
              spawned: false
              bytes: 453960
              windows:
                - start: "2026-10-15T16:51:14.597Z"
                  end: "2026-10-15T12:23:26-07:00"
            - harness: codex
              id: 01a1407a-af6f-7152-833b-a43cafac9231
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 467ba6fd9787954503097fe5d6f76e5f3eacf111156a527937edb4a48121c222
              started: "2026-10-15T16:52:26.427Z"
              spawned: false
              bytes: 462233
              windows:
                - start: "2026-10-15T16:52:26.427Z"
                  end: "2026-10-15T12:23:26-07:00"
            - harness: codex
              id: 01a1407b-1424-7c31-964f-c6839b6c5951
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 9bfc53e280ec295615d2f254082bc30c1f99379efc2f80040203fcef658d93a0
              started: "2026-10-15T16:52:54.44Z"
              spawned: false
              bytes: 420615
              windows:
                - start: "2026-10-15T16:52:54.44Z"
                  end: "2026-10-15T12:23:26-07:00"
            - harness: codex
              id: 01a1407b-ee59-71d1-98d5-f8a27ea6615a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 9efef5aa5f9e50eef2d508357bf71aee275445db8b2a49fb239a8fb0e8da70a8
              started: "2026-10-15T16:53:46.292Z"
              spawned: true
              bytes: 611271
              windows:
                - start: "2026-10-15T16:53:46.292Z"
                  end: "2026-10-15T12:23:26-07:00"
            - harness: codex
              id: 01a1407d-b788-7921-97ed-2aa273e67570
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 23ed2409e504187554c6d6aa4eba3868f59c74027970bb50e97cb1bd40c15e76
              started: "2026-10-15T16:55:43.336Z"
              spawned: true
              bytes: 806710
              windows:
                - start: "2026-10-15T16:55:43.336Z"
                  end: "2026-10-15T12:23:26-07:00"
            - harness: codex
              id: 01a1407e-0e37-7f82-9b51-1e2425d58944
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: a2df7d0ece42d275e671ab4812dd242df37c8008a1493a1217993701c48e7beb
              started: "2026-10-15T16:56:05.521Z"
              spawned: true
              bytes: 1414506
              windows:
                - start: "2026-10-15T16:56:05.521Z"
                  end: "2026-10-15T12:23:26-07:00"
            - harness: codex
              id: 01a14080-72b2-7e43-bc09-63ffe1e09102
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: e1a877e295cf76c410c5c74c7a817e582583faf1e507b98fe977b821fe7d9247
              started: "2026-10-15T16:58:42.319Z"
              spawned: true
              bytes: 1553084
              windows:
                - start: "2026-10-15T16:58:42.319Z"
                  end: "2026-10-15T12:23:26-07:00"
            - harness: codex
              id: 01a140bb-446b-78a2-bad0-8a87ab45e0d6
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: bea2c41d92607fceb2ea2d50a0cf2458e402a022f020adbe379546da89cb68ea
              started: "2026-10-15T18:02:57.072Z"
              spawned: true
              bytes: 811720
              windows:
                - start: "2026-10-15T18:02:57.072Z"
                  end: "2026-10-15T12:23:26-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.03448333333333333
        cuts:
            - dispatched: "2026-10-15T11:42:18-07:00"
              finalized: "2026-10-15T11:45:12-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-15T12:26:00-07:00"
      dispatched: "2026-10-15T12:23:40-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: qoder.go checks all four placeholders, including streaming rows. Regression tests cover nonzero and invalid values; removing the check through a temporary overlay makes the placeholder and streaming tests fail.
          round: 2
        - id: BR-2
          disposition: addressed
          note: README.md:130–149 now documents Qoder credits, calibration and schema-version-2 nullable measurements, and links to the matching Local usage heading in atlas/workflow/sdlc-binary.md. The collector and calibration implementation support the documented behavior.
          round: 2
      findings:
        - id: BR-3
          severity: Critical
          title: Malformed prefixes allow close to adopt incomplete Qoder actuals
          detail: 'cmd/sdlc/internal/transcripts/qoder.go:119–121 returns on a malformed row before later workspace metadata can establish relevance. discovery.go:54–59 then treats the failed source as skippable. A temporary extension of TestQoderActualDiscoveryRefusesPartialOwnedEvidence with "not-json\n" before valid owned rows returned the good source with no error. A production actual/close overlay confirmed actualMeasured, 0.1 hours and adopted=true despite that omitted owned source. Preserve uncertainty until exclusion is proven: continue bounded relevance discovery after recoverable parse failures, and refuse automatic adoption when scan failures or limits leave ownership unresolved. Sweep parse, read and bound failures before ownership resolution, with discovery and close regressions (ARCH-SECURE, ARCH-ORDER, ARCH-PURPOSE).'
          family: incomplete-evidence-adoption
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: deb475f4-837f-428e-b764-9569698eb952
            repository: github.com/xianxu/ariadne
            issue_id: "000385"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/201d08d1f0de78ef
            reviewed_head: c907a4d0ba566153263b2334826a9eeb5779f1dd
            slot: ariadne:3
            requested_at: "2026-10-15T20:04:59.847852Z"
          phase: finished
          pid: 9588
          started_at: "2026-10-15T20:04:59.956079Z"
          finished_at: "2026-10-15T20:07:23.068597Z"
          session_id: 01a1412b-029d-7fc0-aa89-d0362e8e17f2
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: c907a4d0ba566153263b2334826a9eeb5779f1dd
        tracker_head: 50b325faaa7337121199f235859f0aa8b54027f5
        branch_point: 8e637c6e2dd431b4c82c42c2dea680f93918c785
        since: "2026-10-15T09:33:04-07:00"
        until: "2026-10-15T13:04:44-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T09:33:04-07:00"
              end: "2026-10-15T13:04:44-07:00"
              began: "2026-10-15T09:33:04-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a138fd-1ac0-7a83-826e-60623c226386
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a584a19e995ca1e921b1f3791b1044b4fa2c02533e2cfc0229b55973c8474416
              started: "2026-10-14T05:58:30.184Z"
              spawned: false
              bytes: 44288249
              windows:
                - start: "2026-10-15T09:33:04-07:00"
                  end: "2026-10-15T13:04:44-07:00"
            - harness: codex
              id: 01a14074-981e-74f2-af8c-a6ddf477fc6d
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a9886dc5ea07325421724dbd964a263d52c38a01666ce05f66b4eb483be6e957
              started: "2026-10-15T16:45:47.228Z"
              spawned: false
              bytes: 758658
              windows:
                - start: "2026-10-15T16:45:47.228Z"
                  end: "2026-10-15T13:04:44-07:00"
            - harness: codex
              id: 01a14079-9246-7ea3-91fb-c457e3886da9
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 1f1689dceec0617335228ae3ffa7158942cc8f392d39eed6bd1e8e0323cd5c72
              started: "2026-10-15T16:51:14.597Z"
              spawned: false
              bytes: 453960
              windows:
                - start: "2026-10-15T16:51:14.597Z"
                  end: "2026-10-15T13:04:44-07:00"
            - harness: codex
              id: 01a1407a-af6f-7152-833b-a43cafac9231
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 467ba6fd9787954503097fe5d6f76e5f3eacf111156a527937edb4a48121c222
              started: "2026-10-15T16:52:26.427Z"
              spawned: false
              bytes: 462233
              windows:
                - start: "2026-10-15T16:52:26.427Z"
                  end: "2026-10-15T13:04:44-07:00"
            - harness: codex
              id: 01a1407b-1424-7c31-964f-c6839b6c5951
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 9bfc53e280ec295615d2f254082bc30c1f99379efc2f80040203fcef658d93a0
              started: "2026-10-15T16:52:54.44Z"
              spawned: false
              bytes: 420615
              windows:
                - start: "2026-10-15T16:52:54.44Z"
                  end: "2026-10-15T13:04:44-07:00"
            - harness: codex
              id: 01a1407b-ee59-71d1-98d5-f8a27ea6615a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 9efef5aa5f9e50eef2d508357bf71aee275445db8b2a49fb239a8fb0e8da70a8
              started: "2026-10-15T16:53:46.292Z"
              spawned: true
              bytes: 611271
              windows:
                - start: "2026-10-15T16:53:46.292Z"
                  end: "2026-10-15T13:04:44-07:00"
            - harness: codex
              id: 01a1407d-b788-7921-97ed-2aa273e67570
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 23ed2409e504187554c6d6aa4eba3868f59c74027970bb50e97cb1bd40c15e76
              started: "2026-10-15T16:55:43.336Z"
              spawned: true
              bytes: 806710
              windows:
                - start: "2026-10-15T16:55:43.336Z"
                  end: "2026-10-15T13:04:44-07:00"
            - harness: codex
              id: 01a1407e-0e37-7f82-9b51-1e2425d58944
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: 888e8c69e57908465103bf8fe4ab74d0205318c884c073cf5c4e53ca373a20f0
              started: "2026-10-15T16:56:05.521Z"
              spawned: true
              bytes: 1906848
              windows:
                - start: "2026-10-15T16:56:05.521Z"
                  end: "2026-10-15T13:04:44-07:00"
            - harness: codex
              id: 01a14080-72b2-7e43-bc09-63ffe1e09102
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: e1a877e295cf76c410c5c74c7a817e582583faf1e507b98fe977b821fe7d9247
              started: "2026-10-15T16:58:42.319Z"
              spawned: true
              bytes: 1553084
              windows:
                - start: "2026-10-15T16:58:42.319Z"
                  end: "2026-10-15T13:04:44-07:00"
            - harness: codex
              id: 01a140bb-446b-78a2-bad0-8a87ab45e0d6
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a138fd-1ac0-7a83-826e-60623c226386
              digest: bea2c41d92607fceb2ea2d50a0cf2458e402a022f020adbe379546da89cb68ea
              started: "2026-10-15T18:02:57.072Z"
              spawned: true
              bytes: 811720
              windows:
                - start: "2026-10-15T18:02:57.072Z"
                  end: "2026-10-15T13:04:44-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.03448333333333333
        cuts:
            - dispatched: "2026-10-15T11:42:18-07:00"
              finalized: "2026-10-15T11:45:12-07:00"
              finalizing: false
            - dispatched: "2026-10-15T12:23:40-07:00"
              finalized: "2026-10-15T12:26:00-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-15T13:07:23-07:00"
      dispatched: "2026-10-15T13:04:59-07:00"
      agent: codex
      dispose:
        - id: BR-3
          disposition: addressed
          note: Failed scans retain unresolved ownership unless complete evidence proves exclusion. Production tracked and legacy actual/close regressions pass. Restoring the former early parse return and error-selection rule in a disposable copy makes the discovery failure matrix and unreadable-candidate regression fail; the copy was removed.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: c907a4d0ba566153263b2334826a9eeb5779f1dd
      blocked: false
---

# Gate ledger — ariadne#385 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T11:45:12-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `unsupported-format-visibility` Qoder nonzero token counters lack the promised format diagnostic
  cmd/sdlc/internal/usage/qoder.go:117 ignores the parsed token fields. A temporary overlay test with input_tokens=123 and output_tokens=456 returned null tokens but only sources/records coverage and spend_complete=true. The plan at lines 137–140 explicitly requires a visible diagnostic for unrecognized counter semantics. Validate all four placeholder fields, retain null tokens, and add regression coverage for nonzero and invalid values (ARCH-SECURE, ARCH-PURPOSE).
- **BR-2** [Important] `user-surface-documentation` README omits Qoder usage and links to a removed atlas anchor
  README.md:130–145 still describes Claude/Codex tokens only, omitting Qoder credit reporting, calibration support, and schema-version-2 nullable measurements. Its local-token-usage-365 link no longer matches the renamed atlas heading. Update this passage and link in the same boundary, as required by the README docs gate (ARCH-PURPOSE).

## Round 2 — 2026-10-15T12:26:00-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — qoder.go checks all four placeholders, including streaming rows. Regression tests cover nonzero and invalid values; removing the check through a temporary overlay makes the placeholder and streaming tests fail.
- BR-2 — addressed — README.md:130–149 now documents Qoder credits, calibration and schema-version-2 nullable measurements, and links to the matching Local usage heading in atlas/workflow/sdlc-binary.md. The collector and calibration implementation support the documented behavior.

### Raised

- **BR-3** [Critical] `incomplete-evidence-adoption` Malformed prefixes allow close to adopt incomplete Qoder actuals
  cmd/sdlc/internal/transcripts/qoder.go:119–121 returns on a malformed row before later workspace metadata can establish relevance. discovery.go:54–59 then treats the failed source as skippable. A temporary extension of TestQoderActualDiscoveryRefusesPartialOwnedEvidence with "not-json\n" before valid owned rows returned the good source with no error. A production actual/close overlay confirmed actualMeasured, 0.1 hours and adopted=true despite that omitted owned source. Preserve uncertainty until exclusion is proven: continue bounded relevance discovery after recoverable parse failures, and refuse automatic adoption when scan failures or limits leave ownership unresolved. Sweep parse, read and bound failures before ownership resolution, with discovery and close regressions (ARCH-SECURE, ARCH-ORDER, ARCH-PURPOSE).

## Round 3 — 2026-10-15T13:07:23-07:00 (codex) — passed

### Disposed

- BR-3 — addressed — Failed scans retain unresolved ownership unless complete evidence proves exclusion. Production tracked and legacy actual/close regressions pass. Restoring the former early parse return and error-selection rule in a disposable copy makes the discovery failure matrix and unreadable-candidate regression fail; the copy was removed.

## Open findings

(none — every finding has been disposed)
