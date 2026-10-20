---
gate: boundary-review
issue: 494
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 3de28267-02fc-40b6-88b1-b4c7a977d59e
            repository: github.com/xianxu/ariadne
            issue_id: "000494"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/290a1e60eb02aaa2
            reviewed_head: 56a0a9c84276d7bb914673cec39fc6983424f55d
            slot: ariadne:1
            requested_at: "2026-10-20T13:54:47.73622Z"
          phase: finished
          pid: 11561
          started_at: "2026-10-20T13:54:47.903612Z"
          finished_at: "2026-10-20T13:57:45.832965Z"
          session_id: 01a15997-e170-7552-803f-7f29c8b22307
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 56a0a9c84276d7bb914673cec39fc6983424f55d
        tracker_head: b87aeaf90da3d46653d08b4a79e00208927424b0
        branch_point: dc76211e27578aca7aa62d434db9ca2bd43eed5b
        since: "2026-10-20T06:34:01-07:00"
        until: "2026-10-20T06:54:19-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T06:34:01-07:00"
              end: "2026-10-20T06:54:19-07:00"
              began: "2026-10-20T06:34:01-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 01cf322156be0f543b850058799973e209310b12807b244938ef3c9f188c9e46
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 17933733
              windows:
                - start: "2026-10-20T06:34:01-07:00"
                  end: "2026-10-20T06:54:19-07:00"
            - harness: codex
              id: 01a15987-902d-73f2-91b7-018e8795a468
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: aa63fccdcc635495609e2d61cfded7ccf3fe14cbfb231b82ac73d2300ea65139
              started: "2026-10-20T13:36:59.09Z"
              spawned: true
              bytes: 412469
              windows:
                - start: "2026-10-20T13:36:59.09Z"
                  end: "2026-10-20T06:54:19-07:00"
            - harness: codex
              id: 01a15989-d3c8-7ee2-ac79-bbae313eefb4
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 9e4c160056e551f11ec9c8f671cda9fdaead218840d661e07366c2b77f1bfe16
              started: "2026-10-20T13:39:27.486Z"
              spawned: true
              bytes: 512867
              windows:
                - start: "2026-10-20T13:39:27.486Z"
                  end: "2026-10-20T06:54:19-07:00"
            - harness: codex
              id: 01a15990-dd6f-7fb0-b4f9-4633f6b8d536
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: ecafaae3b5a79142257325c4eebccf6c0c08525aa6229483116defaf5a79e6de
              started: "2026-10-20T13:47:08.625Z"
              spawned: true
              bytes: 434978
              windows:
                - start: "2026-10-20T13:47:08.625Z"
                  end: "2026-10-20T06:54:19-07:00"
      "n": 1
      timestamp: "2026-10-20T06:57:47-07:00"
      dispatched: "2026-10-20T06:54:47-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      recipe: small-diff-review
      isolation: fenced
      reviewed: 56a0a9c84276d7bb914673cec39fc6983424f55d
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 78aae31f-417c-4df2-a0da-364eb7144f9a
            repository: github.com/xianxu/ariadne
            issue_id: "000494"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/dd29bca80050a876
            reviewed_head: 058f4f15587adf7659fed2095b9b8f9750a1de02
            slot: ariadne:1
            requested_at: "2026-10-20T20:03:44.821981Z"
          phase: finished
          pid: 11322
          started_at: "2026-10-20T20:03:44.999779Z"
          finished_at: "2026-10-20T20:05:55.456497Z"
          session_id: 01a15ae9-aaab-72e3-845d-46be382b4fe1
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 058f4f15587adf7659fed2095b9b8f9750a1de02
        tracker_head: e9e6ec35026a49a40a827ddf0ce29bccfd758fcd
        branch_point: 8a64a18ecef987a8d09cd4af0c34bd5f3f6fe24e
        since: "2026-10-20T06:34:01-07:00"
        until: "2026-10-20T13:02:49-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T06:34:01-07:00"
              end: "2026-10-20T13:02:49-07:00"
              began: "2026-10-20T06:34:01-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 46a22a4efeb7410abd6243ba94794164d2bef0f325bfaa102fefa1f61e12b25e
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 24311511
              windows:
                - start: "2026-10-20T06:34:01-07:00"
                  end: "2026-10-20T13:02:49-07:00"
            - harness: codex
              id: 01a15987-902d-73f2-91b7-018e8795a468
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: aa63fccdcc635495609e2d61cfded7ccf3fe14cbfb231b82ac73d2300ea65139
              started: "2026-10-20T13:36:59.09Z"
              spawned: true
              bytes: 412469
              windows:
                - start: "2026-10-20T13:36:59.09Z"
                  end: "2026-10-20T13:02:49-07:00"
            - harness: codex
              id: 01a15989-d3c8-7ee2-ac79-bbae313eefb4
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 9e4c160056e551f11ec9c8f671cda9fdaead218840d661e07366c2b77f1bfe16
              started: "2026-10-20T13:39:27.486Z"
              spawned: true
              bytes: 512867
              windows:
                - start: "2026-10-20T13:39:27.486Z"
                  end: "2026-10-20T13:02:49-07:00"
            - harness: codex
              id: 01a15990-dd6f-7fb0-b4f9-4633f6b8d536
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: ecafaae3b5a79142257325c4eebccf6c0c08525aa6229483116defaf5a79e6de
              started: "2026-10-20T13:47:08.625Z"
              spawned: true
              bytes: 434978
              windows:
                - start: "2026-10-20T13:47:08.625Z"
                  end: "2026-10-20T13:02:49-07:00"
            - harness: codex
              id: 01a159e8-298d-7f52-a603-c9c35b716c2a
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 2caa53b33cb5d8029afe9b057070f46bbbc2c19fb6dab328a05e98df59de3031
              started: "2026-10-20T15:22:29.718Z"
              spawned: true
              bytes: 537312
              windows:
                - start: "2026-10-20T15:22:29.718Z"
                  end: "2026-10-20T13:02:49-07:00"
            - harness: codex
              id: 01a15ae3-7ae6-7872-898c-580cdcac92bd
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              parent_id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              digest: 50acdd0bcda9260c327e4b964acce32b569d20b7f76175d4195eb0b0b46e781a
              started: "2026-10-20T19:57:00.112Z"
              spawned: true
              bytes: 449797
              windows:
                - start: "2026-10-20T19:57:00.112Z"
                  end: "2026-10-20T13:02:49-07:00"
        queue_waits:
            - attempt_id: 11cab56f89656e3bab43c5a95e196394
              slot: ariadne:1
              record_digest: a14a1fd38d2cddeec4a4bbcc47b2d79676fec93c00afba725105ea7c53ff8acf
              start: "2026-10-20T13:53:08.007281Z"
              end: "2026-10-20T14:26:34.902966Z"
              queue_seconds: 2006.895685
              raw_seconds: 2006.895685
              exact: true
            - attempt_id: 5e357c6c26d3f4022155d9300359e757
              slot: ariadne:1
              record_digest: 9c0315c2d8bda80ea34a3d63dba272e2752297584c7ad17cbaab1da00f72bde5
              start: "2026-10-20T14:40:00.049669Z"
              end: "2026-10-20T14:55:12.813107Z"
              queue_seconds: 912.763438
              raw_seconds: 912.763438
              exact: true
            - attempt_id: b403134eaf325843104a5d4affdff08e
              slot: ariadne:1
              record_digest: 913ed486ec443dfaacaf4d46395b63498597ee13e9573c5053e4c284327cd5e8
              start: "2026-10-20T17:57:54.109065Z"
              end: "2026-10-20T18:09:37.248032Z"
              queue_seconds: 703.138967
              raw_seconds: 703.138967
              exact: true
            - attempt_id: 3f382a7f2b15693b8f3cf2f6a3bc1f41
              slot: ariadne:1
              record_digest: 5c521bc5142abcb69d69e0acf956199798825cbd9be7dac7486b1f266db7c43b
              start: "2026-10-20T19:37:34.742541Z"
              end: "2026-10-20T19:43:45.495171Z"
              queue_seconds: 370.75263
              raw_seconds: 370.75263
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 43.6245
        cuts:
            - dispatched: "2026-10-20T06:54:47-07:00"
              finalized: "2026-10-20T06:57:47-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-20T13:05:56-07:00"
      dispatched: "2026-10-20T13:03:44-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      recipe: fix-review
      isolation: fenced
      reviewed: 058f4f15587adf7659fed2095b9b8f9750a1de02
      blocked: false
---

# Gate ledger — ariadne#494 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T06:57:47-07:00 (codex) — passed

## Round 2 — 2026-10-20T13:05:56-07:00 (codex) — passed

## Open findings

(none — every finding has been disposed)
