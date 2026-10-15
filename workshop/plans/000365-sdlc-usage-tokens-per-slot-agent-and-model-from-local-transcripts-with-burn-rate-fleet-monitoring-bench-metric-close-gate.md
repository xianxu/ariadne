---
gate: boundary-review
issue: 365
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 2723e1f5892053ae932f5c86bf14818454ae62ad
        tracker_head: 38322ec576dd1b9bd4c58fb2c847ffbfafd16239
        branch_point: 18915f6c17d9b4aeb5e618d3f736705ec355d2b3
        since: "2026-10-14T07:39:58-07:00"
        until: "2026-10-14T21:06:52-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T07:39:58-07:00"
              end: "2026-10-14T21:06:52-07:00"
              began: "2026-10-14T07:39:58-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a138fd-1ac0-7a83-826e-60623c226386
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 4f4e740df2b643518d372c881bb8c4e2d6b22b0b02a1f7a0343fd0376d2c439f
              started: "2026-10-14T05:58:30.184Z"
              spawned: false
              bytes: 26756698
              windows:
                - start: "2026-10-14T07:39:58-07:00"
                  end: "2026-10-14T21:06:52-07:00"
            - harness: codex
              id: 01a13b1c-bf44-71e2-81bd-5d24934b826e
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: e92e937bd36804a2b768db3419fb4733c214f05436d1f73f506d20e37182ed82
              started: "2026-10-14T15:51:42.429Z"
              spawned: false
              bytes: 747338
              windows:
                - start: "2026-10-14T15:51:42.429Z"
                  end: "2026-10-14T21:06:52-07:00"
            - harness: codex
              id: 01a13b23-a09c-7d03-9c26-d94c5e01b8a5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 30703dd1110e0f2b3d1f227c181abee39e19bca5930e07c83f34010170f69d20
              started: "2026-10-14T15:59:13.3Z"
              spawned: false
              bytes: 442525
              windows:
                - start: "2026-10-14T15:59:13.3Z"
                  end: "2026-10-14T21:06:52-07:00"
            - harness: codex
              id: 01a13b26-2c15-7842-8c20-78ad9adc63ca
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: fb19b09075ef1b879ef0d82858e34e66e2e8954dc02673929d4c561bb98c32a6
              started: "2026-10-14T16:02:00.404Z"
              spawned: false
              bytes: 439211
              windows:
                - start: "2026-10-14T16:02:00.404Z"
                  end: "2026-10-14T21:06:52-07:00"
            - harness: codex
              id: 01a13b26-c5e7-7711-8e4a-fc316e11fce7
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: deb59b626263063d9512ef2be32e8b120cd0d3e4b275cd73aab82b52717d5268
              started: "2026-10-14T16:02:39.61Z"
              spawned: false
              bytes: 411362
              windows:
                - start: "2026-10-14T16:02:39.61Z"
                  end: "2026-10-14T21:06:52-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 75.24986666666666
      "n": 1
      timestamp: "2026-10-14T21:11:21-07:00"
      dispatched: "2026-10-14T21:07:30-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Modern cumulative evidence hides later uncovered Codex spend
          detail: cmd/sdlc/internal/usage/codex.go:218 uses the legacy cumulative baseline only when no modern thread_token_usage exists. A production-decoder probe containing a modern response and thread total of 100 input/20 output, followed by a legacy cumulative snapshot of 500/100, returned only 120 tokens with spend_complete=true and no uncovered_segments diagnostic. This can authorize calibration using an incomplete denominator. Reconcile all applicable cumulative evidence without adding it to response totals, and add this mixed-format regression. ARCH-PURPOSE, ARCH-SECURE.
          family: cumulative-coverage-reconciliation
          round: 1
        - id: BR-2
          severity: Critical
          title: Invalid reviewer bindings disappear without incomplete coverage
          detail: cmd/sdlc/usage_reviewers.go:95 ignores ledger parse failures, while :136 silently discards records failing reviewusage.Validate. A production-loader probe with a version-99 reviewer record returned empty coverage, attribution_complete=true, and worker attribution for the affected owned session. Validate retained records at ingestion and report failures; usable identity evidence must not silently become ordinary worker evidence. Sweep ledger and comment readers, preferably sharing the duplicated decoding in reviewer_accounting.go:136. ARCH-SECURE, ARCH-DRY.
          family: invalid-accounting-evidence-must-remain-visible
          round: 1
        - id: BR-3
          severity: Critical
          title: Core concepts tables contradict the delivered paths and statuses
          detail: workshop/plans/000365-usage-plan.md:30 lists provenance.go as modified, but it is unchanged in the pinned range. Line 31 names nonexistent internal/gatestate/reviewer.go, and line 44 names nonexistent usage_render.go. Binding validation actually lives in internal/reviewusage/model.go and rendering in usage.go. Under the requested Core concepts gate, these contradictions are blocking. Append a Revisions entry and reconcile the tables with the delivered implementation and PURE/INTEGRATION ownership. ARCH-PURE.
          family: core-concept-table-must-match-code
          round: 1
        - id: BR-4
          severity: Important
          title: The promised gate-dispatch-to-usage regression is missing
          detail: workshop/plans/000365-usage-plan.md:139 marks production boundary/plan/estimate dispatch-through-usage coverage complete. The judge tests supply UsageIntent and UsageRecorder directly; root accounting tests exercise cleanup and serialization, and usage_attribution_test.go:159 fabricates bindings and worker windows. None verifies the production gate attachment and persistence through usage loading, including interruption without semantic gate advancement. Add the promised integration regressions and make omission of production accounting wiring fail them. ARCH-PURPOSE.
          family: production-path-regression-evidence
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#365 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T21:11:21-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `cumulative-coverage-reconciliation` Modern cumulative evidence hides later uncovered Codex spend
  cmd/sdlc/internal/usage/codex.go:218 uses the legacy cumulative baseline only when no modern thread_token_usage exists. A production-decoder probe containing a modern response and thread total of 100 input/20 output, followed by a legacy cumulative snapshot of 500/100, returned only 120 tokens with spend_complete=true and no uncovered_segments diagnostic. This can authorize calibration using an incomplete denominator. Reconcile all applicable cumulative evidence without adding it to response totals, and add this mixed-format regression. ARCH-PURPOSE, ARCH-SECURE.
- **BR-2** [Critical] `invalid-accounting-evidence-must-remain-visible` Invalid reviewer bindings disappear without incomplete coverage
  cmd/sdlc/usage_reviewers.go:95 ignores ledger parse failures, while :136 silently discards records failing reviewusage.Validate. A production-loader probe with a version-99 reviewer record returned empty coverage, attribution_complete=true, and worker attribution for the affected owned session. Validate retained records at ingestion and report failures; usable identity evidence must not silently become ordinary worker evidence. Sweep ledger and comment readers, preferably sharing the duplicated decoding in reviewer_accounting.go:136. ARCH-SECURE, ARCH-DRY.
- **BR-3** [Critical] `core-concept-table-must-match-code` Core concepts tables contradict the delivered paths and statuses
  workshop/plans/000365-usage-plan.md:30 lists provenance.go as modified, but it is unchanged in the pinned range. Line 31 names nonexistent internal/gatestate/reviewer.go, and line 44 names nonexistent usage_render.go. Binding validation actually lives in internal/reviewusage/model.go and rendering in usage.go. Under the requested Core concepts gate, these contradictions are blocking. Append a Revisions entry and reconcile the tables with the delivered implementation and PURE/INTEGRATION ownership. ARCH-PURE.
- **BR-4** [Important] `production-path-regression-evidence` The promised gate-dispatch-to-usage regression is missing
  workshop/plans/000365-usage-plan.md:139 marks production boundary/plan/estimate dispatch-through-usage coverage complete. The judge tests supply UsageIntent and UsageRecorder directly; root accounting tests exercise cleanup and serialization, and usage_attribution_test.go:159 fabricates bindings and worker windows. None verifies the production gate attachment and persistence through usage loading, including interruption without semantic gate advancement. Add the promised integration regressions and make omission of production accounting wiring fail them. ARCH-PURPOSE.

## Open findings

- **BR-1** [Critical] `cumulative-coverage-reconciliation` Modern cumulative evidence hides later uncovered Codex spend
- **BR-2** [Critical] `invalid-accounting-evidence-must-remain-visible` Invalid reviewer bindings disappear without incomplete coverage
- **BR-3** [Critical] `core-concept-table-must-match-code` Core concepts tables contradict the delivered paths and statuses
- **BR-4** [Important] `production-path-regression-evidence` The promised gate-dispatch-to-usage regression is missing
