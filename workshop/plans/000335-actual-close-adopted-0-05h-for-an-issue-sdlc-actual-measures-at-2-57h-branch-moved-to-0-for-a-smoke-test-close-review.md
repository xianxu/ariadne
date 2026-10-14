# Boundary Review — ariadne#335 (whole-issue close)

| field | value |
|-------|-------|
| issue | 335 — actual: close adopted 0.05h for an issue sdlc actual measures at 2.57h (branch moved to :0 for a smoke test) |
| repo | ariadne |
| issue file | workshop/issues/000335-actual-close-adopted-0-05h-for-an-issue-sdlc-actual-measures-at-2-57h-branch-moved-to-0-for-a-smoke-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | ca9b09087f9c0992d0171fa691df8819336edbbf..c4bdbe1b1cb54c7a157200064de62b8da01f728f |
| command | sdlc close --issue 335 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T04:21:39-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The shared resolver, retained measurement inputs, and correction recovery model are substantial improvements. Targeted tests passed, but two correctness gaps remain: Claude task spans can cross engagement boundaries, and excluding the newest calibration sample can revive an older sample. The close-time warning also lacks the planned production-path regression.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: dispatch-engagement-attribution
    title: |
      Claude task spans inherit the parent session's later engagements
    detail: |
      cmd/sdlc/internal/activetime/compute.go:152 merges task spans into session activity before ownership clipping. Claude spans carry no dispatch-engagement binding, so an Agent dispatched during a released interval and returning after re-claim contributes time inside the new engagement. Retain the original dispatch provenance and admit each span only within its dispatch engagement before unioning activity. ARCH-PURPOSE and ARCH-ORDER: the plan explicitly requires Claude Agent spans to retain dispatch engagement, alongside Codex child exclusion.
  - id: new
    severity: Critical
    kind: defect
    family: authoritative-calibration-exclusion
    title: |
      Excluding the newest sample revives an older sample for the same issue
    detail: |
      cmd/sdlc/actualcorrection.go:769 targets only one row, and cmd/sdlc/internal/estimate/correction.go:175 removes both that target and the exclusion record. With two ordinary rows for one issue followed by an exclusion targeting the newest, downstream NewestPerIssue selects the older row. A production converge-report probe returned n=1 and the older row's build=1.00h after exclusion. Preserve authoritative exclusion semantics through consumer selection, while allowing a legitimate later close to supersede the exclusion. ARCH-PURPOSE: the plan explicitly forbids an excluded newest repair from exposing a superseded trusted sample.
  - id: new
    severity: Important
    kind: defect
    family: production-adoption-warning-coverage
    title: |
      The close-time plausibility warning has no production-path regression
    detail: |
      cmd/sdlc/close.go:526 introduces the warning, but cmd/sdlc/actualwiring_test.go:59 only tests actualEstimateWarning directly. No test asserts that omitted-actual whole-issue close emits the warning; deleting its production call leaves the comparator test green. Add a close-path regression for an implausible adopted measurement and a milestone control, and confirm the regression fails when the warning call is removed.
```

1. **Strengths**

   - Actual, close, and milestone adoption share the resolver and engine; production consumer parity is exercised in `actualmeasurement_test.go`.
   - Retained inputs bind replay to Git pins, transcript prefixes, and phase cuts. Tests cover later appends and changed-prefix refusal.
   - Correction recovery uses explicit phases and exact operation/candidate evidence, with tests for interruption and completed retries.
   - README, atlas, help, recovery catalog, and modeled field ownership were updated.

2. **Critical findings**

   - **Dispatch attribution:** [compute.go:152](cmd/sdlc/internal/activetime/compute.go:152). Preserve Claude span dispatch identity before activity union and clipping. Test both dispatch during release and dispatch in an earlier engagement.
   - **Calibration exclusion:** [correction.go:175](cmd/sdlc/internal/estimate/correction.go:175). Keep exclusion authority through selection. Test multiple historical rows through forecast, convergence, and drift consumers.

3. **Important findings**

   - **Warning coverage:** [close.go:526](cmd/sdlc/close.go:526). Exercise the real omitted-actual close path and assert its diagnostic.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Eight targeted internal packages passed. Selected root command tests passed in 58.418s. The full suite was not rerun.

   The existing `TestClipActivityTaskSpanAcrossRelease` expects a span to contribute to both ownership windows; it does not enforce dispatch lineage. Existing exclusion consumer fixtures have only one original row, missing the resurrection case. No prior findings required disposition.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared measurement execution and correction resolution are reused by consumers.
   - **ARCH-PURE — pass:** derivation, eligibility, clipping, and correction transitions have direct unit coverage.
   - **ARCH-PURPOSE — flag:** Claude dispatch attribution and exclusion authority remain incomplete.
   - **ARCH-MOCK — pass:** integration tests use portable repositories, transcript fixtures, and existing Git seams.
   - **ARCH-CONSTRAINTS — pass:** history and retained snapshots have explicit bounds; cancellation is tested.
   - **ARCH-SECURE — pass:** repository identity, retained prefixes, malformed correction evidence, and exact publication bindings receive validation.
   - **ARCH-ORDER — flag:** correction ordering is explicit, but Claude span attribution loses the engagement associated with dispatch.
   - **ARCH-FUNERAL — pass:** pending refs have completion cleanup; permanent evidence costs are documented.

7. **Plan revision recommendations**

   Add `## Revisions` entries specifying Claude span admission before activity union and exclusion precedence across multiple historical samples. Refresh the discovery file locations to `discovery.go` and the comparator location to `actualmeasurement.go`. Keep post-landing data repairs explicitly pending.

---

## Re-review — 2026-10-14T04:39:07-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 335 — actual: close adopted 0.05h for an issue sdlc actual measures at 2.57h (branch moved to :0 for a smoke test) |
| repo | ariadne |
| issue file | workshop/issues/000335-actual-close-adopted-0-05h-for-an-issue-sdlc-actual-measures-at-2-57h-branch-moved-to-0-for-a-smoke-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | ca9b09087f9c0992d0171fa691df8819336edbbf..e7fe9c345a32e5d5577c622e888c7cb72e2d2b1c |
| command | sdlc close --issue 335 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T04:39:07-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

All three prior findings are addressed, with regressions independently confirmed to fail when their fixes are removed. One new lifecycle defect blocks shipping: re-closing a corrected issue retains correction metadata from the previous completion generation, causing subsequent actual measurement to refuse. README and atlas updates are present; post-landing data repairs remain explicitly pending as documented.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Task spans retain original dispatch and are admitted within that engagement before union. Removing admission failed four dispatch regression cases; restoring it passed.
  - id: BR-2
    disposition: addressed
    note: |
      Validated exclusions suppress preceding issue samples before consumer selection. Removing the cutoff failed ordinary selection, forecast, convergence and drift assertions; restoring it passed. Production convergence and forecast regressions also passed.
  - id: BR-3
    disposition: addressed
    note: |
      TestOmittedActualCloseEstimateWarningWiring exercises production whole-issue close and a milestone control. Removing the production warning failed both low/high issue cases; restoring it passed.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: completion-bound-correction-lifecycle
    title: |
      Re-close retains a correction bound to the previous completion generation
    detail: |
      cmd/sdlc/actualinputs.go:106 rejects corrections whose completion binding differs, but internal/issue/handoff.go:217 replaces Completion without clearing ActualCorrection. A disposable-head regression using production tracker.CodecompleteCard followed by resolveOwnedActualInputs failed for both numeric replacement and exclusion with "actual correction belongs to a different completion binding". ARCH-ORDER/ARCH-PURPOSE: retire prior-generation correction metadata atomically when a new completion token is installed, preserve it for same-generation landing updates, and cover both correction dispositions with regressions.
```

1. **Strengths**

   - Dispatch provenance is enforced before activity union in `internal/activetime/compute.go:152`.
   - Shared correction resolution protects all inspected calibration consumers.
   - Retained transcript prefixes detect changed evidence while permitting later appends.
   - Production warning tests verify successful close and milestone behavior.

2. **Critical findings**

   - **Stale correction after re-close:** `cmd/sdlc/actualinputs.go:106` and `cmd/sdlc/internal/issue/handoff.go:217`. Clear the current correction when replacing its completion generation; retain historical evidence in Git. Preserve correction metadata when merely adding the same generation’s landing commit.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Six core packages passed, as did targeted ownership, replay, correction, warning, convergence and forecast tests. All three prior fixes passed removal/restoration mutation checks. The new re-close probe failed for both correction dispositions. Disposable copies were removed; the reviewed tree remains clean. No full suite was rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared measurement and calibration resolution.
   - **ARCH-PURE — pass:** ownership, eligibility, clipping and correction transitions have direct pure tests.
   - **ARCH-PURPOSE — flag:** correction lifecycle breaks legitimate subsequent close generations.
   - **ARCH-MOCK — pass:** portable Git repositories/remotes and transcript fixtures exercise integration boundaries.
   - **ARCH-CONSTRAINTS — pass:** explicit history, source-count and snapshot bounds refuse incomplete authority.
   - **ARCH-SECURE — pass:** persisted evidence is validated; changed prefixes refuse visibly.
   - **ARCH-ORDER — flag:** completion replacement leaves authoritative state bound to an obsolete generation.
   - **ARCH-FUNERAL — pass:** completed recovery refs are removed; permanent evidence costs are documented.

7. **Plan revision recommendations**

   Add a `## Revisions` entry defining correction lifetime across completion changes: new completion token retires the current correction; same-token landing preserves it. Require numeric-replacement and exclusion sequence tests through production card completion and measurement replay.

---

## Re-review — 2026-10-14T04:54:41-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 335 — actual: close adopted 0.05h for an issue sdlc actual measures at 2.57h (branch moved to :0 for a smoke test) |
| repo | ariadne |
| issue file | workshop/issues/000335-actual-close-adopted-0-05h-for-an-issue-sdlc-actual-measures-at-2-57h-branch-moved-to-0-for-a-smoke-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | ca9b09087f9c0992d0171fa691df8819336edbbf..e8db458421fcf144425095f1d10abfe89fe455e3 |
| command | sdlc close --issue 335 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T04:54:41-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

BR-4 is addressed: replacing a completion generation retires its correction metadata, while same-generation landing preserves it. Removing the fix made both replacement and exclusion regressions fail. Targeted tests passed. One inexpensive contract gap remains: correction previews discard the measurement diagnostics needed to assess historical repairs.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Dispatch admission occurs before activity union; passing engine tests cover released dispatch, prior engagements, clipped original dispatch, and continuous relocation.
  - id: BR-2
    disposition: addressed
    note: |
      ResolveCorrections suppresses earlier issue samples after exclusion; passing tests cover forecast, convergence, drift, later closes, and identical exclusion retries.
  - id: BR-3
    disposition: addressed
    note: |
      TestOmittedActualCloseEstimateWarningWiring passes through production whole-issue close, checks both ratio warnings, and verifies the milestone control.
  - id: BR-4
    disposition: addressed
    note: |
      SetCardCompletion atomically clears prior-token correction metadata. Removing that guard in a disposable pinned-head copy made both replacement/exclusion unit and production CodecompleteCard-to-resolver regressions fail; same-token preservation is covered.
findings:
  - id: new
    severity: Important
    kind: defect
    family: measurement-diagnostics-preservation
    title: |
      Correction previews discard measurement provenance and quality warnings
    detail: |
      cmd/sdlc/actualcorrection.go:672 consumes the measurement, but its successful numeric path prints only hours and hashes at line 776, discarding measured.Window and measured.Warnings. This hides historical reconstruction and attribution warnings despite the plan requiring the repair dry-run to name its reconstruction basis (workshop/plans/000335-actual-measurement-plan.md:120–124). ARCH-PURPOSE: render the bounded measurement diagnostics for successful previews and corrections, and add a command-path regression; a disposable extension of TestActualCorrectionDryRunAndUncertainRefusal failed because its supplied provenance warning appeared in neither stdout nor stderr.
```

### Strengths

- Shared measurement wiring pins ownership, code history, source prefixes, and phase cuts across preview and adoption.
- Completion retirement is centralized in `cmd/sdlc/internal/issue/handoff.go:219`, covering both correction dispositions.
- Correction recovery preserves confirmed progress and verifies operation identities before advancing effects.
- README, atlas, and command help document the new correction and replay surfaces.

### Critical findings

None.

### Important findings

- `cmd/sdlc/actualcorrection.go:672`: preserve successful measurement diagnostics in correction output. Reuse the existing diagnostic rendering where practical; test reconstruction warnings, window output, and unchanged dry-run write behavior.

### Minor findings

None.

### Test coverage notes

Eight affected internal packages passed, along with targeted command tests covering measurement, corrections, calibration consumers, and adoption. BR-4’s fix-removal experiment failed as expected for both dispositions. The additional diagnostic assertion failed against pinned head. Disposable copies were removed; the reviewed checkout remains unchanged. The full suite was not rerun.

### Architectural notes

- **ARCH-DRY — pass:** measurement execution and calibration resolution are shared.
- **ARCH-PURE — pass:** derivation, eligibility, clipping, validation, and correction transitions have direct pure tests.
- **ARCH-PURPOSE — flag:** correction output loses required repair diagnostics.
- **ARCH-MOCK — pass:** portable repositories and transcript fixtures exercise production boundaries.
- **ARCH-CONSTRAINTS — pass:** history and retained evidence have explicit limits; cancellation is covered.
- **ARCH-SECURE — pass:** generation, repository, prefix, and persisted-operation validation reject inconsistent evidence.
- **ARCH-ORDER — pass:** correction phases and confirmation identities are explicit; BR-4 closes the generation lifecycle gap.
- **ARCH-FUNERAL — pass:** completed recovery refs are removed; permanent evidence has documented bounded per-operation cost.

### Plan revision recommendations

None required. The diagnostic fix fulfills the existing contract. Landing, shared-binary rebuild, and peer-data repairs remain explicitly post-landing acceptance work.
