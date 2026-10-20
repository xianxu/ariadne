---
gate: boundary-review
issue: 489
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 107f4f76-91ba-4082-b761-bd1381005a95
            repository: github.com/xianxu/ariadne
            issue_id: "000489"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/8b9699b35ff09a6e
            reviewed_head: 497ca73e8dd4c7923e0c4a071656065480f858a6
            slot: ariadne:2
            requested_at: "2026-10-20T12:33:25.42607Z"
          phase: finished
          pid: 96433
          started_at: "2026-10-20T12:33:25.582099Z"
          finished_at: "2026-10-20T12:39:48.701623Z"
          session_id: 0cba0e04-56ac-47e9-8e74-aaa6156b5f8f
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 497ca73e8dd4c7923e0c4a071656065480f858a6
        tracker_head: 590710092c264e352c84aa2d4ba25e6c7c9929c1
        branch_point: bf7eaa4d17eaf48b7b36b71982f65bdcf5f9d114
        since: "2026-10-20T02:02:22-07:00"
        until: "2026-10-20T05:32:43-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T02:02:22-07:00"
              end: "2026-10-20T02:02:31-07:00"
              began: "2026-10-20T02:02:22-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            - start: "2026-10-20T05:21:51-07:00"
              end: "2026-10-20T05:32:43-07:00"
              began: "2026-10-20T05:21:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f8a7ba0715ff72ba471ba8775a00af875944a29ceb9df5b2b8b3d995a236eecf
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 5972683
              windows:
                - start: "2026-10-20T02:02:22-07:00"
                  end: "2026-10-20T02:02:31-07:00"
                - start: "2026-10-20T05:21:51-07:00"
                  end: "2026-10-20T05:32:43-07:00"
        protected:
            - issue: "473"
              slot: ariadne:2
              start: "2026-10-20T02:02:22-07:00"
              end: "2026-10-20T02:07:22-07:00"
      "n": 1
      timestamp: "2026-10-20T05:39:49-07:00"
      dispatched: "2026-10-20T05:33:25-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: reviewusage identity.go duplicate-key Scan has no test that fails without it
          detail: Replacing jsonstrict.Scan(..., 16) at cmd/sdlc/internal/reviewusage/identity.go:59 with nil leaves the reviewusage suite green. The gap predates this change, but a session-event line with a duplicate key should be rejected by a test.
          family: strict-check-untested-caller
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: json-recursive-scanner misses method receivers, selector recursion and iterative stack walks
          detail: recursiveJSONScanners skips FuncDecls with a receiver, only treats a bare identifier call as recursion, and cannot see iterative walks (the Log notes the last). A future method-based scanner would get past the single-owner guard.
          family: lint-rule-shape-coverage
          round: 1
      recipe: milestone-review
      isolation: checkout
      reviewed: 497ca73e8dd4c7923e0c4a071656065480f858a6
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 662b2dd4-e2dc-4d3f-9eb3-54cc7fd99bab
            repository: github.com/xianxu/ariadne
            issue_id: "000489"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/99a02a4ee0898403
            reviewed_head: 07e0d9724226706fbe4751c3de68aef6eedbe852
            slot: ariadne:2
            requested_at: "2026-10-20T12:41:18.055295Z"
          phase: finished
          pid: 98317
          started_at: "2026-10-20T12:41:18.155523Z"
          finished_at: "2026-10-20T12:42:05.805988Z"
          session_id: 86212ef0-bdb5-45dd-b6c1-84d37d023039
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 07e0d9724226706fbe4751c3de68aef6eedbe852
        tracker_head: f0fca7fcc5563e7481425c8ca3ac74b6e165884a
        branch_point: bf7eaa4d17eaf48b7b36b71982f65bdcf5f9d114
        since: "2026-10-20T02:02:22-07:00"
        until: "2026-10-20T05:40:50-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T02:02:22-07:00"
              end: "2026-10-20T02:02:31-07:00"
              began: "2026-10-20T02:02:22-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            - start: "2026-10-20T05:21:51-07:00"
              end: "2026-10-20T05:40:50-07:00"
              began: "2026-10-20T05:21:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 0973f648-94f5-4a35-b868-c9b26ef649ac
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 034ccf96755533008a7a721bbce389b589c88a73497422e8c595eb037821e696
              started: "2026-10-20T03:20:54.668Z"
              spawned: false
              bytes: 6137532
              windows:
                - start: "2026-10-20T02:02:22-07:00"
                  end: "2026-10-20T02:02:31-07:00"
                - start: "2026-10-20T05:21:51-07:00"
                  end: "2026-10-20T05:40:50-07:00"
        protected:
            - issue: "473"
              slot: ariadne:2
              start: "2026-10-20T02:02:22-07:00"
              end: "2026-10-20T02:07:22-07:00"
        cuts:
            - dispatched: "2026-10-20T05:33:25-07:00"
              finalized: "2026-10-20T05:39:49-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-20T05:42:06-07:00"
      dispatched: "2026-10-20T05:41:18-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Disabling jsonstrict.Scan in identity.go turns TestIdentityStrictProtocol red (identity_test.go:40); both an exact and a different-case duplicate are covered.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: Selector recursion is fixed and tested, since removing rules.go:477 turns TestRules red. Methods are scanned now that the Recv filter is gone (rules.go:489), but putting that filter back keeps pkg/golint green. The harness wraps snippets in a receiver-less func f(), so a method needs a whole-file test case (as in TestRulesFollowImportNames).
          round: 2
      recipe: fix-review
      isolation: checkout
      reviewed: 07e0d9724226706fbe4751c3de68aef6eedbe852
      blocked: false
---

# Gate ledger — ariadne#489 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T05:39:49-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `strict-check-untested-caller` (kind: defect) reviewusage identity.go duplicate-key Scan has no test that fails without it
  Replacing jsonstrict.Scan(..., 16) at cmd/sdlc/internal/reviewusage/identity.go:59 with nil leaves the reviewusage suite green. The gap predates this change, but a session-event line with a duplicate key should be rejected by a test.
- **BR-2** [Minor] `lint-rule-shape-coverage` (kind: defect) json-recursive-scanner misses method receivers, selector recursion and iterative stack walks
  recursiveJSONScanners skips FuncDecls with a receiver, only treats a bare identifier call as recursion, and cannot see iterative walks (the Log notes the last). A future method-based scanner would get past the single-owner guard.

## Round 2 — 2026-10-20T05:42:06-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Disabling jsonstrict.Scan in identity.go turns TestIdentityStrictProtocol red (identity_test.go:40); both an exact and a different-case duplicate are covered.
- BR-2 — not-addressed — Selector recursion is fixed and tested, since removing rules.go:477 turns TestRules red. Methods are scanned now that the Recv filter is gone (rules.go:489), but putting that filter back keeps pkg/golint green. The harness wraps snippets in a receiver-less func f(), so a method needs a whole-file test case (as in TestRulesFollowImportNames).

## Open findings

- **BR-2** [Minor] `lint-rule-shape-coverage` (kind: defect) json-recursive-scanner misses method receivers, selector recursion and iterative stack walks
