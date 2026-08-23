---
gate: plan-quality
issue: 203
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-08-22T10:08:30-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Critical
          title: The class enumeration misses at least two in-class sites, so the plan patches instances of its own rule
          detail: |-
            Spec defines the class as every surface handing findings to the fixer, but the table lists four
            code sites and omits cmd/sdlc/changecode.go:572 (classifyFallback, the same plan gate's prose
            path) and cmd/sdlc/milestoneclose.go:626 ("address before crossing the boundary" — verbatim the
            phrasing the Problem section diagnoses). changecode.go:722 (estimate-quality) and judge.go:172
            (ad-hoc sdlc judge) need an explicit in-or-out ruling in Non-goals. Update the enumeration table,
            Done-when item 1, and the Plan's wire-the-call-sites row before implementation. ARCH-PURPOSE at-plan.
          family: incomplete-class-enumeration
          round: 1
        - id: PQ-2
          severity: Important
          title: A table-driven test over four named funcs cannot deliver Done-when 2's "a fifth call site cannot quietly ship"
          detail: |-
            The Plan's "table-driven over the emitting funcs" pins regression on the listed funcs and is
            structurally blind to a new one — demonstrated by the two unlisted sites already in the tree,
            over which that test would run green. Either specify an enumeration guard (a source scan over
            cmd/sdlc asserting the formatter is the only producer of a fixer-facing findings refusal), or
            weaken Done-when 2 to what an instance table actually buys. Do not keep the strong claim over
            the weak mechanism.
          family: guard-pinned-to-instance-list
          round: 1
        - id: PQ-3
          severity: Minor
          title: Test plan names no functions and no seam, though the emitting funcs differ in testability
          detail: |-
            Name the formatter and the emitting funcs (runPlanQualityJudge changecode.go:416, classifyFallback
            changecode.go:563, finalizeBoundaryReview close.go:1144, formatFixThenShipProtocol close.go:1806,
            the milestoneclose.go Failure arm), plus one line on which seam the new test rides:
            runPlanQualityJudge is already driven directly by changecode_test.go:100, while finalizeBoundaryReview
            has no direct driver and is reached via runCloseWithReview (close_finalize_test.go:335) and
            TestRunClose_LedgerGuardWiring (close_ledger_test.go:114).
          family: test-surface-unnamed
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-08-22T10:14:14-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Mechanical scan reproduced; all eight in-class sites and seven reasoned exclusions verified at the cited lines.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Plan added a source-scanning enumeration guard rather than weakening Done-when 2; sequenced red-first.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Emitting funcs, formatter file, and the two existing test seams are now named.
          round: 2
      blocked: false
content_hash: e2e6f1d7f5b781aa5d85a756e767698d7cc7e317825dc420b06633769a55b783
---

# Gate ledger — ariadne#203 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-22T10:08:30-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Critical] `incomplete-class-enumeration` The class enumeration misses at least two in-class sites, so the plan patches instances of its own rule
  Spec defines the class as every surface handing findings to the fixer, but the table lists four
  code sites and omits cmd/sdlc/changecode.go:572 (classifyFallback, the same plan gate's prose
  path) and cmd/sdlc/milestoneclose.go:626 ("address before crossing the boundary" — verbatim the
  phrasing the Problem section diagnoses). changecode.go:722 (estimate-quality) and judge.go:172
  (ad-hoc sdlc judge) need an explicit in-or-out ruling in Non-goals. Update the enumeration table,
  Done-when item 1, and the Plan's wire-the-call-sites row before implementation. ARCH-PURPOSE at-plan.
- **PQ-2** [Important] `guard-pinned-to-instance-list` A table-driven test over four named funcs cannot deliver Done-when 2's "a fifth call site cannot quietly ship"
  The Plan's "table-driven over the emitting funcs" pins regression on the listed funcs and is
  structurally blind to a new one — demonstrated by the two unlisted sites already in the tree,
  over which that test would run green. Either specify an enumeration guard (a source scan over
  cmd/sdlc asserting the formatter is the only producer of a fixer-facing findings refusal), or
  weaken Done-when 2 to what an instance table actually buys. Do not keep the strong claim over
  the weak mechanism.
- **PQ-3** [Minor] `test-surface-unnamed` Test plan names no functions and no seam, though the emitting funcs differ in testability
  Name the formatter and the emitting funcs (runPlanQualityJudge changecode.go:416, classifyFallback
  changecode.go:563, finalizeBoundaryReview close.go:1144, formatFixThenShipProtocol close.go:1806,
  the milestoneclose.go Failure arm), plus one line on which seam the new test rides:
  runPlanQualityJudge is already driven directly by changecode_test.go:100, while finalizeBoundaryReview
  has no direct driver and is reached via runCloseWithReview (close_finalize_test.go:335) and
  TestRunClose_LedgerGuardWiring (close_ledger_test.go:114).

## Round 2 — 2026-08-22T10:14:14-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Mechanical scan reproduced; all eight in-class sites and seven reasoned exclusions verified at the cited lines.
- PQ-2 — addressed — Plan added a source-scanning enumeration guard rather than weakening Done-when 2; sequenced red-first.
- PQ-3 — addressed — Emitting funcs, formatter file, and the two existing test seams are now named.

## Open findings

(none — every finding has been disposed)
