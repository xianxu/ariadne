# Boundary Review — ariadne#182 (whole-issue close)

| field | value |
|-------|-------|
| issue | 182 — project calendar estimator: mechanize the commit reality-check (effort→calendar bridge) |
| repo | ariadne |
| issue file | workshop/issues/000182-project-calendar-estimator.md |
| boundary | whole-issue close |
| milestone | — |
| window | 449a25159856a7652e4a3c8c04dcea86098a7e49..HEAD |
| command | sdlc close --issue 182 |
| reviewer | claude |
| timestamp | 2026-07-19T01:56:42-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. The refactor is behavior-preserving (the `seen` map now dedups all visited files rather than only matches, but the output set is identical, and `DiscoverByIssueRef`'s own tests pin it). Let me write the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** ariadne#182 delivers its stated purpose cleanly at the whole-issue boundary: a measured throughput baseline (`SpanThroughput` + blessed TSV), a genuinely-pure forecast core (`ComputeForecast`/`RenderForecast`), fleet-load IO seams, and three "inform, never block" consumers — commit-transition evidence + `planned_finish` derivation, live `show`/`status` drift, and a `## Calendar ledger` calibration row at close. Every "Done-when" item is delivered, the Core-concepts table matches the code entity-for-entity at every path, docs are updated in-range (README + helptext + atlas), and the three prior milestone FIX-THEN-SHIP findings are all resolved in-tree (burned-down contention, symlink test, README, end-to-end calendar-row assertion). I found **no correctness bug that blocks shipping**. What keeps this from a clean SHIP is one plan↔code taxonomy contradiction (`ListActiveProjectFiles` is an IO function filed under the plan's "Pure entities" table) plus a handful of Minor nits. **Caveat:** the harness `EPERM` on `~/.claude/session-env` makes Bash unavailable this session (the known boundary-review limitation), so this is static analysis — the main agent must run `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ -count=1` and confirm green before crossing.

I verified the one real runtime dependency by hand: the live brain `estimate-logic-project-v1.md` **does** carry a well-formed `## Calendar ledger` heading + table (header + divider), so real `sdlc project close` will not error on the second ledger append.

### 1. Strengths

- **ARCH-PURE is exemplary in the true core.** `forecast.go` and `throughput.go` are IO-free (value inputs, `today` a parameter not `time.Now()`); `forecast_test.go`/`throughput_test.go` run on fixture structs/strings with zero mocks. All filesystem/clock lives in package `main` (`projectforecast.go`, `projectthroughput.go`) and is injected as `[]ProjectLoad`/`ThroughputBaseline`.
- **ARCH-DRY consolidation is real, not cosmetic.** `walkFleetProjects` (`discover.go:78`) is now the single fleet-walk source for both `DiscoverByIssueRef` and `ListActiveProjectFiles`; `appendProjectLedgerRow` is heading-parameterized for both tables (`projectclose.go:393`); `SpanThroughput` reuses `estimate.ParseRows`/`LedgerRow` (no second parser); `computeBoard` supplies remaining hours.
- **ARCH-PURPOSE fulfilled via one render path.** All three inform-surfaces derive from the single `RenderForecast` (`projectsetstatus.go:103`, `projectforecast.go:141`), and `planned_finish` precedence (manual > pre-existing > derived) is correct and provenance-logged (`projectsetstatus.go:198-217`).
- **The guard-ordering subtlety is handled correctly.** Derived `planned_finish` is applied to the in-memory doc *before* the guard loop so `baseline-set` (`guards.go:39-44`) sees it, and nothing is written on a guard failure (write is at the end) — the comment states exactly why (`projectsetstatus.go:136-141`).
- **The M2/M3 fixes landed and are pinned.** `boardRowsResolved` + `isActiveContention`'s `RemainingHours > 0` correctly make a burned-down-but-open project read board/0 and not contend (`forecast.go:39-44`, `projectforecast.go:97-113`), and the end-to-end calendar row is now asserted (`projectclose_test.go` fixture sets `planned_finish: 2026-07-30` → `-14` slip), closing the "delete the wiring and every test still passes" gap.

### 2. Critical findings

None.

### 3. Important findings

- **Plan Core-concepts table mislabels `ListActiveProjectFiles` as a Pure entity (plan↔code contradiction).** `workshop/plans/000182-…-plan.md:30` lists `ProjectFile + ListActiveProjectFiles (shared fleet walk)` under **"Pure entities (the conceptual core)"**, but `ListActiveProjectFiles` (`discover.go:159`) does filesystem IO (`SiblingRepoDirs`, `filepath.Glob`, `EvalSymlinks`) and its test (`listfleet_test.go`) requires real `t.TempDir()` files to run — by the checklist's own definition it "isn't really PURE." I'm rating this **Important, not Critical**, because the *code's* pure/IO separation is actually correct: the genuinely-pure math (`ComputeForecast`/`RenderForecast`/`SpanThroughput`) is IO-free, and `ListActiveProjectFiles` is properly an injected IO seam consumed by `ListFleetProjects`. This is a taxonomy slip in the plan table, not hidden coupling in the code. *Fix:* move that row from the "Pure entities" table to the "Integration points" table (its sibling `ListFleetProjects` already lives there), via a `## Revisions` entry — so the plan stops claiming a filesystem walk is pure.

### 4. Minor findings

- **Commit forecast + no-baseline refusal fire for *any* `f.To == "committed"`, before transition legality/no-op is checked** (`projectsetstatus.go:65-69`). Re-running `set-status --to committed` on an already-committed project with no blessed baseline now errors "no throughput baseline blessed" instead of being an idempotent no-op; an illegal `→committed` transition with a baseline prints a spurious forecast before `applyProjectStatus` rejects it. No wrong state is persisted (writes are gated on legal transition), so this is cosmetic — but a one-line gate (`prev == "defined"` is the only legal source of `committed`) would make the error path honest.
- **`slipDays` (`projectclose.go:381`) and `dayDelta` (`forecast.go:146`) are near-identical ISO date-diff helpers** in two packages (`main` vs `project`). Same `int(a.Sub(b).Hours()/24)` math, different sign semantics. Acceptable to leave per the M3 note; consolidate into a shared exported helper if a third copy appears (ARCH-DRY watch).
- **Trailing-window is 29 inclusive days, framed as "4 weeks"** (`projectthroughput.go`: `from := t.AddDate(0,0,-28)` → `today-28..today` = 29 inclusive days ÷ 7 = 4.14wk). Divides by slightly more weeks than a clean 28-day window, so a stable rate reads marginally low. Informational-only (never substituted); harmless.
- **ISO layout literal triplicated** as `isoDate` (estimate), `forecastISO` (project), `isoLayout` (main) plus the bare literal in `slipDays`/`renderBoard`. Pre-existing cross-package pattern (~15×), not a #182 regression.
- **Double doc-read on the commit/show cold path** (`computeCommitForecast` → `applyProjectStatus`; `runProjectShow` → `forecastLine`). CLI cold path, not hot; noted only.

### 5. Test coverage notes

Coverage is strong and pins real logic, not the implementation. Pure core: solo/multi-active/paused/unknown-other/zero-remaining/zero-baseline + render variants, with arithmetic pinned exactly (36h ÷ 55/3 → +14d). Fleet assembly: board/phase-a/unknown sources, the all-terminal-reads-board-0 regression, symlink exclusion, env-override + `errNoBaseline` for absent/unparsable. Consumers: precedence (derived/pre-existing/manual), no-baseline refusal + legacy `--reality` pass, non-commit regression pin, the vantage-bug regression (`TestForecastForProject_RelativePathResolvesVantage`), and the now-meaningful end-to-end calendar-row assertion. Remaining gaps are all low-value/unreachable-when-a-baseline-is-present: `RenderForecast`'s "under"/"on time"/"unparsable-deadline" branches (only "over" asserted), and the two defensive `cwarn`-on-compute-error branches (`projectforecast.go:141`, `projectsetstatus.go:99-101`) that the earlier `phase-a-estimate` guard makes hard to reach. Note for future, don't force.

### 6. Architectural notes for upcoming work

- The roadmap rollup (#15/#185) remains "call `ComputeForecast` per project and sum" — the pure core needs no change, consistent with the plan. ARCH-PURPOSE holds: no consumer is a deferred point-of-the-issue.
- Every `show`/`status` now walks the full fleet (N doc reads + `computeBoard` each) best-effort. Fine at current fleet size; if it ever shows on a read verb, a cheaper "active contention count" could replace the full assembly for the drift line.
- `forecastForProject` returns a **raw `Metadata()` error** (not `errNoBaseline`) at `projectforecast.go:172-174`; consumers already treat "any non-`errNoBaseline` error → quiet/cwarn fallback," which is correct — keep that invariant when adding a fourth consumer.

### 7. Plan revision recommendations

- **`workshop/plans/000182-project-calendar-estimator-plan.md:30` — reclassify `ListActiveProjectFiles`.** Add a `## Revisions` entry moving `ProjectFile + ListActiveProjectFiles (shared fleet walk)` out of the **"Pure entities"** table into the **"Integration points"** table (it does filesystem IO; only `ProjectFile`, the data struct, is pure). This is the one table↔code contradiction; the existing Revisions section (items 1–5) already records the other execution deltas accurately.

The verdict is **FIX-THEN-SHIP** at the gate: the sole Important is a plan-doc taxonomy fix (one `## Revisions` line), the code itself is sound, and the Minors are all note-for-future. Per the resolve-bundled-fixes convention (#174) these can be folded in before the close without a re-run — but the main agent must first execute the test suite and confirm green, since I could not run it this session.
