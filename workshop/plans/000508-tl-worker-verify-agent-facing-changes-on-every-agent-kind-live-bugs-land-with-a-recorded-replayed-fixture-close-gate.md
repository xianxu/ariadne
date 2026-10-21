---
gate: boundary-review
issue: 508
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 22491b38-4ec5-48a6-a021-0c483c2e9492
            repository: github.com/xianxu/ariadne
            issue_id: "000508"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/de81860f86943930
            reviewed_head: 07c888e640c0d215f8e069ef16f55dc6264eb9df
            slot: ariadne:5
            requested_at: "2026-10-21T15:46:44.129187Z"
          phase: finished
          pid: 47932
          started_at: "2026-10-21T15:46:44.230844Z"
          finished_at: "2026-10-21T15:48:13.447033Z"
          session_id: 01a15f24-b889-7640-8ef8-f5bd98e816b0
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 07c888e640c0d215f8e069ef16f55dc6264eb9df
        tracker_head: faff77da2f16e2d6d46ac2ad106500df001f2765
        branch_point: bef84c5c2cb245869723e7aef6640c90b157fdfd
        since: "2026-10-21T01:53:24-07:00"
        until: "2026-10-21T08:45:12-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T01:53:24-07:00"
              end: "2026-10-21T08:45:12-07:00"
              began: "2026-10-21T01:53:24-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 893481dca5e6d7b55853999c83be3acfaa9383f03a178fe7f796898955894cf8
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 15325634
              windows:
                - start: "2026-10-21T01:53:24-07:00"
                  end: "2026-10-21T08:45:12-07:00"
      "n": 1
      timestamp: "2026-10-21T08:48:14-07:00"
      dispatched: "2026-10-21T08:46:44-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      recipe: small-diff-review
      isolation: fenced
      reviewed: 07c888e640c0d215f8e069ef16f55dc6264eb9df
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 3ada0742-73ad-4b4f-a14f-7254f8f862b7
            repository: github.com/xianxu/ariadne
            issue_id: "000508"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/7ee10392ec374e8e
            reviewed_head: c7885cd4c826bd6e894ac7732283d51510c6418d
            slot: ariadne:5
            requested_at: "2026-10-21T19:45:26.452876Z"
          phase: finished
          pid: 14789
          started_at: "2026-10-21T19:45:26.552483Z"
          finished_at: "2026-10-21T19:46:55.97343Z"
          session_id: 01a15fff-4330-7d83-b213-d1ad811ba84e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: c7885cd4c826bd6e894ac7732283d51510c6418d
        tracker_head: 705062eb768281682799530fe4a3959e666a1aec
        branch_point: 510aab6ff3a55bfdf5472e45eedf5512f6d04420
        since: "2026-10-21T01:53:24-07:00"
        until: "2026-10-21T11:46:43-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T01:53:24-07:00"
              end: "2026-10-21T11:46:43-07:00"
              began: "2026-10-21T01:53:24-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: ac81afb69c7acd41085350cb4d5e86822a310343820e6c9353a9107ac3063789
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 17920617
              windows:
                - start: "2026-10-21T01:53:24-07:00"
                  end: "2026-10-21T11:46:43-07:00"
        queue_waits:
            - attempt_id: 8f216198d08914b50ddc8c2567941156
              slot: ariadne:5
              record_digest: 98ecb57e9a297610667aeb8c498ba32d9acceca859596475dc239b5760779bd6
              start: "2026-10-21T15:46:24.832536Z"
              end: "2026-10-21T16:08:17.992856Z"
              queue_seconds: 1313.16032
              raw_seconds: 1313.16032
              exact: true
        cuts:
            - dispatched: "2026-10-21T08:46:44-07:00"
              finalized: "2026-10-21T08:48:14-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-21T12:46:56-07:00"
      dispatched: "2026-10-21T12:45:26-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Important
          title: README update appears missing for the harness test command
          detail: cmd/tl-worker/help.md:46 newly directs workers to run `make test-couch-fleet` in Pair, but README.md is unchanged in this range and contains no corresponding usage guidance. Add a concise, Pair-scoped reference near the existing TL tooling documentation, linking to the updated atlas.
          family: user-facing-surface-documentation
          round: 2
      recipe: fix-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#508 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T08:48:14-07:00 (codex) — passed

## Round 2 — 2026-10-21T12:46:56-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-facing-surface-documentation` README update appears missing for the harness test command
  cmd/tl-worker/help.md:46 newly directs workers to run `make test-couch-fleet` in Pair, but README.md is unchanged in this range and contains no corresponding usage guidance. Add a concise, Pair-scoped reference near the existing TL tooling documentation, linking to the updated atlas.

## Open findings

- **BR-1** [Important] `user-facing-surface-documentation` README update appears missing for the harness test command
