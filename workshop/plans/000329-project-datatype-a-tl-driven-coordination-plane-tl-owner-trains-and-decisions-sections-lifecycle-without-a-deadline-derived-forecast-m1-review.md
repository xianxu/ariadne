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
