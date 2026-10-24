---
gate: boundary-review
issue: 580
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 4698faab-b6cf-401f-9716-a2124871edff
            repository: github.com/xianxu/ariadne
            issue_id: "000580"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/c5dff2b32ce509f2
            reviewed_head: 8d3d5948cfdfaf4eefa19957d73be00e09442d4e
            slot: ariadne:3
            requested_at: "2026-10-24T03:46:53.993784Z"
          phase: finished
          pid: 71707
          started_at: "2026-10-24T03:46:54.106611Z"
          finished_at: "2026-10-24T03:51:24.903377Z"
          session_id: 01a16c04-c552-7e60-a70d-b66340c5d8f3
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 8d3d5948cfdfaf4eefa19957d73be00e09442d4e
        tracker_head: fa1f09c67d791c4f7663565d9ed41b3aa2706823
        branch_point: 84b68e2fc5c0773ff38f13771507d93aaa08efd4
        since: "2026-10-23T19:27:24-07:00"
        until: "2026-10-23T20:46:18-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T19:27:24-07:00"
              end: "2026-10-23T20:46:18-07:00"
              began: "2026-10-23T19:27:24-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: bd04f53300c518ac99422b89e07fdb15d2b4a2fe7104e5014fbcd6b8825614fa
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 143419147
              windows:
                - start: "2026-10-23T19:27:24-07:00"
                  end: "2026-10-23T20:46:18-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.07971666666666667
      "n": 1
      timestamp: "2026-10-23T20:51:25-07:00"
      dispatched: "2026-10-23T20:46:53-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Ownership-only replay accepts a witnessed foreign issue branch
          detail: 'cmd/sdlc/actualbranch.go:311–319 validates branch/foreign-branch identity but accepts ownership-only intervals regardless of their canonical branch ID; measurement_branch.go:85 also leaves this case unchecked. A temporary test overlay relabelled the incident fixture''s foreign-branch interval as ownership-only and widened its saved selected window: resolveOwnedReplay and computeOwnedActualFor accepted it, crediting B with 29 minutes instead of 11. ARCH-SECURE requires contradictory persisted evidence to fail visibly. Reject ownership-only evidence naming any canonical issue branch, matching SelectBranchActivity''s rule, and add a regression through saved-input replay.'
          family: retained-attribution-consistency
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#580 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T20:51:25-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `retained-attribution-consistency` (kind: defect) Ownership-only replay accepts a witnessed foreign issue branch
  cmd/sdlc/actualbranch.go:311–319 validates branch/foreign-branch identity but accepts ownership-only intervals regardless of their canonical branch ID; measurement_branch.go:85 also leaves this case unchecked. A temporary test overlay relabelled the incident fixture's foreign-branch interval as ownership-only and widened its saved selected window: resolveOwnedReplay and computeOwnedActualFor accepted it, crediting B with 29 minutes instead of 11. ARCH-SECURE requires contradictory persisted evidence to fail visibly. Reject ownership-only evidence naming any canonical issue branch, matching SelectBranchActivity's rule, and add a regression through saved-input replay.

## Open findings

- **BR-1** [Important] `retained-attribution-consistency` (kind: defect) Ownership-only replay accepts a witnessed foreign issue branch
