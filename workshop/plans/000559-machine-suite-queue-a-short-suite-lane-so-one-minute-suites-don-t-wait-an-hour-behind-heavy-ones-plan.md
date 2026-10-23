---
collaboration-mode: discussion
---

# Short-suite lane in the machine suite queue — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A recorded suite whose recent recorded wall times are short runs in one dedicated, niced light slot.
It does not queue behind heavy suites. Heavy suites keep their FIFO admission untouched.

**Architecture:**
- **Eligibility.** Go (`sdlc local-check`) decides eligibility from the repo's own recorded suite attempts. This is
  measured, never declared.
- **Admission.** Before any work starts, Go asks the checkout's `test-lock.py` for a light admission (`--lane light`).
  The helper keeps a second, separate FIFO line (its own ticket directory) and one extra lock file. The normal slots
  and the normal line never see either, so heavy admission cannot change.
- **Overrun.** A light run that overruns its bound is stopped by local-check's existing timeout and cancel path,
  recorded, and rerun once in the normal line.

**Tech Stack:** Go (`cmd/sdlc`), Python 3 (`scripts/test-lock.py`), flock(2).

Builds on #560 (the aging projection `suite_order`) and #570 (the `ordering_policy` marker on each ticket). Neither
changes; the lane is a separate line beside them.

---

## Scope choices (for TL approval)

- **(a) Recorded suites only.** Only `sdlc local-check` (the recorded suite that every landing waits for) asks for the
  light lane. Plain `make test` always takes the normal line.
  - Eligibility is measured from local-check's attempt records.
  - Only local-check can stop an overrun, record it and requeue it.
  - A hand-run `make test` is never killed at a deadline.
- **(b) Light waiters use only the light slot.** They never take or wait on a normal slot, and they never hold a
  normal ticket. The light line is plain FIFO, with no grants, because waits in it are bounded by the light bound.
  This is what makes "the light slot never borrows a normal slot or delays a heavy suite's admission" structural
  rather than a rule inside the election.
- **(c) Numbers.**

  | Constant | Value |
  |---|---|
  | `LaneSamples` | 3 |
  | `LaneThreshold` | 120 s |
  | `LaneBound` | 240 s |
  | Niceness | +10 |

  - Eligible: the last 3 recorded `suite` attempts of this repo all finished with a wall time under 120 s. Fewer
    than 3 attempts means not eligible.
  - The light run is stopped at 240 s, twice the threshold. Echo measures 61–88 s under today's contention; a niced
    run must not trip the bound on noise.
  - The light run is niced +10.
- **(d) Heavy-impact bound.** ariadne's suite wall time with the light slot continuously busy stays within 10% of
  its recent recorded median. This is measured once before landing (Task 6) and the numbers go in the Log.
- **(e) Rollout.** The light lane is opt-in on both ends, so a mixed fleet degrades to today's behavior:
  - A new `sdlc` with an older checkout helper probes the helper's capability and gets "no lane", so it uses the
    normal line.
  - An older `sdlc` never asks for the light lane.
  - An older helper never opens the light paths.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `LaneEligibility` / `suiteLaneEligibility` | `cmd/sdlc/suitelane.go` | new |
| `suiteLaneStep` / `nextSuiteLaneStep` | `cmd/sdlc/suitelane.go` | new |
| `light_order` | `scripts/test-lock.py` | new |
| `SuiteAttempt.Lane` | `cmd/sdlc/teststats.go` | modified (additive field) |

- **suiteLaneEligibility(attempts, repo) LaneEligibility.** Selects this repo's `check == "suite"`, finished,
  non-main-baseline attempts, newest first by `finished_at`. Eligible iff at least `LaneSamples` exist and each of
  the newest `LaneSamples` has `wall_seconds < LaneThreshold`.
  - An overrun attempt is recorded with its wall time, which is at least `LaneBound`, so it disqualifies the repo by
    measurement alone. No flag and no sticky state: three short normal-lane runs restore eligibility.
  - It returns the samples it used, for status and the Log.
  - ARCH-DRY: it reads through the existing `readSuiteEvidence` loader. It adds no second reader of the store.
- **nextSuiteLaneStep(state, event) (state, effect).** This is the explicit order of one recorded-suite run (ARCH-ORDER;
  table below). local-check's run loop executes its effects and does nothing else.
- **light_order(records).** Light tickets in number order (FIFO). Its head is the only light waiter that may try the
  light slot. It is a pure projection, like `suite_order`, used for both admission and inspection (ARCH-DRY).
- **SuiteAttempt.Lane.** Adds `lane` (`"light"` or absent) and outcome `lane_overrun` to the attempt record.
  - The change is additive (ARCH-EVOLVE): `decodeSuiteAttempt` already preserves unknown members and stays strict on
    known ones. An older reader sees an unknown outcome string, which `teststats` buckets as it does other non-pass
    outcomes. A test checks this.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| helper `acquire --lane light` | `scripts/test-lock.py` | new | flock on `test-suite.lock.light` + `test-suite.lock.light-tickets/` |
| `helperSupportsLane` | `cmd/sdlc/helpercapability.go` | new | the checkout helper's advertised capabilities |
| `acquireSuiteSlot(..., lane)` | `cmd/sdlc/localcheck_run.go` | modified | the helper subprocess |
| inspect `lane` section | `scripts/test-lock.py` → `cmd/sdlc/queue_suite_status.go` | modified | the light lock file and line |

- **Helper light acquire.**
  1. Takes a numbered ticket in `<base>.light-tickets/`, using the same `Tickets` class and dead-waiter pruning, with
     a different directory (ARCH-DRY).
  2. When this waiter is the head of `light_order`, takes `<base>.light` non-blocking under the light allocation
     mutex.
  3. Answers `acquired-light`.
  4. The watcher, slot record and release work exactly as for a normal slot: the kernel frees the flock when the
     watched holder exits.
- **Capability probe.** The helper advertises `lane-v1`, through the same channel #563 uses for the inspect
  capability.
  - When the helper lacks `lane-v1`, `acquireSuiteSlot` never passes `--lane`.
  - A probe failure counts as "no lane", never as an error: the lane is optional.
  - The fake for tests is the real helper, run against a temporary lock root (`WF_TEST_LOCK_FILE`). This is the
    existing pattern in the helper tests, not a function mock (ARCH-MOCK).
- **Status.** `inspect` gains an additive `lane` object: `{holder, waiters[], bound_seconds}`. `sdlc queue status`
  prints a "light lane" line with the holder (repo, slot, how long it has run) and the waiters in order.
  - `suiteQueueObservation` already tolerates added helper fields (its `wire:` note).
  - An older helper's inspect output has no `lane` object, and status then says "light lane: not supported by this
    helper".

### Run order (ARCH-ORDER)

The state lives inside one local-check run; it is in memory and dies with the run.

| State | Event | Next state | Effect |
|---|---|---|---|
| start | eligible and helper has `lane-v1` | waitLight | acquire `--lane light` |
| start | not eligible, or no capability | waitNormal | acquire normally (today's path) |
| waitLight | `acquired-light` | runLight | run `nice -n 10 make test` with a `LaneBound` deadline |
| waitLight | ctx cancelled or lease cancelled | done | release; nothing recorded (today's behavior) |
| runLight | exit before deadline | done | record attempt `lane: light`, normal outcome |
| runLight | deadline | waitNormal | kill the group (existing timeout path); record `lane_overrun` with its wall time; release the light slot |
| runLight | ctx cancelled | done | existing cancel path; record cancelled, `lane: light` |
| waitNormal | acquired | runNormal | run `make test` (today's path) |
| runNormal | any | done | today's recording |

- **Invariants:**
  - At most one light attempt per run, so an overrun cannot loop.
  - The light slot is released before the normal wait begins, so no run holds both.
  - A run never holds a normal slot and the light slot together.
- **Event most likely to be mishandled:** the deadline firing as the suite exits by itself. The existing timeout
  path decides `TimedOut` from the context's error, not from the exit code. The table maps `TimedOut` with lane
  light to an overrun, so a completed run that raced the deadline reruns once. That wastes some time but never
  records a false pass.
- **Process death mid-run:** the kernel releases the flock (the existing watcher design), and no attempt is
  recorded, as today.
- **Tests** drive `nextSuiteLaneStep` over every row, plus rejected events. Each rejected event (for example
  `acquired-light` while in `runNormal`) returns an error. An integration test runs the loop with a fake suite
  command whose duration is controlled and a deadline injected through `localRunDeps`.

### Operating envelope (ARCH-CONSTRAINTS)

- **Workload:** batch. Admission is a once-a-second poll, as today.
- **Eligibility read:** one read of the repo's attempt store per local-check run. It is bounded by the existing
  `readSuiteEvidence` limit and byte budget (445 files in ariadne today; pruned by `pruneSuiteAttempts`).
- **CPU:** the one extra concurrent suite is niced +10 and bounded at 240 s of wall time. Its impact on a heavy suite
  is measured, not assumed (Task 6, bound (d)).
- **Concurrency:** exactly one light slot. Light waiters poll like normal waiters, with no new process per waiter.
- **Overload:** when many short suites arrive, the light line grows FIFO. Each holder is bounded at 240 s, so the
  k-th waiter waits at most k × 240 s, and the heavy line is unaffected.

### Trust and residue

- **ARCH-SECURE.** The light ticket and slot records are untrusted, exactly like normal ones. Decoding reuses
  `holder()` and the inspection snapshot bounds.
  - Attempt records come from the repo's own store. They pass through the strict decoder, and a malformed attempt
    counts as "not eligible", never as short.
  - No credentials are involved.
- **ARCH-FUNERAL.** The design creates two durable things:
  - **Light lock file:** one file, rewritten per holder, never more.
  - **Light ticket files:** removed on release and pruned when their waiter dies, by the same code as normal tickets.

  Attempt records gain one short field each, and the existing prune removes them.

## Tasks

### Task 1: eligibility (pure) and the attempt field

**Files:** create `cmd/sdlc/suitelane.go` and `cmd/sdlc/suitelane_test.go`; modify `cmd/sdlc/teststats.go`.

- [ ] Table tests for `suiteLaneEligibility`, with literal 120 and 3 in the cases as an independent oracle:
  - fewer than 3 attempts
  - all three under 120 s
  - one at 120 s exactly
  - an overrun among the last three
  - an older long attempt pushed out by three short ones
  - other repos and other checks ignored
  - main-baseline attempts ignored
  - a malformed attempt
- [ ] Add the `Lane` field and the `lane_overrun` outcome, and test that an older-schema record round-trips with an
  unknown `lane` preserved.
- [ ] Commit.

### Task 2: the helper's light line

**Files:** modify `scripts/test-lock.py` and its tests (`cmd/sdlc/*testlock*_test.go`, the existing helper suite).

- [ ] Failing tests against the real helper in a temporary lock root:
  - With both normal slots held by a weight-2 holder, `acquire --lane light` answers `acquired-light` within a
    poll. (Done-when 1.)
  - Two light waiters are admitted in ticket order.
  - A normal waiter's position and admission are unchanged by 10 queued light waiters (the heavy FIFO stream test,
    Done-when 4): the normal line's `suite_order` output is identical with and without them.
  - A dead light waiter's ticket is pruned.
- [ ] Implement `--lane light`, `light_order`, the `lane-v1` capability and the inspect `lane` object.
- [ ] Commit.

### Task 3: capability probe and lane-aware acquisition

**Files:** modify `cmd/sdlc/helpercapability.go` and `cmd/sdlc/localcheck_run.go`, with tests beside them.

- [ ] Tests:
  - A helper without `lane-v1` (a copy of the pre-#559 helper from git) causes a normal acquire.
  - With `lane-v1`, an eligible run passes `--lane light` and handles `acquired-light`.
- [ ] Commit.

### Task 4: run order and overrun requeue

**Files:** `cmd/sdlc/suitelane.go` (`nextSuiteLaneStep`) and `cmd/sdlc/localcheck_run.go` (the loop).

- [ ] Table test over every row of the run-order table, plus rejected events.
- [ ] Integration test: an eligible repo whose fake suite sleeps past an injected short bound is stopped. The test
  checks four things:
  - one `lane_overrun` attempt is recorded, with `lane: light`;
  - the run is requeued in the normal line;
  - the second run passes;
  - eligibility is now false. (Done-when 2.)
- [ ] Light runs execute as `nice -n 10`; assert the argv.
- [ ] Commit.

### Task 5: status

**Files:** `cmd/sdlc/queue_suite_status.go` and its test; `atlas/workflow/sdlc-binary.md` (the queue section).

- [ ] The status output shows a "light lane" line with the holder and the waiters (Done-when 3), and a
  "not supported by this helper" line for an older inspect payload.
- [ ] Update the atlas text, stating that the light lane adds a third concurrent run beyond `WF_TEST_SUITES=2`, bounded to light suites (one slot, niced, stopped at 240 s).
- [ ] Commit.

### Task 6: verification

- [ ] Run the targeted tests.
- [ ] Run the recorded suite (`sdlc local-check run --name suite`).
- [ ] Heavy-impact measurement: run ariadne's suite, both slots, while a niced short suite loops in the light slot
  (echo's suite if its checkout is free, otherwise a looping `go test` of a small ariadne package). Compare the wall
  time to the median of the last 5 recorded ariadne suite attempts and log both numbers. Bound: at most 10% slower.
  If the bound fails, stop and re-plan; do not land.

## Revisions

- 2026-10-23: TL approved scope choices (a)–(e) as written; Task 5's atlas update must state that the lane adds a third concurrent run beyond `WF_TEST_SUITES=2`, bounded to light suites.
- 2026-10-23 (close review round 1, BR-1 and BR-2): what was built differs from the text above in four places.
  - The overrun is evidence on the attempt that retried it: `lane_overrun_seconds`, beside `lane`. There is no separate light attempt and no `lane_overrun` outcome (ARCH-EVOLVE: no new outcome value). Eligibility reads the field, so an overrun still takes the repo out of the lane by measurement. The stopped light run's log is kept beside the retry's, as `<log>.light`. A retry that cannot run names the overrun in its error.
  - The capability probe reads `--lane` in the helper's `acquire --help`, through the same `probeHelperHelp` grammar as the other options (`helperSupportsLane`). There is no `lane-v1` token. The inspect `lane` object carries `version` for the status view.
  - The eligibility function is `laneEligibility`, returning `suiteLaneEligibility`.
  - Typed-runner suites (local-checks v3 with flake runners, `StrictGroupCleanup`) never ask for the lane. Their strict cleanup cannot confirm a run that the light deadline stopped, so an overrun could be neither retried nor recorded. No repo declares typed runners today: echo is on v1, pair on v2, ariadne on v1.
