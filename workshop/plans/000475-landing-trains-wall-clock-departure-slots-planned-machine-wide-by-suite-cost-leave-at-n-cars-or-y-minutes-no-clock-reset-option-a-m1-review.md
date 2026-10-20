# Boundary Review — ariadne#475 (milestone M1)

| field | value |
|-------|-------|
| issue | 475 — landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A) |
| repo | ariadne |
| issue file | workshop/issues/000475-landing-trains-wall-clock-departure-slots-planned-machine-wide-by-suite-cost-leave-at-n-cars-or-y-minutes-no-clock-reset-option-a.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 7918d3d63c4c12e9a4537b51680e2143b29fd286..b97bc6dd7d96fcd482b2fc57c6231b394cb0bf64 |
| command | sdlc milestone-close --issue 475 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T04:51:27-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Tests pass. Now I'll undo the reset fix to confirm the grant regression test fails without it.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 does what it says. Lone, granted and refused landings no longer push the next train back. A train now also leaves early when it is full (N cars) or when a car has waited too long (Y minutes). `Slots` computes the wall-clock grid correctly. The new config key `train_max_wait_minutes` is parsed with bounds. Targeted tests pass: `go test ./internal/landingline/` and `go test -run 'Landing|TrainAdmission|ParseLandingConfig' .` are both green. I also checked the reset-fix regression test by putting back `NextDue(now, s, …)` in `serial()`. With that, `TestLandingTrainGrantDepartsAloneWhenNoTrainIsDue` and `TestLandingTrainOfOneIsTodaysLanding` both fail, so the fix is reachable and pinned. Two cheap gaps keep this from SHIP:
- The README's description of departure is now out of date and doesn't mention the new config key.
- T5 is ticked, but `landing status` doesn't print the reason format the plan promised, and the golden test it promised doesn't exist.

1. **Strengths**
   - `Slots.Prev` (`cmd/sdlc/internal/landingline/slots.go:100`) floors correctly for times before the epoch and takes the offset modulo the interval. A 2000-sample property test checks `Prev ≤ t < Next`, that `Next − Prev` is one interval, and that every slot is on the grid.
   - `landing status` and admission both call `DepartureDue` (`landingcmd.go:192-193`), so status shows the same decision admission makes (ARCH-DRY).
   - Lone landings keep the "due at once" signal when they leave ready cars (`AloneNextDue`, `train.go:260`), so a lone landing can only bring the next train forward.
   - The format changes are additive. `Event.Departed` and `Readiness.Since` use `omitzero`, `since` is wired through `readinessFields` so `Extra` still keeps unknown members, and the round-trip tests and old-sidecar fallback tests exist (ARCH-EVOLVE).
   - `Waiters` is computed once per pass, under the line mutex, with the same `r.For(t, now)` check boarding uses (`line.go:268`).

2. **Critical findings:** none.

3. **Important findings**
   - **README is stale and lacks the new key.** `README.md:521-531` still lists only `train_interval_minutes` and `train_max_cars`. It still says a train is due "on the first train, one interval after the last departure, or at once after a train that left ready cars behind". It doesn't mention the full-train or max-wait early departures, `train_max_wait_minutes`, or that lone and granted landings no longer reset the clock. `helptext/landing.md` is equally stale; its update is planned for M2 (T10), but the key ships now. Fix: add the key and the new departure triggers to the README Departure bullet in this milestone.
   - **T5 is ticked but not delivered as specified.** The plan promised `train due (<reason>)` or `next train at HH:MM (<slot|max-wait>)`, with states like `train due (cars left ready)` and `(slot :MM)`, plus a golden test. In the code, a slot or left-ready departure prints a bare `"train due"`, and the next time is a full RFC3339 timestamp with no hint of whether it is the slot or the wait limit (`train.go:231-237`). The only status test change adds `"1 ready"` to an existing substring check (`landingcmd_test.go:196`). Fix: either add reasons for slot and left-ready departures and for what set the next time, with a status test covering the full, max-wait and next-time lines, or write a plan Revision narrowing T5.

4. **Minor findings**
   - The plan's departure rule is not what the code does. The plan says `DepartureDue` checks `Prev(now) > lastDeparture`, using `lastDeparture` read from `Event.Departed`. The code instead stores `NextDue` at departure time and compares against that. As a result `Departed` is written (`landingtrain.go:695,705`) but nothing reads it. Once M2 can change offsets, an offset change only takes effect after the next train.
   - The plan's Ordering section says the new binary ignores old-format `alone` records by their `Result`. `NextDeparture` (`train.go:84`) actually filters only on a non-zero `NextDue`. So an `alone` record written before the upgrade (NextDue = departed + interval) still delays the next train once.
   - ARCH-DRY: the max-wait test (`s.MaxWait > 0 && w.Ready > 0 && !w.OldestSince.IsZero()`) appears twice (`train.go:228,234`). One helper returning the wait deadline would do.
   - Mixed binaries: an older sdlc re-vouching keeps `since` through `Extra` even when the head changed. The new binary would then trust that stale `Since` and could depart early on max-wait.
   - Admission uses two clocks: `ReadyWaiters` uses `l.o.Now()` while `DepartureDue` uses the `now` passed into `trainAdmission`. Both are `time.Now` in production; it's harmless but inconsistent.

5. **Test coverage notes**
   - Covered: `DepartureDue` reasons and precedence with a fake clock, an early departure leaving the grid unchanged (via `NextDue`), the grant and lone-landing reset fix at unit and e2e level (confirmed red when the fix is reverted), `Since` carry and reset through the real `Line.Vouch`, and the config bounds.
   - Missing: status output for each reason (above), and full or max-wait admission through `tryAdmit` end to end. The latter is covered only at the `trainAdmission` unit level.

6. **Architectural notes for upcoming work**
   - ARCH-DRY: pass, apart from the duplicated predicate.
   - ARCH-PURE: pass. All decisions are pure functions over events, waiters and an injected clock.
   - ARCH-PURPOSE: pass for M1's scope. The grid isn't wired into admission yet (Offset is nil in production); that is M2's T8 by plan.
   - ARCH-MOCK: N/A. No external dependency is added.
   - ARCH-CONSTRAINTS: pass. An O(waiters) scan per pass.
   - ARCH-SECURE: pass. The config key is typed and range-checked.
   - ARCH-ORDER: pass. The decision is recomputed under `head.lock` from the log, with no state of its own between events.
   - ARCH-FUNERAL: pass. Nothing new is durable, and `alone` events get smaller.
   - ARCH-EVOLVE: pass, with the mixed-binary `since` caveat.
   - For M2: decide whether the slot check reads `Departed` (the plan's design) or the stored `NextDue`. Reading `Departed` would let a `--write` offset change take effect immediately.

7. **Plan revision recommendations**
   - "M1: the slot check is the stored `NextDue = Slots.Next(departed)` (equivalent to `Prev(now) > departed` at write time). `Departed` is recorded for M2 and history, and nothing reads it in M1."
   - "Ordering: the new binary filters `alone` records by NextDue presence, not by Result. An `alone` record written before the upgrade can delay the next train once."
   - If T5 is narrowed: "M1 status prints a reason for full and max-wait departures only. Slot and left-ready reasons and the `HH:MM` format are deferred to M2 T10."

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: user-docs-track-new-surface
    title: |
      README departure rule is stale and omits train_max_wait_minutes
    detail: |
      README.md:521-531 lists only train_interval_minutes/train_max_cars and describes departure as first train / one interval after the last / at once after left-ready. It omits the full-train and max-wait early departures, the new key, and that lone and granted landings no longer reset the clock. helptext/landing.md is equally stale.
  - id: new
    severity: Important
    kind: defect
    family: plan-claim-matches-delivery
    title: |
      T5 is ticked, but landing status lacks the promised reasons and golden test
    detail: |
      The plan promised train due (reason) / next train at HH:MM (slot or max-wait), including cars-left-ready, plus a golden test. The code prints a bare train due for slot and left-ready, and an RFC3339 time with no source. The only test change is a 1-ready substring (landingcmd_test.go:196). Add the reasons and a status test, or narrow T5 in Revisions.
  - id: new
    severity: Minor
    kind: defect
    family: plan-claim-matches-delivery
    title: |
      Event.Departed has no reader; the slot check uses the stored NextDue, not the plan's lastDeparture
    detail: |
      PQ-2 says lastDeparture reads Departed. The code compares a NextDue fixed at write time, so Departed is written but never read, and a later offset change only applies after the next train.
  - id: new
    severity: Minor
    kind: defect
    family: plan-claim-matches-delivery
    title: |
      NextDeparture does not ignore old-format alone records by Result as the Ordering section claims
    detail: |
      train.go:84 filters on non-zero NextDue only, so an alone record written before the upgrade (departed + interval) still delays the next train once.
  - id: new
    severity: Minor
    kind: defect
    family: shared-helper-not-repeated-predicate
    title: |
      The max-wait predicate is duplicated in DepartureDue
    detail: |
      train.go:228 and :234 repeat the same three-part condition; one wait-deadline helper would do.
  - id: new
    severity: Minor
    kind: defect
    family: unknown-field-carry-stale-semantics
    title: |
      An older sdlc re-vouching carries a stale since through Extra across a head change
    detail: |
      mergeExtra keeps since as an unknown member for an old binary, so after a head change the new binary trusts the stale Since and may depart early on max-wait.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e23a1d03-611b-4c3a-8af2-99eb9b7564b9","repository":"github.com/xianxu/ariadne","issue_id":"000475","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/5b6b7dbec9edb75f","reviewed_head":"b97bc6dd7d96fcd482b2fc57c6231b394cb0bf64","slot":"ariadne:4","requested_at":"2026-10-20T11:40:01.533863Z"},"phase":"finished","pid":39821,"started_at":"2026-10-20T11:40:01.627835Z","finished_at":"2026-10-20T11:51:26.839134Z","session_id":"ce54466e-eca4-404e-8a65-78527918ab82","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T05:08:04-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 475 — landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A) |
| repo | ariadne |
| issue file | workshop/issues/000475-landing-trains-wall-clock-departure-slots-planned-machine-wide-by-suite-cost-leave-at-n-cars-or-y-minutes-no-clock-reset-option-a.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 7918d3d63c4c12e9a4537b51680e2143b29fd286..fcc0be143bb399a3b0176477be2ed52ca16da8e0 |
| command | sdlc milestone-close --issue 475 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T05:08:04-07:00 |
| verdict | SHIP |

## Review

Still reviewing: the targeted test run timed out in the foreground, so I'm checking what it printed.

```verdict
verdict: SHIP
confidence: high
```

I recommend shipping M1. It delivers what the plan asks for: the `Slots` grid, the new `DepartureDue` (full, then max-wait, then the departure time, with `waitDeadline` as one helper), `train_max_wait_minutes`, and the reset fix (lone and granted landings record no `NextDue` unless they left cars ready). `landing status` now gives the same decision as admission, with its reason.

All six round-1 findings are fixed, and I checked each against the code. I reverted two of the fixes in the disposable checkout: the old-format `alone` filter and the `SinceHead` guard. With each reverted, its regression test fails (`TestNextDepartureReadsTheDeparture`, `TestStaleSinceFromAnOlderSdlc`). I then restored the file.

These tests pass at the pinned head:
- `go test ./internal/landingline/`
- the five changed `cmd/sdlc` tests
- `go vet`

One new Minor remains, in the same family as BR-6. It doesn't block.

**1. Strengths**
- `slots.go:181-199`: `Prev`/`Next` are anchored to the epoch, handle times before the epoch with a floor, and are pinned by a boundary table plus a 2000-sample property test (`slots_test.go:242`).
- `train.go:357-380`: `nextDeparture` works the next slot out from the recorded `Departed` when it decides, so a changed offset applies at once. This also closes BR-3.
- `train.go:405` `AloneNextDue`: a lone landing can only bring the next train forward, never push it back, which matches the PQ Minor revision.
- `landingcmd.go:189-199`: status calls the same `DepartureDue` that admission uses, so the two can't drift apart.
- The persisted fields are additive (`departed`, `since`, `since_head` with omitzero/omitempty) and registered in `readinessFields`, and `Extra` is still preserved.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- **Stale `since` from an older sdlc at the same head** (`train.go:448-465`). An older sdlc that vouches not-ready and then ready again at the same head carries `since`/`since_head` through `Extra` unchanged. The head check still matches, so the new binary trusts an old `Since` and max-wait can fire early. `SinceHead` guards only one of the two ways `Since` gets reset; the readiness reset has no guard. This is the 2nd finding in `unknown-field-carry-stale-semantics` (see the findings block).
- Nit: the slot construction `Slots{Interval: s.Interval, Offset: *s.Offset}` appears twice (`train.go:373`, `:400`). A `Schedule.grid()` helper would remove the repeat.
- Nit: a train record from before #475 (no `departed`) whose run outlasted the interval reads as "cars left ready". The decision (due at once) is the same as before; only the label is wrong, and only for that one record.

**5. Test coverage notes**
- Covered:
  - each departure reason and its precedence;
  - the grid staying put after an early departure;
  - a grant or lone landing leaving the next departure unchanged, both as a unit test and in the e2e `TestLandingTrainGrantDepartsAloneWhenNoTrainIsDue`;
  - the `Since` lifecycle through `Line.Vouch`;
  - `departed` round-tripping;
  - the config bounds;
  - the status reason string.
- The T5 test checks substrings rather than a full golden output, but it pins the new reason text and would fail without the change. I accept it.

**6. Architecture**
- **ARCH-DRY:** pass. The BR-5 fix is in, and the only repeat left is the slot-construction nit.
- **ARCH-PURE:** pass. The decision logic is pure in `landingline`, and status and admission are thin callers.
- **ARCH-PURPOSE:** pass for M1. The plan offset is nil in production until M2, as planned. The README and help text say "with a machine plan", which describes M2 behaviour ahead of time; that's acceptable.
- **ARCH-MOCK:** N/A. M1 adds no external dependency.
- **ARCH-CONSTRAINTS:** pass. Each admission poll adds one scan over the waiters (O(waiters)).
- **ARCH-SECURE:** pass, with the Minor above. Sidecars are parsed into typed values, and a stale `since` falls back to `At`.
- **ARCH-ORDER:** mostly pass. The `head.lock` serialisation is unchanged. But the first-ready state is now three fields (`Ready`, `Since`, `SinceHead`) whose legal combinations aren't written down anywhere. M2 could fold the validity stamp into one record.
- **ARCH-FUNERAL:** pass. Nothing new is durable. Events gain `departed`, `alone` events lose `next_due`, and the existing 30-day retention covers both.
- **ARCH-EVOLVE:** pass. The new fields are additive and survive a round trip.

**7. Plan revision recommendations**
- Optional: note in Revisions that T5 delivered a substring status test, and that "interval" is a reason value when there is no plan (the plan lists only slot|max-wait).

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:521-538 and helptext/landing.md now list train_max_wait_minutes, full and max-wait early departures, and that lone/granted landings never push the train back.
  - id: BR-2
    disposition: addressed
    note: |
      Status calls DepartureDue and prints the reason (slot, interval, max wait, full, cars left ready); landingcmd_test.go:852-862 pins "next train at 2026-10-20 10:00 (interval)".
  - id: BR-3
    disposition: addressed
    note: |
      nextDeparture works out the slot from departedOf(e) when it decides (train.go:373); TestNextDepartureReadsTheDeparture covers it.
  - id: BR-4
    disposition: addressed
    note: |
      The alone case is skipped by Result (train.go:370); reverting it makes TestNextDepartureReadsTheDeparture fail (checked in the checkout).
  - id: BR-5
    disposition: addressed
    note: |
      waitDeadline helper (train.go:330) used for both the due check and the reported next time.
  - id: BR-6
    disposition: addressed
    note: |
      SinceHead guard in firstReady (train.go:449); reverting it makes TestStaleSinceFromAnOlderSdlc fail (checked). A sibling case is raised separately.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: unknown-field-carry-stale-semantics
    title: |
      An older sdlc's not-ready then ready re-vouch at the same head still carries a stale since
    detail: |
      This is the 2nd finding in this family. The rule: a derived field that an older writer can carry through Extra must be stamped against every condition that resets it, not just one. SinceHead covers a head change but not a not-ready vouch, so mergeExtra keeps since across old-binary not-ready then ready at the same head, and firstReady trusts it (train.go:448-465). Fix: stamp the reset conditions too, for example a since_at equal to the At of the vouch that set Since, read as valid only while every vouch since then was ready. Prevalence: 2 of 2 carried fields in this family (since; head-reset fixed, readiness-reset open). The window needs mixed binaries on one line, so this is Minor.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"c159b511-dfc9-4d80-91e3-f34150a94d55","repository":"github.com/xianxu/ariadne","issue_id":"000475","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/30ef4d045dbeb182","reviewed_head":"fcc0be143bb399a3b0176477be2ed52ca16da8e0","slot":"ariadne:4","requested_at":"2026-10-20T11:55:52.686773Z"},"phase":"finished","pid":58606,"started_at":"2026-10-20T11:55:52.902227Z","finished_at":"2026-10-20T12:08:04.201777Z","session_id":"dfa6f6eb-bdcf-4fe1-87d5-aa1bbe132208","outcome":"returned"}] -->
