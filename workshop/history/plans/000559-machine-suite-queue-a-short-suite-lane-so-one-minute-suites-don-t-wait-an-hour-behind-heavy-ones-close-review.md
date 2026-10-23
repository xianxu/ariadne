# Boundary Review — ariadne#559 (whole-issue close)

| field | value |
|-------|-------|
| issue | 559 — Machine suite queue: a short-suite lane so one-minute suites don't wait an hour behind heavy ones |
| repo | ariadne |
| issue file | workshop/issues/000559-machine-suite-queue-a-short-suite-lane-so-one-minute-suites-don-t-wait-an-hour-behind-heavy-ones.md |
| boundary | whole-issue close |
| milestone | — |
| window | a561b0aba633d3103cbf9343c08117b8c0a1fab3..43ebab480fae890e82ce4523aca1699dd86de858 |
| command | sdlc close --issue 559 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T08:33:07-07:00 |
| verdict | REWORK |

## Review

Reviewing the helper's admission gate and the Go run path for the light lane; a few more checks to go.

Checking how attempt decoding treats unknown fields (backward-compat for the new `lane` keys).

Running the new targeted Go tests and the helper shell test.

```verdict
verdict: REWORK
confidence: medium
```

The lane is mostly sound, but one bug blocks the boundary. The helper keeps the light line separate from the normal line in its structure: light waiters only ever see `fds=[light]` and their own `Tickets` directory, so heavy admission cannot change. Eligibility is pure, measured from recorded attempts, and covered by table tests. Status rendering and the atlas both match the code. All targeted Go tests pass at the pinned head, and so do the 16 lane checks in `scripts/test/test-lock.test.sh`, which I ran here. The blocker is a repo that declares typed flake runners (local-checks v3, `flakeContract != ""`). For that repo, a light-lane overrun does not retry in the normal line. Instead it surfaces as an unconfirmed-cleanup error and records no attempt. The repo therefore stays eligible, and every later run repeats the 240 s overrun and fails. That breaks Done-when 2 for that class of repo. The plan also no longer matches the code in two places: how the overrun is recorded, and how the capability is probed.

### 1. Strengths
- **Light line kept apart by construction** (`scripts/test-lock.py:1127-1129`, `elect_light`). A light waiter never scans or takes a normal slot or ticket. The shell test checks this end to end: a heavy waiter stays at position 1 while light waiters queue, then takes both slots while the light slot is held.
- **Re-entrancy covers the light holder.** `every = fds + [light]` is used for the `nested()` scan, so a `make test` nested under a light run goes re-entrant. This is tested.
- **Eligibility is pure and measured.** `laneEligibility` (`cmd/sdlc/suitelane.go:40`) handles recency ordering, overrun or timeout disqualification, and baseline or legacy filtering. Its tests use literal 3 and 120 values as an independent oracle.
- **The state machine is tested in full.** `nextSuiteLaneStep` is checked over the complete states × events grid, including rejected events and the "overrun never returns to light" invariant.
- **Strict decoding of the new fields** (`teststats_store.go:976`). The decoder rejects an unknown lane value, a zero or null overrun, and case-aliased keys, and each rejection is tested.

### 2. Critical findings
- **`cmd/sdlc/localcheck_run.go:82` (`lightLaneFor`) together with `runUnderResourceBatch`'s deferred `finishResumeScratch`.** `bindTypedCheck` sets `StrictGroupCleanup = true` for any typed-runner suite.
  - When the light deadline fires, `run.TimedOut = true`.
  - The deferred `finishResumeScratch` (`localcheck_resume.go:362`) then returns `stop(nil)`, which is a non-nil `runnerCleanupUnconfirmed`, and it sets `runnerLifetime.retirable = false`.
  - `runSuiteLanes` returns on that error with no retry, and `runLocalCheck` takes the "uncertain" path, so no attempt or `lane_overrun_seconds` is recorded.
  - The repo stays eligible and loops: each run overruns for 240 s and fails. The `runnerLifetime` pointer is also shared through `dl := d`, so even a retry would inherit `retirable=false`.
  - **Fix sketch:** exclude `d.flakeContract != ""` (or `StrictGroupCleanup`) in `lightLaneFor`, or give lane overrun its own cleanup confirmation. Add a test of a typed-runner suite under the light lane that overruns, asserting either "never asks light" or "retries and records".

### 3. Important findings
- **The plan no longer matches the code (ARCH-PURPOSE / traceability).** The plan's Core concepts and run-order table say three things the code does not do:
  - The overrun is recorded as a separate attempt with outcome `lane_overrun`. The code records no light attempt; it puts `lane_overrun_seconds` on the retry's attempt instead.
  - The helper advertises `lane-v1` "through the same channel #563 uses". The code probes for `--lane` in `acquire --help`.
  - The plan names `suiteLaneEligibility` as the function. The code's function is `laneEligibility`.

  The atlas describes the code correctly. The plan's `## Revisions` has no entry for any of these.

### 4. Minor findings
- If the normal retry errors after a light overrun (paused, interrupted, or a startup error), the overrun length is dropped. No record carries it, so the repo stays eligible (`suitelane.go:631`).
- The recorded `queue_seconds` and `wall_seconds` cover only the retry. The light run's queue wait and its partial log are lost, because `os.Create` truncates the log on retry.
- `TestRunLocalCheckLight*` uses a stateless fake that logs argv and answers `acquired-light`. The plan said "the fake is the real helper". The real helper is exercised elsewhere (the shell test and `TestSuiteQueueRealHelperShowsLightLane`), so this is a small gap (ARCH-MOCK).
- Plan Task 2's "a dead light waiter's ticket is pruned" has no direct test. The shell test only checks cleanup after a normal release.
- `inspect_light` takes the free light slot's flock for an instant, as do non-head `scan(fds,0)` polls. The head light waiter may miss a poll. This is documented and harmless.

### 5. Test coverage notes
- `TestLaneEligibility`, `TestNextSuiteLaneStep`, `TestSuiteAttemptLaneFieldsRoundTrip`, `TestRunLocalCheckLight*` and `TestSuiteQueue*Lane*` all pass. The shell test reports "all passed".
- The missing test is the class behind the Critical finding: a light run combined with the existing typed or strict-cleanup paths.
- The round-trip test covers the known fields only. An attempt is never rewritten, so preserving unknown members is not exercised, but nothing loses them either.

### 6. Architecture
| Marker | Result | Notes |
|---|---|---|
| ARCH-DRY | pass | Reuses `Tickets`, `scan`, `inspection_snapshot` and `validateSuiteTickets`. |
| ARCH-PURE | pass | Eligibility and step are pure; the IO stays in the run closure. |
| ARCH-PURPOSE | flag | Done-when 2 fails for typed-runner repos (Critical). |
| ARCH-MOCK | minor | See Minor findings. |
| ARCH-CONSTRAINTS | pass | Heavy impact was measured: 951 s against a 1003 s bound. The load was un-niced, which is harsher than the real lane. |
| ARCH-SECURE | pass | Holder and waiter records are rendered with `%q`, and inspection is bounded. |
| ARCH-ORDER | partial | The step function is used, but cancel events never reach it at runtime (errors short-circuit). The overrun-with-strict-cleanup interleaving is unmodelled. |
| ARCH-FUNERAL | pass | One `.light` file, plus tickets removed on release or prune. |
| ARCH-EVOLVE | pass | Additive fields; older readers tolerate unknown keys; strict known-field validation. |

### 7. Plan revision recommendations
- `## Revisions`: "Overrun evidence is `lane_overrun_seconds` on the retrying attempt (no separate `lane_overrun` outcome or light attempt); the capability probe is `--lane` in `acquire --help`, not a `lane-v1` token; the eligibility function is `laneEligibility`."
- `## Revisions`: "Typed-runner (StrictGroupCleanup) suites are excluded from the light lane, or their overrun is confirmed and retried; see BR-n."

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: lane-overrun-strict-cleanup-path
    title: |
      A light-lane overrun in a typed-runner suite errors out without retrying or recording, so the repo loops in the lane
    detail: |
      bindTypedCheck sets StrictGroupCleanup. On the light deadline run.TimedOut is true, so the deferred finishResumeScratch (localcheck_resume.go:362) returns a non-nil runnerCleanupUnconfirmed and clears retirable on the shared runnerLifetime. runSuiteLanes then returns that error with no normal retry, no attempt and no lane_overrun_seconds, so laneEligibility stays true and every later run overruns again. Fix: exclude flakeContract/StrictGroupCleanup in lightLaneFor (localcheck_run.go:83), or confirm and retry the overrun, plus a regression test.
  - id: new
    severity: Important
    kind: defect
    family: plan-code-drift
    title: |
      The plan's overrun recording, capability probe and eligibility name no longer match the code; no Revisions entry
    detail: |
      The plan says a separate attempt is recorded with outcome lane_overrun, the helper advertises lane-v1, and the function is suiteLaneEligibility. The code puts lane_overrun_seconds on the retry's attempt, probes for --lane in acquire --help, and names the function laneEligibility. The atlas matches the code; the plan needs a dated Revisions entry.
  - id: new
    severity: Minor
    kind: defect
    family: lane-overrun-evidence-loss
    title: |
      The overrun length is dropped when the normal retry errors, and the light phase's queue wait and log are not kept
    detail: |
      runSuiteLanes returns the error with the overrun only in memory (suitelane.go:631). A paused or interrupted retry leaves the repo eligible. os.Create truncates the light run's log on retry.
  - id: new
    severity: Minor
    kind: defect
    family: plan-test-commitments
    title: |
      Plan-promised tests missing: dead light waiter pruning, and the run loop exercised against the real helper
    detail: |
      The shell test checks ticket cleanup only after a normal release. TestRunLocalCheckLight* uses a stateless argv-logging fake helper rather than the real helper the plan names.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f50cf827-dee8-4196-bedb-d0eede06b58b","repository":"github.com/xianxu/ariadne","issue_id":"000559","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/46c59f09c10a22d1","reviewed_head":"43ebab480fae890e82ce4523aca1699dd86de858","slot":"ariadne:5","requested_at":"2026-10-23T15:29:33.805137Z"},"phase":"finished","pid":11490,"started_at":"2026-10-23T15:29:33.909432Z","finished_at":"2026-10-23T15:33:07.086281Z","session_id":"6227478e-b7cb-4b6b-b305-10a1afe5c288","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T08:51:05-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 559 — Machine suite queue: a short-suite lane so one-minute suites don't wait an hour behind heavy ones |
| repo | ariadne |
| issue file | workshop/issues/000559-machine-suite-queue-a-short-suite-lane-so-one-minute-suites-don-t-wait-an-hour-behind-heavy-ones.md |
| boundary | whole-issue close |
| milestone | — |
| window | a561b0aba633d3103cbf9343c08117b8c0a1fab3..c021d115ae1ce19f3afa780de8c71b9659cdfe16 |
| command | sdlc close --issue 559 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T08:51:05-07:00 |
| verdict | REWORK |

## Review

Checking the BR-1 fix now: I'll run the new tests, then revert the guard to confirm the regression test fails without it.

The new log-keeping test fails deterministically at head: publish deletes the inflight directory along with the `.light` log. Next I'll run the full package and the helper shell test.

BR-1's regression test goes red without the guard, so it's confirmed fixed. Last, I'm checking that the atlas and plan wording match the code.

```verdict
verdict: REWORK
confidence: high
```

The BR-1 fix is correct and backed by a real test. With the guard removed, `TestLightLaneForSkipsTypedRunnerSuites` fails ("a typed-runner suite asked for the light lane"). With the guard restored, it passes. BR-2 and BR-4 are also done. One thing blocks SHIP: the fix for the second half of BR-3, keeping the stopped light run's log, does not work, and its own test fails every time at the pinned head. `TestRunLocalCheckLightOverrunRetriesInNormalLine` fails at `suitelane_run_test.go:118` with "the light run's log was not kept: []". The repo's `suite` check would therefore fail on this head. The plan's new Revisions entry also says the log is kept, which the code does not do.

**1. Strengths**
- `cmd/sdlc/localcheck_run.go:167`: typed-runner suites never ask for the lane. This closes the loop BR-1 described, and the regression test is real (checked by removing the guard).
- `cmd/sdlc/suitelane.go:126`: when the retry cannot run, its error now names the overrun ("stopped the suite at 240s"), so the length is no longer lost silently. `TestRunSuiteLanesRetryErrorNamesTheOverrun` pins this.
- `scripts/test/test-lock.test.sh`: the new case where a dead light waiter's ticket is pruned and the next waiter heads the line passes. The whole shell suite passes ("all passed").
- `suitelane_run_test.go`: the new subtest "the real helper grants the light slot" runs `runLocalCheck` against the real `scripts/test-lock.py` and passes.
- The atlas section in `atlas/workflow/sdlc-binary.md` matches the code: `laneEligibility`, the `--lane` probe, `lane_overrun_seconds`, and the third slot beyond `WF_TEST_SUITES=2`.

**2. Critical findings**
- **The kept light log is deleted when the attempt is published, and its test fails.**
  - **Where:** `cmd/sdlc/localcheck_run.go:309` renames the light log to `<writer.LogPath>.light`, which sits inside the inflight attempt directory.
  - **What goes wrong:** `publishSuiteAttempt` (`teststats_store.go:420-432`) copies only the main log to `local-checks/<sha>/suite.log` and then removes the inflight directory. The `.light` file goes with it.
  - **Evidence:** a directory walk after the run shows no `.light` file anywhere under `.git/sdlc`.
  - **Family:** this is the 2nd finding in family `lane-overrun-evidence-loss`. Don't fix just this site. The rule that covers all of them is: **any evidence about the light phase must travel through the attempt's publish path.** That means the `SuiteAttempt` record or a file `publishSuiteAttempt` places next to the published log. Nothing should live only in the inflight directory or only in memory.
  - **Fix sketch:** have `publishSuiteAttempt` carry a sibling `.light` log to both places it publishes: `local-checks/<sha>/suite.log.light`, and `failed/<id>.light.log` for attempts kept unchanged. Then sweep the other sites in this family: the light phase's queue wait (still dropped), and the error-only overrun length (still not recorded).
  - **Plan and atlas:** fix the Revisions bullet, or drop the claim if the log is not going to be kept. Name the `.light` sibling in the atlas.

**3. Important findings**
- None new. The Revisions entry stating something the code does not do is part of the Critical finding above; I have not raised it separately as `plan-code-drift`.

**4. Minor findings**
- `lightLaneFor` tests `d.flakeContract != ""` as a stand-in for strict cleanup. Checking `d.StrictGroupCleanup` directly would state the actual rule. Today the two are equivalent on this path, since `bindTypedCheck` sets both.

**5. Test coverage notes**
- Run at head:
  - `go test -run 'Lane|Light|SuiteLane|QueueSuite'`: one failure, the one above. Everything else passes.
  - `scripts/test/test-lock.test.sh`: all passed.
- Not run: the full `go test ./cmd/sdlc`. It ran past 10 minutes and I stopped it, so I can't say whether anything outside the lane tests fails.
- The BR-1 regression was checked by removing the guard (red) and restoring it (green).

**6. Architectural notes**
- **ARCH-DRY:** pass.
- **ARCH-PURE:** pass. `nextSuiteLaneStep` and `laneEligibility` are pure and tested without IO.
- **ARCH-PURPOSE:** flag. The evidence the overrun leaves behind is still partial (the Critical finding).
- **ARCH-MOCK:** pass. A test now runs against the real helper.
- **ARCH-CONSTRAINTS:** pass. One light slot, a 240 s bound, niced +10.
- **ARCH-SECURE:** pass.
- **ARCH-ORDER:** pass. The lane state lives in memory inside one run and changes only through `nextSuiteLaneStep`.
- **ARCH-FUNERAL:** flag. The `.light` artifact's lifetime ends at publish without anyone deciding that, while the plan says it is kept. A real fix also has to bound it under `pruneLocalCheckLogs` and attempt pruning.
- **ARCH-EVOLVE:** pass. `lane` and `lane_overrun_seconds` are additive, and the round-trip test exists.

**7. Plan revision recommendations**
- Correct the 2026-10-23 round-1 Revisions bullet, "The stopped light run's log is kept beside the retry's, as `<log>.light`". Either describe where the published `.light` log really ends up once the fix lands, or remove the claim.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      lightLaneFor excludes flakeContract suites (localcheck_run.go:167); TestLightLaneForSkipsTypedRunnerSuites goes red with the guard removed and green with it.
  - id: BR-2
    disposition: addressed
    note: |
      Plan Revisions 2026-10-23 round-1 entry records lane_overrun_seconds, the --lane probe and laneEligibility; the atlas matches the code.
  - id: BR-3
    disposition: not-addressed
    note: |
      The overrun is now named in the retry error, but the kept .light log is deleted at publish and its test fails; see the new Critical finding.
  - id: BR-4
    disposition: addressed
    note: |
      The dead light waiter pruning shell case and the real-helper runLocalCheck subtest were added; both pass.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: lane-overrun-evidence-loss
    title: |
      The kept light-run log is deleted on publish, so TestRunLocalCheckLightOverrunRetriesInNormalLine fails at head
    detail: |
      This is the 2nd finding in family lane-overrun-evidence-loss. localcheck_run.go:309 renames the light log to <writer.LogPath>.light inside the inflight attempt directory. publishSuiteAttempt (teststats_store.go:420-432) copies only the main log to local-checks/<sha>/suite.log and removes that directory, so the .light file is lost. The test fails every time at suitelane_run_test.go:118 ("the light run's log was not kept: []"), and the plan's Revisions entry claims the log is kept. Rule to fix rather than this instance: every piece of light-phase evidence (overrun length, light log, light queue wait) must travel through the attempt publish path (the SuiteAttempt record or a file publishSuiteAttempt places next to the published log), never only the inflight directory or memory. Prevalence so far: 3 sites (log truncated on retry, now deleted on publish; overrun length only in the error text; light queue wait dropped).
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"466aeaa1-5dc4-43e3-b9c9-70c83812d9ff","repository":"github.com/xianxu/ariadne","issue_id":"000559","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/85eec7f191200bb0","reviewed_head":"c021d115ae1ce19f3afa780de8c71b9659cdfe16","slot":"ariadne:5","requested_at":"2026-10-23T15:37:52.939971Z"},"phase":"finished","pid":6299,"started_at":"2026-10-23T15:37:53.044573Z","finished_at":"2026-10-23T15:51:05.257443Z","session_id":"e8d1a633-db38-48d2-bc9e-e6178e039fa5","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T09:04:24-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 559 — Machine suite queue: a short-suite lane so one-minute suites don't wait an hour behind heavy ones |
| repo | ariadne |
| issue file | workshop/issues/000559-machine-suite-queue-a-short-suite-lane-so-one-minute-suites-don-t-wait-an-hour-behind-heavy-ones.md |
| boundary | whole-issue close |
| milestone | — |
| window | a561b0aba633d3103cbf9343c08117b8c0a1fab3..547c6fa0e5bb7b2ede22caa970feb3dfde405be2 |
| command | sdlc close --issue 559 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T09:04:24-07:00 |
| verdict | SHIP |

## Review

Let me look at the most recent commit (the fix for BR-5/BR-3) first.

I'm re-running the mutation check so it compiles, to confirm the queue-time fix has a regression test that fails without it.

Checking the plan's Core concepts table against the code (the glob didn't expand; retrying with the exact path).

```verdict
verdict: SHIP
confidence: high
```

BR-5, the Critical from round 2, is fixed. The fix applies the rule rather than patching the one site. The `.light` rename into the inflight directory, which publishing deleted, is gone. Everything the light phase leaves now has to travel on the one published attempt:
- the overrun's length is stored in `lane_overrun_seconds`;
- the light phase's wait is folded into the attempt's queue time, which now runs from the light request (`suitelane.go:146-149`);
- the partial log is discarded on purpose (ARCH-FUNERAL), and the atlas and the plan's Revisions both say so.

I checked this in the disposable checkout. The targeted lane tests and `scripts/test/test-lock.test.sh` pass at head. `TestRunLocalCheckLightOverrunRetriesInNormalLine`, which failed at the last round's head, now passes. I also reverted the queue-time fix (lines 146-149), and both `TestRunSuiteLanesQueueTimeCoversTheLightPhase` and `TestRunLocalCheckLightOverrunRetriesInNormalLine` failed; the file is restored and the tree is clean.

The lane gate now keys on `StrictGroupCleanup` instead of `flakeContract`. That is a superset of the old check: `bindTypedCheck` sets it with the contract (`localcheck_run.go:92-93`) and resume sets it too (`localcheck_resume.go:126`), so the BR-1 fix still holds and also covers resume runs.

One thing is still open, and it is Minor. If the normal-line retry fails or is interrupted, the overrun's length is still only in the error message (part of BR-3). It doesn't block shipping.

1. **Strengths**
   - `runSuiteLanes` (`suitelane.go:119-158`) stays a pure driver over `nextSuiteLaneStep`, with the IO injected. That let the queue-time fix get a fake-clock unit test (`suitelane_test.go:140-160`).
   - The fix answers the evidence-loss family with a rule written into the doc comment, the atlas (`sdlc-binary.md` "One record for the whole run") and the plan's Revisions, rather than another one-site patch.
   - Adding `lane` and `lane_overrun_seconds` is safe for older binaries. The strict decoder accepts keys it doesn't know (`canonicalSuiteKey`), and only the new fields' values are strictly checked (`teststats_store.go:187-189`).
   - The test for the lane gate (`suitelane_run_test.go:122-143`) checks `StrictGroupCleanup` itself rather than one way of setting it.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - BR-3, remaining part: when the retry errors (paused, interrupted, or can't run), no attempt is published. The overrun therefore isn't recorded, and the repo stays eligible for the lane. The cost is bounded: the next run wastes at most one niced 240-second light attempt before it overruns and records. The cheapest close is one atlas sentence saying an interrupted retry records nothing, the same as any interrupted run. The alternative is to carry the overrun into the retained interruption record.
   - The plan's Core concepts row still names `acquireSuiteSlot(..., lane)`. The code added a separate `acquireSuiteSlotLane` (`localcheck_run.go:852`).

5. **Test coverage**
   - The queue-time fix fails its tests when reverted (checked above).
   - Removing the rename has no positive test of its own, but there is nothing left to assert: the partial log is discarded on purpose.
   - Nothing tests that the overrun survives a retry that errors. That matches the remaining Minor.

6. **Architecture (all nine principles)**

   | Principle | Result | Why |
   |---|---|---|
   | ARCH-DRY | pass | The capability check reuses `probeHelperHelp` and `missingHelperOptions`; the light line reuses `Tickets` and its pruning. |
   | ARCH-PURE | pass | Eligibility and the state machine are pure; `runSuiteLanes` takes its IO as an injected function. |
   | ARCH-PURPOSE | pass | The rule is applied to all three evidence sites, not just the one named. |
   | ARCH-MOCK | pass | Tests run against the real helper script in temporary directories. |
   | ARCH-CONSTRAINTS | pass | The heavy-impact measurement is in the issue log, and the light slot is bounded at one. |
   | ARCH-SECURE | pass | The new attempt fields are range-checked on decode. |
   | ARCH-ORDER | pass | The order is one explicit transition function, tested event by event. The interrupted-retry gap is noted under Minor. |
   | ARCH-FUNERAL | pass | The partial log is discarded on purpose, and light tickets reuse the existing dead-waiter pruning. |
   | ARCH-EVOLVE | pass | The fields are additive, older readers accept them, and attempt records are never rewritten, so nothing gets dropped. |

7. **Plan revision recommendations**
   - Add a Revisions line: the slot acquisition with a lane is `acquireSuiteSlotLane`, not a changed `acquireSuiteSlot`.
   - Optionally, add a line saying an interrupted normal-line retry records no overrun.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      The .light rename is gone. The partial log is discarded on purpose, and the plan Revisions and atlas say so. TestRunLocalCheckLightOverrunRetriesInNormalLine passes at head. Reverting suitelane.go:146-149 turns it and TestRunSuiteLanesQueueTimeCoversTheLightPhase red.
  - id: BR-3
    disposition: not-addressed
    note: |
      The light queue wait (now in the queue time) and the truncated log (now discarded on purpose) are resolved. One part remains: when the normal retry errors, the overrun is still only in the error text (suitelane.go:131-135). The cost is bounded at one wasted 240s light run. Minor, non-blocking.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"72c144d3-0970-4369-a5bb-fce00c285ac6","repository":"github.com/xianxu/ariadne","issue_id":"000559","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/df7d7a9c630e4adc","reviewed_head":"547c6fa0e5bb7b2ede22caa970feb3dfde405be2","slot":"ariadne:5","requested_at":"2026-10-23T16:00:44.821418Z"},"phase":"finished","pid":11795,"started_at":"2026-10-23T16:00:44.984053Z","finished_at":"2026-10-23T16:04:23.434052Z","session_id":"b54a5d1c-9b71-4761-b877-3ae28da37076","outcome":"returned"}] -->
