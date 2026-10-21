# Boundary Review — ariadne#498 (whole-issue close)

| field | value |
|-------|-------|
| issue | 498 — flake: TestLandingTrainLandsEveryCarOnOneSuite/train |
| repo | ariadne |
| issue file | workshop/issues/000498-flake-testlandingtrainlandseverycarononesuite-train.md |
| boundary | whole-issue close |
| milestone | — |
| window | d3614beac4564191b7ef008e35d1a73a44e4bdb8..e51983dee5a21f532fa8d1ae83749d55b59043f3 |
| command | sdlc close --issue 498 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T16:56:20-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation matches the shared-budget design, and the affected controls pass independently. No code defect was found. Whole-issue acceptance remains blocked by missing evidence for the explicitly required recorded full suite.

```findings
findings:
  - id: new
    severity: Important
    kind: unverifiable
    family: boundary-verification-evidence
    title: |
      Required recorded full-suite pass is not established
    detail: |
      workshop/issues/000498-flake-testlandingtrainlandseverycarononesuite-train.md:43 requires a passing full suite; line 49 leaves that verification unchecked, and line 65 explicitly says it remains next. Searching /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/attempts for pinned head e51983dee5a21f532fa8d1ae83749d55b59043f3 returned no matching completed record. These are all instances of this acceptance-evidence gap in the window. Supply the recorded sharded suite result and inspectable output for this head, resolve any failures, and update the verification log before closing or unregistering the flakes.
```

1. **Strengths**

   - `finished()` and `wantSettled()` share one completion policy while retaining their existing error handling.
   - Budget arithmetic covers short, loaded, absent, expired, and exact-deadline cases without IO.
   - Diagnostic tests exercise both actual helpers and verify another goroutine’s stack appears before Go’s outer alarm.
   - The lint regression covers aliased imports and preserves intentional readiness/deadline tests.

2. **Critical findings:** None.

3. **Important findings:** The full-suite evidence gap above; no implementation change is currently indicated.

4. **Minor findings:** None.

5. **Test coverage notes**

   Independently passed:

   - `go test ./pkg/golint`
   - Focused archive, all-car train/null/head, malformed-car, and completion regressions: **71.636s**
   - Pinned-range `git diff --check`

   These focused passes do not establish the required full-suite result.

6. **Architectural notes**

   - **ARCH-DRY — pass:** completion consumers share the budget and diagnostic helper.
   - **ARCH-PURE — pass:** deterministic arithmetic is separated from clock, timer, and diagnostic IO.
   - **ARCH-PURPOSE — flag:** implementation serves the stated purpose, but required whole-suite verification remains unestablished.
   - **ARCH-EVOLVE — pass:** no persisted format changes.

   Atlas documents the new lint rule. No new command, flag, or configuration surface requires a README change.

7. **Plan revision recommendations:** None. Complete the existing verification checklist and preserve the verification-before-unregistration ordering.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"07f86530-57d7-42fe-bfda-a04e3a701986","repository":"github.com/xianxu/ariadne","issue_id":"000498","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/0ec0960e6cad7516","reviewed_head":"e51983dee5a21f532fa8d1ae83749d55b59043f3","slot":"ariadne:1","requested_at":"2026-10-20T23:54:33.043266Z"},"phase":"finished","pid":61564,"started_at":"2026-10-20T23:54:33.194792Z","finished_at":"2026-10-20T23:56:20.035795Z","session_id":"01a15bbc-f8d0-7e20-bc53-55a6052723ef","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T17:43:35-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 498 — flake: TestLandingTrainLandsEveryCarOnOneSuite/train |
| repo | ariadne |
| issue file | workshop/issues/000498-flake-testlandingtrainlandseverycarononesuite-train.md |
| boundary | whole-issue close |
| milestone | — |
| window | d3614beac4564191b7ef008e35d1a73a44e4bdb8..78cc0cbd3b00141880769731fb24bd164a1e8fbb |
| command | sdlc close --issue 498 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T17:43:35-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation satisfies the documented shared-budget design. Focused controls pass, regression tests fail when the correction is removed, and the recorded full-suite evidence resolves BR-1. No new findings.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Inspected completed attempt 706d1c6db69cb8835e2befae20926eb5: exit 0, outcome passed, all eight shards and 75 other packages. Verified suite.log SHA256 bb46b4829d2fde6e2c242130fe31c4b04ac7cd32de7b3947559489b7b65278df. The tested implementation commit e51983d differs from pinned head only in tracker/review documentation; issue lines 69–70 record verification before unregistration.
```

1. **Strengths**

   - `finished()` and `wantSettled()` share one completion policy and preserve error handling.
   - Budget tests cover short, loaded, absent, expired, and exact-deadline cases.
   - Diagnostics exercise both actual helpers and include an independently blocked goroutine’s stack.
   - The narrow lint guard covers the completion helpers and competing archive timeout while allowing intentional readiness/deadline tests.

2. **Critical findings:** None.

3. **Important findings:** None remaining.

4. **Minor findings:** None.

5. **Test coverage notes**

   - `go test ./pkg/golint -count=1` passed.
   - Focused archive, train/null/head, malformed-car, budget, error, and diagnostic controls passed: **52.927s**.
   - Temporary overlays restoring both old completion helpers caused both diagnostic regressions to fail at Go’s outer alarm.
   - Disabling the lint matcher caused its regression cases to fail.
   - Recorded full suite verified: **1,877 cmd/sdlc tests, 75 other packages, eight shards**, all passing.
   - Pinned-range `git diff --check` passed.

6. **Architectural notes**

   - **ARCH-DRY — pass:** all identified completion consumers share the policy.
   - **ARCH-PURE — pass:** deterministic arithmetic is separate from clock, timer, and diagnostic IO.
   - **ARCH-PURPOSE — pass:** both incidents and sibling completion waits are covered; required verification is established.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.

   Atlas documents the new lint rule. No new command, flag, or configuration surface requires a README update.

7. **Plan revision recommendations:** None. Existing close and landing checklist items remain appropriate.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"264afb2e-c7c8-47af-a6f9-afd7e7b22a4e","repository":"github.com/xianxu/ariadne","issue_id":"000498","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2c274af53f43b5fa","reviewed_head":"78cc0cbd3b00141880769731fb24bd164a1e8fbb","slot":"ariadne:1","requested_at":"2026-10-21T00:41:38.693942Z"},"phase":"finished","pid":45075,"started_at":"2026-10-21T00:41:38.8013Z","finished_at":"2026-10-21T00:43:34.82298Z","session_id":"01a15be8-160b-7492-a4ad-6b159e2fa13f","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T20:17:57-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 498 — flake: TestLandingTrainLandsEveryCarOnOneSuite/train |
| repo | ariadne |
| issue file | workshop/issues/000498-flake-testlandingtrainlandseverycarononesuite-train.md |
| boundary | whole-issue close |
| milestone | — |
| window | b68b00d48c91bf77d214ae4487b60d68510e5619..18ef5d0ed97b172b1e1d02eb0b204b4adf76a475 |
| command | sdlc close --issue 498 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T20:17:57-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The conflict resolutions preserve all three lint rules and their atlas documentation. The interdiff leaves matcher behavior, Spec, and Done-when unchanged. No new defects found; independently inspected suite evidence confirms BR-1 remains addressed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Verified attempt 706d1c6db69cb8835e2befae20926eb5 passed with exit 0 and its log matches SHA256 bb46b4829d2fde6e2c242130fe31c4b04ac7cd32de7b3947559489b7b65278df. Also verified integrated attempt a1a8983f0805bb5c183a550dedd34044 passed all eight shards and 75 other packages; its log matches SHA256 1afa234db8460f35ecdf14bd44cc23f05073e4fa671df181772131ecbda81984. Tested integrated commit 73f4695 differs from pinned head only in the issue log.
```

1. **Strengths**
   - `pkg/golint/rules.go:15` retains the completion-wait rule and both session-move rules, with their original matchers.
   - `atlas/workflow/sdlc-binary.md:2215` documents all three retained entries.
   - Existing tests exercise each rule through catalog lookup, detecting an omitted registration.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes:** `go test ./pkg/golint -count=1` passed, including rule, catalog-hygiene, atlas, and repository checks. Pinned-range `git diff --check` passed. Recorded integrated suite passed 1,877 cmd/sdlc tests and 75 other packages.

6. **Architectural notes**
   - **ARCH-DRY — pass:** retains shared matchers without duplication.
   - **ARCH-PURE — pass:** catalog resolution introduces no IO or business logic.
   - **ARCH-PURPOSE — pass:** preserves both sides’ enforcement and documentation.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.

   Atlas is updated; no new command, flag, or configuration surface requires README changes.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5a66117f-6e58-4b27-b47c-ba4b535fe18e","repository":"github.com/xianxu/ariadne","issue_id":"000498","recipe":"fix-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/054024344283c22c","reviewed_head":"18ef5d0ed97b172b1e1d02eb0b204b4adf76a475","slot":"ariadne:1","requested_at":"2026-10-21T03:17:00.101433Z"},"phase":"finished","pid":24819,"started_at":"2026-10-21T03:17:00.214Z","finished_at":"2026-10-21T03:17:56.749141Z","session_id":"01a15c76-51d2-7571-924d-f706c936353b","outcome":"returned"}] -->
