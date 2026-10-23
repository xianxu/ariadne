# Boundary Review — ariadne#572 (whole-issue close)

| field | value |
|-------|-------|
| issue | 572 — local-check: refuse to record a suite attempt invalidated by sandbox denials |
| repo | ariadne |
| issue file | workshop/issues/000572-local-check-refuse-to-record-a-suite-attempt-invalidated-by-sandbox-denials.md |
| boundary | whole-issue close |
| milestone | — |
| window | ee4dd7fa1e8488e3230dc462971be8796f317e41..214e4300379c9dded9180059db910a9a02232be1 |
| command | sdlc close --issue 572 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T05:00:29-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation follows the revised Spec: independent capability failures prevent execution, while denial-looking test output retains normal failure attribution. Focused tests pass. Shipping is blocked by a failing required lint check, missing README guidance, and a Core concepts classification that contradicts the implementation.

```findings
findings:
  - id: new
    severity: Critical
    family: core-concept-classification
    title: |
      ARCH-PURE: the plan names a pure diagnosis entity that does not exist at its stated location
    detail: |
      workshop/plans/000572-suite-capability-preflight-plan.md:31-34 lists a new pure capability-error diagnosis in suitecapability.go. That file contains three integration functions, each performing or invoking IO; the described EPERM discrimination actually resides in suiteEnvironmentError.Error at suiteenvironment.go:57. Revise the table and explanatory paragraph to identify the actual pure formatter and integration entities, or implement and directly test the promised pure helper. The review contract explicitly makes Core concepts contradictions Critical.
  - id: new
    severity: Important
    family: required-check-regression
    title: |
      The new unused suppression fails the required golint check
    detail: |
      cmd/sdlc/suitecapability.go:70 adds a tmp-literal suppression, but that rule only examines calls through os, filepath.Join, and ioutil; scratchdisk.Roots produces no finding. Running go test ./pkg/golint -count=1 fails TestTrackedGoFilesPassTheCatalog with “suppression of tmp-literal matches no finding; remove it”. Remove the suppression, retain its useful explanatory comment, and rerun the check.
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README update appears missing for host-capability refusal and whole-call retry
    detail: |
      cmd/sdlc/helptext/local-check.md:347-359 introduces a user-facing refusal condition and specific retry instructions, but README.md is unchanged in the pinned range. Update the recorded-local-check usage near README.md:765 to explain could-not-run, the whole-call retry requirement, and preservation of existing failures. Atlas and command help already cover this surface.
```

1. **Strengths**

   - Shared preflight precedes ordinary execution allocation and affected graph preparation.
   - Tests verify no new attempt/status on refusal, same-head repair, parent preservation, and saved-pass repost bypass.
   - Captured mixed output is hash-checked; both mixed and all-denial controls remain failures.
   - Deferred cleanup preserves errors through `errors.Join` instead of reporting false readiness.

2. **Critical findings**

   - Core concepts mismatch above. A documentation correction can resolve it; extracting another helper is unnecessary if the plan accurately describes the existing formatter.

3. **Important findings**

   - Required golint failure and missing README update above.

4. **Minor findings**

   - None.

5. **Test coverage notes**

   - PASS: `go test ./cmd/sdlc -run 'TestSuiteCapability|TestSuiteEnvironment|TestSuiteInfrastructure|TestResume' -count=1 -timeout 20m` — 54.885s.
   - FAIL: `go test ./pkg/golint -count=1` — unused suppression.
   - Native conformance could not rerun: both children failed at `sandbox_apply: Operation not permitted`. Retained native evidence pins match the reviewed source blobs; this review independently confirmed deterministic refusal behavior.
   - The planned resume refusal → capability restoration → successful retry sequence is not explicitly exercised by the new resume test.
   - Checkout remained clean.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — pass:** one shared capability probe across execution entry points.
   - **ARCH-PURE — flag:** plan classification finding above; the actual formatter is directly tested without IO.
   - **ARCH-PURPOSE — pass:** fulfills the revised independent-evidence contract without granting log text waiver authority.
   - **ARCH-MOCK — pass:** injected resource ledger and existing repository/status fixtures exercise production seams; native conformance is supplied.
   - **ARCH-CONSTRAINTS — pass:** fixed capability operations; reused sweep population cost is explicitly acknowledged.
   - **ARCH-SECURE — pass:** wrapped error identity controls advice; private scratch and argv-based native invocation preserve boundaries.
   - **ARCH-ORDER — pass:** synchronous ownership and cleanup; resume checks precede child/link creation.
   - **ARCH-FUNERAL — pass:** normal cleanup plus existing stale-directory reclamation.
   - **ARCH-EVOLVE — pass:** no persisted-format change.

7. **Plan revision recommendations**

   Add a `## Revisions` entry correcting the pure diagnosis location/classification and naming the actual functions. Include affected graph preflight in the integration table. Keep the outstanding recorded-suite and landing gates explicitly pending.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3c09dc2b-aeb5-421c-8509-7db103d3f721","repository":"github.com/xianxu/ariadne","issue_id":"000572","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/db28c916f831cd4a","reviewed_head":"214e4300379c9dded9180059db910a9a02232be1","slot":"ariadne:4","requested_at":"2026-10-23T11:58:00.170968Z"},"phase":"finished","pid":72330,"started_at":"2026-10-23T11:58:00.335119Z","finished_at":"2026-10-23T12:00:28.757222Z","session_id":"01a168a0-084b-7812-b4ba-816e2d6c6f72","outcome":"returned"}] -->
