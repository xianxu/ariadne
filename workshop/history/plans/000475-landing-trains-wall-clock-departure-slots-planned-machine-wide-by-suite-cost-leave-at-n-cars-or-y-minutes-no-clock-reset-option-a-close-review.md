# Boundary Review — ariadne#475 (whole-issue close)

| field | value |
|-------|-------|
| issue | 475 — landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A) |
| repo | ariadne |
| issue file | workshop/issues/000475-landing-trains-wall-clock-departure-slots-planned-machine-wide-by-suite-cost-leave-at-n-cars-or-y-minutes-no-clock-reset-option-a.md |
| boundary | whole-issue close |
| milestone | — |
| window | c6fb9859856d942bef3be366946737aa3c47ba64..fc32559e438281ea63d14ab4fa4905e0c299e0b7 |
| command | sdlc close --issue 475 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T06:41:01-07:00 |
| verdict | SHIP |

## Review

Checking BR-7's `SinceStamp` fix and the test behind it, then giving the verdict.

```verdict
verdict: SHIP
confidence: medium
```

Both open findings are fixed. **BR-7** is fixed in code. `Readiness.SinceStamp` stores the `At` of the vouch that wrote `Since`, and `firstReady` trusts `Since` only while that stamp still equals `At` (`train.go:189-194, 275-280`). This handles every reset an older sdlc can't apply, so it fixes the whole family rule and not just the one reported case. I confirmed the regression test is real: when I removed the stamp check, `TestStaleSinceFromAnOlderSdlc` failed ("ready again, same: stale since read as 09:00, want 09:30"), and it passed again once I restored it. **BR-15** is fixed by a Revisions entry (`plan.md:248-259`), and every test it cites exists and checks what it claims. The 10-minute modulo case and the `offset_minutes:1440` refusal are in `plan_test.go:75,93`. The "warned once" count is asserted in `landingplan_test.go:131-133` (the test runs the source three times with a 0 TTL). Nothing new blocks SHIP. One caveat on confidence: the `cmd/sdlc` integration tests did not finish within the 600 s time limit, so I only verified the `landingline` unit package by running it.

1. **Strengths**
   - `firstReady`/`carrySince` (`train.go:273-291`): the stamp fixes the stale-`Since` problem as a class, and the test covers both a new head and a not-ready vouch followed by a ready one at the same head.
   - `landingScheduleSource` (`landingqueue.go:240-270`): one cached constructor through `effectiveSchedule`, separate one-time warnings for a bad config and a bad plan, and the warning count is tested.
   - `DecodePlan`'s strict codec: it refuses repeated keys at every level and keeps unknown members through a rewrite, at the top level and per repo (`plan_test.go:63-105`).
   - The Done-when fixture `TestPlannedReposNeverDepartInEachOthersWindow` steps the real pure `DepartureDue` with a fake clock.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none new.

5. **Test coverage notes:** I ran `go test ./cmd/sdlc/internal/landingline/` and it passed. I did not finish running the `cmd/sdlc` tests (`Landing|Plan|Schedule`) because they timed out, so the e2e side is not verified by this review. The checkout was restored to clean afterwards.

6. **Architectural notes:**
   - **ARCH-DRY: pass.** One `effectiveSchedule` and one max-wait helper.
   - **ARCH-PURE: pass.** Slots, DepartureDue, PackOffsets and the plan codec are pure; IO sits in `landingconfig`/`landingplan`.
   - **ARCH-PURPOSE: pass.** BR-7 and BR-15 were fixed at the rule level.
   - **ARCH-MOCK: pass.** Tests use real temp files via `XDG_STATE_HOME`.
   - **ARCH-CONSTRAINTS: pass.** Overloads are named, and the 30 s TTL is honoured.
   - **ARCH-SECURE: pass.** The plan is parsed strictly, and a bad plan falls back visibly.
   - **ARCH-ORDER: pass.** The departure decision is recomputed from the log under `head.lock`.
   - **ARCH-FUNERAL: pass.** The plan file is a single overwritten file, and events keep their existing retention.
   - **ARCH-EVOLVE: pass.** `since`/`since_stamp` are additive, and an older writer's carry-through is now detected.

   For upcoming work: T11 (live dry run and the hub's one-hour observation) is still unticked. That is correct, because it happens after landing.

7. **Plan revision recommendations:** none. The plan matches the code.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      SinceStamp (train.go:189-194) makes firstReady trust Since only when stamped by the current vouch; reverting the stamp check turns TestStaleSinceFromAnOlderSdlc red (ready again, same head case).
  - id: BR-15
    disposition: addressed
    note: |
      Revisions entry plan.md:248-259 replaces both claims; cited tests exist and assert them (plan_test.go:75 modulo case, :93 offset 1440 refusal; landingplan_test.go:131 warned exactly once).
```

The verdict stays **SHIP**, and I now have a test result to back it.

- **Wide run timed out:** I had run every `cmd/sdlc` test with "Landing", "Plan" or "Schedule" in its name. That run stopped at Go's 10-minute limit. I only kept the last few lines of output, so I can't name the test that ran long, and that run is unverified.
- **Narrow run passed:** I then ran the tests this issue added or extended, and all seven passed in about 15 s: `TestParseLandingConfig`, `TestLandingPlanCommand`, `TestScheduleSourceReadsThePlan`, `TestPlannedReposNeverDepartInEachOthersWindow`, `TestMachinePlanPath`, `TestFleetPlanInputs` and the grant end-to-end test `TestLandingTrainGrantDepartsAloneWhenNoTrainIsDue`.
- **BR-7 test is real:** with its fix removed, `TestStaleSinceFromAnOlderSdlc` fails. With the fix restored, the `landingline` unit tests pass.
- **Checkout is clean:** I restored `train.go`, and `git status` shows no changes.

Not covered: the full `cmd/sdlc` suite. If it should be checked, run it with a longer `-timeout` or split it up.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ac1c618b-8c05-464a-a239-1388169a2a59","repository":"github.com/xianxu/ariadne","issue_id":"000475","recipe":"milestone-review","round":6,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/f41731c89a9d1958","reviewed_head":"fc32559e438281ea63d14ab4fa4905e0c299e0b7","slot":"ariadne:4","requested_at":"2026-10-20T13:29:31.273042Z"},"phase":"finished","pid":48153,"started_at":"2026-10-20T13:29:31.587198Z","finished_at":"2026-10-20T13:41:00.853025Z","session_id":"37e6231c-1a97-4ea7-9a3c-7967a3f85c3f","outcome":"returned"}] -->
