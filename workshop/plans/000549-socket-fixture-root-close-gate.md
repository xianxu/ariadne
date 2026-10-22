---
gate: boundary-review
issue: 549
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e15553c4-5a33-4a7b-b632-ebfd71c54de2
            repository: github.com/xianxu/ariadne
            issue_id: "000549"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/e22ccfaaec5d768e
            reviewed_head: 41ca34e209e77a44661d746a1dbbb21ad78d29d7
            slot: ariadne:5
            requested_at: "2026-10-22T21:42:36.101603Z"
          phase: finished
          pid: 38840
          started_at: "2026-10-22T21:42:36.411579Z"
          finished_at: "2026-10-22T21:44:54.952988Z"
          session_id: 01a16590-e4c5-7c11-8123-92c0072c1ea5
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 41ca34e209e77a44661d746a1dbbb21ad78d29d7
        tracker_head: cf26cf9227cec304e182763edc62718f1bb478af
        branch_point: cb3952669997cf4835f5ce139a4e0b1b72840f6f
        since: "2026-10-22T13:51:39-07:00"
        until: "2026-10-22T14:10:04-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T13:51:39-07:00"
              end: "2026-10-22T14:10:04-07:00"
              began: "2026-10-22T13:51:39-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: decb59fba081bf65b25456d661f95191d083f80739522e948a3934616e091005
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 54235669
              windows:
                - start: "2026-10-22T13:51:39-07:00"
                  end: "2026-10-22T14:10:04-07:00"
            - harness: codex
              id: 01a16569-2ac8-7262-8166-f38064a3678e
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 00091c72291162edcfd3f8b20e4c1f8f73f060f182df88d3001027b34b0ca5b5
              started: "2026-10-22T20:59:13.935Z"
              spawned: false
              bytes: 591831
              windows:
                - start: "2026-10-22T20:59:13.935Z"
                  end: "2026-10-22T14:10:04-07:00"
            - harness: codex
              id: 01a1656b-d8d4-7f62-b4e0-d964ce2b3bff
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 6ad1e1d54f7f5ab4af350c7bfe5467c3149e091b45cce574b281f0caaa4d28d3
              started: "2026-10-22T21:02:09.88Z"
              spawned: false
              bytes: 402780
              windows:
                - start: "2026-10-22T21:02:09.88Z"
                  end: "2026-10-22T14:10:04-07:00"
            - harness: codex
              id: 01a1656d-6400-7a00-8f94-942f722b558c
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 579b9383a6ccf4aa56a11916801de954ae79a68f68d769bccc31e7bf1384f10c
              started: "2026-10-22T21:03:51.064Z"
              spawned: false
              bytes: 427949
              windows:
                - start: "2026-10-22T21:03:51.064Z"
                  end: "2026-10-22T14:10:04-07:00"
            - harness: codex
              id: 01a1656e-33d9-7091-8b81-5f329965ee2e
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 3c51dfa79ef1d3fc6341061c460c69a2ef802c7dc2350ad3c29b6c78f5bc1329
              started: "2026-10-22T21:04:44.002Z"
              spawned: false
              bytes: 387138
              windows:
                - start: "2026-10-22T21:04:44.002Z"
                  end: "2026-10-22T14:10:04-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.09348333333333333
      "n": 1
      timestamp: "2026-10-22T14:45:27-07:00"
      dispatched: "2026-10-22T14:42:36-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Parentheses bypass supported assignment and multi-result propagation
          detail: pkg/golint/socket_root.go:143 requires an identifier LHS, and line 150 requires an unparenthesized call RHS. Consequently, both `var root string; (root)=t.TempDir(); net.Listen("unix",root)` and `root,err:=(os.MkdirTemp("","s-")); _=err; net.Listen("unix",root)` produce no finding. Independently verified that both compile. Normalize transparent parentheses before binding identifiers and recognizing first-result calls, and add regressions through Check and CheckSources. This violates the plan's assignment/parentheses contract (ARCH-PURPOSE).
          family: transparent-expression-provenance
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Explicitly typed local testing receivers are silently excluded
          detail: pkg/golint/socket_root.go:84 accepts receiver declarations only as ast.Field. Thus `var tt *testing.T=t; net.Listen("unix",tt.TempDir())` produces no finding, despite the receiver type being directly visible and the plan promising visible testing T/B/F/TB receivers. Recognize explicit local variable types for TempDir and Setenv, with positive tests and unrelated-type controls; document any remaining receiver-alias limitation precisely (ARCH-PURPOSE).
          family: visible-receiver-recognition
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#549 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T14:45:27-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `transparent-expression-provenance` (kind: defect) Parentheses bypass supported assignment and multi-result propagation
  pkg/golint/socket_root.go:143 requires an identifier LHS, and line 150 requires an unparenthesized call RHS. Consequently, both `var root string; (root)=t.TempDir(); net.Listen("unix",root)` and `root,err:=(os.MkdirTemp("","s-")); _=err; net.Listen("unix",root)` produce no finding. Independently verified that both compile. Normalize transparent parentheses before binding identifiers and recognizing first-result calls, and add regressions through Check and CheckSources. This violates the plan's assignment/parentheses contract (ARCH-PURPOSE).
- **BR-2** [Important] `visible-receiver-recognition` (kind: defect) Explicitly typed local testing receivers are silently excluded
  pkg/golint/socket_root.go:84 accepts receiver declarations only as ast.Field. Thus `var tt *testing.T=t; net.Listen("unix",tt.TempDir())` produces no finding, despite the receiver type being directly visible and the plan promising visible testing T/B/F/TB receivers. Recognize explicit local variable types for TempDir and Setenv, with positive tests and unrelated-type controls; document any remaining receiver-alias limitation precisely (ARCH-PURPOSE).

## Open findings

- **BR-1** [Critical] `transparent-expression-provenance` (kind: defect) Parentheses bypass supported assignment and multi-result propagation
- **BR-2** [Important] `visible-receiver-recognition` (kind: defect) Explicitly typed local testing receivers are silently excluded
