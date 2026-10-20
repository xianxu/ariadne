# Boundary Review — ariadne#494 (whole-issue close)

| field | value |
|-------|-------|
| issue | 494 — Train car's own merge warns 'active issue missing without confirmed archive' after the train already archived it |
| repo | ariadne |
| issue file | workshop/issues/000494-train-car-s-own-merge-warns-active-issue-missing-without-confirmed-archive-after-the-train-already-archived-it.md |
| boundary | whole-issue close |
| milestone | — |
| window | dc76211e27578aca7aa62d434db9ca2bd43eed5b..56a0a9c84276d7bb914673cec39fc6983424f55d |
| command | sdlc close --issue 494 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T06:57:47-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned change fulfills the revised Spec: a validated train outcome enables archive discovery, while the shared verifier supplies the archive proof. Both focused test runs passed. No blocking findings; the separately required recorded-suite pass remains a landing prerequisite.

```findings
{}
```

1. **Strengths**
   - Retains the validated waiter outcome through cleanup (`cmd/sdlc/landing.go:420`).
   - Reuses the complete archive-transition verifier across both provenance paths (`cmd/sdlc/landingarchive.go:731`).
   - Exercises two real merge callers with deterministic cleanup ordering, including output, resting-branch preservation, and ref cleanup (`cmd/sdlc/landingtrain_archive_test.go:21`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Passed archive proof controls and the two-car regression (77.200s); passed no-outcome recovery and tracked serial landing checks across merge, squash, and rebase (53.289s). `git diff --check` passed. The full recorded suite was not run during this review.

6. **Architecture:** ARCH-DRY passes through verifier reuse. ARCH-PURE passes: history lookup remains in the existing injectable integration boundary. ARCH-PURPOSE passes: the actual waiter failure and safe cleanup are exercised. ARCH-EVOLVE passes: no persisted format changes. The atlas documents the flow; no new runnable surface requires a README change.

7. **Plan revisions:** None required.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3de28267-02fc-40b6-88b1-b4c7a977d59e","repository":"github.com/xianxu/ariadne","issue_id":"000494","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/290a1e60eb02aaa2","reviewed_head":"56a0a9c84276d7bb914673cec39fc6983424f55d","slot":"ariadne:1","requested_at":"2026-10-20T13:54:47.73622Z"},"phase":"finished","pid":11561,"started_at":"2026-10-20T13:54:47.903612Z","finished_at":"2026-10-20T13:57:45.832965Z","session_id":"01a15997-e170-7552-803f-7f29c8b22307","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T13:05:56-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 494 — Train car's own merge warns 'active issue missing without confirmed archive' after the train already archived it |
| repo | ariadne |
| issue file | workshop/issues/000494-train-car-s-own-merge-warns-active-issue-missing-without-confirmed-archive-after-the-train-already-archived-it.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7cb5a0a481b915a3bb48cf997d284b3a2566bfa5..058f4f15587adf7659fed2095b9b8f9750a1de02 |
| command | sdlc close --issue 494 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T13:05:56-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff correctly stabilizes the positive-outcome regression without changing production behavior or expanding the Spec. No prior findings required disposition, and no new defects were found. The corrected fixture still detects the original archive failure.

```findings
{}
```

1. **Strengths**
   - Outcome observation now controls listing visibility, closing the readiness race (`cmd/sdlc/landingtrain_archive_test.go:122`).
   - The wrapper reuses the stateful fake and changes returned copies, preserving actual merge/archive effects (`cmd/sdlc/landingtrain_archive_test.go:202`).
   - Existing assertions retain archive-output, resting-branch, and ref-cleanup coverage.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage**
   - Focused regression, archive-proof controls, and no-outcome recovery passed: 33.034s.
   - Three race-enabled regression repetitions passed: 38.439s.
   - Disabling the production archive fallback through a temporary Go overlay made the corrected regression fail with the exact original missing-archive error after observing the landed outcome.
   - `git diff --check` passed. Full suite not rerun during this narrow review.

6. **Architecture**
   - **ARCH-DRY — pass:** reuses the existing fake and synchronized writer.
   - **ARCH-PURE — pass:** ordering remains within the injected integration fixture.
   - **ARCH-PURPOSE — pass:** makes the intended interleaving deterministic while retaining independent no-outcome coverage.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.
   - No new surface requires atlas or README updates.

7. **Plan revision recommendations:** None; Spec, Plan, and Done-when remain unchanged.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"78aae31f-417c-4df2-a0da-364eb7144f9a","repository":"github.com/xianxu/ariadne","issue_id":"000494","recipe":"fix-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/dd29bca80050a876","reviewed_head":"058f4f15587adf7659fed2095b9b8f9750a1de02","slot":"ariadne:1","requested_at":"2026-10-20T20:03:44.821981Z"},"phase":"finished","pid":11322,"started_at":"2026-10-20T20:03:44.999779Z","finished_at":"2026-10-20T20:05:55.456497Z","session_id":"01a15ae9-aaab-72e3-845d-46be382b4fe1","outcome":"returned"}] -->
