# Boundary Review — ariadne#527 (whole-issue close)

| field | value |
|-------|-------|
| issue | 527 — landing split: don't hold a car for a registered flake or a failure in a package it doesn't touch |
| repo | ariadne |
| issue file | workshop/issues/000527-landing-split-don-t-hold-a-car-for-a-registered-flake-or-a-failure-in-a-package-it-doesn-t-touch.md |
| boundary | whole-issue close |
| milestone | — |
| window | e3a4c617f55f29aa489e4718aecddcd4a8679254..f2ad2f5c25200df5e05e4791b172466bfa7fc1ca |
| command | sdlc close --issue 527 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T10:32:07-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation fulfills the issue’s no-false-blame contract while preserving passing-suite requirements. Exact-prefix attribution now precedes suite Holds, fresh registration and dependency evidence inform the decision, and unresolved cars stop without automatic retraining. Focused tests passed, including the production registration, untouched-package, genuine-culprit, and reintegration regressions. One minor plan classification needs correction.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: pure-integration-classification
    title: |
      Plan labels the IO-bearing splitImpact function as pure
    detail: |
      workshop/plans/000527-landing-split-attribution-plan.md:89 calls splitImpact a pure helper, but cmd/sdlc/landingattribution_impact.go:20 performs Git operations, creates worktrees, executes graph preparation, and loads Go graphs. Update its classification to INTEGRATION and add it to the integration concepts table; the implementation already keeps this IO outside the pure reducer (ARCH-PURE).
```

1. **Strengths**

   - `trainstate.go` introduces an explicit attribution phase and rejects mismatched, duplicate, and out-of-phase assessment events.
   - `landingtrain_attribution.go` checks prefix SHA, log hash, timeout status, and parser completeness before using failure evidence.
   - `landingattribution_impact.go` unions both endpoint graphs, preserving dependency evidence when imports or resources change.
   - The production regressions verify observable outcomes: no false Hold, no failed-prefix landing, explicit retry guidance, and preserved genuine-culprit handling.
   - Code-main reintegration invalidates stale prefix evidence and stops before unsafe bisection.

2. **Critical findings**

   None.

3. **Important findings**

   None.

4. **Minor findings**

   The plan’s `splitImpact` classification is inaccurate, as recorded above.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/sdlc/internal/landingline ./cmd/sdlc/internal/flakes ./cmd/sdlc/internal/affected`
   - Focused root-package production, impact, evidence, recovery, and reintegration tests: **54.299s**
   - `git diff --check` for the pinned range.

   Tests cover fresh external registration through an explicit barrier, untouched packages, genuine culprit holding, dependency/resource impact, corrupt evidence, cancellation cleanup, and stale-prefix reintegration. The full repository suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Reuses flake coverage, affected graph selection, and existing refusal/landing effects.
   - **ARCH-PURE — pass in code; documentation flag:** Classification and transitions remain pure; assessment IO stays in the shell.
   - **ARCH-PURPOSE — pass:** Suite-blame paths require assessment, including first-car and single-car cases.
   - **ARCH-MOCK — pass:** Production-path tests use the existing stateful GitHub fake and isolated real Git/tracker fixtures.
   - **ARCH-CONSTRAINTS — pass:** Assessment has an outer deadline; tests exercise cancellation and constant endpoint-probe counts.
   - **ARCH-SECURE — pass:** Invalid or incomplete evidence becomes visibly unverified without granting passing authority.
   - **ARCH-ORDER — pass:** Exact-pair events, terminal refusal, and reintegration invalidation make ordering explicit.
   - **ARCH-FUNERAL — pass:** Evidence is conductor-scoped; temporary worktrees have cleanup paths.
   - **ARCH-EVOLVE — pass:** No persisted format or outcome enum changes.

   Atlas documentation covers the new phase and policy. No new command, flag, or configuration key requires a README update.

7. **Plan revision recommendation**

   Add a `## Revisions` entry stating that `splitImpact` is an **INTEGRATION** helper using the existing pure affected graph functions, and correct its test-strategy row and concepts inventory accordingly.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9d197645-f0d2-4104-bf34-93a7a546009f","repository":"github.com/xianxu/ariadne","issue_id":"000527","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/c6e5f47456665513","reviewed_head":"f2ad2f5c25200df5e05e4791b172466bfa7fc1ca","slot":"ariadne:2","requested_at":"2026-10-21T17:13:08.619365Z"},"phase":"finished","pid":87289,"started_at":"2026-10-21T17:13:08.780536Z","finished_at":"2026-10-21T17:32:06.726307Z","session_id":"01a15f73-d51c-7210-a41a-fc8e51adb6c6","outcome":"returned"}] -->
