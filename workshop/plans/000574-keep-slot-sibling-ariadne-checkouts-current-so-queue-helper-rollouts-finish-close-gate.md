---
gate: boundary-review
issue: 574
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 95463c2e-f4cd-4c59-98c8-01e947963491
            repository: github.com/xianxu/ariadne
            issue_id: "000574"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/8000f35af00287d4
            reviewed_head: a492fb052f9d058acef4f8eeabc41f62c8003482
            slot: ariadne:3
            requested_at: "2026-10-23T16:35:27.002991Z"
          phase: finished
          pid: 44319
          started_at: "2026-10-23T16:35:27.180476Z"
          finished_at: "2026-10-23T16:37:47.399515Z"
          session_id: 01a1699e-0be8-7082-ba69-ec46c70c3b3e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: a492fb052f9d058acef4f8eeabc41f62c8003482
        tracker_head: dcb26ebe943dd431ed85b44eb08269b1baa66bae
        branch_point: db77404e3257507f41c9d6fc3d819a8723bb36fa
        since: "2026-10-23T09:02:18-07:00"
        until: "2026-10-23T09:33:55-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T09:02:18-07:00"
              end: "2026-10-23T09:33:55-07:00"
              began: "2026-10-23T09:02:18-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 60dbe56852f8dacaed075324a46ed00c6a2357ac6e59f7d5d8e62a0933e17c0e
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 132099312
              windows:
                - start: "2026-10-23T09:02:18-07:00"
                  end: "2026-10-23T09:33:55-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 0.1062
      "n": 1
      timestamp: "2026-10-23T09:37:48-07:00"
      dispatched: "2026-10-23T09:35:27-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Important
          title: README update appears missing for weave refresh --suite-helper
          detail: cmd/weave/refresh.go:95 introduces a user-facing flag, but README.md is unchanged in the pinned range. Add the flag, automatic suite-setup boundary, eligibility limits, and skip/recovery behavior to README.md's existing refresh documentation around line 405.
          family: user-surface-documentation
          round: 1
        - id: BR-2
          severity: Important
          title: Supplier provenance tests do not deliver the controls marked complete in the plan
          detail: workshop/plans/000574-sibling-helper-refresh-plan.md:82 marks multiple revisions, fast-forward after ticket launch, missing worktree, symlink/identity mismatch, and Git failure controls complete. cmd/sdlc/queue_suite_supplier_test.go exercises one native dirty checkout plus injected constant observations; it does not exercise those production resolver/read paths. Add private fixtures covering distinct real supplier revisions, a supplier advancing while its legacy ticket remains unchanged, and unavailable results for missing or redirected owners and failed Git observations.
          family: planned-behavior-regression-coverage
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#574 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T09:37:48-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-surface-documentation` README update appears missing for weave refresh --suite-helper
  cmd/weave/refresh.go:95 introduces a user-facing flag, but README.md is unchanged in the pinned range. Add the flag, automatic suite-setup boundary, eligibility limits, and skip/recovery behavior to README.md's existing refresh documentation around line 405.
- **BR-2** [Important] `planned-behavior-regression-coverage` Supplier provenance tests do not deliver the controls marked complete in the plan
  workshop/plans/000574-sibling-helper-refresh-plan.md:82 marks multiple revisions, fast-forward after ticket launch, missing worktree, symlink/identity mismatch, and Git failure controls complete. cmd/sdlc/queue_suite_supplier_test.go exercises one native dirty checkout plus injected constant observations; it does not exercise those production resolver/read paths. Add private fixtures covering distinct real supplier revisions, a supplier advancing while its legacy ticket remains unchanged, and unavailable results for missing or redirected owners and failed Git observations.

## Open findings

- **BR-1** [Important] `user-surface-documentation` README update appears missing for weave refresh --suite-helper
- **BR-2** [Important] `planned-behavior-regression-coverage` Supplier provenance tests do not deliver the controls marked complete in the plan
