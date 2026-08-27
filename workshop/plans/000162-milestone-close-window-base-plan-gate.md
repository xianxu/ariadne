---
gate: plan-quality
issue: 162
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-08-26T21:58:21-07:00"
      agent: codex
      findings:
        - id: PQ-1
          severity: Critical
          title: 'ARCH-PURPOSE: the preserved first-milestone resolver still violates the required branch-point window'
          detail: The revision preserves resolveReviewWindow and milestonewindow_test.go unchanged, but boundaryWindowBase routes a first milestone through branchStartByIssue at cmd/sdlc/milestoneclose.go:303, while the existing test at cmd/sdlc/milestonewindow_test.go:107 explicitly expects the first issue commit's parent. Refine the plan to use the feature-branch merge-base for first milestones and add a regression where the issue was filed on main before unrelated commits and branching.
          family: deliver-full-stated-purpose
          round: 1
        - id: PQ-2
          severity: Important
          title: Compress enumerated test cases into named risky-function strategies
          detail: Tasks 1 through 4 enumerate fixtures, cases, and expected assertions instead of giving one strategy line per risky function. Name the unit-tested functions, including RenderReviewWindow and resolveBoundaryReviewManifest, and state each adversarial input class plus its mechanical guard; leave the concrete case inventory to Go tests.
          family: test-strategy-contract
          round: 1
      blocked: true
---

# Gate ledger — ariadne#162 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-26T21:58:21-07:00 (codex) — BLOCKED

### Raised

- **PQ-1** [Critical] `deliver-full-stated-purpose` ARCH-PURPOSE: the preserved first-milestone resolver still violates the required branch-point window
  The revision preserves resolveReviewWindow and milestonewindow_test.go unchanged, but boundaryWindowBase routes a first milestone through branchStartByIssue at cmd/sdlc/milestoneclose.go:303, while the existing test at cmd/sdlc/milestonewindow_test.go:107 explicitly expects the first issue commit's parent. Refine the plan to use the feature-branch merge-base for first milestones and add a regression where the issue was filed on main before unrelated commits and branching.
- **PQ-2** [Important] `test-strategy-contract` Compress enumerated test cases into named risky-function strategies
  Tasks 1 through 4 enumerate fixtures, cases, and expected assertions instead of giving one strategy line per risky function. Name the unit-tested functions, including RenderReviewWindow and resolveBoundaryReviewManifest, and state each adversarial input class plus its mechanical guard; leave the concrete case inventory to Go tests.

## Open findings

- **PQ-1** [Critical] `deliver-full-stated-purpose` ARCH-PURPOSE: the preserved first-milestone resolver still violates the required branch-point window
- **PQ-2** [Important] `test-strategy-contract` Compress enumerated test cases into named risky-function strategies
