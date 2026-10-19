---
gate: boundary-review
issue: 469
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d6bab811-356f-4993-bbdb-fef48e6915e9
            repository: github.com/xianxu/ariadne
            issue_id: "000469"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/69b25147586cac0d
            reviewed_head: 7879351a725d5ac52d9ea00e2ced1c362ba0dde7
            slot: ariadne:2
            requested_at: "2026-10-19T18:43:08.990555Z"
          phase: finished
          pid: 60831
          started_at: "2026-10-19T18:43:09.147762Z"
          finished_at: "2026-10-19T18:44:14.90636Z"
          session_id: f2a77207-4be8-42be-b365-bddac8c8c3b9
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7879351a725d5ac52d9ea00e2ced1c362ba0dde7
        tracker_head: 5ab2120b88b7b3d614471399ee21ee2d07acb9d0
        branch_point: cabcee34f5008627158452f85ee7b811dc3a0bf2
        since: "2026-10-19T11:26:36-07:00"
        until: "2026-10-19T11:29:14-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T11:26:36-07:00"
              end: "2026-10-19T11:29:14-07:00"
              began: "2026-10-19T11:26:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: b9b76aea819faeeac70f909118790dcd231c0e68ed43f45d799edaff08867879
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 330505450
              windows:
                - start: "2026-10-19T11:26:36-07:00"
                  end: "2026-10-19T11:29:14-07:00"
      "n": 1
      timestamp: "2026-10-19T11:44:15-07:00"
      dispatched: "2026-10-19T11:43:08-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: The atlas Go lint catalog list leaves out the new test-reads-workshop rule (#469)
          detail: 'atlas/workflow/sdlc-binary.md:2027-2030 names every catalog rule with its incident. It lists all 8 earlier rules but not test-reads-workshop, which pkg/golint/rules.go:64 adds. Add "`test-reads-workshop` #469" to that list. No other Catalog-restating list exists in this window.'
          family: atlas-restatement-drift
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: 'The atlas still links the archived #449 plan at workshop/plans'
          detail: 'atlas/workflow/sdlc-binary.md:462 says "Plan: `workshop/plans/000449-landing-train-plan.md`", but the file now lives at workshop/history/plans/. Other atlas lines with the same pattern, all older than this change: harness-integration.md:104, introspect.md:325, issue-tracker-migration.md:98.'
          family: doc-links-archived-workshop-path
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: The atlas says the test keeps the Phases list in step, but it only checks events and effects
          detail: atlas/workflow/sdlc-binary.md:457 (Phases) falls under the claim "TestPlanNamesEveryTrainEventAndEffect keeps this list in step". The test matches only backtick-prefixed event and effect type names. Either narrow the claim or extend the test to phases.
          family: doc-claim-exceeds-enforcement
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: test-reads-workshop misses paths split into Join segments
          detail: filepath.Join("..", "..", "workshop", "plans", "x.md") passes the rule, because only a single literal that climbs into workshop/plans or workshop/issues is matched. This is acceptable for the Done-when; consider matching adjacent "workshop","plans"/"issues" arguments that follow ".." in the same Join call.
          family: lint-rule-literal-only
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2ed2f97d-e321-4c11-8d18-425e27c612fb
            repository: github.com/xianxu/ariadne
            issue_id: "000469"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/e9b052e10f1f941a
            reviewed_head: 7b2bff5878d86def1d2aa555929b5136b90e1ceb
            slot: ariadne:2
            requested_at: "2026-10-19T18:45:01.198229Z"
          phase: finished
          pid: 83073
          started_at: "2026-10-19T18:45:01.31092Z"
          finished_at: "2026-10-19T18:45:30.061751Z"
          session_id: f7ca9c3a-9cb9-4173-9f7d-b91f549db18e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 7b2bff5878d86def1d2aa555929b5136b90e1ceb
        tracker_head: 5ab2120b88b7b3d614471399ee21ee2d07acb9d0
        branch_point: cabcee34f5008627158452f85ee7b811dc3a0bf2
        since: "2026-10-19T11:26:36-07:00"
        until: "2026-10-19T11:44:42-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T11:26:36-07:00"
              end: "2026-10-19T11:44:42-07:00"
              began: "2026-10-19T11:26:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: d237cd39764e21fcbb2515cfa0c4f1ad1fd741e9e6dfbbc230b2da873a83d62a
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 331171418
              windows:
                - start: "2026-10-19T11:26:36-07:00"
                  end: "2026-10-19T11:44:42-07:00"
        queue_waits:
            - attempt_id: fe5fa2d62500c79bcedef05f0b82516e
              slot: ariadne:2
              record_digest: d9143902a070d8f22cd30019237a1b5f75858e7b276da98c3979fcccc93cb169
              start: "2026-10-19T18:29:25.729801Z"
              end: "2026-10-19T18:29:43.012858Z"
              queue_seconds: 17.283057
              raw_seconds: 17.283057
              exact: true
        cuts:
            - dispatched: "2026-10-19T11:43:08-07:00"
              finalized: "2026-10-19T11:44:15-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-19T11:45:30-07:00"
      dispatched: "2026-10-19T11:45:01-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'sdlc-binary.md:2030 now names test-reads-workshop with incident #469. TestAtlasNamesEveryRule pins it: removing the backticks turned the test red, verified at the pinned head and then restored.'
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: 'sdlc-binary.md:462 still says Plan: workshop/plans/000449-landing-train-plan.md, but the file is gone from workshop/plans. Change it to workshop/history/plans/.'
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: sdlc-binary.md:457 still claims the test keeps the whole list in step, including Phases, but the test checks only backticked Event and Effect names.
          round: 2
        - id: BR-4
          disposition: not-addressed
          note: 'Still literal-only: filepath.Join("..","..","workshop","plans",...) passes the rule. This stays acceptable under the Done-when.'
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: 7b2bff5878d86def1d2aa555929b5136b90e1ceb
      blocked: false
---

# Gate ledger — ariadne#469 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T11:44:15-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `atlas-restatement-drift` (kind: defect) The atlas Go lint catalog list leaves out the new test-reads-workshop rule (#469)
  atlas/workflow/sdlc-binary.md:2027-2030 names every catalog rule with its incident. It lists all 8 earlier rules but not test-reads-workshop, which pkg/golint/rules.go:64 adds. Add "`test-reads-workshop` #469" to that list. No other Catalog-restating list exists in this window.
- **BR-2** [Minor] `doc-links-archived-workshop-path` (kind: defect) The atlas still links the archived #449 plan at workshop/plans
  atlas/workflow/sdlc-binary.md:462 says "Plan: `workshop/plans/000449-landing-train-plan.md`", but the file now lives at workshop/history/plans/. Other atlas lines with the same pattern, all older than this change: harness-integration.md:104, introspect.md:325, issue-tracker-migration.md:98.
- **BR-3** [Minor] `doc-claim-exceeds-enforcement` (kind: defect) The atlas says the test keeps the Phases list in step, but it only checks events and effects
  atlas/workflow/sdlc-binary.md:457 (Phases) falls under the claim "TestPlanNamesEveryTrainEventAndEffect keeps this list in step". The test matches only backtick-prefixed event and effect type names. Either narrow the claim or extend the test to phases.
- **BR-4** [Minor] `lint-rule-literal-only` (kind: defect) test-reads-workshop misses paths split into Join segments
  filepath.Join("..", "..", "workshop", "plans", "x.md") passes the rule, because only a single literal that climbs into workshop/plans or workshop/issues is matched. This is acceptable for the Done-when; consider matching adjacent "workshop","plans"/"issues" arguments that follow ".." in the same Join call.

## Round 2 — 2026-10-19T11:45:30-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — sdlc-binary.md:2030 now names test-reads-workshop with incident #469. TestAtlasNamesEveryRule pins it: removing the backticks turned the test red, verified at the pinned head and then restored.
- BR-2 — not-addressed — sdlc-binary.md:462 still says Plan: workshop/plans/000449-landing-train-plan.md, but the file is gone from workshop/plans. Change it to workshop/history/plans/.
- BR-3 — not-addressed — sdlc-binary.md:457 still claims the test keeps the whole list in step, including Phases, but the test checks only backticked Event and Effect names.
- BR-4 — not-addressed — Still literal-only: filepath.Join("..","..","workshop","plans",...) passes the rule. This stays acceptable under the Done-when.

## Open findings

- **BR-2** [Minor] `doc-links-archived-workshop-path` (kind: defect) The atlas still links the archived #449 plan at workshop/plans
- **BR-3** [Minor] `doc-claim-exceeds-enforcement` (kind: defect) The atlas says the test keeps the Phases list in step, but it only checks events and effects
- **BR-4** [Minor] `lint-rule-literal-only` (kind: defect) test-reads-workshop misses paths split into Join segments
