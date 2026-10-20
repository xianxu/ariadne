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
