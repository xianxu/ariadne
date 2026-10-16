---
gate: boundary-review
issue: 404
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 15d8660e-4fc6-4c09-8edd-593b3e85a473
            repository: github.com/xianxu/ariadne
            issue_id: "000404"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/003f954c8df6e8ac
            reviewed_head: a49380b2b5404f45e6deef1ef85b53c8295f84ac
            slot: ariadne:3
            requested_at: "2026-10-16T16:40:36.364378Z"
          phase: finished
          pid: 91578
          started_at: "2026-10-16T16:40:36.515997Z"
          finished_at: "2026-10-16T16:41:42.902533Z"
          session_id: 01a14596-3ece-78a0-a790-c1d4f0764341
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: a49380b2b5404f45e6deef1ef85b53c8295f84ac
        tracker_head: 9810284dca05164b0488516cb291a07beae3fa1b
        branch_point: 6a5ce138e751df52a0606eb851d190673d13beb8
        since: "2026-10-16T09:16:33-07:00"
        until: "2026-10-16T09:40:11-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T09:16:33-07:00"
              end: "2026-10-16T09:40:11-07:00"
              began: "2026-10-16T09:16:33-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: a7a76a9c726147e582d9f757846ee12317670b68af041ed678ce8e909fde01ac
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 1587332
              windows:
                - start: "2026-10-16T09:16:33-07:00"
                  end: "2026-10-16T09:40:11-07:00"
      "n": 1
      timestamp: "2026-10-16T09:41:44-07:00"
      dispatched: "2026-10-16T09:40:36-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: 'ARCH-DRY: both commands independently maintain the same planning contract'
          detail: The entire eight-line planning body is duplicated verbatim at cmd/tl-lead/help.md:66–73 and cmd/tl-worker/help.md:18–25; these are all instances in this window. These prompts drive agent behavior, so a future correction can leave the roles following different rules. Embed the common planning text from one shared source and compose it into both help surfaces, retaining independent output assertions for both commands.
          family: shared-contract-single-source
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: aefc6d28-d008-4174-ba4c-0a332ef8411e
            repository: github.com/xianxu/ariadne
            issue_id: "000404"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/6b8f39ca9915ba9b
            reviewed_head: f845245242b0895b87cdc1e688db15fb3f1750bc
            slot: ariadne:3
            requested_at: "2026-10-16T16:43:46.111625Z"
          phase: finished
          pid: 9849
          started_at: "2026-10-16T16:43:46.253643Z"
          finished_at: "2026-10-16T16:44:37.823416Z"
          session_id: 01a14599-2400-7361-837b-ab009d1cfe54
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: f845245242b0895b87cdc1e688db15fb3f1750bc
        tracker_head: 9810284dca05164b0488516cb291a07beae3fa1b
        branch_point: 6a5ce138e751df52a0606eb851d190673d13beb8
        since: "2026-10-16T09:16:33-07:00"
        until: "2026-10-16T09:43:14-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T09:16:33-07:00"
              end: "2026-10-16T09:43:14-07:00"
              began: "2026-10-16T09:16:33-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: c7047d45851817d332c085bba4a368e3c0a19564e6a0bc9aa973a1a3e4436936
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 1851541
              windows:
                - start: "2026-10-16T09:16:33-07:00"
                  end: "2026-10-16T09:43:14-07:00"
        cuts:
            - dispatched: "2026-10-16T09:40:36-07:00"
              finalized: "2026-10-16T09:41:44-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-16T09:44:41-07:00"
      dispatched: "2026-10-16T09:43:46-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: pkg/tlrules/planning.md now owns the complete contract; cmd/tl-lead/main.go:47 and cmd/tl-worker/main.go:34 both call ComposeHelp. Independent output assertions remain. In a disposable pinned-head copy, replacing ComposeHelp with an identity function made all four planning checks fail for both commands under --help and no-argument invocation; the copy was removed.
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: f845245242b0895b87cdc1e688db15fb3f1750bc
      blocked: false
---

# Gate ledger — ariadne#404 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T09:41:44-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `shared-contract-single-source` (kind: defect) ARCH-DRY: both commands independently maintain the same planning contract
  The entire eight-line planning body is duplicated verbatim at cmd/tl-lead/help.md:66–73 and cmd/tl-worker/help.md:18–25; these are all instances in this window. These prompts drive agent behavior, so a future correction can leave the roles following different rules. Embed the common planning text from one shared source and compose it into both help surfaces, retaining independent output assertions for both commands.

## Round 2 — 2026-10-16T09:44:41-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — pkg/tlrules/planning.md now owns the complete contract; cmd/tl-lead/main.go:47 and cmd/tl-worker/main.go:34 both call ComposeHelp. Independent output assertions remain. In a disposable pinned-head copy, replacing ComposeHelp with an identity function made all four planning checks fail for both commands under --help and no-argument invocation; the copy was removed.

## Open findings

(none — every finding has been disposed)
