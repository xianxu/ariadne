# Boundary Review — ariadne#329 (milestone M1)

| field | value |
|-------|-------|
| issue | 329 — Project datatype: a TL-driven coordination plane (tl: owner, Trains and Decisions sections, lifecycle without a deadline, derived forecast) |
| repo | ariadne |
| issue file | workshop/issues/000329-project-datatype-a-tl-driven-coordination-plane-tl-owner-trains-and-decisions-sections-lifecycle-without-a-deadline-derived-forecast.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 44a185047879f707fdf805044c1ce5f77d2afdf8..3ebce5b3e3e8c61150a00af2e3ac4c260b089f52 |
| command | sdlc milestone-close --issue 329 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T17:25:29-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Diff read; now checking a few specifics (stale `done` sites, the archived-record test D1 promises, ledger eligibility) and running the tests.

Package tests pass; the `cmd/sdlc` subset is still running. Meanwhile I'm checking a suspected cycle-naming edge case in the scheduler.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what it claims. The lifecycle is now `defined → executing ⇄ closing → closed`, plus `dropped`, and it is driven by the cue transition table. The deadline fields are banned with a sentinel diagnostic that names the field. The trains parser and the critical-path forecast are pure and tested without IO. `set-status` and `close` were reworked, and the calendar forecast, Phase-A fog/calendar ledgers and the throughput baseline were removed cleanly, with no orphaned callers. README, help, datatype prose and atlas were updated in the same range. The targeted tests pass: `pkg/vocab`, `internal/{project,estimate,processmanual}`, `cmd/vocabulary` (cue on PATH), `vet_test.sh`, and the project/forecast/close/retro tests in `cmd/sdlc`. A broader `-run 'Close'` sweep hung in `TestClose_TrailingUnclosedMilestones_AcceptedByCloseReview`, a judge-dispatch test outside this diff that stalls when run inside a review, so I left it out. Two cheap fixes stand between this and SHIP. The cycle diagnostic can name a car that is not on the cycle, which breaks D6's "names a car on it". And the archived-record test D1 promises was never written.

**1. Strengths**
- `cmd/sdlc/internal/project/trainforecast.go` is a cleanly pure scheduler. It runs Kahn's algorithm with three parallel finish vectors (low, point, high), skips terminal cars, and uses the prior with a note for unresolved or unestimated cars. Tests are invariant-based (DAG properties, fuzz seeds) rather than restating the implementation.
- `estimate.CalibrateForecast` (`forecastcal.go:30`) reuses the `NewestPerIssue` and `ExcludedFromCalibration` eligibility rules, so there is one definition of an observation (ARCH-DRY). It also clamps low ≤ point ≤ high when the aggregate falls outside the per-issue quartiles.
- The `mvp-landed` guard takes `GuardCtx.IssueTerminal` as an injected function, built in `projectsetstatus.go` from the shared `projectIssueLookupFn`. Tests swap in a map fake at the same seam (ARCH-PURE, ARCH-MOCK).
- `set-status` refuses the close edge through the vocabulary (`FirstTransitionForEvent("close").To`) instead of a hardcoded `"done"` (`projectsetstatus.go:97`).
- `commitProjectClose` is simpler now that there is one artifact, and its rollback is still tested (`TestProjectCloseArchiveRenameFailureRestoresProject`).

**2. Critical findings**
None.

**3. Important findings**
- **The cycle error can name a car that isn't on the cycle** (`trainforecast.go:150-155`). After Kahn's algorithm stalls, the code reports the first car in `order` with `indeg>0`. That can be a car that is only downstream of the cycle. I reproduced it with a scratch test: trains `Y: r#5` and `X: r#3 → r#4`, edges `r#4 ⇒ r#5` and `r#4 ⇒ r#3`. The error says "cycle through r#5", but the cycle is r#3↔r#4. D6 says the error "names a car on it", and the existing test only checks for the substring "cycle". Fix sketch: from any stalled node, keep walking to a predecessor that is still stalled until a node repeats, then report that node or the whole loop. Add a test where a downstream car comes first in `order`.
- **The archived-record test D1 promised is missing.** D1 says "A test pins this: `project validate` over a fixture with an archived `status: done` record plus `deadline:` passes." Nothing in the range does that. `TestValidateInstance_ProjectDeadlineBan` only checks the live `workshop/projects/` corpus, and no `cmd/sdlc` test puts an archived record next to a live one. The behavior holds today only because `projectFiles(ProjectsDir)` happens to read just the live directory. A later change that widens `--all` or the changed-instance gate to the archive would start failing the eight frozen `done` records, and nothing would catch it. Fix: add a `runProjectValidate --all` test whose temp repo has `workshop/history/projects/x.md` with `status: done` and `deadline:`, and assert it passes.

**4. Minor findings**
- `runProjectForecast` writes parser warnings to `os.Stderr` instead of an injected writer, so tests can't see them (`projectforecast.go:59`).
- `projectIssueTerminalFn` re-runs `filepath.Abs` and `resolveWorkspace` for every ref. Resolve once outside the closure. The Abs + resolveWorkspace sequence also duplicates `forecastForProject` (ARCH-DRY).
- `calibrationLedgerPath` is a third copy of the `WF_CALIB_LEDGER`-else-`VelocityPath` rule, next to `close.go:1025` and `estimatesource.go:44` (ARCH-DRY). Extract one helper.
- `ExcludedFromCalibration` only excludes `Estimate == 0`. A hand-edited negative estimate in the ledger would give a negative multiplier in `CalibrateForecast` (ARCH-SECURE: the persisted ledger is untrusted). Guard with `Estimate > 0` there.
- D6 specifies `--path P`, but the verb takes a positional `[file]`. Fix either the plan or the flag.
- `forecastLine` parses Trains twice: once for the quiet check, then again inside `forecastForProject`. This is negligible.
- `atlas/process-manual.md` still describes the old project help. The Log notes it is generated and will be refreshed later; that's acceptable.

**5. Test coverage notes**
Coverage is good on the vocabulary predicates and transitions, each guard (including `mvp-landed` with a fake, and refusal on an empty `mvp_scope`), the `set-status` wrap refusing on an open MVP card, `close` from closing and the drop rules, the cue deadline ban end to end with the field named, and the forecast verb printing a range against a fixture ledger and fake lookup. The gaps are the two Important items above.

**6. Architecture checks**
- **ARCH-DRY: pass**, apart from the Minor helper duplication. Deleting the second forecast was the right call.
- **ARCH-PURE: pass.** The parser, scheduler and calibration are pure, and the IO is in `projectforecast.go` and `projectsetstatus.go`.
- **ARCH-PURPOSE: pass for M1.** The forecast was moved into M1 so `show` and `status` still have a consumer. find/resolve train and car, and ops conformance, are explicitly M2 and are not the point of M1.
- **ARCH-MOCK: pass.** Card reads go through the `projectIssueLookupFn` var, and the ledger through `WF_CALIB_LEDGER`.
- **ARCH-CONSTRAINTS: pass.** The plan sets an envelope of about 50 cars, one card lookup each, best-effort on `show`/`status`. Lookup errors turn into notes and never fail the verb.
- **ARCH-SECURE: pass with one Minor.** Malformed trains rows are skipped with a warning. The only gap is the negative estimate above.
- **ARCH-ORDER: pass.** Status changes go only through the cue transition table and the guard registry.
- **ARCH-FUNERAL: pass.** The fog/calendar ledger writers and the throughput baseline writer end here, and their files stay as history. Nothing new is durable.

For M2: `TrainPlan.Locate` already returns the 1-based car position. Keep `ProjectMatch` reading through `Doc.Trains(repo)` with the same default-repo rule `forecastForProject` uses, so a bare `#N` resolves the same way in find/resolve as in the forecast.

**7. Plan revision recommendations**
- Fix D6 one way or the other: change the verb to `--path P`, or note in `## Revisions` that the verb takes `forecast [file]`.
- Either add the archived `done` + `deadline` validate test D1 promises, or record a Revision explaining why it was dropped.

```findings
findings:
  - id: new
    severity: Important
    family: diagnostic-names-wrong-entity
    title: |
      Cycle error can name a car downstream of the cycle, not on it (trainforecast.go:150)
    detail: |
      After Kahn's algorithm stalls, the first car in order with indeg>0 is reported, and that car may only be reachable from the cycle. Repro: trains Y: r#5 and X: r#3 → r#4, edges r#4 ⇒ r#5 and r#4 ⇒ r#3 give "cycle through r#5". D6 promises a car on the cycle. Fix: walk stalled predecessors until a node repeats, and add a test where the downstream car comes first.
  - id: new
    severity: Important
    family: plan-promised-test-missing
    title: |
      D1's promised test (validate passes with an archived status done + deadline record) was not written
    detail: |
      TestValidateInstance_ProjectDeadlineBan only checks the live corpus. Nothing pins that project validate --all and the changed-instance gate skip workshop/history/projects, so widening them would start rejecting the eight frozen done records with nothing to catch it.
  - id: new
    severity: Minor
    family: duplicated-path-resolution
    title: |
      Third copy of the WF_CALIB_LEDGER-else-VelocityPath rule, plus Abs+resolveWorkspace repeated per ref in projectIssueTerminalFn
  - id: new
    severity: Minor
    family: untrusted-ledger-values
    title: |
      CalibrateForecast divides by r.Estimate, but the eligibility check excludes only Estimate == 0, so a hand-edited negative estimate gives a negative multiplier
  - id: new
    severity: Minor
    family: injected-io-writer
    title: |
      runProjectForecast writes warnings to os.Stderr instead of an injected writer
  - id: new
    severity: Minor
    family: plan-code-drift
    title: |
      D6 specifies --path P, but the forecast verb takes a positional file argument
```

---

## Re-review — 2026-10-12T17:46:23-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 329 — Project datatype: a TL-driven coordination plane (tl: owner, Trains and Decisions sections, lifecycle without a deadline, derived forecast) |
| repo | ariadne |
| issue file | workshop/issues/000329-project-datatype-a-tl-driven-coordination-plane-tl-owner-trains-and-decisions-sections-lifecycle-without-a-deadline-derived-forecast.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 44a185047879f707fdf805044c1ce5f77d2afdf8..6d8a469e0a9daac749bfe29a126ef37ef05e25c6 |
| command | sdlc milestone-close --issue 329 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T17:46:23-07:00 |
| verdict | SHIP |

## Review

I'm checking whether the prior-round fixes landed; next I'll run the tests and do the spec cross-check.

Now run tests and verify BR-1 / BR-2 regressions go red when reverted.

Project tests pass, apart from one tracker-reachability failure that looks environmental. Next I'm reverting the BR-1 fix to confirm its regression test goes red.

BR-1 is confirmed fixed: reverting it turns the test red. Next I'm sweeping the diff for other `os.Stderr` writes and checking D6 (BR-6).

Plan read. Next I'm going through the new trains parser and forecast scheduler.

Next, the calibration code, to check the low ≤ point ≤ high invariant the plan claims.

Calibration and scheduler look sound. Next: the guards, the CUE schema, set-status and close.

Pre-M1 set-status had the same drop behaviour, so that's out of scope. Next I'm checking that the plan's promised M1 tests exist.

Promised tests are all present. Next I'm sweeping the docs for stale references to the removed surfaces.

Prior findings and docs check out. Last check: how the ledger parser handles non-finite numbers like `Inf`/`NaN` from a hand-edited ledger.

```verdict
verdict: SHIP
confidence: high
```

All six prior-round findings are fixed, and I checked each one against the code at the pinned head, not the commit message. Reverting the BR-1 cycle fix turns `TestForecastTrainsCycleAndEmpty` red ("cycle through r#5"); restoring it turns it green. The M1 boundary does what the plan's revised M1 section says:
- the vocabulary pivot to defined → executing → closing → closed, plus dropped;
- the deadline ban;
- the trains parser and the pure scheduler;
- calibration;
- the `mvp-landed` guard on the shared card-lookup seam;
- removal of the #182 calendar, throughput and Phase-A code;
- matching updates to README, atlas, helptext and the datatype prose.

`go build`, `go vet` and the tests pass for `internal/project`, `internal/estimate`, `pkg/vocab` and `cmd/vocabulary`. The project-related tests in `cmd/sdlc` pass too, with one exception: `TestLegacyRepoLockSetStatusMutationWaits` fails. That test covers issue set-status, which this diff does not touch, and it fails because no issue tracker is reachable from this checkout, so I treat it as environmental. One caveat: the full `cmd/sdlc` package run hit Go's 10-minute default timeout, so I covered that package with the filtered run only. Nothing blocks SHIP. One new Minor finding and one Minor note remain.

**1. Strengths**
- `trainforecast.go` keeps the core pure. The forecast core takes a `TrainPlan`, a car-estimate map and a calibration, and has no IO. The three multipliers (low, point, high) run through one Kahn pass, and iteration order is fixed so output is stable. `carOnCycle` (trainforecast.go:198) walks stalled predecessors to a repeated car, which is provably a car on the cycle.
- `CalibrateForecast` (forecastcal.go:51-56) clamps low ≤ point ≤ high when a few large issues pull the aggregate point outside the per-issue quartiles. That is a real edge case, and it is handled and commented.
- One lookup seam (ARCH-DRY, ARCH-MOCK): both `mvp-landed` (`projectIssueTerminalFn`) and the forecast's car hours go through `projectIssueLookupFn`. `projectCheckout` now holds the absolute-path-to-workspace rule in one place, and `calibrationLedgerPath` has three callers: close, converge-report and forecast.
- The lifecycle is an explicit table in `project.cue`. The cue laws enforce it: every status has a description, every status except the entry is reachable, and every non-terminal status can be left. `set-status` sends close to its dedicated verb by looking up the `close` event (`FirstTransitionForEvent("close")`) instead of hardcoding `"closed"`.
- The deadline ban is tested at both layers: the cue fixtures in `vet_test.sh:39-48` and `TestValidateInstance_ProjectDeadlineBan`.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- **Ledger floats accept `Inf`/`NaN`** (family `untrusted-ledger-values`, 2nd finding). `parseLedgerFloat` (ledger.go:271) accepts `Inf` and `NaN`. A hand-edited `Inf` estimate passes `Estimate > 0`, gives that row a multiplier of 0, drives the point multiplier to 0, and makes the prior infinite. The forecast then reports a meaningless 0h or Inf with no warning.
  - BR-4 fixed one instance. The rule covering the class is ARCH-SECURE: a ledger number becomes a typed value once, at the parse boundary. `parseLedgerFloat` should reject non-finite values so the row reads as absent, and filters downstream should stop re-checking.
- **Deadline ban has one escape value.** The sentinel-string ban (`project.cue`, `deadline?: "retired by …"`) accepts a record whose `deadline` is exactly that sentinel string. The diagnostic trade-off is documented, so this is a note only.
- **Abs error dropped.** `loadForecastCalibration` discards the error from `filepath.Abs` (`brainAbs, _ :=`). It is harmless in practice.

**5. Test coverage notes**
- BR-2: `TestProjectValidationScopeExcludesArchive` pins that both `validate --all` and the changed-instance gate (`gateForPath`) ignore `workshop/history/projects/`. It stubs `validateFrontmatterFn`, which is fine because the property under test is scope, not schema.
- BR-4: the regression rows `splitRow("7", -2, …)` and `splitRow("8", 2, 1, -1)` land in a test that asserts N == 3, so a regression fails it.
- All the plan's M1 tests exist:
  - vocab predicates and transitions: `pkg/vocab/project_test.go`;
  - each guard, including `mvp-landed` with a fake lookup: `guards_test.go`;
  - wrap refusing an open MVP issue: `TestProjectSetStatusWrapChecksMVPCards`;
  - `closing→closed` and drop: `projectclose_test.go`;
  - validate refusing `deadline:`: `TestValidateInstance_ProjectDeadlineBan`.

**6. Architecture**

| Marker | Result | Notes |
|---|---|---|
| ARCH-DRY | pass | The BR-3 consolidation is verified. |
| ARCH-PURE | pass | Parser, scheduler, calibration and guards are pure. IO is confined to `forecastForProject` and `projectIssueTerminalFn`. |
| ARCH-PURPOSE | pass | The calendar forecast is deleted rather than kept alongside. Find/resolve train and car fields are explicitly re-scoped to M2 in Revisions. |
| ARCH-MOCK | pass | The existing `projectIssueLookupFn` var seam with a map-backed fake; no new external binary. |
| ARCH-CONSTRAINTS | pass | The declared envelope is about 50 lookups per forecast, best-effort on show/status. It is not measured; measure once on a real ops project in M2. |
| ARCH-SECURE | pass with Minor | The parse boundary for ledger numbers (the Minor finding above). |
| ARCH-ORDER | pass | Lifecycle state is the cue transition table. Set-status reads the status, checks guards and writes, with no persisted flags in between. Drop via `set-status` not archiving predates this diff. |
| ARCH-FUNERAL | pass | Nothing durable is created. The fog, calendar and throughput writers end; their brain files stay as history with no writer or reader, and the plan states this. |

For M2: `ProjectMatch` train/car should reuse `TrainPlan.Locate`. Feed the D10 ops patch through the real `ParseTrains` warnings path.

**7. Plan revision recommendations:** none. The Revisions section already records the M1 boundary move, the sentinel-string ban, BR-1, BR-2 and the D6 positional-argument change.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      carOnCycle walks stalled preds; reverting it turns TestForecastTrainsCycleAndEmpty red ("cycle through r#5"), restoring it turns it green.
  - id: BR-2
    disposition: addressed
    note: |
      TestProjectValidationScopeExcludesArchive pins that validate --all and gateForPath both skip workshop/history/projects.
  - id: BR-3
    disposition: addressed
    note: |
      close.go:1030 and convergereport now call calibrationLedgerPath; projectCheckout consolidates Abs+resolveWorkspace and resolves once per closure.
  - id: BR-4
    disposition: addressed
    note: |
      Eligibility now requires Estimate > 0 and ConvergeHours >= 0; forecastcal_test rows 7 and 8 pin it via the N == 3 assertion.
  - id: BR-5
    disposition: addressed
    note: |
      runProjectForecast takes an injected stderr (cmd.ErrOrStderr()); no os.Stderr remains in the added non-test code.
  - id: BR-6
    disposition: addressed
    note: |
      Plan Revisions records D6's --path P becoming a positional [file]; it matches Use "forecast [file]".
findings:
  - id: new
    severity: Minor
    family: untrusted-ledger-values
    title: |
      parseLedgerFloat accepts Inf/NaN, so a hand-edited Inf estimate zeroes the forecast multiplier
    detail: |
      2nd finding in this family. BR-4 filtered negatives at one reader. The rule covering the class (ARCH-SECURE): a ledger number becomes a typed value at the parse boundary, so parseLedgerFloat (ledger.go:271) should reject non-finite values (the row reads as absent) and downstream filters stop re-checking. Today an Inf estimate passes Estimate > 0, gives the row m = 0 and an infinite prior, and the forecast prints 0h or Inf with no warning.
  - id: new
    severity: Minor
    family: schema-ban-bypassable
    title: |
      deadline ban accepts the sentinel string itself as a valid value
    detail: |
      deadline?: "retired by ariadne#329 ..." rejects every other value, but a record carrying exactly that string validates. The trade-off for a readable cue diagnostic is documented, so this is a note only.
```
