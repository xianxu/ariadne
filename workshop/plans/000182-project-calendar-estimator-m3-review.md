# Boundary Review — ariadne#182 (milestone M3)

| field | value |
|-------|-------|
| issue | 182 — project calendar estimator: mechanize the commit reality-check (effort→calendar bridge) |
| repo | ariadne |
| issue file | workshop/issues/000182-project-calendar-estimator.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 6907844b37ed5d84ee288fa374df4de5a6856231..HEAD |
| command | sdlc milestone-close --issue 182 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-07-19T01:45:18-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have enough to complete the review. I verified the code, the pure/IO separation, signature consistency across all call sites, the vocab guard contract, the live brain ledger dependency, and traced the tests (Bash was unavailable in this session, so I could not execute `go test`; the diff is internally consistent and the assertions trace correctly).

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3 wires the three consumers cleanly and faithfully to the Spec's "inform, never block" contract: `set-status →committed` computes the forecast, injects it as reality-check evidence only when the operator gave no `--reality`, and derives `planned_finish` with correct precedence; `show`/`status` append a best-effort forecast line that never fails the read; `close` records a planned-vs-actual `## Calendar ledger` row. The single-source discipline holds (`RenderForecast`/`forecastForProject` shared across consumers; `appendProjectLedgerRow` heading-parameterized for both ledgers). No correctness bug blocks the boundary. What keeps it from a clean SHIP is a test-coverage gap (the calendar-row wiring is not asserted end-to-end) and a README Docs-gate omission — both non-blocking and cheap.

**1. Strengths**
- **Verification caught a real bug and pinned it.** The relative-path vantage fix (`projectforecast.go:154-172`) resolves `projectPath` to absolute before deriving repoDir/parentDir — without it the default relative `--projects-dir` collapses to `.`/`..` and silently reads 0 remaining for cross-repo issue lookups. Regression test `TestForecastForProject_RelativePathResolvesVantage` (`projectforecast_test.go:180+`) exercises the exact chdir+relative-path path. Good discipline.
- **Guard-ordering subtlety handled correctly** (`projectsetstatus.go:137-142`): `planned_finish` is applied to the in-memory doc *before* guards run so `baseline-set` (`guards.go:39-44`) sees the derived value, and on guard failure nothing is written (the file write is at the end). The comment states exactly why.
- **Inform-never-block is precisely implemented** (`projectsetstatus.go:87-111`): the only refusal is on *absence of any evidence* (no baseline + no `--reality` + no force); a genuine compute error degrades to a `cwarn` and proceeds. Matches Spec §"Consumers — inform, never block."
- **ARCH-DRY reuse** of `appendProjectLedgerRow` for both tables, with each error naming its own heading (`projectclose.go:393-417`).
- Live brain ledger `estimate-logic-project-v1.md` already carries the `## Calendar ledger` heading + table (verified), so real `project close` won't error on the new second append.

**2. Critical findings** — none.

**3. Important findings**
- **Calendar-row landing is not asserted end-to-end.** `TestProjectCloseRecordsFogAndArchives` (`cmd/sdlc/projectclose_test.go:48`) is the only full `runProjectClose` test, and it asserts the fog row (`:89`) but never the calendar row. The fixture (`:507`) now includes the `## Calendar ledger` heading so the close won't error, but if the wiring at `projectclose.go:139` (`prepareCalendarLedgerRow`) were deleted, every test — including this one and the pure-helper unit tests — would still pass. That's exactly the regression class this diff could ship. *Fix:* the fixture has `planned_finish: 2026-07-30`, close date `2026-07-16` → add an assertion that the ledger contains `| alpha | 2026-07-30 | 2026-07-16 | -14 | 2026-07-16 |`.
- **README Docs gate: `README.md:33-35` not extended for the forecast surfacing.** The "Bridge effort to calendar" block stops at bless-and-show-baseline; it doesn't mention that `show`/`status` now print the forecast-vs-deadline line, that commit derives `planned_finish`, or the new `--planned-finish` override flag. M1's review flagged README for the `throughput` verb and it was fixed; the same gate applies to the surface that is the actual payoff of #182. *Fix:* one or two lines under that block (or in the `set-status` prose at `:42`).

**4. Minor findings**
- ARCH-DRY (minor): `slipDays` (`projectclose.go:381`) and `dayDelta` (`forecast.go:146`) are near-identical ISO date-diff helpers in two packages (`main` vs `project`, latter unexported). Consolidation crosses a package boundary; acceptable to leave, but worth a shared exported helper if a third copy appears.
- Double doc read: `runProjectShow`/`runProjectStatus` parse the doc, then `forecastLine` re-reads it (`project.go:168`+`:173`; `projectstatus.go:42`+`:55`); likewise `computeCommitForecast` then `applyProjectStatus`. CLI cold path, not hot — thread the parsed `*Doc` if convenient.
- `computeCommitForecast` re-reads `d.FM("deadline")` (`:103`) instead of using the deadline already returned by `forecastForProject` (`:92` discards it as `_`); harmless, slightly inconsistent with `forecastLine` which uses the returned value.

**5. Test coverage notes**
- Pure core (`forecast.go`) and the two new surface test files are thorough on the happy + no-baseline paths, precedence (derived/pre-existing/manual), and the non-commit regression pin.
- Two defensive branches are untested but effectively unreachable when a baseline + `**phase-a:**` are present (the committed transition's `phase-a-estimate` guard runs first): `forecastLine`'s `"forecast: unavailable (...)"` branch (`projectforecast.go:141`) and `computeCommitForecast`'s cwarn-on-compute-error path (`projectsetstatus.go:99`). Low priority; note them rather than force coverage.

**6. Architectural notes for upcoming work**
- Every `show`/`status` now walks the full fleet (N project-doc reads + `computeBoard` each) best-effort. Fine at current fleet size; if it becomes noticeable on a read verb, a cheaper "active contention count" could replace the full assembly for the drift line.
- The roadmap rollup (#15/#185) remains "call `ComputeForecast` per project and sum" — the pure core needs no change, consistent with the plan. ARCH-PURPOSE: the three surfaces all derive from the one core; purpose fulfilled, not a cheap subset.

**7. Plan revision recommendations**
- Add a `## Revisions` note to `workshop/plans/000182-project-calendar-estimator-plan.md`: Task 3.1 Step 3 specifies `plannedFinish projectdoc.PlannedFinishDecision`, but the implementation places `plannedFinishDecision` in package `main` (`cmd/sdlc/projectsetstatus.go:25`) — the better location (a main-package orchestration concern, not a doc-package type). Update the step text so the plan matches the code.
