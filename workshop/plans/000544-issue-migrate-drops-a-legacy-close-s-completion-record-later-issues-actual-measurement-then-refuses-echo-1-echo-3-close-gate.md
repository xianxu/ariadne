---
gate: boundary-review
issue: 544
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 6cc64491-1677-4d7c-9e5b-980b3d51a3fe
            repository: github.com/xianxu/ariadne
            issue_id: "000544"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/024361c488a18dd0
            reviewed_head: 912d87634143d023fab266ded7e212c53082db70
            slot: ariadne:1
            requested_at: "2026-10-22T09:01:05.161845Z"
          phase: finished
          pid: 89706
          started_at: "2026-10-22T09:01:05.39856Z"
          finished_at: "2026-10-22T09:09:07.304627Z"
          session_id: 01a162d7-b377-7da1-8d72-74abf5a5fd94
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 912d87634143d023fab266ded7e212c53082db70
        tracker_head: 6a3d4b72c73ee54127467599c79b190aa0176bc2
        branch_point: 2dd8537609f7442797016b44bdf1e957de342ba5
        since: "2026-10-22T00:02:32-07:00"
        until: "2026-10-22T01:41:37-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T00:02:32-07:00"
              end: "2026-10-22T01:41:37-07:00"
              began: "2026-10-22T00:02:32-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 208b36c649dd4699f543f6cf92279a1ff156aa4b0f7be9ea3dcb082d1b8a70eb
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 65584650
              windows:
                - start: "2026-10-22T00:02:32-07:00"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a1626c-1a91-7e01-b21a-4f131070165b
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 2b9bff51e39c67cb39137014ea2049289d6a4d8fbae94e5e7400d1ed1e2a5515
              started: "2026-10-22T07:03:34.399Z"
              spawned: true
              bytes: 800497
              windows:
                - start: "2026-10-22T07:03:34.399Z"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a1626c-43f6-75e3-9968-d12e3a19c422
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 5383d8e584f3050957e7a71db426eaeb9c6d21c5cb42b2cfd3edd933ec9a257b
              started: "2026-10-22T07:03:44.992Z"
              spawned: true
              bytes: 597694
              windows:
                - start: "2026-10-22T07:03:44.992Z"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a16272-8090-73f2-ba81-77e28978e17c
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 1780f733c3d86e7f93f2efee0550b4e7e9900e611fe6348a31d57d7b9fb55008
              started: "2026-10-22T07:10:34.266Z"
              spawned: false
              bytes: 590520
              windows:
                - start: "2026-10-22T07:10:34.266Z"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a16274-a97e-73b3-b06a-933e39ee1bcc
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 4d0cc4cb97cb5e8072b70a26dd05bdf917abd6f7488b82206d23bd1d90097619
              started: "2026-10-22T07:12:55.599Z"
              spawned: false
              bytes: 388932
              windows:
                - start: "2026-10-22T07:12:55.599Z"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a16276-7782-7033-8ceb-4e2afe871c23
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 84f6d8139a795fc0f45f48c53046ffcee5a9c4e244a31bc77626c0f507f7d727
              started: "2026-10-22T07:14:53.928Z"
              spawned: false
              bytes: 446062
              windows:
                - start: "2026-10-22T07:14:53.928Z"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a16278-05e4-76a0-a7d1-bba7766466c4
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: ed09ce704552035c9d2052c2de119e880e531e45e97de90055b5b0aeb8db3c5a
              started: "2026-10-22T07:16:35.895Z"
              spawned: false
              bytes: 451758
              windows:
                - start: "2026-10-22T07:16:35.895Z"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a16278-88f7-7140-941b-a5b6f5ef1974
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: fbf540f33e186327cf28a6d3d169367353c93ff60ab30ff96b1aded3a87d913e
              started: "2026-10-22T07:17:09.451Z"
              spawned: false
              bytes: 419665
              windows:
                - start: "2026-10-22T07:17:09.451Z"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a16279-d678-7660-8790-536c9fd83e03
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: dcb8b268e0fa7411b09d1678e4d7226bdeb18cd6d0a3b580f167eaeb344b99e0
              started: "2026-10-22T07:18:34.528Z"
              spawned: true
              bytes: 1070182
              windows:
                - start: "2026-10-22T07:18:34.528Z"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a1627a-4865-7cb1-b889-9e318730d432
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: b2c22758f79bfa1dc87fd7ad55ff855a7c8b310f06f53fc75e1c54f60f25d673
              started: "2026-10-22T07:19:03.719Z"
              spawned: true
              bytes: 2070314
              windows:
                - start: "2026-10-22T07:19:03.719Z"
                  end: "2026-10-22T01:41:37-07:00"
            - harness: codex
              id: 01a1627a-8393-7781-957d-89729e223fa6
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: b913ce1f83a8ca02d4f72a5a5f53f3e39a4e2a665af95ed3954dbff08410cf64
              started: "2026-10-22T07:19:18.875Z"
              spawned: true
              bytes: 1966911
              windows:
                - start: "2026-10-22T07:19:18.875Z"
                  end: "2026-10-22T01:41:37-07:00"
      "n": 1
      timestamp: "2026-10-22T02:09:08-07:00"
      dispatched: "2026-10-22T02:01:05-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Hybrid measurements remain readable as incomplete measurements by older binaries
          detail: cmd/sdlc/internal/gatestate/measurement.go:48 adds legacy_prefix and legacy_cutover while cmd/sdlc/actualinputs.go:233 retains version 1 and the existing engagement method. A temporary regression using the pinned base decoder accepted a valid hybrid snapshot, passed Validate, and silently discarded both fields; the base computation consequently has only suffix evidence to replay, and rewrites lose the prefix permanently. ARCH-EVOLVE requires an explicit method/version boundary that older readers reject because ignoring this evidence changes the total. Add cross-version regression coverage alongside unknown-field preservation checks.
          family: persisted-measurement-compatibility
          round: 1
        - kind: defect
          id: BR-2
          severity: Critical
          title: Repair uses a weaker history fence than the consumer of its repaired evidence
          detail: cmd/sdlc/issuemigrate_repair.go:217 accepts histories containing removal of a matching completion and does not inspect provenance transitions. A production-fixture probe imported an old card, repaired it, restored its original unbound bytes, and successfully repaired it again; verifyLegacyAdoption then rejected that published card with “legacy adoption tracker history removes completion evidence” at legacycompletion.go:608. Consolidate repair and exclusion onto the same complete, monotone generation validator, including completion/provenance removal and replacement. Add sequence regressions through repair followed by actual exclusion. This violates ARCH-DRY, ARCH-ORDER, and ARCH-PURPOSE.
          family: legacy-generation-continuity
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#544 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T02:09:08-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `persisted-measurement-compatibility` (kind: defect) Hybrid measurements remain readable as incomplete measurements by older binaries
  cmd/sdlc/internal/gatestate/measurement.go:48 adds legacy_prefix and legacy_cutover while cmd/sdlc/actualinputs.go:233 retains version 1 and the existing engagement method. A temporary regression using the pinned base decoder accepted a valid hybrid snapshot, passed Validate, and silently discarded both fields; the base computation consequently has only suffix evidence to replay, and rewrites lose the prefix permanently. ARCH-EVOLVE requires an explicit method/version boundary that older readers reject because ignoring this evidence changes the total. Add cross-version regression coverage alongside unknown-field preservation checks.
- **BR-2** [Critical] `legacy-generation-continuity` (kind: defect) Repair uses a weaker history fence than the consumer of its repaired evidence
  cmd/sdlc/issuemigrate_repair.go:217 accepts histories containing removal of a matching completion and does not inspect provenance transitions. A production-fixture probe imported an old card, repaired it, restored its original unbound bytes, and successfully repaired it again; verifyLegacyAdoption then rejected that published card with “legacy adoption tracker history removes completion evidence” at legacycompletion.go:608. Consolidate repair and exclusion onto the same complete, monotone generation validator, including completion/provenance removal and replacement. Add sequence regressions through repair followed by actual exclusion. This violates ARCH-DRY, ARCH-ORDER, and ARCH-PURPOSE.

## Open findings

- **BR-1** [Critical] `persisted-measurement-compatibility` (kind: defect) Hybrid measurements remain readable as incomplete measurements by older binaries
- **BR-2** [Critical] `legacy-generation-continuity` (kind: defect) Repair uses a weaker history fence than the consumer of its repaired evidence
