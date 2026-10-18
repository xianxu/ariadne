---
gate: boundary-review
issue: 392
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 49dd21a3-fb69-4088-afe2-636d048e9f91
            repository: github.com/xianxu/ariadne
            issue_id: "000392"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/db824b5e234d4c2b
            reviewed_head: fb020b30e17631086f207423ca8f7fe816f8393c
            slot: ariadne:1
            requested_at: "2026-10-18T06:35:26.311244Z"
          phase: finished
          pid: 68460
          started_at: "2026-10-18T06:35:26.543529Z"
          finished_at: "2026-10-18T06:39:15.83681Z"
          session_id: 01a14db8-eb42-7903-8efb-4a97ccd1bc08
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: fb020b30e17631086f207423ca8f7fe816f8393c
        tracker_head: 7125e51ec0b3295d2dd23e4ceeb39ac5e79b4a1c
        branch_point: 826a55dd130c59a4ae255c147ffafb3e7a2f2c1b
        since: "2026-10-17T21:39:59-07:00"
        until: "2026-10-17T22:59:04-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T21:39:59-07:00"
              end: "2026-10-17T22:59:04-07:00"
              began: "2026-10-17T21:39:59-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 0a61c96f8663d6c1299e85e2da3dd0827c15ace36466310b79f8d1384c4ba49b
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 86288849
              windows:
                - start: "2026-10-17T21:39:59-07:00"
                  end: "2026-10-17T22:59:04-07:00"
            - harness: codex
              id: 01a14d52-f902-7160-ae71-1d6bd5ab83ff
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: e700e45a097bc930282b19bca022b8a402a7bd8aa22228998eb9be94b165a57c
              started: "2026-10-18T04:44:06.484Z"
              spawned: false
              bytes: 643546
              windows:
                - start: "2026-10-18T04:44:06.484Z"
                  end: "2026-10-17T22:59:04-07:00"
            - harness: codex
              id: 01a14d54-ead2-7b20-9c53-1b5e41ff0a87
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 183a0a8310c77c5ca4d3b213fe00d9ec7f456478a5df65dafc28c5881df5993c
              started: "2026-10-18T04:46:13.747Z"
              spawned: false
              bytes: 470979
              windows:
                - start: "2026-10-18T04:46:13.747Z"
                  end: "2026-10-17T22:59:04-07:00"
            - harness: codex
              id: 01a14d55-c2d0-7f43-a126-be76a08cf111
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: a1ac38d2c9c59fc7d56886bf55a108f4775c048d73eab3ccfa8164aed61ef357
              started: "2026-10-18T04:47:09.105Z"
              spawned: false
              bytes: 466461
              windows:
                - start: "2026-10-18T04:47:09.105Z"
                  end: "2026-10-17T22:59:04-07:00"
            - harness: codex
              id: 01a14d56-3dca-75a2-948b-5e7e712adae8
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: e56ca979b8be42265aa0d96251a325cd4072d997f88fea885c019fef0862b751
              started: "2026-10-18T04:47:40.659Z"
              spawned: false
              bytes: 387320
              windows:
                - start: "2026-10-18T04:47:40.659Z"
                  end: "2026-10-17T22:59:04-07:00"
            - harness: codex
              id: 01a14d56-bf3b-7c82-8b91-5a5be2fb1f2b
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: 86185125629b7d383c00e7b6238a7c623cef4dcc5bf76e29919efd5c38de33c2
              started: "2026-10-18T04:48:13.227Z"
              spawned: true
              bytes: 490250
              windows:
                - start: "2026-10-18T04:48:13.227Z"
                  end: "2026-10-17T22:59:04-07:00"
            - harness: codex
              id: 01a14d56-cf5b-7db1-9543-eb45bb98cbcb
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: 4fa4085d29254011615cdf6ef16aefce8c1756d391690aa8030d943176c36638
              started: "2026-10-18T04:48:17.367Z"
              spawned: true
              bytes: 444810
              windows:
                - start: "2026-10-18T04:48:17.367Z"
                  end: "2026-10-17T22:59:04-07:00"
            - harness: codex
              id: 01a14d57-7379-7843-b4f2-4ca6b77c3594
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              digest: 9e52c69d1a95ae4ba9912d4022c6b3fd5695db1b8b262b8c18badba1bde892da
              started: "2026-10-18T04:48:59.355Z"
              spawned: true
              bytes: 822077
              windows:
                - start: "2026-10-18T04:48:59.355Z"
                  end: "2026-10-17T22:59:04-07:00"
        queue_waits:
            - attempt_id: 9d175589065c9c45d54faa4ad4b66fc3
              slot: ariadne:1
              record_digest: 24cd3ec4ce03d1a40a52337e6ab9403a4c312f82788306e3a4262d2d5b578b4e
              start: "2026-10-18T05:33:10.08748Z"
              end: "2026-10-18T05:48:28.13186Z"
              queue_seconds: 918.04438
              raw_seconds: 918.04438
              exact: true
      "n": 1
      timestamp: "2026-10-17T23:39:16-07:00"
      dispatched: "2026-10-17T23:35:26-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Expired source history can be certified as a complete numeric ratio.
          detail: cmd/sdlc/weeklyattribution.go:326 publishes a ratio whenever no explicit warning exists, but the reader supplies no retention or collection-coverage boundary. A temporary regression test with a first enqueue 40 days before the week and no retained attempts produces ratio=0 with no unknown coverage, despite the plan at line 21 requiring null when earlier evidence might have expired. Existing pruning removes old attempt records at cmd/sdlc/teststats_store.go:473. Carry explicit history coverage into validation and require it to cover the entire cohort interval before certifying a ratio; test expired attempts and potentially expired earlier admissions (ARCH-SECURE, ARCH-PURPOSE).
          family: evidence-history-completeness
          round: 1
        - id: BR-2
          severity: Critical
          title: Earlier enqueue validation skips intermediate retries after finding the earliest enqueue.
          detail: cmd/sdlc/weeklyattribution.go:193 compares each prior enqueue against enqueued, which line 216 mutates during the scan. For chronological events containing a refused enqueue at -3h, an unrefused enqueue at -2h, and the successful enqueue at -1h, the first record moves the threshold to -3h and the incomplete intermediate ticket is skipped. A temporary regression test confirms that this contradictory lifecycle is accepted without unknown coverage. Enumerate all admissions against the original successful enqueue boundary, validate each independently, then calculate the earliest timestamp; test input permutations and multiple retries (ARCH-ORDER, ARCH-SECURE).
          family: evidence-boundary-validation
          round: 1
        - id: BR-3
          severity: Critical
          title: Legacy attempts counted in weekly totals vanish from reconciliation.
          detail: cmd/sdlc/weeklyattribution.go:262 selects in-week rows using FinishedAt directly, while aggregateWeeklyRepo uses suiteAttemptCompletedAt at cmd/sdlc/weeklystats.go:101. Production legacy records populate StartedAt but leave FinishedAt zero, so they contribute to physical suite totals yet never enter unassignable_attempts or unknown_attempts. A temporary regression test produces physical runs=1 and unassignable=0. Reconcile the same selected population while preserving the legacy timestamp's uncertainty and excluding it from exact attribution (ARCH-PURPOSE).
          family: metric-population-consistency
          round: 1
        - id: BR-4
          severity: Important
          title: README documentation is missing for the new snapshot command.
          detail: README.md:498 documents only the existing test-stats view; README.md is unchanged in the pinned range despite adding snapshot, --end, and --brain-root. Add invocation examples, weekly-boundary behavior, the immutable output path, and coverage/conflict semantics. Embedded help and the atlas update do not satisfy the explicit README gate.
          family: readme-surface-documentation
          round: 1
        - id: BR-5
          severity: Important
          title: The durable pilot snapshot and rollout acceptance evidence are still outstanding.
          detail: workshop/issues/000392-weekly-test-economy-snapshot.md:73 explicitly says the pilot wrote only to a temporary directory, and line 75 describes the saved snapshot as future work. No durable snapshot link, ops-project confirmation, or successful execution from both propagated pilot checkouts is recorded, although Done when lines 33–36 and plan line 68 require these. Complete the capture and record its exact path/commit, source coverage, stable-rerun evidence, project link confirmation, and pilot consumer verification before whole-issue close (ARCH-PURPOSE, ARCH-MOCK).
          family: default-rollout-completeness
          round: 1
        - id: BR-6
          severity: Important
          title: Git collection expands beyond the declared weekly cohort.
          detail: cmd/sdlc/weeklystats_inputs.go:191 inventories every reachable historical merge, and lines 246–282 request graphs for every completion with any retained landed event, irrespective of the requested week. This contradicts plan line 41's bound by weekly landed commits and candidate attempts; an irrelevant historical graph failure also nulls the requested week's ratio through global CoverageUnknown. Pass the window into collection, restrict graph work to relevant candidates, and test that unrelated historical landings cannot consume the budget or invalidate current coverage (ARCH-CONSTRAINTS).
          family: explicit-operating-envelope
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#392 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T23:39:16-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `evidence-history-completeness` Expired source history can be certified as a complete numeric ratio.
  cmd/sdlc/weeklyattribution.go:326 publishes a ratio whenever no explicit warning exists, but the reader supplies no retention or collection-coverage boundary. A temporary regression test with a first enqueue 40 days before the week and no retained attempts produces ratio=0 with no unknown coverage, despite the plan at line 21 requiring null when earlier evidence might have expired. Existing pruning removes old attempt records at cmd/sdlc/teststats_store.go:473. Carry explicit history coverage into validation and require it to cover the entire cohort interval before certifying a ratio; test expired attempts and potentially expired earlier admissions (ARCH-SECURE, ARCH-PURPOSE).
- **BR-2** [Critical] `evidence-boundary-validation` Earlier enqueue validation skips intermediate retries after finding the earliest enqueue.
  cmd/sdlc/weeklyattribution.go:193 compares each prior enqueue against enqueued, which line 216 mutates during the scan. For chronological events containing a refused enqueue at -3h, an unrefused enqueue at -2h, and the successful enqueue at -1h, the first record moves the threshold to -3h and the incomplete intermediate ticket is skipped. A temporary regression test confirms that this contradictory lifecycle is accepted without unknown coverage. Enumerate all admissions against the original successful enqueue boundary, validate each independently, then calculate the earliest timestamp; test input permutations and multiple retries (ARCH-ORDER, ARCH-SECURE).
- **BR-3** [Critical] `metric-population-consistency` Legacy attempts counted in weekly totals vanish from reconciliation.
  cmd/sdlc/weeklyattribution.go:262 selects in-week rows using FinishedAt directly, while aggregateWeeklyRepo uses suiteAttemptCompletedAt at cmd/sdlc/weeklystats.go:101. Production legacy records populate StartedAt but leave FinishedAt zero, so they contribute to physical suite totals yet never enter unassignable_attempts or unknown_attempts. A temporary regression test produces physical runs=1 and unassignable=0. Reconcile the same selected population while preserving the legacy timestamp's uncertainty and excluding it from exact attribution (ARCH-PURPOSE).
- **BR-4** [Important] `readme-surface-documentation` README documentation is missing for the new snapshot command.
  README.md:498 documents only the existing test-stats view; README.md is unchanged in the pinned range despite adding snapshot, --end, and --brain-root. Add invocation examples, weekly-boundary behavior, the immutable output path, and coverage/conflict semantics. Embedded help and the atlas update do not satisfy the explicit README gate.
- **BR-5** [Important] `default-rollout-completeness` The durable pilot snapshot and rollout acceptance evidence are still outstanding.
  workshop/issues/000392-weekly-test-economy-snapshot.md:73 explicitly says the pilot wrote only to a temporary directory, and line 75 describes the saved snapshot as future work. No durable snapshot link, ops-project confirmation, or successful execution from both propagated pilot checkouts is recorded, although Done when lines 33–36 and plan line 68 require these. Complete the capture and record its exact path/commit, source coverage, stable-rerun evidence, project link confirmation, and pilot consumer verification before whole-issue close (ARCH-PURPOSE, ARCH-MOCK).
- **BR-6** [Important] `explicit-operating-envelope` Git collection expands beyond the declared weekly cohort.
  cmd/sdlc/weeklystats_inputs.go:191 inventories every reachable historical merge, and lines 246–282 request graphs for every completion with any retained landed event, irrespective of the requested week. This contradicts plan line 41's bound by weekly landed commits and candidate attempts; an irrelevant historical graph failure also nulls the requested week's ratio through global CoverageUnknown. Pass the window into collection, restrict graph work to relevant candidates, and test that unrelated historical landings cannot consume the budget or invalidate current coverage (ARCH-CONSTRAINTS).

## Open findings

- **BR-1** [Critical] `evidence-history-completeness` Expired source history can be certified as a complete numeric ratio.
- **BR-2** [Critical] `evidence-boundary-validation` Earlier enqueue validation skips intermediate retries after finding the earliest enqueue.
- **BR-3** [Critical] `metric-population-consistency` Legacy attempts counted in weekly totals vanish from reconciliation.
- **BR-4** [Important] `readme-surface-documentation` README documentation is missing for the new snapshot command.
- **BR-5** [Important] `default-rollout-completeness` The durable pilot snapshot and rollout acceptance evidence are still outstanding.
- **BR-6** [Important] `explicit-operating-envelope` Git collection expands beyond the declared weekly cohort.
