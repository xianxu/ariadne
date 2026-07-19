---
id: 000182
status: working
deps: [ariadne#180]
github_issue:
created: 2026-07-16
updated: 2026-07-19
estimate_hours: 2.65
started: 2026-07-18T22:35:59-07:00
---

# project calendar estimator: mechanize the commit reality-check (effort→calendar bridge)

## Problem

A project's load-bearing attribute is its `deadline:` — a DATE — but both
estimators produce HOURS: per-issue estimate-logic-v3.1 (Phase B) and #180's
Phase-A PRD-level estimator (workstreams × fog). Nothing bridges effort to
calendar. The bridge currently lives in exactly one place: the `reality-check`
guard at project commit (defined→committed), which as designed in #180 is an
evidence flag — the operator types `--reality "fits July"` and the gate takes
their word. A mandate, not a mechanism — off-brand for the vocabulary lift,
whose whole point is computed gates over attestation.

## Spec

*(Settled at the 2026-07-18 brainstorm — supersedes the 2026-07-16 seed. One
material change from the seed: the reality-check INFORMS, it never BLOCKS —
"estimation is a funny business, often wrong; when behind we work overtime,
shift resources. Always track and inform the operator" (operator). See
`## Revisions`.)*

### Layering: issue facts, project math

Projects don't burn down directly — issues do. The forecast divides two
numbers, and issue-level data supplies both:

- **Numerator — this project's remaining hours:** Σ `estimate_hours` over
  unfinished breakdown rows (`computeBoard.RemainingHours`); before the
  breakdown resolves, fall back to the Phase-A total (#180) with a note in
  the statement. Issue estimates progressively replace the coarse Phase-A
  number as the project matures.
- **Denominator — throughput (hours/week):** measured from the calibration
  ledger (`brain/data/life/42shots/velocity/calibration-ledger.tsv`, one row
  per issue close): Σ `actual` per week over the blessed span.

**Unit identity (load-bearing):** both sides are SHIP WALL-CLOCK hours for
one engineer + AI (#118). Weekly Σ actual can exceed 168h because concurrent
agent sessions overlap wall-clock — the unit is *issue-hours produced*, not
human attention-hours. The division is sound because numerator and
denominator share the unit; we never convert to human hours. (Consequence:
parallelism is already priced into measured throughput, so the attention
ceiling is a warning, not arithmetic.) MVP uses issue estimates raw; the
est/actual ratio bias (~0.4–1.7× recently) is a future refinement once the
calendar ledger shows it matters.

### Blessed throughput baseline (new, in brain)

Trailing windows skew under vacations and life events (measured weekly
volatility: 14 → 56 → 153 → 91 → 139h across the last 6 active weeks). The
operator instead **blesses a representative span**; the machinery measures
the number — operator picks the SPAN, never types the RATE:

- `sdlc project throughput --bless 2026-06-16..2026-07-13` reads the ledger,
  computes Σ actual ÷ span-weeks, and APPENDS to
  `brain/data/life/42shots/velocity/throughput-baseline.tsv`:
  `blessed_date  span_start  span_end  hours_per_week  rows  ceiling`.
  Last row = current baseline; append-only = free history of re-blessings.
  `ceiling` defaults to 2 (#117), settable at bless time. Refuses an empty
  span (no ledger rows).
- Bare `sdlc project throughput` prints the current baseline + a trailing-4-
  week comparison, so staleness/divergence is visible on demand. The gate
  never auto-substitutes the trailing number.
- Same write-to-brain class as the calibration ledger (measurement — the
  #171 residency charter's brain).
- Test override: `WF_THROUGHPUT_BASELINE` env (mirrors `WF_CALIB_LEDGER`).

### Pure core

`internal/project/forecast.go` — `ComputeForecast(baseline, thisRemaining,
others []ProjectLoad, today) Forecast`, no IO:

- `n` = this project + others with status `committed`/`executing` (fleet-wide;
  enumerate via a small `ListFleetProjects` reusing #171's sibling walk; each
  project's remaining via its own `computeBoard` from its own repo vantage).
- share = `hours_per_week ÷ n`; projected finish = today + remaining ÷ share
  weeks. Auditable one-line arithmetic — no hidden knobs.
- **Ceiling = warning threshold, not arithmetic:** `n > ceiling` adds
  "n active projects exceed your ~2-session attention ceiling — forecast
  degrades".
- **Paused projects: weight 0, named risk lines** ("paused: metis-v2, 14h
  remaining — resuming it invalidates this forecast").
- Output struct: projected date, the arithmetic trail (n, share, remaining,
  source of remaining), divergence note, ceiling warning, risk lines.
  `RenderForecast` produces the one-paragraph statement every surface prints
  identically.

### Consumers — inform, never block

1. **`set-status →committed`:** compute → print → derive `planned_finish:` →
   record the rendered statement in the Log as the reality-check evidence.
   The guard passes on HAVING COMPUTED, never on the answer (feasible or
   not). Only when computation is impossible (no blessed baseline) does it
   fall back to requiring the legacy `--reality` prose — a process fallback,
   not a feasibility gate. Explicit `--planned-finish` overrides the derived
   date and is recorded as a manual override.
2. **`project show` / `project status`:** recompute live; render forecast-
   vs-deadline drift ("at baseline pace you land Oct 3; deadline Sep 1").
3. **`project close`:** append planned-vs-actual finish + slip days as a
   calendar calibration row beside the fog-factor row (exact file/columns
   decided at plan time after reading projectclose.go's ledger shape). This
   creates the project-level feedback loop that issue-level est/actual
   already has.

Future roadmap rollup (#15/#185 territory) = "call `ComputeForecast` per
project and sum" — deliberately nothing reserved for it.

### Errors

No baseline → legacy `--reality` fallback + bless hint. Empty bless span →
refuse. Unresolvable issue estimates → Phase-A fallback, noted in the
statement. Ledger unreadable → same fallback as no-baseline.

### Out of scope

Roadmap-level rollup, per-render forecast history, multi-operator
throughput, ratio-corrected numerator.

## Done when

- `set-status →committed` prints a computed feasibility statement
  (throughput ÷ contention → projected finish vs proposed deadline) and
  records it as the reality-check evidence; it INFORMS and never refuses on
  the answer (`--reality` prose survives only as the no-baseline fallback).
- `planned_finish:` derives from the computation (manual override recorded
  as such).
- Throughput comes from measured ledger data over an operator-BLESSED span
  (operator picks the span, never types the rate); baseline stored in brain
  with provenance.
- `project show`/`status` surface the live forecast-vs-deadline drift.
- Project close records planned-vs-actual finish as a calibration row.

## Revisions

### 2026-07-18 — brainstorm settles the four forks; inform-not-gate

Done-when amended from the seed: the computed statement no longer *replaces*
`--reality` as a gate — it never blocks at all (operator direction: track
and inform; slippage is recoverable by means the math can't see). Forks
settled: blessed-span baseline over trailing window (volatility evidence);
contention = committed+executing, paused listed at weight 0; ceiling =
constant 2 in the baseline record, warning-only; `--reality` survives as the
no-baseline process fallback. (Restored 2026-07-19: an editing slip dropped
this section when the Spec was rewritten; caught by the plan review.)

## Estimate

Derived from the durable plan's three-milestone decomposition (design
pre-resolved by the spec+plan → items carry only the residual design cost;
impl at v3.1's 40% scale of the v2 primitive table; +0.15 flat design buffer).
One `milestone-review` per boundary (three auto-dispatched fresh-context
reviews). M3 carries the heavier impl (three consumers + the
`applyProjectStatus` signature threading).

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: smaller-go-module         design=0.05 impl=0.20
item: smaller-go-module         design=0.05 impl=0.20
item: greenfield-go-module      design=0.20 impl=0.30
item: smaller-go-module         design=0.05 impl=0.20
item: cross-cutting-refactor    design=0.10 impl=0.20
item: smaller-go-module         design=0.05 impl=0.25
item: atlas-docs                design=0.05 impl=0.15
item: milestone-review          design=0.0  impl=0.15
item: milestone-review          design=0.0  impl=0.15
item: milestone-review          design=0.0  impl=0.15
total: 2.65
```

Item→milestone map: M1 = `smaller-go-module` (pure span throughput) +
`smaller-go-module` (throughput verb); M2 = `greenfield-go-module` (forecast
core) + `smaller-go-module` (fleet load assembly); M3 =
`cross-cutting-refactor` (commit hook + `applyProjectStatus` threading) +
`smaller-go-module` (show/status/close surfaces) + `atlas-docs`; plus three
`milestone-review`.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only. (Source [stale] per estimate-source
— treat per-primitive hours as provisional, #127.)*

## Plan

- [x] M1 — throughput measurement: pure span math over the existing ledger parser + `sdlc project throughput` (`--bless`, append-only baseline TSV in brain)
- [x] M2 — pure `ComputeForecast`/`RenderForecast` core + fleet load assembly (`ListFleetProjects` reusing #171 sibling walk)
- [x] M3 — three consumers: commit-transition forecast (informs, derives `planned_finish`), show/status live drift, close calendar-ledger row + atlas

## Log




- 2026-07-19: closed M3 — go test ./cmd/sdlc/... ./pkg/vocab/ -count=1 green. Commit hook: TestSetStatusCommit_{ComputesAndDerivesPlannedFinish,PreExistingPlannedFinishKept,ExplicitPlannedFinishWins,NoBaselineRefusesWithHint,NoBaselineWithRealityPasses} + TestSetStatusNonCommit_Untouched (informs-never-blocks; planned_finish precedence manual>existing>derived; derived applied before baseline-set guard). Surfaces: TestProjectShow_IncludesForecast, TestProjectStatus_IncludesForecast, TestProjectShow_NoBaselineQuietLine. Calendar ledger: TestPrepareCalendarLedgerRow{,_NoPlannedFinish,_MissingHeadingErrors}, TestAppendProjectLedgerRow_FogHeadingUnchanged (fog regression). Vantage bug (verification-caught) fixed + pinned: TestForecastForProject_RelativePathResolvesVantage. Live: sdlc project show on the real tracking project → 2.6h ÷ 110.6 h/wk forecast. Atlas + helptext updated; Calendar ledger heading added to real brain file. actual 0.9h = M3 increment (3 consumers + vantage fix); review verdict: FIX-THEN-SHIP
- 2026-07-19: M3 FIX-THEN-SHIP resolved (bundled per #174, no re-close): Important 1 — the calendar-row wiring wasn't asserted end-to-end (deleting prepareCalendarLedgerRow would pass every test); added a calendar-row assertion to TestProjectCloseRecordsFogAndArchives (planned 2026-07-30, closed 2026-07-16 → -14 slip). Important 2 — README not extended for the forecast surfacing; added the commit/--planned-finish/show lines under the calendar block. Minor — computeCommitForecast now uses the deadline returned by forecastForProject instead of re-reading d.FM (consistent with forecastLine). Left as noted: slipDays/dayDelta cross-package dup (consolidate on 3rd copy), double doc-read on cold CLI path, two unreachable defensive branches. Plan Revisions records the plannedFinishDecision package placement + the 4 execution deltas. Re-ran go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ -count=1 — green.
- 2026-07-19: closed M2 — go test ./cmd/sdlc/... ./pkg/vocab/ -count=1 green. Pure forecast core: TestComputeForecast_{Solo,TwoActiveOthers,PausedExcludedButListed,UnknownOtherWeightZero,ZeroRemainingErrors,ZeroBaselineErrors} (exact date arithmetic: 55h/55hpw/n1=+7d, 36h/55hpw/n3=+14d), TestRenderForecast_{FullStatement,NoDeadline,PhaseAFallbackNoted}. Fleet assembly: TestListActiveProjectFiles (shared walk, subject excluded, archive/.bak skipped, DiscoverByIssueRef regression green), TestListFleetProjects (board/phase-a/unknown sources), TestLoadThroughputBaseline_{EnvOverride,AbsentIsErrNoBaseline,UnparsableIsErrNoBaseline}, TestForecastForProject_{NoBaseline,WithBaseline}. Atlas: forecast section added (throughput baseline, pure core, fleet assembly). actual 0.6h = M2 increment; review verdict: FIX-THEN-SHIP
- 2026-07-19: M2 FIX-THEN-SHIP resolved (bundled per #174, no re-close): Important — a fully-burned-down-but-open project fell through to its stale Phase-A total (predicate was board `RemainingHours > 0`), reading e.g. 12h remaining and inflating N for every fleet forecast. Two-part fix: (1) `boardRowsResolved` — the board is authoritative once the breakdown resolves into ANY issue row, even at 0 remaining (a complete project is maximally mature, reads ~0, not the coarse PRD number); Phase-A is a fallback only when nothing resolved; (2) `isActiveContention` now also requires `RemainingHours > 0` so a 0-remaining project consumes no throughput. Pinned by TestListFleetProjects_AllTerminalReadsBoardZero (board/0, N stays 1) + TestComputeForecast_ZeroRemainingOtherDoesNotContend. Minors: subject `Repo` now the repo basename (consistent with siblings, was project name); ListFleetProjects warns on a fleet-walk error instead of silent solo-degrade; copy-paste test comment fixed + the EvalSymlinks subject-exclusion now actually exercised (TestListActiveProjectFiles_ExcludeResolvesSymlink). Left as noted: 3× Metadata decode per project (cheap, not hot). Re-ran go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ green.
- 2026-07-19: closed M1 — go test ./cmd/sdlc/... ./pkg/vocab/ -count=1 green. Pure: TestSpanThroughput_{28DaySpan,PartialWeek,EmptySpan,BadSpanBounds} (untrusted counted, bad-date skipped, days/7 partial weeks), TestBaselineTSV_RoundTrip + BadFloat + Empty. Verb: TestProjectThroughput_{Registered,Bless,BlessAppends,EmptySpanRefuses,BareShowsCurrentAndTrailing,NoBaselineBareHints}. Live: sdlc project throughput --bless 2026-06-22..2026-07-19 → 110.60 h/wk (280 rows), bare form shows trailing-4wk -3.81 delta; baseline written to brain velocity dir. actual 0.5h = M1 increment (2 tasks: pure span math + verb); review verdict: FIX-THEN-SHIP
- 2026-07-19: M1 FIX-THEN-SHIP resolved (bundled per #174, no re-close): Important — README project quick-start omitted the new `throughput` verb (Docs gate); added `--bless` + bare forms. Minor — `showThroughput` swallowed ALL baseline read errors as "not blessed yet"; now distinguishes os.IsNotExist (informational) from a real IO error (surfaced). Left as noted: isoLayout/isoDate cross-package const dup (pre-existing ~15× pattern, not a #182 regression), --ceiling no-op in show mode (documented). Re-ran go build ./... && go test ./cmd/sdlc/... green.
### 2026-07-16

Filed from the #180 plan-approval session. Operator: "we need a higher level
time estimator" — the effort→calendar gap identified while reviewing #180's
two-phase (hours-only) estimation design. Kept out of #180 scope deliberately:
needs its own design and consumes #180's model the same way #171 does.

Same day, operator placed this IN the project-management-primitive MVP
(initially proposed as post-MVP): "that is the key differentiator between a
project and an issue, the timeline aspect, of managing something higher
level, longer running." The taxonomy consequence for design: the calendar
computation isn't an optional refinement of the reality-check — it IS the
project noun's defining capability; without it a project degenerates to an
issue container with a hand-attested date.
