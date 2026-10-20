---
collaboration-mode: discussion
---

# Landing trains: wall-clock departure slots, N-or-Y, no clock reset — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A repo's landing trains leave on a wall-clock grid (`hh:00 + offset`, every interval), with the offsets planned machine-wide from suite cost so heavy suites don't collide. A train also leaves early when N cars are ready or the oldest has waited Y minutes. Lone and granted landings never move the grid.

**Architecture:** The departure decision stays one pure function in `cmd/sdlc/internal/landingline` (`DepartureDue`). Its inputs change from "the last `NextDue`" to: the slot grid (interval plus the repo's offset), the departure time of the last real train, and the ready waiters. The offset comes from a machine-wide plan file next to the suite lock, read with the same 30 s cache as the repo's schedule. `sdlc landing plan` shows the plan and, with `--write`, recomputes it from measured suite cost (`test-stats` p90, suite weight). A repo with no plan entry keeps today's clock.

**Tech Stack:** Go; the existing landingline package and its injected clock (`landingNow`, `Options.Now`); test-stats store.

**Source:** operator decision "option A" (ops `tl-status-xian-xu.side/2026-10-19-11-landing-schedule.md` §3 and the adaptive-departure note). Code map: the explore digest in `## Log` (2026-10-20).

**Edges:**
- ariadne#480 (ariadne:1) changes per-suite cost (the shard count). The plan reads measured p90, so a recompute picks it up and nothing here depends on #480.
- ariadne#473 M2 (ariadne:2) decides what each car must prove before it boards. It is orthogonal: this issue only decides *when* a train leaves.

---

## Today (from the code map)

- `NextDue(departed, s, leftReady)` gives `departed + interval`, or `departed` when ready cars were left behind (`train.go:72`). It is persisted on every `EventTrain` (`log.go:57`).
- `DepartureDue` takes the last `EventTrain`'s `NextDue` (`train.go:42-67`).
- **The reset bug:**
  - `maybeConductTrain`'s `serial()` (`landingtrain.go:667-670`) records an `EventTrain` (Result `alone`) with a fresh `NextDue` for every lone landing.
  - It does the same for granted landings (`:681`), own-gate refusals (`:685`), and "no other car ready" (`:706`).
  - So a lone or granted landing pushes the next train a whole interval out. The comment at `train.go:40` says a grant is not a departure; the code disagrees.
- Config `.sdlc/landing.json` has `train_interval_minutes` and `train_max_cars` (default 8); unknown keys are tolerated.

## Decisions (flag any to change)

- **D1. The grid never shifts.** Slots are `epoch_hour + offset + k·interval`. An early departure (N cars or Y minutes) does not move later slots. This is the strongest reading of "no clock reset", and it keeps the machine plan's spacing true. (The issue's parenthesis, "the slot then shifts to the next free offset", reads to me as describing which slot the *next* train takes, and that is still the grid's next slot.)
- **D2. Only a conductor's train is a departure.** `alone` records (lone, granted, refused-self, no-adapter) keep being logged for history but carry no `NextDue`, and the due check ignores them.
- **D3. Y is new config:** `train_max_wait_minutes` in `.sdlc/landing.json` (additive; 0 or absent means no bound). It is measured from the oldest current ready car's first vouch.
- **D4. The plan file** is `${XDG_STATE_HOME:-~/.local/state}/ariadne/landing-plan.json`, beside `test-suite.lock` (resolved like `scripts/test-lock.py`, honouring `WF_TEST_LOCK_FILE`'s directory). It is per machine and owned by the steward; `sdlc landing plan --write` is the steward's verb.
- **D5. The cost basis** is suite p90 wall time from test-stats (default `--since 7d`) times the declared suite weight (`local-checks.json` `weight`, default 1).
  - Packing is greedy: heaviest first, each at the earliest 5-minute offset where the summed weight of the suites overlapping its p90 window stays ≤ the machine's suite slots (`WF_TEST_SUITES`, default 2).
  - A repo with no samples is planned at its declared weight × 15 minutes.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Slots` (offset + interval → next/prev slot) | `cmd/sdlc/internal/landingline/slots.go` | new |
| `DepartureDue` (slot / full / max-wait / left-ready) | `cmd/sdlc/internal/landingline/train.go` | modified |
| `NextDue` / `NextDeparture` (trains only) | `cmd/sdlc/internal/landingline/train.go` | modified |
| `MachinePlan` (decode, validate, offsets) | `cmd/sdlc/internal/landingline/plan.go` | new |
| `PackOffsets` (costs → offsets) | `cmd/sdlc/internal/landingline/plan.go` | new |
| `Schedule.MaxWait` | `cmd/sdlc/internal/landingline/train.go` | modified |

- **Slots**: `Slots{Interval, Offset}`. `Prev(t)` is the latest slot ≤ t, and `Next(t)` the first slot > t; slots are anchored at the Unix-epoch hour and taken modulo the interval, so every process computes the same grid. `Offset == nil` means no plan entry: the caller keeps today's "last departure + interval".
- **DepartureDue(now, s, events, waiters)** returns `(due bool, why Reason)` with why ∈ {off, first, slot, full, max-wait, left-ready, next(t)}:
  - `left-ready`: the last train left ready cars behind (unchanged from today).
  - `full`: ready waiters ≥ MaxCars.
  - `max-wait`: oldest ready ≥ MaxWait.
  - `slot`: a slot passed since the last train departure (`Prev(now) > lastDeparture`).
  - else `next(min(Next(now), oldest+MaxWait))`.
  - Ready waiters are the line's vouched tickets; the input already reaches `trainAdmission`.
- **MachinePlan** is JSON `{"version":1,"slots":2,"repos":{"<repo>":{"offset_minutes":N,"basis":"…"}}}`. Unknown members are kept on decode and written back by `--write` (ARCH-EVOLVE). An offset outside 0..interval-1, or a bad version, is an error that names the file. A reader that fails falls back to today's clock, with one warning per TTL.
- **PackOffsets(costs []RepoCost, interval, slots)** gives offsets. It is deterministic: ties go by repo name.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `readMachinePlan` (+ 30 s cache) | `cmd/sdlc/landingconfig.go` | new | the plan file |
| `trainAdmission` | `cmd/sdlc/landingcar.go` | modified | passes waiters + plan offset |
| `maybeConductTrain.serial` | `cmd/sdlc/landingtrain.go` | modified | records `alone` without NextDue |
| `sdlc landing plan [--write]` | `cmd/sdlc/landingcmd.go` | new | test-stats store, local-checks weight, the plan file |
| `sdlc landing status` | `cmd/sdlc/landingcmd.go` | modified | shows the slot and the reason |

No external service is involved. The clock is the existing injected `landingNow` / `Options.Now`. The plan file and test-stats are portable folders that tests build in `t.TempDir()` (via `XDG_STATE_HOME`, and the store dir under the fixture's git common dir), so ARCH-MOCK is met by real files.

### State design table (per repo's departure)

Actions: a car vouches or leaves, `sdlc landing grant`, `sdlc landing plan --write`, editing `.sdlc/landing.json`.

| Internal state / event | Visible (`landing status`) | Recovery | Possible loss |
|---|---|---|---|
| no ready car | `next train at HH:MM (slot :MM)` | — | — |
| ready < N, oldest < Y, before slot | `next train at min(slot, oldest+Y)` | wait, or grant to go alone | — |
| slot passed since last train | `train due (slot :MM)` | the head conducts | — |
| ready ≥ N | `train due (full: N ready)` | the head conducts early; grid unchanged (D1) | — |
| oldest waited ≥ Y | `train due (max wait Ym)` | the head conducts early; grid unchanged | — |
| last train left ready cars | `train due (cars left ready)` | the head conducts now | — |
| granted / lone / refused-self landing | unchanged from before it | none needed: it no longer touches the grid (D2) | — |
| train abandoned or refused | as by slot/full/Y; released cars count as ready | — | the time of one train |
| plan file missing | `no machine plan: every <interval> after the last train` | `sdlc landing plan --write` | — |
| plan file unreadable or invalid | the same fallback, plus `[!] plan … (reason)` | fix or rewrite the file | — |
| config unreadable | `trains off (serial): <reason>` (unchanged) | fix main's config | — |

Covered: every departure trigger, both landings that used to reset the clock, and plan absence or corruption. Not covered: two machines landing the same repo; trains are one machine's, as today.

### Ordering (ARCH-ORDER)

The departure decision holds no state between events of its own: it is recomputed on every admission poll from the line's events log (the last train departure), the ready waiters and the clock. The only persisted change is that `alone` events stop carrying `NextDue`.

- **Most likely to be mishandled:** two heads racing at a slot boundary. That is already serialised: `tryAdmit` takes `head.lock` and reads events under it, and the train's own `EventTrain` record ends the window, so the second head sees `Prev(now) ≤ lastDeparture`.
- **Clock skew between processes** on one machine: none (one clock). An NTP step backwards can only delay one slot: `Prev(now) > last` is false until the clock passes it again.
- **An old binary reading new events** (`alone` without NextDue): `NextDeparture` already skips a zero NextDue, so an old reader sees the last train's due time. That is safe.
- **A new binary reading old events** (an `alone` with NextDue): the trains-only filter ignores `alone` by Result.

### Operating envelope (ARCH-CONSTRAINTS)

- The admission poll runs every 2 s. It adds one cached plan read (30 s TTL, one small file) and an O(waiters) scan; waiters number at most a few dozen.
- `landing plan --write` is interactive: it reads test-stats per train-enabled repo (sequential, about 1 s each).

### Funeral (ARCH-FUNERAL)

- **The plan file:** one per machine. It is overwritten by `--write` and removed by hand when trains are retired; it never grows.
- **Events:** the existing events.jsonl, with 30-day retention; `alone` events get smaller (no NextDue).
- **Nothing else** new is durable.

### Trust (ARCH-SECURE)

The plan file is hand-editable, per-user state. It is parsed into typed values: offsets are range-checked, and unknown members are kept but never interpreted. An invalid plan degrades to today's clock, visibly. No credentials are involved.

## Milestones

- **M1: the departure rule.**
  - `Slots`, the new `DepartureDue` with its full, max-wait and slot reasons, and `train_max_wait_minutes`.
  - The `alone` records stop carrying `NextDue` (the reset fix), with the plan offset still nil, so today's clock remains when there's no plan.
  - `landing status` shows the reason.
  - Fake-clock fixtures in `internal/landingline/train_test.go` and an e2e grant test asserting that the next departure is unchanged.
- **M2: the machine plan.**
  - `MachinePlan`, `PackOffsets`, `readMachinePlan` wired into admission, and `sdlc landing plan [--write]` with its cost basis.
  - The two-repo fixture: their train suites never start within each other's expected window.
  - Help and atlas.

## Tasks

### M1

- [x] **T1 `Slots`** (`internal/landingline/slots.go`, `slots_test.go`):
  - a table of interval 60/30, offsets 0/15/45, times around each boundary (exactly on a slot counts as due);
  - property: `Prev(t) ≤ t < Next(t)` and `Next(t) − Prev(t) = interval`, for random t.
- [x] **T2 `DepartureDue`**:
  - Extend `Schedule` with `MaxWait` and `Offset *time.Duration`; pass the ready waiters (oldest vouch time and count).
  - Test each reason, the precedence (left-ready, then full, then max-wait, then slot), and that the grid is unchanged after an early departure: a train that left at :05 because it was full still has :15 as the next slot.
- [x] **T3 the reset fix:** `serial()` records `alone` with no `NextDue`, and `NextDeparture` / `lastTrainDeparture` consider only non-`alone` trains.
  - Unit test: events [train@:00, alone@:10 (grant)] give the next slot :00+interval, not :10+interval.
  - e2e: `TestLandingTrainGrantDepartsAloneWhenNoTrainIsDue` gains an assertion that the next departure is unchanged by the grant.
- [x] **T4 config:** `train_max_wait_minutes` (0..24*60) in `landingconfig.go`; `TestParseLandingConfig` gains its cases, including the unknown-key case already there.
- [x] **T5 `landing status`:** prints `train due (<reason>)` or `next train at HH:MM (<slot|max-wait>)`; golden test.
- [x] **M1 close:** targeted tests, then `sdlc milestone-close --milestone M1`.

### M2

- [x] **T6 `MachinePlan` and `PackOffsets`** (`internal/landingline/plan.go`, `plan_test.go`):
  - decode/encode round-trip that keeps unknown members, top-level and per repo;
  - invalid offset, version and JSON cases;
  - packing with today's measured costs (ariadne 9.5 min × 2, pair 12.5 × 1, parley 2.7 × 1) gives ariadne :00, pair :15, parley :15;
  - the overlap property for random costs: summed weight ≤ slots at every minute.
- [x] **T7 `readMachinePlan`** (`landingconfig.go`): path resolution (XDG, WF_TEST_LOCK_FILE), a 30 s cache that shares `landingScheduleSource`'s pattern, and a fallback with one warning.
- [x] **T8 admission:** `trainAdmission` takes the repo's offset from the plan. Fixture: two repos, offsets 0 and 15, interval 60, a fake clock stepping through an hour. Each repo's train becomes due only in its own slot, so their suites never start within each other's p90 window.
- [x] **T9 `sdlc landing plan [--write]`:** (revised in M2 review round 1)
  - It reads the train-enabled repos from the test-stats fleet discovery (`fleetRepoPaths`) (each repo's `.sdlc/landing.json` on its main), their test-stats p90, and their declared suite weight.
  - It prints `repo offset interval weight p90 basis`; `--write` replaces the file and keeps unknown members.
  - Golden test over a fixture fleet.
- [x] **T10 help and atlas:**
  - `helptext/landing.md`: slots, N-or-Y, the plan and its fallback.
  - `atlas/workflow/sdlc-binary.md`: the landing section.
  - tl-lead help: the steward runs `landing plan --write` after cost changes, for example once #480 lands.
- [ ] **T11 verify:**
  - Full suite, then close.
  - **Live:** `sdlc landing plan` dry run on this machine, sent to the hub. The hub writes the plan and watches one hour of departures, recording slot-wait p50/p90 and ready→landed before and after in test-economy.

## Done when

- Fake-clock fixtures prove:
  - two repos with offsets never start train suites within each other's expected window;
  - a full train leaves before its slot;
  - the oldest-car bound fires;
  - a granted (and a lone) landing leaves the next departure unchanged.
- `sdlc landing plan` prints the machine plan with each repo's offset and its cost basis; `--write` keeps unknown members.
- Without a plan entry a repo behaves exactly as today, except for the clock-reset fix, which applies everywhere.

## Revisions

### 2026-10-20: plan-quality round 1

**PQ-1: the admission seam changes.** `AdmissionRule` becomes `func(events []Event, t Ticket, r *Readiness, granted bool, w Waiters) (bool, string)`.
- `Waiters{Ready int; OldestSince time.Time}` is computed once per `tryAdmit` pass under `head.lock` (`line.go:262`). It counts only the vouches that are valid now (`r.For(t, now) == ""`, the check boarding uses), and takes the oldest `Since` among them.
- `trainAdmission` hands `w` to `DepartureDue`.
- Every rule implementation and test stub is updated in M1. They are found by the compiler, and the enumeration goes in the M1 log.

**PQ-2: the departure time is persisted.**
- `Event.Departed time.Time` (`json:"departed,omitzero"`) is added on `EventTrain`; `recordTrain` sets it to `c.departed`.
- `lastDeparture(events)` is the newest conductor train's `Departed`.
- An older event without the field falls back to its `At` (completion). That is conservative: for the single pre-upgrade record, at most one slot that fell during that train is absorbed.
- The change is additive (ARCH-EVOLVE): old readers ignore the field, and `NextDue` stays written for them.
- Test: a round-trip of an event with and without `departed`, plus the fallback case.

**PQ-3: early departures and the plan.**
- The machine plan schedules **slot** departures only.
- A full or max-wait departure leaves early by design (bounded waits beat strict spacing), and may overlap another repo's window. Such an overlap is bounded by the machine suite lock (2 slots): a suite that doesn't fit waits in the lock's FIFO, exactly as today. This is a stated rule, not a gap.
- The Done-when and T8 are narrowed to match: "two repos' **slot-triggered** departures never start within each other's planned window". A separate fixture shows that an early departure is admitted regardless of the other repo's window.

**Minor: a lone landing that left ready cars** keeps its "due at once" signal.
- An `alone` record carries `NextDue` only when it left ready cars (`NextDue = departed`, never `departed + interval`). The left-ready check reads the newest `NextDue` of any kind.
- So a lone landing can only ever bring the next train *forward*, never push it back.
- Test: [train@:00 (next :60), alone@:10 left-ready] is due at :10; [train@:00, alone@:10 grant] is next at :60.

**Minor: D3's first vouch.**
- `Readiness.Since time.Time` (`json:"since,omitzero"`, additive; `Extra` already preserves unknown members) is set when a car first vouches ready at a head. It is kept across re-vouches at the same head and reset when the head moves or the vouch turns not-ready.
- Old sidecars without it fall back to `At`.
- Test: re-vouch keeps `Since`; a head change resets it; decoding an old sidecar works.

**Minor: test prose.** T1 and T6 are compressed to one strategy each:
- `Slots`: a table at the boundaries, plus the `Prev ≤ t < Next` property;
- `PackOffsets`: the measured-cost golden plus the overlap property.

**Done when (revised):**
- Fake-clock fixtures show four things:
  - two repos' slot-triggered departures never start within each other's planned window;
  - a full train leaves before its slot;
  - the oldest-car bound fires;
  - a granted or lone landing leaves the next departure unchanged, unless it left ready cars, which brings it forward.
- `sdlc landing plan` prints each repo's offset and cost basis, and `--write` keeps unknown members.
- Without a plan entry, a repo behaves as today except for the reset fix.

### 2026-10-20: M2 review round 1

**BR-8 (the rule: every consumer gets a repository's effective schedule from one constructor that combines setting and plan).** `effectiveSchedule(git, mainRef, repo)` is that constructor (`landingconfig.go`). The sweep of `readLandingSchedule` callers:
- `landingScheduleSource` (admission) and `landing status` now use it; status also names the grid in use ("departure grid: …").
- `fleetPlanInputs` keeps the bare setting, because the plan is what it computes.

**BR-9 (the rule: a task is ticked only when every artifact it names exists as named).**
- T7: the `WF_TEST_LOCK_FILE` and XDG paths are now tested (`TestMachinePlanPath`).
- T9's text is revised:
  - The source is `fleetRepoPaths`, the test-stats fleet discovery split out in M2, not `sdlc fleet inventory`. It is the same set of repositories test-stats measures, so the cost basis and the repository list come from one discovery.
  - `fleetPlanInputs` takes its sources injected and is tested with fakes (`TestFleetPlanInputs`): the trains-on filter, the declared weight, the no-sample default, and the name join.
- The live dry run is T11's, after landing; it was claimed in the issue's M2 line, which is corrected.

**BR-10 (the rule: every new subcommand, flag, config key, env var or state path gets a README line in the same range).** The README now has `sdlc landing plan [--write]`, the plan file path, `WF_TEST_LOCK_FILE`, `WF_TEST_SUITES` and `train_max_wait_minutes` (M1).

**BR-11.** `PackOffsets` returns each placement that still exceeds the slots, with its peak; `landing plan` warns, naming the repository; `TestPackOffsetsNamesAnOverload`.

**Minors:**
- Every landing test gets an isolated plan path through the shared `landingFixture` (`isolateMachinePlan`).
- The plan decoder refuses a repeated key at every level.

### 2026-10-20: M2 closed (SHIP). The plan now matches the code

Rule (family "plan claim matches delivery"): every behavioural sentence in Core concepts names its enforcing test or code, or is revised here.

- **MachinePlan offsets:** the decoder refuses an offset outside 0..1439 (`decodeEntry`, `TestMachinePlanCodec`). `Offset(repo, interval)` then takes it modulo the interval, so a plan written for a longer interval still places the repository, and never refuses (`TestMachinePlanCodec`, the 10-minute case). This replaces "outside 0..interval-1 is an error".
- **Plan read failure:** warned once for the life of a schedule source (one merge process), not once per TTL (`TestScheduleSourceReadsThePlan`). `landing status` names the failure on every run.
- **Sweep of the other Core-concepts behaviour sentences:**
  - `Slots` Prev/Next (`TestSlots`).
  - The `DepartureDue` reasons and precedence (`TestDepartureDueEarlyReasons`, `TestNextDepartureReadsTheDeparture`).
  - Packing determinism and the overload (`TestPackOffsets`, `TestPackOffsetsNamesAnOverload`).
  - `Since` across re-vouches (`TestVouchKeepsSince`, `TestStaleSinceFromAnOlderSdlc`).
  - Every remaining Core-concepts behaviour sentence names its test above or in the earlier Revisions.
