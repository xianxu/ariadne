# Boundary Review — ariadne#182 (milestone M2)

| field | value |
|-------|-------|
| issue | 182 — project calendar estimator: mechanize the commit reality-check (effort→calendar bridge) |
| repo | ariadne |
| issue file | workshop/issues/000182-project-calendar-estimator.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 8ec574de67fea445902127c6fc0ba5347b274d3a..HEAD |
| command | sdlc milestone-close --issue 182 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-07-19T01:20:42-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have enough to complete the review. Note: I attempted to run `go build`/`go test` but the shell is unavailable in this session (harness `EPERM` creating its session-env dir, independent of the sandbox flag), so verification below is **static** — I read every entity in the diff against its call sites and the Spec/Plan rather than executing the suite.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The M2 boundary delivers exactly its scope — a pure `ComputeForecast`/`RenderForecast` core plus the fleet-load IO seam — with clean pure/IO separation, a genuine DRY refactor of the fleet walk, and a decision-table test that pins the calendar arithmetic exactly. The Core-concepts table matches the code entity-for-entity, and the atlas was updated in-range (docs gate satisfied; no README surface lands until M3). Nothing blocks shipping. The one substantive finding is a contention-classification edge in `projectLoadFromDoc` (a fully-burned-down-but-open project falls back to its stale Phase-A total), which is within the plan's stated fallback order but reverses the Spec's maturity intent and silently inflates fleet `N` — worth a guard or a pinning test before the loop closes in M3.

### 1. Strengths
- **ARCH-PURE, exemplary.** `forecast.go` is genuinely pure — value inputs, deterministic output, no clock/fs/net — and every test in `forecast_test.go` runs on structs with zero mocks. All IO (baseline read, fleet walk, `computeBoard`) lives in `projectforecast.go` (package `main`) and is injected as `[]ProjectLoad`/`ThroughputBaseline` values. `today` is a parameter, not `time.Now()`. Clean seam.
- **ARCH-DRY, real consolidation.** `discover.go:78` factors the fleet walk into `walkFleetProjects`, and both `DiscoverByIssueRef` and the new `ListActiveProjectFiles` drive off it — "where the fleet's projects live" now has one source. `computeBoard`, `ParsePhaseA`, and `estimate.ParseBaselineTSV` are all reused rather than re-implemented.
- **`errNoBaseline` as a single fallback signal** (`projectforecast.go:34`): absence *and* parse failure both collapse to it, matching the Spec's "ledger unreadable → same fallback as no-baseline," and it's bubbled untouched so M3 consumers pick their own fallback.
- **Unknown load is never silently dropped** — weight 0 + `Warning` (`projectforecast.go:103`, `forecast.go:37`), exactly as the Spec demands ("a silent drop reads as 'no contention'").
- **The decision-table test pins arithmetic exactly** (`forecast_test.go:44` — 36h ÷ 55/3 = 1.9636wk → ceil 14 → today+14), not a tautology of the implementation.

### 2. Critical findings
None.

### 3. Important findings
- **`projectforecast.go:95` — a burned-down-but-open project reports its stale Phase-A total, inflating contention.** The board-vs-Phase-A predicate is `berr == nil && b.RemainingHours > 0`. A `committed`/`executing` project whose breakdown is *fully resolved but all-terminal* yields `RemainingHours == 0`, so it falls through to `ParsePhaseA` and reports the coarse PRD number (e.g. 12h) as remaining. That reverses the Spec's maturity logic ("issue estimates progressively replace the coarse Phase-A number as the project matures" — a *complete* project is maximally mature and should read ~0), and because such a load is `isActiveContention()`, it bumps `N` for **every** other project's forecast fleet-wide (`forecast.go:88`). The window is narrow (complete-but-not-yet-closed, since close archives), and the current plan's Step-3 wording ("board remaining … → Phase-A fallback") does sanction the simple ordering — so this is a judgment call, not a plan deviation. Fix sketch: gate the Phase-A fallback on "breakdown did not resolve into live rows" rather than "remaining ≤ 0" (e.g. only fall back when there are no resolved non-terminal frontier/blocked rows), or at minimum add a test pinning the intended weight-0 for an all-terminal open project so M3 doesn't cement the surprise.

### 4. Minor findings
- `projectforecast.go:120` — `Repo` is set to `meta.Name` (project name) for the subject but to the repo basename (`f.Repo`) for siblings. Currently unused by `ComputeForecast`/`RenderForecast`, so harmless today, but the field's meaning is inconsistent and M3 rendering may read it.
- `projectforecast.go:60-62` — `ListFleetProjects` returns `nil` on a fleet-walk error (unreadable parent dir), silently degrading the forecast to solo (`N=1`). Contrast with the deliberate per-file "never silently dropped" handling just below; a one-line `log`/note would keep the two consistent. Edge-only (parentDir is normally `..`).
- `listfleet_test.go:51` — the doc comment on `TestListActiveProjectFiles_NoParent` describes an `ExcludeByRealPath` symlink test, but the body tests a missing parent dir. Copy-paste comment mismatch — and, relatedly, the `EvalSymlinks`-based subject exclusion (`discover.go:165`) that the comment claims is covered is not actually exercised by any test.
- `projectforecast.go:82,116` + `projectstatus.go:85` — `d.Metadata()` is decoded up to three times per project (`forecastForProject` → `projectLoadFromDoc` → `computeBoard`). Cheap, not a hot path; noting only for tidiness.

### 5. Test coverage notes
Pure core is well covered (solo, multi-active, paused, unknown-other, zero-remaining error, zero-baseline error, render variants). Gaps, all Minor: (a) `RenderForecast`'s "under" / "on time" / "unparsable-deadline" branches (`forecast.go:121-127`) — only "over" is asserted; (b) a non-contending non-paused *other* (`done`/`defined` with a load) is never fed to `ComputeForecast` to confirm it's ignored; (c) the Important edge above (all-terminal open project → weight-0) has no pinning test; (d) the symlink exclusion path is untested (see Minor).

### 6. Architectural notes for upcoming work
- The three M3 consumers (`set-status →committed`, `show`/`status`, `close`) all funnel through `forecastForProject`, which returns `(Forecast, deadline, error)`. Note that `forecastForProject` returns a **raw `Metadata()` error** (not `errNoBaseline`) at `projectforecast.go:117-118`; M3 consumers must not assume every non-nil error is `errNoBaseline` when choosing the legacy-`--reality` vs quiet-line fallback.
- The `Repo`-field inconsistency (Minor above) is worth settling before M3 renders it.
- ARCH-PURPOSE for M2 is satisfied — this milestone is legitimately the reusable core, not a deferred point-of-the-issue. The shadow-sweep (does every consumer derive from `RenderForecast`?) applies at the M3 boundary, where all three surfaces must print the identical statement.

### 7. Plan revision recommendations
None required — the Plan's Core-concepts table matches the shipped code at every path, and the M2 checkboxes correspond to delivered entities. If the operator agrees the Important finding should change behavior (rather than stay plan-sanctioned), add a one-line `## Revisions` note to the issue recording that the Phase-A fallback fires only when the breakdown is unresolved, not merely when remaining is zero — so M3's `projectLoadFromDoc` contract is unambiguous.
