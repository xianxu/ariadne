---
gate: boundary-review
issue: 543
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 32790977-e9cb-4c29-8c50-5f680f5c59c8
            repository: github.com/xianxu/ariadne
            issue_id: "000543"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/5ad7586956a8926b
            reviewed_head: cf6b8a95b62cb4547f2ec5a8031ab72187619115
            slot: ariadne:3
            requested_at: "2026-10-22T06:18:30.566774Z"
          phase: finished
          pid: 81680
          started_at: "2026-10-22T06:18:30.772652Z"
          finished_at: "2026-10-22T06:21:43.262979Z"
          session_id: 01a16242-dbb6-7432-a192-5fdfdbd64e3e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: cf6b8a95b62cb4547f2ec5a8031ab72187619115
        tracker_head: 31dd15bcb583ebcd5cb609c8fd562a142b782686
        branch_point: a4600f5050712a27c7979cfd035cccd8b5a1ceca
        since: "2026-10-21T22:43:36-07:00"
        until: "2026-10-21T23:18:01-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T22:43:36-07:00"
              end: "2026-10-21T23:18:01-07:00"
              began: "2026-10-21T22:43:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6ba73a5ee7f8e38a0200111a1dc5bc3a81cb87d33f3119e7f874935c96c057d2
              started: "2026-10-21T14:24:08.011Z"
              spawned: false
              bytes: 29078723
              windows:
                - start: "2026-10-21T22:43:36-07:00"
                  end: "2026-10-21T23:18:01-07:00"
            - harness: codex
              id: 01a16223-465f-7072-8a6d-000b13cd7d1f
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 19fc09b896a7525e7a128b9022e4c0378bdacac76b3fdbb72e212653c928e591
              started: "2026-10-22T05:44:01.478Z"
              spawned: true
              bytes: 906026
              windows:
                - start: "2026-10-22T05:44:01.478Z"
                  end: "2026-10-21T23:18:01-07:00"
            - harness: codex
              id: 01a16228-99a6-7fa1-b81c-ce96606a3570
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 3b346945906489566e93112696f2ab52099fd105e4099321862e028143892c2b
              started: "2026-10-22T05:49:50.708Z"
              spawned: false
              bytes: 654735
              windows:
                - start: "2026-10-22T05:49:50.708Z"
                  end: "2026-10-21T23:18:01-07:00"
            - harness: codex
              id: 01a1622b-9e56-77e3-9f46-1cdfdd4d1f74
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 509abcda94a2f6a4d8ddf1031ec804142396bf64286dea86ba15cc74c1b705e2
              started: "2026-10-22T05:53:08.564Z"
              spawned: false
              bytes: 517988
              windows:
                - start: "2026-10-22T05:53:08.564Z"
                  end: "2026-10-21T23:18:01-07:00"
            - harness: codex
              id: 01a1622d-d3a2-7533-ae5e-39ecf8100f7a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 024b057aedf79116cd6bb15f171cd67034c1fe7d75f3683aff1310c618d4c466
              started: "2026-10-22T05:55:33.176Z"
              spawned: false
              bytes: 468701
              windows:
                - start: "2026-10-22T05:55:33.176Z"
                  end: "2026-10-21T23:18:01-07:00"
            - harness: codex
              id: 01a1622e-518e-78d3-8696-6dc1300c5de2
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 624647cc1e9d8dfc3d5e546c6d0a0a66227eefd9679aa5170d9b266d457cac86
              started: "2026-10-22T05:56:05.446Z"
              spawned: false
              bytes: 407857
              windows:
                - start: "2026-10-22T05:56:05.446Z"
                  end: "2026-10-21T23:18:01-07:00"
            - harness: codex
              id: 01a1622f-5ea5-7f63-8b0d-e1ceaf9f513c
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 407ce72e62eb8fdc2acaf3a7885b5c779750aa038bfc691a8e48cea192ad8268
              started: "2026-10-22T05:57:14.066Z"
              spawned: true
              bytes: 662894
              windows:
                - start: "2026-10-22T05:57:14.066Z"
                  end: "2026-10-21T23:18:01-07:00"
      "n": 1
      timestamp: "2026-10-21T23:21:44-07:00"
      dispatched: "2026-10-21T23:18:30-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Initialization treats cached remote branches as the complete publication inventory
          detail: cmd/sdlc/issueinit.go:166-172 authorizes empty initialization using migrationInventory, whose migrationBranchRefs implementation (issuemigrate.go:219-237) enumerates cached refs/remotes entries. The preceding snapshots fetch only main and issue-tracker. In a disposable two-clone reproduction, clone B pushed legacy-work containing workshop/issues/000001-old.md after clone A last fetched; clone A's issue init returned 0 and published an empty tracker and cutover marker. This violates the complete-inventory refusal contract (ARCH-PURPOSE, ARCH-SECURE, ARCH-ORDER). Refresh and pin the publication remote's branch inventory before authorizing initialization, refuse failed observations, and add regressions for remote branches created or advanced since the last fetch.
          family: authoritative-inventory-completeness
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#543 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T23:21:44-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `authoritative-inventory-completeness` (kind: defect) Initialization treats cached remote branches as the complete publication inventory
  cmd/sdlc/issueinit.go:166-172 authorizes empty initialization using migrationInventory, whose migrationBranchRefs implementation (issuemigrate.go:219-237) enumerates cached refs/remotes entries. The preceding snapshots fetch only main and issue-tracker. In a disposable two-clone reproduction, clone B pushed legacy-work containing workshop/issues/000001-old.md after clone A last fetched; clone A's issue init returned 0 and published an empty tracker and cutover marker. This violates the complete-inventory refusal contract (ARCH-PURPOSE, ARCH-SECURE, ARCH-ORDER). Refresh and pin the publication remote's branch inventory before authorizing initialization, refuse failed observations, and add regressions for remote branches created or advanced since the last fetch.

## Open findings

- **BR-1** [Critical] `authoritative-inventory-completeness` (kind: defect) Initialization treats cached remote branches as the complete publication inventory
