---
id: 000475
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-20
estimate_hours: 2.67
collaboration-mode: collaboration
card_mirror: 'b41c86ee837a1928e35246bab104bceb37595ada' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T03:03:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A)

## Problem

Landing trains (#449) depart per repo on independent clocks: a train leaves `train_interval_minutes` after the previous one,
takes up to `train_max_cars`. Nothing coordinates repos on one machine, so heavy suites collide at random in the 2-slot suite
budget (ariadne's suite takes both slots; measured 7-day medians: ariadne 9.5 min, pair 12.5, parley 2.7; slot waits 11–19 min).
A full train waits for the clock even when N cars are ready, and a lone or granted landing resets the train clock
(`landingtrain.go` ~l.661, `NextDue = now + interval`), pushing the next train a whole interval back. Operator discussion and
decision (option A, 10-19): ops `teams/devinfra/tl-status-xian-xu.side/2026-10-19-11-landing-schedule.md`.

## Spec

- **Wall-clock departure slots, planned machine-wide.** Each train-enabled repo departs at `hh:00 + offset (mod interval)`.
  Offsets come from one machine-wide plan file next to the suite lock (the steward's to change), computed from measured suite
  cost (`sdlc test-stats` p50/p90 and suite weight) so heavy suites don't overlap: e.g. ariadne :00 (weight 2), pair and parley
  :15 together, the rest free for branch/affected suites. A repo without a plan entry keeps today's clock (back-compatible,
  ARCH-EVOLVE: readers tolerate unknown fields). `sdlc landing plan` shows and recomputes it.
- **N-or-Y departure:** a train also leaves early when `train_max_cars` cars are ready; and a car's wait is bounded: a train
  leaves when the oldest ready car has waited `train_max_wait_minutes`, whichever comes first (the slot then shifts to the next
  free offset).
- **Fix the clock reset:** lone and granted landings don't move the next train's departure.
- Measured outcome: slot-wait p50/p90 and ready → landed before/after (test-stats).

## Done when

- Fixtures (fake clock): two repos with offsets never start train suites within each other's expected window; a full train
  leaves before its slot; the oldest-car bound fires; a granted landing leaves the next departure unchanged.
- `sdlc landing plan` prints the machine plan with each repo's offset and its cost basis.

## Plan

Durable plan: [workshop/plans/000475-landing-train-departure-slots-plan.md](../plans/000475-landing-train-departure-slots-plan.md). It has decisions D1–D5, a state table, and ARCH notes.

- [x] M1 — departure rule: wall-clock `Slots`, `DepartureDue` with full/max-wait/slot reasons, `train_max_wait_minutes`, and the clock-reset fix (`alone` records carry no NextDue), plus the reason in `landing status`
- [x] M2 — machine plan: `MachinePlan` + `PackOffsets` (p90 × weight, ≤ 2 suite slots), plan file next to the suite lock, offsets wired into admission and status (`effectiveSchedule`), `sdlc landing plan [--write]`, help, README and atlas
- [ ] Live, after landing: `sdlc landing plan` dry run sent to the hub; the hub writes the plan and records slot-wait and ready→landed before and after

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec           design=0.6 impl=0.06
item: smaller-go-module    design=0.05 impl=0.1
item: smaller-go-module    design=0.2 impl=0.3
item: greenfield-go-module design=0.3 impl=0.35
item: atlas-docs           design=0.1 impl=0.1
item: milestone-review     design=0.0 impl=0.16
item: milestone-review     design=0.0 impl=0.16
design-buffer: 0.15
total: 2.67
```

Items, in order:
1. the spec and plan, with two plan-quality rounds;
2. `Slots`;
3. M1's departure rule: `DepartureDue`, the admission seam (`Waiters`), `Readiness.Since`, `Event.Departed`, and the reset fix;
4. M2's machine plan (greenfield): `MachinePlan`, `PackOffsets`, the plan file read, and `sdlc landing plan`;
5. help and atlas;
6. two milestone reviews.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-10-19

### 2026-10-20
- 2026-10-20: closed M1 — M1 departure rule with round-1 fixes: Slots grid; DepartureDue reasons (slot/interval/max wait/full/cars left ready) shown by landing status (TestLandingStatusShowsTrainsAndReadiness); Departed read at decision time so an offset change applies at once and old-format alone records never schedule (TestNextDepartureReadsTheDeparture); reset fix (grant e2e fails with the old serial NextDue); Since with SinceHead guards against stale carry (TestStaleSinceFromAnOlderSdlc, TestVouchKeepsSince); README and landing help updated; atlas updated. landingline and landing sdlc tests pass.; review verdict: SHIP
- Claimed and started from ariadne:4. Parked for #490 at the hub's request, then resumed.
- Code map (explore digest), kept for the plan:
  - The only schedule is the last `EventTrain`'s `NextDue` (`train.go:58-77`).
  - `maybeConductTrain.serial()` (`landingtrain.go:667-670`) records `alone` with a fresh NextDue for lone, granted, refused-self and no-adapter landings: that is the reset.
  - Clock seams: `landingNow` and `Options.Now`.
  - Config: `train_interval_minutes` and `train_max_cars` only; unknown keys are tolerated.
  - The suite lock lives in `${XDG_STATE_HOME:-~/.local/state}/ariadne/`.
  - `test-stats` gives p50/p90 wall time per repo but no weight; the weight comes from the local-checks declaration and the make probe.
- The hub approved the plan; D1–D5 confirmed. It noted that the plan must re-read the weight, since #480 may change ariadne's shard count: `landing plan --write` reads the declared weight live, and the 30 s cache picks up a rewritten file. Only pair has trains on today.
- **M1 implemented:**
  - `Slots` (epoch-anchored grid).
  - `DepartureDue(now, s, events, Waiters)`, with the full and max-wait reasons ahead of the time check; the reported next time is min(slot, oldest + Y).
  - `NextDue` on the grid when a plan offset is set; `AloneNextDue` gives the reset fix (an `alone` record carries NextDue only when it left cars ready).
  - `Event.Departed`; `Readiness.Since`, through `carrySince` in `Vouch`.
  - `ReadyWaiters`, computed once per `tryAdmit` pass.
  - `train_max_wait_minutes`; `landing status` shows the decision and its reason.
- **`AdmissionRule` implementations updated** (the compiler's enumeration): `trainAdmission`'s two closures (`landingcar.go`); the stubs in `internal/landingline/train_test.go` (`TestAdmissibleGatesTheHead`, `TestAdmissionSeesTheTrainTheLastHeadRecorded`); and the calls in `TestTrainAdmission`.
- **Red-checked:** the grant e2e fails with the old `serial()` NextDue. `TestLandingTrainOfOneIsTodaysLanding` now asserts that a lone landing has no NextDue and a Departed time.
- **M1 review round 1 (FIX-THEN-SHIP), all six fixed in this round:**
  - BR-1: the README and `landing --help` state the new departure rule and `train_max_wait_minutes`.
  - BR-2: `DepartureDue` names its reason (slot, interval, max wait, full, cars left ready) with a dated time, and `landing status` prints it; `TestLandingStatusShowsTrainsAndReadiness` pins it.
  - `Departed` now has a reader: `nextDeparture` computes the grid's next slot from it at read time, so an offset change applies at once.
  - `alone` records that didn't leave cars ready, old-format ones included, are never a schedule (`TestNextDepartureReadsTheDeparture`).
  - One `waitDeadline` helper replaces the duplicated condition.
  - `Readiness.SinceHead` makes a `Since` carried by an older sdlc across a head change read as stale (`TestStaleSinceFromAnOlderSdlc`).
- **M1 closed (SHIP, two rounds).** Its advisory was the second finding in the family "an unknown-field carry keeps stale meaning". Rule: trust a derived field only on a record written by an sdlc that maintains it. `since_head` became `since_stamp` (= the At of the vouch that wrote Since). An older sdlc rewriting the vouch for any reason (a new head, not ready, ready again) changes At, so the stamp no longer matches and the wait counts from the vouch (`TestStaleSinceFromAnOlderSdlc` covers both reset cases).
- **M2 implemented:**
  - `MachinePlan` (strict decode; unknown members kept, top-level and per repo); `PackOffsets` (golden: ariadne :00, pair and parley :15; overlap property).
  - `machinePlanPath` and `readMachinePlan`, applied in `landingScheduleSource` under its 30 s cache: no entry keeps the interval; an unreadable plan is warned once.
  - `sdlc landing plan [--write]` (`landingplan.go`): costs from `buildTestStatsReport` and `fleetRepoPaths`, both split out of test-stats with no change in behaviour, plus each repo's declared suite weight on main. `--write` keeps unknown members and other repos' entries, and refuses to replace an unreadable plan.
  - The Done-when fixture `TestPlannedReposNeverDepartInEachOthersWindow`.
  - Help, atlas, and tl-lead's steward duties.
  - The landing, train and test-stats tests pass.
- **M2 review round 1 (FIX-THEN-SHIP):** BR-8 to BR-11 and both Minors fixed by their class rules; details in the plan's Revisions.
  - `effectiveSchedule` is the one constructor for admission and status.
  - `fleetPlanInputs` is injectable and tested.
  - Overloads are named.
  - The README is updated.
  - Landing tests are isolated from the user's plan.
  - Duplicate keys are refused.
