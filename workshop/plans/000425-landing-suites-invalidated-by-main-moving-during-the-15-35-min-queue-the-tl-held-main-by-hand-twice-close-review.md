# Boundary Review — ariadne#425 (whole-issue close)

| field | value |
|-------|-------|
| issue | 425 — Landing suites invalidated by main moving during the 15–35 min queue; the TL held main by hand twice |
| repo | ariadne |
| issue file | workshop/issues/000425-landing-suites-invalidated-by-main-moving-during-the-15-35-min-queue-the-tl-held-main-by-hand-twice.md |
| boundary | whole-issue close |
| milestone | — |
| window | 146796d884ccf9a9d589a8766d7627762b4571e6..7d30d9a34392db47163f8225152dc4d1a4a31b25 |
| command | sdlc close --issue 425 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T07:10:59-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The existing landing mechanism supports the approved workflow: concurrent landings serialize integration, and docs-only main publication permits safe suite reuse. All new tests passed, including the concurrent test under the race detector. Two Important gaps remain: README does not document the changed usage sequence, and regression checks do not protect the embedded instructions that deliver this issue’s behavior change.

```findings
findings:
  - id: new
    severity: Important
    family: user-facing-workflow-documentation
    title: |
      README update appears missing for the changed pre-close and landing sequence
    detail: |
      cmd/sdlc/helptext/local-check.md:63–68 changes the user-facing procedure to close the tested branch without chasing newer main, then let sdlc merge integrate under admission. README.md:466–474 remains unchanged and omits that distinction. Family enumeration: the affected README passages are landing-line usage at lines 444–452 and pre-close suite usage at lines 466–474. Update these passages with the approved sequence, preserving freshness and branch-fix verification requirements.
  - id: new
    severity: Important
    family: executable-guidance-regression-coverage
    title: |
      The changed embedded workflow instructions have no regression checks
    detail: |
      The production change is executable agent guidance, but all added tests exercise unchanged merge behavior. Existing help checks do not assert the new phase ownership or reject the removed instruction to merge main manually before landing. Family enumeration: cmd/tl-worker/help.md:34–40,83–85; cmd/tl-lead/help.md:327–333; cmd/sdlc/helptext/local-check.md:63–68. Extend rendered-help contract checks to protect the branch-suite → close → merge sequence, prohibition on manual chasing/holding, and freshness/code-fix exceptions; demonstrate that restoring the previous guidance fails those checks.
```

1. **Strengths**

   - The two-landing test runs genuine merge invocations against a shared bare remote and checks admission order, tested commit ancestry, and merged-head identity.
   - Docs-only drift coverage uses `TrunkFile.UpdateMany`, then verifies both incorporation of the publication and inheritance of the original suite record.
   - Adversarial reuse tests cover cumulative code, configuration, and embedded-help changes followed by prose.
   - Guidance retains the noon freshness gate and verification after branch code fixes.

2. **Critical findings**

   None.

3. **Important findings**

   The two findings above require a README update and focused regression coverage for the shipped instructions.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Three new tests passed: `go test ./cmd/sdlc -run '^(TestLandingLineTwoLandingsIntegrateInTheirOwnTurns|TestLandingReusesSuiteWhenProsePublishedDuringRun|TestLocalRecordsOnAncestorsRejectsNonProseDrift)$'`.
   - Concurrent landing test passed under `-race`.
   - Worker and lead help/instruction checks passed.
   - `git diff --check` passed. The full suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY: pass.** Reuses existing landing ownership, Git fixtures, CI seams, and code-surface classification.
   - **ARCH-PURE: pass.** Adds integration characterization through existing injected seams; no production business logic is mixed into IO.
   - **ARCH-PURPOSE: pass.** Implements the approved workflow correction and tests both concurrent landing and docs-publication cases. No additional reservation mechanism is needed for that revised scope.

7. **Plan revision recommendations**

   Append a `## Revisions` entry adding README synchronization and executable-guidance regression checks. Also correct the planned test location from `landingqueue_test.go` to `landingeconomy_test.go` and describe the actual readiness/release file protocol.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8adc1145-9547-4852-903b-1cee43ff154f","repository":"github.com/xianxu/ariadne","issue_id":"000425","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/626dcca66105ab49","reviewed_head":"7d30d9a34392db47163f8225152dc4d1a4a31b25","slot":"ariadne:5","requested_at":"2026-10-17T14:08:30.996756Z"},"phase":"finished","pid":39858,"started_at":"2026-10-17T14:08:31.181331Z","finished_at":"2026-10-17T14:10:58.80651Z","session_id":"01a14a31-5dd6-7133-84d6-57bd37e5723a","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T07:56:33-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 425 — Landing suites invalidated by main moving during the 15–35 min queue; the TL held main by hand twice |
| repo | ariadne |
| issue file | workshop/issues/000425-landing-suites-invalidated-by-main-moving-during-the-15-35-min-queue-the-tl-held-main-by-hand-twice.md |
| boundary | whole-issue close |
| milestone | — |
| window | 146796d884ccf9a9d589a8766d7627762b4571e6..40142a581f1c75bc04c7af76240f45ce433d5db6 |
| command | sdlc close --issue 425 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T07:56:33-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range delivers the approved workflow correction and covers concurrent landings, docs-only publication, and unsafe suite reuse. Both prior findings are addressed. No new blocking findings were found.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:444–449 distinguishes branch verification from admitted integration; lines 468–479 preserve freshness, branch-fix verification, and conservative reuse. These match landing.go admission/gating and cigate.go ancestor-reuse behavior.
  - id: BR-2
    disposition: addressed
    note: |
      Rendered-help checks cover worker, lead, and local-check guidance. All pass at pinned head and all fail when the three help files are restored to base versions in a temporary copy, removed after verification.
```

1. **Strengths**

   - Two real merge invocations verify admission order, integrated ancestry, and tested-head identity.
   - Docs-publication coverage uses the actual `TrunkFile.UpdateMany` transaction.
   - Cumulative code, configuration, and embedded-help changes correctly invalidate reuse.
   - README, atlas, and embedded guidance consistently preserve freshness and branch-fix requirements.

2. **Critical findings:** None.

3. **Important findings:** None remaining.

4. **Minor findings:** None.

5. **Test coverage notes**

   - All six new tests passed.
   - Both concurrent integration scenarios passed under `-race`.
   - Broader rendered-help and flag checks passed.
   - Old-guidance mutation confirmed all three regression checks fail.
   - `git diff --check` passed. Full suite was not rerun during review.

6. **Architectural notes**

   - **ARCH-DRY: pass.** Reuses existing admission, fixtures, CI seams, and classifier.
   - **ARCH-PURE: pass.** Integration coverage uses existing injected boundaries; no production logic gains IO coupling.
   - **ARCH-PURPOSE: pass.** Delivers the approved measured scope without adding another reservation mechanism.

7. **Plan revision recommendations:** None. Existing revisions document the corrected test location, synchronization protocol, README updates, and guidance regression checks.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"921afc49-2378-48e2-b1c1-3c556afe97d3","repository":"github.com/xianxu/ariadne","issue_id":"000425","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a74746d19fbbafd6","reviewed_head":"40142a581f1c75bc04c7af76240f45ce433d5db6","slot":"ariadne:5","requested_at":"2026-10-17T14:55:11.582659Z"},"phase":"finished","pid":30153,"started_at":"2026-10-17T14:55:11.683823Z","finished_at":"2026-10-17T14:56:32.691828Z","session_id":"01a14a5c-1823-7da1-a011-c39f6de459e6","outcome":"returned"}] -->
