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

## Open findings

- **BR-2** [Important] `cas-mutator-uses-stale-read` (kind: defect) flake register computes the merged flake_tests from an earlier read than its compare-and-swap snapshot, so concurrent registrations are lost (ARCH-ORDER)
- **BR-8** [Minor] `duplicated-io-glue` (kind: defect) defaultFlakeDeps.registry duplicates loadFlakeRegistry, and attempts/logPath duplicate the store-dir setup (ARCH-DRY)
