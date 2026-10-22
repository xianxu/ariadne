---
gate: boundary-review
issue: 556
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 0bf5421d-be3c-4afd-b999-a81535afaa9d
            repository: github.com/xianxu/ariadne
            issue_id: "000556"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/04d23021471cfcdd
            reviewed_head: 337a9cb6f5eb0cfd07430333b1d961f90f58084e
            slot: ariadne:3
            requested_at: "2026-10-22T21:45:33.061585Z"
          phase: finished
          pid: 21188
          started_at: "2026-10-22T21:45:33.26658Z"
          finished_at: "2026-10-22T21:49:26.319171Z"
          session_id: 01a16593-979b-7120-a1b5-9f2499081837
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 337a9cb6f5eb0cfd07430333b1d961f90f58084e
        tracker_head: cf26cf9227cec304e182763edc62718f1bb478af
        branch_point: ffd859eb5d7298b0844bc1c6876edad9b7574fd2
        since: "2026-10-22T13:11:41-07:00"
        until: "2026-10-22T14:44:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T13:11:41-07:00"
              end: "2026-10-22T14:44:16-07:00"
              began: "2026-10-22T13:11:41-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: f6b4095b2f7e0632975ca2a0b7f8862fb72fe9f4ebc5fdd5498e0b8cff1f1a1d
              started: "2026-10-21T14:24:08.011Z"
              spawned: false
              bytes: 51410676
              windows:
                - start: "2026-10-22T13:11:41-07:00"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a1653e-4411-7831-9d5b-1813fc3e522f
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 0bde33775247350f13ab107315b69ab56d708f0356f0caf8cb6ea6fdb586f917
              started: "2026-10-22T20:12:22.044Z"
              spawned: true
              bytes: 752228
              windows:
                - start: "2026-10-22T20:12:22.044Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a1653e-60ee-7213-b204-5ae2542a899e
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 1628837dd3bf02f13896feffc734079c1fed194801f3828328a30f6a343a0246
              started: "2026-10-22T20:12:29.448Z"
              spawned: true
              bytes: 869759
              windows:
                - start: "2026-10-22T20:12:29.448Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a1654c-cfd0-7572-af52-de00856fd1e6
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6b0851e24ab66494f110c0169ea137dcedc8f872eaed3efcf09a04726d8c80ac
              started: "2026-10-22T20:28:15.604Z"
              spawned: false
              bytes: 809306
              windows:
                - start: "2026-10-22T20:28:15.604Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a16551-f979-7fa3-ac0c-51d8dcae9374
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6472574e46d5ff0d94582788c0b7db2e2726bb5db3405f3cad135f88680b160c
              started: "2026-10-22T20:33:54.093Z"
              spawned: false
              bytes: 508905
              windows:
                - start: "2026-10-22T20:33:54.093Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a16553-ed37-7112-966f-bf4812180f38
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 2e093a529e8715d8c56e0b8a94c50e4b8583fb77cd217adfe71428232040c60f
              started: "2026-10-22T20:36:02.188Z"
              spawned: false
              bytes: 445816
              windows:
                - start: "2026-10-22T20:36:02.188Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a16554-80c4-7241-8a61-38c69fe096d9
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 49accf3592e85cdc0df2b63bb214af4c1e8d01c1edd17688491161ddbd8aca1d
              started: "2026-10-22T20:36:39.781Z"
              spawned: false
              bytes: 467326
              windows:
                - start: "2026-10-22T20:36:39.781Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a16555-921c-7d40-8490-7e30b3f119ac
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: a9fd5cb1dfc8c1b098ba661c36e33e3d1c7f9cfd21f1f8adece5b1ec497f3cc6
              started: "2026-10-22T20:37:49.439Z"
              spawned: true
              bytes: 508110
              windows:
                - start: "2026-10-22T20:37:49.439Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a16556-69a2-7031-b722-3952fbfd8b4d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 1dc98f4baa932f44913d71982d1c474ea46f6d50817d3d3178a384a251b96b2f
              started: "2026-10-22T20:38:44.633Z"
              spawned: true
              bytes: 469381
              windows:
                - start: "2026-10-22T20:38:44.633Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a1655c-c7b8-7630-b557-133e3d96c99b
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 26ec003baf9b67d9417dfbd9d355ba71a184be97c8aa4fe5c7c041098f33f3f9
              started: "2026-10-22T20:45:41.915Z"
              spawned: true
              bytes: 546354
              windows:
                - start: "2026-10-22T20:45:41.915Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a1656c-0b84-7dd0-bea7-74bf5c980a12
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: bc97bd8837a4faeab22554dd61c4bb8fc180c8c0c34400ebed967801b3307d55
              started: "2026-10-22T21:02:22.281Z"
              spawned: true
              bytes: 442270
              windows:
                - start: "2026-10-22T21:02:22.281Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a16570-b18f-70f3-8927-d3e3e419857d
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 8e9380c4e1922e970960ab9f6134ee7d9ddcb20d0fd9be5983f8d5813a0c829a
              started: "2026-10-22T21:07:26.905Z"
              spawned: true
              bytes: 557425
              windows:
                - start: "2026-10-22T21:07:26.905Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a16577-95dc-7e60-bd1b-fd7b4d904a9b
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: d899d022e2eaaf9c2a5bf64b7a4b3fc050974bddbc25f446fcc1b3e4a95abcc5
              started: "2026-10-22T21:14:58.512Z"
              spawned: true
              bytes: 1102461
              windows:
                - start: "2026-10-22T21:14:58.512Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a16577-d01f-73d3-af37-16eff7930e2f
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: afce8cb1752a86c6985fc13567fc284dcd8b1d2e3383386448938b9a96ce2fc0
              started: "2026-10-22T21:15:13.402Z"
              spawned: true
              bytes: 1406948
              windows:
                - start: "2026-10-22T21:15:13.402Z"
                  end: "2026-10-22T14:44:16-07:00"
            - harness: codex
              id: 01a1657a-0d05-76a1-a001-d189ae418a2c
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: cd0e45ff12238ed3fa89935cc1bd6285f65208d6017472d35bd743196960e83e
              started: "2026-10-22T21:17:40.048Z"
              spawned: true
              bytes: 1324367
              windows:
                - start: "2026-10-22T21:17:40.048Z"
                  end: "2026-10-22T14:44:16-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 47.757466666666666
      "n": 1
      timestamp: "2026-10-22T14:49:30-07:00"
      dispatched: "2026-10-22T14:45:33-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: Typed registration accepts a targeted rerun as its confirming suite pass
          detail: cmd/sdlc/teststats.go:256 marks only affected checks as Subset. flakeEvidence includes successful suite-rerun records through isFlakeEvidence, so SameHeadFailPass accepts a matching v4 targeted child as confirmation for script/Lua registration without a later complete suite pass. This contradicts plan contract item 6. Preserve check scope in the shared evidence model and reject targeted confirmations for typed identities while retaining legacy Go behavior; add script and Lua regressions covering registration and Confirmed. ARCH-PURPOSE.
          family: typed-confirmation-requires-complete-suite
          round: 1
        - id: BR-2
          severity: Critical
          title: The Core concepts table names missing and incorrectly located entities
          detail: workshop/plans/000556-typed-flake-identities-plan.md:35 lists RerunPlan and RerunMissingJSON as modified in rerun.go, but that file is unchanged and RerunPlan is new in rerun_typed.go:12. Line 37 names typedAttemptCapability, which does not exist in the pinned code. Correct the inventory and verification strategies to identify the delivered helpers and their actual PURE/INTEGRATION classification, recording the correction under Revisions. Critical under the required Core concepts cross-check. ARCH-PURPOSE.
          family: core-concept-traceability
          round: 1
        - id: BR-3
          severity: Important
          title: README update is missing for typed identities and runner adoption
          detail: README.md:850 still documents Go-only registration and reruns, and README.md is unchanged in the review range. Add script/Lua identity syntax, flake run --unit, version-3 flake_runners configuration, and the fleet-upgrade/adoption prerequisite, linking the detailed atlas instructions. Help and atlas updates do not satisfy the separate README gate.
          family: user-surface-documentation
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#556 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T14:49:30-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `typed-confirmation-requires-complete-suite` Typed registration accepts a targeted rerun as its confirming suite pass
  cmd/sdlc/teststats.go:256 marks only affected checks as Subset. flakeEvidence includes successful suite-rerun records through isFlakeEvidence, so SameHeadFailPass accepts a matching v4 targeted child as confirmation for script/Lua registration without a later complete suite pass. This contradicts plan contract item 6. Preserve check scope in the shared evidence model and reject targeted confirmations for typed identities while retaining legacy Go behavior; add script and Lua regressions covering registration and Confirmed. ARCH-PURPOSE.
- **BR-2** [Critical] `core-concept-traceability` The Core concepts table names missing and incorrectly located entities
  workshop/plans/000556-typed-flake-identities-plan.md:35 lists RerunPlan and RerunMissingJSON as modified in rerun.go, but that file is unchanged and RerunPlan is new in rerun_typed.go:12. Line 37 names typedAttemptCapability, which does not exist in the pinned code. Correct the inventory and verification strategies to identify the delivered helpers and their actual PURE/INTEGRATION classification, recording the correction under Revisions. Critical under the required Core concepts cross-check. ARCH-PURPOSE.
- **BR-3** [Important] `user-surface-documentation` README update is missing for typed identities and runner adoption
  README.md:850 still documents Go-only registration and reruns, and README.md is unchanged in the review range. Add script/Lua identity syntax, flake run --unit, version-3 flake_runners configuration, and the fleet-upgrade/adoption prerequisite, linking the detailed atlas instructions. Help and atlas updates do not satisfy the separate README gate.

## Open findings

- **BR-1** [Critical] `typed-confirmation-requires-complete-suite` Typed registration accepts a targeted rerun as its confirming suite pass
- **BR-2** [Critical] `core-concept-traceability` The Core concepts table names missing and incorrectly located entities
- **BR-3** [Important] `user-surface-documentation` README update is missing for typed identities and runner adoption
