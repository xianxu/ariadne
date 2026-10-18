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

## Open findings

- **BR-1** [Important] `argv-identifier-validation` (kind: defect) Test-name validation accepts '-' and '.'/'..' path segments, so RerunCommand can emit go test flags (ARCH-SECURE)
- **BR-2** [Important] `cas-mutator-uses-stale-read` (kind: defect) flake register computes the merged flake_tests from an earlier read than its compare-and-swap snapshot, so concurrent registrations are lost (ARCH-ORDER)
- **BR-3** [Important] `deferred-purpose-subset` (kind: defect) flake list omits owner, age and the log that D3 and Done-when 3 require (ARCH-PURPOSE)
- **BR-4** [Important] `docs-gate-readme` (kind: defect) README update missing for sdlc flake register / sdlc flake list
- **BR-5** [Minor] `evidence-recorded-verbatim` (kind: defect) --evidence with --issue N is only printed, not recorded as the plan says
- **BR-6** [Minor] `register-target-state-check` (kind: defect) flake register --issue N does not refuse a terminal (closed) issue
- **BR-7** [Minor] `unexplained-outcome-policy` (kind: defect) Unexplained treats failed-twice, too-many and rule-off attempts as explained once their failures are registered
