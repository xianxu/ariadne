# Boundary Review — ariadne#449 (milestone M1)

| field | value |
|-------|-------|
| issue | 449 — landing train: batch ready PRs on an interval and land them with one full suite (extends #393) |
| repo | ariadne |
| issue file | workshop/issues/000449-landing-train-batch-ready-prs-on-an-interval-and-land-them-with-one-full-suite-extends-393.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | e279f713f66ebde2d11e8662e05d4e481369d34c..4babdef455239df1cb2354f0ce56d58425d204f0 |
| command | sdlc milestone-close --issue 449 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T01:43:53-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: done with the pure train core (state machine, bisect, boarding, sidecar). Next I'm reading the car-side vouching, config, and admission wiring in `cmd/sdlc`.

Checking whether the read-only CI gate path can hit the nil local-check runner, then running the tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 delivers what it claims, with a few cheap fixes left. The pure core is in place and tested: `Schedule`/`DepartureDue`, the `Readiness` sidecar codec, `Board`, the `Train`/`Transition` state machine and `Bisect`. The car's vouching runs today's publish and CI gates read-only. `.sdlc/landing.json` parses strictly, `sdlc landing status` shows the train setting and each waiter's vouch, and the atlas has a matching section. When the config is absent or the interval is 0, `Admissible` stays nil and the serial line is unchanged. I re-ran the landingline package and all new cmd tests under `-race`; they pass. Nothing is Critical. Five things should be fixed before crossing:
- **A data race I confirmed with `-race`:** the vouching goroutine reads `Handle.file` while admission (`dropTicket`) writes it. Only an flock separates the two, and the Go race detector does not count that as synchronization.
- **`Built` can silently drop a car:** a car listed in neither `Kept` nor `Conflicts` gets no outcome.
- **`Board` trusts the sidecar's branch:** it boards `r.Branch` without checking it against the ticket's branch.
- **The plan's Core concepts table names entities and states the code doesn't have.**
- **README:** it does not mention the new `.sdlc/landing.json` setting.

### 1. Strengths
- **Typed sidecar parsing.** `DecodeReadiness` (`train.go:127`) requires one object, matches member names exactly, rejects duplicates and nulls, refuses trailing data, and keeps unknown members in `Extra`. `MarshalJSON` never lets `Extra` override a known field. `Vouch` merges the previous `Extra` (`line.go:271-279`) and has a round-trip test. This satisfies both ARCH-SECURE and ARCH-EVOLVE.
- **Explicit state machine.** `Transition` (`trainstate.go:159`) has a finite set of phases and events, returns effects as values, and rejects illegal events without changing state. `TestTrainInvariantsOverRandomHistories` checks invariants stated separately from the code (land only a passing prefix, each car ends exactly once, terminal states accept nothing) over 2000 random histories that include illegal events.
- **Bisect cost is counted, not assumed.** `TestBisectCostIsLogarithmic` checks every culprit position at N=8 and N=16 against ⌈log2 N⌉. The edge test with contradicting results pins that `lo` can never pass the first failure.
- **Sidecar lifecycle is complete.** Sidecars live in their own `ready/` directory, which avoids the bug flagged earlier where `scan` would have deleted them. `sweepReady` removes orphans and leftover `.tmp-` files, and `dropTicket` removes the sidecar on exit.
- **Gates are reused read-only.** `productionCarGates` (`landingcar.go:78`) passes a nil local-check runner. I checked `runPreparedCIGate` (`cigate.go:606`): it is safe with a nil runner, so a missing suite record means "not ready" and never starts a suite run.

### 2. Critical findings
None.

### 3. Important findings
- **Data race on `Handle.file` (ARCH-ORDER)** — `line.go:265` (read in `Vouch`) against `line.go:383` (write in `dropTicket`, reached from `tryAdmit`).
  - In production, `vouchWhileWaiting` (`landingcar.go:96`) calls `Vouch(h, …)` in one goroutine while `line.Wait(ctx, h)` admits and drops the ticket in another. The only exclusion is the `line.lock` flock, which creates no ordering the Go memory model or race detector recognises.
  - I reproduced it with a throwaway test (Vouch in a loop alongside Wait), which `-race` flags at those two lines.
  - The shipped test `TestEnterLandingLineVouchesWhileWaiting` misses it only because both goroutines happen to share the `watchWriter` mutex.
  - There is a second symptom: a vouch that runs after admission gets "only a waiting merge vouches" and prints the misleading "could not record readiness… boards no train" warning.
  - Fix: add a `sync.Mutex` to `Handle`, held by `dropTicket`, `Release` and `Vouch`'s `h.file` check. Treat the post-admission error as a quiet stop. Keep the probe as a regression test.
- **`Built` drops cars with no outcome (ARCH-ORDER)** — `trainstate.go:181-188`.
  - A car in neither `Kept` nor `Conflicts` vanishes from `n.Cars` with no Hold, Release or Refuse effect, which breaks the "every car ends exactly once" invariant.
  - The random-history generator always splits every car into one of the two sets, so the invariant test never reaches this input.
  - Fix: reject a `Built` whose sets don't partition `Cars` (or Release the unaccounted cars), and let the generator produce such inputs.
- **`Board` uses the sidecar's branch, not the ticket's (ARCH-SECURE)** — `train.go:293`.
  - `Car.Branch` comes from `r.Branch`, and `Valid` (`train.go:204`) never compares it with `Ticket.Branch`.
  - A hand-edited or buggy sidecar can make the M2 conductor merge a different branch than the waiting merge owns. The tip re-check would still pass, because it re-reads that same untrusted branch.
  - Fix: `Board` refuses (adds to `Why`) when `r.Branch != t.Branch`, or builds the car from `t.Branch`. Add a table row for it.
- **Plan's Core concepts table contradicts the code (plan drift)** — plan lines 22, 27 and 42.
  - The table names `TrainState`, `TrainEvent` and `Transition`; the code has `Train`, `Phase`, `TrainEvent` and `Transition`.
  - The plan's states include `Held`, but the code has no `Held`; it adds `Reintegrating`, `Merging` and `Refused`.
  - The plan's events (`CarConflict`, `CarGone`, `TipMoved`, `FlakeRerunPass`, `CIGreen`/`CIRed`, `ConductorDead`) differ from the code's (`Boarded`, `Built`, `SuiteDone`, `CIDone`, `Reintegrated`, `Probed`, …).
  - The plan's signature `DepartureDue(now, lastDeparture, sched, line)` differs from the code's `(now, s, granted, events)`, which reads `NextDue` from the event log.
  - Fix: a Revisions entry (see section 7).
- **README update appears missing for `.sdlc/landing.json`** (`train_interval_minutes`, `train_max_cars`) and for the new train lines in `sdlc landing status`.
  - The plan puts README in M3, but the config key is user-visible now (status prints "`.sdlc/landing.json` sets an interval").
  - Fix: add a short README note now, or record the deferral explicitly in the plan's Revisions.

### 4. Minor findings
- **Events log read once per ticket per poll (ARCH-CONSTRAINTS):** `trainAdmission` calls `events()` for every live ticket on every poll, inside the line mutex. Each call reads all of `events.jsonl` (30 days). Read it once per `tryAdmit`.
- **Grant bypass written twice (ARCH-DRY):** `trainAdmission` handles `granted` itself and always passes `granted=false` to `DepartureDue`, so that branch of `DepartureDue` is dead in production.
- **Two definitions of "main" (ARCH-DRY):** status uses `landingMainRef` (`main@{upstream}`), while `runDurableMerge` uses `t.Remote+"/main"`.
- **Stale config read before the wait:** `runDurableMerge` reads the schedule from the remote-tracking main without fetching, and keeps it for the whole wait, which can be an hour. A missing `<remote>/main` ref now refuses even interval-0 merges.
- **Skipped tickets drop out of the audit:** when `Admissible` filters out not-ready tickets, the tickets passed over are missing from `OnAdmit`'s `displaced` list.
- **The schedule cannot advance in M1:** with the config turned on, nothing writes a `train` event until the M2 conductor exists. Every admission reads as "first train", and status reports an interval that does nothing. That is acceptable only because the default is off.

### 5. Test coverage notes
- The pure entities are tested without IO: `DepartureDue`, `Board`, `Bisect`, `Transition` and `DecodeReadiness`.
- Line-level tests use temporary directories, which is acceptable for this flock-based component.
- **Missing:**
  - a concurrent Vouch-versus-admission race test;
  - a `Built` input whose sets don't partition the cars;
  - a sidecar whose branch differs from its ticket's.
- I did not re-run the 1922s interval-0 regression suite; I relied on the unchanged-behaviour argument that `Admissible` is nil when trains are off.

### 6. Architectural notes for upcoming work
- **ARCH-DRY:** two minor duplications (see section 4).
- **ARCH-PURE:** pass. The clock and events are injected, the gates are injected through `carGates`, and the core is pure.
- **ARCH-PURPOSE:** pass for an M1 subset. The conductor, the landed-commit resolver and the fake are correctly left to M2. Note the fifth `MergeOID` consumer (`parseLandingPRs`) recorded in the Log: the M2 sweep must cover it.
- **ARCH-MOCK:** N/A for M1. The gh and git seams are unchanged, and the reachability fake belongs to M2.
- **ARCH-CONSTRAINTS:** minor flag (repeated events read).
- **ARCH-SECURE:** flag (branch not cross-checked). Otherwise the parse at the boundary is good.
- **ARCH-ORDER:** flags (Handle race; `Built` drop). Also, after `Reintegrated{NeedSuite}` a failing suite bisects "prefix k". M2 must define which commit prefix k means after reintegration: old-base prefixes are not the same tree. Also, `MainMoved` while `Merging` assumes the merge was refused; the M2 shell must only send it when the merge is confirmed not done, or route it through `Probe`.
- **ARCH-FUNERAL:** pass. Sidecars and temp files are swept, and events are bounded by `KeepEvents`.
- **ARCH-EVOLVE:** pass. `Readiness` round-trip preservation is tested, `Event` fields are additive, and the config is read-only so ignoring unknown keys is fine. Refusing another `version` degrades to "not ready", which is safe.

### 7. Plan revision recommendations
- Add a Revisions entry: "M1 as built: the state machine is `Train{Phase}` with phases forming/integrating/testing/bisecting/reintegrating/landing/merging and terminals landed/abandoned/refused (there is no Held state; a held car is a `Hold` effect). Events are Boarded/Built/SuiteDone/CIDone/MainMoved/Reintegrated/Merged/MergeUnknown/Probed. `DepartureDue(now, sched, granted, events)` reads the latest `train` event's `next_due` rather than a `lastDeparture` argument. `Alone` comes from the `local-gates-verified` label." Update the Core concepts table to match.
- Add a Revisions note stating whether README for `.sdlc/landing.json` is deferred to M3 or added now.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: shared-handle-unsynchronized
    title: |
      Data race on Handle.file between Vouch (vouch goroutine) and dropTicket (Wait/admission)
    detail: |
      line.go:265 reads h.file in Vouch while line.go:383 writes it via tryAdmit; only an flock separates them, which the Go memory model does not see. A temporary probe test (Vouch in a loop alongside Wait) is flagged by go test -race. vouchWhileWaiting runs exactly this in production; it also prints a misleading "could not record readiness" warning after admission. Fix: a sync.Mutex on Handle covering file/dropTicket/Vouch, treat the post-admission error as a quiet stop, and keep the probe as a regression test.
  - id: new
    severity: Important
    kind: defect
    family: state-machine-event-partition
    title: |
      Transition(Built) silently drops cars that are in neither Kept nor Conflicts
    detail: |
      trainstate.go:181-188 removes such cars from n.Cars with no Hold/Release/Refuse effect, breaking the every-car-ends-once invariant. The random-history generator always partitions the cars, so the invariant test cannot reach this input. Reject a non-partitioning Built (return an error) or Release the unaccounted cars, and make the generator produce such inputs.
  - id: new
    severity: Important
    kind: defect
    family: untrusted-sidecar-cross-check
    title: |
      Board builds Car.Branch from the sidecar's branch without checking it against the ticket's branch
    detail: |
      train.go:293 uses r.Branch, and Readiness.Valid never compares it with Ticket.Branch. A hand-edited or buggy sidecar can steer the M2 conductor to a branch the waiting merge does not own, and the tip re-check passes because it reads the same untrusted branch. Refuse a mismatch in Board (add it to Why) or use t.Branch, and add a table test.
  - id: new
    severity: Important
    kind: defect
    family: plan-core-concepts-drift
    title: |
      Plan Core concepts and state lists do not match the code (TrainState, Held, event names, DepartureDue signature)
    detail: |
      The plan names TrainState, a Held state and the events CarConflict/CarGone/TipMoved/FlakeRerunPass/CIGreen/CIRed/ConductorDead, plus DepartureDue(now, lastDeparture, sched, line). The code has Train{Phase} with Reintegrating/Merging/Refused, the events Boarded/Built/SuiteDone/CIDone/Reintegrated/Probed, and DepartureDue(now, s, granted, events) reading next_due from the log. Add a Revisions entry and update the table.
  - id: new
    severity: Important
    kind: defect
    family: docs-gate-readme
    title: |
      README update appears missing for the .sdlc/landing.json config keys and the train lines in landing status
    detail: |
      train_interval_minutes and train_max_cars are a new user-facing config surface, and status already points users to the file; README.md is unchanged in the range. Add a short README note, or record the deferral to M3 explicitly in the plan's Revisions.
  - id: new
    severity: Minor
    kind: defect
    family: repeated-io-in-hot-loop
    title: |
      trainAdmission reads the whole events.jsonl once per live ticket per poll, inside the line mutex
    detail: |
      Read the events log once per tryAdmit and pass it in.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-rule
    title: |
      The grant bypass is written in both trainAdmission and DepartureDue, and "main" has two definitions (main@{upstream} vs remote/main)
    detail: |
      trainAdmission always passes granted=false, so DepartureDue's granted branch is dead in production. landingMainRef (status) and runDurableMerge (t.Remote+"/main") resolve main differently.
  - id: new
    severity: Minor
    kind: defect
    family: stale-config-read
    title: |
      The schedule is read from an unfetched remote-tracking main before the wait and held for the whole wait
    detail: |
      A missing remote/main ref now refuses even interval-0 merges, and a config change made during a long wait is not seen.
  - id: new
    severity: Minor
    kind: defect
    family: audit-displaced-filtered
    title: |
      Not-ready tickets skipped by Admissible are missing from OnAdmit's displaced audit
    detail: |
      NextHead runs over the filtered candidates, so the tickets passed over are not recorded.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2b877597-7f69-4ea8-a79f-543a37ca2461","repository":"github.com/xianxu/ariadne","issue_id":"000449","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/739c47422e6f2298","reviewed_head":"4babdef455239df1cb2354f0ce56d58425d204f0","slot":"ariadne:2","requested_at":"2026-10-19T08:40:06.232542Z"},"phase":"finished","pid":33940,"started_at":"2026-10-19T08:40:06.463357Z","finished_at":"2026-10-19T08:43:52.200262Z","session_id":"416d5331-7349-441f-b806-74aadb741ffd","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T02:57:26-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 449 — landing train: batch ready PRs on an interval and land them with one full suite (extends #393) |
| repo | ariadne |
| issue file | workshop/issues/000449-landing-train-batch-ready-prs-on-an-interval-and-land-them-with-one-full-suite-extends-393.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | e279f713f66ebde2d11e8662e05d4e481369d34c..452ccde50bfc1f2304db66a1530d11bfa83446bd |
| command | sdlc milestone-close --issue 449 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T02:57:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Now run tests with race and verify reverts.

Running the full suite: two archive tests fail, and the run hits the 10-minute timeout. Next I'll check whether both failures also happen at the base commit.

The merge/landing run timed out. That could be a real hang, so I'll find which test was running.

No hang: the tests are just slow (about 24s each), so the run hit the timeout. I'll split them into parallel groups.

All 219 landing, merge, train and vouch tests pass across the 10 shards. Next I'll revert each round-1 fix in this disposable checkout and confirm its regression test goes red.

BR-1 is confirmed: without the mutex the race detector fails the test. Now BR-2, BR-3 and BR-9.

BR-3 is confirmed. Now checking BR-9 (audit of skipped tickets) the same way, and looking for other `NextHead` callers.

BR-9 is confirmed. One new suspect: `landingScheduleSource` keeps a `warned` flag that both the vouching goroutine and the admission poll call. I'll write a race probe to check it.

The probe shows a data race at `landingqueue.go:206-207`. Now checking that production really shares one `sched` closure between the two goroutines.

No deadlock: `Events()` reads the file without the mutex. Last, checking that `landing status` shows what the new README paragraph claims.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All nine round-1 findings are fixed. For each behavior fix I checked the regression evidence by reverting the fix in this checkout: the matching test went red for BR-1, BR-2, BR-3 and BR-9. For BR-6 and BR-8 I read the new counting and turn-on tests but did not revert those fixes. The landing-line package passes under `-race`. All 219 landing, merge, train, vouch and car tests in `cmd/sdlc` pass, run in 10 parallel shards. In one run of the whole package, two archive tests failed for an unrelated reason: the sandbox's read-only git remote breaks `git fetch`. One new problem should be fixed first, and cheaply. It is a second instance of the BR-1 family: the `warned` flag in `landingScheduleSource` is shared without a lock between the vouching goroutine and the admission poll, and a two-goroutine `-race` probe flags it. Three Minor repeats of earlier families are also noted.

**1. Strengths**
- `Readiness.For` (`train.go:224`) is now the one check that a vouch matches its ticket, used by both `Board` and `trainAdmission`. A car's branch now comes from the ticket, not the sidecar.
- `partitions` (`trainstate.go:315`) rejects a `Built` event that loses, reorders, double-counts or invents a car. The random-history generator now produces such reports, and the "every car ends exactly once" invariant fails if the check is removed.
- `Handle.mu` plus `ErrNotWaiting` gives a clean stop after admission. `TestVouchConcurrentWithAdmission` is a real race regression test: it fails with `DATA RACE` when the mutex in `Vouch` is removed.
- If the schedule can't be read, landing falls back to the serial line with one warning instead of refusing. The schedule is re-read on each attempt, so a config change made during a wait now counts.

**2. Critical findings**
None.

**3. Important findings**
- **New finding — data race on `warned` in `landingScheduleSource` (`cmd/sdlc/landingqueue.go:206-207`). ARCH-ORDER.**
  - **Where it happens:** `enterLandingLine` passes the same `sched` closure to two places that run at the same time: `openLandingLine` (called from `Wait` → `tryAdmit` → `Admission()`) and the `vouchWhileWaiting` goroutine.
  - **Evidence:** a probe that calls `src()` from two goroutines while the config is unreadable is flagged by `go test -race`.
  - **Family:** this is the second finding in family `shared-handle-unsynchronized`.
  - **The rule:** anything handed to both the vouching goroutine and `Line.Wait` must be safe to use from two goroutines.
  - **Sweep of what the two goroutines share:**
    - the line: protected by the flock and `Handle.mu`;
    - `readiness`: used only by the vouching goroutine;
    - `stderr`: an `*os.File` in production, and the test writer has its own mutex;
    - `sched`: unsynchronized, and the only remaining instance.
  - **Fix:** use `sync.Once` or `atomic.Bool` for the warning, and add a concurrent `-race` test of `landingScheduleSource`.

**4. Minor findings**
- **`sdlc landing status` reads vouches without `Readiness.For`** (`landingcmd.go:199-203`).
  - **Impact:** a vouch naming another branch shows as ready in status while admission refuses it.
  - **Family:** second finding in `untrusted-sidecar-cross-check`. The rule: every reader of a sidecar goes through `For`.
  - **Related:** status prints "trains: unknown" when the config can't be read, but merges then land serially. Status should say so.
- **Every waiter re-reads the schedule every 2 seconds** (`trainAdmission` → `sched()`, `landingqueue.go:202`).
  - **Impact:** that is two `git` processes per waiter, inside the line flock, even with trains off (the default). Before this change, the serial line ran no git there.
  - **Family:** second finding in `repeated-io-in-hot-loop`. The rule: inputs to the admission rule are read outside the mutex, and slow-changing ones are cached (for example by a time limit or by main's commit id).
- **The audit's "passed over" rule is now written twice** (`line.go:244-257`).
  - **Impact:** this loop restates `NextHead`'s rule, and `NextHead`'s `displaced` return now has no production caller.
  - **Family:** second finding in `duplicated-rule`. The rule: one pure function in `order.go` computes who was passed over, and `NextHead` returns only the head.

**5. Test coverage notes**
- Revert-to-red was confirmed for BR-1 (`TestVouchConcurrentWithAdmission` under `-race`), BR-2 (`TestTrainRejectsIllegalEvents` and `TestTrainInvariantsOverRandomHistories`), BR-3 (`TestBoard` and `TestTrainAdmission`) and BR-9 (`TestAdmissibleGatesTheHead`).
- BR-6 is pinned by the read counter in `TestTrainAdmission` and the per-attempt counter in `TestAdmissibleGatesTheHead`.
- BR-8 is pinned by `TestLandingScheduleSourceFailsToSerial` and `TestVouchingStartsWhenTrainsTurnOn`.
- **Gap:** nothing runs the schedule source from two goroutines at once (see the new Important finding).

**6. Architectural notes for upcoming work**
- **ARCH-DRY: flag** (Minor finding on the audit rule).
- **ARCH-PURE: pass.** `DepartureDue`, `Board`, `Transition`, `Bisect` and `For` are pure; the IO is injected as functions.
- **ARCH-PURPOSE: pass.** M1's scope is delivered; the conductor is M2, as the plan says.
- **ARCH-MOCK: pass for M1.** Config reads go through an injected git function. The GitHub fake is M2's.
- **ARCH-CONSTRAINTS: flag** (Minor finding on the 2-second git reads).
- **ARCH-SECURE: pass**, apart from the status gap.
- **ARCH-ORDER: flag** (Important finding). The vouching goroutine's lifetime is bounded by `stopVouching`.
- **ARCH-FUNERAL: pass.** `scan` sweeps the `ready/` sidecars, and `KeepEvents` sweeps the events.
- **ARCH-EVOLVE: pass.** `Vouch` keeps a newer version's unknown members, and a test covers it.

For M2: the conductor will add more goroutines and closures. Apply the shared-state rule above to everything they share, not just to new handles.

**7. Plan revision recommendations**
None. The BR-4 Revisions entry ("M1 as built") matches the code: `Train{Phase}`, the event and effect names, `DepartureDue(now, s, events)` and `Options.Admission`.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Handle.mu guards file/head; removing it from Vouch makes TestVouchConcurrentWithAdmission fail with DATA RACE under -race.
  - id: BR-2
    disposition: addressed
    note: |
      partitions() rejects non-partitioning Built; reverting fails TestTrainRejectsIllegalEvents and the random-history invariant.
  - id: BR-3
    disposition: addressed
    note: |
      Readiness.For checks the branch and Board uses t.Branch; reverting fails TestBoard and TestTrainAdmission.
  - id: BR-4
    disposition: addressed
    note: |
      Plan Revisions entry M1 as built (BR-4) matches trainstate.go/train.go names, events, effects and the DepartureDue signature.
  - id: BR-5
    disposition: addressed
    note: |
      README.md:491-498 documents train_interval_minutes, train_max_cars, the local-gates guard, and the status lines.
  - id: BR-6
    disposition: addressed
    note: |
      Options.Admission is built once per attempt; TestTrainAdmission counts one events read across two tickets.
  - id: BR-7
    disposition: addressed
    note: |
      The grant branch is gone from DepartureDue; one landingMainRef(git, remote) is used by both merge and status, and is tested.
  - id: BR-8
    disposition: addressed
    note: |
      landingScheduleSource re-reads on each attempt and falls back to serial; TestLandingScheduleSourceFailsToSerial and TestVouchingStartsWhenTrainsTurnOn cover it.
  - id: BR-9
    disposition: addressed
    note: |
      The audit walks the full order; reverting to NextHead's displaced fails TestAdmissibleGatesTheHead.
findings:
  - id: new
    severity: Important
    kind: defect
    family: shared-handle-unsynchronized
    title: |
      Data race on the warned flag in landingScheduleSource, shared by the vouching goroutine and the admission poll
    detail: |
      landingqueue.go:206-207. enterLandingLine passes one sched closure to both openLandingLine (Admission, Wait goroutine) and vouchWhileWaiting; a two-goroutine probe with an unreadable config is flagged by go test -race. This is the 2nd finding in this family. The rule: everything handed to both vouchWhileWaiting and Line.Wait must be safe for concurrent use. Swept: the line (flock and Handle.mu), readiness (vouch goroutine only), stderr (os.File) are fine; sched is the only remaining instance. Fix with sync.Once or atomic.Bool and a concurrent -race test.
  - id: new
    severity: Minor
    kind: defect
    family: untrusted-sidecar-cross-check
    title: |
      landing status reads readiness sidecars without Readiness.For, so a vouch for another branch shows as ready
    detail: |
      landingcmd.go:199-203 uses readinessLine and Current only. This is the 2nd finding in this family; the rule is that every sidecar reader goes through Readiness.For(ticket, now). Status also prints trains: unknown when the config is unreadable, while merges then land serially.
  - id: new
    severity: Minor
    kind: defect
    family: repeated-io-in-hot-loop
    title: |
      Every waiter runs two git processes each 2s poll under the line flock to re-read the schedule, even with trains off
    detail: |
      trainAdmission calls sched() (git ls-tree and git show) inside tryAdmit's mutex on every attempt. Before this change the serial default ran none. This is the 2nd finding in this family; the rule is that admission inputs are read outside the mutex and slow-changing ones are cached (time limit or main's commit id).
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-rule
    title: |
      The audit's passed-over rule is restated in line.go while NextHead's displaced return has no production caller
    detail: |
      line.go:244-257 duplicates order.go:95-99. This is the 2nd finding in this family; the rule is one pure function in order.go computes who was passed over, and NextHead returns only the head.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d58ee59d-3e8d-4879-9140-3a040c2dea60","repository":"github.com/xianxu/ariadne","issue_id":"000449","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e5305a206114e446","reviewed_head":"452ccde50bfc1f2304db66a1530d11bfa83446bd","slot":"ariadne:2","requested_at":"2026-10-19T09:09:30.136993Z"},"phase":"finished","pid":14812,"started_at":"2026-10-19T09:09:30.243292Z","finished_at":"2026-10-19T09:57:26.154955Z","session_id":"f7f523a8-fecb-4230-a343-b3686b0ea28c","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T03:50:39-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 449 — landing train: batch ready PRs on an interval and land them with one full suite (extends #393) |
| repo | ariadne |
| issue file | workshop/issues/000449-landing-train-batch-ready-prs-on-an-interval-and-land-them-with-one-full-suite-extends-393.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | e279f713f66ebde2d11e8662e05d4e481369d34c..96d1a9d01bb99b19bac8f2f3c9fd9666e46142c6 |
| command | sdlc milestone-close --issue 449 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T03:50:39-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

The last commit is the round-2 answer, so I'll look at it first.

Verifying the round-2 fixes now: running the new tests under `-race`, then reverting the BR-10 fix to confirm the test catches the race.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

Round 2 fixed all four open findings, and the fixes are backed by tests. The schedule source is now behind a mutex with a 30s cache (BR-10, BR-12). When I removed the lock, `TestLandingScheduleSourceIsSharedAndCached` failed under `-race` with three `DATA RACE` reports and "400 calls … ran 4 config reads". With the lock restored it passes. Status now checks vouches with `Readiness.For` (BR-11). `Passed` in `order.go` is now the only place the audit rule is written, and `NextHead` returns only the head (BR-13). The landingline package passes under `-race`, and the narrowed landing test set in `cmd/sdlc` passes (119s). I did not run the full `cmd/sdlc` suite to completion. A broad `-race` run with `-run 'Landing|Merge|Train|Vouch|Car'` hit the 600s foreground limit. I then tried to stop it, and its saved output holds no test result, so I can't say whether that run passed. The two new findings are both Minor. The more important one is an ordering hazard that the BR-12 fix opens for M2, because `Admission()` now reads the events log before the line mutex is taken. Nothing blocks the M1 gate.

1. **Strengths**
   - `landingqueue.go:209-231`: the schedule source is one lock-protected closure shared by the vouching goroutine and admission. The cache also handles a clock that goes backwards (`!now.Before(readAt)`).
   - `order.go:96-112`: `Passed` derives the audit from the full `Order(live, grants)`. The audit therefore names both older tickets a grant passed and tickets skipped as not admissible, and `TestPassedIncludesTicketsThatWereNotAdmissible` checks this.
   - `landingcmd.go:201-206`: status uses the same `For` check as admission and boarding. The test at `landingcmd_test.go:184-196` uses a vouch that names another branch, so the check is exercised.
   - `trainAdmission` (`landingcar.go`) is built once per attempt, and with trains off it is a constant "admit". Together with the 30s cache, the serial default runs no git processes on most polls.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings**
   - **The departure decision can use a stale events log (ARCH-ORDER).** `line.go:188-193` builds `admissible` from `Events()` before taking `withMutex` and `head.lock`. `Events()` does not take the mutex, so it would be safe to call inside it.
     - **Why it doesn't bite yet:** nothing in M1 writes `EventTrain`.
     - **The M2 failure:** waiter B reads the events, then conductor A records `train{next_due: +60m}` and releases `head.lock`. B then gets the head on its old "train due" reading, so a second train leaves immediately and the interval is skipped.
     - **Fix:** keep the cached schedule read outside the mutex. Read the events, and evaluate `DepartureDue`, inside `withMutex` after `head.lock` is acquired. This means splitting the `Admission` contract into a slow part read outside (the schedule) and a part decided under the lock (the departure ledger).
   - **The 30s cache is not reflected in two descriptions.** `landingqueue.go:39` (the `openLandingLine` doc) and plan Revisions line 205 still say the schedule is "re-read … on each attempt". The code now reuses a read for up to 30s (`landingScheduleTTL`). This is the 2nd finding in `plan-core-concepts-drift`. The rule: when a commit changes behavior that the "M1 as built" Revisions entry or a function doc describes, it updates that text too. Check those passages as part of every answer-review commit.

5. **Test coverage notes**
   - The BR-10 regression test was tried with the fix removed, and it failed: three race reports plus the wrong read count.
   - BR-11's status test fails without the `For` call, because the "not counted:" line never appears.
   - No test controls the order between the events read and `head.lock` acquisition. M2 needs such a seam for the departure gate (see the first Minor finding).

6. **Architectural notes for upcoming work**

   I went through each principle:

   | Principle | Result | Note |
   |---|---|---|
   | ARCH-DRY | pass | `Passed` is the only audit rule; `For` is the only vouch check |
   | ARCH-PURE | pass | `DepartureDue`, `Passed`, `NextHead` are pure; `trainAdmission` is a thin composition |
   | ARCH-PURPOSE | pass | Each Minor was fixed as a class: every reader goes through `For`; slow admission inputs are cached and read outside the mutex |
   | ARCH-MOCK | pass for M1 | git is injected into `landingScheduleSource`; the GitHub fake is M2 |
   | ARCH-CONSTRAINTS | pass | At most one config read per 30s; no git under the line mutex |
   | ARCH-SECURE | pass | Sidecars are cross-checked with `For` everywhere they are read |
   | ARCH-ORDER | flag (Minor) | Stale events snapshot ahead of the head decision, above |
   | ARCH-FUNERAL | pass | No new durable artifact this round |
   | ARCH-EVOLVE | pass | No format change this round |

   Before M2 writes `EventTrain`, the departure-ledger read needs to move under the lock.

7. **Plan revision recommendations**
   - In "M1 as built", replace "The schedule is re-read from the local remote-tracking main on each attempt" with: "The schedule is read from the local remote-tracking main and reused for 30s (`landingScheduleTTL`), outside the line mutex. The events log, the departure ledger, is read under the mutex once `head.lock` is held (M2 requirement)."

```findings
dispose:
  - id: BR-10
    disposition: addressed
    note: |
      Mutex in landingScheduleSource; with the lock removed, TestLandingScheduleSourceIsSharedAndCached fails under -race with 3 DATA RACE reports and 4 reads.
  - id: BR-11
    disposition: addressed
    note: |
      landingcmd.go:201-206 uses r.For(t, now); the test vouches with another ticket's branch and expects "not counted"; the unreadable-config line now says merges land serially.
  - id: BR-12
    disposition: addressed
    note: |
      Admission() is now built before withMutex; the schedule is cached for 30s, so the serial default runs no git on most polls; the cache test pins one read per TTL.
  - id: BR-13
    disposition: addressed
    note: |
      Passed() in order.go:96-112 is the only rule; NextHead returns (head, ok); line.go calls Passed; TestPassedIncludesTicketsThatWereNotAdmissible.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: admission-decision-toctou
    title: |
      Admission reads the events log (the departure ledger) before the line mutex and head.lock, so M2 can admit a second train on a stale "due"
    detail: |
      line.go:188-193 builds admissible from Events() outside withMutex. Once M2 records EventTrain and then releases head.lock, a waiter that read the events just before can take the head and depart early, skipping the interval. Keep the cached schedule outside the mutex, but read the events and decide DepartureDue under the mutex after head.lock is acquired (Events() does not take the mutex). Latent in M1: nothing writes EventTrain yet.
  - id: new
    severity: Minor
    kind: defect
    family: plan-core-concepts-drift
    title: |
      The openLandingLine doc and plan Revisions still say the schedule is re-read on every admission attempt; the code caches it for 30s
    detail: |
      landingqueue.go:39 and plan line 205. 2nd in this family. The rule: a commit that changes behavior described by the "M1 as built" Revisions entry or a function doc updates that text in the same commit; check those passages as part of every answer-review commit.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"aa544094-7b48-4f49-a757-d30d2ca2d389","repository":"github.com/xianxu/ariadne","issue_id":"000449","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/5c4c5ec244a460f9","reviewed_head":"96d1a9d01bb99b19bac8f2f3c9fd9666e46142c6","slot":"ariadne:2","requested_at":"2026-10-19T10:35:22.460561Z"},"phase":"finished","pid":3970,"started_at":"2026-10-19T10:35:22.690705Z","finished_at":"2026-10-19T10:50:07.782477Z","session_id":"ac6ce1ff-52cd-4ab8-95e5-9a0927bf0faa","outcome":"returned"}] -->
