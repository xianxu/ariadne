# Project Calendar Estimator Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bridge effort to calendar — a measured throughput baseline plus a pure forecast core that INFORMS (never blocks) at project commit, renders live drift on show/status, and closes the loop with a planned-vs-actual calendar ledger row.

**Architecture:** One pure forecast core (`ComputeForecast`/`RenderForecast`) fed by three IO seams: a blessed throughput baseline (operator picks the span, machinery measures the rate from the calibration ledger), a fleet project enumeration (reusing #171's sibling walk + the existing `computeBoard`), and the project doc itself. Three thin consumers print the same rendered statement: `set-status →committed` (records it as reality-check evidence and derives `planned_finish:`), `project show`/`status` (live drift), `project close` (calendar ledger row beside the fog row). The guard registry and cue model are untouched — the caller fills the evidence, so the existing `evidenceGuard` machinery records it.

**Tech Stack:** Go (`cmd/sdlc`, `cmd/sdlc/internal/{project,estimate}`), TSV ledgers in brain, markdown-table ledger in `estimate-logic-project-v1.md`.

**Spec:** `workshop/issues/000182-project-calendar-estimator.md` `## Spec` (2026-07-18 brainstorm; inform-never-block).

**Arch principles in play** (`sdlc arch-principles`):
- **ARCH-PURE** — the calendar math is a pure function over value inputs; every IO (ledger read, fleet walk, doc write) is a caller-side seam. Span math is pure too (rows in, h/wk out).
- **ARCH-DRY** — reuse `computeBoard` (remaining hours), the #171 sibling walk, `evidenceGuard`'s recording (caller injects computed evidence — zero guard/model change), `VelocityPath`, and a parameterized `appendProjectLedgerRow` for the second ledger table.
- **ARCH-PURPOSE** — the purpose is *informing at the right surfaces*: commit, show/status, close all render the SAME statement via one `RenderForecast`. A commit-only wiring would under-deliver (the operator asked for surfacing, explicitly naming `sdlc project show`).

---

## Core concepts

The system reasons about a **blessed throughput baseline** (measured issue-hours/week over an operator-designated span), a **project load** (one project's status + remaining hours), and a **forecast** (projected finish + the arithmetic trail). Unit identity is load-bearing: numerator (remaining issue-estimates) and denominator (ledger actuals/week) are both ship wall-clock engineer+AI hours (#118) — no conversion to human hours anywhere.

### Pure entities (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| `SpanThroughput` | `cmd/sdlc/internal/estimate/throughput.go` | new |
| `ThroughputBaseline` (+ TSV parse/render) | `cmd/sdlc/internal/estimate/throughput.go` | new |
| `ProjectLoad` | `cmd/sdlc/internal/project/forecast.go` | new |
| `Forecast` / `ComputeForecast` | `cmd/sdlc/internal/project/forecast.go` | new |
| `RenderForecast` | `cmd/sdlc/internal/project/forecast.go` | new |
| `appendProjectLedgerRow` (heading-parameterized) | `cmd/sdlc/projectclose.go` | modified |

- **SpanThroughput** — `SpanThroughput(rows []LedgerRow, from, to string) (SpanMeasure, error)`: pure math over parsed calibration-ledger rows → `{HoursPerWeek, Rows, UntrustedRows, Days}`. Inclusive day count; divide by `days/7.0` (partial weeks well-defined). Includes `window_trusted=no` rows (their measured hours are real; excluding them under-counts worse) and reports their count so the bless output can warn. `LedgerRow` (`{Issue string, Actual float64, Date string, WindowTrusted bool}`) + `ParseCalibrationLedger(text string)` live beside it — parsing is pure string→struct.
  - **Relationships:** N ledger rows → 1 SpanMeasure; SpanMeasure 1:1 feeds a ThroughputBaseline row.
  - **DRY rationale:** one parser/measure shared by `--bless` and the bare verb's trailing-4-week comparison.
  - **Future extensions:** per-repo filters (roadmap-level throughput split) widen the row predicate, not the math.
- **ThroughputBaseline** — `{BlessedDate, SpanStart, SpanEnd string, HoursPerWeek float64, Rows int, Ceiling int}` + `ParseBaselineTSV(text) ([]ThroughputBaseline, error)` (last row = current) + `RenderBaselineRow`. Append-only TSV = free history of re-blessings.
  - **Relationships:** N historical rows, 1 current; consumed by value in `ComputeForecast`.
  - **DRY rationale:** single TSV schema for writer (bless) and readers (forecast, bare verb).
- **ProjectLoad** — `{Name, Repo, Status string, RemainingHours float64, RemainingSource string, Warning string}`. `RemainingSource` ∈ `board|phase-a|unknown` — the Phase-A fallback applies UNIFORMLY (this project and fleet siblings alike); `unknown` (no board rows resolve AND no Phase-A) carries a warning and weighs 0.
  - **Relationships:** N:1 with a fleet project file; the forecast's `others` input is `[]ProjectLoad`.
- **Forecast / ComputeForecast** — `ComputeForecast(b ThroughputBaseline, this ProjectLoad, others []ProjectLoad, today string) Forecast`. `n` = 1 + count(others with status `committed`|`executing`); `share = b.HoursPerWeek / n`; `projectedFinish = today + ceil(this.RemainingHours/share × 7) days`. Ceiling is a WARNING threshold only (`n > b.Ceiling` → warning line), never arithmetic. Paused others → zero-weight named risk lines. Output: `{ProjectedFinish string, N int, SharePerWeek, Remaining float64, RemainingSource string, CeilingWarning string, PausedRisks []ProjectLoad, Notes []string}`. Zero/absent baseline hours or zero remaining → `error` (callers fall back; the pure core never guesses).
  - **DRY rationale:** the one place calendar math exists; the future roadmap rollup is "call per project and sum" with no changes here.
- **RenderForecast** — `RenderForecast(f Forecast, deadline string) string`: the identical one-paragraph statement every surface prints, e.g. `forecast: 36.0h remaining (board) ÷ 55.0h/wk ÷ 2 active → ~27.5h/wk share → lands ~2026-09-20 (deadline 2026-09-01: 19 days over). paused: metis-v2 (14h) — resuming invalidates this. [3 active projects exceed ceiling 2]`. Deadline absent → "no deadline set" clause instead of the comparison.
- **appendProjectLedgerRow** — existing fog-table appender generalized to take the section heading (`"Fog ledger"` | `"Calendar ledger"`); behavior for the fog path is pinned unchanged.

### Integration points (where pure meets the world)

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `runProjectThroughput` (bless + show) | `cmd/sdlc/projectthroughput.go` | new | calibration ledger + baseline TSV |
| `loadThroughputBaseline` | `cmd/sdlc/projectforecast.go` | new | baseline TSV + `WF_THROUGHPUT_BASELINE` |
| `ListFleetProjects` | `cmd/sdlc/projectforecast.go` | new | fleet fs walk + `computeBoard` |
| `forecastForProject` | `cmd/sdlc/projectforecast.go` | new | assembles the pure inputs |
| commit hook in `runProjectSetStatus` | `cmd/sdlc/projectsetstatus.go` | modified | evidence injection + `planned_finish` |
| forecast section in show/status | `cmd/sdlc/project.go` / `cmd/sdlc/projectstatus.go` | modified | render only |
| calendar row at close | `cmd/sdlc/projectclose.go` | modified | `estimate-logic-project-v1.md` |

- **runProjectThroughput** — `sdlc project throughput [--bless FROM..TO] [--ceiling N]`. Bless: read ledger (`estimate.VelocityPath(brainDir, "calibration-ledger.tsv")`), `SpanThroughput`, refuse empty span, append `RenderBaselineRow` to `VelocityPath(brainDir, "throughput-baseline.tsv")` (create with header comment if absent), print measure + untrusted-row warning. Bare: print current baseline + trailing-28-day comparison (computed live from the ledger; comparison only — never substituted).
  - **Injected into:** nothing — it's a leaf verb; the baseline file is the interface.
- **loadThroughputBaseline** — resolves `WF_THROUGHPUT_BASELINE` env override, else `VelocityPath(f.BrainDir, "throughput-baseline.tsv")`; parses; returns current baseline or a typed `errNoBaseline` every consumer maps to its fallback.
- **ListFleetProjects** — **deliberately in package `main`** (`computeBoard` lives there): walk `SiblingRepoDirs(parent)` filtered by `isFleetSibling`, glob each repo's `vocab.Project().Discovery().Home` (active only — archived/terminal projects hold no load; brain legacy home included while metis-v2 lives there, same `gitx.IsBrainRepo` branch as `DiscoverByIssueRef`), parse each doc, build `ProjectLoad` via `computeBoard` with that repo's vantage (`projectIssueLookupFn(ref, repoDir)`), Phase-A fallback, `unknown` → weight 0 + warning. Excludes the subject project by resolved path.
  - **Injected into:** `ComputeForecast` via `[]ProjectLoad` — the pure core never walks anything.
- **forecastForProject** — the shared assembly used by all three consumers: `(doc, path, projectsDir, brainDir, today) → (Forecast, deadline string, error)`; maps `errNoBaseline` through untouched so each consumer picks its own fallback.
- **Commit hook** — in `runProjectSetStatus`, only when `f.To == "committed"`: try `forecastForProject`; on success, `f.Reality` empty → `ctx.Evidence["reality-check"] = "computed: " + RenderForecast(...)` (the EXISTING evidence machinery records it in the Log — zero guard change); `planned_finish` absent → `d.SetFM("planned_finish", f.ProjectedFinish)` + note; `planned_finish` present or `--planned-finish` given → kept, logged as manual override. On `errNoBaseline` → legacy behavior exactly (operator must pass `--reality` prose; error message gains the bless hint).
- **Calendar row at close** — after the fog row: `| name | planned_finish | actual_finish(today) | slip_days | date |` appended under `## Calendar ledger` in the same `estimate-logic-project-v1.md`; `planned_finish` empty → row records `n/a` slip (informational; never refuses). Covered by the existing `--no-ledger` skip.

**Test surface.** Pure entities: colocated `_test.go`, no mocks (fixture strings/structs). Integration: temp-dir fleets + `WF_THROUGHPUT_BASELINE`/`WF_CALIB_LEDGER`-style overrides + real files, the established pattern (`resolve_test.go`, `peerwrite_apply_test.go`); no external service → no process fake needed.

---

## Chunk 1: M1 — Throughput measurement + blessed baseline

### Task 1.1: Pure span math + ledger parse (`internal/estimate/throughput.go`)

**Files:** Create `cmd/sdlc/internal/estimate/throughput.go`, `cmd/sdlc/internal/estimate/throughput_test.go`

- [ ] **Step 1: Write the failing tests** — table tests: `ParseCalibrationLedger` (skips `#` comments + header line, reads `issue/actual/window_trusted/date` columns, tolerates short lines); `SpanThroughput` over fixture rows: 28-day span → `Σactual/4.0`; 27-day span divides by `27/7.0`; rows outside span excluded; untrusted counted but included in the sum; empty span → error. `ParseBaselineTSV` (last row current, comments skipped, bad float → error) + `RenderBaselineRow` round-trip.
- [ ] **Step 2: Run to verify failure.** `go test ./cmd/sdlc/internal/estimate/ -run 'Throughput|Baseline|CalibrationLedger' -v` → FAIL undefined.
- [ ] **Step 3: Implement** — `LedgerRow`, `ParseCalibrationLedger(text)`, `SpanMeasure{HoursPerWeek float64; Rows, UntrustedRows, Days int}`, `SpanThroughput(rows, from, to)` (dates ISO `YYYY-MM-DD`, inclusive; `days = to-from+1`; `HoursPerWeek = sum / (float64(days)/7.0)`), `ThroughputBaseline`, `ParseBaselineTSV`, `RenderBaselineRow`. Baseline TSV columns: `blessed_date span_start span_end hours_per_week rows ceiling` (tab-separated, `#` comments).
- [ ] **Step 4: Run to verify pass.**
- [ ] **Step 5: Commit** — `#182 M1: pure span throughput + baseline TSV codec`.

### Task 1.2: The `sdlc project throughput` verb

**Files:** Create `cmd/sdlc/projectthroughput.go`, `cmd/sdlc/projectthroughput_test.go`; Modify `cmd/sdlc/project.go` (register), `cmd/sdlc/helptext/project.md`

- [ ] **Step 1: Write the failing process-level test** — temp brain dir with a fixture `calibration-ledger.tsv`; `--bless 2026-06-16..2026-07-13` → baseline file created under the velocity dir with one row, printed h/wk matches, untrusted rows warned; second bless appends (2 rows, last wins); empty-span bless refuses; bare verb prints current baseline + trailing comparison; no baseline → bare verb says so with the bless hint. Use a `--brain-dir` flag (mirrors close) so tests stay hermetic.
- [ ] **Step 2: Run to verify failure.**
- [ ] **Step 3: Implement** — flags `--bless FROM..TO`, `--ceiling N` (default 2, only meaningful with `--bless`), `--brain-dir` (default `../brain`). Read-only bare form; bless is `markMutatingCommand`. Wire `newProjectThroughputCmd()` into `NewProjectCmd`; add the subcommand row + a short section to `helptext/project.md`.
- [ ] **Step 4: Run to verify pass; run `go test ./cmd/sdlc/ -run Helptext`** (embed tests).
- [ ] **Step 5: Commit** — `#182 M1: sdlc project throughput — bless + show the measured baseline`.
- [ ] **Step 6: Milestone-close** — `sdlc milestone-close --issue 182 --milestone M1 --no-project` (project tracks at issue granularity).

## Chunk 2: M2 — Pure forecast core + fleet load assembly

### Task 2.1: `ComputeForecast` / `RenderForecast` (pure)

**Files:** Create `cmd/sdlc/internal/project/forecast.go`, `cmd/sdlc/internal/project/forecast_test.go`

- [ ] **Step 1: Write the failing decision-table test** — solo project (n=1, full share, date = today + remaining/hpw weeks); two active others (n=3, share/3); paused other → excluded from n, present in `PausedRisks`; `unknown`-source other → weight 0 + note; `n > ceiling` → `CeilingWarning` set; zero remaining → error; zero baseline h/wk → error; date arithmetic pinned exactly (e.g. 36h remaining, 55h/wk, n=2 → 27.5h/wk → 36/27.5 = 1.309 wk → ceil(9.16)=10 days → today+10). `RenderForecast`: statement contains remaining+source, share arithmetic, projected date, deadline delta in days (over/under), paused lines, ceiling warning; absent deadline → "no deadline set".
- [ ] **Step 2: Run to verify failure.** `go test ./cmd/sdlc/internal/project/ -run Forecast -v`
- [ ] **Step 3: Implement** per the Core-concepts contract. Date math on `time.Time` parsed from ISO; output dates ISO.
- [ ] **Step 4: Run to verify pass.**
- [ ] **Step 5: Commit** — `#182 M2: pure calendar forecast core`.

### Task 2.2: Fleet load assembly + baseline loader (IO, package main)

**Files:** Create `cmd/sdlc/projectforecast.go`, `cmd/sdlc/projectforecast_test.go`

- [ ] **Step 1: Write the failing temp-fleet test** — parent with `ariadne` (subject project, executing, breakdown rows → issues with estimates), `metis` (executing project, remaining via its own issues), `nous` (paused project), non-project repo ignored; `ListFleetProjects` returns loads with correct statuses/remaining/vantage; subject excluded by path; sibling with unresolvable rows + `**phase-a:** 12h` → `phase-a` source 12h; neither → `unknown` weight 0. `loadThroughputBaseline`: env override wins; typed `errNoBaseline` when file absent.
- [ ] **Step 2: Run to verify failure.**
- [ ] **Step 3: Implement** — `ListFleetProjects(parentDir, excludePath string) []ProjectLoad` (best-effort per repo: unreadable/unparsable file → skipped with a warning load? No — skip silently is a silent cap; append `unknown` load with Warning so the statement can surface it); `loadThroughputBaseline(brainDir string)`; `forecastForProject(...)` assembling this-project load (board → Phase-A fallback via `ParsePhaseA`) + others + baseline + today.
- [ ] **Step 4: Run to verify pass.**
- [ ] **Step 5: Commit** — `#182 M2: fleet load assembly + baseline loader (IO seams)`.
- [ ] **Step 6: Milestone-close** — `sdlc milestone-close --issue 182 --milestone M2 --no-project`.

## Chunk 3: M3 — The three consumers

### Task 3.1: Commit hook in set-status

**Files:** Modify `cmd/sdlc/projectsetstatus.go`; Test `cmd/sdlc/projectsetstatus_test.go` (extend)

- [ ] **Step 1: Write the failing tests** — with baseline env + temp fleet: `set-status --to committed` (no `--reality`) succeeds, Log gains `- reality-check: computed: …` evidence block, `planned_finish:` set to the projected date, statement printed to stdout; pre-existing `planned_finish` → kept + `manual planned_finish kept …` Log note; no baseline → refuses with legacy `guard reality-check` error + bless hint; no baseline + `--reality 'fits'` → passes exactly as today (regression pin); non-committed transitions untouched (regression pin).
- [ ] **Step 2: Run to verify failure.**
- [ ] **Step 3: Implement** — in `runProjectSetStatus` before `applyProjectStatus` (keeping `applyProjectStatus`'s signature): when `f.To == "committed"`, call `forecastForProject`; on success print `RenderForecast`, inject evidence when `f.Reality == ""`, and pass the derived date through a new optional field on `GuardCtx`? NO — `applyProjectStatus` owns the doc write; add a `derivedPlannedFinish string` parameter threaded to it (set FM only when currently empty, append the manual-override note otherwise). On `errNoBaseline`: leave `f.Reality` handling exactly as-is; wrap the eventual guard error with the bless hint.
- [ ] **Step 4: Run to verify pass; full `go test ./cmd/sdlc/ -run ProjectSetStatus -count=1`.**
- [ ] **Step 5: Commit** — `#182 M3: commit transition computes, informs, derives planned_finish — never blocks on the answer`.

### Task 3.2: show/status surfaces + close calendar row

**Files:** Modify `cmd/sdlc/project.go` (`runProjectShow`), `cmd/sdlc/projectstatus.go` (`renderBoard` or its caller), `cmd/sdlc/projectclose.go`; helptext `project.md`, `set-status` help if present; Tests: extend `projectstatus_test.go` / `project_crud_test.go` / `projectclose_test.go`

- [ ] **Step 1: Write the failing tests** — `project show` and `project status` on a committed/executing project with baseline present include the forecast line (drift vs deadline); without baseline → one quiet `forecast: no blessed baseline (sdlc project throughput --bless …)` line, no error. Close: fog row still appended (regression pin); NEW calendar row under `## Calendar ledger` with `slip_days = actual_finish − planned_finish` (negative = early); `planned_finish` empty → `n/a` row; `--no-ledger` skips both; missing `## Calendar ledger` heading → same add-the-heading error contract as fog. `appendProjectLedgerRow` heading parameterization pinned for both tables.
- [ ] **Step 2: Run to verify failure.**
- [ ] **Step 3: Implement** — show/status call `forecastForProject` best-effort (any error → the one quiet line; never fail a read verb); generalize `appendProjectLedgerRow(text, heading, row)`; close computes slip from `metadata.PlannedFinish` vs `today`. Add `## Calendar ledger` heading + table to the live `estimate-logic-project-v1.md` in brain (one-time data edit, commit rides nous's auto-commit like #171 M6).
- [ ] **Step 4: Run to verify pass; full suite `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ -count=1`.**
- [ ] **Step 5: Update atlas** — `atlas/workflow/sdlc-binary.md`: forecast section (blessed baseline, pure core, three consumers, unit identity); tick issue Plan rows.
- [ ] **Step 6: Commit** — `#182 M3: forecast surfaces on show/status + calendar calibration row at close`.
- [ ] **Step 7: Milestone-close** — `sdlc milestone-close --issue 182 --milestone M3 --no-project`.

## Manual Verification

1. **Bless a real baseline:** `sdlc project throughput --bless 2026-06-22..2026-07-19` against the live brain ledger; sanity-check the printed h/wk against the known weekly sums (≈110h/wk); confirm `throughput-baseline.tsv` lands in brain's velocity dir.
2. **Live forecast:** `sdlc project show --slug project-management-primitive` shows the forecast line with the real remaining hours and deadline drift.
3. **Commit transition (when a project next reaches it):** statement printed + recorded; `planned_finish` derived.

## Issue close (after M3)

`sdlc close --issue 182 --verified '<test names + manual results>'` — omit `--actual` (measure+adopt). Estimate to set at change-code: v3.1 derivation in the issue `## Estimate` (M1 smaller-go-module, M2 greenfield-go-module, M3 cross-cutting-refactor — see #171's block for the format).
