---
gate: boundary-review
issue: 335
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-14T04:21:39-07:00"
      dispatched: "2026-10-14T04:18:28-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Claude task spans inherit the parent session's later engagements
          detail: 'cmd/sdlc/internal/activetime/compute.go:152 merges task spans into session activity before ownership clipping. Claude spans carry no dispatch-engagement binding, so an Agent dispatched during a released interval and returning after re-claim contributes time inside the new engagement. Retain the original dispatch provenance and admit each span only within its dispatch engagement before unioning activity. ARCH-PURPOSE and ARCH-ORDER: the plan explicitly requires Claude Agent spans to retain dispatch engagement, alongside Codex child exclusion.'
          family: dispatch-engagement-attribution
          round: 1
        - kind: defect
          id: BR-2
          severity: Critical
          title: Excluding the newest sample revives an older sample for the same issue
          detail: 'cmd/sdlc/actualcorrection.go:769 targets only one row, and cmd/sdlc/internal/estimate/correction.go:175 removes both that target and the exclusion record. With two ordinary rows for one issue followed by an exclusion targeting the newest, downstream NewestPerIssue selects the older row. A production converge-report probe returned n=1 and the older row''s build=1.00h after exclusion. Preserve authoritative exclusion semantics through consumer selection, while allowing a legitimate later close to supersede the exclusion. ARCH-PURPOSE: the plan explicitly forbids an excluded newest repair from exposing a superseded trusted sample.'
          family: authoritative-calibration-exclusion
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: The close-time plausibility warning has no production-path regression
          detail: cmd/sdlc/close.go:526 introduces the warning, but cmd/sdlc/actualwiring_test.go:59 only tests actualEstimateWarning directly. No test asserts that omitted-actual whole-issue close emits the warning; deleting its production call leaves the comparator test green. Add a close-path regression for an implausible adopted measurement and a milestone control, and confirm the regression fails when the warning call is removed.
          family: production-adoption-warning-coverage
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#335 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T04:21:39-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `dispatch-engagement-attribution` (kind: defect) Claude task spans inherit the parent session's later engagements
  cmd/sdlc/internal/activetime/compute.go:152 merges task spans into session activity before ownership clipping. Claude spans carry no dispatch-engagement binding, so an Agent dispatched during a released interval and returning after re-claim contributes time inside the new engagement. Retain the original dispatch provenance and admit each span only within its dispatch engagement before unioning activity. ARCH-PURPOSE and ARCH-ORDER: the plan explicitly requires Claude Agent spans to retain dispatch engagement, alongside Codex child exclusion.
- **BR-2** [Critical] `authoritative-calibration-exclusion` (kind: defect) Excluding the newest sample revives an older sample for the same issue
  cmd/sdlc/actualcorrection.go:769 targets only one row, and cmd/sdlc/internal/estimate/correction.go:175 removes both that target and the exclusion record. With two ordinary rows for one issue followed by an exclusion targeting the newest, downstream NewestPerIssue selects the older row. A production converge-report probe returned n=1 and the older row's build=1.00h after exclusion. Preserve authoritative exclusion semantics through consumer selection, while allowing a legitimate later close to supersede the exclusion. ARCH-PURPOSE: the plan explicitly forbids an excluded newest repair from exposing a superseded trusted sample.
- **BR-3** [Important] `production-adoption-warning-coverage` (kind: defect) The close-time plausibility warning has no production-path regression
  cmd/sdlc/close.go:526 introduces the warning, but cmd/sdlc/actualwiring_test.go:59 only tests actualEstimateWarning directly. No test asserts that omitted-actual whole-issue close emits the warning; deleting its production call leaves the comparator test green. Add a close-path regression for an implausible adopted measurement and a milestone control, and confirm the regression fails when the warning call is removed.

## Open findings

- **BR-1** [Critical] `dispatch-engagement-attribution` (kind: defect) Claude task spans inherit the parent session's later engagements
- **BR-2** [Critical] `authoritative-calibration-exclusion` (kind: defect) Excluding the newest sample revives an older sample for the same issue
- **BR-3** [Important] `production-adoption-warning-coverage` (kind: defect) The close-time plausibility warning has no production-path regression
