---
gate: boundary-review
issue: 387
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e6b4e583-9d58-43fa-be54-405be5c59207
            repository: github.com/xianxu/ariadne
            issue_id: "000387"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c440b4c4351bdb59
            reviewed_head: 83cf89adaae2526d9c5d35ae228dda063a84f75a
            slot: ariadne:4
            requested_at: "2026-10-15T22:00:38.937282Z"
          phase: finished
          pid: 6587
          started_at: "2026-10-15T22:00:39.429975Z"
          finished_at: "2026-10-15T22:02:59.007037Z"
          session_id: 773f0f51-7b19-4c99-971b-18a7a7ade280
          outcome: returned
        - version: 1
          intent:
            disposable_root: true
            id: 0488afdc-9e99-478a-b261-e51655f9586d
            repository: github.com/xianxu/ariadne
            issue_id: "000387"
            recipe: small-diff-review
            round: 1
            attempt: 2
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c440b4c4351bdb59
            reviewed_head: 83cf89adaae2526d9c5d35ae228dda063a84f75a
            slot: ariadne:4
            requested_at: "2026-10-15T22:03:00.147758Z"
          phase: finished
          pid: 38483
          started_at: "2026-10-15T22:03:00.280015Z"
          finished_at: "2026-10-15T22:03:35.779504Z"
          session_id: daea5e58-2cc2-495e-993f-8c6409485dba
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 83cf89adaae2526d9c5d35ae228dda063a84f75a
        tracker_head: 4b45626545c12585b0568fdac7d6af6a83f0ff84
        branch_point: 84d3981d4d486611e70c77a20848d5270c24a13c
        since: "2026-10-15T11:01:43-07:00"
        until: "2026-10-15T14:58:29-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T11:01:43-07:00"
              end: "2026-10-15T14:58:29-07:00"
              began: "2026-10-15T11:01:43-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 24c8d79e-33f9-4ab1-a6b5-8cb02e00580a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: aafd3fc8ef75cf94079399d0a6812c0302b814e7eeb6e28dcee8442b1533411a
              started: "2026-10-12T14:18:17.059Z"
              spawned: false
              bytes: 13014436
              windows:
                - start: "2026-10-15T11:01:43-07:00"
                  end: "2026-10-15T14:58:29-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.029516666666666667
      "n": 1
      timestamp: "2026-10-15T15:03:36-07:00"
      dispatched: "2026-10-15T15:00:38-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: Legacy-parse test references the Legacy constant rather than the literal "sdlc"
          detail: agentid_test.go:24 calls Parse(Legacy). If someone edits the constant, the test changes with it and passes even though existing trailers in history no longer parse. Asserting Parse("sdlc") pins the history format. This is the only instance in the window.
          family: legacy-compat-pinned-by-literal
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 83cf89adaae2526d9c5d35ae228dda063a84f75a
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: a4d74aea-ed98-42e2-b118-2bccd9a96b21
            repository: github.com/xianxu/ariadne
            issue_id: "000387"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/9c79a10366ccb264
            reviewed_head: 870f183161f3b9a02677cb355caeaa3afe7f99a2
            slot: ariadne:4
            requested_at: "2026-10-15T22:05:02.964772Z"
          phase: finished
          pid: 50768
          started_at: "2026-10-15T22:05:03.131021Z"
          finished_at: "2026-10-15T22:05:28.024626Z"
          session_id: f23d282e-3d09-44d2-a1d9-e44a184e0c6a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 870f183161f3b9a02677cb355caeaa3afe7f99a2
        tracker_head: a66e844c4ee49aa0c82ae7780925357f98603df1
        branch_point: e055f2a3f7102135beb6bb966fe87f9da33fdd83
        since: "2026-10-15T11:01:43-07:00"
        until: "2026-10-15T15:04:01-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T11:01:43-07:00"
              end: "2026-10-15T15:04:01-07:00"
              began: "2026-10-15T11:01:43-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 24c8d79e-33f9-4ab1-a6b5-8cb02e00580a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 2a392f5f402866121dc97df422f6a2ce76d29eb726cbd9b33723b1a01778b919
              started: "2026-10-12T14:18:17.059Z"
              spawned: false
              bytes: 13100235
              windows:
                - start: "2026-10-15T11:01:43-07:00"
                  end: "2026-10-15T15:04:01-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.029516666666666667
        cuts:
            - dispatched: "2026-10-15T15:00:38-07:00"
              finalized: "2026-10-15T15:03:36-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-15T15:05:34-07:00"
      dispatched: "2026-10-15T15:05:02-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: agentid_test.go:25 now loops over the literals "human" and "sdlc". Changing Legacy to "sdlc2" turns the test red at line 27, and restoring it turns it green.
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: 870f183161f3b9a02677cb355caeaa3afe7f99a2
      blocked: false
---

# Gate ledger — ariadne#387 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T15:03:36-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `legacy-compat-pinned-by-literal` (kind: defect) Legacy-parse test references the Legacy constant rather than the literal "sdlc"
  agentid_test.go:24 calls Parse(Legacy). If someone edits the constant, the test changes with it and passes even though existing trailers in history no longer parse. Asserting Parse("sdlc") pins the history format. This is the only instance in the window.

## Round 2 — 2026-10-15T15:05:34-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — agentid_test.go:25 now loops over the literals "human" and "sdlc". Changing Legacy to "sdlc2" turns the test red at line 27, and restoring it turns it green.

## Open findings

(none — every finding has been disposed)
