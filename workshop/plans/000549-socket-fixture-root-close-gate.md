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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7610a9e7-1229-4b89-bace-0b643f674870
            repository: github.com/xianxu/ariadne
            issue_id: "000549"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/0351ca0f625590d6
            reviewed_head: c677d1f53ba04b9fcbed79490981f9a0e0d7abfa
            slot: ariadne:5
            requested_at: "2026-10-22T21:50:12.7083Z"
          phase: finished
          pid: 61776
          started_at: "2026-10-22T21:50:12.986045Z"
          finished_at: "2026-10-22T21:51:47.671904Z"
          session_id: 01a16597-dc12-72e0-8e90-75d1113a8afe
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: c677d1f53ba04b9fcbed79490981f9a0e0d7abfa
        tracker_head: 8c7c9357429f47b1eedf062515acbda4593bd08d
        branch_point: cb3952669997cf4835f5ce139a4e0b1b72840f6f
        since: "2026-10-22T13:51:39-07:00"
        until: "2026-10-22T14:49:01-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T13:51:39-07:00"
              end: "2026-10-22T14:49:01-07:00"
              began: "2026-10-22T13:51:39-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 3423cc64b49f80286a94d286d45452322038091e04a0352b1e33ac5994cf2356
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 54632377
              windows:
                - start: "2026-10-22T13:51:39-07:00"
                  end: "2026-10-22T14:49:01-07:00"
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
                  end: "2026-10-22T14:49:01-07:00"
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
                  end: "2026-10-22T14:49:01-07:00"
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
                  end: "2026-10-22T14:49:01-07:00"
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
                  end: "2026-10-22T14:49:01-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 15.376216666666666
        cuts:
            - dispatched: "2026-10-22T14:42:36-07:00"
              finalized: "2026-10-22T14:45:27-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-22T14:51:48-07:00"
      dispatched: "2026-10-22T14:50:12-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: socket_root.go:173 and :180 normalize assignment targets and tuple-return expressions. Check/CheckSources regressions cover both; separate overlays removing each normalization made the corresponding regression fail with zero findings instead of one.
          round: 2
        - id: BR-2
          disposition: addressed
          note: socket_root.go:115 recognizes explicitly declared local receiver types. Check/CheckSources tests cover TempDir, Setenv, TB, parentheses, and unrelated-type controls. Removing ValueSpec recognition failed the positive regressions. Atlas and plan revisions document unresolved inferred and named-type aliases.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: c677d1f53ba04b9fcbed79490981f9a0e0d7abfa
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d54fdd75-b51a-4227-b865-456663167e3a
            repository: github.com/xianxu/ariadne
            issue_id: "000549"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/402dfbbabf77178e
            reviewed_head: d32c7b0811908755c8cdbb0ce2d7a3b8b17d2cda
            slot: ariadne:5
            requested_at: "2026-10-23T03:29:51.751302Z"
          phase: finished
          pid: 24544
          started_at: "2026-10-23T03:29:51.843393Z"
          finished_at: "2026-10-23T03:30:49.063712Z"
          session_id: 01a166ce-cff7-7c73-b2e4-d9849ce84561
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: d32c7b0811908755c8cdbb0ce2d7a3b8b17d2cda
        tracker_head: daf8af593e898eca3ea337d36f8872f934ecdf0b
        branch_point: 36544319c0c14a714e70f729ff73d21e8dab0df6
        since: "2026-10-22T13:51:39-07:00"
        until: "2026-10-22T20:29:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T13:51:39-07:00"
              end: "2026-10-22T20:29:16-07:00"
              began: "2026-10-22T13:51:39-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 2b17dad810d95de5aab5f5f7a88d1aa3b8761824bdaf90f85b8ec00bd5263beb
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 59589986
              windows:
                - start: "2026-10-22T13:51:39-07:00"
                  end: "2026-10-22T20:29:16-07:00"
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
                  end: "2026-10-22T20:29:16-07:00"
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
                  end: "2026-10-22T20:29:16-07:00"
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
                  end: "2026-10-22T20:29:16-07:00"
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
                  end: "2026-10-22T20:29:16-07:00"
        queue_waits:
            - attempt_id: 15fe86ae8c6cddc0f314b7ab9dea1737
              slot: ariadne:5
              record_digest: 4b0006fb773ec6ff3b2846d2d75dc9210a8921a5b44a44d25485ad196bcf9f17
              start: "2026-10-22T21:49:06.906083Z"
              end: "2026-10-22T23:37:57.776303Z"
              queue_seconds: 6530.87022
              raw_seconds: 6530.87022
              exact: true
            - attempt_id: 785aff8da0351ad139dcbf4df040e59f
              slot: ariadne:5
              record_digest: dd578d596267515eb632d0eaccaefdc9f61a435cd1ddd0f36fa2a4b011fd5e19
              start: "2026-10-23T02:52:40.710304Z"
              end: "2026-10-23T03:14:27.736706Z"
              queue_seconds: 1307.026402
              raw_seconds: 1307.026402
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 200.73011666666667
        cuts:
            - dispatched: "2026-10-22T14:42:36-07:00"
              finalized: "2026-10-22T14:45:27-07:00"
              finalizing: false
            - dispatched: "2026-10-22T14:50:12-07:00"
              finalized: "2026-10-22T14:51:48-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-22T20:30:49-07:00"
      dispatched: "2026-10-22T20:29:51-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Parenthesis normalization and Check/CheckSources regressions remain unchanged in this window; those regressions pass at the pinned head.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Explicit local testing receiver recognition and positive/unrelated-type controls remain unchanged in this window; those regressions pass at the pinned head.
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: d32c7b0811908755c8cdbb0ce2d7a3b8b17d2cda
      blocked: false
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

## Round 2 — 2026-10-22T14:51:48-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — socket_root.go:173 and :180 normalize assignment targets and tuple-return expressions. Check/CheckSources regressions cover both; separate overlays removing each normalization made the corresponding regression fail with zero findings instead of one.
- BR-2 — addressed — socket_root.go:115 recognizes explicitly declared local receiver types. Check/CheckSources tests cover TempDir, Setenv, TB, parentheses, and unrelated-type controls. Removing ValueSpec recognition failed the positive regressions. Atlas and plan revisions document unresolved inferred and named-type aliases.

## Round 3 — 2026-10-22T20:30:49-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Parenthesis normalization and Check/CheckSources regressions remain unchanged in this window; those regressions pass at the pinned head.
- BR-2 — addressed — Explicit local testing receiver recognition and positive/unrelated-type controls remain unchanged in this window; those regressions pass at the pinned head.

## Open findings

(none — every finding has been disposed)
