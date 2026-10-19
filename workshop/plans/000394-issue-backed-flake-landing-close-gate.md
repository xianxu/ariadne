---
gate: boundary-review
issue: 394
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 86e468cc-1abc-40f2-9b8e-6476db969e02
            repository: github.com/xianxu/ariadne
            issue_id: "000394"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/6f0fffb25c67a2a1
            reviewed_head: 1a2e64e114bcf3f97920592e4a4f9350be51c5ca
            slot: ariadne:2
            requested_at: "2026-10-18T23:42:59.604244Z"
          phase: finished
          pid: 15555
          started_at: "2026-10-18T23:42:59.697805Z"
          finished_at: "2026-10-18T23:45:20.68225Z"
          session_id: a95c51ac-f943-44e8-9dc9-1eb426ba6e61
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 1a2e64e114bcf3f97920592e4a4f9350be51c5ca
        tracker_head: 56f80c2788ea9f42471eb8dd56476107b4ec697b
        branch_point: 046e0ebb03d7c78bf46c8511fa6d6084d42137f4
        since: "2026-10-18T15:44:55-07:00"
        until: "2026-10-18T16:42:37-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T15:44:55-07:00"
              end: "2026-10-18T16:42:37-07:00"
              began: "2026-10-18T15:44:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0191d03b-74fe-4185-8091-8374ef70c0c9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6f878e4449d6a70c6d117dcc8e3e2dc7eb921bedd5cb164ae329d9abc8dbd463
              started: "2026-10-18T22:59:31.062Z"
              spawned: false
              bytes: 393720
              windows:
                - start: "2026-10-18T22:59:31.062Z"
                  end: "2026-10-18T16:42:37-07:00"
            - harness: claude
              id: 1baa8f6b-092f-413a-83f9-0b25ed003914
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 15568299dbfd540e29417f414005c5fb2212d08913ca02800e9b5f9ab1030c6c
              started: "2026-10-18T22:57:31.956Z"
              spawned: false
              bytes: 519489
              windows:
                - start: "2026-10-18T22:57:31.956Z"
                  end: "2026-10-18T16:42:37-07:00"
            - harness: claude
              id: 23a2a6e8-f430-49c9-938f-6a09729f7067
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6c2a2ebaf3bf9384772532e724d4644737187d5f9aab48909773ef49f7eecb93
              started: "2026-10-18T22:53:00.413Z"
              spawned: false
              bytes: 571220
              windows:
                - start: "2026-10-18T22:53:00.413Z"
                  end: "2026-10-18T16:42:37-07:00"
            - harness: claude
              id: aa450fd7-6a6d-4cd2-abcc-3339322077dc
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 776fdfac606030c3cf04f41a719a0ad32e2a31cec8e0f54a07ed46e3daf64dd0
              started: "2026-10-18T22:59:17.58Z"
              spawned: false
              bytes: 502106
              windows:
                - start: "2026-10-18T22:59:17.58Z"
                  end: "2026-10-18T16:42:37-07:00"
            - harness: claude
              id: d11b9adb-1d2c-4872-81c3-b6aa47a04715
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: fb35a51e74ad315fa11eb15dc5dd4cd0f8c6a8a811c7d9cafac6943d15e78dfc
              started: "2026-10-18T22:56:29.354Z"
              spawned: false
              bytes: 556087
              windows:
                - start: "2026-10-18T22:56:29.354Z"
                  end: "2026-10-18T16:42:37-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 230aef7e545dd13303b05ed9aa029fbe963de4eb022e19934de3444b141187b1
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 227684245
              windows:
                - start: "2026-10-18T15:44:55-07:00"
                  end: "2026-10-18T16:42:37-07:00"
      "n": 1
      timestamp: "2026-10-18T16:45:21-07:00"
      dispatched: "2026-10-18T16:42:59-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Test-name validation accepts '-' and '.'/'..' path segments, so RerunCommand can emit go test flags (ARCH-SECURE)
          detail: registry.go importPath regex lets "-exec.TestA" and "./x.TestB" pass ValidTest. A probe test showed RerunCommand returns [go test -count=1 -run ^(TestA)$ -exec]. The card check uses the same rule and any slot can write the card. Reject segments starting with '-' or '.', and add the hostile cases to TestRerunCommand and TestBuildRegistry.
          family: argv-identifier-validation
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: flake register computes the merged flake_tests from an earlier read than its compare-and-swap snapshot, so concurrent registrations are lost (ARCH-ORDER)
          detail: flake_cmd.go:285-294 merges from the Fresh registry read, and the setTests decide callback (:139-143) ignores card.Raw. Merge from the card passed to decide, check MaxRerun there, and test a card that changes between the two reads.
          family: cas-mutator-uses-stale-read
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: flake list omits owner, age and the log that D3 and Done-when 3 require (ARCH-PURPOSE)
          detail: flake_cmd.go:341-343 prints test, issue and status only. Attempt lines carry no LogPath. Add the claimant and age to flakes.Card, print the log path, and assert all three in TestFlakeList.
          family: deferred-purpose-subset
          round: 1
        - kind: defect
          id: BR-4
          severity: Important
          title: README update missing for sdlc flake register / sdlc flake list
          detail: The atlas row was added, but README.md, which documents local-check and test-stats, has nothing for the new user-facing command.
          family: docs-gate-readme
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: --evidence with --issue N is only printed, not recorded as the plan says
          detail: Only the path that files a new issue keeps the evidence text, in the issue body.
          family: evidence-recorded-verbatim
          round: 1
        - kind: defect
          id: BR-6
          severity: Minor
          title: flake register --issue N does not refuse a terminal (closed) issue
          detail: It reports success, but the test never enters the registry.
          family: register-target-state-check
          round: 1
        - kind: defect
          id: BR-7
          severity: Minor
          title: Unexplained treats failed-twice, too-many and rule-off attempts as explained once their failures are registered
          detail: A blind full rerun after "known flake failed twice" would then pass M3's guard. Decide this explicitly in M3.
          family: unexplained-outcome-policy
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e48c0ccd-fa0e-42bb-bf56-99e70549daea
            repository: github.com/xianxu/ariadne
            issue_id: "000394"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2ccbf0b21a1c09d2
            reviewed_head: 93687c932b8b1220194fa048d949abdab7bc99d7
            slot: ariadne:2
            requested_at: "2026-10-18T23:49:32.691167Z"
          phase: finished
          pid: 76127
          started_at: "2026-10-18T23:49:32.922656Z"
          finished_at: "2026-10-18T23:52:49.010462Z"
          session_id: fa33b215-2823-4fd1-b6c8-9d7016d91b53
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 93687c932b8b1220194fa048d949abdab7bc99d7
        tracker_head: 56f80c2788ea9f42471eb8dd56476107b4ec697b
        branch_point: 046e0ebb03d7c78bf46c8511fa6d6084d42137f4
        since: "2026-10-18T15:44:55-07:00"
        until: "2026-10-18T16:48:36-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T15:44:55-07:00"
              end: "2026-10-18T16:48:36-07:00"
              began: "2026-10-18T15:44:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0191d03b-74fe-4185-8091-8374ef70c0c9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6f878e4449d6a70c6d117dcc8e3e2dc7eb921bedd5cb164ae329d9abc8dbd463
              started: "2026-10-18T22:59:31.062Z"
              spawned: false
              bytes: 393720
              windows:
                - start: "2026-10-18T22:59:31.062Z"
                  end: "2026-10-18T16:48:36-07:00"
            - harness: claude
              id: 1baa8f6b-092f-413a-83f9-0b25ed003914
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 15568299dbfd540e29417f414005c5fb2212d08913ca02800e9b5f9ab1030c6c
              started: "2026-10-18T22:57:31.956Z"
              spawned: false
              bytes: 519489
              windows:
                - start: "2026-10-18T22:57:31.956Z"
                  end: "2026-10-18T16:48:36-07:00"
            - harness: claude
              id: 23a2a6e8-f430-49c9-938f-6a09729f7067
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6c2a2ebaf3bf9384772532e724d4644737187d5f9aab48909773ef49f7eecb93
              started: "2026-10-18T22:53:00.413Z"
              spawned: false
              bytes: 571220
              windows:
                - start: "2026-10-18T22:53:00.413Z"
                  end: "2026-10-18T16:48:36-07:00"
            - harness: claude
              id: aa450fd7-6a6d-4cd2-abcc-3339322077dc
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 776fdfac606030c3cf04f41a719a0ad32e2a31cec8e0f54a07ed46e3daf64dd0
              started: "2026-10-18T22:59:17.58Z"
              spawned: false
              bytes: 502106
              windows:
                - start: "2026-10-18T22:59:17.58Z"
                  end: "2026-10-18T16:48:36-07:00"
            - harness: claude
              id: d11b9adb-1d2c-4872-81c3-b6aa47a04715
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: fb35a51e74ad315fa11eb15dc5dd4cd0f8c6a8a811c7d9cafac6943d15e78dfc
              started: "2026-10-18T22:56:29.354Z"
              spawned: false
              bytes: 556087
              windows:
                - start: "2026-10-18T22:56:29.354Z"
                  end: "2026-10-18T16:48:36-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 750c3773037d57756bc7df6e084cfee7e7dae61411e9be96eaa9775f06fe5ae1
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 230026645
              windows:
                - start: "2026-10-18T15:44:55-07:00"
                  end: "2026-10-18T16:48:36-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T16:42:59-07:00"
              finalized: "2026-10-18T16:45:21-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T16:52:51-07:00"
      dispatched: "2026-10-18T16:49:32-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: registry.go:20 segment regex rejects leading - and . ; hostile cases in TestBuildRegistry and TestRerunCommand; card check shares ValidTest.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: 'Code is right (addFlakeTests merges onto the CAS card) but nothing tests it: changing it to tests := add keeps every flake test green, because the race test''s fake re-implements the merge and the real-card test never checks that the old test survives. Add that assertion on a real card.'
          round: 2
        - id: BR-3
          disposition: addressed
          note: flake list prints owner, age and log path (flake_cmd.go:369,396,406); TestFlakeList asserts all three.
          round: 2
        - id: BR-4
          disposition: addressed
          note: README.md:498-505 documents flake register and flake list, matching the help and code.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Plan Revisions replace 'recorded verbatim' with filed-issue body or printed for N's Log; help (flake_cmd.go:210-213) and code (:321-323) match.
          round: 2
        - id: BR-6
          disposition: addressed
          note: addFlakeTests refuses a terminal status inside the CAS callback; TestFlakeRegistryFromTrackerCards goes red when the guard is reverted.
          round: 2
        - id: BR-7
          disposition: addressed
          note: neverExplained in evidence.go writes the policy down; TestEvidence covers failed-twice, too-many, rule-off and unattributable.
          round: 2
      findings:
        - kind: defect
          id: BR-8
          severity: Minor
          title: defaultFlakeDeps.registry duplicates loadFlakeRegistry, and attempts/logPath duplicate the store-dir setup (ARCH-DRY)
          detail: flake_cmd.go:159-170 repeats loadFlakeRegistry (:104-114) except for returning the cards; :130-157 builds localCheckLogDir twice. Have loadFlakeRegistry return the cards and pull out a suiteStoreDir helper.
          family: duplicated-io-glue
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2a2cafb7-c774-4abd-9cb6-7c16c4ebd1ac
            repository: github.com/xianxu/ariadne
            issue_id: "000394"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/ea18e1af9ed61283
            reviewed_head: 9baa43c83bf3abb2623fd48cdd7acec651730e05
            slot: ariadne:2
            requested_at: "2026-10-18T23:54:42.537406Z"
          phase: finished
          pid: 27945
          started_at: "2026-10-18T23:54:42.646538Z"
          finished_at: "2026-10-18T23:56:18.771914Z"
          session_id: dbb666a8-4eb0-499d-9bdb-08e228ed9f4e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 9baa43c83bf3abb2623fd48cdd7acec651730e05
        tracker_head: 56f80c2788ea9f42471eb8dd56476107b4ec697b
        branch_point: 046e0ebb03d7c78bf46c8511fa6d6084d42137f4
        since: "2026-10-18T15:44:55-07:00"
        until: "2026-10-18T16:54:26-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T15:44:55-07:00"
              end: "2026-10-18T16:54:26-07:00"
              began: "2026-10-18T15:44:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0191d03b-74fe-4185-8091-8374ef70c0c9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6f878e4449d6a70c6d117dcc8e3e2dc7eb921bedd5cb164ae329d9abc8dbd463
              started: "2026-10-18T22:59:31.062Z"
              spawned: false
              bytes: 393720
              windows:
                - start: "2026-10-18T22:59:31.062Z"
                  end: "2026-10-18T16:54:26-07:00"
            - harness: claude
              id: 1baa8f6b-092f-413a-83f9-0b25ed003914
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 15568299dbfd540e29417f414005c5fb2212d08913ca02800e9b5f9ab1030c6c
              started: "2026-10-18T22:57:31.956Z"
              spawned: false
              bytes: 519489
              windows:
                - start: "2026-10-18T22:57:31.956Z"
                  end: "2026-10-18T16:54:26-07:00"
            - harness: claude
              id: 23a2a6e8-f430-49c9-938f-6a09729f7067
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6c2a2ebaf3bf9384772532e724d4644737187d5f9aab48909773ef49f7eecb93
              started: "2026-10-18T22:53:00.413Z"
              spawned: false
              bytes: 571220
              windows:
                - start: "2026-10-18T22:53:00.413Z"
                  end: "2026-10-18T16:54:26-07:00"
            - harness: claude
              id: aa450fd7-6a6d-4cd2-abcc-3339322077dc
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 776fdfac606030c3cf04f41a719a0ad32e2a31cec8e0f54a07ed46e3daf64dd0
              started: "2026-10-18T22:59:17.58Z"
              spawned: false
              bytes: 502106
              windows:
                - start: "2026-10-18T22:59:17.58Z"
                  end: "2026-10-18T16:54:26-07:00"
            - harness: claude
              id: d11b9adb-1d2c-4872-81c3-b6aa47a04715
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: fb35a51e74ad315fa11eb15dc5dd4cd0f8c6a8a811c7d9cafac6943d15e78dfc
              started: "2026-10-18T22:56:29.354Z"
              spawned: false
              bytes: 556087
              windows:
                - start: "2026-10-18T22:56:29.354Z"
                  end: "2026-10-18T16:54:26-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: d682fc4e8261d27d3fce922c8332e621296fefa18aad017e2f2b0c01501911ee
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 231550677
              windows:
                - start: "2026-10-18T15:44:55-07:00"
                  end: "2026-10-18T16:54:26-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T16:42:59-07:00"
              finalized: "2026-10-18T16:45:21-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T16:49:32-07:00"
              finalized: "2026-10-18T16:52:51-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-18T16:56:28-07:00"
      dispatched: "2026-10-18T16:54:42-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-2
          disposition: addressed
          note: addFlakeTests merges, caps and checks closed status on the card the CAS hands to decide (cardsetters.go:68); making it ignore the card's tests turns TestAddFlakeTestsMergesOntoTheCard and TestFlakeRegistryFromTrackerCards red.
          round: 3
        - id: BR-8
          disposition: addressed
          note: registry dep now calls loadFlakeRegistry (which returns cards); suiteStoreDir (flake_cmd.go:128) is the single store-dir helper used by attempts and logPath.
          round: 3
      findings:
        - kind: defect
          id: BR-9
          severity: Minor
          title: The one-owner-per-test check reads the registry, not the card under compare-and-swap, so concurrent registrations on two issues can both pass
          detail: 'This is the 2nd finding in family cas-mutator-uses-stale-read. The rule: an invariant a mutator enforces must either be checked on the snapshot the CAS protects, or be enforced again where the data is read. The cross-card uniqueness here can only be the second kind. BuildRegistry (registry.go:71) already refuses a test named by two issues, which fails closed and visibly. The fix is to state that rule in the register help and atlas row, not to add another check at the write site. Measured prevalence: 1 cross-card site (flake_cmd.go:294-304).'
          family: cas-mutator-uses-stale-read
          round: 3
        - kind: defect
          id: BR-10
          severity: Minor
          title: The suite-attempt writer now records version 2, which older sdlc binaries reject as an unsupported version
          detail: 'teststats_store.go:50 and localcheck_run.go:190: an older binary reading the same store skips or flags v2 records. Note this as a breaking change for downstream repos when the release ships.'
          family: breaking-change-undocumented
          round: 3
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: 9baa43c83bf3abb2623fd48cdd7acec651730e05
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 9513e4ab-6515-4927-931e-466daee25d25
            repository: github.com/xianxu/ariadne
            issue_id: "000394"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2a8edeac52931d01
            reviewed_head: a837ec2812c669e71c4e2c66e0e5c53a64a617dc
            slot: ariadne:2
            requested_at: "2026-10-19T00:38:24.489004Z"
          phase: finished
          pid: 64171
          started_at: "2026-10-19T00:38:24.604582Z"
          finished_at: "2026-10-19T00:42:32.941472Z"
          session_id: 2edc7a8a-36c0-4e66-bd0f-ae35e1cbce20
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: a837ec2812c669e71c4e2c66e0e5c53a64a617dc
        tracker_head: 202d6ff1fcab1d503009681d88c2f222037e6a0c
        branch_point: 046e0ebb03d7c78bf46c8511fa6d6084d42137f4
        since: "2026-10-18T15:44:55-07:00"
        until: "2026-10-18T17:37:56-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T15:44:55-07:00"
              end: "2026-10-18T17:37:56-07:00"
              began: "2026-10-18T15:44:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0191d03b-74fe-4185-8091-8374ef70c0c9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6f878e4449d6a70c6d117dcc8e3e2dc7eb921bedd5cb164ae329d9abc8dbd463
              started: "2026-10-18T22:59:31.062Z"
              spawned: false
              bytes: 393720
              windows:
                - start: "2026-10-18T22:59:31.062Z"
                  end: "2026-10-18T17:37:56-07:00"
            - harness: claude
              id: 1baa8f6b-092f-413a-83f9-0b25ed003914
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 15568299dbfd540e29417f414005c5fb2212d08913ca02800e9b5f9ab1030c6c
              started: "2026-10-18T22:57:31.956Z"
              spawned: false
              bytes: 519489
              windows:
                - start: "2026-10-18T22:57:31.956Z"
                  end: "2026-10-18T17:37:56-07:00"
            - harness: claude
              id: 23a2a6e8-f430-49c9-938f-6a09729f7067
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6c2a2ebaf3bf9384772532e724d4644737187d5f9aab48909773ef49f7eecb93
              started: "2026-10-18T22:53:00.413Z"
              spawned: false
              bytes: 571220
              windows:
                - start: "2026-10-18T22:53:00.413Z"
                  end: "2026-10-18T17:37:56-07:00"
            - harness: claude
              id: aa450fd7-6a6d-4cd2-abcc-3339322077dc
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 776fdfac606030c3cf04f41a719a0ad32e2a31cec8e0f54a07ed46e3daf64dd0
              started: "2026-10-18T22:59:17.58Z"
              spawned: false
              bytes: 502106
              windows:
                - start: "2026-10-18T22:59:17.58Z"
                  end: "2026-10-18T17:37:56-07:00"
            - harness: claude
              id: d11b9adb-1d2c-4872-81c3-b6aa47a04715
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: fb35a51e74ad315fa11eb15dc5dd4cd0f8c6a8a811c7d9cafac6943d15e78dfc
              started: "2026-10-18T22:56:29.354Z"
              spawned: false
              bytes: 556087
              windows:
                - start: "2026-10-18T22:56:29.354Z"
                  end: "2026-10-18T17:37:56-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c00a4e9f5870754eca1f150a3d025524bfd14db5c8b77d1a60a2c065318a6dcf
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 242378725
              windows:
                - start: "2026-10-18T15:44:55-07:00"
                  end: "2026-10-18T17:37:56-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T16:42:59-07:00"
              finalized: "2026-10-18T16:45:21-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T16:49:32-07:00"
              finalized: "2026-10-18T16:52:51-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T16:54:42-07:00"
              finalized: "2026-10-18T16:56:28-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-18T17:42:33-07:00"
      dispatched: "2026-10-18T17:38:24-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-11
          severity: Critical
          title: A suite killed by timeout_minutes lands green when its partial log names only a known flake
          detail: 'knownFlakeStep builds flakes.Result from log facts only; rec.TimedOut never reaches Classify, and unattributedFailures only matches go''s in-process timeout panic. Reproduced in the checkout: a suite that prints a FAIL for a registered test and then sleeps past its timeout gives timedout=true, outcome known-flake-rerun, status success. That contradicts the help text and the D1 union argument. Rule: any failure signal the log does not name (a timeout, a run that did not start, an unparseable log) goes into Unattributed before Classify. Add a gate test for the timed-out case.'
          family: failure-signal-dropped-before-classify
          round: 4
        - kind: defect
          id: BR-12
          severity: Important
          title: The known-flake pass is posted without rereading the registry or policy, unlike the plan's ARCH-ORDER claim
          detail: The registry is read once at localcheck_flakes.go:52 and reused after a rerun of up to 15 minutes. republishSavedPass and savedAttemptRecord post a saved known-flake-rerun or known-flake-accepted outcome with no registry or policy check, so a closed issue or a policy now set to off still turns green. Recheck the outcome against a fresh registry and policy just before posting success, or revise the plan.
          family: effect-revalidates-fresh-state
          round: 4
        - kind: defect
          id: BR-13
          severity: Important
          title: Rerun attempts are counted as full suite runs by test-stats readers
          detail: Rerun attempts are saved with check suite and only rerun_of to mark them. teststats.go:118-131 counts them in Runs and FlakeDenominator, adds their wall time to the suite wall distribution and their units to slow-unit times, and nothing reads RerunOf apart from validation. List every reader of the store (summary, snapshot, flake list) and decide how each treats rerun attempts.
          family: new-record-kind-unswept-readers
          round: 4
        - kind: defect
          id: BR-14
          severity: Important
          title: README does not cover known-flake landing or the known_flakes key added to the local-check and merge help
          detail: 'This is the 2nd finding in family docs-gate-readme. Rule: any range that changes a helptext''s user-facing behaviour or config key updates the matching README section, and the atlas ci-merge-check recorded-local-checks section, in the same commit. Here README.md lines 486-505 still describe the registry only.'
          family: docs-gate-readme
          round: 4
        - kind: defect
          id: BR-15
          severity: Important
          title: The new declaration refusals and the persistent-main-failure label have no tests
          detail: There is no test for an invalid known_flakes value, for known_flakes on a check other than the suite in parseLocalChecks, or for the diagnoseSuiteFailure PersistentMain label in suitebaseline.go.
          family: claimed-behavior-without-test
          round: 4
        - kind: defect
          id: BR-16
          severity: Minor
          title: The rerun timeout is a fixed 15 minutes, not bounded by the check's timeout_minutes as the plan states
          detail: knownFlakeRerunTimeout in localcheck_flakes.go is a constant; a rerun that times out is also labelled known-flake-failed-twice.
          family: declared-envelope-not-enforced
          round: 4
        - kind: defect
          id: BR-17
          severity: Minor
          title: A rerun passes on exit 0 alone, so a run with no tests to run would count as landing evidence
          detail: Nothing confirms that the named tests appear as passed in the rerun log.
          family: rerun-evidence-unverified
          round: 4
        - kind: defect
          id: BR-18
          severity: Minor
          title: The persistent-main-failure label repeats the pre-existing list it sits beside
          detail: This is the 2nd finding in family duplicated-io-glue. Here it is one shared-failure list printed twice, in suitebaseline.go around line 52. Compute the list once and label it once.
          family: duplicated-io-glue
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a7e29a06-5a6f-4f1c-9db7-24ddfd9e11c6
            repository: github.com/xianxu/ariadne
            issue_id: "000394"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/db2a032ce7a226a8
            reviewed_head: 7054740a11fd5047c99319f6a4905cc84bb801f2
            slot: ariadne:2
            requested_at: "2026-10-19T01:10:30.323486Z"
          phase: finished
          pid: 75766
          started_at: "2026-10-19T01:10:30.477473Z"
          finished_at: "2026-10-19T01:24:30.648801Z"
          session_id: 320c2191-2b48-4742-816a-423fc177b85d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7054740a11fd5047c99319f6a4905cc84bb801f2
        tracker_head: 74a115c29467038eb7946caa7c125c20a7dd383c
        branch_point: 046e0ebb03d7c78bf46c8511fa6d6084d42137f4
        since: "2026-10-18T15:44:55-07:00"
        until: "2026-10-18T18:10:13-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T15:44:55-07:00"
              end: "2026-10-18T18:10:13-07:00"
              began: "2026-10-18T15:44:55-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0191d03b-74fe-4185-8091-8374ef70c0c9
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6f878e4449d6a70c6d117dcc8e3e2dc7eb921bedd5cb164ae329d9abc8dbd463
              started: "2026-10-18T22:59:31.062Z"
              spawned: false
              bytes: 393720
              windows:
                - start: "2026-10-18T22:59:31.062Z"
                  end: "2026-10-18T18:10:13-07:00"
            - harness: claude
              id: 1baa8f6b-092f-413a-83f9-0b25ed003914
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 15568299dbfd540e29417f414005c5fb2212d08913ca02800e9b5f9ab1030c6c
              started: "2026-10-18T22:57:31.956Z"
              spawned: false
              bytes: 519489
              windows:
                - start: "2026-10-18T22:57:31.956Z"
                  end: "2026-10-18T18:10:13-07:00"
            - harness: claude
              id: 23a2a6e8-f430-49c9-938f-6a09729f7067
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6c2a2ebaf3bf9384772532e724d4644737187d5f9aab48909773ef49f7eecb93
              started: "2026-10-18T22:53:00.413Z"
              spawned: false
              bytes: 571220
              windows:
                - start: "2026-10-18T22:53:00.413Z"
                  end: "2026-10-18T18:10:13-07:00"
            - harness: claude
              id: aa450fd7-6a6d-4cd2-abcc-3339322077dc
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 776fdfac606030c3cf04f41a719a0ad32e2a31cec8e0f54a07ed46e3daf64dd0
              started: "2026-10-18T22:59:17.58Z"
              spawned: false
              bytes: 502106
              windows:
                - start: "2026-10-18T22:59:17.58Z"
                  end: "2026-10-18T18:10:13-07:00"
            - harness: claude
              id: d11b9adb-1d2c-4872-81c3-b6aa47a04715
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: fb35a51e74ad315fa11eb15dc5dd4cd0f8c6a8a811c7d9cafac6943d15e78dfc
              started: "2026-10-18T22:56:29.354Z"
              spawned: false
              bytes: 556087
              windows:
                - start: "2026-10-18T22:56:29.354Z"
                  end: "2026-10-18T18:10:13-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: b0f647a3e3619420ea05292638efd143ef11448dabcecde5f63bb6d0b89bcd44
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 249820253
              windows:
                - start: "2026-10-18T15:44:55-07:00"
                  end: "2026-10-18T18:10:13-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T16:42:59-07:00"
              finalized: "2026-10-18T16:45:21-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T16:49:32-07:00"
              finalized: "2026-10-18T16:52:51-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T16:54:42-07:00"
              finalized: "2026-10-18T16:56:28-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-18T17:38:24-07:00"
              finalized: "2026-10-18T17:42:33-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-18T18:24:31-07:00"
      dispatched: "2026-10-18T18:10:30-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-11
          disposition: addressed
          note: Timeout and no-exit are unattributed before Classify (localcheck_run.go:151); reverting it turns "a suite killed by its timeout never lands (BR-11)" red.
          round: 5
        - id: BR-12
          disposition: addressed
          note: flakes.Revalidate runs at post, at reuse and at republish; removing the run-path call turns the issue-closes-during-rerun test red. Policy recheck is per run (documented).
          round: 5
        - id: BR-13
          disposition: addressed
          note: The stats aggregate and the weekly snapshot now skip rerun rows; TestSuiteStatsRerunIsEvidenceNotARun goes red on revert. The missed reader (loadSavedAttempt) is raised as a new Critical.
          round: 5
        - id: BR-14
          disposition: addressed
          note: README.md:506-512 and atlas/workflow/ci-merge-check.md:166-187 now describe the gate and the known_flakes values, matching the help text.
          round: 5
        - id: BR-15
          disposition: addressed
          note: TestParseLocalChecks covers the bad value, the off-suite key and the default; TestSuiteBaselineRetainsRealLockOwner asserts the persistent-main label.
          round: 5
        - id: BR-16
          disposition: not-addressed
          note: min(15m, timeout_minutes) is in the code, but no test pins the bound.
          round: 5
        - id: BR-17
          disposition: addressed
          note: RerunMissing plus go test -v; reverting the gate turns "a rerun that shows no PASS proves nothing" red. A bypass via saved-pass reuse is raised as new.
          round: 5
        - id: BR-18
          disposition: addressed
          note: The label is computed once inside the pre-existing branch; the reentry test asserts TestB appears exactly once.
          round: 5
      findings:
        - kind: defect
          id: BR-19
          severity: Critical
          title: loadSavedAttempt reuses a rerun row as the head's suite pass, so merge posts success for an inconclusive rerun
          detail: 'loadSavedAttempt filters only on SHA and check. A rerun row has Check suite and a later FinishedAt, and passed() is true on exit 0, while revalidateKnownFlake passes non-landing kinds through. Reproduced: suiteFailing plus a rerun printing [no tests to run]; run posts failure, then republishSavedPass returns ok=true and posts success, publishing the rerun''s log as the suite''s. 2nd in family. Rule: a RerunOf row is never a suite attempt for any reader; enforce it with one shared predicate or a distinct rerun representation, and decide reuse from the parent attempt''s outcome. Add a regression test.'
          family: new-record-kind-unswept-readers
          round: 5
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#394 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T16:45:21-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `argv-identifier-validation` (kind: defect) Test-name validation accepts '-' and '.'/'..' path segments, so RerunCommand can emit go test flags (ARCH-SECURE)
  registry.go importPath regex lets "-exec.TestA" and "./x.TestB" pass ValidTest. A probe test showed RerunCommand returns [go test -count=1 -run ^(TestA)$ -exec]. The card check uses the same rule and any slot can write the card. Reject segments starting with '-' or '.', and add the hostile cases to TestRerunCommand and TestBuildRegistry.
- **BR-2** [Important] `cas-mutator-uses-stale-read` (kind: defect) flake register computes the merged flake_tests from an earlier read than its compare-and-swap snapshot, so concurrent registrations are lost (ARCH-ORDER)
  flake_cmd.go:285-294 merges from the Fresh registry read, and the setTests decide callback (:139-143) ignores card.Raw. Merge from the card passed to decide, check MaxRerun there, and test a card that changes between the two reads.
- **BR-3** [Important] `deferred-purpose-subset` (kind: defect) flake list omits owner, age and the log that D3 and Done-when 3 require (ARCH-PURPOSE)
  flake_cmd.go:341-343 prints test, issue and status only. Attempt lines carry no LogPath. Add the claimant and age to flakes.Card, print the log path, and assert all three in TestFlakeList.
- **BR-4** [Important] `docs-gate-readme` (kind: defect) README update missing for sdlc flake register / sdlc flake list
  The atlas row was added, but README.md, which documents local-check and test-stats, has nothing for the new user-facing command.
- **BR-5** [Minor] `evidence-recorded-verbatim` (kind: defect) --evidence with --issue N is only printed, not recorded as the plan says
  Only the path that files a new issue keeps the evidence text, in the issue body.
- **BR-6** [Minor] `register-target-state-check` (kind: defect) flake register --issue N does not refuse a terminal (closed) issue
  It reports success, but the test never enters the registry.
- **BR-7** [Minor] `unexplained-outcome-policy` (kind: defect) Unexplained treats failed-twice, too-many and rule-off attempts as explained once their failures are registered
  A blind full rerun after "known flake failed twice" would then pass M3's guard. Decide this explicitly in M3.

## Round 2 — 2026-10-18T16:52:51-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — registry.go:20 segment regex rejects leading - and . ; hostile cases in TestBuildRegistry and TestRerunCommand; card check shares ValidTest.
- BR-2 — not-addressed — Code is right (addFlakeTests merges onto the CAS card) but nothing tests it: changing it to tests := add keeps every flake test green, because the race test's fake re-implements the merge and the real-card test never checks that the old test survives. Add that assertion on a real card.
- BR-3 — addressed — flake list prints owner, age and log path (flake_cmd.go:369,396,406); TestFlakeList asserts all three.
- BR-4 — addressed — README.md:498-505 documents flake register and flake list, matching the help and code.
- BR-5 — addressed — Plan Revisions replace 'recorded verbatim' with filed-issue body or printed for N's Log; help (flake_cmd.go:210-213) and code (:321-323) match.
- BR-6 — addressed — addFlakeTests refuses a terminal status inside the CAS callback; TestFlakeRegistryFromTrackerCards goes red when the guard is reverted.
- BR-7 — addressed — neverExplained in evidence.go writes the policy down; TestEvidence covers failed-twice, too-many, rule-off and unattributable.

### Raised

- **BR-8** [Minor] `duplicated-io-glue` (kind: defect) defaultFlakeDeps.registry duplicates loadFlakeRegistry, and attempts/logPath duplicate the store-dir setup (ARCH-DRY)
  flake_cmd.go:159-170 repeats loadFlakeRegistry (:104-114) except for returning the cards; :130-157 builds localCheckLogDir twice. Have loadFlakeRegistry return the cards and pull out a suiteStoreDir helper.

## Round 3 — 2026-10-18T16:56:28-07:00 (claude) — passed

### Disposed

- BR-2 — addressed — addFlakeTests merges, caps and checks closed status on the card the CAS hands to decide (cardsetters.go:68); making it ignore the card's tests turns TestAddFlakeTestsMergesOntoTheCard and TestFlakeRegistryFromTrackerCards red.
- BR-8 — addressed — registry dep now calls loadFlakeRegistry (which returns cards); suiteStoreDir (flake_cmd.go:128) is the single store-dir helper used by attempts and logPath.

### Raised

- **BR-9** [Minor] `cas-mutator-uses-stale-read` (kind: defect) The one-owner-per-test check reads the registry, not the card under compare-and-swap, so concurrent registrations on two issues can both pass
  This is the 2nd finding in family cas-mutator-uses-stale-read. The rule: an invariant a mutator enforces must either be checked on the snapshot the CAS protects, or be enforced again where the data is read. The cross-card uniqueness here can only be the second kind. BuildRegistry (registry.go:71) already refuses a test named by two issues, which fails closed and visibly. The fix is to state that rule in the register help and atlas row, not to add another check at the write site. Measured prevalence: 1 cross-card site (flake_cmd.go:294-304).
- **BR-10** [Minor] `breaking-change-undocumented` (kind: defect) The suite-attempt writer now records version 2, which older sdlc binaries reject as an unsupported version
  teststats_store.go:50 and localcheck_run.go:190: an older binary reading the same store skips or flags v2 records. Note this as a breaking change for downstream repos when the release ships.

## Round 4 — 2026-10-18T17:42:33-07:00 (claude) — BLOCKED

### Raised

- **BR-11** [Critical] `failure-signal-dropped-before-classify` (kind: defect) A suite killed by timeout_minutes lands green when its partial log names only a known flake
  knownFlakeStep builds flakes.Result from log facts only; rec.TimedOut never reaches Classify, and unattributedFailures only matches go's in-process timeout panic. Reproduced in the checkout: a suite that prints a FAIL for a registered test and then sleeps past its timeout gives timedout=true, outcome known-flake-rerun, status success. That contradicts the help text and the D1 union argument. Rule: any failure signal the log does not name (a timeout, a run that did not start, an unparseable log) goes into Unattributed before Classify. Add a gate test for the timed-out case.
- **BR-12** [Important] `effect-revalidates-fresh-state` (kind: defect) The known-flake pass is posted without rereading the registry or policy, unlike the plan's ARCH-ORDER claim
  The registry is read once at localcheck_flakes.go:52 and reused after a rerun of up to 15 minutes. republishSavedPass and savedAttemptRecord post a saved known-flake-rerun or known-flake-accepted outcome with no registry or policy check, so a closed issue or a policy now set to off still turns green. Recheck the outcome against a fresh registry and policy just before posting success, or revise the plan.
- **BR-13** [Important] `new-record-kind-unswept-readers` (kind: defect) Rerun attempts are counted as full suite runs by test-stats readers
  Rerun attempts are saved with check suite and only rerun_of to mark them. teststats.go:118-131 counts them in Runs and FlakeDenominator, adds their wall time to the suite wall distribution and their units to slow-unit times, and nothing reads RerunOf apart from validation. List every reader of the store (summary, snapshot, flake list) and decide how each treats rerun attempts.
- **BR-14** [Important] `docs-gate-readme` (kind: defect) README does not cover known-flake landing or the known_flakes key added to the local-check and merge help
  This is the 2nd finding in family docs-gate-readme. Rule: any range that changes a helptext's user-facing behaviour or config key updates the matching README section, and the atlas ci-merge-check recorded-local-checks section, in the same commit. Here README.md lines 486-505 still describe the registry only.
- **BR-15** [Important] `claimed-behavior-without-test` (kind: defect) The new declaration refusals and the persistent-main-failure label have no tests
  There is no test for an invalid known_flakes value, for known_flakes on a check other than the suite in parseLocalChecks, or for the diagnoseSuiteFailure PersistentMain label in suitebaseline.go.
- **BR-16** [Minor] `declared-envelope-not-enforced` (kind: defect) The rerun timeout is a fixed 15 minutes, not bounded by the check's timeout_minutes as the plan states
  knownFlakeRerunTimeout in localcheck_flakes.go is a constant; a rerun that times out is also labelled known-flake-failed-twice.
- **BR-17** [Minor] `rerun-evidence-unverified` (kind: defect) A rerun passes on exit 0 alone, so a run with no tests to run would count as landing evidence
  Nothing confirms that the named tests appear as passed in the rerun log.
- **BR-18** [Minor] `duplicated-io-glue` (kind: defect) The persistent-main-failure label repeats the pre-existing list it sits beside
  This is the 2nd finding in family duplicated-io-glue. Here it is one shared-failure list printed twice, in suitebaseline.go around line 52. Compute the list once and label it once.

## Round 5 — 2026-10-18T18:24:31-07:00 (claude) — BLOCKED

### Disposed

- BR-11 — addressed — Timeout and no-exit are unattributed before Classify (localcheck_run.go:151); reverting it turns "a suite killed by its timeout never lands (BR-11)" red.
- BR-12 — addressed — flakes.Revalidate runs at post, at reuse and at republish; removing the run-path call turns the issue-closes-during-rerun test red. Policy recheck is per run (documented).
- BR-13 — addressed — The stats aggregate and the weekly snapshot now skip rerun rows; TestSuiteStatsRerunIsEvidenceNotARun goes red on revert. The missed reader (loadSavedAttempt) is raised as a new Critical.
- BR-14 — addressed — README.md:506-512 and atlas/workflow/ci-merge-check.md:166-187 now describe the gate and the known_flakes values, matching the help text.
- BR-15 — addressed — TestParseLocalChecks covers the bad value, the off-suite key and the default; TestSuiteBaselineRetainsRealLockOwner asserts the persistent-main label.
- BR-16 — not-addressed — min(15m, timeout_minutes) is in the code, but no test pins the bound.
- BR-17 — addressed — RerunMissing plus go test -v; reverting the gate turns "a rerun that shows no PASS proves nothing" red. A bypass via saved-pass reuse is raised as new.
- BR-18 — addressed — The label is computed once inside the pre-existing branch; the reentry test asserts TestB appears exactly once.

### Raised

- **BR-19** [Critical] `new-record-kind-unswept-readers` (kind: defect) loadSavedAttempt reuses a rerun row as the head's suite pass, so merge posts success for an inconclusive rerun
  loadSavedAttempt filters only on SHA and check. A rerun row has Check suite and a later FinishedAt, and passed() is true on exit 0, while revalidateKnownFlake passes non-landing kinds through. Reproduced: suiteFailing plus a rerun printing [no tests to run]; run posts failure, then republishSavedPass returns ok=true and posts success, publishing the rerun's log as the suite's. 2nd in family. Rule: a RerunOf row is never a suite attempt for any reader; enforce it with one shared predicate or a distinct rerun representation, and decide reuse from the parent attempt's outcome. Add a regression test.

## Open findings

- **BR-9** [Minor] `cas-mutator-uses-stale-read` (kind: defect) The one-owner-per-test check reads the registry, not the card under compare-and-swap, so concurrent registrations on two issues can both pass
- **BR-10** [Minor] `breaking-change-undocumented` (kind: defect) The suite-attempt writer now records version 2, which older sdlc binaries reject as an unsupported version
- **BR-16** [Minor] `declared-envelope-not-enforced` (kind: defect) The rerun timeout is a fixed 15 minutes, not bounded by the check's timeout_minutes as the plan states
- **BR-19** [Critical] `new-record-kind-unswept-readers` (kind: defect) loadSavedAttempt reuses a rerun row as the head's suite pass, so merge posts success for an inconclusive rerun
