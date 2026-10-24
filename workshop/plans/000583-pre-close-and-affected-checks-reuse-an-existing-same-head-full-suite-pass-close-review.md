# Boundary Review — ariadne#583 (whole-issue close)

| field | value |
|-------|-------|
| issue | 583 — Pre-close and affected checks reuse an existing same-head full-suite pass instead of queueing a new full run |
| repo | ariadne |
| issue file | workshop/issues/000583-pre-close-and-affected-checks-reuse-an-existing-same-head-full-suite-pass.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa9b49183e8924fb5b034678c696dd8a1ddedd26..8f64a172590ab44dcb07a1f059e4183c80571608 |
| command | sdlc close --issue 583 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T20:32:43-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking boarding semantics, then I'll run the new tests.

Now reverting the fix to confirm the test goes red.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change does what the issue asks. `runLocalCheckCmd` now calls the existing `republishSavedPass` (#431) before it plans or runs anything. That means a reuse inherits the same safeguards merge already uses: the attempt must be the newest one, its log hash must match, known-flake passes are rechecked, and the #394 guard still runs. Lookups are keyed to the exact sha, so a pass on another head never counts. I ran the three relevant tests and they pass. I also put the pre-#583 `localcheck_cmd.go` back and reran `TestPreCloseReusesASameHeadSuitePass`: it fails, and the `--pre-close` subcase posts a new `sdlc-local/affected:success`, which proves the test catches the bug this issue fixes. What stops a plain SHIP is test coverage. The reuse condition has two branches the help text describes, and neither is tested: `--pre-close` resolving to `suite` (a repo without trains), and a plain `--name suite` that must still rerun. Both tests are cheap to add.

1. **Strengths**
   - `cmd/sdlc/localcheck_cmd.go:240`: the fix calls the existing `republishSavedPass` instead of writing a second "saved pass" check (ARCH-DRY pass). It also returns the error when `reused && err != nil`, so a guard refusal or a "needs newer sdlc" refusal is never followed by a blind rerun. That matches what merge does.
   - The condition `wantAffected && len(names) == 0` correctly leaves out `--all`, where `splitAffectedName` gives `wantAffected=false`, and any mix like `--name affected,lint`. Neither case can skip other declared checks.
   - `cmd/sdlc/precloseck_test.go:113`: the tests check real effects, not just output. They count saved suite attempts in the store (`suiteRuns`), look for the absence of an `sdlc-local/affected` status, and cover a different head and a failed saved suite. The failed-suite case also checks that the #394 guard still refuses.
   - The close help change is pinned by `TestCloseHelpNamesThePreCloseCheck`, which checks for the stale phrases as well as the new one.

2. **Critical findings:** none.

3. **Important findings**
   - **Untested branches of the reuse condition** (`cmd/sdlc/localcheck_cmd.go:240`; family `done-when-mode-coverage`). The test fixture always turns trains on and declares `affected.json`, so `--pre-close` always resolves to `affected`. These cases in the same family are never run:
     - (a) `--pre-close` resolving to `suite`, where `names=["suite"]` and only the `f.PreClose` disjunct allows reuse;
     - (b) plain `--name suite` with a same-head pass, which must rerun, as both the help text (`local-check.md:267`) and the code comment say;
     - (c) `--boarding --name affected` reuse, which the plan names ("boarding included");
     - (d) a reused known-flake-qualified pass, which the Spec names explicitly.

     Fix: in `TestPreCloseReusesASameHeadSuitePass`, add a fixture without trains (or without `affected.json`) where `--pre-close` reuses. Assert that `--name suite` after a pass produces 2 suite attempts. Add `{Boarding: true, Names: ["affected"]}` to the reuse loop. Case (d) can rely on the existing #431 revalidation tests, but name that reliance in the issue Log.

4. **Minor findings**
   - In a `--pre-close` reuse, stderr first prints `==> local-check --pre-close: running affected (...)` (`localcheck_cmd.go:171`) and then "nothing to run". The output contradicts itself; consider "resolved to" wording.
   - The reuse message joins `f.Names`, so a pre-close run reports `local-check affected on …`, not `--pre-close`. Cosmetic.

5. **Test coverage notes:** the revert check succeeded: the new test fails without the fix, at the `--pre-close` subcase. The regression evidence is real. The gaps are the mode coverage listed above.

6. **Architectural notes**
   - ARCH-DRY: pass. It reuses `republishSavedPass`, and no logic is duplicated.
   - ARCH-PURE: pass. The new code is a short decision in the CLI glue that hands off to an existing helper. The condition could become a small pure predicate (`reusesSuitePass(preClose, wantAffected, names)`) if it grows.
   - ARCH-PURPOSE: pass. Both `--pre-close` and `--name affected` use the saved pass, and the help text and atlas `ci-merge-check.md` are updated in the same range.
   - ARCH-EVOLVE: pass. No persisted format changes.
   - README: no new flag or subcommand, so no update is needed.

7. **Plan revision recommendations:** none. The code matches the Plan's checklist. If (d) is left to the existing #431 tests, add a Log line saying so.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: done-when-mode-coverage
    title: |
      Reuse condition branches untested: pre-close resolving to suite, plain --name suite rerun, boarding reuse
    detail: |
      cmd/sdlc/localcheck_cmd.go:240 reuses when f.PreClose or (wantAffected and no other names). The fixture always turns trains on and declares affected.json, so pre-close always resolves to affected. Never run in this family: (a) pre-close resolving to suite, where only the f.PreClose disjunct allows reuse; (b) plain --name suite with a same-head pass must rerun (claimed at local-check.md:267 and in the code comment); (c) --boarding --name affected reuse (named in the Plan); (d) reuse of a known-flake-qualified pass (named in the Spec). Add a fixture without trains asserting pre-close reuse, assert 2 suite attempts after a second --name suite, add the boarding flags to the reuse loop, and cite or add known-flake reuse coverage.
  - id: new
    severity: Minor
    kind: defect
    family: reuse-output-wording
    title: |
      Pre-close reuse prints "running affected" and then "nothing to run"
    detail: |
      localcheck_cmd.go:171 announces running the resolved check before the reuse short-circuit at :240. The message at :246 also names f.Names rather than --pre-close. Reword the first line as "resolved to", or print it after the reuse check.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3b13a1d9-17ac-4ee7-bbfa-949bce702227","repository":"github.com/xianxu/ariadne","issue_id":"000583","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/6fa214c240f6e3a9","reviewed_head":"8f64a172590ab44dcb07a1f059e4183c80571608","slot":"ariadne:2","requested_at":"2026-10-24T03:31:13.978277Z"},"phase":"finished","pid":53625,"started_at":"2026-10-24T03:31:14.117277Z","finished_at":"2026-10-24T03:32:42.918145Z","session_id":"07a3ef5e-5397-419d-8683-8b623e5bb8cb","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T20:35:51-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 583 — Pre-close and affected checks reuse an existing same-head full-suite pass instead of queueing a new full run |
| repo | ariadne |
| issue file | workshop/issues/000583-pre-close-and-affected-checks-reuse-an-existing-same-head-full-suite-pass.md |
| boundary | whole-issue close |
| milestone | — |
| window | fa9b49183e8924fb5b034678c696dd8a1ddedd26..d4a79fa2665c5d708e3807f554b27fa0202c32b7 |
| command | sdlc close --issue 583 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T20:35:51-07:00 |
| verdict | SHIP |

## Review

Both prior findings are fixed and I'm finishing the verdict.

```verdict
verdict: SHIP
confidence: high
```

Both findings from round 1 are fixed, and I checked each fix by breaking the code and watching a test fail. The change does what #583 asks. Before running anything, `runLocalCheckCmd` now asks the existing `republishSavedPass` (#431) for a saved, verified suite pass on the exact head. It asks only for `--pre-close` and for runs that name `affected` alone (with or without `--boarding`). When it finds one, it posts that pass and runs nothing. A plain `--name suite` still reruns. A run that names `affected` together with another check, or uses `--all`, also behaves as before, because `names` is non-empty or `wantAffected` is false. The targeted tests pass at the head. Nothing blocks shipping.

**What I ran:**
- **The new tests at the head:** `TestPreCloseReusesASameHeadSuitePass`, `TestLocalCheckCmdPreClose` and `TestCloseHelpNamesThePreCloseCheck` all pass.
- **Removing the `f.PreClose ||` part of the condition** (`localcheck_cmd.go:240`): the `pre-close resolving to suite reuses` subtest fails.
- **Reducing the condition to bare `ok`, so every run reuses:** the subtest fails with "1 suite attempts after an explicit --name suite, want 2".

I restored the code after each change, and the checkout is clean.

1. **Strengths**
   - It reuses the existing code instead of adding new logic (ARCH-DRY). `republishSavedPass` already checks the saved log hash, the host, that the pass is the newest attempt, and known-flake status. The CLI picks all of that up and shares one source of truth with merge's runner.
   - The reuse condition (`localcheck_cmd.go:240`) is narrow: `--all`, a mix like `affected,lint`, and a plain `suite` request all skip the reuse path.
   - Errors from republishing are passed back to the caller. Unsupported or unverifiable evidence and an unexplained failure on the head are never turned into a quiet rerun or a quiet success.
   - The tests use a real git fixture and count real suite attempts in test-stats, so they check behaviour rather than mocks.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none new.

5. **Test coverage notes**
   - Each mode in the Done-when clause is covered both ways:
     - **Reuse:** `--pre-close` resolving to affected, `--pre-close` resolving to suite, `--name affected`, and `--name affected` with `--boarding`.
     - **No reuse:** another head (via `--pre-close`) and a failed saved suite (via `--name affected`).
   - Reuse of a known-flake pass is covered at the function level in `localcheck_flakes_test.go:255`. That test goes through the same `republishSavedPass`, which is a fair citation.

6. **Architecture**
   - **ARCH-DRY:** pass. It reuses `republishSavedPass` and duplicates nothing.
   - **ARCH-PURE:** pass. The change is a 16-line guard in the existing command layer, with no new logic mixed into the IO code.
   - **ARCH-PURPOSE:** pass. Every path named in the Spec and Plan is delivered and tested, and the help, atlas and close prose were updated in the same change.
   - **ARCH-EVOLVE:** pass. No stored format changes.

7. **Plan revisions:** none needed. The plan matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      precloseck_test.go now covers pre-close resolving to suite (no trains), the plain --name suite rerun (2 attempts) and boarding reuse; known-flake reuse is cited to localcheck_flakes_test.go:255. Removing the f.PreClose disjunct, or widening the condition to bare ok, turns the test red.
  - id: BR-2
    disposition: addressed
    note: |
      localcheck_cmd.go:171 now prints "resolved to X" (pinned at precloseck_test.go:108), and the reuse line at :246 names --pre-close when it applies.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"481e473a-78d1-4206-9637-b74f56dc2a55","repository":"github.com/xianxu/ariadne","issue_id":"000583","recipe":"small-diff-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/96233add20fe1add","reviewed_head":"d4a79fa2665c5d708e3807f554b27fa0202c32b7","slot":"ariadne:2","requested_at":"2026-10-24T03:34:30.408878Z"},"phase":"finished","pid":7675,"started_at":"2026-10-24T03:34:30.512238Z","finished_at":"2026-10-24T03:35:50.87265Z","session_id":"78f7ae90-86b1-44a0-aa89-bb5b0c7adf61","outcome":"returned"}] -->
