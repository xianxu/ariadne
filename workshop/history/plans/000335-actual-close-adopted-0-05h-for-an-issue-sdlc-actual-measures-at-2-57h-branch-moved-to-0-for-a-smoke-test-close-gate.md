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
    - "n": 2
      timestamp: "2026-10-14T04:39:07-07:00"
      dispatched: "2026-10-14T04:34:54-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Task spans retain original dispatch and are admitted within that engagement before union. Removing admission failed four dispatch regression cases; restoring it passed.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Validated exclusions suppress preceding issue samples before consumer selection. Removing the cutoff failed ordinary selection, forecast, convergence and drift assertions; restoring it passed. Production convergence and forecast regressions also passed.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestOmittedActualCloseEstimateWarningWiring exercises production whole-issue close and a milestone control. Removing the production warning failed both low/high issue cases; restoring it passed.
          round: 2
      findings:
        - kind: defect
          id: BR-4
          severity: Critical
          title: Re-close retains a correction bound to the previous completion generation
          detail: 'cmd/sdlc/actualinputs.go:106 rejects corrections whose completion binding differs, but internal/issue/handoff.go:217 replaces Completion without clearing ActualCorrection. A disposable-head regression using production tracker.CodecompleteCard followed by resolveOwnedActualInputs failed for both numeric replacement and exclusion with "actual correction belongs to a different completion binding". ARCH-ORDER/ARCH-PURPOSE: retire prior-generation correction metadata atomically when a new completion token is installed, preserve it for same-generation landing updates, and cover both correction dispositions with regressions.'
          family: completion-bound-correction-lifecycle
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 3
      timestamp: "2026-10-14T04:54:41-07:00"
      dispatched: "2026-10-14T04:50:51-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Dispatch admission occurs before activity union; passing engine tests cover released dispatch, prior engagements, clipped original dispatch, and continuous relocation.
          round: 3
        - id: BR-2
          disposition: addressed
          note: ResolveCorrections suppresses earlier issue samples after exclusion; passing tests cover forecast, convergence, drift, later closes, and identical exclusion retries.
          round: 3
        - id: BR-3
          disposition: addressed
          note: TestOmittedActualCloseEstimateWarningWiring passes through production whole-issue close, checks both ratio warnings, and verifies the milestone control.
          round: 3
        - id: BR-4
          disposition: addressed
          note: SetCardCompletion atomically clears prior-token correction metadata. Removing that guard in a disposable pinned-head copy made both replacement/exclusion unit and production CodecompleteCard-to-resolver regressions fail; same-token preservation is covered.
          round: 3
      findings:
        - kind: defect
          id: BR-5
          severity: Important
          title: Correction previews discard measurement provenance and quality warnings
          detail: 'cmd/sdlc/actualcorrection.go:672 consumes the measurement, but its successful numeric path prints only hours and hashes at line 776, discarding measured.Window and measured.Warnings. This hides historical reconstruction and attribution warnings despite the plan requiring the repair dry-run to name its reconstruction basis (workshop/plans/000335-actual-measurement-plan.md:120–124). ARCH-PURPOSE: render the bounded measurement diagnostics for successful previews and corrections, and add a command-path regression; a disposable extension of TestActualCorrectionDryRunAndUncertainRefusal failed because its supplied provenance warning appeared in neither stdout nor stderr.'
          family: measurement-diagnostics-preservation
          round: 3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - "n": 4
      timestamp: "2026-10-14T05:08:48-07:00"
      dispatched: "2026-10-14T05:05:21-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Dispatch-engagement admission remains enforced before span union; active-time regression tests passed.
          round: 4
        - id: BR-2
          disposition: addressed
          note: ResolveCorrections suppresses preceding issue samples before eligibility filtering; internal and command consumer regressions passed.
          round: 4
        - id: BR-3
          disposition: addressed
          note: TestOmittedActualCloseEstimateWarningWiring passed its low/high whole-close cases and milestone control.
          round: 4
        - id: BR-4
          disposition: addressed
          note: SetCardCompletion retires correction metadata on token replacement; production re-close and same-generation preservation regressions passed.
          round: 4
        - id: BR-5
          disposition: addressed
          note: actualcorrection.go calls printActualMeasurement after numeric preflight. TestActualCorrectionNumericDiagnostics passed; deleting that call in a disposable pinned-head copy failed both preview and execution, and restoration passed.
          round: 4
      recipe: milestone-review
      isolation: fenced
      reviewed: 25c8e1661a59c361139f9433a7ef3534081b746e
      blocked: false
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

## Round 2 — 2026-10-14T04:39:07-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Task spans retain original dispatch and are admitted within that engagement before union. Removing admission failed four dispatch regression cases; restoring it passed.
- BR-2 — addressed — Validated exclusions suppress preceding issue samples before consumer selection. Removing the cutoff failed ordinary selection, forecast, convergence and drift assertions; restoring it passed. Production convergence and forecast regressions also passed.
- BR-3 — addressed — TestOmittedActualCloseEstimateWarningWiring exercises production whole-issue close and a milestone control. Removing the production warning failed both low/high issue cases; restoring it passed.

### Raised

- **BR-4** [Critical] `completion-bound-correction-lifecycle` (kind: defect) Re-close retains a correction bound to the previous completion generation
  cmd/sdlc/actualinputs.go:106 rejects corrections whose completion binding differs, but internal/issue/handoff.go:217 replaces Completion without clearing ActualCorrection. A disposable-head regression using production tracker.CodecompleteCard followed by resolveOwnedActualInputs failed for both numeric replacement and exclusion with "actual correction belongs to a different completion binding". ARCH-ORDER/ARCH-PURPOSE: retire prior-generation correction metadata atomically when a new completion token is installed, preserve it for same-generation landing updates, and cover both correction dispositions with regressions.

## Round 3 — 2026-10-14T04:54:41-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Dispatch admission occurs before activity union; passing engine tests cover released dispatch, prior engagements, clipped original dispatch, and continuous relocation.
- BR-2 — addressed — ResolveCorrections suppresses earlier issue samples after exclusion; passing tests cover forecast, convergence, drift, later closes, and identical exclusion retries.
- BR-3 — addressed — TestOmittedActualCloseEstimateWarningWiring passes through production whole-issue close, checks both ratio warnings, and verifies the milestone control.
- BR-4 — addressed — SetCardCompletion atomically clears prior-token correction metadata. Removing that guard in a disposable pinned-head copy made both replacement/exclusion unit and production CodecompleteCard-to-resolver regressions fail; same-token preservation is covered.

### Raised

- **BR-5** [Important] `measurement-diagnostics-preservation` (kind: defect) Correction previews discard measurement provenance and quality warnings
  cmd/sdlc/actualcorrection.go:672 consumes the measurement, but its successful numeric path prints only hours and hashes at line 776, discarding measured.Window and measured.Warnings. This hides historical reconstruction and attribution warnings despite the plan requiring the repair dry-run to name its reconstruction basis (workshop/plans/000335-actual-measurement-plan.md:120–124). ARCH-PURPOSE: render the bounded measurement diagnostics for successful previews and corrections, and add a command-path regression; a disposable extension of TestActualCorrectionDryRunAndUncertainRefusal failed because its supplied provenance warning appeared in neither stdout nor stderr.

## Round 4 — 2026-10-14T05:08:48-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Dispatch-engagement admission remains enforced before span union; active-time regression tests passed.
- BR-2 — addressed — ResolveCorrections suppresses preceding issue samples before eligibility filtering; internal and command consumer regressions passed.
- BR-3 — addressed — TestOmittedActualCloseEstimateWarningWiring passed its low/high whole-close cases and milestone control.
- BR-4 — addressed — SetCardCompletion retires correction metadata on token replacement; production re-close and same-generation preservation regressions passed.
- BR-5 — addressed — actualcorrection.go calls printActualMeasurement after numeric preflight. TestActualCorrectionNumericDiagnostics passed; deleting that call in a disposable pinned-head copy failed both preview and execution, and restoration passed.

## Open findings

(none — every finding has been disposed)
