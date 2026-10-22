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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 5af370ed-a007-4239-8c71-db0b25d9145e
            repository: github.com/xianxu/ariadne
            issue_id: "000553"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/02338d9c9a8acc8f
            reviewed_head: 954dea6bc67eecfd312bd92ccefa1affeb0ac827
            slot: ariadne:5
            requested_at: "2026-10-22T14:14:25.581358Z"
          phase: finished
          pid: 45058
          started_at: "2026-10-22T14:14:25.682028Z"
          finished_at: "2026-10-22T14:16:31.417191Z"
          session_id: 01a163f6-91ab-7410-81aa-a8a74d4b6613
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 954dea6bc67eecfd312bd92ccefa1affeb0ac827
        tracker_head: 1a2d2d252fa96aea23db7d93101e309b7a9b150b
        branch_point: e4bd907ae47d7a94b31b6067df0976f96eee3ade
        since: "2026-10-22T06:34:21-07:00"
        until: "2026-10-22T07:13:15-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T06:34:21-07:00"
              end: "2026-10-22T07:13:15-07:00"
              began: "2026-10-22T06:34:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 8f11f68826c087916ccd34de42f4d3c6b951a6c9afd99ab9d07f11224c8f356c
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 44502476
              windows:
                - start: "2026-10-22T06:34:21-07:00"
                  end: "2026-10-22T07:13:15-07:00"
        cuts:
            - dispatched: "2026-10-22T07:03:05-07:00"
              finalized: "2026-10-22T07:05:27-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-22T07:16:32-07:00"
      dispatched: "2026-10-22T07:14:25-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Both bookkeeping merge regressions pass on pinned HEAD and fail with the prior parser restored through a temporary Go overlay. Single-parent bookkeeping, foreign-parent, and train controls also pass.
          round: 2
      findings:
        - kind: defect
          id: BR-2
          severity: Critical
          title: Qualified invalid leads still bypass structural ancestry proof
          detail: 'cmd/sdlc/actualqueue_head_git.go:40 calls parseQueueHead(fact, ""), while membership classification uses the actual repository name. For ariadne#0, the reader sees definite foreign authority and skips ancestry proof; the classifier sees an invalid local lead and returns unknown with MainParents unset. Real-Git overlay cases fail for all three tested forms: ariadne#0 alone, ariadne#0 followed by #10, and #10 followed by ariadne#0. The affected class includes zero-padded zero IDs and qualified invalid refs anywhere in a recognized lead list, for any self-repository name. This is the 2nd finding in family message-independent-structural-attribution. Earlier rounds fixed instances. Do NOT fix only these spellings: enforce the rule that proof selection and membership classification use identical authority context, preserving definite authority and malformed-train precedence (ARCH-PURPOSE). Retain regression controls for this enumeration.'
          family: message-independent-structural-attribution
          round: 2
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: b1243126-de31-4afb-a41d-658ecfd4e69d
            repository: github.com/xianxu/ariadne
            issue_id: "000553"
            recipe: small-diff-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/c3aa707f7edada33
            reviewed_head: fe9a4c9e4baff739a319d5969a7adeb8ca683212
            slot: ariadne:5
            requested_at: "2026-10-22T14:25:44.286438Z"
          phase: finished
          pid: 86853
          started_at: "2026-10-22T14:25:44.426758Z"
          finished_at: "2026-10-22T14:28:54.026648Z"
          session_id: 01a16400-edbd-75a2-8458-69769e094846
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: fe9a4c9e4baff739a319d5969a7adeb8ca683212
        tracker_head: 1a2d2d252fa96aea23db7d93101e309b7a9b150b
        branch_point: e4bd907ae47d7a94b31b6067df0976f96eee3ade
        since: "2026-10-22T06:34:21-07:00"
        until: "2026-10-22T07:24:49-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T06:34:21-07:00"
              end: "2026-10-22T07:24:49-07:00"
              began: "2026-10-22T06:34:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 2d815c8cb7b6d31d9a205636adabf09e8053aaebb5ae0475e7366bfedefeda06
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 44902565
              windows:
                - start: "2026-10-22T06:34:21-07:00"
                  end: "2026-10-22T07:24:49-07:00"
        cuts:
            - dispatched: "2026-10-22T07:03:05-07:00"
              finalized: "2026-10-22T07:05:27-07:00"
              finalizing: false
            - dispatched: "2026-10-22T07:14:25-07:00"
              finalized: "2026-10-22T07:16:32-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-22T07:28:55-07:00"
      dispatched: "2026-10-22T07:25:44-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-2
          disposition: addressed
          note: actualqueue_head_git.go:40 and :169 classify with the same selfRepo passed through the typed reader interface. The 45-case real-Git matrix passes on pinned HEAD; restoring the empty proof-selection context through a temporary Go overlay makes all 45 cases fail. Definite-authority, foreign-parent, and malformed-train controls pass.
          round: 3
      recipe: small-diff-review
      isolation: fenced
      reviewed: fe9a4c9e4baff739a319d5969a7adeb8ca683212
      blocked: false
---

# Gate ledger — ariadne-553#553 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T07:05:27-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `message-independent-structural-attribution` Bookkeeping subjects bypass proven structural integration attribution
  cmd/sdlc/actualqueue_head.go:192–196 returns unknown before the new structural fallback for both instances in this family: "issue-sync: update issues" and "archive completed issues to history". A merge with an issue-attributable first parent and all side parents on pinned main therefore remains unknown solely because of its message (ARCH-PURPOSE). Added real-Git cases fail for both subjects; allowing the unsuccessful bookkeeping match to fall through makes the complete structural-control matrix pass. Preserve single-parent bookkeeping recognition and train precedence, allow structurally valid merges through, and retain regression cases for both subjects.

## Round 2 — 2026-10-22T07:16:32-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Both bookkeeping merge regressions pass on pinned HEAD and fail with the prior parser restored through a temporary Go overlay. Single-parent bookkeeping, foreign-parent, and train controls also pass.

### Raised

- **BR-2** [Critical] `message-independent-structural-attribution` (kind: defect) Qualified invalid leads still bypass structural ancestry proof
  cmd/sdlc/actualqueue_head_git.go:40 calls parseQueueHead(fact, ""), while membership classification uses the actual repository name. For ariadne#0, the reader sees definite foreign authority and skips ancestry proof; the classifier sees an invalid local lead and returns unknown with MainParents unset. Real-Git overlay cases fail for all three tested forms: ariadne#0 alone, ariadne#0 followed by #10, and #10 followed by ariadne#0. The affected class includes zero-padded zero IDs and qualified invalid refs anywhere in a recognized lead list, for any self-repository name. This is the 2nd finding in family message-independent-structural-attribution. Earlier rounds fixed instances. Do NOT fix only these spellings: enforce the rule that proof selection and membership classification use identical authority context, preserving definite authority and malformed-train precedence (ARCH-PURPOSE). Retain regression controls for this enumeration.

## Round 3 — 2026-10-22T07:28:55-07:00 (codex) — passed

### Disposed

- BR-2 — addressed — actualqueue_head_git.go:40 and :169 classify with the same selfRepo passed through the typed reader interface. The 45-case real-Git matrix passes on pinned HEAD; restoring the empty proof-selection context through a temporary Go overlay makes all 45 cases fail. Definite-authority, foreign-parent, and malformed-train controls pass.

## Open findings

(none — every finding has been disposed)
