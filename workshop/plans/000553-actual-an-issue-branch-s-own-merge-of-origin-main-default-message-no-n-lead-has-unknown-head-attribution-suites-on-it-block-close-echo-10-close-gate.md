---
gate: boundary-review
issue: 553
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d44de121-8df4-4df6-a18c-62e8623b9c94
            repository: github.com/xianxu/ariadne
            issue_id: "000553"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/da68810f4a8591fd
            reviewed_head: db9a3d2a43dfdc6f0743022b47dbaae326b13d07
            slot: ariadne:5
            requested_at: "2026-10-22T14:03:05.841506Z"
          phase: finished
          pid: 82990
          started_at: "2026-10-22T14:03:06.011797Z"
          finished_at: "2026-10-22T14:05:26.673335Z"
          session_id: 01a163ec-3379-7142-b581-650feb0e101e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: db9a3d2a43dfdc6f0743022b47dbaae326b13d07
        tracker_head: 1a2d2d252fa96aea23db7d93101e309b7a9b150b
        branch_point: e4bd907ae47d7a94b31b6067df0976f96eee3ade
        since: "2026-10-22T06:34:21-07:00"
        until: "2026-10-22T07:01:14-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T06:34:21-07:00"
              end: "2026-10-22T07:01:14-07:00"
              began: "2026-10-22T06:34:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: bcac952314e07a1caaa9dd08e443ad0b4ae4f51a8311cfa710426a14cfee447d
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 44070070
              windows:
                - start: "2026-10-22T06:34:21-07:00"
                  end: "2026-10-22T07:01:14-07:00"
      "n": 1
      timestamp: "2026-10-22T07:05:27-07:00"
      dispatched: "2026-10-22T07:03:05-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: Bookkeeping subjects bypass proven structural integration attribution
          detail: 'cmd/sdlc/actualqueue_head.go:192–196 returns unknown before the new structural fallback for both instances in this family: "issue-sync: update issues" and "archive completed issues to history". A merge with an issue-attributable first parent and all side parents on pinned main therefore remains unknown solely because of its message (ARCH-PURPOSE). Added real-Git cases fail for both subjects; allowing the unsuccessful bookkeeping match to fall through makes the complete structural-control matrix pass. Preserve single-parent bookkeeping recognition and train precedence, allow structurally valid merges through, and retain regression cases for both subjects.'
          family: message-independent-structural-attribution
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne-553#553 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T07:05:27-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `message-independent-structural-attribution` Bookkeeping subjects bypass proven structural integration attribution
  cmd/sdlc/actualqueue_head.go:192–196 returns unknown before the new structural fallback for both instances in this family: "issue-sync: update issues" and "archive completed issues to history". A merge with an issue-attributable first parent and all side parents on pinned main therefore remains unknown solely because of its message (ARCH-PURPOSE). Added real-Git cases fail for both subjects; allowing the unsuccessful bookkeeping match to fall through makes the complete structural-control matrix pass. Preserve single-parent bookkeeping recognition and train precedence, allow structurally valid merges through, and retain regression cases for both subjects.

## Open findings

- **BR-1** [Critical] `message-independent-structural-attribution` Bookkeeping subjects bypass proven structural integration attribution
