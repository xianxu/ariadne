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

---

## Re-review — 2026-10-23T05:08:49-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 572 — local-check: refuse to record a suite attempt invalidated by sandbox denials |
| repo | ariadne |
| issue file | workshop/issues/000572-local-check-refuse-to-record-a-suite-attempt-invalidated-by-sandbox-denials.md |
| boundary | whole-issue close |
| milestone | — |
| window | ee4dd7fa1e8488e3230dc462971be8796f317e41..7dad3c36d67a7f86404356c9b2b2cf9888c5b12a |
| command | sdlc close --issue 572 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T05:08:49-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the revised Spec: independent host-capability failures refuse execution without creating attempt or status evidence, while denial-looking test output remains ordinary failure evidence. All three prior findings are addressed; no new findings. The recorded final-source suite remains a separate prerequisite before landing.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plan lines 27–42 now identify the pure suiteEnvironmentError formatter in suiteenvironment.go and separately enumerate the actual integration functions and entry points. The formatter and its direct, IO-free advice tests support the corrected classification.
  - id: BR-2
    disposition: addressed
    note: |
      suitecapability.go:70 retains the short-path explanation without the unused suppression. The existing TestTrackedGoFilesPassTheCatalog regression check passes; pkg/golint/check.go:132–135 still rejects unused suppressions.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:772–777 now documents could-not-run, whole-call retry, preservation of earlier failures, and normal attribution on healthy hosts. These statements match preflightSuite, the error formatter, and the passing execution tests.
```

1. **Strengths**

   - Shared preflight precedes execution allocation, affected graph preparation, and flake/resume child creation.
   - Tests verify absent side effects, same-head repair, genuine-parent preservation, and saved-pass repost bypass.
   - Hash-checked captured mixed output and synthetic all-denial output both retain failure attribution.
   - Deferred cleanup joins errors rather than reporting false readiness.

2. **Critical findings**

   None.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - PASS: focused capability/environment/infrastructure/resume tests — 35.559s.
   - PASS: affected and baseline tests — 46.057s.
   - PASS: `go test ./pkg/golint -count=1` — 1.554s.
   - Native rerun was unavailable: both children failed at `sandbox_apply: Operation not permitted`. Retained host-run evidence matches the pinned production/probe blobs; this review does not claim a local native PASS.
   - Checkout remained clean at the pinned head.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one shared probe across the enumerated entry points.
   - **ARCH-PURE — pass:** pure formatting is directly tested; IO remains in integration functions.
   - **ARCH-PURPOSE — pass:** fulfills the revised independent-evidence contract without granting log text waiver authority.
   - **ARCH-MOCK — pass:** stateful resource fixtures and repository/status doubles exercise production seams; native conformance covers kernel behavior.
   - **ARCH-CONSTRAINTS — pass:** fixed capability operations; sweep population cost is explicitly acknowledged.
   - **ARCH-SECURE — pass:** wrapped error identity controls advice; scratch is private; captured output cannot authorize invalidation.
   - **ARCH-ORDER — pass:** synchronous resource lifetime, cleanup ordering, and resume proof-before-child behavior are covered.
   - **ARCH-FUNERAL — pass:** normal cleanup and existing stale-directory reclamation cover probe residue.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.

7. **Plan revision recommendations**

   None. The existing revisions accurately describe the implementation and preserve the outstanding suite and landing gates.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2d5e394d-436d-4718-8dfb-aa5bdfaf00e5","repository":"github.com/xianxu/ariadne","issue_id":"000572","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/64d08af6eb9061eb","reviewed_head":"7dad3c36d67a7f86404356c9b2b2cf9888c5b12a","slot":"ariadne:4","requested_at":"2026-10-23T12:06:01.958125Z"},"phase":"finished","pid":41180,"started_at":"2026-10-23T12:06:02.107042Z","finished_at":"2026-10-23T12:08:48.670038Z","session_id":"01a168a7-6215-70e2-8e96-899ad48bf7ba","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T06:22:44-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 572 — local-check: refuse to record a suite attempt invalidated by sandbox denials |
| repo | ariadne |
| issue file | workshop/issues/000572-local-check-refuse-to-record-a-suite-attempt-invalidated-by-sandbox-denials.md |
| boundary | whole-issue close |
| milestone | — |
| window | 48b621314745466eb9f5c9c92c76614ca42a4633..37a4536e423b774dd71d6ca746783de445b5b3c4 |
| command | sdlc close --issue 572 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-23T06:22:44-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff repairs the existing short-socket contract without expanding scope. Canonical paths are checked before bind, rejected directories are removed, and operational errors still refuse immediately. All prior dispositions remain valid; no new defects found. A recorded suite PASS remains required before landing.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The plan identifies suiteEnvironmentError.Error in suiteenvironment.go as pure and separately lists the IO functions and execution boundaries. Direct advice tests exercise formatting without IO.
  - id: BR-2
    disposition: addressed
    note: |
      suitecapability.go retains an ordinary explanatory comment without the unused suppression. The tracked-file golint regression passes; unused suppressions remain rejected.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:772–777 documents could-not-run, whole-call retry, preservation of earlier failures, and healthy-host failure attribution, matching the formatter and execution contract.
```

1. **Strengths**

   - `cmd/sdlc/suitecapability.go:87–95` checks the canonical final socket path and removes rejected scratch before fallback.
   - `cmd/sdlc/suitecapability.go:79–101` bounds placement selection and preserves creation, canonicalization, and cleanup errors.
   - Regression coverage exercises long configured roots and symlink targets. Removing the production fix makes the new test fail with `bind: invalid argument`.

2. **Critical findings**

   None.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Focused capability, environment, infrastructure, and original scratch tests: PASS, 23.209s.
   - `go test ./pkg/golint -count=1`: PASS, 1.242s.
   - Without-fix regression: expected FAIL in a disposable clone; clone removed afterward.
   - Native conformance rerun could not apply Seatbelt: `sandbox_apply: Operation not permitted`. Retained host evidence pins production and native-test blobs matching this head; no local native PASS is claimed.
   - Reviewed checkout remains clean.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared probe remains the single owner; no duplicated caller logic.
   - **ARCH-PURE — pass:** filesystem placement remains in the existing IO boundary.
   - **ARCH-PURPOSE — pass:** covers canonical configured-root lengths, including symlinks, while preserving denial refusal.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.

   The correction adds no user-facing command, flag, or architectural surface requiring further README or atlas changes.

7. **Plan revision recommendations**

   None. The appended revision records the bounded correction; Spec and Done-when remain unchanged.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"20e3d60d-2989-460e-a207-a60c08db42d7","repository":"github.com/xianxu/ariadne","issue_id":"000572","recipe":"fix-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/454dc8e333029bbd","reviewed_head":"37a4536e423b774dd71d6ca746783de445b5b3c4","slot":"ariadne:4","requested_at":"2026-10-23T13:20:47.524863Z"},"phase":"finished","pid":73874,"started_at":"2026-10-23T13:20:47.652634Z","finished_at":"2026-10-23T13:22:42.567384Z","session_id":"01a168eb-d359-7693-9ba7-aa22f8080541","outcome":"returned"}] -->
