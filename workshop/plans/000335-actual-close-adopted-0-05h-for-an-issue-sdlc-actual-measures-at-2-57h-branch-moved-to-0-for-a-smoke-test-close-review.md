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
