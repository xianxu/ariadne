# Boundary Review — ariadne#630 (milestone M2)

| field | value |
|-------|-------|
| issue | 630 — Shard timeouts with no failing test are 'incomplete', not failures: retry once, skip the main-baseline diagnosis |
| repo | ariadne |
| issue file | workshop/issues/000630-shard-timeouts-with-no-failing-test-are-incomplete-not-failures-retry-once-skip.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | f4ddeb648dd808b1c223a11d897840bbb79b7ad0..28d76ccae32151f528c30eb1636d03c5608435a6 |
| command | sdlc milestone-close --issue 630 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T13:45:59-07:00 |
| verdict | SHIP |

## Review

Reviewing the M2 diff; now checking the suite config and running the tests.

Targeted tests hung in a pre-existing train test; isolating the new tests from it now.

Both new guards go red when removed. Checking that retention handles the new version-5 suite links.

```verdict
verdict: SHIP
confidence: high
```

M2 does what the plan and the issue's Decisions promise. When a recorded suite run is incomplete, it retries once through the #561 resume composite. The proof takes the package list from the parent's own log and binds it in a version 5 suite link. The retry runs with the check's own weight and timeout (2 slots, 60 minutes for ariadne's suite), not a targeted rerun's single slot and 15 minutes. A clean retry posts green as `incomplete-retry-passed`. A retry cut short again is reported with the load and posts nothing, and a later run of the same head still posts. A retry that fails for real posts as an ordinary failure from the retry's log, which the merge gate and the train diagnose against main. A main baseline never retries.

BR-1 is fixed. ariadne's suite is `make test` → `python3 scripts/test-shard.py`, and that script now honors `SDLC_AFFECTED_GO_PKGS`, so the retry really runs only the selection.

All the #630 tests pass, the resume/saved/republish/known-flake set passes (294s), and the Python tests pass (15). I removed two of the new guards one at a time and confirmed a test fails each time. One pre-existing train test (`TestLandingTrainSplitHoldsTheCulpritAndLandsThePrefix`) hung under a broad parallel run but passes alone in 17.5s; this diff doesn't touch it. Nothing blocks SHIP. There is one Minor.

1. **Strengths**
   - `localcheck_resume_evidence.go:242-249`: the retry's package list comes from `classifyIncomplete(parentFacts)`, never from the link. The link must match it exactly (`:278`), so a forged, wider list is refused. The "a wider selection" and "a group" cases pin this.
   - `suiteincomplete.go:482` `retryPassed` is the single definition of a clean retry, shared by the child record and the composite verdict. It also closes the case of an exit-0 run that prints `--- FAIL`.
   - "No retry of a retry" is enforced by structure: the child runs through `runUnderSlotBatch`, not `runLocalCheck`. No flag is needed for it.
   - `TestLandingTrainIncompleteSuite` drives the real `conductor.runSuite` over `newLocalCheckRunner` and covers three branches: the retry passes, it is incomplete again, and it fails for real.
   - `scripts/test-shard.py` `select_packages` refuses a package outside the module instead of skipping it. It also leaves the test timings untouched when `cmd/sdlc` is not run.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `localcheck_resume_saved.go:349`: the code decides to point `LogPath` at the child's log by checking whether `rec.LogSHA256 == in.ChildLogSHA256`. That works today but is fragile. The code that builds the record already knows it took the failed-retry branch, so it should return that fact as an explicit flag.

5. **Test coverage notes**
   - Mutations I checked by hand:
     - removing the "a known-flake suite link must be version 3" check makes `TestKnownFlakeSuiteLinkStaysVersion3` fail;
     - removing `r.Version != l.Version` makes `TestResolveIncompleteResumeProof` fail on "a receipt version 3".
   - Already covered:
     - the envelope (the retry's deadline is 30m, not the 5m targeted cap);
     - a baseline never retries (no package-list marker file is written);
     - a later pass on the head after two incomplete runs is still posted.
   - Gap: no test exercises retention (`localcheck_resume_retention.go:141,190`) on a v5 parent with a v2 child. I read the code and the pairing is correct.

6. **Architectural notes**
   - ARCH-DRY: pass. The retry reuses `affectedEnvCommand`, `runUnderSlotBatch`, the resume link and `savedAttemptRecord`; there is no parallel path.
   - ARCH-PURE: pass. The verdicts (`resolveResumeProof`, `incompleteRetryOutcome`, `retryPassed`, `incompleteRetryCommands`) are pure and table-tested. The IO is the thin `resumeSaved` and `affectedCommandOnMain`.
   - ARCH-PURPOSE: pass. All three post-retry branches are wired to the merge gate and the train.
   - ARCH-MOCK: pass. The tests use the existing `ciFakeGH` and git fixtures through the production seams.
   - ARCH-CONSTRAINTS: pass. The retry takes the declared weight and timeout of the parent check. In practice an alarm in `cmd/sdlc` reruns all of `cmd/sdlc`'s shards, which is what the TL decided.
   - ARCH-SECURE: pass. The link is untrusted input, and the package list is re-derived from the digest-bound parent log.
   - ARCH-ORDER: pass. State moves through `resumeAdvance` (eligible → … → child published). An interrupted retry is handled by the existing reconciliation receipt, which now carries the link's version.
   - ARCH-FUNERAL: pass. The child attempts and links fall under the existing resume retention, and nothing new is created.
   - ARCH-EVOLVE: pass. A suite link that carries a package list is version 5, so an older sdlc treats it as unsupported proof for that head only. `TestResumeEnvelopeVersionsByCheck` pins the admitted version matrix.
   - Docs: the atlas is updated (`atlas/workflow/ci-merge-check.md`). README needs no change, since there are no new flags or subcommands and `SDLC_AFFECTED_GO_PKGS` is an existing contract.

7. **Plan revision recommendations:** none. The "M2 as built" Revisions entry matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      scripts/test-shard.py now honors SDLC_AFFECTED_GO_PKGS (select_packages/build_jobs, PackageSelectionTest); ariadne's suite is make test -> test-shard.py, so the retry runs only the selection (TestIncompleteRetryRunsOnlyItsPackagesAndLands asserts the list is exactly o/r/pkg).
findings:
  - id: new
    severity: Minor
    kind: defect
    family: state-inferred-from-coincidence
    title: |
      resolvedResumeRecords picks the child's log by comparing digests rather than by an explicit flag
    detail: |
      localcheck_resume_saved.go:349 sets LogPath to the child's log when rec.LogSHA256 equals ChildLogSHA256. resolvedSavedRecord already knows it took the failed-retry branch; it should return that as an explicit flag instead of the caller inferring it from digest equality.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"51ad9ff9-8cd3-4249-8920-8a318c6f0cd1","repository":"github.com/xianxu/ariadne","issue_id":"000630","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/025820607d37a94e","reviewed_head":"28d76ccae32151f528c30eb1636d03c5608435a6","slot":"ariadne:5","requested_at":"2026-10-25T20:26:56.719699Z"},"phase":"finished","pid":49535,"started_at":"2026-10-25T20:26:56.899879Z","finished_at":"2026-10-25T20:45:57.948925Z","session_id":"a1095458-d877-4e93-8018-f5856f17f25c","outcome":"returned"}] -->
